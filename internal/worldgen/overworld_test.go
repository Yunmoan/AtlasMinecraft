package worldgen

import (
	"reflect"
	"sync"
	"testing"
)

func TestOverworldGeneratorConcurrentDeterminism(t *testing.T) {
	generator, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	var results [2]*GeneratedOverworldChunk
	var errors [2]error
	var workers sync.WaitGroup
	for i := range results {
		workers.Add(1)
		go func(i int) { defer workers.Done(); results[i], errors[i] = generator.Generate(-1, 0) }(i)
	}
	workers.Wait()
	for _, err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(results[0], results[1]) {
		t.Fatal("concurrent requests generated different blocks or biomes")
	}
	if len(results[0].SkippedFeatures) != 0 {
		t.Fatalf("validated top-level features unexpectedly skipped: %v", results[0].SkippedFeatures)
	}
	if len(results[0].SkippedStages) == 0 {
		t.Fatal("incomplete world generation stages were advertised as complete")
	}
	if _, err := generator.Generate(0, 0); err != nil {
		t.Fatal(err)
	}
	// Mutating a returned result must not poison the reusable base cache.
	results[0].Terrain.Blocks[0] = NoiseLava
	results[0].Terrain.SurfaceBlocks[0] = SurfaceBlockState{Name: "minecraft:diamond_block"}
	for _, state := range results[0].Terrain.SurfaceBlocks {
		if len(state.Properties) > 0 {
			state.Properties["test-poison"] = "true"
			break
		}
	}
	again, err := generator.Generate(-1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again, results[1]) {
		t.Fatal("generation depends on request order or previous result edits")
	}
}
