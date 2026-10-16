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

func netherForestProbeBase(p, origin FeaturePosition, nylium string, mode int) SurfaceBlockState {
	s := func(name string) SurfaceBlockState { return SurfaceBlockState{Name: "minecraft:" + name} }
	if p.Y < 0 || p.Y >= 256 || mode == 7 {
		return s("air")
	}
	dx, dy, dz := p.X-origin.X, p.Y-origin.Y, p.Z-origin.Z
	floor := origin.Y - 1
	if mode == 13 {
		floor += ((dx+dz)%3 + 3) % 3
	}
	if p.Y <= floor {
		if p.Y < floor {
			return s("netherrack")
		}
		switch mode {
		case 3:
			return s("netherrack")
		case 4:
			if nylium == "crimson_nylium" {
				return s("warped_nylium")
			}
			return s("crimson_nylium")
		case 8:
			return s("bedrock")
		case 21:
			if dx != 0 || dz != 0 {
				return s([]string{nylium, "dirt", "soul_soil", "mycelium", "netherrack"}[((dx+dz)%5+5)%5])
			}
		}
		return s(nylium)
	}
	switch mode {
	case 5:
		if dy == 0 {
			return s("water")
		}
	case 6:
		if dy >= 0 && dy <= 5 {
			return s("crimson_roots")
		}
	case 9, 10, 18, 22:
		if dy >= 2 && abs32(dx) <= 2 && abs32(dz) <= 2 {
			return s(map[int]string{9: "netherrack", 10: "nether_wart_block", 18: "stone", 22: "short_grass"}[mode])
		}
	case 11:
		if dy >= 0 && dy <= 12 {
			return s("lava")
		}
	case 12:
		if dy == 0 {
			return s("warped_fungus")
		}
	case 19:
		if dx == 0 && dy == 0 && dz == 0 {
			return s("crimson_fungus")
		}
	case 20:
		if dy == 3 {
			return s("netherrack")
		}
	}
	return s("air")
}

func netherForestProbeRegion(t *testing.T, origin FeaturePosition, nylium string, mode int) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	for cz := (origin.Z >> 4) - 1; cz <= (origin.Z>>4)+1; cz++ {
		for cx := (origin.X >> 4) - 1; cx <= (origin.X>>4)+1; cx++ {
			ch := &NoiseChunkResult{X: cx, Z: cz, MinY: 0, Height: 256, Blocks: make([]NoiseMaterial, 256*256), SurfaceBlocks: map[int]SurfaceBlockState{}}
			for y := int32(0); y < 256; y++ {
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						s := netherForestProbeBase(FeaturePosition{cx*16 + x, y, cz*16 + z}, origin, nylium, mode)
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
	region, err := NewDimensionFeatureRegion(128, chunks...)
	if err != nil {
		t.Fatal(err)
	}
	region.borrowed = map[featureChunkPosition]bool{}
	for key := range region.chunks {
		region.borrowed[key] = true
	}
	return region
}

func TestNetherForestFeaturesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/nether_forest_1.21.1.txt")
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
		if len(f) != 11 || f[0] != "nether_forest" {
			t.Fatalf("invalid Nether forest vector %s", scanner.Text())
		}
		seed, mode := parse(f[2]), int(parse(f[3]))
		origin := FeaturePosition{int32(parse(f[4])), int32(parse(f[5])), int32(parse(f[6]))}
		nylium := "crimson_nylium"
		if strings.Contains(f[1], "warped") {
			nylium = "warped_nylium"
		}
		key := fmt.Sprintf("%s/%d", nylium, mode)
		if regions[key] == nil {
			regions[key] = netherForestProbeRegion(t, origin, nylium, mode)
		}
		region := regions[key]
		region.overlays = nil
		place, err := compileFeaturePlan(f[1], 0)
		if err != nil {
			t.Fatal(err)
		}
		r := NewWorldgenRandom(seed)
		placed, err := place(region, r, origin)
		if err != nil {
			t.Fatal(err)
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
	if cases != 1840 || positive != 1264 {
		t.Fatalf("got %d vectors/%d positive cases, want 1840/1264", cases, positive)
	}
	t.Logf("matched %d official JAR cases, including %d positive forest/fungus cases", cases, positive)
}
