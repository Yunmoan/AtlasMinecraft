package worldgen

import (
	"fmt"
	"math"
	"testing"
)

func TestDensityCellFillMatchesScalar(t *testing.T) {
	for _, dimension := range []string{"overworld", "nether", "end"} {
		settings, err := LoadDensitySettings(dimension)
		if err != nil {
			t.Fatal(err)
		}
		for _, seed := range []int64{0, -987654321} {
			fast := NewNoiseChunkDensitySampler(seed, settings, -17, 23)
			scalar := NewNoiseChunkDensitySampler(seed, settings, -17, 23)
			filler := newDensityCellFiller(fast, settings.Router["final_density"])
			for _, x := range []int32{-272, -260} {
				// Include the build floor, sea level and upper gradient. Align
				// even when the End's horizontal cell width differs.
				x = floorCell(x, settings.CellWidth)
				for _, y := range []int32{settings.MinY, 0, 64, settings.MinY + settings.Height - settings.CellHeight} {
					y = floorCell(y, settings.CellHeight)
					origin := densityPosition{x, y, 368}
					values, err := filler.fill(origin)
					if err != nil {
						t.Fatal(err)
					}
					if fast.chunk.fillingCell {
						t.Fatal("cell fill leaked its context")
					}
					scalar.chunk.fillingCell = true
					i := 0
					for dy := settings.CellHeight - 1; dy >= 0; dy-- {
						for dx := int32(0); dx < settings.CellWidth; dx++ {
							for dz := int32(0); dz < settings.CellWidth; dz++ {
								want, err := scalar.Sample(settings.Router["final_density"], x+dx, y+dy, 368+dz)
								if err != nil || math.Float64bits(values[i]) != math.Float64bits(want) {
									t.Fatalf("%s seed %d origin %v index %d: %x/%x (%v)", dimension, seed, origin, i, math.Float64bits(values[i]), math.Float64bits(want), err)
								}
								i++
							}
						}
					}
				}
			}
		}
	}
}

func TestDensityCellFillPreservesBranchErrorsAndContext(t *testing.T) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	s := NewNoiseChunkDensitySampler(0, settings, 0, 0)
	constant := &DensityNode{Kind: "constant", Numbers: map[string]float64{"value": 2}}
	node := &DensityNode{Kind: "range_choice", Numbers: map[string]float64{"min_inclusive": 1, "max_exclusive": 3}, Children: map[string]*DensityNode{"input": constant, "when_in_range": constant}}
	filler := newDensityCellFiller(s, node)
	values, err := filler.fill(densityPosition{})
	if err != nil || values[0] != 2 {
		t.Fatalf("unselected invalid branch: %v/%v", values, err)
	}
	node.Children["when_in_range"] = nil
	filler = newDensityCellFiller(s, node)
	if _, err := filler.fill(densityPosition{}); err == nil || s.chunk.fillingCell {
		t.Fatalf("error/context: %v/%t", err, s.chunk.fillingCell)
	}
}

func TestVanillaDensityMultiplyShortCircuit(t *testing.T) {
	for _, test := range []struct {
		name, expression string
		bits             uint64
		failure          bool
	}{
		{"zero", `{"type":"minecraft:mul","argument1":{"type":"minecraft:half_negative","argument":0},"argument2":{"type":"minecraft:y_clamped_gradient","from_y":0,"to_y":0,"from_value":0,"to_value":1}}`, 0, false},
		{"negative_zero", `{"type":"minecraft:mul","argument1":{"type":"minecraft:half_negative","argument":-0.0},"argument2":{"type":"minecraft:y_clamped_gradient","from_y":0,"to_y":1,"from_value":2,"to_value":2}}`, 0, false},
		{"constant_operand", `{"type":"minecraft:mul","argument1":{"type":"minecraft:half_negative","argument":-0.0},"argument2":2}`, 1 << 63, false},
		{"constant_zero", `{"type":"minecraft:mul","argument1":0,"argument2":{"type":"minecraft:y_clamped_gradient","from_y":0,"to_y":0,"from_value":0,"to_value":1}}`, 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			compiler, err := newDensityCompiler()
			if err != nil {
				t.Fatal(err)
			}
			node, err := compiler.parse([]byte(test.expression), "short_circuit_probe")
			if err != nil {
				t.Fatal(err)
			}
			s := NewDensitySampler(0)
			s.program = compiler.program
			value, err := s.Sample(node, 0, 0, 0)
			if (err != nil) != test.failure || (!test.failure && math.Float64bits(value) != test.bits) {
				t.Fatalf("got %x/%v, want %x/failure=%t", math.Float64bits(value), err, test.bits, test.failure)
			}
			if test.name == "zero" {
				if _, err := NewDensitySampler(0).Sample(node, 0, 0, 0); err == nil {
					t.Fatal("unbound probe unexpectedly skipped its error")
				}
			}
		})
	}
}

func BenchmarkDensityCellFill(b *testing.B) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		b.Fatal(err)
	}
	for _, batch := range []bool{false, true} {
		b.Run(fmt.Sprintf("batch=%t", batch), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				s := NewNoiseChunkDensitySampler(1234, settings, 40, -40)
				var filler *densityCellFiller
				if batch {
					filler = newDensityCellFiller(s, settings.Router["final_density"])
				}
				for x := int32(0); x < 16; x += settings.CellWidth {
					for z := int32(0); z < 16; z += settings.CellWidth {
						for y := settings.MinY + settings.Height - settings.CellHeight; y >= settings.MinY; y -= settings.CellHeight {
							if batch {
								if _, err := filler.fill(densityPosition{640 + x, y, -640 + z}); err != nil {
									b.Fatal(err)
								}
								continue
							}
							s.chunk.fillingCell = true
							for dy := settings.CellHeight - 1; dy >= 0; dy-- {
								for dx := int32(0); dx < settings.CellWidth; dx++ {
									for dz := int32(0); dz < settings.CellWidth; dz++ {
										if _, err := s.Sample(settings.Router["final_density"], 640+x+dx, y+dy, -640+z+dz); err != nil {
											b.Fatal(err)
										}
									}
								}
							}
						}
					}
				}
			}
		})
	}
}

// Includes final density, aquifers, ore veins and block/postprocessing arrays;
// deliberately excludes surface, carvers, decoration, lighting and persistence.
func BenchmarkNoiseChunkGeneration(b *testing.B) {
	noise, err := NewOverworldNoiseGenerator(1234)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := noise.GenerateNoise(40, -40); err != nil {
			b.Fatal(err)
		}
	}
}
