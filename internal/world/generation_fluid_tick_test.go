package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// The frozen original FULL chunks contain both delay-zero ProtoChunkTicks and
// delay-five water updates from chunk activation. Only the former belong to
// the generation stage; the latter still require runtime fluid processing.
func TestGenerationStageFluidTicksPersistAndPreserveEdits(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/native_generation_fluid_ticks_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := make(map[[2]int32][]nbt.Tag)
	var chunkKey [2]int32
	zero, later := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		number := func(i int) int32 {
			t.Helper()
			v, e := strconv.ParseInt(f[i], 10, 32)
			if e != nil {
				t.Fatal(e)
			}
			return int32(v)
		}
		switch f[0] {
		case "native_fluid_chunk":
			if f[1] != "4511451270791457643" {
				t.Fatal("wrong native seed")
			}
			chunkKey = [2]int32{number(2), number(3)}
			want[chunkKey] = []nbt.Tag{}
		case "native_fluid_tick":
			if number(5) != 0 {
				later++
				continue
			}
			entry := nbt.Object{"i": {Type: nbt.String, Value: f[1]}}
			for i, key := range []string{"x", "y", "z", "t", "p"} {
				entry[key] = nbt.Tag{Type: nbt.Int, Value: number(i + 2)}
			}
			want[chunkKey] = append(want[chunkKey], nbt.Tag{Type: nbt.Compound, Value: entry})
			zero++
		default:
			t.Fatal("invalid native tick row", line)
		}
	}
	if len(want) != 2 || zero != 1 || later != 2 {
		t.Fatal("incomplete native fluid queue reference")
	}
	g, err := worldgen.NewOverworldGeneratorWithOptions(4511451270791457643, worldgen.OverworldGeneratorOptions{BeehiveShuffleSalt: 5393695769601225306})
	if err != nil {
		t.Fatal(err)
	}
	for key, ticks := range want {
		t.Run(fmt.Sprint(key), func(t *testing.T) {
			generated, e := g.Generate(key[0], key[1])
			if e != nil {
				t.Fatal(e)
			}
			chunk, e := chunkFromGeneratedOverworld(generated)
			if e != nil {
				t.Fatal(e)
			}
			dir := t.TempDir()
			for pass := 0; pass < 2; pass++ {
				name, root, e := chunk.updateAnvilBlocks()
				if e != nil {
					t.Fatal(e)
				}
				object, _ := root.Object()
				got, _ := object["fluid_ticks"].Elements()
				if !reflect.DeepEqual(got, ticks) {
					t.Fatalf("pass %d fluid ticks: got %v want %v", pass, got, ticks)
				}
				if e = anvil.WriteChunk(dir, key[0], key[1], name, root); e != nil {
					t.Fatal(e)
				}
				name, root, exists, e := anvil.ReadChunk(dir, key[0], key[1])
				if e != nil || !exists {
					t.Fatal("saved chunk missing", e)
				}
				chunk, e = chunkFromAnvil(name, root, key[0], key[1])
				if e != nil {
					t.Fatal(e)
				}
				// Removing a scheduled source must not silently delete/retype its
				// pending request; the runtime will check the current fluid type.
				chunk.SetBlock(9, 8, -53, StateStone)
			}
		})
	}
}
