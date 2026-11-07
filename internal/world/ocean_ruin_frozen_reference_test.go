package world

import (
	"atlas/internal/nbt"
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

// Actual official-server FULL saves with structures enabled. The native server
// is frozen before either remote chunk is loaded, preserving generation queues
// and archaeology loot tables before any scheduled callback executes.
func TestOceanRuinGenerationAgainstFrozenVanillaServer(t *testing.T) {
	checkFrozenStructureChunks(t,
		"../../testdata/worldgen/frozen_ocean_ruin_chunk_1.21.1.txt",
		"../../testdata/worldgen/native_ocean_ruin_generation_block_entities_1.21.1.txt",
		"../../testdata/worldgen/native_ocean_ruin_generation_block_ticks_1.21.1.txt",
		"../../testdata/worldgen/native_ocean_ruin_generation_fluid_ticks_1.21.1.txt")
}

func TestShipwreckGenerationAgainstFrozenVanillaServer(t *testing.T) {
	checkFrozenStructureChunks(t,
		"../../testdata/worldgen/frozen_shipwreck_chunk_1.21.1.txt",
		"../../testdata/worldgen/native_shipwreck_generation_block_entities_1.21.1.txt",
		"../../testdata/worldgen/native_shipwreck_generation_block_ticks_1.21.1.txt",
		"../../testdata/worldgen/native_shipwreck_generation_fluid_ticks_1.21.1.txt")
}

func checkFrozenStructureChunks(t *testing.T, blockPath, entityPath, tickPath, fluidPath string) {
	t.Helper()
	read := func(path string) []string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Split(strings.TrimSpace(string(data)), "\n")
	}
	blocks := read(blockPath)
	entities := read(entityPath)
	ticks := read(tickPath)
	fluidTicks := read(fluidPath)
	decode := func(s string) []byte {
		t.Helper()
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	wantEntities := map[[2]int32][]byte{}
	for _, row := range entities {
		var label, encoded string
		var seed int64
		var x, z int32
		if _, err := fmt.Sscan(row, &label, &seed, &x, &z, &encoded); err != nil || label != "native_block_entities" || seed != 1234 {
			t.Fatal("invalid native block entities", err)
		}
		wantEntities[[2]int32{x, z}] = decode(encoded)
	}
	wantTicks := parseFrozenRuinTicks(t, ticks, "block")
	wantFluids := parseFrozenRuinTicks(t, fluidTicks, "fluid")
	if len(blocks) != 2 || len(wantEntities) != 2 || len(wantTicks) != 2 || len(wantFluids) != 2 {
		t.Fatal("incomplete native structure fixture")
	}
	for _, row := range blocks {
		var label, encoded string
		var seed int64
		var x, z int32
		if _, err := fmt.Sscan(row, &label, &seed, &x, &z, &encoded); err != nil || label != "final_chunk" || seed != 1234 {
			t.Fatal("invalid frozen structure blocks", err)
		}
		t.Run(fmt.Sprintf("%d,%d", x, z), func(t *testing.T) {
			gz, err := gzip.NewReader(bytes.NewReader(decode(encoded)))
			if err != nil {
				t.Fatal(err)
			}
			states, err := io.ReadAll(io.LimitReader(gz, 98304*2+1))
			gz.Close()
			if err != nil || len(states) != 98304*2 {
				t.Fatal("invalid state array", err)
			}
			dir := t.TempDir()
			if err := CreateNormalWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			w, err := Open(noViewers{}, 2, dir)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			w.tickRandom = worldgen.NewLegacyRandom(0)
			// The native forceload target enters postProcessGeneration first
			// (independently observed in the original activation trace). Match
			// that ticket/load order before activating its surrounding ring.
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
			for dx := int32(-2); dx <= 2; dx++ {
				for dz := int32(-2); dz <= 2; dz++ {
					if _, err := w.LoadChunk(x+dx, z+dz); err != nil {
						t.Fatal(err)
					}
				}
			}
			eligible := func(cx, cz int32) bool { return cx >= x-1 && cx <= x+1 && cz >= z-1 && cz <= z+1 }
			w.blockUpdateMu.Lock()
			runtime := &fluidRuntime{world: w, settings: w.fluidSettings()}
			activationErr := w.scheduledTicks.activate(eligible, runtime.postProcess)
			w.blockUpdateMu.Unlock()
			if activationErr != nil {
				t.Fatal(activationErr)
			}
			if err := w.Save(); err != nil {
				t.Fatal(err)
			}
			chunk, err := w.store.vanilla.loadChunk(x, z)
			if err != nil {
				t.Fatal(err)
			}
			mismatches := 0
			for i := 0; i < 98304; i++ {
				want := uint32(binary.BigEndian.Uint16(states[i*2:]))
				got := chunk.GetBlock(i&15, (i>>4)&15, int32(i/256)-64)
				if got != want {
					if mismatches < 12 {
						a, _ := BlockStateByID(got)
						b, _ := BlockStateByID(want)
						t.Logf("%d,%d,%d got %+v want %+v", x*16+int32(i&15), int32(i/256)-64, z*16+int32((i>>4)&15), a, b)
					}
					mismatches++
				}
			}
			if mismatches != 0 {
				t.Errorf("%d frozen structure block states differ", mismatches)
			}
			_, root, err := chunk.updateAnvilBlocks()
			if err != nil {
				t.Fatal(err)
			}
			obj, _ := root.Object()
			byPosition := map[string]nbt.Object{}
			list, _ := obj["block_entities"].Elements()
			for _, entry := range list {
				o, _ := entry.Object()
				px, _ := o.Number("x")
				py, _ := o.Number("y")
				pz, _ := o.Number("z")
				byPosition[fmt.Sprintf("%d,%d,%d", px, py, pz)] = o
			}
			actual, err := json.Marshal(byPosition)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, wantEntities[[2]int32{x, z}]) {
				t.Errorf("block entities differ:\n got %s\nwant %s", actual, wantEntities[[2]int32{x, z}])
			}
			gotTicks := tickqueue.DecodeSavedNBT(obj["block_ticks"], x, z, func(s string) (string, bool) { return s, true })
			want := wantTicks[[2]int32{x, z}]
			if !slices.Equal(gotTicks, want) {
				first := 0
				for first < min(len(gotTicks), len(want)) && gotTicks[first] == want[first] {
					first++
				}
				t.Errorf("ordered block ticks differ: got %d want %d, first difference %d: got %v want %v", len(gotTicks), len(want), first, gotTicks[max(0, first-1):min(len(gotTicks), first+3)], want[max(0, first-1):min(len(want), first+3)])
			}
			gotFluids := tickqueue.DecodeSavedNBT(obj["fluid_ticks"], x, z, func(s string) (string, bool) { return s, true })
			wantFluid := wantFluids[[2]int32{x, z}]
			if !slices.Equal(gotFluids, wantFluid) {
				actualSet := map[tickqueue.Saved[string]]bool{}
				for _, tick := range gotFluids {
					actualSet[tick] = true
				}
				missingTypes := map[string]int{}
				for _, tick := range wantFluid {
					if !actualSet[tick] {
						id := chunk.GetBlock(int(tick.Position.X&15), int(tick.Position.Z&15), tick.Position.Y)
						state, _ := BlockStateByID(id)
						missingTypes[state.Name]++
					}
				}
				t.Logf("missing fluid requests by final block: %v", missingTypes)
				first := 0
				for first < min(len(gotFluids), len(wantFluid)) && gotFluids[first] == wantFluid[first] {
					first++
				}
				t.Errorf("ordered fluid ticks differ: got %d want %d, first difference %d: got %v want %v", len(gotFluids), len(wantFluid), first, gotFluids[max(0, first-1):min(len(gotFluids), first+3)], wantFluid[max(0, first-1):min(len(wantFluid), first+3)])
			}
			t.Logf("compared 98304 frozen states, %d block entities, %d block ticks and %d fluid ticks", len(byPosition), len(gotTicks), len(gotFluids))
		})
	}
}

func parseFrozenRuinTicks(t *testing.T, rows []string, kind string) map[[2]int32][]tickqueue.Saved[string] {
	t.Helper()
	result := map[[2]int32][]tickqueue.Saved[string]{}
	counts := map[[2]int32]int{}
	var key [2]int32
	for _, row := range rows {
		var label string
		if strings.HasPrefix(row, "native_"+kind+"_chunk ") {
			var seed int64
			var count int
			if _, err := fmt.Sscan(row, &label, &seed, &key[0], &key[1], &count); err != nil || seed != 1234 {
				t.Fatal("invalid native tick owner", err)
			}
			if _, exists := counts[key]; exists {
				t.Fatal("duplicate native tick chunk")
			}
			counts[key] = count
			result[key] = nil
		} else {
			var tick tickqueue.Saved[string]
			var priority int32
			if _, err := fmt.Sscan(row, &label, &tick.Type, &tick.Position.X, &tick.Position.Y, &tick.Position.Z, &tick.Delay, &priority); err != nil || label != "native_"+kind+"_tick" {
				t.Fatal("invalid native tick", err)
			}
			tick.Priority = tickqueue.ClampPriority(priority)
			result[key] = append(result[key], tick)
		}
	}
	for key, count := range counts {
		if len(result[key]) != count {
			t.Fatal("incomplete native tick rows", kind, key)
		}
	}
	return result
}
