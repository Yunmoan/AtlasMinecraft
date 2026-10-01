package worldgen

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

type stoneOreWorld struct {
	changed map[FeaturePosition]SurfaceBlockState
}

func (world *stoneOreWorld) OceanFloorHeightAt(int32, int32) (int32, error) {
	return 320, nil
}

func (world *stoneOreWorld) BlockAt(position FeaturePosition) (SurfaceBlockState, error) {
	if state, exists := world.changed[position]; exists {
		return state, nil
	}
	if position.X >= 0 && position.X < 16 && position.Z >= 0 && position.Z < 16 && position.Y >= 0 && position.Y < 40 {
		return SurfaceBlockState{Name: "minecraft:stone"}, nil
	}
	return SurfaceBlockState{Name: "minecraft:air"}, nil
}

func (world *stoneOreWorld) SetBlock(position FeaturePosition, state SurfaceBlockState) error {
	world.changed[position] = state
	return nil
}

func (world *stoneOreWorld) InWorld(position FeaturePosition) bool {
	return position.Y >= -64 && position.Y < 320
}

func TestOreFeatureAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/ore_feature_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type key struct {
		seed int64
		size int32
	}
	type summary struct {
		placed bool
		count  int
		next   int64
	}
	summaries := make(map[key]summary)
	expected := make(map[key][]FeaturePosition)
	type oreCheck struct {
		seed    int64
		chance  float32
		exposed bool
		allowed bool
		next    int64
	}
	var checks []oreCheck
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		parse := func(value string) int64 {
			result, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return result
		}
		if len(fields) < 3 {
			t.Fatalf("invalid JAR ore vector: %q", scanner.Text())
		}
		switch fields[0] {
		case "ore_count":
			if len(fields) != 6 {
				t.Fatalf("invalid JAR ore count: %q", scanner.Text())
			}
			caseKey := key{parse(fields[1]), int32(parse(fields[2]))}
			summaries[caseKey] = summary{parse(fields[3]) != 0, int(parse(fields[4])), parse(fields[5])}
		case "ore_block":
			if len(fields) != 6 {
				t.Fatalf("invalid JAR ore block: %q", scanner.Text())
			}
			caseKey := key{parse(fields[1]), int32(parse(fields[2]))}
			expected[caseKey] = append(expected[caseKey], FeaturePosition{
				int32(parse(fields[3])), int32(parse(fields[4])), int32(parse(fields[5]))})
		case "ore_check":
			if len(fields) != 6 {
				t.Fatalf("invalid JAR ore predicate: %q", scanner.Text())
			}
			checks = append(checks, oreCheck{parse(fields[1]), math.Float32frombits(uint32(parse(fields[2]))),
				parse(fields[3]) != 0, parse(fields[4]) != 0, parse(fields[5])})
		default:
			t.Fatalf("unknown JAR ore vector: %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 12 {
		t.Fatalf("got %d JAR ore cases, want 12", len(summaries))
	}
	if len(checks) != 32 {
		t.Fatalf("got %d JAR ore predicate cases, want 32", len(checks))
	}
	for caseKey, want := range summaries {
		t.Run(fmt.Sprintf("%d/%d", caseKey.seed, caseKey.size), func(t *testing.T) {
			world := &stoneOreWorld{changed: make(map[FeaturePosition]SurfaceBlockState)}
			random := NewWorldgenRandom(caseKey.seed)
			config := OreConfiguration{Size: caseKey.size, Targets: []OreTarget{{
				Predicate: "minecraft:block_match", Block: "minecraft:stone",
				State: SurfaceBlockState{Name: "minecraft:dirt"},
			}}}
			placed, err := PlaceOreFeature(world, random, FeaturePosition{8, 20, 8}, config)
			if err != nil {
				t.Fatal(err)
			}
			next := int64(random.NextLong())
			if placed != want.placed || len(world.changed) != want.count || next != want.next {
				t.Fatalf("got placed=%t blocks=%d next=%d, want %+v", placed,
					len(world.changed), next, want)
			}
			for _, position := range expected[caseKey] {
				if world.changed[position].Name != "minecraft:dirt" {
					t.Fatalf("missing ore block at %+v", position)
				}
			}
		})
	}
	for _, check := range checks {
		world := &stoneOreWorld{changed: make(map[FeaturePosition]SurfaceBlockState)}
		random := NewWorldgenRandom(check.seed)
		config := OreConfiguration{DiscardOnAir: check.chance, Targets: []OreTarget{{
			Predicate: "minecraft:block_match", Block: "minecraft:stone",
			State: SurfaceBlockState{Name: "minecraft:dirt"},
		}}}
		position := FeaturePosition{8, 20, 8}
		if check.exposed {
			position.Y = 39
		}
		_, allowed, err := oreMayReplace(world, random, position,
			SurfaceBlockState{Name: "minecraft:stone"}, config)
		if err != nil {
			t.Fatal(err)
		}
		if next := int64(random.NextLong()); allowed != check.allowed || next != check.next {
			t.Fatalf("ore predicate seed=%d chance=%f exposed=%t: got allowed=%t next=%d, want allowed=%t next=%d",
				check.seed, check.chance, check.exposed, allowed, next, check.allowed, check.next)
		}
	}
}

func TestLoadOreConfiguration(t *testing.T) {
	config, err := LoadOreConfiguration("minecraft:ore_iron")
	if err != nil {
		t.Fatal(err)
	}
	if config.Size != 9 || len(config.Targets) != 2 || !config.replacementTags["minecraft:stone_ore_replaceables"]["minecraft:stone"] ||
		!config.replacementTags["minecraft:deepslate_ore_replaceables"]["minecraft:deepslate"] {
		t.Fatalf("unexpected iron ore configuration: %+v", config)
	}
	definitions, err := loadFeatureDefinitions()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for name, definition := range definitions.Configured {
		if definition.Type != "minecraft:ore" {
			continue
		}
		if _, err := LoadOreConfiguration(name); err != nil {
			t.Fatalf("ore %s: %v", name, err)
		}
		count++
	}
	if count != 30 {
		t.Fatalf("got %d configured ores, want 30", count)
	}
}
