package worldgen

import (
	"atlas/internal/nbt"
	"reflect"
	"sync"
	"testing"
)

func TestStrongholdRingsLazyNearSpawnAndConcurrentQueries(t *testing.T) {
	cache, err := newOverworldStructureCache(1234)
	if err != nil {
		t.Fatal(err)
	}
	g := &OverworldGenerator{structures: cache}
	distance := cache.placement.Sets["minecraft:strongholds"].Placement.Distance
	for x := -2 * distance; x <= 2*distance; x++ {
		for z := -2 * distance; z <= 2*distance; z++ {
			if int64(x)*int64(x)+int64(z)*int64(z) >= 4*int64(distance)*int64(distance) {
				continue
			}
			if found, err := g.strongholdPlacementAt(x, z); err != nil || found {
				t.Fatal("spawn-area shortcut reported a nonexistent stronghold", x, z, found, err)
			}
		}
	}
	if len(cache.placement.RingPositions) != 0 {
		t.Fatal("spawn-area queries eagerly sampled distant stronghold rings")
	}
	type queryResult struct {
		position StructureChunkPosition
		found    bool
		err      error
	}
	positions := []StructureChunkPosition{{-100, -125}, {0, 0}, {64, 0}, {2000, -2000}, {-100, -125}, {-64, 0}, {10, 10}, {1874997, -1874997}}
	results := make(chan queryResult, len(positions)*4)
	begin := make(chan struct{})
	var ready sync.WaitGroup
	for i := 0; i < 4; i++ {
		for _, position := range positions {
			ready.Add(1)
			go func(p StructureChunkPosition) {
				defer ready.Done()
				<-begin
				found, err := g.strongholdPlacementAt(p.X, p.Z)
				results <- queryResult{p, found, err}
			}(position)
		}
	}
	close(begin)
	ready.Wait()
	close(results)
	actualRings := cache.placement.RingPositions["minecraft:strongholds"]
	if len(actualRings) != 128 || actualRings[0] != (StructureChunkPosition{-100, -125}) {
		t.Fatal("concurrent lazy preparation lost original biome-relocated rings", len(actualRings))
	}
	wanted := make(map[StructureChunkPosition]bool, len(actualRings))
	for _, p := range actualRings {
		wanted[p] = true
		if int64(p.X)*int64(p.X)+int64(p.Z)*int64(p.Z) < 4*int64(distance)*int64(distance) {
			t.Fatal("lazy spawn-area bound excluded an actual relocated ring", p)
		}
	}
	for result := range results {
		if result.err != nil || result.found != wanted[result.position] {
			t.Fatal("concurrent placement differs from original relocated rings", result.position, result.found, result.err)
		}
	}
}

func TestOverworldStrongholdConcurrentClipReplayAfterEviction(t *testing.T) {
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	g.SetTerrainWorkers(1)
	want, err := g.Generate(-101, -126)
	if err != nil {
		t.Fatal(err)
	}
	wantStart, found := g.StrongholdStartSnapshot(-100, -125)
	if !found {
		t.Fatal("actual portal generation has no saved start")
	}
	// Replay after eviction must use the immutable original layout, even when
	// the separately retained save snapshot already says Mob/Chest were placed.
	g.structures.mu.Lock()
	clear(g.structures.strongholdBlocks)
	g.structures.mu.Unlock()
	type generatedResult struct {
		chunk *GeneratedOverworldChunk
		err   error
	}
	results := make(chan generatedResult, 2)
	begin := make(chan struct{})
	var ready sync.WaitGroup
	for i := 0; i < 2; i++ {
		ready.Add(1)
		go func() {
			defer ready.Done()
			<-begin
			chunk, err := g.Generate(-101, -126)
			results <- generatedResult{chunk, err}
		}()
	}
	ready.Add(1)
	go func() {
		defer ready.Done()
		<-begin
		for i := 0; i < 100; i++ {
			if snapshot, exists := g.StrongholdStartSnapshot(-100, -125); exists {
				snapshot.Pieces[0].Data["poison"] = nbt.Tag{Type: nbt.String, Value: "bad"}
			}
		}
	}()
	close(begin)
	ready.Wait()
	close(results)
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if !reflect.DeepEqual(result.chunk.Terrain, want.Terrain) || !reflect.DeepEqual(result.chunk.BlockEntities, want.BlockEntities) || !reflect.DeepEqual(result.chunk.Biomes, want.Biomes) {
			t.Fatal("concurrent evicted clip replay changed actual portal geometry or complete chest/spawner NBT")
		}
	}
	if got, exists := g.StrongholdStartSnapshot(-100, -125); !exists || !reflect.DeepEqual(got, wantStart) {
		t.Fatal("concurrent replay or detached snapshot edit changed the saved original start")
	}
}
