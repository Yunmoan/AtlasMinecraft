package worldgen

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
)

func TestParallelOverworldAreaMatchesPostprocessedOutput(t *testing.T) {
	generator, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	positions := []featureChunkPosition{{-1, 0}, {0, 0}, {1, 1}}
	wanted := []string{
		"0840b7b9dd99dd16ab04c6218423a9bf2a37390511bf3766d3cd27649f6cdf87",
		"de996af4874ae8daab1afb716e501834eb4fc265d8c62962cf58f671182d8124",
		"aa37794fa094d75ee7e57ac6f7b773d016f1c0686264e4a7aabe7e2bb96b045a",
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
			digest := sha256.Sum256([]byte(fmt.Sprintf("%v %v %v %v", *result.Terrain, result.Biomes, result.SkippedFeatures, result.BlockEntities)))
			if fmt.Sprintf("%x", digest) != wanted[i] {
				t.Errorf("parallel chunk %d,%d output changed: %x", p.chunkX, p.chunkZ, digest)
			}
		}(i, p)
	}
	wg.Wait()
}
