package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalWorldCreateGenerateEditReload(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "normal-world")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	if w.IsFlat() || w.Seed() != seed {
		t.Fatal("normal world metadata changed")
	}
	if err := w.PreGenerateSpawn(0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := w.LoadChunk(-1, 0); err != nil {
		t.Fatal(err)
	}
	_, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), -1, 0)
	if err != nil || !exists {
		t.Fatalf("negative-coordinate Anvil chunk: %v", err)
	}
	object, _ := root.Object()
	coverage, ok := object.Object("AtlasWorldgen")
	if !ok {
		t.Fatal("generated chunk has no coverage marker")
	}
	if complete, _ := coverage.Number("Complete"); complete != 0 {
		t.Fatal("incomplete decoration was advertised as exact")
	}
	sections, _ := object["sections"].Elements()
	foundOre := false
	for _, section := range sections {
		o, _ := section.Object()
		states, _ := o.Object("block_states")
		palette, _ := states["palette"].Elements()
		for _, entry := range palette {
			state, _ := entry.Object()
			name, _ := state.Text("Name")
			if strings.Contains(name, "_ore") {
				foundOre = true
			}
		}
	}
	if !foundOre {
		t.Fatal("normal generation produced no ores")
	}
	postProcessing, _ := object["PostProcessing"].Elements()
	if len(postProcessing) != NumSections {
		t.Fatal("fluid post-processing sections missing")
	}
	// The halo is scratch terrain, not a partially committed neighbor save.
	if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), -2, 0); err != nil || exists {
		t.Fatalf("scratch neighbor was saved: %v", err)
	}
	y, err := w.SurfaceY(-1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(-1, y, 0, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.PreGenerateSpawn(-1, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := w.LoadChunk(-2, 0); err != nil {
		t.Fatal(err)
	}
	before, err := w.LoadChunk(-1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	after, err := reopened.LoadChunk(-1, 0)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("generated/edited packet changed on restart: %v", err)
	}
	name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), -1, 0)
	if err != nil || !exists {
		t.Fatalf("saved chunk missing: %v", err)
	}
	chunk, err := chunkFromAnvil(name, root, -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if chunk.GetBlock(15, 0, y) != StateStone {
		t.Fatal("neighbor generation or restart overwrote a block edit")
	}
}

func TestNormalWorldRefusesUnfinishedExistingChunk(t *testing.T) {
	dir := t.TempDir()
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	chunk := NewChunk(0, 0)
	if err := chunk.initializeAnvilRoot(); err != nil {
		t.Fatal(err)
	}
	root, _ := chunk.anvilRoot.Object()
	root["Status"] = nbt.Tag{Type: nbt.String, Value: "minecraft:carvers"}
	if err := anvil.WriteChunk(filepath.Join(dir, "region"), 0, 0, "", chunk.anvilRoot); err != nil {
		t.Fatal(err)
	}
	save, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := save.loadChunk(0, 0); err == nil {
		t.Fatal("unfinished vanilla chunk was treated as full or overwritten")
	}
}

func treeStatesInChunk(t *testing.T, chunk *Chunk) map[[3]int32]uint32 {
	t.Helper()
	states := make(map[[3]int32]uint32)
	for y := int32(MinY); y < MaxY; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				id := chunk.GetBlock(x, z, y)
				state, err := BlockStateByID(id)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(state.Name, "_log") || strings.HasSuffix(state.Name, "_leaves") {
					states[[3]int32{int32(x), y, int32(z)}] = id
				}
			}
		}
	}
	if len(states) < 20 {
		t.Fatalf("generated chunk contains only %d tree blocks", len(states))
	}
	return states
}

func vegetationStatesInChunk(t *testing.T, chunk *Chunk) map[[3]int32]uint32 {
	t.Helper()
	states := make(map[[3]int32]uint32)
	for y := int32(MinY); y < MaxY; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				id := chunk.GetBlock(x, z, y)
				state, err := BlockStateByID(id)
				if err != nil {
					t.Fatal(err)
				}
				if worldgen.SupportedPlant(state.Name) {
					states[[3]int32{int32(x), y, int32(z)}] = id
				}
			}
		}
	}
	if len(states) == 0 {
		t.Fatal("generated chunk has no vegetation")
	}
	return states
}

func TestNormalWorldTreesSaveEditNeighborAndReload(t *testing.T) {
	dir := t.TempDir()
	seed := int64(0)
	const cx, cz = int32(-96), int32(-128)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	save, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := save.loadChunk(cx, cz)
	if err != nil {
		t.Fatal(err)
	}
	trees := treeStatesInChunk(t, chunk)
	plants := vegetationStatesInChunk(t, chunk)
	var marker [3]int32
	for position := range trees {
		marker = position
		break
	}
	chunk.SetBlock(int(marker[0]), int(marker[2]), marker[1], StateStone)
	trees[marker] = StateStone
	for position, id := range plants {
		trees[position] = id
	}
	if err := save.persistChunk(chunk); err != nil {
		t.Fatal(err)
	}
	if _, err := save.loadChunk(cx+1, cz); err != nil {
		t.Fatal(err)
	}
	reopened, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := reopened.loadChunk(cx, cz)
	if err != nil {
		t.Fatal(err)
	}
	for position, id := range trees {
		if loaded.GetBlock(int(position[0]), int(position[2]), position[1]) != id {
			t.Fatalf("tree/edit changed after neighbor generation and reload at %+v", position)
		}
	}
	t.Logf("preserved %d tree/vegetation blocks (%d vegetation) and one edit in normal chunk %d,%d", len(trees), len(plants), cx, cz)
}
