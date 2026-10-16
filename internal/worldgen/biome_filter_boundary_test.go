package worldgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// The oracle invokes the actual 1.21.1 BiomeFilter through LevelReader's
// default getBiome and BiomeManager, using nonuniform quart biome sources.
// Uniform-biome feature fixtures cannot detect sampling the wrong quart.
func TestBiomeFilterBoundariesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/biome_filter_boundaries_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	parse := func(text string) int64 {
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	type binding struct {
		template *placementTemplate
		order    *featureOrder
	}
	bindings := make(map[string]binding)
	groups := make(map[string]int)
	checked, quartDifferences, biomeDifferences, mismatches := 0, 0, 0, 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 16 || fields[0] != "biome_filter" {
			t.Fatalf("invalid biome filter vector: %s", scanner.Text())
		}
		dimension, name := fields[1], fields[2]
		seed, mode := parse(fields[3]), parse(fields[4])
		anchorX, anchorZ := int32(parse(fields[5]))>>2, int32(parse(fields[6]))>>2
		position := FeaturePosition{int32(parse(fields[7])), int32(parse(fields[8])), int32(parse(fields[9]))}
		wantQuart := FeaturePosition{int32(parse(fields[10])), int32(parse(fields[11])), int32(parse(fields[12]))}
		wantBiome := fields[13]
		wantAllowed, err := strconv.ParseBool(fields[14])
		if err != nil {
			t.Fatal(err)
		}
		wantNext := parse(fields[15])
		key := dimension + "/" + name
		bound, exists := bindings[key]
		if !exists {
			var order *featureOrder
			if dimension == "overworld" {
				order, err = loadOverworldFeatureOrder()
			} else {
				order, err = loadDimensionFeatureOrder(dimension)
			}
			if err != nil {
				t.Fatal(err)
			}
			var feature *PlacedFeature
			for _, candidate := range order.features {
				if candidate.Name == name {
					copy := candidate
					feature = &copy
					break
				}
			}
			if feature == nil {
				t.Fatalf("feature %s missing from %s schedule", name, dimension)
			}
			definition := PlacedFeatureDefinition{Placement: []json.RawMessage{json.RawMessage(`{"type":"minecraft:biome"}`)}}
			template, err := compilePlacementTemplate(definition, feature)
			if err != nil {
				t.Fatal(err)
			}
			bound = binding{template, order}
			bindings[key] = bound
		}
		var pair [2]string
		switch dimension {
		case "overworld":
			pair = [2]string{"minecraft:river", "minecraft:ocean"}
		case "nether":
			pair = [2]string{"minecraft:nether_wastes", "minecraft:crimson_forest"}
		case "end":
			pair = [2]string{"minecraft:end_highlands", "minecraft:small_end_islands"}
		default:
			t.Fatalf("unexpected dimension %s", dimension)
		}
		// This synthetic noise source defines boundaries only. The expected
		// selected quart and admission decision both come from the Java oracle.
		biomeForQuart := func(qx, qy, qz int32) string {
			left := false
			switch mode {
			case 0:
				left = qx < anchorX
			case 1:
				left = qz < anchorZ
			case 2:
				left = (qx+qy+qz)%2 == 0
			default:
				t.Fatalf("invalid biome pattern %d", mode)
			}
			if left {
				return pair[0]
			}
			return pair[1]
		}
		if biomeForQuart(wantQuart.X, wantQuart.Y, wantQuart.Z) != wantBiome {
			t.Fatalf("oracle quart/biome disagree: %s", scanner.Text())
		}
		rawQuart := FeaturePosition{position.X >> 2, position.Y >> 2, position.Z >> 2}
		if rawQuart != wantQuart {
			quartDifferences++
		}
		if biomeForQuart(rawQuart.X, rawQuart.Y, rawQuart.Z) != wantBiome {
			biomeDifferences++
		}
		region := &OverworldFeatureRegion{biomeSeed: seed, featureOrder: bound.order}
		var gotQuart FeaturePosition
		var gotBiome string
		calls := 0
		operations, err := bound.template.bind(region, func(qx, qy, qz int32) (string, error) {
			calls++
			gotQuart = FeaturePosition{qx, qy, qz}
			gotBiome = biomeForQuart(qx, qy, qz)
			return gotBiome, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		random := NewWorldgenRandom(seed)
		placed := 0
		err = walkPlacement(operations, random, position, func(got FeaturePosition, _ *WorldgenRandom) error {
			if got != position {
				return fmt.Errorf("biome filter moved candidate from %v to %v", position, got)
			}
			placed++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		next := int64(random.NextLong())
		if calls != 1 || gotQuart != wantQuart || gotBiome != wantBiome || placed > 1 || (placed == 1) != wantAllowed || next != wantNext {
			if mismatches < 8 {
				t.Errorf("%s seed=%d mode=%d block=%v: calls=%d quart=%v biome=%s allowed=%t RNG=%d; JAR quart=%v biome=%s allowed=%t RNG=%d",
					key, seed, mode, position, calls, gotQuart, gotBiome, placed == 1, next, wantQuart, wantBiome, wantAllowed, wantNext)
			}
			mismatches++
		}
		groups[key]++
		checked++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if checked != 18432 || len(groups) != 4 || quartDifferences == 0 || biomeDifferences == 0 {
		t.Fatalf("incomplete boundary coverage: cases=%d groups=%v raw-quart differences=%d raw-biome differences=%d", checked, groups, quartDifferences, biomeDifferences)
	}
	for group, count := range groups {
		if count != 4608 {
			t.Errorf("%s: got %d vectors, want 4608", group, count)
		}
	}
	t.Logf("checked %d JAR cases; %d raw-quart differences, %d raw-biome/admission differences, %d mismatches", checked, quartDifferences, biomeDifferences, mismatches)
	if mismatches != 0 {
		t.Fatalf("%d biome filter boundary cases disagree with Java", mismatches)
	}
}
