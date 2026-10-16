package worldgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestDimensionPlacementStreamAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/dimension_placement_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type key struct {
		seed    int64
		cx, cz  int32
		feature PlacedFeature
	}
	type vector struct {
		positions []FeaturePosition
		count     int
		next      int64
	}
	vectors := make(map[key]*vector)
	parse := func(text string) int64 {
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 9 && len(fields) != 10 {
			t.Fatalf("invalid stream vector: %s", scanner.Text())
		}
		k := key{parse(fields[1]), int32(parse(fields[2])), int32(parse(fields[3])), PlacedFeature{int(parse(fields[4])), int(parse(fields[5])), fields[6]}}
		if vectors[k] == nil {
			vectors[k] = &vector{}
		}
		if fields[0] == "stream_candidate" && len(fields) == 10 {
			vectors[k].positions = append(vectors[k].positions, FeaturePosition{int32(parse(fields[7])), int32(parse(fields[8])), int32(parse(fields[9]))})
		} else if fields[0] == "stream_count" && len(fields) == 9 {
			vectors[k].count, vectors[k].next = int(parse(fields[7])), parse(fields[8])
		} else {
			t.Fatalf("invalid stream vector: %s", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 99 {
		t.Fatalf("got %d stream cases, want 99", len(vectors))
	}
	for key, vector := range vectors {
		t.Run(fmt.Sprintf("%s/%d/%d,%d", key.feature.Name, key.seed, key.cx, key.cz), func(t *testing.T) {
			terrain := &NoiseChunkResult{X: key.cx, Z: key.cz, MinY: 0, Height: 256, Blocks: make([]NoiseMaterial, 256*256)}
			for y := int32(0); y <= 120; y++ {
				name := "minecraft:air"
				switch {
				case y == 0 || y >= 111 && y <= 120:
					name = "minecraft:bedrock"
				case y > 0 && y <= 20 || y >= 61 && y <= 70:
					name = "minecraft:netherrack"
				case y >= 21 && y <= 30:
					name = "minecraft:lava"
				case y >= 51 && y <= 55:
					name = "minecraft:soul_sand"
				}
				if name == "minecraft:air" {
					continue
				}
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						terrain.setSurfaceState(x, y, z, SurfaceBlockState{Name: name})
					}
				}
			}
			region, err := NewDimensionFeatureRegion(128, terrain)
			if err != nil {
				t.Fatal(err)
			}
			var definition PlacedFeatureDefinition
			if key.feature.Name == "atlas:variable_layer" {
				definition.Placement = []json.RawMessage{json.RawMessage(`{"type":"minecraft:count_on_every_layer","count":{"type":"minecraft:uniform","min_inclusive":0,"max_inclusive":5}}`)}
			} else {
				definition, err = LoadPlacedFeatureDefinition(key.feature.Name)
				if err != nil {
					t.Fatal(err)
				}
				definition.Placement = definition.Placement[:len(definition.Placement)-1]
			}
			operations, err := bindPlacement(region, definition, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			random := OverworldPlacedFeatureRandom(OverworldDecorationSeed(key.seed, key.cx, key.cz), key.feature)
			var got []FeaturePosition
			err = walkPlacement(operations, random, FeaturePosition{X: key.cx * 16, Z: key.cz * 16}, func(position FeaturePosition, random *WorldgenRandom) error {
				got = append(got, position)
				calls := random.NextIntBound(7) + 1
				for range calls {
					random.NextLong()
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, vector.positions) || len(got) != vector.count {
				t.Fatalf("stream positions differ: got %v, want %v", got, vector.positions)
			}
			if next := int64(random.NextLong()); next != vector.next {
				t.Fatalf("next random long=%d, vanilla=%d", next, vector.next)
			}
		})
	}
}
