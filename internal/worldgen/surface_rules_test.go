package worldgen_test

import (
	"atlas/internal/world"
	"atlas/internal/worldgen"
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestSurfaceRulesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "surface_rules_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	settings, err := worldgen.LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	samplers := make(map[int64]*worldgen.SurfaceRuleSampler)
	scanner := bufio.NewScanner(file)
	rules, bands := 0, 0
	mismatches := 0
	for scanner.Scan() {
		var label string
		var seed int64
		var context worldgen.SurfaceContext
		if _, err := fmt.Sscan(scanner.Text(), &label, &seed, &context.X, &context.Y, &context.Z); err != nil {
			t.Fatal(err)
		}
		sampler := samplers[seed]
		if sampler == nil {
			sampler, err = worldgen.NewSurfaceRuleSampler(seed, settings)
			if err != nil {
				t.Fatal(err)
			}
			samplers[seed] = sampler
		}
		var expected int64
		var state *worldgen.SurfaceBlockState
		if label == "band" {
			if _, err := fmt.Sscan(scanner.Text(), &label, &seed, &context.X, &context.Y, &context.Z, &expected); err != nil {
				t.Fatal(err)
			}
			state, err = sampler.BandAt(context.X, context.Y, context.Z)
			bands++
		} else if label == "rule" {
			var secondaryBits uint64
			if _, err := fmt.Sscan(scanner.Text(), &label, &seed, &context.X, &context.Y, &context.Z, &context.Biome,
				&context.SurfaceDepth, &secondaryBits, &context.StoneDepthAbove, &context.StoneDepthBelow,
				&context.WaterHeight, &context.MinSurfaceLevel, &context.Cold, &context.Steep, &expected); err != nil {
				t.Fatal(err)
			}
			context.SecondaryDepth = math.Float64frombits(secondaryBits)
			depth, err := sampler.SurfaceDepth(context.X, context.Z)
			if err != nil || depth != context.SurfaceDepth {
				t.Fatalf("surface depth seed %d %d,%d: %d, want %d (%v)", seed, context.X, context.Z, depth, context.SurfaceDepth, err)
			}
			secondary, err := sampler.SecondaryDepth(context.X, context.Z)
			if err != nil || math.Float64bits(secondary) != secondaryBits {
				t.Fatalf("secondary depth seed %d %d,%d: bits %d, want %d (%v)", seed, context.X, context.Z, math.Float64bits(secondary), secondaryBits, err)
			}
			state, err = sampler.Apply(context)
			rules++
		} else {
			t.Fatalf("unknown reference label %q", label)
		}
		if err != nil {
			t.Fatal(err)
		}
		got := int64(-1)
		if state != nil {
			id, err := world.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
			got = int64(id)
		}
		if got != expected {
			if mismatches < 15 {
				t.Errorf("%s seed %d context %+v: state %d, want %d", label, seed, context, got, expected)
			}
			mismatches++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if mismatches != 0 {
		t.Errorf("%d surface reference mismatches", mismatches)
	}
	if rules != 14592 || bands != 4608 {
		t.Fatalf("reference count: %d rules, %d bands; want 14592 and 4608", rules, bands)
	}
}

func TestSurfaceRuleSamplerAcceptsLegacySettings(t *testing.T) {
	settings, err := worldgen.LoadDensitySettings("nether")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worldgen.NewSurfaceRuleSampler(1, settings); err != nil {
		t.Fatal(err)
	}
}
