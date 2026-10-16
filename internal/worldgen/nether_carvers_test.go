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

func TestNetherCarversAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/nether_carver_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	decode := func(text string, size int) []byte {
		compressed, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		data, err := io.ReadAll(io.LimitReader(reader, int64(size+1)))
		if err != nil || len(data) != size {
			t.Fatalf("invalid reference length %d (%v)", len(data), err)
		}
		return data
	}
	ids := make(map[string]uint32)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 13 || f[0] != "carved_chunk" {
			t.Fatalf("invalid carving reference: %s", scanner.Text())
		}
		seed, err := strconv.ParseInt(f[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		x, err := strconv.ParseInt(f[2], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		z, err := strconv.ParseInt(f[3], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		mask, expected := decode(f[11], 128*256/8), decode(f[12], 128*256*2)
		g, err := worldgen.NewDimensionNoiseGenerator(seed, "nether")
		if err != nil {
			t.Fatal(err)
		}
		chunk, err := g.GenerateSurface(int32(x), int32(z))
		if err != nil {
			t.Fatal(err)
		}
		if err := g.ApplyNetherCarvers(chunk); err != nil {
			t.Fatal(err)
		}
		maskCount := 0
		for i, got := range chunk.CarvingMask {
			want := mask[i/8]&(1<<uint(i&7)) != 0
			if got != want {
				t.Fatalf("seed %d chunk %d,%d mask at %d,%d,%d: got %v, want %v", seed, x, z, i%16, i/256, i/16%16, got, want)
			}
			if got {
				maskCount++
			}
		}
		wantCount, _ := strconv.Atoi(f[7])
		if maskCount != wantCount || chunk.CarvingMaskByCarver[0] != wantCount {
			t.Fatalf("mask count %d, want %d", maskCount, wantCount)
		}
		for i := range chunk.Blocks {
			state, err := chunk.BlockStateAt(int32(i%16), int32(i/256), int32(i/16%16))
			if err != nil {
				t.Fatal(err)
			}
			key := state.Name + fmt.Sprint(state.Properties)
			id, exists := ids[key]
			if !exists {
				id, err = blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				ids[key] = id
			}
			want := uint32(binary.BigEndian.Uint16(expected[i*2:]))
			if id != want {
				t.Fatalf("seed %d chunk %d,%d at %d,%d,%d: got %s %v (%d), want %d", seed, x, z, i%16, i/256, i/16%16, state.Name, state.Properties, id, want)
			}
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 20 {
		t.Fatalf("got %d carved chunks, want 20", count)
	}
}
