package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestIndependentBeehiveReplay(t *testing.T) {
	config,err:=LoadTreeConfiguration("minecraft:oak_bees_005");if err!=nil { t.Fatal(err) }
	for i:=range config.decorators { if config.decorators[i].kind=="minecraft:beehive" { config.decorators[i].probability=1 } }
	origin:=FeaturePosition{X:15,Y:41,Z:-1}
	context:=treeShuffleContext{seed:1234,salt:7654321,sourceX:0,sourceZ:-1,step:9,index:7}
	generate:=func(seed int64, context treeShuffleContext) (*decoratorProbeWorld,uint64) {
		world:=&decoratorProbeWorld{changed:make(map[FeaturePosition]SurfaceBlockState),initial:make(map[FeaturePosition]SurfaceBlockState),entities:make(map[FeaturePosition]nbt.Object),origin:origin}
		random:=NewWorldgenRandom(seed)
		invocation:=config.WithBeehiveShuffle(context.forTree("minecraft:oak_bees_005",origin,random))
		if success,err:=PlaceTreeFeature(world,random,origin,invocation);err!=nil || !success { t.Errorf("tree invocation failed %v",err) }
		return world,random.NextLong()
	}
	for seed:=int64(0);seed<16;seed++ {
		baseline,next:=generate(seed,context)
		if len(baseline.entities)!=1 { t.Fatalf("forced hive absent at seed %d",seed) }
		var wg sync.WaitGroup
		for worker:=0;worker<8;worker++ {
			wg.Add(1)
			go func() { defer wg.Done();replay,replayNext:=generate(seed,context)
				if !reflect.DeepEqual(baseline.changed,replay.changed) || !reflect.DeepEqual(baseline.entities,replay.entities) || next!=replayNext { t.Errorf("independent tree replay changed across workers at seed %d",seed) }
			}()
		}
		wg.Wait()
		// Recreating the context after a generator restart has no shared cache.
		restarted,restartedNext:=generate(seed,treeShuffleContext{seed:1234,salt:7654321,sourceX:0,sourceZ:-1,step:9,index:7})
		if !reflect.DeepEqual(baseline.changed,restarted.changed) || !reflect.DeepEqual(baseline.entities,restarted.entities) || next!=restartedNext { t.Fatal("tree replay changed after context recreation") }
	}
}

func TestIndependentBeehiveShuffleDoesNotConsumeWorldgenRandom(t *testing.T) {
	context:=treeShuffleContext{seed:1234,salt:7654321,sourceX:-1,sourceZ:0,step:9,index:11}
	random:=NewWorldgenRandom(123)
	random.NextGaussian()
	before:=fmt.Sprint(random.source.lo,random.source.hi,random.gaussian,random.hasGaussian)
	positions:=[]FeaturePosition{{X:1},{X:2},{X:3},{X:4},{X:5},{X:6},{X:7},{X:8}}
	shuffle:=context.forTree("oak_bees_005",FeaturePosition{X:-1,Y:64,Z:15},random)
	shuffle(positions)
	if after:=fmt.Sprint(random.source.lo,random.source.hi,random.gaussian,random.hasGaussian);after!=before { t.Fatal("independent shuffle consumed worldgen RNG or Gaussian cache") }
	seen:=map[string]bool{}
	for salt:=int64(0);salt<16;salt++ {
		context.salt=salt
		p:=[]FeaturePosition{{X:1},{X:2},{X:3},{X:4},{X:5},{X:6},{X:7},{X:8}}
		context.forTree("oak_bees_005",FeaturePosition{X:-1,Y:64,Z:15},random)(p)
		seen[fmt.Sprint(p)]=true
	}
	if len(seen)<2 { t.Fatal("persisted independent salts did not affect hive candidate order") }
}
