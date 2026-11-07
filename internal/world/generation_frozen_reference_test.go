package world

import (
	"atlas/internal/anvil"
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

// FULL saves already contain additional activation ticks and a heap layout.
// Compare delay-zero generation entries by identity here; independent feature
// probes separately verify requests and ProtoChunkTicks insertion order.
func TestGeneratedChunksAndTicksAgainstFrozenVanillaWorld(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/native_generation_block_ticks_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	wantTicks := map[[2]int32]map[tickqueue.Position]tickqueue.Saved[string]{}
	var key [2]int32
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var label string
		if strings.HasPrefix(line, "native_block_chunk ") {
			var seed int64
			var count int
			if _, err := fmt.Sscan(line, &label, &seed, &key[0], &key[1], &count); err != nil || seed != 4511451270791457643 {
				t.Fatal("invalid chunk", err)
			}
			wantTicks[key] = map[tickqueue.Position]tickqueue.Saved[string]{}
		} else {
			var tick tickqueue.Saved[string]
			var priority int32
			if _, err := fmt.Sscan(line, &label, &tick.Type, &tick.Position.X, &tick.Position.Y, &tick.Position.Z, &tick.Delay, &priority); err != nil {
				t.Fatal(err)
			}
			tick.Priority = tickqueue.ClampPriority(priority)
			if tick.Delay == 0 {
				wantTicks[key][tick.Position] = tick
			}
		}
	}
	frozen, err := os.ReadFile("../../testdata/worldgen/frozen_ocean_chunk_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	g, err := worldgen.NewOverworldGeneratorWithOptions(4511451270791457643, worldgen.OverworldGeneratorOptions{BeehiveShuffleSalt: 5393695769601225306})
	if err != nil {
		t.Fatal(err)
	}
	leaves, ticks, cases := 0, 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(frozen)), "\n") {
		var label, payload string
		var seed int64
		if _, err := fmt.Sscan(line, &label, &seed, &key[0], &key[1], &payload); err != nil || label != "final_chunk" || seed != 4511451270791457643 {
			t.Fatal("invalid frozen block row", err)
		}
		compressed, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			t.Fatal(err)
		}
		gz, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatal(err)
		}
		states, err := io.ReadAll(gz)
		gz.Close()
		if err != nil || len(states) != 98304*2 {
			t.Fatal("invalid frozen states", err)
		}
		generated, err := g.Generate(key[0], key[1])
		if err != nil {
			t.Fatal(err)
		}
		chunk, err := chunkFromGeneratedOverworld(generated)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 98304; i++ {
			want := uint32(binary.BigEndian.Uint16(states[i*2:]))
			got := chunk.GetBlock(i&15, (i>>4)&15, int32(i/256)-64)
			a, _ := BlockStateByID(want)
			b, _ := BlockStateByID(got)
			if got != want {
				t.Fatalf("frozen block %v index=%d: got %+v want %+v", key, i, b, a)
			}
			if strings.HasSuffix(a.Name, "_leaves") {
				leaves++
			}
		}
		dir := t.TempDir()
		for pass := 0; pass < 2; pass++ {
			name, root, err := chunk.updateAnvilBlocks()
			if err != nil {
				t.Fatal(err)
			}
			obj, _ := root.Object()
			rows := tickqueue.DecodeSavedNBT(obj["block_ticks"], key[0], key[1], func(name string) (string, bool) { return name, true })
			gotTicks := map[tickqueue.Position]tickqueue.Saved[string]{}
			for _, row := range rows {
				gotTicks[row.Position] = row
			}
			if !reflect.DeepEqual(gotTicks, wantTicks[key]) {
				t.Fatalf("generation block queue %v pass %d:\n got %v\nwant %v", key, pass, gotTicks, wantTicks[key])
			}
			if pass == 0 {
				ticks += len(gotTicks)
			}
			if err := anvil.WriteChunk(dir, key[0], key[1], name, root); err != nil {
				t.Fatal(err)
			}
			var exists bool
			name, root, exists, err = anvil.ReadChunk(dir, key[0], key[1])
			if err != nil || !exists {
				t.Fatal("missing saved chunk", err)
			}
			chunk, err = chunkFromAnvil(name, root, key[0], key[1])
			if err != nil {
				t.Fatal(err)
			}
			// Replacing a queued block must retain its original request on disk.
			for p := range gotTicks {
				chunk.SetBlock(int(p.X&15), int(p.Z&15), p.Y, StateStone)
				break
			}
		}
		cases++
	}
	if cases != 2 || leaves != 47 || ticks != 133 {
		t.Fatal("incomplete frozen reference", cases, leaves, ticks)
	}
	t.Logf("matched 196608 frozen block states (including %d leaves) and %d generation block ticks, including Anvil reload after edits", leaves, ticks)
}
