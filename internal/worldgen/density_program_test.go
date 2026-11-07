package worldgen

import (
	"fmt"
	"math"
	"testing"
)

func TestDensityMissingInterpolationArgumentReturnsError(t *testing.T) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []*DensitySampler{NewCellDensitySampler(1, settings), NewNoiseChunkDensitySampler(1, settings, 0, 0)} {
		_, err := s.Sample(&DensityNode{Kind: "interpolated"}, 5, 63, 9)
		if err == nil || err.Error() != "nil density node" {
			t.Fatalf("missing interpolation argument: %v", err)
		}
	}
}

func TestDensityRuntimeSlotsMatchMapState(t *testing.T) {
	for _, dimension := range []string{"overworld", "nether", "end"} {
		settings, err := LoadDensitySettings(dimension)
		if err != nil {
			t.Fatal(err)
		}
		foreign, err := LoadDensitySettings("overworld")
		if err != nil {
			t.Fatal(err)
		}
		fast := NewNoiseChunkDensitySampler(-987654321, settings, -17, 23)
		maps := NewNoiseChunkDensitySampler(-987654321, settings, -17, 23)
		maps.runtime = nil
		// Foreign and synthetic nodes cannot alias a bound program's slots.
		roots := []*DensityNode{settings.Router["final_density"], settings.Router["vein_toggle"], foreign.Router["temperature"],
			{Kind: "noise", Noise: "minecraft:temperature", Numbers: map[string]float64{"xz_scale": 0.25, "y_scale": 0.25}}, settings.Router["final_density"]}
		for _, filling := range []bool{false, true} {
			fast.chunk.fillingCell, maps.chunk.fillingCell = filling, filling
			for _, node := range roots {
				for x := int32(-4); x <= 20; x += 3 {
					for z := int32(20); z >= -4; z -= 3 {
						for y := settings.MinY; y <= settings.MinY+settings.Height; y += 11 {
							a, errA := fast.Sample(node, -17*16+x, y, 23*16+z)
							b, errB := maps.Sample(node, -17*16+x, y, 23*16+z)
							if errA != nil || errB != nil || math.Float64bits(a) != math.Float64bits(b) {
								t.Fatalf("%s filling=%t at %d,%d,%d: %x/%x (%v/%v)", dimension, filling, x, y, z, math.Float64bits(a), math.Float64bits(b), errA, errB)
							}
						}
					}
				}
			}
		}
	}
}

func TestPowerOfTwoCellFloorMatchesDivision(t *testing.T) {
	for _, size := range []int32{1, 2, 3, 4, 7, 8, 12, 16, 32, 1 << 30} {
		for _, v := range []int32{math.MinInt32, -30000001, -19, -8, -1, 0, 1, 19, 30000001, math.MaxInt32} {
			q := v / size
			if v%size < 0 {
				q--
			}
			if got := floorCell(v, size); got != q*size {
				t.Fatalf("floor(%d,%d)=%d, want %d", v, size, got, q*size)
			}
		}
	}
}

func BenchmarkDensityRuntimeSlots(b *testing.B) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		b.Fatal(err)
	}
	for _, maps := range []bool{false, true} {
		b.Run(fmt.Sprintf("maps=%t", maps), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				s := NewNoiseChunkDensitySampler(1234, settings, -17, 23)
				if maps {
					s.runtime = nil
				}
				s.chunk.fillingCell = true
				for x := int32(0); x < 16; x++ {
					for z := int32(0); z < 16; z++ {
						for y := int32(-64); y < 96; y++ {
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
