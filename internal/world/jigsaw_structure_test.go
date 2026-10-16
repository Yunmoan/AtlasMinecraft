package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"path/filepath"
	"reflect"
	"testing"
)

func TestOverworldJigsawStructureAnvilRestart(t *testing.T) {
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
	positions := [][2]int32{{44, 38}, {43, 35}, {13, 21}, {14, 22}, {13, 60}, {10, 53}}
	for _, position := range positions {
		if _, err := w.LoadChunk(position[0], position[1]); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.SetBlock(44*16+8, 280, 38*16+8, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	saved := make(map[[2]int32]nbt.Object)
	blockTypes := make(map[string]int)
	for _, position := range positions {
		_, tag, found, err := anvil.ReadChunk(filepath.Join(dir, "region"), position[0], position[1])
		if err != nil || !found {
			t.Fatal("Jigsaw chunk not saved", err)
		}
		root, _ := tag.Object()
		saved[position] = root
		entities, _ := root["block_entities"].Elements()
		for _, tag := range entities {
			object, _ := tag.Object()
			id, _ := object.Text("id")
			blockTypes[id]++
		}
	}
	for _, entry := range []struct {
		position [2]int32
		name     string
		pieces   int
	}{
		{[2]int32{44, 38}, "minecraft:trail_ruins", 17},
		{[2]int32{13, 21}, "minecraft:trial_chambers", 226},
		{[2]int32{13, 60}, "minecraft:ancient_city", 87},
	} {
		structures, _ := saved[entry.position].Object("structures")
		starts, _ := structures.Object("starts")
		start, found := starts.Object(entry.name)
		children, _ := start["Children"].Elements()
		if !found || len(children) != entry.pieces {
			t.Fatalf("saved%s children%d !=%d", entry.name, len(children), entry.pieces)
		}
		refs, _ := structures.Object("References")
		if refs[entry.name].Type != nbt.LongArray {
			t.Fatal("Jigsaw references not saved")
		}
	}
	if blockTypes["minecraft:brushable_block"] == 0 || blockTypes["minecraft:trial_spawner"] == 0 {
		t.Fatalf("missing archaeology/trial block entities:%v", blockTypes)
	}
	reopened, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for _, position := range positions {
		if _, err := reopened.LoadChunk(position[0], position[1]); err != nil {
			t.Fatal(err)
		}
		chunk, err := reopened.store.vanilla.loadBlockChunk(position[0], position[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := chunk.anvilRoot.Object()
		if !reflect.DeepEqual(root["structures"], saved[position]["structures"]) || !reflect.DeepEqual(root["block_entities"], saved[position]["block_entities"]) {
			t.Fatalf("restart changed Jigsaw structures/block entities at%v", position)
		}
	}
	id, err := reopened.BlockAt(44*16+8, 280, 38*16+8)
	if err != nil || id != StateStone {
		t.Fatal("restart regenerated edited Jigsaw chunk", err)
	}
	t.Logf("saved/reopened six actual Jigsaw chunks; block entities:%v", blockTypes)
}
