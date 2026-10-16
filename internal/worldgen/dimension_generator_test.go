package worldgen

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestDimensionGeneratorCoverageAndReplay(t *testing.T) {
	for _, dimension := range []string{"nether", "end"} {
		t.Run(dimension, func(t *testing.T) {
			g, err := NewDimensionGenerator(1234, dimension)
			if err != nil {
				t.Fatal(err)
			}
			for _, feature := range g.order.features {
				if _, supported := g.plans[feature.Name]; !supported {
					t.Log("unsupported: " + feature.Name)
				}
			}
			cx, cz := int32(0), int32(0)
			if dimension == "end" {
				cx, cz = 80, -70
			}
			first, err := g.Generate(cx, cz)
			if err != nil {
				t.Fatal(err)
			}
			if first.Terrain.MinY != 0 || first.Terrain.Height != 256 || len(first.Biomes) != 16 {
				t.Fatal("dimension build height/palettes incorrect")
			}
			for _, section := range first.Biomes {
				for _, biome := range section {
					if !g.order.allowed[biome] {
						t.Fatalf("invalid biome %q", biome)
					}
				}
			}
			if _, err := g.Generate(cx+1, cz); err != nil {
				t.Fatal(err)
			}
			second, err := g.Generate(cx, cz)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(first, second) {
				t.Fatal("neighbor exploration changed chunk replay")
			}
			first.Terrain.Blocks[0] = NoiseWater
			first.Biomes[0][0] = "minecraft:plains"
			third, err := g.Generate(cx, cz)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(second, third) {
				t.Fatal("caller mutation changed cached terrain or biomes")
			}
		})
	}
}

func TestDimensionGeneratorConcurrentReplay(t *testing.T) {
	for _, dimension := range []string{"nether", "end"} {
		t.Run(dimension, func(t *testing.T) {
			g, err := NewDimensionGenerator(1234, dimension)
			if err != nil {
				t.Fatal(err)
			}
			var want [3]*GeneratedDimensionChunk
			g.SetTerrainWorkers(1)
			for i := range want {
				want[i], err = g.Generate(int32(i-1), 0)
				if err != nil {
					t.Fatal(err)
				}
			}
			shared, err := NewDimensionGenerator(1234, dimension)
			if err != nil {
				t.Fatal(err)
			}
			shared.SetTerrainWorkers(4)
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					index := i % len(want)
					got, err := shared.Generate(int32(index-1), 0)
					if err != nil {
						t.Error(err)
						return
					}
					if !reflect.DeepEqual(got, want[index]) {
						t.Errorf("worker %d output differs", i)
					}
				}(i)
			}
			wg.Wait()
		})
	}
}

func BenchmarkDimensionGeneration(b *testing.B) {
	for _, dimension := range []string{"nether", "end"} {
		b.Run(dimension, func(b *testing.B) {
			g, err := NewDimensionGenerator(1234, dimension)
			if err != nil {
				b.Fatal(err)
			}
			if _, err := g.Generate(0, 0); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := g.Generate(0, 0); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// One operation creates a new generator and its first complete chunk, including
// cold terrain halos. The existing benchmark measures only warm replay.
func BenchmarkDimensionColdGeneration(b *testing.B) {
	for _, dimension := range []string{"nether", "end"} {
		for _, workers := range []int{1, 4} {
			b.Run(fmt.Sprintf("%s/workers=%d", dimension, workers), func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					g, err := NewDimensionGenerator(1234, dimension)
					if err != nil {
						b.Fatal(err)
					}
					g.SetTerrainWorkers(workers)
					cx, cz := int32(0), int32(0)
					if dimension == "end" {
						cx, cz = 80, -70
					}
					if _, err := g.Generate(cx, cz); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
