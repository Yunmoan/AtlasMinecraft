package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"bufio"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Preserve comparisons of every persisted field while expecting the runtime
// provenance bit to change on an actual Anvil read.
func loadedEntitySnapshot(entity WorldEntity) WorldEntity {
	entity.LoadedFromDisk = true
	return entity
}

func loadedEntitySnapshots(entities []WorldEntity) []WorldEntity {
	result := append([]WorldEntity(nil), entities...)
	for i := range result {
		result[i] = loadedEntitySnapshot(result[i])
	}
	return result
}

func canonicalEntity(data nbt.Object) nbt.Object {
	attributes, _ := data["attributes"].Elements()
	sort.Slice(attributes, func(i, j int) bool {
		a, _ := attributes[i].Object()
		b, _ := attributes[j].Object()
		an, _ := a.Text("id")
		bn, _ := b.Text("id")
		return an < bn
	})
	return data
}

func TestGeneratedEntityDefaultsAndFinalizationAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/generated_entities_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cases := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 4 || fields[0] != "entity_default" && fields[0] != "entity_finalized" && fields[0] != "entity_finalized_fullmoon" {
			continue
		}
		seed, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := hex.DecodeString(fields[3])
		if err != nil {
			t.Fatal(err)
		}
		tag, used, err := nbt.ParseNetwork(encoded)
		if err != nil || used != len(encoded) {
			t.Fatal(err)
		}
		want, _ := tag.Object()
		got, ok, err := defaultGeneratedEntity(fields[1])
		if err != nil || !ok {
			t.Fatal(err)
		}
		if strings.HasPrefix(fields[0], "entity_finalized") {
			random := worldgen.NewXoroshiro(seed)
			data, err := worldgen.FinalizeSwampHutEntity(fields[1], random, fields[0] == "entity_finalized_fullmoon")
			if err != nil {
				t.Fatal(err)
			}
			for name, value := range data {
				got[name] = value
			}
			if !scanner.Scan() {
				t.Fatal("missing metadata fixture")
			}
			if !scanner.Scan() {
				t.Fatal("missing next random fixture")
			}
			next := strings.Fields(scanner.Text())
			expected, _ := strconv.ParseInt(next[3], 10, 64)
			gotNext := int64(random.NextLong())
			if next[0] != "entity_next" || gotNext != expected {
				t.Fatalf("entity random stream differs for %s seed%d: got%d want%d", fields[1], seed, gotNext, expected)
			}
		}
		if !reflect.DeepEqual(canonicalEntity(got), canonicalEntity(want)) {
			t.Fatalf("%s %s seed%d: got=%#v want=%#v", fields[0], fields[1], seed, got, want)
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 48 {
		t.Fatalf("matched %d fixtures, want48", cases)
	}
	t.Logf("matched %d actual JAR Entity.save/Mob.finalizeSpawn fixtures", cases)
}

func pendingEntityTag(id string, x, y, z int32, flag string) nbt.Tag {
	return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: id}, "x": {Type: nbt.Int, Value: x}, "y": {Type: nbt.Int, Value: y}, "z": {Type: nbt.Int, Value: z}, "PieceFlag": {Type: nbt.String, Value: flag}, "Persistent": {Type: nbt.Byte, Value: int8(1)}}}
}

func TestGeneratedEntityAnvilCommitReplayAndRestart(t *testing.T) {
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
	pending := []nbt.Tag{pendingEntityTag("minecraft:witch", -16, 70, 16, "Witch"), pendingEntityTag("minecraft:cat", -16, 70, 16, "Cat")}
	root["AtlasPendingStructureEntities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: pending}
	piece := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:tesh"}, "BB": {Type: nbt.IntArray, Value: []int32{-16, 68, 16, -10, 75, 24}}, "Witch": {Type: nbt.Byte, Value: int8(0)}, "Cat": {Type: nbt.Byte, Value: int8(0)}}
	root["structures"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"starts": {Type: nbt.Compound, Value: nbt.Object{"minecraft:swamp_hut": {Type: nbt.Compound, Value: nbt.Object{"Children": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: piece}}}}}}}}}
	if err := v.persistChunk(chunk); err != nil {
		t.Fatal(err)
	}
	if err := v.loadAndRealizeEntities(chunk); err != nil {
		t.Fatal(err)
	}
	if len(chunk.entities) != 2 {
		t.Fatal("entities not realized")
	}
	if _, ok := root["AtlasPendingStructureEntities"]; ok {
		t.Fatal("fulfilled requests were not cleared")
	}
	for _, flag := range []string{"Witch", "Cat"} {
		value, _ := piece.Number(flag)
		if value != 1 {
			t.Fatalf("%s flag was not committed", flag)
		}
	}
	_, entityRoot, exists, err := anvil.ReadChunk(filepath.Join(dir, "entities"), -1, 1)
	if err != nil || !exists {
		t.Fatal("entity Anvil file missing", err)
	}
	entityObject, _ := entityRoot.Object()
	position, _ := entityObject["Position"].Value.([]int32)
	if !reflect.DeepEqual(position, []int32{-1, 1}) {
		t.Fatal("wrong official entity chunk Position")
	}
	want := loadedEntitySnapshots(chunk.entities)
	// Simulate the block save remaining pending after the entity write reached
	// disk. A retry must merge UUIDs and retain unknown vanilla root fields.
	root["AtlasPendingStructureEntities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: pending}
	piece["Witch"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	if err := v.persistChunk(chunk); err != nil {
		t.Fatal(err)
	}
	loaded, err := v.loadChunk(-1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.entities, want) {
		t.Fatal("pending replay duplicated or changed saved entities")
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	entities, err := w.EntitiesInChunk(-1, 1)
	if err != nil || !reflect.DeepEqual(entities, want) {
		t.Fatal("world actor lost saved entities", err)
	}
	entities[0].Data["id"] = nbt.Tag{Type: nbt.String, Value: "corrupted-snapshot"}
	clean, err := w.EntitiesInChunk(-1, 1)
	if err != nil || clean[0].Type != "minecraft:witch" {
		t.Fatal("snapshot mutated actor state", err)
	}
	id, _ := clean[0].Data.Text("id")
	if id != "minecraft:witch" {
		t.Fatal("snapshot Data aliases actor")
	}
	if _, err := w.UnloadIdle(0); err != nil {
		t.Fatal(err)
	}
	reloaded, err := w.EntitiesInChunk(-1, 1)
	if err != nil || !reflect.DeepEqual(reloaded, want) {
		t.Fatal("unload/reload lost entities", err)
	}
}

func TestGeneratedEntityWriteFailureRetainsPendingAndFlags(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := v.generateFlat(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := chunk.anvilRoot.Object()
	root["AtlasPendingStructureEntities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{pendingEntityTag("minecraft:witch", 0, 70, 0, "Witch")}}
	// A file blocks creation of the entities directory, exercising a real
	// disk failure before any pending request is acknowledged.
	if err := os.WriteFile(filepath.Join(dir, "entities"), []byte("blocked"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := v.loadAndRealizeEntities(chunk); err == nil {
		t.Fatal("entity disk failure was ignored")
	}
	pending, _ := root["AtlasPendingStructureEntities"].Elements()
	if len(pending) != 1 || len(chunk.entities) != 0 {
		t.Fatal("failed entity write acknowledged requests")
	}
}
