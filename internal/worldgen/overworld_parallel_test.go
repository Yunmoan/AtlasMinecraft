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
		// The same version-19 baseline as TestOverworldPerformanceOutput.
		"be55837590fb68ae70c2b942b37acd209a44c25b3a9aa9e1afaf4d289852df8f",
		"900ddd758a3da55b1024f3f6e172d99310b873ded4744138badb3df66a045994",
		"df646254ffc2de48d935ef3ab8827cf68079d6fd8338833deaf69063c59e0884",
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
