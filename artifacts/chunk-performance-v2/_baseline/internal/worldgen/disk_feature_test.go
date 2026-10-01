package worldgen

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type dirtDiskWorld struct {
	changed map[FeaturePosition]SurfaceBlockState
}

func (world *dirtDiskWorld) BlockAt(position FeaturePosition) (SurfaceBlockState, error) {
	if state, exists := world.changed[position]; exists {
		return state, nil
	}
	if position.Y == 40 {
		return SurfaceBlockState{Name: "minecraft:dirt"}, nil
	}
	return SurfaceBlockState{Name: "minecraft:air"}, nil
}

func (world *dirtDiskWorld) SetBlock(position FeaturePosition, state SurfaceBlockState) error {
	world.changed[position] = state
	return nil
}

func (world *dirtDiskWorld) InWorld(position FeaturePosition) bool {
	return position.Y >= -64 && position.Y < 320
}

func TestDiskFeaturesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/disk_feature_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cases := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 10 || fields[0] != "disk_case" {
			t.Fatalf("invalid JAR disk vector: %q", scanner.Text())
		}
		parse := func(value string) int64 {
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return parsed
		}
		placedDefinition, err := LoadPlacedFeatureDefinition(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		configuration, err := LoadDiskConfiguration(placedDefinition.Feature)
		if err != nil {
			t.Fatal(err)
		}
		world := &dirtDiskWorld{changed: make(map[FeaturePosition]SurfaceBlockState)}
		random := NewWorldgenRandom(parse(fields[2]))
		origin := FeaturePosition{int32(parse(fields[3])), int32(parse(fields[4])), int32(parse(fields[5]))}
		placed, err := PlaceDiskFeature(world, random, origin, configuration)
		if err != nil {
			t.Fatal(err)
		}
		positions := make([]string, 0, len(world.changed))
		for position, state := range world.changed {
			if !reflect.DeepEqual(state, configuration.State) {
				t.Fatalf("wrong disk block at %+v: %+v", position, state)
			}
			positions = append(positions, fmt.Sprintf("%d %d %d", position.X, position.Y, position.Z))
		}
		sort.Strings(positions)
		hash := sha256.Sum256([]byte(strings.Join(positions, "\n")))
		if placed != (parse(fields[6]) != 0) || len(world.changed) != int(parse(fields[7])) ||
			int64(random.NextLong()) != parse(fields[8]) || hex.EncodeToString(hash[:]) != fields[9] {
			t.Fatalf("%s seed %s origin %+v: placed=%t count=%d hash=%x, want %s", fields[1],
				fields[2], origin, placed, len(world.changed), hash, scanner.Text())
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 48 {
		t.Fatalf("got %d JAR disk cases, want 48", cases)
	}
}

func TestDiskRulesCompile(t *testing.T) {
	for _, name := range []string{"minecraft:disk_sand", "minecraft:disk_grass"} {
		if _, err := LoadDiskConfiguration(name); err != nil {
			t.Fatalf("rule-based disk %s failed: %v", name, err)
		}
	}
}
