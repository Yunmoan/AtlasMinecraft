package worldgen

import (
	"fmt"
	"math"
	"testing"
)

func TestCornerCacheFallbackRetainsProbeSemantics(t *testing.T) {
	for _, settings := range []*DensitySettings{
		{CellWidth: 3, CellHeight: 5, MinY: -7, Height: 15},
		{CellWidth: 0, CellHeight: 0},
		{CellWidth: 1, CellHeight: 1, Height: 1000000},
	} {
		s := NewNoiseChunkDensitySampler(0, settings, 0, 0)
		if s.chunk.denseCorners != nil {
			t.Fatal("unsupported probe allocated a dense lattice")
		}
		node := &DensityNode{Kind: "constant", Numbers: map[string]float64{"value": 2.5}}
		for pass := 0; pass < 2; pass++ {
			got, err := s.sampleCorner(node, -1, -7, 20)
			if err != nil || got != 2.5 {
				t.Fatalf("fallback: %v, %v", got, err)
			}
		}
	}
}

func TestDenseCornerCacheMatchesCoordinateMap(t *testing.T) {
	for _, dimension := range []string{"overworld", "nether", "end"} {
		settings, err := LoadDensitySettings(dimension)
		if err != nil {
			t.Fatal(err)
		}
		for _, seed := range []int64{0, 1234, -987654321} {
			fast := NewNoiseChunkDensitySampler(seed, settings, -17, 23)
			direct := NewNoiseChunkDensitySampler(seed, settings, -17, 23)
			direct.chunk.denseCorners = nil
			// Repeated traversal revisits corners, boundaries and the fallback
			// beyond the current chunk. Compare bits before quantization.
			for pass := 0; pass < 2; pass++ {
				for x := int32(-4); x <= 20; x += 3 {
					for z := int32(20); z >= -4; z -= 3 {
						for y := settings.MinY - 1; y <= settings.MinY+settings.Height; y += 17 {
							a, errA := fast.Sample(settings.Router["final_density"], -17*16+x, y, 23*16+z)
							b, errB := direct.Sample(settings.Router["final_density"], -17*16+x, y, 23*16+z)
							if errA != nil || errB != nil || math.Float64bits(a) != math.Float64bits(b) {
								t.Fatalf("%s seed %d at %d,%d,%d: %x/%x (%v/%v)", dimension, seed, x, y, z, math.Float64bits(a), math.Float64bits(b), errA, errB)
							}
						}
					}
				}
			}
		}
	}
}

func BenchmarkNoiseCornerCache(b *testing.B) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		b.Fatal(err)
	}
	for _, coordinateMap := range []bool{false, true} {
		b.Run(fmt.Sprintf("coordinate_map=%t", coordinateMap), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				s := NewNoiseChunkDensitySampler(1234, settings, -17, 23)
				if coordinateMap {
					s.chunk.denseCorners = nil
				}
				for x := int32(0); x < 16; x += 4 {
					for z := int32(0); z < 16; z += 4 {
						for y := settings.MinY; y < settings.MinY+settings.Height; y++ {
							if _, err := s.Sample(settings.Router["final_density"], -17*16+x, y, 23*16+z); err != nil {
								b.Fatal(err)
							}
						}
					}
				}
			}
		})
	}
}
