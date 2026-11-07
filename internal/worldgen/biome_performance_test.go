package worldgen

import (
	"fmt"
	"math"
	"testing"
)

// Keep the direct sampler as an oracle: it evaluates every density expression
// afresh, including the horizontal fields at every height. Also compare the
// selected biome so the previous-point tie preference stays in the loop.
func TestBiomeVerticalTraversalMatchesDirect(t *testing.T) {
	for _, seed := range []int64{0, 1234, -918273645} {
		fast, err := NewOverworldBiomeSampler(seed)
		if err != nil {
			t.Fatal(err)
		}
		direct, err := NewOverworldBiomeSampler(seed)
		if err != nil {
			t.Fatal(err)
		}
		direct.density = NewDensitySampler(seed)
		for _, column := range [][2]int32{{0, 0}, {-17, 23}, {29999984, -29999984}, {-17, 23}} {
			for y := int32(320); y >= -80; y -= 4 {
				a, errA := fast.ClimateAtBlock(column[0], y, column[1])
				b, errB := direct.ClimateAtBlock(column[0], y, column[1])
				if errA != nil || errB != nil || a != b {
					t.Fatalf("seed %d at %v,%d: %v/%v (%v/%v)", seed, column, y, a, b, errA, errB)
				}
				got, errA := fast.BiomeAtQuart(column[0]>>2, y>>2, column[1]>>2)
				want, errB := direct.BiomeAtQuart(column[0]>>2, y>>2, column[1]>>2)
				if errA != nil || errB != nil || got != want {
					t.Fatalf("biome seed %d at %v,%d: %s/%s (%v/%v)", seed, column, y, got, want, errA, errB)
				}
			}
		}
	}
}

func TestBiomeHorizontalCacheDensityBitsAndEviction(t *testing.T) {
	s, err := NewOverworldBiomeSampler(-7122209284156038298)
	if err != nil {
		t.Fatal(err)
	}
	direct := NewDensitySampler(s.density.Seed)
	for x := int32(-500); x < 500; x++ {
		for _, y := range []int32{-64, 63, 319, -64} {
			for _, field := range climateFields {
				node := s.settings.Router[field]
				a, errA := s.density.Sample(node, x*4, y, -x*8)
				b, errB := direct.Sample(node, x*4, y, -x*8)
				if errA != nil || errB != nil || math.Float64bits(a) != math.Float64bits(b) {
					t.Fatalf("%s at %d,%d: %x/%x (%v/%v)", field, x, y, math.Float64bits(a), math.Float64bits(b), errA, errB)
				}
			}
		}
		if len(s.density.horizontal) > 4096 {
			t.Fatal("unbounded horizontal cache")
		}
	}
}

func BenchmarkBiomeVerticalTraversal(b *testing.B) {
	for _, direct := range []bool{false, true} {
		b.Run(fmt.Sprintf("direct=%t", direct), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				s, err := NewOverworldBiomeSampler(1234)
				if err != nil {
					b.Fatal(err)
				}
				if direct {
					s.density = NewDensitySampler(1234)
				}
				for x := int32(-1); x <= 4; x++ {
					for z := int32(-1); z <= 4; z++ {
						for y := int32(-16); y < 80; y++ {
							if _, err := s.BiomeAtQuart(x, y, z); err != nil {
								b.Fatal(err)
							}
						}
					}
				}
			}
		})
	}
}
