package worldgen

import "testing"

func BenchmarkStrongholdPlacement(b *testing.B) {
	b.Run("near_spawn", func(b *testing.B) {
		cache, err := newOverworldStructureCache(1234)
		if err != nil {
			b.Fatal(err)
		}
		g := &OverworldGenerator{structures: cache}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if found, err := g.strongholdPlacementAt(int32(i%32)-16, int32(i%31)-15); err != nil || found {
				b.Fatal(found, err)
			}
		}
	})
	b.Run("cold_relocated_rings", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache, err := newOverworldStructureCache(1234)
			if err != nil {
				b.Fatal(err)
			}
			g := &OverworldGenerator{structures: cache}
			if found, err := g.strongholdPlacementAt(-100, -125); err != nil || !found {
				b.Fatal(found, err)
			}
		}
	})
	b.Run("layout", func(b *testing.B) {
		sources := StrongholdStartSources{MinY: -64, SeaLevel: 63, BiomeAtQuart: func(int32, int32, int32) (string, error) { return "minecraft:plains", nil }}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if start, err := GenerateStrongholdStructureStart(1234, -100, -125, sources); err != nil || start == nil || len(start.Pieces) != 129 {
				b.Fatal("pinned original layout changed", err)
			}
		}
	})
}

// One operation creates a real normal chunk. The structures=false branch
// measures the same location without any structures, not a v11 baseline.
func BenchmarkOverworldStronghold(b *testing.B) {
	for _, location := range []struct {
		name string
		x, z int32
	}{{"owner", -100, -125}, {"portal_room", -101, -126}} {
		b.Run(location.name, func(b *testing.B) {
			for _, enabled := range []bool{false, true} {
				name := "cold_with_structures"
				if !enabled {
					name = "cold_without_structures"
				}
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: enabled})
						if err != nil {
							b.Fatal(err)
						}
						g.SetTerrainWorkers(1)
						if _, err := g.Generate(location.x, location.z); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
			b.Run("cached", func(b *testing.B) {
				g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
				if err != nil {
					b.Fatal(err)
				}
				g.SetTerrainWorkers(1)
				if _, err := g.Generate(location.x, location.z); err != nil {
					b.Fatal(err)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := g.Generate(location.x, location.z); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
