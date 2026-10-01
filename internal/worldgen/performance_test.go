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
	for _, workers := range []int{1, 4} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				generator, err := NewOverworldGenerator(1234)
				if err != nil {
					b.Fatal(err)
				}
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

func TestOverworldPerformanceOutput(t *testing.T) {
	generator, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	wanted := []string{
		"6b4edc8da3bdfdeb128064366a4e10ccc7f759fbe1565cbfc7f1430848f1bc8b",
		"262ac55c689c43dcb4080cd9a3fee8adadebbde6f85d116dada1882871d4e06e",
		"17d4aa33bd328f8f8c251e2c8d3d000089c62b50a7ec9f95c82fa2270e5cc2b0",
	}
	for i, p := range []featureChunkPosition{{-1, 0}, {0, 0}, {1, 1}} {
		result, err := generator.Generate(p.chunkX, p.chunkZ)
		if err != nil {
			t.Fatal(err)
		}
		// fmt serializes slices and sorted maps consistently, including biomes,
		// material/fluid arrays, surface properties and unsupported coverage.
		digest := sha256.Sum256([]byte(fmt.Sprintf("%v %v %v", *result.Terrain, result.Biomes, result.SkippedFeatures)))
		if fmt.Sprintf("%x", digest) != wanted[i] {
			t.Fatalf("chunk %d,%d output changed: %x", p.chunkX, p.chunkZ, digest)
		}
		t.Logf("chunk %d,%d sha256=%x", p.chunkX, p.chunkZ, digest)
	}
}
