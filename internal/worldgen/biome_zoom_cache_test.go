package worldgen

import (
	"fmt"
	"testing"
)

func TestCachedBiomeZoomRetainsCornerSelection(t *testing.T) {
	for _, seed := range []int64{0, 1234, -987654321} {
		for _, bounds := range [][4]int32{{0, 0, -64, 384}, {-17, 23, 0, 128}, {1874999, -1875000, 0, 256}} {
			zoom := newChunkBiomeZoom(seed, bounds[0], bounds[1], bounds[2], bounds[3])
			for x := int32(-3); x <= 18; x++ {
				for z := int32(-3); z <= 18; z++ {
					for y := bounds[2] - 4; y <= bounds[2]+bounds[3]+4; y += 3 {
						bx, bz := bounds[0]*16+x, bounds[1]*16+z
						a, b, c := zoom.quartAtBlock(bx, y, bz)
						d, e, f := BiomeQuartAtBlock(seed, bx, y, bz)
						if a != d || b != e || c != f {
							t.Fatalf("seed %d at %d,%d,%d: %v/%v", seed, bx, y, bz, [3]int32{a, b, c}, [3]int32{d, e, f})
						}
					}
				}
			}
		}
	}
}

func BenchmarkSurfaceBiomeZoom(b *testing.B) {
	for _, cached := range []bool{false, true} {
		b.Run(fmt.Sprintf("cached=%t", cached), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				var zoom biomeZoom
				if cached {
					zoom = newChunkBiomeZoom(1234, -17, 23, -64, 384)
				}
				for x := int32(0); x < 16; x++ {
					for z := int32(0); z < 16; z++ {
						for y := int32(160); y >= -64; y-- {
							if cached {
								zoom.quartAtBlock(-17*16+x, y, 23*16+z)
							} else {
								BiomeQuartAtBlock(1234, -17*16+x, y, 23*16+z)
							}
						}
					}
				}
			}
		})
	}
}
