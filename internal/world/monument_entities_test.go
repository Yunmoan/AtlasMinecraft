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

func TestMonumentEntitiesAnvilReplayAndRestart(t *testing.T) {
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
	for i := int32(0); i < 3; i++ {
		position := worldgen.FeaturePosition{X: -16 + i*4, Y: 42 + i, Z: 16 + i*3}
		request := worldgen.PendingStructureEntity{ID: "minecraft:elder_guardian", Position: position}
		data, err := worldgen.FinalizeMonumentEntity(request, random)
		if err != nil {
			t.Fatal(err)
		}
		tag := pendingEntityTag(request.ID, position.X, position.Y, position.Z, fmt.Sprintf("MonumentGuardian:0,0:%d,%d,%d", position.X, position.Y, position.Z))
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
	if len(chunk.entities) != 3 {
		t.Fatal("missing monument guardians")
	}
	if _, ok := root["AtlasPendingStructureEntities"]; ok {
		t.Fatal("fulfilled requests retained")
	}
	for i, entity := range chunk.entities {
		if entity.Type != "minecraft:elder_guardian" {
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
			t.Fatal("realization changed native guardian NBT", i)
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
		t.Fatal("guardian replay changed or duplicated actors", err)
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	entities, err := w.EntitiesInChunk(-1, 1)
	if err != nil || !reflect.DeepEqual(entities, want) {
		t.Fatal("guardian restart lost saved NBT", err)
	}
	entities[0].Data["Health"] = nbt.Tag{Type: nbt.Float, Value: float32(-1)}
	if _, err := w.UnloadIdle(0); err != nil {
		t.Fatal(err)
	}
	reloaded, err := w.EntitiesInChunk(-1, 1)
	if err != nil || !reflect.DeepEqual(reloaded, want) {
		t.Fatal("guardian snapshots alias saved state", err)
	}
}
