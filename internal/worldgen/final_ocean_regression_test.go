package worldgen_test

import (
	"atlas/internal/blockstate"
	"atlas/internal/worldgen"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

// These are complete final block arrays from an isolated official 1.21.1
// server save, including vegetation and settled bubble columns. Terrain-only
// fixtures cannot detect missing scheduled generation updates.
func TestFinalOceanChunksAgainstVanillaServer(t *testing.T) {
	f, err := os.Open("../../docs/worldgen/final_ocean_chunk_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 65536), 2*1024*1024)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if len(fields) != 5 || fields[0] != "final_chunk" {
			t.Fatal("invalid final ocean reference")
		}
		parse := func(i, bits int) int64 {
			n, err := strconv.ParseInt(fields[i], 10, bits)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		seed, cx, cz := parse(1, 64), int32(parse(2, 32)), int32(parse(3, 32))
		compressed, err := base64.StdEncoding.DecodeString(fields[4])
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		expected, err := io.ReadAll(io.LimitReader(reader, 384*256*2+1))
		reader.Close()
		if err != nil || len(expected) != 384*256*2 {
			t.Fatalf("invalid final reference size %d: %v", len(expected), err)
		}
		t.Run(fmt.Sprintf("%d/%d,%d", seed, cx, cz), func(t *testing.T) {
			generator, err := worldgen.NewOverworldGeneratorWithOptions(seed, worldgen.OverworldGeneratorOptions{BeehiveShuffleSalt: 5393695769601225306})
			if err != nil {
				t.Fatal(err)
			}
			generated, err := generator.Generate(cx, cz)
			if err != nil {
				t.Fatal(err)
			}
			terrain := generated.Terrain
			if terrain.MinY != -64 || terrain.Height != 384 {
				t.Fatal("unexpected Overworld bounds")
			}
			mismatches := 0
			for i := 0; i < 384*256; i++ {
				x, y, z := int32(i&15), int32(i/256)-64, int32((i>>4)&15)
				state, err := terrain.BlockStateAt(x, y, z)
				if err != nil {
					t.Fatal(err)
				}
				id, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				want := uint32(binary.BigEndian.Uint16(expected[i*2:]))
				if id != want {
					if mismatches < 10 {
						t.Errorf("block %d,%d,%d: %s %v (ID %d), want ID %d", cx*16+x, y, cz*16+z, state.Name, state.Properties, id, want)
					}
					mismatches++
				}
			}
			if mismatches > 0 {
				t.Fatalf("%d final blocks differ from the official server", mismatches)
			}
			t.Log("all 98304 final block states match the official server")
		})
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("final ocean reference count %d, want 2", count)
	}
}
