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
	"math"
	"os"
	"strings"
	"testing"
)

func TestDimensionNoiseAgainstVanillaJAR(t *testing.T) {
	for _, dimension := range []string{"nether", "end"} {
		t.Run(dimension, func(t *testing.T) {
			file, err := os.Open("../../docs/worldgen/"+dimension+"_noise_1.21.1.txt")
			if err != nil { t.Fatal(err) }
			defer file.Close()
			settings, err := worldgen.LoadDensitySettings(dimension)
			if err != nil { t.Fatal(err) }
			var ids [12]uint32
			for i := range ids {
				ids[i], err = blockstate.BlockStateID(worldgen.NoiseMaterial(i).BlockName(), nil)
				if err != nil { t.Fatal(err) }
			}
			scan := bufio.NewScanner(file)
			scan.Buffer(make([]byte,65536), 4*1024*1024)
			chunks, densities := 0, 0
			for scan.Scan() {
				line := scan.Text()
				var kind string
				var seed int64
				var cx, cz int32
				if _, err := fmt.Sscan(line, &kind, &seed, &cx, &cz); err != nil { t.Fatal(err) }
				if kind == "density" {
					var y int32
					var bits uint64
					if _, err := fmt.Sscan(line, &kind, &seed, &cx, &cz, &y, &bits); err != nil { t.Fatal(err) }
					density := &worldgen.DensityNode{Kind:"cache_all_in_cell", Children: map[string]*worldgen.DensityNode{"argument":settings.Router["final_density"]}}
					v, err := worldgen.NewNoiseChunkDensitySampler(seed, settings, cx, cz).Sample(density,cx*16+5,y,cz*16+9)
					if err != nil || math.Float64bits(v) != bits { t.Fatalf("%s %s: got %016x want %016x (%v)", dimension,line,math.Float64bits(v),bits,err) }
					densities++
					continue
				}
				if kind != "chunk" { t.Fatalf("unknown reference row %s", kind) }
				compressed, err := base64.StdEncoding.DecodeString(strings.Fields(line)[4])
				if err != nil { t.Fatal(err) }
				reader, err := gzip.NewReader(bytes.NewReader(compressed))
				if err != nil { t.Fatal(err) }
				expected, err := io.ReadAll(io.LimitReader(reader, 128*256*3+1))
				reader.Close()
				if err != nil || len(expected) != 128*256*3 { t.Fatalf("invalid dimension reference size %d: %v",len(expected),err) }
				g, err := worldgen.NewDimensionNoiseGenerator(seed, dimension)
				if err != nil { t.Fatal(err) }
				chunk, err := g.GenerateNoise(cx, cz)
				if err != nil { t.Fatal(err) }
				if chunk.MinY != 0 || chunk.Height != 128 { t.Fatalf("invalid %s noise bounds",dimension) }
				for i, material := range chunk.Blocks {
					want := uint32(binary.BigEndian.Uint16(expected[i*3:]))
					update := expected[i*3+2] != 0
					if ids[material] != want || chunk.FluidUpdates[i] != update { t.Fatalf("%s seed %d chunk %d,%d at %d,%d,%d: state %d update %v; want %d %v",dimension,seed,cx,cz,i%16,i/256,i/16%16,ids[material],chunk.FluidUpdates[i],want,update) }
				}
				chunks++
			}
			if err := scan.Err(); err != nil { t.Fatal(err) }
			if chunks != 20 || densities != 40 { t.Fatalf("%s coverage: chunks %d, densities %d",dimension,chunks,densities) }
		})
	}
}
