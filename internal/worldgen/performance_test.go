package worldgen

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
)

// Fixed adjacent chunks measure cold terrain plus shared neighborhood reuse.
// Run with -benchtime=1x; one operation is the entire nine-chunk area.
func BenchmarkOverworldArea(b *testing.B) {
	for _, structures := range []bool{false, true} {
		for _, workers := range []int{1, 4} {
			b.Run(fmt.Sprintf("structures=%t/workers=%d", structures, workers), func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					generator, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: structures})
					if err != nil {
						b.Fatal(err)
					}
					generator.SetTerrainWorkers(workers)
					jobs := make(chan featureChunkPosition, 9)
					for x := int32(-1); x <= 1; x++ {
						for z := int32(-1); z <= 1; z++ {
							jobs <- featureChunkPosition{x, z}
						}
					}
					close(jobs)
					var wg sync.WaitGroup
					for i := 0; i < workers; i++ {
						wg.Add(1)
						go func() {
							defer wg.Done()
							for p := range jobs {
								if _, err := generator.Generate(p.chunkX, p.chunkZ); err != nil {
									b.Error(err)
								}
							}
						}()
					}
					wg.Wait()
				}
			})
		}
	}
}

func TestOverworldPerformanceOutput(t *testing.T) {
	generator, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	wanted := []string{
		"0840b7b9dd99dd16ab04c6218423a9bf2a37390511bf3766d3cd27649f6cdf87",
		// Version 10 settles generation-time bubble columns and leaf distances.
		"de996af4874ae8daab1afb716e501834eb4fc265d8c62962cf58f671182d8124",
		"aa37794fa094d75ee7e57ac6f7b773d016f1c0686264e4a7aabe7e2bb96b045a",
	}
	for i, p := range []featureChunkPosition{{-1, 0}, {0, 0}, {1, 1}} {
		result, err := generator.Generate(p.chunkX, p.chunkZ)
		if err != nil {
			t.Fatal(err)
		}
		// fmt serializes slices and sorted maps consistently, including biomes,
		// material/fluid arrays, surface properties and unsupported coverage.
		digest := sha256.Sum256([]byte(fmt.Sprintf("%v %v %v %v", *result.Terrain, result.Biomes, result.SkippedFeatures, result.BlockEntities)))
		if fmt.Sprintf("%x", digest) != wanted[i] {
			t.Errorf("chunk %d,%d output changed: %x", p.chunkX, p.chunkZ, digest)
		}
		t.Logf("chunk %d,%d sha256=%x", p.chunkX, p.chunkZ, digest)
	}
}
