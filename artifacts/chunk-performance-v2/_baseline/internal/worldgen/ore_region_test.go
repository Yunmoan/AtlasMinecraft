package worldgen

import (
	world "atlas/internal/blockstate"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func newStoneNoiseChunk(chunkX, chunkZ int32) *NoiseChunkResult {
	chunk := &NoiseChunkResult{X: chunkX, Z: chunkZ, MinY: -64, Height: 384,
		Blocks: make([]NoiseMaterial, 384*256)}
	for index := range chunk.Blocks {
		chunk.Blocks[index] = NoiseStone
	}
	return chunk
}

func TestOverworldFeatureRegionCrossChunkBlocks(t *testing.T) {
	left := newStoneNoiseChunk(-1, -1)
	right := newStoneNoiseChunk(0, 0)
	region, err := NewOverworldFeatureRegion(left, right)
	if err != nil {
		t.Fatal(err)
	}
	position := FeaturePosition{-1, 20, -1}
	if err := region.SetBlock(position, SurfaceBlockState{Name: "minecraft:dirt"}); err != nil {
		t.Fatal(err)
	}
	state, err := region.BlockAt(position)
	if err != nil || state.Name != "minecraft:dirt" {
		t.Fatalf("cross-chunk block: %+v, %v", state, err)
	}
	if len(left.SurfaceBlocks) != 1 || len(right.SurfaceBlocks) != 0 {
		t.Fatalf("cross-chunk write reached wrong chunk: %d, %d", len(left.SurfaceBlocks), len(right.SurfaceBlocks))
	}
	initialHeight, err := region.OceanFloorHeightAt(0, 0)
	if err != nil || initialHeight != 320 {
		t.Fatalf("initial ocean floor height: %d, %v", initialHeight, err)
	}
	if err := region.SetBlock(FeaturePosition{0, 319, 0}, SurfaceBlockState{Name: "minecraft:water"}); err != nil {
		t.Fatal(err)
	}
	height, err := region.OceanFloorHeightAt(0, 0)
	if err != nil || height != 319 {
		t.Fatalf("ocean floor height: %d, %v", height, err)
	}
}

func TestApplyOverworldOreFeatureNeedsNeighbors(t *testing.T) {
	chunk := newStoneNoiseChunk(0, 0)
	region, err := NewOverworldFeatureRegion(chunk)
	if err != nil {
		t.Fatal(err)
	}
	feature := PlacedFeature{Step: 6, Index: 0, Name: "minecraft:ore_dirt"}
	plains := func(int32, int32, int32) (string, error) { return "minecraft:plains", nil }
	if _, err := ApplyOverworldOreFeature(region, 0, 0, 0, feature, plains); err == nil {
		t.Fatal("missing neighboring chunks should reject ore placement")
	}
	if len(chunk.SurfaceBlocks) != 0 {
		t.Fatal("incomplete neighborhood partially modified the chunk")
	}
	var chunks []*NoiseChunkResult
	for chunkX := int32(-1); chunkX <= 1; chunkX++ {
		for chunkZ := int32(-1); chunkZ <= 1; chunkZ++ {
			chunks = append(chunks, newStoneNoiseChunk(chunkX, chunkZ))
		}
	}
	region, err = NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	successes, err := ApplyOverworldOreFeature(region, 0, 0, 0, feature, plains)
	if err != nil {
		t.Fatal(err)
	}
	if successes == 0 {
		t.Fatal("stone region received no dirt veins")
	}
	changed := 0
	for _, chunk := range chunks {
		changed += len(chunk.SurfaceBlocks)
	}
	if changed == 0 {
		t.Fatal("ore feature did not change any chunk blocks")
	}
}

func TestOreFeatureCrossesChunkBoundary(t *testing.T) {
	var chunks []*NoiseChunkResult
	for chunkX := int32(-1); chunkX <= 1; chunkX++ {
		for chunkZ := int32(-1); chunkZ <= 1; chunkZ++ {
			chunks = append(chunks, newStoneNoiseChunk(chunkX, chunkZ))
		}
	}
	region, err := NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	config := OreConfiguration{Size: 17, Targets: []OreTarget{{
		Predicate: "minecraft:block_match", Block: "minecraft:stone",
		State: SurfaceBlockState{Name: "minecraft:dirt"},
	}}}
	placed, err := PlaceOreFeature(region, NewWorldgenRandom(0), FeaturePosition{15, 20, 15}, config)
	if err != nil || !placed {
		t.Fatalf("cross-chunk ore placement: %t, %v", placed, err)
	}
	neighborBlocks := 0
	for _, chunk := range chunks {
		if chunk.X != 0 || chunk.Z != 0 {
			neighborBlocks += len(chunk.SurfaceBlocks)
		}
	}
	if neighborBlocks == 0 {
		t.Fatal("ore vein did not modify a neighboring chunk")
	}
}

func TestOverworldOreHeightAgainstVanillaJAR(t *testing.T) {
	heightFile, err := os.Open("../../docs/worldgen/ore_height_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer heightFile.Close()
	type chunkKey struct {
		seed           int64
		chunkX, chunkZ int32
	}
	expected := make(map[chunkKey][16][16]int32)
	parse := func(field string) int64 {
		value, err := strconv.ParseInt(field, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	scanner := bufio.NewScanner(heightFile)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 7 || fields[0] != "ore_height" {
			t.Fatalf("invalid JAR height vector: %q", scanner.Text())
		}
		key := chunkKey{parse(fields[1]), int32(parse(fields[2])), int32(parse(fields[3]))}
		heights := expected[key]
		heights[parse(fields[4])][parse(fields[5])] = int32(parse(fields[6]))
		expected[key] = heights
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 3072 || len(expected) != 12 {
		t.Fatalf("got %d heights in %d chunks, want 3072 in 12", count, len(expected))
	}
	carvedFile, err := os.Open("../../docs/worldgen/carver_chunk_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer carvedFile.Close()
	scanner = bufio.NewScanner(carvedFile)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 13 || fields[0] != "carved_chunk" {
			t.Fatalf("invalid carved chunk vector")
		}
		key := chunkKey{parse(fields[1]), int32(parse(fields[2])), int32(parse(fields[3]))}
		compressed, err := base64.StdEncoding.DecodeString(fields[12])
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		blocks, err := io.ReadAll(reader)
		reader.Close()
		if err != nil || len(blocks) != 384*256*2 {
			t.Fatalf("invalid carved chunk block data: %d bytes, %v", len(blocks), err)
		}
		for blockX := int32(0); blockX < 16; blockX++ {
			for blockZ := int32(0); blockZ < 16; blockZ++ {
				height := int32(-64)
				for blockY := int32(319); blockY >= -64; blockY-- {
					index := int((blockY+64)*256 + blockZ*16 + blockX)
					stateID := binary.BigEndian.Uint16(blocks[index*2 : index*2+2])
					state, err := world.BlockStateByID(uint32(stateID))
					if err != nil {
						t.Fatal(err)
					}
					if surfaceSolid(SurfaceBlockState{Name: state.Name}) {
						height = blockY + 1
						break
					}
				}
				if height != expected[key][blockX][blockZ] {
					t.Fatalf("chunk %+v column %d,%d: Go ore height %d, JAR %d", key,
						blockX, blockZ, height, expected[key][blockX][blockZ])
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}
