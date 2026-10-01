package worldgen

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestOverworldFeatureScheduleAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("feature_order_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	groups := map[string][]string{
		"plains":                 {"minecraft:plains"},
		"desert":                 {"minecraft:desert"},
		"plains_desert":          {"minecraft:plains", "minecraft:desert"},
		"frozen_ocean_deep_dark": {"minecraft:frozen_ocean", "minecraft:deep_dark"},
		"savanna_jungle_meadow":  {"minecraft:savanna", "minecraft:jungle", "minecraft:meadow"},
	}
	expected := make(map[string][]PlacedFeature)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || fields[0] != "schedule" {
			continue
		}
		if len(fields) != 5 {
			t.Fatalf("invalid JAR schedule vector: %q", scanner.Text())
		}
		step, stepErr := strconv.Atoi(fields[2])
		index, indexErr := strconv.Atoi(fields[3])
		if stepErr != nil || indexErr != nil {
			t.Fatalf("invalid JAR schedule index: %q", scanner.Text())
		}
		expected[fields[1]] = append(expected[fields[1]], PlacedFeature{step, index, fields[4]})
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(expected) != len(groups) {
		t.Fatalf("got %d JAR biome groups, want %d", len(expected), len(groups))
	}
	for group, biomes := range groups {
		t.Run(group, func(t *testing.T) {
			got, err := OverworldFeatureSchedule(biomes)
			if err != nil {
				t.Fatal(err)
			}
			want := expected[group]
			if len(got) != len(want) {
				t.Fatalf("got %d placed features, want %d", len(got), len(want))
			}
			for index := range want {
				if got[index] != want[index] {
					t.Fatalf("feature %d: got %+v, want %+v", index, got[index], want[index])
				}
			}
		})
	}
}

func TestOverworldFeatureScheduleRejectsOtherDimensions(t *testing.T) {
	if _, err := OverworldFeatureSchedule([]string{"minecraft:nether_wastes"}); err == nil {
		t.Fatal("Nether biome should not use the overworld feature order")
	}
	if _, err := OverworldFeatureSchedule([]string{"minecraft:unknown"}); err == nil {
		t.Fatal("unknown biome should not be accepted")
	}
	first, err := OverworldFeatureSchedule([]string{"minecraft:plains"})
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := OverworldFeatureSchedule([]string{"minecraft:plains", "minecraft:plains"})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(repeated) {
		t.Fatal("repeated biome changed the feature schedule")
	}
}

func TestOverworldFeatureScheduleCoversBiomeSource(t *testing.T) {
	points, err := OverworldBiomePoints()
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, point := range points {
		if seen[point.Biome] {
			continue
		}
		seen[point.Biome] = true
		if _, err := OverworldFeatureSchedule([]string{point.Biome}); err != nil {
			t.Fatalf("biome %s: %v", point.Biome, err)
		}
	}
	if len(seen) < 50 {
		t.Fatalf("only %d overworld biomes checked", len(seen))
	}
}

func TestOverworldFeatureSeedsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/feature_seed_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 5 && len(fields) != 7 && len(fields) != 9 {
			t.Fatalf("invalid JAR seed vector: %q", scanner.Text())
		}
		values := make([]int64, len(fields)-1)
		for index := range values {
			values[index], err = strconv.ParseInt(fields[index+1], 10, 64)
			if err != nil {
				t.Fatalf("invalid JAR seed vector: %q", scanner.Text())
			}
		}
		decoration := OverworldDecorationSeed(values[0], int32(values[1]), int32(values[2]))
		switch fields[0] {
		case "decoration":
			if decoration != values[3] {
				t.Fatalf("decoration seed: got %d, want %d for %q", decoration, values[3], scanner.Text())
			}
		case "feature_seed":
			feature := PlacedFeature{Step: int(values[3]), Index: int(values[4])}
			got := int64(OverworldPlacedFeatureRandom(decoration, feature).NextLong())
			if got != values[5] {
				t.Fatalf("feature random: got %d, want %d for %q", got, values[5], scanner.Text())
			}
		case "feature_random":
			feature := PlacedFeature{Step: int(values[3]), Index: int(values[4])}
			random := OverworldPlacedFeatureRandom(decoration, feature)
			bound := random.NextIntBound(37)
			floatBits := math.Float32bits(random.NextFloat())
			doubleBits := math.Float64bits(random.NextDouble())
			if int64(bound) != values[5] || int64(floatBits) != values[6] || int64(doubleBits) != values[7] {
				t.Fatalf("feature random methods mismatch for %q", scanner.Text())
			}
		default:
			t.Fatalf("unknown JAR seed vector: %q", scanner.Text())
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 180 {
		t.Fatalf("got %d JAR seed vectors, want 180", count)
	}
}
