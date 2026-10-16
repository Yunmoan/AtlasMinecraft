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

func fossilProbeFloor(x, z int32, mode int) int32 {
	switch mode {
	case 1:
		return 38 + ((x+z)%5+5)%5
	case 7:
		return -65
	case 8:
		return 318
	case 9:
		return -60
	default:
		return 40
	}
}

func fossilProbeState(p, origin FeaturePosition, mode int) SurfaceBlockState {
	name := "minecraft:stone"
	mod := func(n int32) int32 { return (n%3 + 3) % 3 }
	if p.Y < -64 || p.Y >= 320 || p.Y > fossilProbeFloor(p.X, p.Z, mode) {
		name = "minecraft:air"
	} else if mode == 4 || (mode == 3 && (p.X+2*p.Y+3*p.Z)%5 == 0) {
		name = "minecraft:bedrock"
	} else if mode == 5 && p.Y < 35 && mod(p.X+p.Z) == 0 {
		name = "minecraft:water"
	} else if mode == 6 && p.Y < 35 && mod(p.X+p.Z) == 0 {
		name = "minecraft:lava"
	} else if mode == 10 && p.Y < 35 && p.X >= origin.X {
		name = "minecraft:air"
	} else if mode == 11 && p.Y < 35 && (p.X >= origin.X || p.Z >= origin.Z) {
		name = "minecraft:air"
	} else if mode == 12 && p.Y < 35 && mod(p.X+p.Z) == 0 {
		return SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "1"}}
	} else if mode == 13 && p.Y < 35 && mod(p.X+p.Z) == 0 {
		return SurfaceBlockState{Name: "minecraft:oak_slab", Properties: map[string]string{"type": "bottom", "waterlogged": "true"}}
	} else if mode == 14 {
		name = "minecraft:deepslate"
	} else if mode == 15 && (p.X+2*p.Y+3*p.Z)%5 == 0 {
		name = "minecraft:end_portal_frame"
	}
	return SurfaceBlockState{Name: name}
}

func fossilProbeRegion(t *testing.T, origin FeaturePosition, mode int) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	for dx := int32(-1); dx <= 1; dx++ {
		for dz := int32(-1); dz <= 1; dz++ {
			ch := &NoiseChunkResult{X: (origin.X >> 4) + dx, Z: (origin.Z >> 4) + dz, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: make(map[int]SurfaceBlockState)}
			// Canonical states are shared by all positions in the synthetic world.
			states := map[string]SurfaceBlockState{}
			for y := int32(-64); y < 320; y++ {
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						state := fossilProbeState(FeaturePosition{ch.X*16 + x, y, ch.Z*16 + z}, origin, mode)
						index, _ := ch.Index(x, y, z)
						switch state.Name {
						case "minecraft:air":
							continue
						case "minecraft:stone":
							ch.Blocks[index] = NoiseStone
						default:
							key := fmt.Sprint(state)
							canonical, ok := states[key]
							if !ok {
								id, err := blockstate.BlockStateID(state.Name, state.Properties)
								if err != nil {
									t.Fatal(err)
								}
								s, err := blockstate.BlockStateByID(id)
								if err != nil {
									t.Fatal(err)
								}
								canonical = SurfaceBlockState{Name: s.Name, Properties: s.Properties}
								states[key] = canonical
							}
							ch.SurfaceBlocks[index] = canonical
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
	return region
}

func fossilProbeHash(t *testing.T, region *OverworldFeatureRegion) (int, string) {
	t.Helper()
	var output []string
	for cp, overlay := range region.overlays {
		for index, state := range overlay.blocks {
			p := FeaturePosition{cp.chunkX*16 + int32(index%16), -64 + int32(index/256), cp.chunkZ*16 + int32((index%256)/16)}
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
			output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
		}
	}
	sort.Strings(output)
	return len(output), fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(output, "\n"))))
}

func TestFossilFeaturesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/fossil_feature_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	worlds := map[int]*OverworldFeatureRegion{}
	plans := map[string]configuredFeaturePlacer{}
	rotations := map[int32]bool{}
	selections := map[int32]bool{}
	pairs := map[[2]int32]bool{}
	cases := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 11 || fields[0] != "fossil_case" {
			t.Fatalf("invalid vector %s", scanner.Text())
		}
		parse := func(i int) int64 {
			n, err := strconv.ParseInt(fields[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		mode := int(parse(3))
		origin := FeaturePosition{int32(parse(4)), int32(parse(5)), int32(parse(6))}
		region := worlds[mode]
		if region == nil {
			region = fossilProbeRegion(t, origin, mode)
			worlds[mode] = region
		}
		region.overlays = nil
		placer := plans[fields[1]]
		if placer == nil {
			placer, err = compileFeaturePlan(fields[1], 0)
			if err != nil {
				t.Fatal(err)
			}
			plans[fields[1]] = placer
		}
		probeRandom := NewWorldgenRandom(parse(2))
		rotation, index := probeRandom.NextIntBound(4), probeRandom.NextIntBound(8)
		rotations[rotation] = true
		selections[index] = true
		pairs[[2]int32{rotation, index}] = true
		r := NewWorldgenRandom(parse(2))
		success, err := placer(region, r, origin)
		if err != nil {
			t.Fatal(err)
		}
		count, hash := fossilProbeHash(t, region)
		next := r.NextLong()
		if success != (parse(7) == 1) || count != int(parse(8)) || next != uint64(parse(9)) || hash != fields[10] {
			t.Fatalf("%s seed=%s scenario=%d rotation=%d template=%d: success=%t count=%d next=%d hash=%s; vector=%s", fields[1], fields[2], mode, rotation, index, success, count, next, hash, scanner.Text())
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 4096 || len(rotations) != 4 || len(selections) != 8 || len(pairs) != 32 {
		t.Fatalf("incomplete fossil coverage: cases=%d rotations=%d templates=%d pairs=%d", cases, len(rotations), len(selections), len(pairs))
	}
}
