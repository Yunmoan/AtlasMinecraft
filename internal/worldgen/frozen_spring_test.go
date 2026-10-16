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
	if mode >= 152 {
		dx, dy, dz := p.X-origin.X, p.Y-origin.Y, p.Z-origin.Z
		if mode == 155 || mode == 164 {
			if dx == 0 && dz == 0 && dy == -1 {
				return "minecraft:stone"
			}
			if mode == 164 && dy == 0 && (dx != 0 || dz != 0) {
				return "minecraft:oak_slab"
			}
			return "minecraft:air"
		}
		if dy <= -1 {
			if dy == -1 && (dx != 0 || dz != 0) && mode <= 154 {
				return "minecraft:" + map[int]string{152: "sculk", 153: "sculk_catalyst", 154: "moving_piston"}[mode]
			}
			return "minecraft:stone"
		}
		if dy >= 3 {
			return "minecraft:stone"
		}
		if mode == 158 {
			return "minecraft:water"
		}
		if mode == 159 {
			return "flowing_water"
		}
		if mode >= 160 && mode <= 165 && dy == 0 && (dx != 0 || dz != 0) {
			return map[int]string{160: "minecraft:lava", 161: "minecraft:fire", 162: "minecraft:short_grass", 163: "minecraft:kelp", 164: "sculk_vein_faces", 165: "sculk_vein_faces"}[mode]
		}
		return "minecraft:air"
	}
	if mode >= 124 {
		dx, dy, dz := p.X-origin.X, p.Y-origin.Y, p.Z-origin.Z
		if mode == 127 {
			return "minecraft:stone"
		}
		if mode == 128 && dy == -1 || mode == 129 && dy == 4 || mode == 130 && dy >= 0 && dy <= 1 || mode == 140 && dy == -2 {
			return "minecraft:air"
		}
		if mode == 131 && dy == -1 {
			return "minecraft:water"
		}
		if mode == 132 && abs32(dx) < 3 && abs32(dz) < 3 && dy >= 0 && dy < 4 {
			return "minecraft:bedrock"
		}
		if dx == 0 && dz == 0 && dy == 0 {
			if mode == 133 {
				return "minecraft:chest"
			}
			if mode == 134 {
				return "minecraft:spawner"
			}
		}
		openings := int32(1)
		if mode >= 147 {
			openings = int32(mode - 145)
		}
		if dx >= 3 && dz >= -2 && dz < -2+openings && dy >= 0 && dy <= 1 {
			if mode == 138 {
				return "minecraft:water"
			}
			return "minecraft:air"
		}
		if name := map[int]string{135: "deepslate", 136: "oak_leaves", 137: "oak_slab", 139: "gravel"}[mode]; name != "" {
			return "minecraft:" + name
		}
		return "minecraft:stone"
	}
	if mode >= 112 {
		floor := int32(40)
		if mode == 119 {
			floor = 318
		}
		if mode == 120 {
			floor = -65
		}
		if mode == 115 && p.X == origin.X && p.Z == origin.Z && (p.Y == floor-1 || p.Y == floor-2) {
			return "minecraft:air"
		}
		if mode == 116 && p.X == origin.X && p.Z == origin.Z && p.Y == floor-2 {
			return "minecraft:air"
		}
		if p.Y < floor {
			return "minecraft:stone"
		}
		if p.Y == floor {
			if mode == 117 {
				return "minecraft:water"
			}
			return "minecraft:sand"
		}
		if mode == 118 && p.Y == floor+1 {
			return "minecraft:water"
		}
		if mode == 121 && p.Y == floor+4 {
			return "minecraft:stone"
		}
		return "minecraft:air"
	}
	if mode >= 104 {
		if p.Y < 40 {
			return "minecraft:stone"
		}
		if p.Y == 40 {
			switch mode {
			case 104:
				return "minecraft:dirt_path"
			case 105:
				return "minecraft:mycelium"
			case 106:
				return "minecraft:farmland"
			case 107:
				return "minecraft:bamboo"
			case 108:
				return "minecraft:bamboo_sapling"
			case 109:
				return "minecraft:gravel"
			}
			return "minecraft:grass_block"
		}
		if mode == 110 && p.Y == 43 || mode == 111 && p.Y == 55 {
			return "minecraft:stone"
		}
		return "minecraft:air"
	}
	if mode >= 88 {
		if mode == 91 {
			return "minecraft:air"
		}
		if mode == 92 {
			return "minecraft:stone"
		}
		floor := int32(40)
		switch mode {
		case 93:
			floor = 61
		case 94:
			floor = 63
		case 95:
			floor = 62 + ((p.X+p.Z)%4+4)%4
		}
		if p.Y <= floor {
			return "minecraft:stone"
		}
		if p.Y == 63 {
			switch mode {
			case 96:
				return "minecraft:ice"
			case 97:
				return "minecraft:packed_ice"
			case 98:
				return "minecraft:snow_block"
			}
		}
		if mode == 99 && p.Y >= 63 && p.Y <= 64 && (p.X+p.Z)%5 == 0 {
			return "minecraft:snow_block"
		}
		if mode == 100 && (p.X+p.Z)%5 == 0 {
			return "minecraft:stone"
		}
		if p.Y < 63 {
			if mode == 103 {
				return "minecraft:lava"
			}
			return "minecraft:water"
		}
		return "minecraft:air"
	}
	if mode >= 64 {
		floor := int32(40)
		if mode == 65 {
			floor = 38 + ((p.X+p.Z)%5+5)%5
		} else if mode == 80 {
			floor = 318
		} else if mode == 81 {
			floor = -65
		}
		if p.Y < floor {
			return "minecraft:stone"
		}
		if p.Y == floor {
			names := map[int]string{67: "sand", 68: "magma_block", 69: "soul_sand", 70: "honey_block", 71: "oak_leaves", 72: "oak_slab", 73: "sea_pickle", 74: "tube_coral_block"}
			if name := names[mode]; name != "" {
				return "minecraft:" + name
			}
			return "minecraft:stone"
		}
		if mode == 79 {
			return "minecraft:air"
		}
		if mode == 84 && (p.X+p.Z)%3 == 0 && p.Y >= 44 {
			return "minecraft:stone"
		}
		if p.Y == 41 {
			if mode == 85 {
				return "minecraft:kelp"
			}
			if mode == 86 {
				return "minecraft:seagrass"
			}
			if mode == 87 {
				return "minecraft:tube_coral"
			}
		}
		top := int32(60)
		if mode >= 76 && mode <= 78 {
			top = int32(mode) - 35
		}
		if p.Y <= top {
			if mode == 75 {
				return "flowing_water"
			}
			return "minecraft:water"
		}
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
						if name == "sculk_vein_faces" {
							ch.SurfaceBlocks[i] = SurfaceBlockState{Name: "minecraft:sculk_vein", Properties: map[string]string{"down": "true", "north": "true"}}
						} else if name == "flowing_water" {
							ch.SurfaceBlocks[i] = SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "1"}}
						} else if name == "minecraft:stone" {
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
	testSurfaceFeatureReference(t, "cave_placement_1.21.1.txt", "cave_placement", 1024, 12)
}

func TestOceanFeaturesAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "ocean_features_1.21.1.txt", "ocean_feature", 2112, 11)
}

func TestOceanPlacementAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "ocean_placement_1.21.1.txt", "ocean_placement", 2496, 11)
}

func TestIcebergFeaturesAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "iceberg_features_1.21.1.txt", "iceberg_feature", 320, 2)
}

func TestIcebergPlacementAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "iceberg_placement_1.21.1.txt", "iceberg_placement", 320, 2)
}

func TestLandFeaturesAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "land_features_1.21.1.txt", "land_feature", 1792, 8)
}

func TestDesertWellAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "desert_well_1.21.1.txt", "well_feature", 256, 1)
}

func TestMonsterRoomAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "monster_room_1.21.1.txt", "room_feature", 224, 1)
}

func TestSculkVeinAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "sculk_vein_1.21.1.txt", "sculk_feature", 640, 1)
}

func TestSculkVeinPlacementAgainstVanillaJAR(t *testing.T) {
	testSurfaceFeatureReference(t, "sculk_vein_placement_1.21.1.txt", "sculk_placement", 128, 1)
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
		wantFields := 12
		if label == "well_feature" || strings.HasPrefix(label, "room_") {
			wantFields = 13
		}
		if len(f) != wantFields || f[0] != label {
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
		chunks := make([]*NoiseChunkResult, 0, len(template.chunks))
		for _, chunk := range template.chunks {
			chunks = append(chunks, chunk)
		}
		region, err := NewOverworldFeatureRegion(chunks...)
		if err != nil {
			t.Fatal(err)
		}
		region.borrowed, region.biomeAtQuart = template.borrowed, template.biomeAtQuart
		region.biomeSeed = number(f[2])
		random := NewWorldgenRandom(number(f[2]))
		var placed bool
		if strings.HasSuffix(label, "_placement") {
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
			if label == "sculk_placement" {
				biome = "minecraft:deep_dark"
			}
			if label == "iceberg_placement" {
				biome = "minecraft:frozen_ocean"
			}
			if label == "ocean_placement" {
				switch f[1] {
				case "minecraft:seagrass_river":
					biome = "minecraft:river"
				case "minecraft:seagrass_swamp":
					biome = "minecraft:swamp"
				case "minecraft:seagrass_normal":
					biome = "minecraft:ocean"
				case "minecraft:seagrass_deep":
					biome = "minecraft:deep_ocean"
				case "minecraft:seagrass_cold", "minecraft:kelp_cold":
					biome = "minecraft:cold_ocean"
				case "minecraft:seagrass_deep_cold":
					biome = "minecraft:deep_cold_ocean"
				case "minecraft:kelp_warm", "minecraft:seagrass_deep_warm":
					biome = "minecraft:deep_lukewarm_ocean"
				default:
					biome = "minecraft:warm_ocean"
				}
			}
			if f[1] == "minecraft:vines" {
				biome = "minecraft:jungle"
			}
			if strings.Contains(f[1], "dripstone") {
				biome = "minecraft:dripstone_caves"
			}
			if f[1] == "minecraft:underwater_magma" || f[1] == "minecraft:glow_lichen" || strings.Contains(f[1], "lake_lava") || f[1] == "minecraft:amethyst_geode" {
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
			if strings.HasPrefix(f[1], "minecraft:coral_") {
				place, err = compileConfiguredFeaturePlan(f[1], ConfiguredFeatureDefinition{Type: f[1], Config: []byte("{}")}, 0)
			}
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
		if label == "well_feature" || strings.HasPrefix(label, "room_") {
			var values []string
			for p, entity := range region.blockEntities {
				id, _ := entity.data.Text("id")
				loot, _ := entity.data.Text("LootTable")
				if id == "minecraft:mob_spawner" {
					data, _ := entity.data.Object("SpawnData")
					mob, _ := data.Object("entity")
					loot, _ = mob.Text("id")
					for _, field := range []string{"Delay", "MinSpawnDelay", "MaxSpawnDelay", "SpawnCount", "MaxNearbyEntities", "RequiredPlayerRange", "SpawnRange"} {
						value, _ := entity.data.Number(field)
						loot += fmt.Sprintf(" %d", value)
					}
				}
				seed, _ := entity.data.Number("LootTableSeed")
				values = append(values, fmt.Sprintf("%d %d %d %s %s %d", p.X, p.Y, p.Z, id, loot, seed))
			}
			sort.Strings(values)
			entityHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(values, "\n"))))
			if entityHash != f[12] {
				t.Errorf("well entities %s: got %s", scanner.Text(), entityHash)
			}
		}
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
