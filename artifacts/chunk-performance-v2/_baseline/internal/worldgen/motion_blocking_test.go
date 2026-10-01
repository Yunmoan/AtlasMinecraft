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

func TestMotionBlockingStatesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/motion_blocking_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatal("missing JAR motion blocking state vector")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) != 3 || fields[0] != "motion_blocking_states" || fields[1] != "26684" {
		t.Fatalf("invalid JAR motion blocking header: %q", scanner.Text())
	}
	flags, err := base64.StdEncoding.DecodeString(fields[2])
	if err != nil || len(flags) != (26684+7)/8 {
		t.Fatalf("invalid JAR motion blocking flags: %d bytes, %v", len(flags), err)
	}
	if !scanner.Scan() {
		t.Fatal("missing JAR ocean floor state vector")
	}
	oceanFields := strings.Fields(scanner.Text())
	if len(oceanFields) != 3 || oceanFields[0] != "ocean_floor_states" || oceanFields[1] != "26684" {
		t.Fatalf("invalid JAR ocean floor header: %q", scanner.Text())
	}
	oceanFlags, err := base64.StdEncoding.DecodeString(oceanFields[2])
	if err != nil || len(oceanFlags) != len(flags) {
		t.Fatalf("invalid JAR ocean floor flags: %d bytes, %v", len(oceanFlags), err)
	}
	for stateID := uint32(0); stateID < 26684; stateID++ {
		blockState, err := world.BlockStateByID(stateID)
		if err != nil {
			t.Fatal(err)
		}
		got, err := motionBlocking(SurfaceBlockState{Name: blockState.Name,
			Properties: blockState.Properties})
		if err != nil {
			t.Fatalf("state %d: %v", stateID, err)
		}
		want := flags[stateID/8]&(1<<(stateID%8)) != 0
		if got != want {
			t.Fatalf("state %d %s %+v: Go=%t JAR=%t", stateID,
				blockState.Name, blockState.Properties, got, want)
		}
		gotOcean, err := oceanFloorSolid(SurfaceBlockState{Name: blockState.Name,
			Properties: blockState.Properties})
		if err != nil {
			t.Fatalf("ocean floor state %d: %v", stateID, err)
		}
		wantOcean := oceanFlags[stateID/8]&(1<<(stateID%8)) != 0
		if gotOcean != wantOcean {
			t.Fatalf("ocean floor state %d %s %+v: Go=%t JAR=%t", stateID,
				blockState.Name, blockState.Properties, gotOcean, wantOcean)
		}
	}
	if _, err := motionBlocking(SurfaceBlockState{Name: "minecraft:not_a_block"}); err == nil {
		t.Fatal("unknown block must not silently count as solid")
	}
	if _, err := motionBlocking(SurfaceBlockState{Name: "minecraft:glow_lichen",
		Properties: map[string]string{"waterlogged": "invalid"}}); err == nil {
		t.Fatal("invalid waterlogged property must be rejected")
	}
	for name, kind := range motionBlockingKinds {
		if !strings.HasPrefix(kind.Motion, "waterlogged_") {
			continue
		}
		stateID, err := world.BlockStateID(name, nil)
		if err != nil {
			t.Fatal(err)
		}
		got, err := motionBlocking(SurfaceBlockState{Name: name})
		if err != nil || got != (flags[stateID/8]&(1<<(stateID%8)) != 0) {
			t.Fatalf("default state %s: Go=%t JAR state %d, error=%v", name, got, stateID, err)
		}
	}
}

func TestOceanFloorRegionHeightInvalidation(t *testing.T) {
	region, err := NewOverworldFeatureRegion(newFlatForestRockChunk(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	check := func(want int32) {
		t.Helper()
		got, err := region.OceanFloorHeightAt(0, 0)
		if err != nil || got != want {
			t.Fatalf("ocean floor height: got %d, want %d, error=%v", got, want, err)
		}
	}
	set := func(y int32, name string) {
		t.Helper()
		if err := region.SetBlock(FeaturePosition{X: 0, Y: y, Z: 0},
			SurfaceBlockState{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	check(40)
	set(50, "minecraft:short_grass")
	check(40)
	set(50, "minecraft:stone")
	check(51)
	set(55, "minecraft:water")
	check(51)
	set(55, "minecraft:unknown")
	if _, err := region.OceanFloorHeightAt(0, 0); err == nil {
		t.Fatal("unknown block must not be silently accepted by ocean floor heightmap")
	}
}

func TestMotionBlockingHeightsAgainstVanillaJAR(t *testing.T) {
	type chunkKey struct {
		seed int64
		x, z int32
	}
	parse := func(value string) int64 {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
	}
	file, err := os.Open("../../docs/worldgen/motion_blocking_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	expected := make(map[chunkKey][16][16]int32)
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || fields[0] != "motion_blocking_height" {
			continue
		}
		if len(fields) != 7 {
			t.Fatalf("invalid JAR motion blocking height: %q", scanner.Text())
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
		t.Fatalf("got %d height samples in %d chunks, want 3072 in 12", count, len(expected))
	}
	carved, err := os.Open("../../docs/worldgen/carver_chunk_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer carved.Close()
	scanner = bufio.NewScanner(carved)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 13 || fields[0] != "carved_chunk" {
			t.Fatal("invalid carved chunk reference vector")
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
					blockState, err := world.BlockStateByID(uint32(stateID))
					if err != nil {
						t.Fatal(err)
					}
					blocksMotion, err := motionBlocking(SurfaceBlockState{Name: blockState.Name,
						Properties: blockState.Properties})
					if err != nil {
						t.Fatal(err)
					}
					if blocksMotion {
						height = blockY + 1
						break
					}
				}
				if height != expected[key][blockX][blockZ] {
					t.Fatalf("chunk %+v column %d,%d: Go=%d JAR=%d", key,
						blockX, blockZ, height, expected[key][blockX][blockZ])
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestMotionBlockingRegionHeightInvalidation(t *testing.T) {
	region, err := NewOverworldFeatureRegion(newFlatForestRockChunk(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := region.HeightmapAt("MOTION_BLOCKING_NO_LEAVES", 0, 0); err == nil {
		t.Fatal("unsupported heightmap must be rejected")
	}
	check := func(want int32) {
		t.Helper()
		got, err := region.HeightmapAt("MOTION_BLOCKING", 0, 0)
		if err != nil || got != want {
			t.Fatalf("motion blocking height: got %d, want %d, error=%v", got, want, err)
		}
	}
	set := func(y int32, state SurfaceBlockState) {
		t.Helper()
		if err := region.SetBlock(FeaturePosition{X: 0, Y: y, Z: 0}, state); err != nil {
			t.Fatal(err)
		}
	}
	check(40)
	set(50, SurfaceBlockState{Name: "minecraft:glow_lichen"})
	check(40)
	set(50, SurfaceBlockState{Name: "minecraft:glow_lichen",
		Properties: map[string]string{"waterlogged": "true"}})
	check(51)
	set(55, SurfaceBlockState{Name: "minecraft:water"})
	check(56)
	set(55, SurfaceBlockState{Name: "minecraft:air"})
	check(51)
	set(60, SurfaceBlockState{Name: "minecraft:unknown"})
	if _, err := region.HeightmapAt("MOTION_BLOCKING", 0, 0); err == nil {
		t.Fatal("unknown block must not be silently accepted by heightmap")
	}
}
