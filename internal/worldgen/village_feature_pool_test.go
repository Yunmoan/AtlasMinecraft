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

func villageFeatureProbeRegion(t *testing.T, origin FeaturePosition, mode int) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	for cx := (origin.X >> 4) - 1; cx <= (origin.X>>4)+1; cx++ {
		for cz := (origin.Z >> 4) - 1; cz <= (origin.Z>>4)+1; cz++ {
			chunk := &NoiseChunkResult{X: cx, Z: cz, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304), SurfaceBlocks: make(map[int]SurfaceBlockState)}
			for x := int32(0); x < 16; x++ {
				for z := int32(0); z < 16; z++ {
					height := int32(40)
					if mode == 7 {
						height += (cx*16 + x - origin.X) >> 2
					}
					for y := int32(-64); y <= min(319, max(height, 43)); y++ {
						state := SurfaceBlockState{Name: "minecraft:air"}
						switch {
						case y == height:
							state.Name = "minecraft:grass_block"
							if mode == 3 {
								state.Name = "minecraft:sand"
							} else if mode == 4 {
								state.Name = "minecraft:stone"
							}
						case y < height:
							state.Name = "minecraft:dirt"
						case (mode == 5 || mode == 6) && y <= 43:
							state.Name = "minecraft:water"
							if mode == 6 {
								state.Properties = map[string]string{"level": "1"}
							}
						}
						index, _ := chunk.Index(x, y, z)
						switch state.Name {
						case "minecraft:air":
						case "minecraft:stone":
							chunk.Blocks[index] = NoiseStone
						case "minecraft:water":
							chunk.Blocks[index] = NoiseWater
							if state.Properties != nil {
								chunk.SurfaceBlocks[index] = state
							}
						default:
							chunk.SurfaceBlocks[index] = state
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
	region.borrowed = make(map[featureChunkPosition]bool)
	for position := range region.chunks {
		region.borrowed[position] = true
	}
	return region
}

func TestVillageFeaturePoolsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/village_feature_pool_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	regions := make(map[int]*OverworldFeatureRegion)
	features := make(map[string]bool)
	cases, crossChunk, successes, maxRadius := 0, 0, 0, 0
	hash := func(rows []string) string {
		sort.Strings(rows)
		return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 15 || fields[0] != "village_feature" {
			t.Fatalf("invalid village feature vector %s", scanner.Text())
		}
		parse := func(i int) int64 {
			value, err := strconv.ParseInt(fields[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		origin := FeaturePosition{int32(parse(4)), int32(parse(5)), int32(parse(6))}
		mode := int(parse(3))
		region := regions[mode]
		if region == nil {
			region = villageFeatureProbeRegion(t, origin, mode)
			regions[mode] = region
		}
		region.overlays = nil
		region.blockEntities = nil
		region.heights = make(map[featureColumnPosition]int32)
		region.motionHeights = make(map[featureColumnPosition]int32)
		region.surfaceHeights = make(map[featureColumnPosition]int32)
		result := &structureBlockResult{blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
		world := &jigsawFeatureProbeWorld{structureRecordingWorld: &structureRecordingWorld{region: region, result: result}}
		raw, _ := json.Marshal(fields[1])
		element := JigsawPoolElement{Type: "minecraft:feature_pool_element", Projection: "rigid", Feature: raw}
		random := NewWorldgenRandom(parse(2))
		_, err := placeJigsawElement(world, element, origin, FeaturePosition{999, 999, 999}, TemplateRotation(parse(12)), "ignored", random, StructureBoundingBox{256, -64, 256, 271, 319, 271})
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		var rows, marked []string
		outside, radius := false, 0
		for p, state := range result.blocks {
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
			outside = outside || p.X>>4 != origin.X>>4 || p.Z>>4 != origin.Z>>4
			radius = max(radius, int(max(abs32(p.X-origin.X), abs32(p.Z-origin.Z))))
		}
		for p := range result.postProcessing {
			marked = append(marked, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
		}
		if world.placed != (parse(7) != 0) || len(rows) != int(parse(8)) || int64(random.NextLong()) != parse(9) || hash(rows) != fields[10] || len(marked) != int(parse(11)) || radius != int(parse(13)) || hash(marked) != fields[14] {
			t.Fatalf("%s: placed=%v writes=%d hash=%s marked=%d markHash=%s radius=%d", scanner.Text(), world.placed, len(rows), hash(rows), len(marked), hash(marked), radius)
		}
		if region.structureRecording != nil {
			t.Fatal("village feature recorder leaked into later decoration")
		}
		features[fields[1]] = true
		cases++
		if outside {
			crossChunk++
		}
		if world.placed {
			successes++
		}
		maxRadius = max(maxRadius, radius)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 3120 || len(features) != 13 || crossChunk == 0 || successes == 0 || maxRadius > 15 {
		t.Fatalf("village features: cases=%d names=%d crossChunk=%d successes=%d maxRadius=%d", cases, len(features), crossChunk, successes, maxRadius)
	}
	t.Logf("matched %d official calls across %d village features; %d successful, %d cross-chunk outputs, maximum observed write radius %d; disjoint clip/pivot and all four rotations", cases, len(features), successes, crossChunk, maxRadius)
}
