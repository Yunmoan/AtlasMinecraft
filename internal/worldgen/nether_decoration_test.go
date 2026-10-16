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

func netherDecorationProbeBase(p, origin FeaturePosition, ceiling bool, mode int) SurfaceBlockState {
	s := func(name string) SurfaceBlockState { return SurfaceBlockState{Name: "minecraft:" + name} }
	mod := func(n, d int32) int32 { return (n%d + d) % d }
	if p.Y < 0 || p.Y >= 256 || mode == 3 {
		return s("air")
	}
	if mode == 4 {
		return s("stone")
	}
	dx, dy, dz := p.X-origin.X, p.Y-origin.Y, p.Z-origin.Z
	floor, roof := origin.Y-1, origin.Y+1
	if ceiling {
		floor = origin.Y - 12
	}
	if mode == 15 {
		floor += mod(dx+dz, 5) - 2
	}
	if mode == 25 {
		floor += mod(dx*dx+dz*dz, 7) - 3
	}
	support := "netherrack"
	if name, ok := map[int]string{5: "stone", 6: "basalt", 7: "blackstone", 8: "bedrock", 9: "nether_wart_block", 10: "warped_wart_block", 11: "warped_nylium", 12: "lava", 28: "soul_soil", 29: "warped_nylium"}[mode]; ok {
		support = name
	}
	if mode == 17 {
		support = []string{"bedrock", "magma_block", "soul_sand", "nether_bricks", "nether_brick_fence", "nether_brick_stairs", "nether_wart", "chest", "spawner", "netherrack"}[mod(dx+dz, 10)]
	}
	if p.Y <= floor {
		if p.Y == floor {
			return s(support)
		}
		return s("netherrack")
	}
	if ceiling && p.Y == roof && !(mode == 16 && (dx != 0 || dz != 0)) {
		return s(support)
	}
	if mode == 13 && dy == 0 {
		return s("lava")
	}
	if mode == 14 && dy == 0 {
		return s("water")
	}
	if mode == 26 && dy >= -6 && dy <= 4 && mod(dx*3+dz, 5) == 0 {
		return s("netherrack")
	}
	if mode == 27 && dy == 0 && dx == 0 && dz == 0 {
		return s("glowstone")
	}
	return s("air")
}

func netherDecorationProbeRegion(t *testing.T, origin FeaturePosition, ceiling bool, mode int, radius ...int32) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	r := int32(1)
	if len(radius) == 1 {
		r = radius[0]
	}
	for cz := (origin.Z >> 4) - r; cz <= (origin.Z>>4)+r; cz++ {
		for cx := (origin.X >> 4) - r; cx <= (origin.X>>4)+r; cx++ {
			ch := &NoiseChunkResult{X: cx, Z: cz, MinY: 0, Height: 256, Blocks: make([]NoiseMaterial, 256*256), SurfaceBlocks: map[int]SurfaceBlockState{}}
			for y := int32(0); y < 256; y++ {
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						s := netherDecorationProbeBase(FeaturePosition{cx*16 + x, y, cz*16 + z}, origin, ceiling, mode)
						i, _ := ch.Index(x, y, z)
						switch s.Name {
						case "minecraft:air":
						case "minecraft:netherrack":
							ch.Blocks[i] = NoiseNetherrack
						case "minecraft:stone":
							ch.Blocks[i] = NoiseStone
						case "minecraft:water":
							ch.Blocks[i] = NoiseWater
						case "minecraft:lava":
							ch.Blocks[i] = NoiseLava
						default:
							ch.SurfaceBlocks[i] = s
						}
					}
				}
			}
			chunks = append(chunks, ch)
		}
	}
	region, err := NewDimensionFeatureRegionWithSeaLevel(128, 32, chunks...)
	if err != nil {
		t.Fatal(err)
	}
	region.borrowed = map[featureChunkPosition]bool{}
	for key := range region.chunks {
		region.borrowed[key] = true
	}
	return region
}

func compileNetherDecorationProbePlan(name string) (configuredFeaturePlacer, error) {
	def, err := LoadConfiguredFeatureDefinition(name)
	if err != nil {
		return nil, err
	}
	switch def.Type {
	case "minecraft:twisting_vines":
		return compileTwistingVinesPlan(def)
	case "minecraft:weeping_vines":
		return compileWeepingVinesPlan(def)
	case "minecraft:glowstone_blob":
		return compileGlowstonePlan(def)
	case "minecraft:basalt_pillar":
		return compileBasaltPillarPlan(def)
	case "minecraft:basalt_columns":
		return compileBasaltColumnsPlan(def)
	case "minecraft:delta_feature":
		return compileDeltaPlan(def)
	case "minecraft:netherrack_replace_blobs":
		return compileReplaceBlobsPlan(def)
	}
	return nil, fmt.Errorf("unexpected Nether decoration %s", def.Type)
}

func TestNetherDecorationsAgainstVanillaJAR(t *testing.T) {
	testNetherDecorationVectors(t, "nether_decoration_1.21.1.txt", "nether_decoration", compileNetherDecorationProbePlan, 2240, 1214)
}

func testNetherDecorationVectors(t *testing.T, fileName, prefix string, compiler func(string) (configuredFeaturePlacer, error), wantCases, wantPositive int, placementCounts ...func() [2]int) {
	t.Helper()
	file, err := os.Open("../../docs/worldgen/" + fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	regions := map[string]*OverworldFeatureRegion{}
	parse := func(s string) int64 {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	scanner := bufio.NewScanner(file)
	cases, positive := 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		var wantCounts [2]int
		if len(f) == 12 && len(placementCounts) == 1 {
			wantCounts = [2]int{int(parse(f[7])), int(parse(f[8]))}
			flag := "0"
			if wantCounts[1] > 0 {
				flag = "1"
			}
			f = append(append(append([]string(nil), f[:7]...), flag), f[9:]...)
		}
		if len(f) != 11 || f[0] != prefix {
			t.Fatalf("invalid Nether decoration vector %s", scanner.Text())
		}
		seed, mode := parse(f[2]), int(parse(f[3]))
		origin := FeaturePosition{int32(parse(f[4])), int32(parse(f[5])), int32(parse(f[6]))}
		ceiling := f[1] == "minecraft:weeping_vines" || f[1] == "minecraft:glowstone_extra" || f[1] == "minecraft:basalt_pillar"
		key := fmt.Sprintf("%v/%d", ceiling, mode)
		if regions[key] == nil {
			radius := int32(1)
			if len(placementCounts) == 1 {
				radius = 2
			}
			regions[key] = netherDecorationProbeRegion(t, origin, ceiling, mode, radius)
		}
		region := regions[key]
		region.overlays = nil
		place, err := compiler(f[1])
		if err != nil {
			t.Fatal(err)
		}
		r := NewWorldgenRandom(seed)
		placed, err := place(region, r, origin)
		if err != nil {
			t.Fatal(err)
		}
		if len(placementCounts) == 1 && placementCounts[0]() != wantCounts {
			t.Errorf("%s got candidate/success counts %v want %v", scanner.Text(), placementCounts[0](), wantCounts)
		}
		var output []string
		for key, overlay := range region.overlays {
			for i, s := range overlay.blocks {
				p := FeaturePosition{key.chunkX*16 + int32(i&15), int32(i / 256), key.chunkZ*16 + int32((i>>4)&15)}
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
		if placed != (parse(f[7]) == 1) || len(output) != int(parse(f[8])) || next != parse(f[9]) || hash != f[10] {
			t.Errorf("%s seed=%d mode=%d placed=%v writes=%d next=%d hash=%s; want %s", f[1], seed, mode, placed, len(output), next, hash, scanner.Text())
		}
		cases++
		if placed {
			positive++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != wantCases || positive != wantPositive {
		t.Fatalf("got %d vectors/%d positive, want %d/%d", cases, positive, wantCases, wantPositive)
	}
	t.Logf("matched %d official JAR cases, including %d positive Nether decorations", cases, positive)
}
