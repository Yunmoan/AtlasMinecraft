package worldgen

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
)

func TestParallelOverworldAreaMatchesVersion4Output(t *testing.T) {
	generator, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	positions := []featureChunkPosition{{-1, 0}, {0, 0}, {1, 1}}
	wanted := []string{
		"6b4edc8da3bdfdeb128064366a4e10ccc7f759fbe1565cbfc7f1430848f1bc8b",
		"262ac55c689c43dcb4080cd9a3fee8adadebbde6f85d116dada1882871d4e06e",
		"17d4aa33bd328f8f8c251e2c8d3d000089c62b50a7ec9f95c82fa2270e5cc2b0",
	}
	var wg sync.WaitGroup
	for i, p := range positions {
		wg.Add(1)
		go func(i int, p featureChunkPosition) {
			defer wg.Done()
			result, err := generator.Generate(p.chunkX, p.chunkZ)
			if err != nil {
				t.Error(err)
				return
			}
			digest := sha256.Sum256([]byte(fmt.Sprintf("%v %v %v", *result.Terrain, result.Biomes, result.SkippedFeatures)))
			if fmt.Sprintf("%x", digest) != wanted[i] {
				t.Errorf("parallel chunk %d,%d output changed: %x", p.chunkX, p.chunkZ, digest)
			}
		}(i, p)
	}
	wg.Wait()
}
