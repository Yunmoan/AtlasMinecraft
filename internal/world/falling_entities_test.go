package world

import (
	"atlas/internal/nbt"
	"bufio"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func fallingTestWorld(t *testing.T, dir string) *World {
	t.Helper()
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	})
	for x := int32(-4); x <= 0; x++ {
		for z := int32(0); z <= 4; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	return w
}

func spawnTestFallingSand(t *testing.T) (*World, string, WorldEntity) {
	return spawnTestFallingBlock(t, "sand")
}
func spawnTestFallingBlock(t *testing.T, block string) (*World, string, WorldEntity) {
	t.Helper()
	dir := t.TempDir()
	seed := int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w := fallingTestWorld(t, dir)
	eligible := func(x, z int32) bool { return x >= -3 && x <= -1 && z >= 1 && z <= 3 }
	if err := w.TickScheduled(eligible); err != nil {
		t.Fatal(err)
	}
	for x := int32(-19); x <= -15; x++ {
		for z := int32(45); z <= 49; z++ {
			if err := w.SetBlock(x, 60, z, StateStone); err != nil {
				t.Fatal(err)
			}
		}
	}
	sand, err := BlockStateID(block, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(-17, 64, 47, sand); err != nil {
		t.Fatal(err)
	}
	w.AdvanceTime(2)
	if err := w.TickScheduled(eligible); err != nil {
		t.Fatal(err)
	}
	entities, err := w.EntitiesInChunk(-2, 2)
	if err != nil || len(entities) != 1 {
		t.Fatalf("spawn: entities=%d err=%v", len(entities), err)
	}
	return w, dir, entities[0]
}

// Compare the persisted world implementation with the native entity's complete
// per-tick NBT, including a reload while airborne and a paused entity ticket.
func TestFallingEntityReloadAgainstVanilla(t *testing.T) {
	runFallingEntityReloadReference(t, "../../testdata/worldgen/falling_block_motion_1.21.1.jsonl.gz", "sand", false)
}
func TestBrushableEntityReloadAgainstVanilla(t *testing.T) {
	for _, block := range []string{"suspicious_sand", "suspicious_gravel"} {
		t.Run(block, func(t *testing.T) {
			runFallingEntityReloadReference(t, "../../testdata/worldgen/brushable_motion_1.21.1.jsonl.gz", block, true)
		})
	}
}
func runFallingEntityReloadReference(t *testing.T, fixture, block string, brushable bool) {
	w, dir, original := spawnTestFallingBlock(t, block)
	f, err := os.Open(fixture)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	sand, _ := BlockStateID(block, nil)
	scanner := bufio.NewScanner(gz)
	count := 0
	for scanner.Scan() {
		var row struct {
			State, Support uint32
			Tick           int
			Entities       []string
			Removed        []bool
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		if row.State != sand || row.Support != StateStone {
			continue
		}
		if row.Tick > 0 {
			w.AdvanceTime(1)
			if err := w.TickFallingEntities(func(x, z int32) bool { return x == -2 && z == 2 }); err != nil {
				t.Fatal(err)
			}
		}
		entities, err := w.EntitiesInChunk(-2, 2)
		if err != nil {
			t.Fatal(err)
		}
		if row.Removed[0] {
			if len(entities) != 0 {
				t.Fatalf("tick %d: landed entity retained", row.Tick)
			}
			id, err := w.BlockAt(-17, 61, 47)
			landing := sand
			if brushable {
				landing = StateAir
			}
			if err != nil || id != landing {
				t.Fatalf("landing state=%d err=%v", id, err)
			}
		} else {
			if len(entities) != 1 || entities[0].UUID != original.UUID {
				t.Fatalf("tick %d: entity identity changed", row.Tick)
			}
			got := entities[0].Data
			delete(got, "id")
			delete(got, "UUID")
			raw, err := hex.DecodeString(row.Entities[0])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("tick %d: persisted entity differs\ngot=%v\nwant=%v", row.Tick, got, want)
			}
		}
		if row.Tick == 5 {
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			w = fallingTestWorld(t, dir)
			before, err := w.EntitiesInChunk(-2, 2)
			if err != nil {
				t.Fatal(err)
			}
			if err := w.TickFallingEntities(func(int32, int32) bool { return false }); err != nil {
				t.Fatal(err)
			}
			after, err := w.EntitiesInChunk(-2, 2)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("non-entity-ticking chunk advanced")
			}
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 31 {
		t.Fatalf("incomplete native motion rows: %d", count)
	}
}

func TestFallingSpawnJournalRecovery(t *testing.T) {
	for _, stage := range []string{"prepared", "source_removed", "entity_written", "later_block_edit"} {
		t.Run(stage, func(t *testing.T) {
			w, dir, entity := spawnTestFallingSand(t)
			sand, _ := BlockStateID("sand", nil)
			state := uint32(StateAir)
			if stage == "prepared" {
				state = sand
			}
			if stage == "later_block_edit" {
				state = StateStone
			}
			w.blockUpdateMu.Lock()
			err := w.fallingEntityOperation(-2, 2, func(actor *Region, ch *Chunk) error {
				ch.SetBlock(15, 15, 64, state)
				entry := nbt.Object{
					"Source":      {Type: nbt.IntArray, Value: []int32{-17, 64, 47}},
					"Expected":    {Type: nbt.Int, Value: int32(sand)},
					"Replacement": {Type: nbt.Int, Value: int32(StateAir)},
					"Entity":      {Type: nbt.Compound, Value: entity.Data},
				}
				root, _ := ch.anvilRoot.Object()
				root[fallingJournal] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: entry}}}
				if stage != "entity_written" {
					if err := persistFallingEntities(actor, ch, nil); err != nil {
						return err
					}
				}
				return actor.world.store.vanilla.persistChunk(ch)
			})
			w.blockUpdateMu.Unlock()
			if err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			w = fallingTestWorld(t, dir)
			for repeat := 0; repeat < 2; repeat++ {
				if err := w.TickFallingEntities(func(int32, int32) bool { return false }); err != nil {
					t.Fatal(err)
				}
				got, err := w.EntitiesInChunk(-2, 2)
				if err != nil || len(got) != 1 || got[0].UUID != entity.UUID || !reflect.DeepEqual(got[0].Data, entity.Data) {
					t.Fatalf("recovery %d changed/duplicated entity: count=%d err=%v", repeat, len(got), err)
				}
				want := uint32(StateAir)
				if stage == "later_block_edit" {
					want = StateStone
				}
				id, err := w.BlockAt(-17, 64, 47)
				if err != nil || id != want {
					t.Fatalf("recovery source=%d want=%d err=%v", id, want, err)
				}
			}
			if err := w.fallingEntityOperation(-2, 2, func(_ *Region, ch *Chunk) error {
				root, _ := ch.anvilRoot.Object()
				if _, ok := root[fallingJournal]; ok {
					t.Error("committed spawn journal retained")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFallingSpawnInvalidJournalDoesNotEditWorld(t *testing.T) {
	w, _, entity := spawnTestFallingSand(t)
	sand, _ := BlockStateID("sand", nil)
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	if err := w.fallingEntityOperation(-2, 2, func(_ *Region, ch *Chunk) error {
		ch.SetBlock(15, 15, 64, sand)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	runtime := &fluidRuntime{world: w, settings: w.fluidSettings()}
	for _, field := range []string{"Source", "Expected", "Replacement", "Entity"} {
		t.Run(field, func(t *testing.T) {
			entry := nbt.Object{
				"Source":      {Type: nbt.IntArray, Value: []int32{-17, 64, 47}},
				"Expected":    {Type: nbt.Int, Value: int32(sand)},
				"Replacement": {Type: nbt.Int, Value: int32(StateAir)},
				"Entity":      {Type: nbt.Compound, Value: entity.Data},
			}
			delete(entry, field)
			if err := runtime.finishFallingSpawn(entry); err == nil {
				t.Fatal("accepted incomplete spawn journal")
			}
			id, err := w.BlockAt(-17, 64, 47)
			if err != nil || id != sand {
				t.Fatalf("invalid recovery removed source: %d %v", id, err)
			}
		})
	}
}
