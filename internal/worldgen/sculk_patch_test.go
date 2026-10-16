package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func sculkPatchProbeBase(p, origin FeaturePosition, mode int) SurfaceBlockState {
	state := func(s string) SurfaceBlockState { return SurfaceBlockState{Name: "minecraft:" + s} }
	if p.Y < -64 || p.Y >= 320 {
		return state("air")
	}
	dx, dy, dz := p.X-origin.X, p.Y-origin.Y, p.Z-origin.Z
	switch mode {
	case 0:
		return state("air")
	case 1:
		return state("stone")
	case 2:
		return state("water")
	case 3:
		if dy == 0 {
			return SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "1"}}
		}
	case 4:
		if dy == 0 {
			return state("lava")
		}
	case 5:
		if dy == -1 {
			return state("sculk")
		}
	case 6:
		if dy <= -1 {
			return state("deepslate_bricks")
		}
	case 7:
		if dy == 0 && dx == 0 && dz == 0 {
			return state("sculk")
		}
	case 8:
		if dy <= -1 {
			return state("oak_slab")
		}
	case 9:
		if dy <= -1 {
			return state("bedrock")
		}
	case 10:
		if dy <= -1 {
			return state("sand")
		}
	case 11:
		if dy >= 0 && dy <= 2 {
			return state("water")
		}
	case 16, 17, 18, 19:
		if dy <= -1 {
			return state(map[int]string{16: "oak_leaves", 17: "glass", 18: "farmland", 19: "magma_block"}[mode])
		}
	}
	if dy <= -1 || dy >= 4 {
		return state("stone")
	}
	return state("air")
}

func sculkPatchProbeRegion(t *testing.T, origin FeaturePosition, mode int) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	for cz := (origin.Z >> 4) - 2; cz <= (origin.Z>>4)+2; cz++ {
		for cx := (origin.X >> 4) - 2; cx <= (origin.X>>4)+2; cx++ {
			ch := &NoiseChunkResult{X: cx, Z: cz, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: map[int]SurfaceBlockState{}}
			// A cursor stays within 15 horizontal blocks of its origin. The
			// remaining 20 vertical blocks include all adjacency/growth reads.
			for y := max(-64, origin.Y-20); y <= min(319, origin.Y+20); y++ {
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						state := sculkPatchProbeBase(FeaturePosition{cx*16 + x, y, cz*16 + z}, origin, mode)
						i, _ := ch.Index(x, y, z)
						switch state.Name {
						case "minecraft:air":
						case "minecraft:stone":
							ch.Blocks[i] = NoiseStone
						case "minecraft:water":
							ch.Blocks[i] = NoiseWater
							if len(state.Properties) > 0 {
								ch.SurfaceBlocks[i] = state
							}
						case "minecraft:lava":
							ch.Blocks[i] = NoiseLava
						default:
							ch.SurfaceBlocks[i] = state
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
	for pos := range region.chunks {
		region.borrowed[pos] = true
	}
	return region
}

func TestSculkPatchAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/sculk_patch_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	regions := map[int]*OverworldFeatureRegion{}
	scanner := bufio.NewScanner(file)
	cases, successes := 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 12 || f[0] != "sculk_patch" {
			t.Fatalf("invalid sculk patch vector %s", scanner.Text())
		}
		parse := func(s string) int64 {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		seed, mode := parse(f[2]), int(parse(f[3]))
		origin := FeaturePosition{int32(parse(f[4])), int32(parse(f[5])), int32(parse(f[6]))}
		if regions[mode] == nil {
			regions[mode] = sculkPatchProbeRegion(t, origin, mode)
		}
		region := regions[mode]
		region.overlays = nil
		region.blockEntities = nil
		var definition ConfiguredFeatureDefinition
		if strings.HasSuffix(f[1], "_probe") {
			config := `{"charge_count":2,"amount_per_charge":120,"spread_attempts":64,"growth_rounds":2,"spread_rounds":2,"extra_rare_growths":2,"catalyst_chance":0}`
			if strings.Contains(f[1], "dense") {
				config = `{"charge_count":32,"amount_per_charge":500,"spread_attempts":64,"growth_rounds":1,"spread_rounds":2,"extra_rare_growths":0,"catalyst_chance":1}`
			}
			definition = ConfiguredFeatureDefinition{Type: "minecraft:sculk_patch", Config: json.RawMessage(config)}
		} else {
			definition, err = LoadConfiguredFeatureDefinition(f[1])
			if err != nil {
				t.Fatal(err)
			}
		}
		place, err := compileSculkPatchPlan(definition)
		if err != nil {
			t.Fatal(err)
		}
		r := NewWorldgenRandom(seed)
		placed, err := place(region, r, origin)
		if err != nil {
			t.Fatal(err)
		}
		var output []string
		marked := 0
		for key, overlay := range region.overlays {
			marked += len(overlay.postProcessing)
			for i, s := range overlay.blocks {
				p := FeaturePosition{key.chunkX*16 + int32(i&15), -64 + int32(i/256), key.chunkZ*16 + int32((i>>4)&15)}
				id, err := blockstate.BlockStateID(s.Name, s.Properties)
				if err != nil {
					t.Fatal(err)
				}
				output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
			}
		}
		sort.Strings(output)
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(output, "\n"))))
		next := int64(r.NextLong())
		if placed != (parse(f[7]) == 1) || len(output) != int(parse(f[8])) || next != parse(f[9]) || hash != f[10] || marked != int(parse(f[11])) {
			t.Errorf("seed=%d mode=%d %s: placed=%v writes=%d next=%d hash=%s marked=%d; want %s", seed, mode, f[1], placed, len(output), next, hash, marked, scanner.Text())
		}
		cases++
		if placed {
			successes++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 264 || successes != 168 {
		t.Fatalf("got %d vectors / %d positive cases, want 264 / 168", cases, successes)
	}
	t.Logf("matched %d official JAR cases including %d positive cases", cases, successes)
}

func TestDecoratedDeepDarkIncludesSculkPatch(t *testing.T) {
	g, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := g.plans["minecraft:sculk_patch_deep_dark"]; !ok {
		t.Fatal("validated deep-dark sculk patch is not enabled")
	}
	result, err := g.Generate(-256, 232)
	if err != nil {
		t.Fatal(err)
	}
	sculk, veins, growths := 0, 0, 0
	for _, state := range result.Terrain.SurfaceBlocks {
		switch state.Name {
		case "minecraft:sculk":
			sculk++
		case "minecraft:sculk_vein":
			veins++
		case "minecraft:sculk_sensor", "minecraft:sculk_shrieker", "minecraft:sculk_catalyst":
			growths++
		}
	}
	if sculk == 0 || veins == 0 {
		t.Fatalf("runtime deep-dark chunk has %d sculk blocks, %d veins, %d growths", sculk, veins, growths)
	}
	for _, name := range result.SkippedFeatures {
		if name == "minecraft:sculk_patch_deep_dark" {
			t.Fatal("supported sculk patch skipped at runtime")
		}
	}
	t.Logf("runtime deep-dark chunk contains %d sculk blocks, %d veins, %d growths", sculk, veins, growths)
}
