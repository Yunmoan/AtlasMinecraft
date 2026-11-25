package worldgen

import (
	"errors"
	"fmt"
	"runtime"
	"testing"
)

func TestVanillaAquiferStatusArrayMatchesCoordinateMap(t *testing.T) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	for _, seed := range []int64{1234, -987654321} {
		for _, pos := range [][2]int32{{-17, 23}, {40, -40}} {
			random := NewXoroshiro(seed).ForkPositional().FromHashOf("minecraft:aquifer").ForkPositional()
			fast := newNoiseAquifer(NewNoiseChunkDensitySampler(seed, settings, pos[0], pos[1]), settings, random)
			old := newNoiseAquifer(NewNoiseChunkDensitySampler(seed, settings, pos[0], pos[1]), settings, random)
			old.centerGrid = nil
			old.gridHeight = 0
			for pass := 0; pass < 2; pass++ {
				for _, y := range []int32{-64, -53, -10, 0, 64, 319} {
					for x := int32(0); x < 16; x += 3 {
						for z := int32(0); z < 16; z += 3 {
							for _, density := range []float64{-0.2, 0, 0.2} {
								a, solidA, tickA := fast.compute(pos[0]*16+x, y, pos[1]*16+z, density)
								b, solidB, tickB := old.compute(pos[0]*16+x, y, pos[1]*16+z, density)
								if a != b || solidA != solidB || tickA != tickB || fast.err != nil || old.err != nil {
									t.Fatalf("aquifer %d %v at %d,%d,%d: %v/%v", seed, pos, x, y, z, []any{a, solidA, tickA, fast.err}, []any{b, solidB, tickB, old.err})
								}
							}
						}
					}
				}
			}
			// Same grid cell, different position: it must not alias a center slot.
			grid := densityPosition{pos[0], 0, pos[1]}
			center := fast.centerAt(grid)
			if center != old.centerAt(grid) {
				t.Fatal("center changed")
			}
			for _, p := range []densityPosition{center, {center.x + 1, center.y, center.z}, {center.x - 64, center.y, center.z + 64}} {
				if a, b := fast.fluidAt(p), old.fluidAt(p); a != b {
					t.Fatalf("fluid status %v: %v/%v", p, a, b)
				}
			}
		}
	}
}

func TestVanillaAquiferFailedStatusIsNotCached(t *testing.T) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	random := NewXoroshiro(1234).ForkPositional().FromHashOf("minecraft:aquifer").ForkPositional()
	a := newNoiseAquifer(NewNoiseChunkDensitySampler(1234, settings, 0, 0), settings, random)
	center := a.centerAt(a.gridMin)
	a.err = errors.New("probe failure")
	a.fluidAt(center)
	if a.centerGrid[0].fluidValid {
		t.Fatal("failed fluid calculation was cached")
	}
	a.err = nil
	fresh := newNoiseAquifer(NewNoiseChunkDensitySampler(1234, settings, 0, 0), settings, random)
	if got, want := a.fluidAt(center), fresh.fluidAt(center); got != want || a.err != nil || fresh.err != nil {
		t.Fatalf("retry returned %v/%v (%v/%v)", got, want, a.err, fresh.err)
	}
}

// Preserve the original coordinate-map path as an independent lookup oracle.
func mappedAquiferStatus(a *noiseAquifer, p densityPosition) aquiferFluid {
	if value, ok := a.fluids[p]; ok {
		return value
	}
	value := a.calculateFluid(p.x, p.y, p.z)
	a.fluids[p] = value
	return value
}

func BenchmarkVanillaAquiferStatusLookup(b *testing.B) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		b.Fatal(err)
	}
	for _, array := range []bool{false, true} {
		b.Run(fmt.Sprintf("array=%t", array), func(b *testing.B) {
			random := NewXoroshiro(1234).ForkPositional().FromHashOf("minecraft:aquifer").ForkPositional()
			a := newNoiseAquifer(NewNoiseChunkDensitySampler(1234, settings, -17, 23), settings, random)
			lookup := func(p densityPosition) aquiferFluid { return mappedAquiferStatus(a, p) }
			if array {
				lookup = a.fluidAt
			}
			positions := make([]densityPosition, 0, 27)
			for x := int32(-18); x <= -16; x++ {
				for z := int32(22); z <= 24; z++ {
					for _, y := range []int32{-2, 0, 3} {
						p := a.centerAt(densityPosition{x, y, z})
						positions = append(positions, p)
						lookup(p)
					}
				}
			}
			if a.err != nil {
				b.Fatal(a.err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			var result aquiferFluid
			for n := 0; n < b.N; n++ {
				result = lookup(positions[n%len(positions)])
			}
			b.StopTimer()
			runtime.KeepAlive(result)
		})
	}
}
