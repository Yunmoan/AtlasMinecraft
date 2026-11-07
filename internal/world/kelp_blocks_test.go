package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func prepareKelpTick(t *testing.T, drops bool) (*World, string, func(int32, int32) bool) {
	t.Helper()
	w, dir := mechanicsWorld(t)
	eligible := func(x, z int32) bool { return x == 0 && z == 0 }
	if err := w.TickScheduled(eligible); err != nil {
		t.Fatal(err)
	}
	v := w.store.vanilla
	v.mu.Lock()
	root, _ := v.root.Object()
	data, _ := root.Object("Data")
	rules, _ := data.Object("GameRules")
	value := "false"
	if drops {
		value = "true"
	}
	rules["doTileDrops"] = nbt.Tag{Type: nbt.String, Value: value}
	v.mu.Unlock()
	w.tickRandom = worldgen.NewLegacyRandom(1234)
	id, _ := BlockStateID("minecraft:kelp_plant", nil)
	if err := w.fallingEntityOperation(0, 0, func(_ *Region, ch *Chunk) error { ch.SetBlock(8, 8, 64, id); return nil }); err != nil {
		t.Fatal(err)
	}
	if err := w.scheduledTicks.scheduleBlock(worldgen.FeaturePosition{X: 8, Y: 64, Z: 8}, "minecraft:kelp_plant", 0, 0); err != nil {
		t.Fatal(err)
	}
	return w, dir, eligible
}

func TestKelpScheduledDestructionAndItemPersistence(t *testing.T) {
	for _, drops := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[drops], func(t *testing.T) {
			w, dir, eligible := prepareKelpTick(t, drops)
			if err := w.TickScheduled(eligible); err != nil {
				t.Fatal(err)
			}
			id, err := w.BlockAt(8, 64, 8)
			water, _ := BlockStateID("water", nil)
			if err != nil || id != water {
				t.Fatal("kelp did not restore source water", id, err)
			}
			entities, err := w.EntitiesInChunk(0, 0)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			if drops {
				count = 1
			}
			if len(entities) != count {
				t.Fatal("incorrect item count", len(entities))
			}
			random := worldgen.NewLegacyRandom(1234)
			want, err := worldgen.NewKelpBlockDropEntity(worldgen.FeaturePosition{X: 8, Y: 64, Z: 8}, drops, random, worldgen.NewLegacyRandom(0))
			if err != nil {
				t.Fatal(err)
			}
			if random.NextLong() != w.tickRandom.NextLong() {
				t.Fatal("runtime block-drop RNG differs")
			}
			if drops {
				for _, field := range []string{"Pos", "Motion", "Item", "PickupDelay", "Age"} {
					if !reflect.DeepEqual(entities[0].Data[field], want[field]) {
						t.Errorf("item %s differs", field)
					}
				}
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(noViewers{}, 1, dir)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			if _, err := reopened.LoadChunk(0, 0); err != nil {
				t.Fatal(err)
			}
			saved, err := reopened.EntitiesInChunk(0, 0)
			if err != nil || len(saved) != count {
				t.Fatal("reload item count", err)
			}
			if drops && (!reflect.DeepEqual(saved[0].Data, entities[0].Data) || saved[0].UUID != entities[0].UUID) {
				t.Fatal("item identity or NBT changed on reload")
			}
			if err := reopened.fallingEntityOperation(0, 0, func(_ *Region, ch *Chunk) error {
				root, _ := ch.anvilRoot.Object()
				if _, exists := root[kelpBreakJournal]; exists {
					t.Error("completed intent was retained")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestKelpItemWriteFailureRecovery(t *testing.T) {
	w, dir, eligible := prepareKelpTick(t, true)
	file := filepath.Join(dir, "entities", "r.0.0.mca")
	// The loaded empty native entity region has no persistent payload yet.
	if data, err := os.ReadFile(file); err == nil && len(data) != 0 {
		t.Fatal("expected empty entity region")
	}
	if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.MkdirAll(file, 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(file) })
	if err := w.TickScheduled(eligible); err == nil {
		t.Fatal("obstructed entity region unexpectedly succeeded")
	}
	id, err := w.BlockAt(8, 64, 8)
	kelp, _ := BlockStateID("kelp_plant", nil)
	if err != nil || id != kelp {
		t.Fatal("failed item write lost source", id, err)
	}
	var intended WorldEntity
	if err := w.fallingEntityOperation(0, 0, func(_ *Region, ch *Chunk) error {
		root, _ := ch.anvilRoot.Object()
		pending, _ := root.Object(kelpBreakJournal)
		entry, _ := pending.Object("8,64,8")
		entity, _ := entry.Object("Entity")
		var err error
		intended, err = worldEntityFromNBT(entity)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w, err = Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	for repeat := 0; repeat < 2; repeat++ {
		if err := w.TickScheduled(eligible); err != nil {
			t.Fatal(err)
		}
	}
	items, err := w.EntitiesInChunk(0, 0)
	if err != nil || len(items) != 1 || items[0].UUID != intended.UUID || !reflect.DeepEqual(items[0].Data, intended.Data) {
		t.Fatal("recovery changed or duplicated item", len(items), err)
	}
}

func TestKelpPublishedItemRecoveryPreservesLaterEdits(t *testing.T) {
	for _, stage := range []string{"source_retained", "source_removed", "later_edit"} {
		t.Run(stage, func(t *testing.T) {
			w, _, eligible := prepareKelpTick(t, true)
			if err := w.TickScheduled(eligible); err != nil {
				t.Fatal(err)
			}
			items, err := w.EntitiesInChunk(0, 0)
			if err != nil || len(items) != 1 {
				t.Fatal("initial item", err)
			}
			kelp, _ := BlockStateID("kelp_plant", nil)
			water, _ := BlockStateID("water", nil)
			state := uint32(water)
			if stage == "source_retained" {
				state = kelp
			}
			if stage == "later_edit" {
				state = StateStone
			}
			if err := w.fallingEntityOperation(0, 0, func(actor *Region, ch *Chunk) error {
				ch.SetBlock(8, 8, 64, state)
				entry := nbt.Object{"Source": {Type: nbt.IntArray, Value: []int32{8, 64, 8}}, "Expected": {Type: nbt.Int, Value: int32(kelp)}, "Entity": {Type: nbt.Compound, Value: items[0].Data}}
				return updateKelpJournal(actor, ch, "8,64,8", entry)
			}); err != nil {
				t.Fatal(err)
			}
			for repeat := 0; repeat < 2; repeat++ {
				if err := w.TickScheduled(eligible); err != nil {
					t.Fatal(err)
				}
			}
			got, err := w.EntitiesInChunk(0, 0)
			if err != nil || len(got) != 1 || got[0].UUID != items[0].UUID {
				t.Fatal("published item was duplicated or replaced", err)
			}
			if state == kelp {
				state = water
			}
			id, err := w.BlockAt(8, 64, 8)
			if err != nil || id != state {
				t.Fatal("recovery lost source state/later edit", id, state, err)
			}
		})
	}
}
