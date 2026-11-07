package world

import (
	"atlas/internal/nbt"
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

// Final native saves include scheduled updates. Keep this complete-array gate
// at the world boundary: terrain generation must not secretly advance time.
func TestFinalOceanChunksAgainstVanillaServer(t *testing.T) {
	runFinalOceanChunksReference(t, "ocean")
}
func TestFinalOceanRuinChunksAgainstVanillaServer(t *testing.T) {
	runFinalOceanChunksReference(t, "ocean_ruin")
}
func TestFinalShipwreckChunksAgainstVanillaServer(t *testing.T) {
	runFinalOceanChunksReference(t, "shipwreck")
}
func runFinalOceanChunksReference(t *testing.T, preset string) {
	structures := preset != "ocean"
	fixture := "final_ocean_chunk_1.21.1.txt"
	if preset == "ocean_ruin" {
		fixture = "final_ocean_ruin_chunk_1.21.1.txt"
	} else if preset == "shipwreck" {
		fixture = "final_shipwreck_chunk_1.21.1.txt"
	} else if preset != "ocean" {
		t.Fatal("unknown final chunk fixture", preset)
	}
	f, err := os.Open("../../testdata/worldgen/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 65536), 2*1024*1024)
	cases := 0
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if len(fields) != 5 || fields[0] != "final_chunk" {
			t.Fatal("invalid original ocean fixture")
		}
		seed, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		x, err := strconv.ParseInt(fields[2], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		z, err := strconv.ParseInt(fields[3], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		cx, cz := int32(x), int32(z)
		compressed, err := base64.StdEncoding.DecodeString(fields[4])
		if err != nil {
			t.Fatal(err)
		}
		gz, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		expected, err := io.ReadAll(io.LimitReader(gz, 384*256*2+1))
		gz.Close()
		if err != nil || len(expected) != 384*256*2 {
			t.Fatal("invalid original block array", err)
		}
		t.Run(fmt.Sprintf("%d,%d", cx, cz), func(t *testing.T) {
			dir := t.TempDir()
			if err := CreateNormalWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			if !structures {
				v, err := openVanilla(dir)
				if err != nil {
					t.Fatal(err)
				}
				root, _ := v.root.Object()
				data, _ := root.Object("Data")
				settings, _ := data.Object("WorldGenSettings")
				settings["generate_features"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
				marker, _ := data.Object("AtlasWorldgen")
				marker["BeehiveShuffleSalt"] = nbt.Tag{Type: nbt.Long, Value: int64(5393695769601225306)}
				if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), v.name, v.root); err != nil {
					t.Fatal(err)
				}
			}
			w, err := Open(noViewers{}, 2, dir)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			w.tickRandom = worldgen.NewLegacyRandom(0)
			if structures {
				if _, err := w.LoadChunk(cx, cz); err != nil {
					t.Fatal(err)
				}
			}
			// Native /forceload uses a level-31 ticket: the center and its
			// one-chunk ring tick blocks, with a further FULL ring for reads.
			for dx := int32(-2); dx <= 2; dx++ {
				for dz := int32(-2); dz <= 2; dz++ {
					if _, err := w.LoadChunk(cx+dx, cz+dz); err != nil {
						t.Fatal(err)
					}
				}
			}
			eligible := func(x, z int32) bool { return x >= cx-1 && x <= cx+1 && z >= cz-1 && z <= cz+1 }
			w.blockUpdateMu.Lock()
			runtime := &fluidRuntime{world: w, settings: w.fluidSettings()}
			activationErr := w.scheduledTicks.activate(eligible, runtime.postProcess)
			w.blockUpdateMu.Unlock()
			if activationErr != nil {
				t.Fatal(activationErr)
			}
			if !structures {
				checkFrozenActivationTicks(t, w, cx, cz)
			}
			var checkRuntime func()
			if preset == "ocean_ruin" {
				checkRuntime = observeStructureRuntime(t, w, cx, cz, "../../testdata/worldgen/native_ocean_ruin_live_ticks_1.21.1.json")
			} else if preset == "shipwreck" {
				checkRuntime = observeStructureRuntime(t, w, cx, cz, "../../testdata/worldgen/native_shipwreck_live_ticks_1.21.1.json")
			}
			for tick := 1; tick <= 100; tick++ {
				w.AdvanceTime(1)
				if err := w.TickScheduled(eligible); err != nil {
					t.Fatalf("tick %d: %v", tick, err)
				}
				if err := w.TickFallingEntities(func(x, z int32) bool { return x == cx && z == cz }); err != nil {
					t.Fatalf("entity tick %d: %v", tick, err)
				}
			}
			if checkRuntime != nil {
				checkRuntime()
			}
			if err := w.Save(); err != nil {
				t.Fatal(err)
			}
			chunk, err := w.store.vanilla.loadChunk(cx, cz)
			if err != nil {
				t.Fatal(err)
			}
			mismatches := 0
			pairs := make(map[string]int)
			for i := 0; i < 384*256; i++ {
				x, y, z := int32(i&15), int32(i/256)-64, int32((i>>4)&15)
				got := chunk.GetBlock(int(x), int(z), y)
				want := uint32(binary.BigEndian.Uint16(expected[i*2:]))
				if got == want {
					continue
				}
				a, _ := BlockStateByID(got)
				b, _ := BlockStateByID(want)
				pairs[a.Name+" -> "+b.Name]++
				if mismatches < 15 {
					t.Logf("block %d,%d,%d got %s %v want %s %v", cx*16+x, y, cz*16+z, a.Name, a.Properties, b.Name, b.Properties)
				}
				mismatches++
			}
			if mismatches != 0 {
				t.Fatalf("%d runtime final blocks differ: %v", mismatches, pairs)
			}
			t.Log("all 98304 final block states match the original server after 100 ticks")
		})
		cases++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 2 {
		t.Fatal("incomplete final ocean fixture", cases)
	}
}
