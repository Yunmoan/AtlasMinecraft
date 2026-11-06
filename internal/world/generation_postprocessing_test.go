package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"reflect"
	"testing"
)

func TestGeneratedPostprocessingAnvilRetainsOrderAndDuplicates(t *testing.T) {
	terrain := &worldgen.NoiseChunkResult{X: -2, Z: 3, MinY: -64, Height: 32, Blocks: make([]worldgen.NoiseMaterial, 8192), FluidUpdates: make([]bool, 8192), PostProcessing: make([]bool, 8192), PostProcessingOrder: []int{4367, 1, 271, 1, 4367}}
	// All final blocks are air. Stale marks must survive; membership arrays
	// must neither filter them nor synthesize extra entries.
	terrain.FluidUpdates[17], terrain.PostProcessing[18] = true, true
	biomes := make([][64]string, 2)
	for i := range biomes {
		for j := range biomes[i] {
			biomes[i][j] = "minecraft:plains"
		}
	}
	chunk, err := chunkFromNoiseBounds(terrain, biomes, true)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int16{{1, 31, 1}, {31, 31}}
	assert := func(root nbt.Tag) {
		t.Helper()
		object, _ := root.Object()
		sections, _ := object["PostProcessing"].Elements()
		got := make([][]int16, len(sections))
		for i, s := range sections {
			entries, _ := s.Elements()
			for _, p := range entries {
				got[i] = append(got[i], p.Value.(int16))
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("postprocessing lists changed", got, want)
		}
	}
	name, root, err := chunk.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	assert(root)
	dir := t.TempDir()
	if err := anvil.WriteChunk(dir, -2, 3, name, root); err != nil {
		t.Fatal(err)
	}
	name, root, exists, err := anvil.ReadChunk(dir, -2, 3)
	if err != nil || !exists {
		t.Fatal("reload", err)
	}
	assert(root)
	reopened, err := chunkFromAnvilBounds(name, root, -2, 3, -64, -32, true)
	if err != nil {
		t.Fatal(err)
	}
	reopened.SetBlock(1, 0, -64, StateStone)
	_, root, err = reopened.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	assert(root)
}
