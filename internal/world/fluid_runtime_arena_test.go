package world

import (
	"atlas/internal/nbt"
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type nativeRuntimeTick struct {
	Type                     string
	X, Y, Z, Delay, Priority int32
}
type nativeRuntimeChunk struct {
	X, Z            int32
	Blocks          []byte
	Ticks           []nativeRuntimeTick
	BlockTicks      []nativeRuntimeTick `json:"block_ticks"`
	BlockEntities   []byte              `json:"block_entities"`
	FallingEntities []byte              `json:"falling_entities"`
}
type nativeRuntimeSnapshot struct {
	Tick   int
	Chunks []nativeRuntimeChunk
}

func (c nativeRuntimeChunk) entities(t *testing.T) nbt.Tag {
	t.Helper()
	_, tag, err := nbt.Parse(c.BlockEntities)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := tag.Object()
	return root["block_entities"]
}

func (c nativeRuntimeChunk) saved(t *testing.T) []tickqueue.Saved[worldgen.FluidType] {
	t.Helper()
	result := make([]tickqueue.Saved[worldgen.FluidType], len(c.Ticks))
	for i, s := range c.Ticks {
		kind, ok := parseScheduledFluid(s.Type)
		if !ok {
			t.Fatalf("unknown reference fluid %q", s.Type)
		}
		result[i] = tickqueue.Saved[worldgen.FluidType]{Type: kind, Position: tickqueue.Position{X: s.X, Y: s.Y, Z: s.Z}, Delay: s.Delay, Priority: tickqueue.ClampPriority(s.Priority)}
	}
	return result
}

func (c nativeRuntimeChunk) savedBlocks(t *testing.T) []tickqueue.Saved[string] {
	t.Helper()
	result := make([]tickqueue.Saved[string], len(c.BlockTicks))
	for i, s := range c.BlockTicks {
		name, ok := parseScheduledBlock(s.Type)
		if !ok {
			t.Fatal("unknown native scheduled block", s.Type)
		}
		result[i] = tickqueue.Saved[string]{Type: name, Position: tickqueue.Position{X: s.X, Y: s.Y, Z: s.Z}, Delay: s.Delay, Priority: tickqueue.ClampPriority(s.Priority)}
	}
	return result
}

// These snapshots come from real ServerLevel ticks, not the callback-recording
// single-fluid-step probe. Every box is closed; none can exchange fluids with
// another box or the surrounding flat terrain.
func TestFluidRuntimeBoxesAgainstVanillaServer(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/fluid_runtime_1.21.1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	var fixture struct {
		Version   int
		MinY      int32 `json:"min_y"`
		Height    int32
		Scenes    []string
		Snapshots []nativeRuntimeSnapshot
	}
	if err := json.NewDecoder(gz).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 3955 || fixture.MinY != 60 || fixture.Height != 16 || len(fixture.Scenes) != 28 || len(fixture.Snapshots) != 17 {
		t.Fatal("incomplete native runtime reference")
	}
	dir := t.TempDir()
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	initial := fixture.Snapshots[0]
	if initial.Tick != 0 || len(initial.Chunks) != len(fixture.Scenes) {
		t.Fatal("missing initial native runtime state")
	}
	for _, source := range initial.Chunks {
		_, tag, err := nbt.Parse(source.FallingEntities)
		if err != nil {
			t.Fatal(err)
		}
		entityRoot, _ := tag.Object()
		entityList, ok := entityRoot["falling_entities"].Elements()
		if !ok || len(entityList) != 0 {
			t.Fatal("fixture must start before falling entities spawn")
		}
		if len(source.Blocks) != int(fixture.Height)*512 {
			t.Fatal("invalid initial native block array")
		}
		c, err := v.loadChunk(source.X, source.Z)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < len(source.Blocks)/2; i++ {
			c.SetBlock(i&15, (i>>4)&15, fixture.MinY+int32(i/256), uint32(binary.BigEndian.Uint16(source.Blocks[2*i:])))
		}
		root, _ := c.anvilRoot.Object()
		root["block_entities"] = source.entities(t)
		root["fluid_ticks"] = tickqueue.EncodeSavedNBT(source.saved(t), func(k worldgen.FluidType) string { return k.String() })
		root["block_ticks"] = tickqueue.EncodeSavedNBT(source.savedBlocks(t), func(name string) string { return name })
		root["PostProcessing"] = nbt.Tag{Type: nbt.List, ListType: nbt.End, Value: []nbt.Tag{}}
		if err := v.persistChunk(c); err != nil {
			t.Fatal(err)
		}
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	w.tickRandom = worldgen.NewLegacyRandom(0)
	// Native forced chunks (1..4,1..7) have a block-ticking ring and an
	// additional FULL ring. This fixture contains no fluids outside its boxes.
	for x := int32(-1); x <= 6; x++ {
		for z := int32(-1); z <= 9; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	eligible := func(x, z int32) bool { return x >= 0 && x <= 5 && z >= 0 && z <= 8 }
	entityEligible := func(x, z int32) bool { return x >= 1 && x <= 4 && z >= 1 && z <= 7 }
	// Native commands created these requests in already activated chunks.
	// Unpack at initial time, before advancing to the first sampled game tick.
	w.blockUpdateMu.Lock()
	runtime := &fluidRuntime{world: w, settings: w.fluidSettings()}
	err = w.scheduledTicks.activate(eligible, runtime.postProcess)
	w.blockUpdateMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	previous := 0
	for _, snapshot := range fixture.Snapshots {
		for step := previous + 1; step <= snapshot.Tick; step++ {
			w.AdvanceTime(1)
			if err := w.TickScheduled(eligible); err != nil {
				t.Fatalf("tick %d: %v", step, err)
			}
			if err := w.TickFallingEntities(entityEligible); err != nil {
				t.Fatalf("entity tick %d: %v", step, err)
			}
		}
		if err := w.Save(); err != nil {
			t.Fatal(err)
		}
		if len(snapshot.Chunks) != len(fixture.Scenes) {
			t.Fatal("incomplete checkpoint")
		}
		for index, expected := range snapshot.Chunks {
			c, err := w.store.vanilla.loadChunk(expected.X, expected.Z)
			if err != nil {
				t.Fatal(err)
			}
			if len(expected.Blocks) != int(fixture.Height)*512 {
				t.Fatal("invalid native block array")
			}
			entities := expected.entities(t)
			root, _ := c.anvilRoot.Object()
			if !reflect.DeepEqual(root["block_entities"], entities) {
				t.Errorf("tick %d %s block entities differ: got %v want %v", snapshot.Tick, fixture.Scenes[index], root["block_entities"], entities)
			}
			_, _, allEntities, err := w.store.vanilla.readEntityChunk(expected.X, expected.Z)
			if err != nil {
				t.Fatal(err)
			}
			falling := []nbt.Tag{}
			for _, entity := range allEntities {
				if entity.Type == "minecraft:falling_block" {
					delete(entity.Data, "UUID")
					falling = append(falling, nbt.Tag{Type: nbt.Compound, Value: entity.Data})
				}
			}
			_, wantTag, err := nbt.Parse(expected.FallingEntities)
			if err != nil {
				t.Fatal(err)
			}
			wantRoot, _ := wantTag.Object()
			gotFalling := nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: falling}
			if !reflect.DeepEqual(gotFalling, wantRoot["falling_entities"]) {
				t.Errorf("tick %d %s falling entities differ: got %v want %v", snapshot.Tick, fixture.Scenes[index], gotFalling, wantRoot["falling_entities"])
			}
			mismatches := 0
			for i := 0; i < len(expected.Blocks)/2; i++ {
				x, z, y := i&15, (i>>4)&15, fixture.MinY+int32(i/256)
				got, want := c.GetBlock(x, z, y), uint32(binary.BigEndian.Uint16(expected.Blocks[i*2:]))
				if got != want {
					if mismatches < 5 {
						a, _ := BlockStateByID(got)
						b, _ := BlockStateByID(want)
						t.Errorf("tick %d %s block %d,%d,%d got %s %v want %s %v", snapshot.Tick, fixture.Scenes[index], expected.X*16+int32(x), y, expected.Z*16+int32(z), a.Name, a.Properties, b.Name, b.Properties)
					}
					mismatches++
				}
			}
			if mismatches > 0 {
				t.Errorf("tick %d %s: %d state mismatches", snapshot.Tick, fixture.Scenes[index], mismatches)
			}
			got := diskFluidTicks(t, dir, expected.X, expected.Z)
			want := expected.saved(t)
			if len(got) != len(want) || len(got) > 0 && !reflect.DeepEqual(got, want) {
				t.Errorf("tick %d %s queue counts: got %d; want %d", snapshot.Tick, fixture.Scenes[index], len(got), len(want))
				for i := 0; i < min(len(got), len(want)); i++ {
					if got[i] != want[i] {
						t.Errorf("first queue difference at %d: got %+v; want %+v", i, got[i], want[i])
						break
					}
				}
			}
			gotBlocks := tickqueue.DecodeSavedNBT(root["block_ticks"], expected.X, expected.Z, parseScheduledBlock)
			if wantBlocks := expected.savedBlocks(t); !reflect.DeepEqual(gotBlocks, wantBlocks) {
				t.Errorf("tick %d %s block queue differs: got %v; want %v", snapshot.Tick, fixture.Scenes[index], gotBlocks, wantBlocks)
			}
		}
		if t.Failed() {
			t.FailNow()
		}
		t.Logf("tick %d: all %d boxes, block entities, falling entities and ordered pending queues match", snapshot.Tick, len(fixture.Scenes))
		previous = snapshot.Tick
	}
}
