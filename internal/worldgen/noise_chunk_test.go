package worldgen_test

import (
	"atlas/internal/world"
	"atlas/internal/worldgen"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func openNoiseReference(t *testing.T) *bufio.Scanner {
	t.Helper()
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "noise_chunk_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { file.Close() })
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	return scanner
}

func TestNoiseChunkDensityReference(t *testing.T) {
	settings, err := worldgen.LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	final := &worldgen.DensityNode{Kind: "cache_all_in_cell", Children: map[string]*worldgen.DensityNode{
		"argument": settings.Router["final_density"],
	}}
	scanner := openNoiseReference(t)
	count := 0
	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "density ") {
			continue
		}
		var label string
		var seed int64
		var cx, cz, y int32
		var bits uint64
		if _, err := fmt.Sscan(scanner.Text(), &label, &seed, &cx, &cz, &y, &bits); err != nil {
			t.Fatal(err)
		}
		sampler := worldgen.NewNoiseChunkDensitySampler(seed, settings, cx, cz)
		value, err := sampler.Sample(final, cx*16+5, y, cz*16+9)
		if err != nil || math.Float64bits(value) != bits {
			t.Errorf("seed %d chunk %d,%d Y %d: bits %d, want %d (%v)", seed, cx, cz, y, math.Float64bits(value), bits, err)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 72 {
		t.Fatalf("density reference count %d, want 72", count)
	}
}

func TestNoiseChunksAgainstVanillaJAR(t *testing.T) {
	testNoiseChunksAgainstVanillaJAR(t, "noise_chunk_1.21.1.txt", 12)
}

func TestMountainPlateauCaveNoiseAgainstVanillaJAR(t *testing.T) {
	testNoiseChunksAgainstVanillaJAR(t, "terrain_noise_1.21.1.txt", terrainCaseCount(t))
}

func terrainCaseCount(t *testing.T) int {
	t.Helper()
	data, err := os.ReadFile("../../docs/worldgen/terrain_cases_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			count++
		}
	}
	if count < 31 {
		t.Fatalf("terrain suite lost coverage: %d cases", count)
	}
	return count
}

func testNoiseChunksAgainstVanillaJAR(t *testing.T, fixture string, wantCount int) {
	var ids [10]uint32
	for index := range ids {
		state, err := world.BlockStateID(worldgen.NoiseMaterial(index).BlockName(), nil)
		if err != nil {
			t.Fatal(err)
		}
		ids[index] = state
	}
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", fixture))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || fields[0] != "chunk" {
			continue
		}
		if len(fields) != 5 {
			t.Fatal("invalid chunk reference")
		}
		seed, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		cx, err := strconv.ParseInt(fields[2], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		cz, err := strconv.ParseInt(fields[3], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		compressed, err := base64.StdEncoding.DecodeString(fields[4])
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		expected, err := io.ReadAll(io.LimitReader(reader, 16*16*384*3+1))
		reader.Close()
		if err != nil || len(expected) != 16*16*384*3 {
			t.Fatalf("invalid decompressed reference: %d bytes (%v)", len(expected), err)
		}
		t.Run(fmt.Sprintf("%d/%d,%d", seed, cx, cz), func(t *testing.T) {
			generator, err := worldgen.NewOverworldNoiseGenerator(seed)
			if err != nil {
				t.Fatal(err)
			}
			chunk, err := generator.GenerateNoise(int32(cx), int32(cz))
			if err != nil {
				t.Fatal(err)
			}
			mismatches := 0
			for index, material := range chunk.Blocks {
				want := uint32(binary.BigEndian.Uint16(expected[index*3:]))
				wantUpdate := expected[index*3+2] != 0
				if ids[material] != want || chunk.FluidUpdates[index] != wantUpdate {
					if mismatches < 8 {
						t.Errorf("%d,%d,%d: state %d update %v, want state %d update %v", index%16, index/256-64, index/16%16,
							ids[material], chunk.FluidUpdates[index], want, wantUpdate)
					}
					mismatches++
				}
			}
			if mismatches != 0 {
				t.Errorf("%d of %d positions differ from the supplied vanilla JAR", mismatches, len(chunk.Blocks))
			}
		})
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != wantCount {
		t.Fatalf("chunk reference count %d, want %d", count, wantCount)
	}
}
