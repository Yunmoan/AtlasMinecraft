package world

import (
	"atlas/internal/anvil"
	"atlas/internal/item"
	"atlas/internal/nbt"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func savedChestWorld(t *testing.T, name, table string) (*World, string, BlockPosition) {
	t.Helper()
	dir := t.TempDir()
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	position := BlockPosition{X: -1, Y: -60, Z: 1}
	ch := newChunkBounds(-1, 0, -64, 384, true)
	state, err := BlockStateID(name, map[string]string{"type": "single", "facing": "north", "waterlogged": "false"})
	if err != nil {
		t.Fatal(err)
	}
	ch.SetBlock(15, 1, -60, state)
	if err := ch.initializeAnvilRoot(); err != nil {
		t.Fatal(err)
	}
	data := nbt.Object{"id": {Type: nbt.String, Value: name}, "x": {Type: nbt.Int, Value: position.X}, "y": {Type: nbt.Int, Value: position.Y}, "z": {Type: nbt.Int, Value: position.Z}, "ForeignContainerField": {Type: nbt.String, Value: "preserved"}, "CustomName": {Type: nbt.String, Value: `{"text":"Original chest","color":"gold"}`}}
	if table != "" {
		data["LootTable"] = nbt.Tag{Type: nbt.String, Value: table}
		data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(1234)}
	}
	root, _ := ch.anvilRoot.Object()
	root["block_entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: data}}}
	chunkName, tag, err := ch.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if err := anvil.WriteChunk(filepath.Join(dir, "region"), -1, 0, chunkName, tag); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	return w, dir, position
}
func chestSavedData(t *testing.T, dir string, p BlockPosition) nbt.Object {
	t.Helper()
	_, tag, found, err := anvil.ReadChunk(filepath.Join(dir, "region"), p.X>>4, p.Z>>4)
	if err != nil || !found {
		t.Fatal("missing chest chunk", err)
	}
	root, _ := tag.Object()
	entities, _ := root["block_entities"].Elements()
	for _, tag := range entities {
		data, _ := tag.Object()
		x, _ := data.Number("x")
		y, _ := data.Number("y")
		z, _ := data.Number("z")
		if x == int64(p.X) && y == int64(p.Y) && z == int64(p.Z) {
			return data
		}
	}
	t.Fatal("missing chest block entity")
	return nil
}
func TestChestLootUnpackedOnceAndSavedBeforeSnapshot(t *testing.T) {
	table := "minecraft:chests/pillager_outpost"
	w, dir, p := savedChestWorld(t, "minecraft:chest", table)
	want, err := FillChestLoot(table, 1234, 0)
	if err != nil {
		t.Fatal(err)
	}
	before := chestSavedData(t, dir, p)
	if _, ok := before["LootTable"]; !ok {
		t.Fatal("loot eagerly unpacked")
	}
	got, err := w.ReadChest(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	for i, stack := range want {
		if got.Slots[i].Count != stack.Count || !stack.Empty() && !got.Slots[i].SameItem(stack) {
			t.Fatal("different loot", i)
		}
	}
	data := chestSavedData(t, dir, p)
	if _, exists := data["LootTable"]; exists {
		t.Fatal("loot table survived first snapshot")
	}
	if _, exists := data["LootTableSeed"]; exists {
		t.Fatal("loot seed not consumed")
	}
	if foreign, _ := data.Text("ForeignContainerField"); foreign != "preserved" {
		t.Fatal("unknown chest NBT lost")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	next, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer next.Close()
	after, err := next.ReadChest(p, 5)
	if err != nil {
		t.Fatal(err)
	}
	if got.Position != after.Position || got.CustomName != after.CustomName {
		t.Fatal("restart lost chest identity/name")
	}
	for i, stack := range got.Slots {
		if stack.Count != after.Slots[i].Count || !stack.Empty() && !stack.SameItem(after.Slots[i]) {
			t.Fatal("restart/luck rerolled or lost saved chest", i)
		}
	}
}
func TestChestConcurrentUpdatesAndSaveFailureRollback(t *testing.T) {
	w, dir, p := savedChestWorld(t, "minecraft:trapped_chest", "")
	stack, err := item.FromNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:stone"}, "count": {Type: nbt.Int, Value: int32(7)}, "ForeignItemField": {Type: nbt.String, Value: "kept"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.UpdateChest(p, 0, func(v *ChestInventory) error { v.Slots[0] = stack; return nil }); err != nil {
		t.Fatal(err)
	}
	if name, _ := chestSavedData(t, dir, p).Text("id"); name != "minecraft:trapped_chest" {
		t.Fatal("trapped chest saved as normal chest")
	}
	var wg sync.WaitGroup
	results := make(chan item.Stack, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var taken item.Stack
			_, err := w.UpdateChest(p, 0, func(v *ChestInventory) error {
				taken = v.Slots[0].WithCount(min(1, v.Slots[0].Count))
				v.Slots[0] = v.Slots[0].WithCount(v.Slots[0].Count - taken.Count)
				return nil
			})
			if err != nil {
				t.Error(err)
			}
			results <- taken
		}()
	}
	wg.Wait()
	close(results)
	count := int32(0)
	for stack := range results {
		count += stack.Count
		if !stack.Empty() {
			tag, err := stack.ToNBT()
			if err != nil {
				t.Fatal(err)
			}
			o, _ := tag.Object()
			if field, _ := o.Text("ForeignItemField"); field != "kept" {
				t.Fatal("opaque item data lost")
			}
		}
	}
	if count != 7 {
		t.Fatal("concurrent chest duplicated or lost items", count)
	}
	if _, err := w.UpdateChest(p, 0, func(v *ChestInventory) error { v.Slots[0] = stack; return nil }); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "region", "r.-1.0.mca")
	previous, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := w.UpdateChest(p, 0, func(v *ChestInventory) error { v.Slots[0] = item.Stack{}; return nil }); err == nil {
		t.Fatal("corrupt source accepted chest save")
	}
	after, err := w.ReadChest(p, 0)
	if err != nil || after.Slots[0].Count != 7 {
		t.Fatal("failed commit mutated live chest", err)
	}
	if err := os.WriteFile(path, previous, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := w.UpdateChest(p, 0, func(v *ChestInventory) error { v.Slots[0] = item.Stack{ID: 999999, Count: 1}; return nil }); err == nil {
		t.Fatal("invalid server ID accepted")
	}
	if _, err := w.UpdateChest(p, 0, func(v *ChestInventory) error { v.Slots[0] = item.Stack{}; return errors.New("abort") }); err == nil {
		t.Fatal("callback abort accepted")
	}
	after, err = w.ReadChest(p, 0)
	if err != nil || after.Slots[0].Count != 7 {
		t.Fatal("invalid/aborted update mutated chest", err)
	}
}
func TestChestRedstoneConductorObstruction(t *testing.T) {
	w, _, p := savedChestWorld(t, "minecraft:chest", "")
	if err := w.SetBlock(p.X, p.Y+1, p.Z, StateStone); err != nil {
		t.Fatal(err)
	}
	if _, err := w.ReadChest(p, 0); !errors.Is(err, ErrChestBlocked) {
		t.Fatal("solid above chest did not block", err)
	}
	glass, err := BlockStateID("minecraft:glass", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(p.X, p.Y+1, p.Z, glass); err != nil {
		t.Fatal(err)
	}
	if _, err := w.ReadChest(p, 0); err != nil {
		t.Fatal("transparent conductor=false above chest blocked open", err)
	}
}
