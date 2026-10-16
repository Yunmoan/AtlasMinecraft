package worldgen

import "testing"

func BenchmarkDimensionEndCity(b *testing.B) {
	for _, entry := range []struct {
		name string
		x, z int32
	}{{"owner", 145, -77}, {"ship", 138, -78}} {
		b.Run(entry.name, func(b *testing.B) {
			warm, err := NewDimensionGeneratorWithOptions(1234, "end", DimensionGeneratorOptions{GenerateStructures: true})
			if err != nil {
				b.Fatal(err)
			}
			if _, err := warm.Generate(entry.x, entry.z); err != nil {
				b.Fatal(err)
			}
			b.Run("cached", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					if _, err := warm.Generate(entry.x, entry.z); err != nil {
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
						g, err := NewDimensionGeneratorWithOptions(1234, "end", DimensionGeneratorOptions{GenerateStructures: enabled})
						if err != nil {
							b.Fatal(err)
						}
						if _, err := g.Generate(entry.x, entry.z); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}
