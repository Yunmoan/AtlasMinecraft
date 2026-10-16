package world

import (
	"atlas/internal/anvil"
	"atlas/internal/item"
	"atlas/internal/nbt"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func armorFixtureStack(t *testing.T, v string) item.Stack {
	t.Helper()
	f := strings.Split(v, ":")
	id, e1 := strconv.Atoi(f[0])
	count, e2 := strconv.Atoi(f[1])
	if e1 != nil || e2 != nil {
		t.Fatal(v)
	}
	return item.New(int32(id), int32(count))
}
func armorFixtureNBT(t *testing.T, slots [6]item.Stack) nbt.Object {
	t.Helper()
	data := nbt.Object{}
	for _, field := range []struct {
		name       string
		start, end int
	}{{"HandItems", 0, 2}, {"ArmorItems", 2, 6}} {
		list := []nbt.Tag{}
		for _, s := range slots[field.start:field.end] {
			o := nbt.Object{}
			if !s.Empty() {
				name, _ := item.Name(s.ID)
				o["id"] = nbt.Tag{Type: nbt.String, Value: name}
				o["count"] = nbt.Tag{Type: nbt.Int, Value: s.Count}
			}
			list = append(list, nbt.Tag{Type: nbt.Compound, Value: o})
		}
		data[field.name] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: list}
	}
	return data
}

func TestArmorStandInteractionAgainstVanillaJAR(t *testing.T) {
	source, err := os.ReadFile("../../docs/worldgen/armor_stand_interaction_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{"diamond_sword", "shield", "iron_boots", "iron_leggings", "iron_chestplate", "iron_helmet"}
	equipment := [6]item.Stack{}
	for i, name := range ids {
		id, _ := item.ID(name)
		equipment[i] = item.New(id, 1)
	}
	cases, slotCases := 0, 0
	for _, line := range strings.Split(string(source), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if f[0] == "armor_item_slot" {
			id, _ := strconv.Atoi(f[1])
			slot, _ := strconv.Atoi(f[2])
			if got := ArmorStandItemSlot(item.New(int32(id), 1)); got != slot {
				t.Fatalf("item %d slot=%d want%d", id, got, slot)
			}
			slotCases++
			continue
		}
		if f[0] != "armor_interact" {
			continue
		}
		if len(f) != 21 {
			t.Fatalf("invalid vector %s", line)
		}
		number := func(i int) int {
			v, e := strconv.Atoi(f[i])
			if e != nil {
				t.Fatal(e)
			}
			return v
		}
		boolean := func(i int) bool {
			v, e := strconv.ParseBool(f[i])
			if e != nil {
				t.Fatal(e)
			}
			return v
		}
		y, e := strconv.ParseFloat(f[11], 64)
		if e != nil {
			t.Fatal(e)
		}
		slots := [6]item.Stack{}
		for i, s := range equipment {
			if number(12)&(1<<i) != 0 {
				slots[i] = s
			}
		}
		data := armorFixtureNBT(t, slots)
		for key, index := range map[string]int{"Small": 2, "ShowArms": 3, "Marker": 4} {
			var v int8
			if boolean(index) {
				v = 1
			}
			data[key] = nbt.Tag{Type: nbt.Byte, Value: v}
		}
		data["DisabledSlots"] = nbt.Tag{Type: nbt.Int, Value: int32(number(5))}
		held := item.New(int32(number(7)), int32(number(8)))
		after, result, changed, err := ArmorStandInteraction(data, held, y, boolean(9), boolean(10))
		if err != nil {
			t.Fatal(line, err)
		}
		wantHeld := armorFixtureStack(t, f[14])
		if result != f[13] || after.ID != wantHeld.ID || after.Count != wantHeld.Count {
			t.Fatalf("case%s result=%s held=%d:%d want%s %s", f[1], result, after.ID, after.Count, f[13], f[14])
		}
		got, err := armorStandEquipment(data)
		if err != nil {
			t.Fatal(err)
		}
		differs := false
		for i, s := range got {
			want := armorFixtureStack(t, f[15+i])
			if s.ID != want.ID || s.Count != want.Count {
				t.Fatalf("case%s slot%d got%d:%d want%s", f[1], i, s.ID, s.Count, f[15+i])
			}
			differs = differs || s.ID != slots[i].ID || s.Count != slots[i].Count
		}
		if differs && !changed {
			t.Fatal("equipment changed without mutation signal", line)
		}
		cases++
	}
	if cases != 4096 || slotCases != 1332 {
		t.Fatalf("incomplete official vectors %d interactions/%d items", cases, slotCases)
	}
	t.Logf("matched %d actual interactAt vectors and %d official item equipment slots", cases, slotCases)
}

func armorStandSavedWorld(t *testing.T) (*World, string, [16]byte, item.Stack) {
	t.Helper()
	dir := t.TempDir()
	seed := int64(1)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	data, ok, err := defaultGeneratedEntity("minecraft:armor_stand")
	if err != nil || !ok {
		t.Fatal(err)
	}
	uuid := [16]byte{44}
	data["UUID"] = entityUUIDTag(uuid)
	data["Pos"] = entityPositionTag([3]float64{-.5, -60, .5})
	id, _ := item.ID("iron_boots")
	stack, err := item.FromNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:iron_boots"}, "count": {Type: nbt.Int, Value: int32(1)}, "ForeignItemField": {Type: nbt.String, Value: "kept"}}})
	if err != nil {
		t.Fatal(err)
	}
	slots := [6]item.Stack{}
	slots[2] = item.New(id, 1)
	equipment := armorFixtureNBT(t, slots)
	data["HandItems"], data["ArmorItems"] = equipment["HandItems"], equipment["ArmorItems"]
	armor, _ := data["ArmorItems"].Elements()
	tag, err := stack.ToNBT()
	if err != nil {
		t.Fatal(err)
	}
	armor[0] = tag
	data["ForeignActorField"] = nbt.Tag{Type: nbt.String, Value: "kept"}
	root := nbt.Object{"DataVersion": {Type: nbt.Int, Value: int32(3955)}, "Position": {Type: nbt.IntArray, Value: []int32{-1, 0}}, "Entities": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: data}}}, "ForeignEntityHeader": {Type: nbt.String, Value: "kept"}}
	if err := anvil.WriteChunk(filepath.Join(dir, "entities"), -1, 0, "foreign", nbt.Tag{Type: nbt.Compound, Value: root}); err != nil {
		t.Fatal(err)
	}
	w, err := Open(nil, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	if _, err := w.LoadChunk(-1, 0); err != nil {
		t.Fatal(err)
	}
	return w, dir, uuid, stack
}

func TestArmorStandConcurrentRemovalAndAnvilRestart(t *testing.T) {
	w, dir, uuid, original := armorStandSavedWorld(t)
	var wg sync.WaitGroup
	results := make(chan ArmorStandSwap, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			swap, err := w.InteractLoadedArmorStand(ArmorStandClick{CX: -1, CZ: 0, UUID: uuid, Point: [3]float32{0, .2, 0}, Eye: [3]float64{.5, -58.38, .5}, Creative: true})
			if err != nil {
				t.Error(err)
				return
			}
			results <- swap
		}()
	}
	wg.Wait()
	close(results)
	taken := 0
	for swap := range results {
		if !swap.Held.Empty() {
			taken++
			_, originalNBT, originalErr := nbt.Parse(original.Saved)
			_, receivedNBT, receivedErr := nbt.Parse(swap.Held.Saved)
			if originalErr != nil || receivedErr != nil || !original.SameItem(swap.Held) || !reflect.DeepEqual(originalNBT, receivedNBT) {
				t.Fatal("saved item fields changed during swap")
			}
		}
	}
	if taken != 1 {
		t.Fatalf("%d players took the same equipment", taken)
	}
	actors, err := w.EntitiesInChunk(-1, 0)
	if err != nil || len(actors) != 1 {
		t.Fatal(actors, err)
	}
	slots, err := armorStandEquipment(actors[0].Data)
	if err != nil || !slots[2].Empty() || actors[0].Revision != 1 {
		t.Fatal("actor did not commit exactly one exchange", err)
	}
	actors[0].Data["ForeignActorField"] = nbt.Tag{Type: nbt.String, Value: "mutated outside actor"}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Open(nil, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reloaded.Close()
	actors, err = reloaded.EntitiesInChunk(-1, 0)
	if err != nil {
		t.Fatal(err)
	}
	slots, err = armorStandEquipment(actors[0].Data)
	if err != nil || !slots[2].Empty() {
		t.Fatal("equipment returned after restart", err)
	}
	if v, _ := actors[0].Data.Text("ForeignActorField"); v != "kept" {
		t.Fatal("unknown actor data lost or snapshot was mutable")
	}
	name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "entities"), -1, 0)
	if err != nil || !exists || name != "foreign" {
		t.Fatal("foreign root lost", name, err)
	}
	o, _ := root.Object()
	if v, _ := o.Text("ForeignEntityHeader"); v != "kept" {
		t.Fatal("unknown entity root field lost")
	}
}

func TestArmorStandRejectsRemoteMissingAndFailedSourceSave(t *testing.T) {
	w, dir, uuid, _ := armorStandSavedWorld(t)
	click := ArmorStandClick{CX: -1, CZ: 0, UUID: uuid, Point: [3]float32{0, .2, 0}, Eye: [3]float64{.5, -58.38, .5}, Creative: true}
	bad := click
	bad.Eye[0] = 100
	if _, err := w.InteractLoadedArmorStand(bad); !errors.Is(err, ErrEntityInteractionRange) {
		t.Fatal(err)
	}
	bad = click
	bad.CX = 123456
	if _, err := w.InteractLoadedArmorStand(bad); !errors.Is(err, ErrChunkNotLoaded) {
		t.Fatal(err)
	}
	bad = click
	bad.UUID[0]++
	if _, err := w.InteractLoadedArmorStand(bad); !errors.Is(err, ErrEntityNotLoaded) {
		t.Fatal(err)
	}
	// A corrupt destination/source file must leave live equipment unchanged.
	path := filepath.Join(dir, "entities", "r.-1.0.mca")
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := w.InteractLoadedArmorStand(click); err == nil {
		t.Fatal("bad source unexpectedly accepted")
	}
	actors, err := w.EntitiesInChunk(-1, 0)
	if err != nil {
		t.Fatal(err)
	}
	slots, err := armorStandEquipment(actors[0].Data)
	if err != nil || slots[2].Empty() || actors[0].Revision != 0 {
		t.Fatal("failed commit mutated live actor", err)
	}
}
