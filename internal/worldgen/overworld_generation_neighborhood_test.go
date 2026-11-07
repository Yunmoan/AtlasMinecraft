package worldgen

import (
	"errors"
	"testing"
)

// These production coordinates failed repeatedly in run/logs/latest.log.
func TestOverworldRuntimeGenerationNeighborhood(t *testing.T) {
	for _, c := range []struct {
		name string
		x, z int32
	}{
		{"ocean_ruin", 1, -32},
		{"large_dripstone", 11, -62},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, err := NewOverworldGeneratorWithOptions(-8981361456740854275, OverworldGeneratorOptions{GenerateStructures: true})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := g.Generate(c.x, c.z); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFeatureRegionLazyChunkAccess(t *testing.T) {
	region, err := NewOverworldFeatureRegion(newStoneNoiseChunk(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	p := FeaturePosition{-17, 70, -1}
	if _, err := region.BlockAt(p); err == nil {
		t.Fatal("standalone region accepted a missing neighbor")
	}
	failure := errors.New("neighbor generation failed")
	calls := 0
	remote := newStoneNoiseChunk(-2, -1)
	region.borrowed = make(map[featureChunkPosition]bool)
	region.loadChunk = func(x, z int32) error {
		if x != -2 || z != -1 {
			t.Fatalf("negative coordinate mapped to %d,%d", x, z)
		}
		calls++
		if calls == 1 {
			return failure
		}
		key := featureChunkPosition{x, z}
		region.chunks[key] = remote
		region.borrowed[key] = true
		return nil
	}
	if _, err := region.BlockAt(p); !errors.Is(err, failure) {
		t.Fatalf("lost load error: %v", err)
	}
	for i := 0; i < 2; i++ {
		if state, err := region.BlockAt(p); err != nil || state.Name != "minecraft:stone" {
			t.Fatalf("retry: %v/%v", state, err)
		}
	}
	if calls != 2 {
		t.Fatalf("loaded cached neighbor %d times", calls)
	}
	if err := region.SetBlock(p, SurfaceBlockState{Name: "minecraft:gold_block"}); err != nil {
		t.Fatal(err)
	}
	if remote.surfaceState(p.X&15, p.Y, p.Z&15).Name != "minecraft:stone" {
		t.Fatal("lazy neighbor write mutated the cached terrain")
	}
	region.loadChunk = func(int32, int32) error { return nil }
	if _, err := region.BlockAt(FeaturePosition{64, 70, 0}); err == nil {
		t.Fatal("loader success without a chunk was accepted")
	}
}

func TestStructureRegionLazyChunkRetainsPrecedingStage(t *testing.T) {
	region, err := NewOverworldFeatureRegion(newStoneNoiseChunk(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	region.borrowed = make(map[featureChunkPosition]bool)
	region.biomeBases = make(map[featureChunkPosition]*overworldBase)
	p := FeaturePosition{-17, 70, -1}
	remote := newStoneNoiseChunk(-2, -1)
	index, _ := remote.Index(p.X&15, p.Y, p.Z&15)
	snapshot := &structureChunkSnapshot{base: &overworldBase{terrain: remote}, overlay: &featureChunkOverlay{blocks: map[int]SurfaceBlockState{index: {Name: "minecraft:glass"}}}}
	ready := make(chan struct{})
	close(ready)
	g := &OverworldGenerator{preceding: map[structureChunkKey]*structureChunkEntry{{-2, -1, 4}: {ready: ready, snapshot: snapshot, complete: true}}}
	g.bindStructureChunkLoader(region, 4)
	if state, err := region.BlockAt(p); err != nil || state.Name != "minecraft:glass" {
		t.Fatalf("lost preceding decoration: %v/%v", state, err)
	}
	if err := region.SetBlock(p, SurfaceBlockState{Name: "minecraft:gold_block"}); err != nil {
		t.Fatal(err)
	}
	if snapshot.overlay.blocks[index].Name != "minecraft:glass" || remote.surfaceState(p.X&15, p.Y, p.Z&15).Name != "minecraft:stone" {
		t.Fatal("structure read view mutated a shared snapshot")
	}
}
