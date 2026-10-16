package world

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStrongholdNormalCreationRemoteFlagsAnvilRestart(t *testing.T) {
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
	w.store.vanilla.normal.SetTerrainWorkers(1)
	positions := [][2]int32{{-100, -125}, {-99, -125}, {-102, -124}, {-101, -123}, {-101, -126}}
	read := func(position [2]int32) nbt.Object {
		t.Helper()
		_, tag, found, err := anvil.ReadChunk(filepath.Join(dir, "region"), position[0], position[1])
		if err != nil || !found {
			t.Fatal("actual stronghold chunk missing from Anvil", position, err)
		}
		root, ok := tag.Object()
		if !ok {
			t.Fatal("saved chunk root is not a compound")
		}
		return root
	}
	ownerStart := func(root nbt.Object) nbt.Object {
		t.Helper()
		structures, _ := root.Object("structures")
		starts, _ := structures.Object("starts")
		start, found := starts.Object("minecraft:stronghold")
		children, _ := start["Children"].Elements()
		if !found || len(children) != 129 {
			t.Fatal("actual first-ring stronghold lost its full 129-piece start")
		}
		return start
	}
	portalFlag := func(start nbt.Object) int64 {
		t.Helper()
		children, _ := start["Children"].Elements()
		for _, tag := range children {
			piece, _ := tag.Object()
			id, _ := piece.Text("id")
			if id == "minecraft:shpr" {
				flag, _ := piece.Number("Mob")
				return flag
			}
		}
		t.Fatal("original portal room missing from saved start")
		return -1
	}
	// Persist the owner before any remote portal/library/chest clip exists.
	if _, err := w.LoadChunk(positions[0][0], positions[0][1]); err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(-100*16+8, 280, -125*16+8, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	initial := ownerStart(read(positions[0]))
	if portalFlag(initial) != 0 {
		t.Fatal("owner saved a placed portal spawner before its remote chunk existed")
	}
	for _, position := range positions[1:] {
		if _, err := w.LoadChunk(position[0], position[1]); err != nil {
			t.Fatal(err)
		}
	}
	wantedStart, found := w.store.vanilla.normal.StrongholdStartSnapshot(-100, -125)
	if !found {
		t.Fatal("real generated clips have no coalesced metadata snapshot")
	}
	wantedNBT, err := wantedStart.NBT()
	if err != nil || portalFlag(wantedNBT) != 1 {
		t.Fatal("real portal clip did not set the original Mob flag", err)
	}
	wantedEntities := make(map[[3]int32]nbt.Object)
	for _, position := range positions {
		generated, err := w.store.vanilla.normal.Generate(position[0], position[1])
		if err != nil {
			t.Fatal(err)
		}
		for _, data := range generated.BlockEntities {
			id, _ := data.Text("id")
			table, _ := data.Text("LootTable")
			if id != "minecraft:mob_spawner" && table != "minecraft:chests/stronghold_library" && table != "minecraft:chests/stronghold_corridor" && table != "minecraft:chests/stronghold_crossing" {
				continue
			}
			x, _ := data.Number("x")
			y, _ := data.Number("y")
			z, _ := data.Number("z")
			wantedEntities[[3]int32{int32(x), int32(y), int32(z)}] = data
		}
	}
	if _, found := wantedEntities[[3]int32{-1606, -7, -2006}]; !found {
		t.Fatal("actual portal room has no original silverfish spawner NBT")
	}
	for _, p := range [][3]int32{{-1620, -5, -1971}, {-1616, 0, -1962}, {-1579, -12, -1994}} {
		if _, found := wantedEntities[p]; !found {
			t.Fatal("actual library/corridor chest missing", p)
		}
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	if saved := ownerStart(read(positions[0])); !reflect.DeepEqual(saved, wantedNBT) {
		t.Fatal("Save failed to refresh already persisted owner flags/full original NBT")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	saved := make(map[[2]int32]nbt.Object)
	seen := make(map[[3]int32]bool)
	for _, position := range positions {
		root := read(position)
		saved[position] = root
		structures, _ := root.Object("structures")
		refs, _ := structures.Object("References")
		if refs["minecraft:stronghold"].Type != nbt.LongArray {
			t.Fatal("stronghold owner references missing from actual Anvil chunk")
		}
		entities, _ := root["block_entities"].Elements()
		for _, tag := range entities {
			data, _ := tag.Object()
			x, _ := data.Number("x")
			y, _ := data.Number("y")
			z, _ := data.Number("z")
			p := [3]int32{int32(x), int32(y), int32(z)}
			if want, found := wantedEntities[p]; found {
				if !reflect.DeepEqual(data, want) {
					t.Fatal("saved full original seeded chest/spawner NBT changed", p)
				}
				seen[p] = true
			}
		}
	}
	if len(seen) != len(wantedEntities) {
		t.Fatal("actual generated stronghold block entities were not all persisted")
	}
	reopened, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for _, position := range positions {
		chunk, err := reopened.store.vanilla.loadBlockChunk(position[0], position[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := chunk.anvilRoot.Object()
		if !reflect.DeepEqual(root["structures"], saved[position]["structures"]) || !reflect.DeepEqual(root["block_entities"], saved[position]["block_entities"]) {
			t.Fatal("restart changed full stronghold start/references/seeded loot/spawner", position)
		}
	}
	for p, data := range wantedEntities {
		id, err := reopened.BlockAt(p[0], p[1], p[2])
		if err != nil {
			t.Fatal(err)
		}
		state, err := blockstate.BlockStateByID(id)
		entityID, _ := data.Text("id")
		want := "minecraft:chest"
		if entityID == "minecraft:mob_spawner" {
			want = "minecraft:spawner"
		}
		if err != nil || state.Name != want {
			t.Fatal("restart loaded stronghold metadata without actual structure blocks", p, err)
		}
	}
	if id, err := reopened.BlockAt(-100*16+8, 280, -125*16+8); err != nil || id != StateStone {
		t.Fatal("restart regenerated the player-edited owner chunk", err)
	}
	t.Logf("created normal world, persisted owner before four remote clips, saved/reopened 129 pieces and %d full original seeded chest/spawner entities; Save merged Mob/Chest flags and player edit survived", len(wantedEntities))
}
