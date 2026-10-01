package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

type waterBlueIceWorld struct {
	origin    FeaturePosition
	supported bool
	changed   map[FeaturePosition]SurfaceBlockState
}

func (world *waterBlueIceWorld) BlockAt(position FeaturePosition) (SurfaceBlockState, error) {
	if state, found := world.changed[position]; found {
		return state, nil
	}
	if world.supported && position == (FeaturePosition{world.origin.X + 1,
		world.origin.Y, world.origin.Z}) {
		return SurfaceBlockState{Name: "minecraft:packed_ice"}, nil
	}
	if position.Y >= 30 && position.Y <= 61 {
		return SurfaceBlockState{Name: "minecraft:water"}, nil
	}
	return SurfaceBlockState{Name: "minecraft:air"}, nil
}

func (world *waterBlueIceWorld) SetBlock(position FeaturePosition, state SurfaceBlockState) error {
	world.changed[position] = state
	return nil
}

func (world *waterBlueIceWorld) InWorld(position FeaturePosition) bool {
	return position.Y >= -64 && position.Y < 320
}

func TestBlueIceAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/blue_ice_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type blueCase struct {
		seed      int64
		origin    FeaturePosition
		supported bool
	}
	type blueResult struct {
		placed bool
		count  int
		next   int64
	}
	results := make(map[blueCase]blueResult)
	blocks := make(map[blueCase][]FeaturePosition)
	parse := func(value string) int64 {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 9 {
			t.Fatalf("invalid JAR blue ice vector: %q", scanner.Text())
		}
		key := blueCase{parse(fields[1]), FeaturePosition{int32(parse(fields[2])),
			int32(parse(fields[3])), int32(parse(fields[4]))}, parse(fields[5]) != 0}
		switch fields[0] {
		case "blue_count":
			results[key] = blueResult{parse(fields[6]) != 0, int(parse(fields[7])), parse(fields[8])}
		case "blue_block":
			blocks[key] = append(blocks[key], FeaturePosition{int32(parse(fields[6])),
				int32(parse(fields[7])), int32(parse(fields[8]))})
		default:
			t.Fatalf("unknown JAR blue ice vector: %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(results) != 16 {
		t.Fatalf("got %d JAR blue ice cases, want 16", len(results))
	}
	for key, want := range results {
		t.Run(fmt.Sprintf("%d/%+v/%t", key.seed, key.origin, key.supported), func(t *testing.T) {
			world := &waterBlueIceWorld{origin: key.origin, supported: key.supported,
				changed: make(map[FeaturePosition]SurfaceBlockState)}
			random := NewWorldgenRandom(key.seed)
			placed, err := PlaceBlueIce(world, random, key.origin)
			if err != nil {
				t.Fatal(err)
			}
			if next := int64(random.NextLong()); placed != want.placed || len(world.changed) != want.count || next != want.next {
				t.Fatalf("got placed=%t count=%d next=%d, want %+v", placed,
					len(world.changed), next, want)
			}
			if len(blocks[key]) != want.count {
				t.Fatalf("JAR listed %d blocks, want %d", len(blocks[key]), want.count)
			}
			for _, position := range blocks[key] {
				if world.changed[position].Name != "minecraft:blue_ice" {
					t.Fatalf("missing blue ice at %+v", position)
				}
			}
		})
	}
}

func TestApplyBlueIceRequiresNeighboringChunks(t *testing.T) {
	feature := PlacedFeature{Step: 4, Index: 3, Name: "minecraft:blue_ice"}
	biome := func(int32, int32, int32) (string, error) {
		return "minecraft:frozen_ocean", nil
	}
	region, err := NewOverworldFeatureRegion(newWaterBlueIceChunk(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyOverworldBlueIceFeature(region, 0, 0, 0, feature, biome); err == nil {
		t.Fatal("missing neighboring chunks must be rejected")
	}
	if len(region.chunks[featureChunkPosition{0, 0}].SurfaceBlocks) != 0 {
		t.Fatal("incomplete region was modified")
	}
	var chunks []*NoiseChunkResult
	for chunkX := int32(-1); chunkX <= 1; chunkX++ {
		for chunkZ := int32(-1); chunkZ <= 1; chunkZ++ {
			chunks = append(chunks, newWaterBlueIceChunk(chunkX, chunkZ))
		}
	}
	region, err = NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	// The JAR's first lazy-stream candidate for seed/chunk 0 is (7,31,9).
	// Candidate-list helpers deliberately collect modifiers eagerly and
	// cannot be used to prepare the runtime placement world's random path.
	if err := region.SetBlock(FeaturePosition{8, 31, 9}, SurfaceBlockState{Name: "minecraft:packed_ice"}); err != nil {
		t.Fatal(err)
	}
	successes, err := ApplyOverworldBlueIceFeature(region, 0, 0, 0, feature, biome)
	if err != nil || successes == 0 {
		t.Fatalf("blue ice placements: %d, %v", successes, err)
	}
	blueCount := 0
	for _, chunk := range chunks {
		for _, state := range chunk.SurfaceBlocks {
			if state.Name == "minecraft:blue_ice" {
				blueCount++
			}
		}
	}
	if blueCount == 0 {
		t.Fatal("blue ice did not modify the region")
	}
}

func newWaterBlueIceChunk(chunkX, chunkZ int32) *NoiseChunkResult {
	chunk := &NoiseChunkResult{X: chunkX, Z: chunkZ, MinY: -64, Height: 384,
		Blocks: make([]NoiseMaterial, 384*256)}
	for y := int32(30); y <= 61; y++ {
		for index := int32(0); index < 256; index++ {
			chunk.Blocks[(y+64)*256+index] = NoiseWater
		}
	}
	return chunk
}
