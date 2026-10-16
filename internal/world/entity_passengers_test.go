package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedPassengerAnvilReplayRestartAndIsolation(t *testing.T) {
	fixture, err := os.ReadFile("../../docs/worldgen/structure_drowned_actors_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var data nbt.Object
	for _, line := range strings.Split(string(fixture), "\n") {
		f := strings.Fields(line)
		if len(f) != 4 || f[0] != "drowned_actor_extra" {
			continue
		}
		raw, err := hex.DecodeString(f[3])
		if err != nil {
			t.Fatal(err)
		}
		tag, _, err := nbt.ParseNetwork(raw)
		if err != nil {
			t.Fatal(err)
		}
		data, _ = tag.Object()
		break
	}
	if data == nil {
		t.Fatal("missing actual vanilla chicken jockey")
	}
	seed := int64(1234)
	dir := filepath.Join(t.TempDir(), "world")
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := v.generateFlat(-1, 1)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := chunk.anvilRoot.Object()
	pending, _ := pendingEntityTag("minecraft:chicken", -16, 50, 16, "OceanRuinDrowned:-1,1:0:0").Object()
	pending["Persistent"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	pending["Data"] = nbt.Tag{Type: nbt.Compound, Value: data}
	pendingList := nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: pending}}}
	root["AtlasPendingStructureEntities"] = pendingList
	if err := v.loadAndRealizeEntities(chunk); err != nil {
		t.Fatal(err)
	}
	if len(chunk.entities) != 1 {
		t.Fatal("jockey was flattened or dropped", len(chunk.entities))
	}
	vehicle := chunk.entities[0]
	children, err := EntityPassengers(vehicle)
	if err != nil || len(children) != 1 {
		t.Fatal(children, err)
	}
	child := children[0]
	if child.Type != "minecraft:drowned" || child.UUID == vehicle.UUID || child.UUID == [16]byte{} {
		t.Fatal("passenger identity", child)
	}
	if _, ok := root["AtlasPendingStructureEntities"]; ok {
		t.Fatal("fulfilled jockey pending retained")
	}
	_, saved, exists, err := anvil.ReadChunk(filepath.Join(dir, "entities"), -1, 1)
	if err != nil || !exists {
		t.Fatal(exists, err)
	}
	savedRoot, _ := saved.Object()
	savedActors, _ := savedRoot["Entities"].Elements()
	if len(savedActors) != 1 {
		t.Fatal("vanilla entities chunk must retain vehicle root")
	}
	root["AtlasPendingStructureEntities"] = pendingList
	if err := v.persistChunk(chunk); err != nil {
		t.Fatal(err)
	}
	loaded, err := v.loadChunk(-1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.entities, loadedEntitySnapshots(chunk.entities)) {
		t.Fatal("retry changed or duplicated nested jockey")
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	check := func() {
		entities, err := w.EntitiesInChunk(-1, 1)
		if err != nil || len(entities) != 1 {
			t.Fatal(entities, err)
		}
		passengers, err := EntityPassengers(entities[0])
		if err != nil || len(passengers) != 1 || !reflect.DeepEqual(passengers[0], loadedEntitySnapshot(child)) {
			t.Fatal("restart/unload lost saved passenger", passengers, err)
		}
		passengers[0].Data["Health"] = nbt.Tag{Type: nbt.Float, Value: float32(0)}
		again, err := EntityPassengers(entities[0])
		if err != nil || !reflect.DeepEqual(again[0], loadedEntitySnapshot(child)) {
			t.Fatal("passenger snapshot aliases saved vehicle")
		}
	}
	check()
	if _, err := w.UnloadIdle(0); err != nil {
		t.Fatal(err)
	}
	check()
}
