package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func frozenSpringProbeBase(p, origin FeaturePosition, mode int) string {
	if p.Y < -64 || p.Y >= 320 {
		return "minecraft:air"
	}
	if mode >= 56 {
		if mode == 56 && p == origin {
			return "minecraft:air"
		}
		switch mode {
		case 57:
			return "minecraft:deepslate"
		case 58:
			return "minecraft:water"
		case 59:
			return "minecraft:bedrock"
		case 60:
			return "minecraft:amethyst_block"
		case 61:
			return "minecraft:lava"
		case 63:
			if p.Y == 41 {
				return "minecraft:water"
			}
		}
		return "minecraft:stone"
	}
	if mode >= 48 {
		if p.Y <= 40 {
			return "minecraft:stone"
		}
		if p.Y < 46 {
			return "minecraft:air"
		}
		if p.Y < 50 {
			return "minecraft:stone"
		}
		if p.Y == 50 {
			switch mode {
			case 52:
				return "minecraft:lava"
			case 53:
				return "minecraft:clay"
			case 54:
				return "minecraft:oak_leaves"
			case 55:
				return "minecraft:sand"
			}
			return "minecraft:grass_block"
		}
		if mode == 51 && p.Y <= 52 {
			return "minecraft:water"
		}
		return "minecraft:air"
	}
	if mode >= 40 {
		floor := int32(40)
		if mode == 47 {
			floor = 40 + ((p.X+p.Z)%3+3)%3
		}
		if p.Y < floor {
			return "minecraft:stone"
		}
		if p.Y == floor {
			switch mode {
			case 40, 41:
				return "minecraft:clay"
			case 42:
				return "minecraft:moss_block"
			case 44:
				return "minecraft:dirt"
			case 46:
				return "minecraft:sand"
			}
			return "minecraft:stone"
		}
		if p.Y >= 46 {
			return "minecraft:stone"
		}
		if mode == 41 && p.Y <= 43 {
			return "minecraft:water"
		}
		if mode == 45 && p.Y == 42 {
			return "minecraft:stone"
		}
		return "minecraft:air"
	}
	if mode >= 20 {
		floor, ceiling := int32(36), int32(46)
		switch mode {
		case 20:
			floor, ceiling = 40, 44
		case 22:
			floor, ceiling = 20, 64
		case 23:
			floor, ceiling = 10, 100
		case 28:
			ceiling = 42
		case 30:
			floor, ceiling = 34+((p.X+p.Z)%5+5)%5, 45+((p.X-p.Z)%7+7)%7
		case 31:
			return "minecraft:air"
		case 34:
			floor, ceiling = 318, 321
		case 35:
			floor, ceiling = -65, -58
		case 36:
			floor = 37
		case 37:
			floor = 36
		case 38:
			floor = 35
		case 39:
			if p == origin {
				return "minecraft:stone"
			}
		}
		if p.Y <= floor && mode != 28 {
			if mode == 25 && p.Y == floor {
				return "minecraft:lava"
			}
			if mode == 26 && p.Y == floor {
				return "minecraft:dripstone_block"
			}
			if mode == 27 {
				return "minecraft:dirt"
			}
			if mode == 29 {
				return "minecraft:deepslate"
			}
			return "minecraft:stone"
		}
		if p.Y >= ceiling {
			if mode == 27 {
				return "minecraft:calcite"
			}
			return "minecraft:stone"
		}
		if mode == 24 || mode >= 36 && mode <= 38 {
			return "minecraft:water"
		}
		return "minecraft:air"
	}
	if mode >= 16 {
		if p == origin || p == treeOffset(origin, 1, 0, 0) && mode != 17 {
			return "minecraft:air"
		}
		if mode == 18 {
			return "minecraft:packed_ice"
		}
		if mode == 19 {
			return "minecraft:netherrack"
		}
		return "minecraft:stone"
	}
	floor := int32(40)
	if mode == 8 {
		floor = 318
	}
	if mode == 9 {
		floor = -65
	}
	if mode == 1 {
		floor = 38 + ((p.X+p.Z)%5+5)%5
	}
	if p.Y < floor {
		return "minecraft:stone"
	}
	if p.Y == floor {
		switch mode {
		case 3:
			return "minecraft:sand"
		case 4:
			return "minecraft:podzol"
		case 5:
			return "minecraft:stone"
		case 7:
			return "minecraft:snow_block"
		case 12:
			return "minecraft:soul_sand"
		case 13:
			return "minecraft:honey_block"
		case 14:
			return "minecraft:ice"
		case 15:
			return "minecraft:oak_leaves"
		}
		return "minecraft:grass_block"
	}
	if mode == 6 && p.Y <= 44 {
		return "minecraft:water"
	}
	return "minecraft:air"
}

func frozenSpringProbeRegion(t *testing.T, origin FeaturePosition, mode int) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	for dx := int32(-1); dx <= 1; dx++ {
		for dz := int32(-1); dz <= 1; dz++ {
			ch := &NoiseChunkResult{X: (origin.X >> 4) + dx, Z: (origin.Z >> 4) + dz, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: make(map[int]SurfaceBlockState)}
			for y := int32(-64); y < 320; y++ {
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						p := FeaturePosition{ch.X*16 + x, y, ch.Z*16 + z}
						name := frozenSpringProbeBase(p, origin, mode)
						i := int((y+64)*256 + z*16 + x)
						if name == "minecraft:stone" {
							ch.Blocks[i] = NoiseStone
						} else if name != "minecraft:air" {
							ch.SurfaceBlocks[i] = SurfaceBlockState{Name: name}
						}
					}
				}
			}
			chunks = append(chunks, ch)
		}
	}
	region, err := NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	region.borrowed = make(map[featureChunkPosition]bool)
	for p := range region.chunks {
		region.borrowed[p] = true
	}
	region.biomeAtQuart = func(int32, int32, int32) (string, error) {
		if mode == 0 {
			return "minecraft:plains", nil
		}
		return "minecraft:snowy_plains", nil
	}
	return region
}

func TestFrozenAndSpringFeaturesAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "frozen_spring_1.21.1.txt", "frozen_spring", 1120, 6)
}

func TestCaveFeaturesAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "cave_features_1.21.1.txt", "cave_feature", 10752, 21)
}

func TestCavePlacementAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "cave_placement_1.21.1.txt", "cave_placement", 320, 10)
}

func testSurfaceFeatureReference(t *testing.T, fixture, label string, wantCases, wantSuccesses int) {
	file, err := os.Open("../../docs/worldgen/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	templates := make(map[[4]int32]*OverworldFeatureRegion)
	cases := 0
	successes := map[string]int{}
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 12 || f[0] != label {
			t.Fatalf("invalid vector %s", scanner.Text())
		}
		number := func(s string) int64 {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		origin := FeaturePosition{int32(number(f[4])), int32(number(f[5])), int32(number(f[6]))}
		mode := int(number(f[3]))
		key := [4]int32{origin.X, origin.Y, origin.Z, int32(mode)}
		template := templates[key]
		if template == nil {
			template = frozenSpringProbeRegion(t, origin, mode)
			templates[key] = template
		}
		region := &OverworldFeatureRegion{chunks: template.chunks, borrowed: template.borrowed, heights: make(map[featureColumnPosition]int32), motionHeights: make(map[featureColumnPosition]int32), surfaceHeights: make(map[featureColumnPosition]int32), biomeAtQuart: template.biomeAtQuart}
		region.biomeSeed = number(f[2])
		random := NewWorldgenRandom(number(f[2]))
		var placed bool
		if label == "cave_placement" {
			order, err := loadOverworldFeatureOrder()
			if err != nil {
				t.Fatal(err)
			}
			var feature PlacedFeature
			for _, candidate := range order.features {
				if candidate.Name == f[1] {
					feature = candidate
				}
			}
			definition, err := LoadPlacedFeatureDefinition(f[1])
			if err != nil {
				t.Fatal(err)
			}
			place, err := compileFeaturePlan(definition.Feature, 0)
			if err != nil {
				t.Fatal(err)
			}
			biome := "minecraft:lush_caves"
			if strings.Contains(f[1], "dripstone") {
				biome = "minecraft:dripstone_caves"
			}
			if f[1] == "minecraft:underwater_magma" || f[1] == "minecraft:glow_lichen" {
				biome = "minecraft:plains"
			}
			operations, err := bindPlacement(region, definition, &feature, func(int32, int32, int32) (string, error) { return biome, nil })
			if err != nil {
				t.Fatal(err)
			}
			cx, cz := origin.X>>4, origin.Z>>4
			random = OverworldPlacedFeatureRandom(OverworldDecorationSeed(number(f[2]), cx, cz), feature)
			err = walkPlacement(operations, random, FeaturePosition{X: cx * 16, Z: cz * 16}, func(p FeaturePosition, r *WorldgenRandom) error {
				ok, err := place(region, r, p)
				placed = placed || ok
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
		} else {
			place, err := compileFeaturePlan(f[1], 0)
			if err != nil {
				t.Fatal(err)
			}
			placed, err = place(region, random, origin)
		}
		if err != nil {
			t.Fatal(err)
		}
		var output []string
		ticks := 0
		for position, overlay := range region.overlays {
			for i, state := range overlay.blocks {
				p := FeaturePosition{position.chunkX*16 + int32(i&15), -64 + int32(i/256), position.chunkZ*16 + int32((i>>4)&15)}
				id, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
			}
			ticks += len(overlay.postProcessing)
		}
		sort.Strings(output)
		hash := sha256.Sum256([]byte(strings.Join(output, "\n")))
		next := int64(random.NextLong())
		if placed != (number(f[7]) != 0) || len(output) != int(number(f[8])) || next != number(f[9]) || fmt.Sprintf("%x", hash) != f[10] || ticks != int(number(f[11])) {
			t.Errorf("%s seed=%s scenario=%d: placed=%v count=%d next=%d hash=%x ticks=%d; want %s", f[1], f[2], mode, placed, len(output), next, hash, ticks, scanner.Text())
		}
		if placed && len(output) > 0 {
			successes[f[1]]++
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != wantCases {
		t.Fatalf("got %d cases, want %d", cases, wantCases)
	}
	if len(successes) < wantSuccesses {
		t.Fatalf("only %d features have successful changed-block cases", len(successes))
	}
	if !t.Failed() {
		t.Logf("matched %d official JAR cases", cases)
	}
}
