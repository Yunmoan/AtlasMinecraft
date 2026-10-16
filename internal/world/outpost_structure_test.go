package world

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestOutpostStructureAnvilRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	metadataBytes, err := readGzip(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	_, metadata, err := nbt.Parse(metadataBytes)
	if err != nil {
		t.Fatal(err)
	}
	positions := [][2]int32{{-249, 327}, {-248, 328}, {-250, 328}, {-248, 326}}
	pending := make(map[[2]int32]nbt.Tag)
	wantActors := make(map[[16]byte]WorldEntity)
	var wantChest nbt.Object
	for _, pos := range positions {
		generated, err := w.store.vanilla.normal.Generate(pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		chunk := newChunkBounds(pos[0], pos[1], -64, 384, false)
		if err := chunk.initializeAnvilRoot(); err != nil {
			t.Fatal(err)
		}
		installGeneratedStructures(chunk, generated)
		root, _ := chunk.anvilRoot.Object()
		pending[pos] = root["AtlasPendingStructureEntities"]
		requests, _ := pending[pos].Elements()
		for _, tag := range requests {
			request, _ := tag.Object()
			id, _ := request.Text("id")
			if id != "minecraft:allay" && id != "minecraft:iron_golem" {
				continue
			}
			actor, supported, err := realizeGeneratedEntity(seed, "minecraft:overworld", request)
			if err != nil || !supported {
				t.Fatal("authored outpost actor unsupported by runtime", err)
			}
			if _, duplicate := wantActors[actor.UUID]; duplicate {
				t.Fatal("two authored actors aliased to one UUID")
			}
			wantActors[actor.UUID] = actor
		}
		for _, data := range generated.BlockEntities {
			table, _ := data.Text("LootTable")
			if table == "minecraft:chests/pillager_outpost" {
				wantChest = data
			}
		}
	}
	if len(wantActors) != 5 || wantChest == nil {
		t.Fatal("pinned outpost lost five authored actors or tower loot")
	}
	errors := make([]error, len(positions))
	var wg sync.WaitGroup
	for i, pos := range positions {
		wg.Add(1)
		go func(i int, pos [2]int32) {
			defer wg.Done()
			_, errors[i] = w.LoadChunk(pos[0], pos[1])
		}(i, pos)
	}
	wg.Wait()
	for _, err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	checkActors := func(w *World, loaded bool) {
		t.Helper()
		seen := make(map[[16]byte]bool)
		allays, golems := 0, 0
		for _, pos := range positions {
			actors, err := w.EntitiesInChunk(pos[0], pos[1])
			if err != nil {
				t.Fatal(err)
			}
			for _, actor := range actors {
				if actor.Type != "minecraft:allay" && actor.Type != "minecraft:iron_golem" {
					continue
				}
				want := wantActors[actor.UUID]
				if loaded {
					want = loadedEntitySnapshot(want)
				}
				if seen[actor.UUID] || !reflect.DeepEqual(actor, want) {
					t.Fatal("outpost actor UUID/full NBT changed or duplicated")
				}
				seen[actor.UUID] = true
				health, _ := actor.Data["Health"].Value.(float32)
				if actor.Type == "minecraft:allay" {
					allays++
					duplicate, _ := actor.Data.Number("CanDuplicate")
					// The original cage template saves Health10 even though a
					// newly constructed allay has Health20.
					if health != 10 || duplicate != 0 {
						t.Fatal("allay lost original authored Health10/load defaults")
					}
				} else {
					golems++
					if health != 100 {
						t.Fatal("cage golem lost Health100")
					}
				}
			}
		}
		if len(seen) != 5 || allays != 4 || golems != 1 {
			t.Fatal("outpost actors missing", allays, golems)
		}
	}
	checkActors(w, false)
	id, err := w.BlockAt(-3974, 109, 5242)
	if err != nil {
		t.Fatal(err)
	}
	state, err := blockstate.BlockStateByID(id)
	if err != nil || state.Name != "minecraft:chest" {
		t.Fatal("outpost start metadata without actual tower chest", err)
	}
	if err := w.SetBlock(-249*16+8, 280, 327*16+8, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	saved := make(map[[2]int32]nbt.Object)
	savedActorCount, chestCount := 0, 0
	for _, pos := range positions {
		_, tag, found, err := anvil.ReadChunk(filepath.Join(dir, "region"), pos[0], pos[1])
		if err != nil || !found {
			t.Fatal("outpost block chunk not saved", err)
		}
		root, _ := tag.Object()
		saved[pos] = root
		if _, retained := root["AtlasPendingStructureEntities"]; retained {
			t.Fatal("outpost actors left pending after realization")
		}
		structures, _ := root.Object("structures")
		refs, _ := structures.Object("References")
		if refs["minecraft:pillager_outpost"].Type != nbt.LongArray {
			t.Fatal("outpost references not saved")
		}
		if pos == positions[0] {
			starts, _ := structures.Object("starts")
			start, found := starts.Object("minecraft:pillager_outpost")
			children, _ := start["Children"].Elements()
			if !found || len(children) != 15 {
				t.Fatal("outpost saved start lost its 15 pieces")
			}
		}
		entities, _ := root["block_entities"].Elements()
		for _, tag := range entities {
			data, _ := tag.Object()
			table, _ := data.Text("LootTable")
			if table == "minecraft:chests/pillager_outpost" {
				chestCount++
				if !reflect.DeepEqual(data, wantChest) {
					t.Fatal("outpost saved loot table/full chest NBT changed")
				}
			}
		}
		_, entityTag, found, err := anvil.ReadChunk(filepath.Join(dir, "entities"), pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		if found {
			entityRoot, _ := entityTag.Object()
			actors, _ := entityRoot["Entities"].Elements()
			for _, tag := range actors {
				data, _ := tag.Object()
				id, _ := data.Text("id")
				if id == "minecraft:allay" || id == "minecraft:iron_golem" {
					savedActorCount++
					if data["UUID"].Type != nbt.IntArray {
						t.Fatal("outpost entity region lost UUID")
					}
				}
			}
		}
		// Recreate a crash window: actors reached entities/ before their
		// requests were cleared in region/. Replaying must merge by UUID.
		requests, _ := pending[pos].Elements()
		if len(requests) > 0 {
			root["AtlasPendingStructureEntities"] = pending[pos]
			if err := w.store.vanilla.regions.WriteChunk(filepath.Join(dir, "region"), pos[0], pos[1], "", tag); err != nil {
				t.Fatal(err)
			}
		}
	}
	if savedActorCount != 5 || chestCount != 1 {
		t.Fatal("outpost actor/chest Anvil coverage changed", savedActorCount, chestCount)
	}
	reopened, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	checkActors(reopened, true)
	for _, pos := range positions {
		chunk, err := reopened.store.vanilla.loadBlockChunk(pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := chunk.anvilRoot.Object()
		if _, retained := root["AtlasPendingStructureEntities"]; retained {
			t.Fatal("outpost crash replay was not acknowledged")
		}
		if !reflect.DeepEqual(root["structures"], saved[pos]["structures"]) || !reflect.DeepEqual(root["block_entities"], saved[pos]["block_entities"]) {
			t.Fatal("restart changed full outpost start/refs/chest NBT")
		}
	}
	id, err = reopened.BlockAt(-249*16+8, 280, 327*16+8)
	if err != nil || id != StateStone {
		t.Fatal("restart regenerated edited outpost chunk", err)
	}
	afterBytes, err := readGzip(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	_, after, err := nbt.Parse(afterBytes)
	if err != nil || !reflect.DeepEqual(metadata, after) {
		t.Fatal("outpost generation changed root level.dat tags", err)
	}
	t.Log("saved/reopened four actual overworld chunks,15 outpost pieces,one seeded tower chest,four allays and one golem with full UUID/NBT; crash-pending replay deduplicated and player edits preserved")
}
