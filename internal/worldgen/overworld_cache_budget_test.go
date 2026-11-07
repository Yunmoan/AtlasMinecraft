package worldgen

import (
	"reflect"
	"testing"
)

func TestOverworldCacheBudgetPreservesCompleteOutput(t *testing.T) {
	baseline, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	positions := [][2]int32{{-1, 0}, {1, 1}}
	wanted := make([]*GeneratedOverworldChunk, len(positions))
	for i, p := range positions {
		wanted[i], err = baseline.Generate(p[0], p[1])
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, limit := range []int{1, 128} {
		g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{TerrainCacheLimit: limit})
		if err != nil {
			t.Fatal(err)
		}
		for i, p := range positions {
			got, err := g.Generate(p[0], p[1])
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, wanted[i]) {
				t.Fatalf("cache limit %d changed full output at %v", limit, p)
			}
		}
		g.mu.Lock()
		count := len(g.cache)
		g.mu.Unlock()
		if count > limit {
			t.Fatalf("completed terrain cache retained %d entries beyond budget %d", count, limit)
		}
	}
}
