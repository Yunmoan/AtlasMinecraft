package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

type flatForestRockWorld struct {
	changed  map[FeaturePosition]SurfaceBlockState
	noGround bool
}

func (world *flatForestRockWorld) BlockAt(position FeaturePosition) (SurfaceBlockState, error) {
	if state, found := world.changed[position]; found {
		return state, nil
	}
	if !world.noGround && position.Y <= 39 {
		return SurfaceBlockState{Name: "minecraft:stone"}, nil
	}
	return SurfaceBlockState{Name: "minecraft:air"}, nil
}

func (world *flatForestRockWorld) SetBlock(position FeaturePosition, state SurfaceBlockState) error {
	world.changed[position] = state
	return nil
}

func (world *flatForestRockWorld) InWorld(position FeaturePosition) bool {
	return position.Y >= -64 && position.Y < 320
}

func TestForestRockAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/forest_rock_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type rockCase struct {
		seed   int64
		origin FeaturePosition
	}
	type rockResult struct {
		placed bool
		count  int
		next   int64
	}
	type candidateCase struct {
		seed int64
		x, z int32
	}
	type candidateResult struct {
		count int
		next  int64
	}
	results := make(map[rockCase]rockResult)
	blocks := make(map[rockCase][]FeaturePosition)
	candidates := make(map[candidateCase]candidateResult)
	positions := make(map[candidateCase][]FeaturePosition)
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
		if len(fields) < 5 {
			t.Fatalf("invalid forest rock vector: %q", scanner.Text())
		}
		switch fields[0] {
		case "rock_count", "rock_block":
			if len(fields) != 8 {
				t.Fatalf("invalid rock vector: %q", scanner.Text())
			}
			key := rockCase{parse(fields[1]), FeaturePosition{int32(parse(fields[2])),
				int32(parse(fields[3])), int32(parse(fields[4]))}}
			if fields[0] == "rock_count" {
				results[key] = rockResult{parse(fields[5]) != 0, int(parse(fields[6])), parse(fields[7])}
			} else {
				blocks[key] = append(blocks[key], FeaturePosition{int32(parse(fields[5])),
					int32(parse(fields[6])), int32(parse(fields[7]))})
			}
		case "rock_candidates":
			if len(fields) != 6 {
				t.Fatal(scanner.Text())
			}
			key := candidateCase{parse(fields[1]), int32(parse(fields[2])), int32(parse(fields[3]))}
			candidates[key] = candidateResult{int(parse(fields[4])), parse(fields[5])}
		case "rock_candidate":
			if len(fields) != 7 {
				t.Fatal(scanner.Text())
			}
			key := candidateCase{parse(fields[1]), int32(parse(fields[2])), int32(parse(fields[3]))}
			positions[key] = append(positions[key], FeaturePosition{int32(parse(fields[4])),
				int32(parse(fields[5])), int32(parse(fields[6]))})
		default:
			t.Fatalf("unknown forest rock vector: %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(results) != 20 || len(candidates) != 8 {
		t.Fatalf("got %d rock cases and %d placement cases, want 20 and 8", len(results), len(candidates))
	}
	state, err := LoadForestRockState("minecraft:forest_rock")
	if err != nil || state.Name != "minecraft:mossy_cobblestone" {
		t.Fatalf("forest rock configuration: %+v, %v", state, err)
	}
	for key, want := range results {
		t.Run(fmt.Sprintf("blocks/%d/%+v", key.seed, key.origin), func(t *testing.T) {
			world := &flatForestRockWorld{changed: make(map[FeaturePosition]SurfaceBlockState),
				noGround: key.origin.Y == 100}
			random := NewWorldgenRandom(key.seed)
			placed, err := PlaceForestRock(world, random, key.origin, state)
			if err != nil {
				t.Fatal(err)
			}
			if next := int64(random.NextLong()); placed != want.placed || len(world.changed) != want.count || next != want.next {
				t.Fatalf("got placed=%t count=%d next=%d, want %+v", placed, len(world.changed), next, want)
			}
			for _, position := range blocks[key] {
				if world.changed[position].Name != state.Name {
					t.Fatalf("missing block at %+v", position)
				}
			}
		})
	}
	feature := PlacedFeature{Step: 2, Index: 4, Name: "minecraft:forest_rock"}
	for key, want := range candidates {
		t.Run(fmt.Sprintf("candidates/%d/%d,%d", key.seed, key.x, key.z), func(t *testing.T) {
			heightmap := func(name string, blockX, blockZ int32) (int32, error) {
				if name != "MOTION_BLOCKING" {
					return 0, fmt.Errorf("unexpected heightmap %s", name)
				}
				return 40 + ((blockX ^ blockZ) & 1), nil
			}
			got, random, err := overworldPreBiomeCandidates(key.seed, key.x, key.z, feature, heightmap)
			if err != nil {
				t.Fatal(err)
			}
			if next := int64(random.NextLong()); len(got) != want.count || next != want.next {
				t.Fatalf("got %d candidates and next=%d, want %+v", len(got), next, want)
			}
			if len(positions[key]) != want.count {
				t.Fatalf("JAR listed %d positions, want %d", len(positions[key]), want.count)
			}
			for index, position := range positions[key] {
				if got[index] != position {
					t.Fatalf("candidate %d: got %+v, want %+v", index, got[index], position)
				}
			}
		})
	}
	if _, err := OverworldPreBiomeCandidates(0, 0, 0, feature); err == nil {
		t.Fatal("missing heightmap must be rejected")
	}
}

func TestApplyForestRockRequiresNeighboringChunks(t *testing.T) {
	feature := PlacedFeature{Step: 2, Index: 4, Name: "minecraft:forest_rock"}
	biome := func(int32, int32, int32) (string, error) {
		return "minecraft:old_growth_pine_taiga", nil
	}
	chunk := newFlatForestRockChunk(0, 0)
	region, err := NewOverworldFeatureRegion(chunk)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyOverworldForestRockFeature(region, 0, 0, 0, feature, biome); err == nil {
		t.Fatal("missing neighbors must be rejected")
	}
	if len(chunk.SurfaceBlocks) != 0 {
		t.Fatal("incomplete region was modified")
	}
	var chunks []*NoiseChunkResult
	for chunkX := int32(-1); chunkX <= 1; chunkX++ {
		for chunkZ := int32(-1); chunkZ <= 1; chunkZ++ {
			chunks = append(chunks, newFlatForestRockChunk(chunkX, chunkZ))
		}
	}
	region, err = NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	successes, err := ApplyOverworldForestRockFeature(region, 0, 0, 0, feature, biome)
	if err != nil || successes != 2 {
		t.Fatalf("forest rock placements: %d, %v", successes, err)
	}
	changed := 0
	for _, generated := range chunks {
		changed += len(generated.SurfaceBlocks)
	}
	if changed == 0 {
		t.Fatal("forest rock did not modify the region")
	}
}

func newFlatForestRockChunk(chunkX, chunkZ int32) *NoiseChunkResult {
	chunk := &NoiseChunkResult{X: chunkX, Z: chunkZ, MinY: -64, Height: 384,
		Blocks: make([]NoiseMaterial, 384*256)}
	for y := int32(-64); y <= 39; y++ {
		for index := int32(0); index < 256; index++ {
			chunk.Blocks[(y+64)*256+index] = NoiseStone
		}
	}
	return chunk
}
