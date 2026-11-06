package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMansionEntitiesAnvilReplayAndRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
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
	random := worldgen.NewXoroshiro(seed)
	var pending []nbt.Tag
	ids := []string{"minecraft:evoker", "minecraft:vindicator", "minecraft:allay", "minecraft:allay"}
	for i, id := range ids {
		position := worldgen.FeaturePosition{X: -16, Y: 70, Z: 16}
		request := worldgen.PendingStructureEntity{ID: id, Position: position}
		data, err := worldgen.FinalizeMansionEntity(request, random)
		if err != nil {
			t.Fatal(err)
		}
		tag := pendingEntityTag(request.ID, position.X, position.Y, position.Z, fmt.Sprintf("Mansion:0,0:piece:marker:%d", i))
		object, _ := tag.Object()
		object["Data"] = nbt.Tag{Type: nbt.Compound, Value: data}
		pending = append(pending, tag)
	}
	root["AtlasPendingStructureEntities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: pending}
	if err := v.persistChunk(chunk); err != nil {
		t.Fatal(err)
	}
	if err := v.loadAndRealizeEntities(chunk); err != nil {
		t.Fatal(err)
	}
	if len(chunk.entities) != 4 {
		t.Fatal("missing mansion actors")
	}
	if _, ok := root["AtlasPendingStructureEntities"]; ok {
		t.Fatal("fulfilled requests retained")
	}
	for i, entity := range chunk.entities {
		if entity.Type != ids[i] {
			t.Fatal(entity.Type)
		}
		data, err := cloneEntityData(entity.Data)
		if err != nil {
			t.Fatal(err)
		}
		delete(data, "UUID")
		request, _ := pending[i].Object()
		expected, _ := request.Object("Data")
		if !reflect.DeepEqual(canonicalEntity(data), canonicalEntity(expected)) {
			t.Fatal("realization changed native mansion actor NBT", i)
		}
	}
	_, entityRoot, exists, err := anvil.ReadChunk(filepath.Join(dir, "entities"), -1, 1)
	if err != nil || !exists {
		t.Fatal("missing entity Anvil", err)
	}
	object, _ := entityRoot.Object()
	if !reflect.DeepEqual(object["Position"].Value, []int32{-1, 1}) {
		t.Fatal("wrong entity ownership")
	}
	want := loadedEntitySnapshots(chunk.entities)
	// Replay an interrupted terrain commit after the entity region reached disk.
	root["AtlasPendingStructureEntities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: pending}
	if err := v.persistChunk(chunk); err != nil {
		t.Fatal(err)
	}
	loaded, err := v.loadChunk(-1, 1)
	if err != nil || !reflect.DeepEqual(loaded.entities, want) {
		t.Fatal("mansion actor replay changed or duplicated actors", err)
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	entities, err := w.EntitiesInChunk(-1, 1)
	if err != nil || !reflect.DeepEqual(entities, want) {
		t.Fatal("mansion actor restart lost saved NBT", err)
	}
	entities[0].Data["Health"] = nbt.Tag{Type: nbt.Float, Value: float32(-1)}
	if _, err := w.UnloadIdle(0); err != nil {
		t.Fatal(err)
	}
	reloaded, err := w.EntitiesInChunk(-1, 1)
	if err != nil || !reflect.DeepEqual(reloaded, want) {
		t.Fatal("mansion actor snapshots alias saved state", err)
	}
}
