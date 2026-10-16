package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func vegetationProbeBase(p FeaturePosition, scenario int) string {
	if scenario == 12 && p.Y == 40 {
		if ((p.X%3)+3)%3 == 0 {
			return "minecraft:water"
		}
		return "minecraft:sand"
	}
	if scenario == 13 && p.Y <= 40 {
		return "minecraft:ice"
	}
	floor := int32(40)
	if scenario == 1 {
		floor = 38 + ((p.X+p.Z)%5+5)%5
	}
	if scenario == 8 {
		floor = 318
	}
	if scenario == 9 {
		floor = -65
	}
	if p.Y <= floor {
		switch scenario {
		case 3:
			return "minecraft:sand"
		case 4:
			return "minecraft:podzol"
		case 5:
			return "minecraft:stone"
		}
		return "minecraft:grass_block"
	}
	if scenario == 6 && p.Y <= 42 {
		return "minecraft:water"
	}
	if scenario == 7 && p.Y == 42 {
		return "minecraft:stone"
	}
	return "minecraft:air"
}

func vegetationProbeRegion(t *testing.T, cx, cz int32, scenario int) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	for dx := int32(-1); dx <= 1; dx++ {
		for dz := int32(-1); dz <= 1; dz++ {
			ch := &NoiseChunkResult{X: cx + dx, Z: cz + dz, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: make(map[int]SurfaceBlockState)}
			for z := int32(0); z < 16; z++ {
				for x := int32(0); x < 16; x++ {
					p := FeaturePosition{X: ch.X*16 + x, Z: ch.Z*16 + z}
					floor := int32(40)
					if scenario == 1 {
						floor = 38 + ((p.X+p.Z)%5+5)%5
					}
					if scenario == 8 {
						floor = 318
					}
					if scenario == 9 {
						floor = -65
					}
					for y := int32(-64); y <= min(floor+3, 319); y++ {
						index, _ := ch.Index(x, y, z)
						if y < floor {
							ch.Blocks[index] = NoiseStone
							continue
						}
						p.Y = y
						name := vegetationProbeBase(p, scenario)
						if name != "minecraft:air" {
							ch.SurfaceBlocks[index] = SurfaceBlockState{Name: name}
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
	return region
}

func TestVegetationAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/vegetation_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		t.Fatal(err)
	}
	features := make(map[string]PlacedFeature)
	biomes := make(map[string]string)
	for _, feature := range order.features {
		features[feature.Name] = feature
	}
	for biome, steps := range order.biomes {
		for step, indices := range steps {
			if step != 9 {
				continue
			}
			for _, index := range indices {
				for _, feature := range order.features {
					if feature.Step == 9 && feature.Index == index {
						biomes[feature.Name] = biome
					}
				}
			}
		}
	}
	// Template regions are immutable; replay each case into a private clone.
	templates := make(map[[3]int32]*OverworldFeatureRegion)
	extra, err := os.Open("../../docs/worldgen/column_vegetation_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer extra.Close()
	scanner := bufio.NewScanner(io.MultiReader(file, extra))
	cases := 0
	successes := make(map[string]int)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 13 || f[0] != "vegetation" {
			t.Fatalf("invalid vegetation vector: %s", scanner.Text())
		}
		parse := func(s string) int64 {
			n, e := strconv.ParseInt(s, 10, 64)
			if e != nil {
				t.Fatal(e)
			}
			return n
		}
		origin := FeaturePosition{int32(parse(f[5])), int32(parse(f[6])), int32(parse(f[7]))}
		key := [3]int32{origin.X >> 4, origin.Z >> 4, int32(parse(f[4]))}
		if templates[key] == nil {
			templates[key] = vegetationProbeRegion(t, key[0], key[1], int(key[2]))
		}
		var chunks []*NoiseChunkResult
		for _, chunk := range templates[key].chunks {
			chunks = append(chunks, cloneTerrain(chunk))
		}
		region, e := NewOverworldFeatureRegion(chunks...)
		if e != nil {
			t.Fatal(e)
		}
		random := NewWorldgenRandom(parse(f[3]))
		var place configuredFeaturePlacer
		var operations []placementOperation
		if f[1] == "configured" {
			place, err = compileFeaturePlan(f[2], 0)
		} else {
			feature := features[f[2]]
			def, e := LoadPlacedFeatureDefinition(f[2])
			if e != nil {
				t.Fatal(e)
			}
			place, err = compileFeaturePlan(def.Feature, 0)
			operations, e = bindPlacement(region, def, &feature, func(int32, int32, int32) (string, error) { return biomes[feature.Name], nil })
			if e != nil {
				t.Fatal(e)
			}
			random = OverworldPlacedFeatureRandom(OverworldDecorationSeed(parse(f[3]), key[0], key[1]), feature)
			origin.Y = 0
		}
		if err != nil {
			t.Fatal(err)
		}
		candidates, placed := 0, 0
		err = walkPlacement(operations, random, origin, func(p FeaturePosition, random *WorldgenRandom) error {
			candidates++
			ok, e := place(region, random, p)
			if ok {
				placed++
			}
			return e
		})
		if err != nil {
			t.Fatal(err)
		}
		var output []string
		for _, ch := range region.chunks {
			for i, state := range ch.SurfaceBlocks {
				p := FeaturePosition{ch.X*16 + int32(i&15), -64 + int32(i/256), ch.Z*16 + int32((i>>4)&15)}
				if state.Name == vegetationProbeBase(p, int(key[2])) {
					continue
				}
				id, e := blockstate.BlockStateID(state.Name, state.Properties)
				if e != nil {
					t.Fatal(e)
				}
				output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
			}
		}
		sort.Strings(output)
		hash := sha256.Sum256([]byte(strings.Join(output, "\n")))
		next := int64(random.NextLong())
		if candidates != int(parse(f[8])) || placed != int(parse(f[9])) || len(output) != int(parse(f[10])) || next != parse(f[11]) || fmt.Sprintf("%x", hash) != f[12] {
			t.Errorf("%s %s seed %s scenario %s: candidates=%d placed=%d changed=%d next=%d hash=%x; want %s", f[1], f[2], f[3], f[4], candidates, placed, len(output), next, hash, scanner.Text())
		}
		if placed > 0 {
			successes[f[2]]++
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 4208 {
		t.Fatalf("got %d cases, want 4208", cases)
	}
	if len(successes) < 25 {
		t.Fatalf("only %d vegetation feature names have successful reference cases", len(successes))
	}
	t.Logf("matched %d JAR cases, %d configured and placed feature names", cases, len(successes))
}
