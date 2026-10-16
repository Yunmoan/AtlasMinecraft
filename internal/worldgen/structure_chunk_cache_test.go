package worldgen

import (
	"reflect"
	"sync"
	"testing"
)

func TestStructureChunkSnapshotConcurrentIsolation(t *testing.T) {
	g, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	snapshots := make([]*structureChunkSnapshot, 4)
	errors := make([]error, len(snapshots))
	for i := range snapshots {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			snapshots[i], errors[i] = g.structureChunkBefore(0, 0, 3)
		}(i)
	}
	wg.Wait()
	for i := range snapshots {
		if errors[i] != nil {
			t.Fatal(errors[i])
		}
		if snapshots[i] != snapshots[0] {
			t.Fatal("concurrent preceding-stage requests did not coalesce")
		}
	}
	uncached, err := g.structureRegionBefore(0, 0, 3)
	if err != nil {
		t.Fatal(err)
	}
	key := featureChunkPosition{0, 0}
	if !reflect.DeepEqual(uncached.overlays[key], snapshots[0].overlay) {
		t.Fatal("cached center changed preceding feature writes or updates")
	}
	regions := make([]*OverworldFeatureRegion, 2)
	for i := range regions {
		region, err := NewOverworldFeatureRegion(snapshots[0].base.terrain)
		if err != nil {
			t.Fatal(err)
		}
		region.borrowed = make(map[featureChunkPosition]bool)
		region.biomeBases = make(map[featureChunkPosition]*overworldBase)
		region.installStructureChunk(key, snapshots[0])
		regions[i] = region
	}
	p := FeaturePosition{0, 70, 0}
	want, err := regions[1].BlockAt(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := regions[0].SetBlock(p, SurfaceBlockState{Name: "minecraft:gold_block"}); err != nil {
		t.Fatal(err)
	}
	if err := regions[0].MarkPostProcessing(p); err != nil {
		t.Fatal(err)
	}
	got, err := regions[1].BlockAt(p)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("structure write escaped private overlay: %v %v", got, err)
	}
	if !reflect.DeepEqual(uncached.cloneChunk(0, 0), regions[1].cloneChunk(0, 0)) {
		t.Fatal("install changed untouched chunk or shared update flags")
	}
	// A later-stage cache entry must never return an earlier-stage snapshot.
	later, err := g.structureChunkBefore(0, 0, 4)
	if err != nil || later == snapshots[0] {
		t.Fatalf("preceding-stage cache reused an incompatible stage: %v", err)
	}
}
