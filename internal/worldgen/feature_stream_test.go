package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestFeaturePlacementStreamAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/feature_stream_1.21.1.txt")
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
	if len(vectors) != 60 {
		t.Fatalf("got %d stream cases, want 60", len(vectors))
	}
	for key, vector := range vectors {
		t.Run(fmt.Sprintf("%s/%d/%d,%d", key.feature.Name, key.seed, key.cx, key.cz), func(t *testing.T) {
			operations, err := compilePlacement(key.feature)
			if err != nil {
				t.Fatal(err)
			}
			// The JAR probe stops before the biome modifier and consumes a
			// variable amount of random state at each leaf, like a feature.
			if operations[len(operations)-1] != nil {
				t.Fatal("expected final biome modifier")
			}
			operations = operations[:len(operations)-1]
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
