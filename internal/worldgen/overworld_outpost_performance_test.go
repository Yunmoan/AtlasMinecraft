package worldgen

import "testing"

func BenchmarkOverworldOutpost(b *testing.B) {
	for _, location := range []struct {
		name string
		x, z int32
	}{{"owner", -249, 327}, {"allay_cage", -250, 328}} {
		b.Run(location.name, func(b *testing.B) {
			warm, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
			if err != nil {
				b.Fatal(err)
			}
			if _, err := warm.Generate(location.x, location.z); err != nil {
				b.Fatal(err)
			}
			b.Run("cached", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					if _, err := warm.Generate(location.x, location.z); err != nil {
						b.Fatal(err)
					}
				}
			})
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
						if _, err := g.Generate(location.x, location.z); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}
