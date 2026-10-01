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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSurfaceChunksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "surface_chunk_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 6 || fields[0] != "surface_chunk" {
			t.Fatalf("invalid surface chunk vector")
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
		biome := fields[4]
		compressed, err := base64.StdEncoding.DecodeString(fields[5])
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		expected, err := io.ReadAll(io.LimitReader(reader, 16*16*384*2+1))
		reader.Close()
		if err != nil || len(expected) != 16*16*384*2 {
			t.Fatalf("invalid surface chunk reference: %d bytes (%v)", len(expected), err)
		}
		t.Run(fmt.Sprintf("%d/%d,%d/%s", seed, cx, cz, biome), func(t *testing.T) {
			generator, err := worldgen.NewOverworldNoiseGenerator(seed)
			if err != nil {
				t.Fatal(err)
			}
			chunk, err := generator.GenerateNoise(int32(cx), int32(cz))
			if err != nil {
				t.Fatal(err)
			}
			biomeAtQuart := func(int32, int32, int32) (string, error) { return biome, nil }
			if biome == "natural" {
				sampler, err := worldgen.NewOverworldBiomeSampler(seed)
				if err != nil {
					t.Fatal(err)
				}
				biomeAtQuart = sampler.BiomeAtQuart
			}
			if err := generator.ApplySurface(chunk, biomeAtQuart); err != nil {
				t.Fatal(err)
			}
			mismatches := 0
			for index := range chunk.Blocks {
				state, err := chunk.BlockStateAt(int32(index%16), int32(index/256-64), int32(index/16%16))
				if err != nil {
					t.Fatal(err)
				}
				got, err := world.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				want := uint32(binary.BigEndian.Uint16(expected[index*2:]))
				if got != want {
					if mismatches < 10 {
						t.Errorf("%d,%d,%d: state %d (%s), want %d", index%16, index/256-64,
							index/16%16, got, state.Name, want)
					}
					mismatches++
				}
			}
			if mismatches != 0 {
				t.Errorf("%d of %d surface blocks differ", mismatches, len(chunk.Blocks))
			}
		})
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 108 {
		t.Fatalf("surface chunk count %d, want 108", count)
	}
}

func TestCarvedChunksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "carver_chunk_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 13 || fields[0] != "carved_chunk" || fields[4] != "natural" {
			t.Fatalf("invalid carved chunk vector")
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
		wantStarts, err := strconv.Atoi(fields[5])
		if err != nil {
			t.Fatal(err)
		}
		wantMask, err := strconv.Atoi(fields[7])
		if err != nil {
			t.Fatal(err)
		}
		var wantByCarver [3]int
		for index := range wantByCarver {
			wantByCarver[index], err = strconv.Atoi(fields[8+index])
			if err != nil {
				t.Fatal(err)
			}
		}
		compressedMask, err := base64.StdEncoding.DecodeString(fields[11])
		if err != nil {
			t.Fatal(err)
		}
		maskReader, err := gzip.NewReader(bytes.NewReader(compressedMask))
		if err != nil {
			t.Fatal(err)
		}
		expectedMask, err := io.ReadAll(io.LimitReader(maskReader, 16*16*384/8+1))
		maskReader.Close()
		if err != nil || len(expectedMask) != 16*16*384/8 {
			t.Fatalf("invalid carving mask reference: %d bytes (%v)", len(expectedMask), err)
		}
		compressed, err := base64.StdEncoding.DecodeString(fields[12])
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		expected, err := io.ReadAll(io.LimitReader(reader, 16*16*384*2+1))
		reader.Close()
		if err != nil || len(expected) != 16*16*384*2 {
			t.Fatalf("invalid carved chunk reference: %d bytes (%v)", len(expected), err)
		}
		t.Run(fmt.Sprintf("%d/%d,%d", seed, cx, cz), func(t *testing.T) {
			if starts := worldgen.OverworldCarverStarts(seed, int32(cx), int32(cz)); len(starts) != wantStarts {
				t.Fatalf("carver starts %d, want %d", len(starts), wantStarts)
			}
			generator, err := worldgen.NewOverworldNoiseGenerator(seed)
			if err != nil {
				t.Fatal(err)
			}
			chunk, err := generator.GenerateCarved(int32(cx), int32(cz))
			if err != nil {
				t.Fatal(err)
			}
			maskCount := 0
			for _, carved := range chunk.CarvingMask {
				if carved {
					maskCount++
				}
			}
			if maskCount != wantMask {
				t.Errorf("carving mask bits %d, want %d", maskCount, wantMask)
			}
			if chunk.CarvingMaskByCarver != wantByCarver {
				t.Errorf("carving masks by carver %v, want %v", chunk.CarvingMaskByCarver, wantByCarver)
			}
			maskMismatches := 0
			for index, carved := range chunk.CarvingMask {
				want := expectedMask[index/8]&(1<<uint(index%8)) != 0
				if carved != want {
					if maskMismatches < 8 {
						t.Errorf("carving mask at %d,%d,%d: got %t, want %t", index%16,
							index/256-64, index/16%16, carved, want)
					}
					maskMismatches++
				}
			}
			if maskMismatches != 0 {
				t.Errorf("%d carving mask bits differ", maskMismatches)
			}
			mismatches := 0
			for index := range chunk.Blocks {
				state, err := chunk.BlockStateAt(int32(index%16), int32(index/256-64), int32(index/16%16))
				if err != nil {
					t.Fatal(err)
				}
				got, err := world.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				want := uint32(binary.BigEndian.Uint16(expected[index*2:]))
				if got != want {
					if mismatches < 12 {
						t.Errorf("%d,%d,%d: state %d (%s), want %d", index%16, index/256-64,
							index/16%16, got, state.Name, want)
					}
					mismatches++
				}
			}
			if mismatches != 0 {
				t.Errorf("%d of %d carved blocks differ", mismatches, len(chunk.Blocks))
			}
		})
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 12 {
		t.Fatalf("carved chunk count %d, want 12", count)
	}
}
