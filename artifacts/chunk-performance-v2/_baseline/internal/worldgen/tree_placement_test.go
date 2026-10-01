package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func treePlacementBase(p FeaturePosition, scenario int) string {
	floorMod := func(value, divisor int32) int32 { return ((value % divisor) + divisor) % divisor }
	floor := int32(40)
	if scenario == 1 {
		floor = 38 + floorMod(p.X+p.Z, 5)
	}
	if p.Y < floor {
		return "minecraft:stone"
	} // Same heightmap predicate as probe dirt below the surface.
	if p.Y == floor {
		if scenario == 3 {
			return "minecraft:stone"
		}
		if scenario >= 4 {
			return "minecraft:snow_block"
		}
		return "minecraft:grass_block"
	}
	if scenario == 2 && p.Y <= 42 {
		return "minecraft:water"
	}
	if scenario == 4 && p.Y <= 41+floorMod(p.X+p.Z, 3) {
		return "minecraft:powder_snow"
	}
	if scenario == 6 && p.Y <= 51+floorMod(p.X+p.Z, 3) {
		return "minecraft:powder_snow"
	}
	return "minecraft:air"
}

func treePlacementRegion(t *testing.T, cx, cz int32, scenario int) *OverworldFeatureRegion {
	t.Helper()
	states := make(map[string]SurfaceBlockState)
	for _, name := range []string{"minecraft:grass_block", "minecraft:water", "minecraft:snow_block", "minecraft:powder_snow"} {
		id, err := blockstate.BlockStateID(name, nil)
		if err != nil {
			t.Fatal(err)
		}
		state, err := blockstate.BlockStateByID(id)
		if err != nil {
			t.Fatal(err)
		}
		states[name] = SurfaceBlockState{Name: state.Name, Properties: state.Properties}
	}
	var chunks []*NoiseChunkResult
	for dx := int32(-1); dx <= 1; dx++ {
		for dz := int32(-1); dz <= 1; dz++ {
			chunk := &NoiseChunkResult{X: cx + dx, Z: cz + dz, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: make(map[int]SurfaceBlockState)}
			for y := int32(-64); y <= 56; y++ {
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						name := treePlacementBase(FeaturePosition{chunk.X*16 + x, y, chunk.Z*16 + z}, scenario)
						if name == "minecraft:stone" {
							index, _ := chunk.Index(x, y, z)
							chunk.Blocks[index] = NoiseStone
						} else if name != "minecraft:air" {
							chunk.setSurfaceState(x, y, z, states[name])
						}
					}
				}
			}
			chunks = append(chunks, chunk)
		}
	}
	region, err := NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	return region
}

func TestTreePlacementAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/tree_placement_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		t.Fatal(err)
	}
	features := make(map[string]PlacedFeature)
	for _, feature := range order.features {
		features[feature.Name] = feature
	}
	biomes := map[string]string{"minecraft:trees_taiga": "minecraft:taiga", "minecraft:trees_grove": "minecraft:grove", "minecraft:trees_snowy": "minecraft:snowy_plains", "minecraft:trees_water": "minecraft:river", "minecraft:trees_windswept_forest": "minecraft:windswept_forest", "minecraft:trees_windswept_hills": "minecraft:windswept_hills", "minecraft:trees_badlands": "minecraft:wooded_badlands"}
	scanner := bufio.NewScanner(file)
	cases, successes := 0, 0
	byFeature := make(map[string]int)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 11 || fields[0] != "tree_placement" {
			t.Fatalf("invalid tree placement vector: %s", scanner.Text())
		}
		parse := func(s string) int64 {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		cx, cz, scenario := int32(parse(fields[3])), int32(parse(fields[4])), int(parse(fields[5]))
		region := treePlacementRegion(t, cx, cz, scenario)
		feature, topLevel := features[fields[1]]
		if !topLevel {
			feature = PlacedFeature{Name: fields[1]}
		}
		definition, err := LoadPlacedFeatureDefinition(feature.Name)
		if err != nil {
			t.Fatal(err)
		}
		placer, err := compileFeaturePlan(definition.Feature, 0)
		if err != nil {
			t.Fatal(err)
		}
		featureContext := &feature
		if !topLevel {
			featureContext = nil
		}
		operations, err := bindPlacement(region, definition, featureContext, func(int32, int32, int32) (string, error) { return biomes[feature.Name], nil })
		if err != nil {
			t.Fatal(err)
		}
		random := OverworldPlacedFeatureRandom(OverworldDecorationSeed(parse(fields[2]), cx, cz), feature)
		origin := FeaturePosition{X: cx * 16, Z: cz * 16}
		if !topLevel {
			random = NewWorldgenRandom(parse(fields[2]))
			origin.Y = 41
		}
		candidates, placed := 0, 0
		err = walkPlacement(operations, random, origin, func(p FeaturePosition, random *WorldgenRandom) error {
			candidates++
			success, err := placer(region, random, p)
			if success {
				placed++
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		var output []string
		for _, chunk := range region.chunks {
			for i, state := range chunk.SurfaceBlocks {
				p := FeaturePosition{chunk.X*16 + int32(i&15), -64 + int32(i/256), chunk.Z*16 + int32((i>>4)&15)}
				id, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				base, err := blockstate.BlockStateID(treePlacementBase(p, scenario), nil)
				if err != nil {
					t.Fatal(err)
				}
				if id != base {
					output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
				}
			}
		}
		sort.Strings(output)
		hash := sha256.Sum256([]byte(strings.Join(output, "\n")))
		next := int64(random.NextLong())
		if candidates != int(parse(fields[6])) || placed != int(parse(fields[7])) || len(output) != int(parse(fields[8])) || next != parse(fields[9]) || hex.EncodeToString(hash[:]) != fields[10] {
			t.Errorf("%s seed %s chunk %d,%d scenario %d: candidates=%d placed=%d blocks=%d next=%d hash=%x; want %s", feature.Name, fields[2], cx, cz, scenario, candidates, placed, len(output), next, hash, scanner.Text())
		}
		successes += placed
		byFeature[feature.Name] += placed
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 1008 || successes < 100 {
		t.Fatalf("insufficient tree placement coverage: %d cases, %d successful trees", cases, successes)
	}
	for _, name := range []string{"trees_taiga", "trees_grove", "trees_snowy", "trees_water", "trees_windswept_forest", "trees_windswept_hills", "trees_badlands", "pine_on_snow", "spruce_on_snow"} {
		if byFeature["minecraft:"+name] == 0 {
			t.Fatalf("no successful placement coverage for %s", name)
		}
	}
}

func TestTreePlanCoverage(t *testing.T) {
	generator, err := NewOverworldGenerator(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"trees_taiga", "trees_grove", "trees_snowy", "trees_water", "trees_windswept_forest", "trees_windswept_hills", "trees_badlands"} {
		if _, exists := generator.plans["minecraft:"+name]; !exists {
			t.Fatalf("supported tree plan %s missing", name)
		}
	}
	for _, name := range []string{"trees_plains", "trees_birch", "trees_birch_and_oak", "trees_swamp", "trees_old_growth_spruce_taiga"} {
		if _, exists := generator.plans["minecraft:"+name]; exists {
			t.Fatalf("unsupported decorated tree plan %s enabled", name)
		}
	}
}
