package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"testing"
)

func TestBrushableProgressLifetimeAndScheduledReset(t *testing.T) {
	w, dir := mechanicsWorld(t)
	p := BlockPosition{X: 8, Y: 64, Z: 8}
	state, _ := BlockStateID("minecraft:suspicious_gravel", map[string]string{"dusted": "3"})
	if err := w.SetBlock(p.X, p.Y-1, p.Z, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(p.X, p.Y, p.Z, state); err != nil {
		t.Fatal(err)
	}
	age, _ := w.Time()
	// Supply a loaded archaeology entity; the nonzero count models an already
	// active brushing session. None of these transient fields belong in NBT.
	if err := w.fallingEntityOperation(0, 0, func(_ *Region, ch *Chunk) error {
		root, _ := ch.anvilRoot.Object()
		root["block_entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: nbt.Object{
			"id": {Type: nbt.String, Value: "minecraft:brushable_block"}, "x": {Type: nbt.Int, Value: p.X}, "y": {Type: nbt.Int, Value: p.Y}, "z": {Type: nbt.Int, Value: p.Z},
			"LootTable": {Type: nbt.String, Value: "minecraft:archaeology/ocean_ruin_cold"}, "LootTableSeed": {Type: nbt.Long, Value: int64(42)}, "keepPacked": {Type: nbt.Byte, Value: int8(0)},
		}}}}
		ch.brushableProgress = map[BlockPosition]worldgen.BrushableProgress{p: {Count: 8, ResetAt: age + 2, CooldownEnd: age + 10, HitDirection: 2}}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	eligible := func(x, z int32) bool { return x == 0 && z == 0 }
	for tick := 1; tick <= 6; tick++ {
		w.AdvanceTime(1)
		if err := w.TickScheduled(eligible); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.fallingEntityOperation(0, 0, func(_ *Region, ch *Chunk) error {
		progress := ch.brushableProgress[p]
		if progress.Count != 4 || progress.ResetAt != age+10 || progress.CooldownEnd != age+10 || progress.HitDirection != 2 {
			t.Errorf("incorrect live progress: %+v", progress)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	id, err := w.BlockAt(p.X, p.Y, p.Z)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := BlockStateID("minecraft:suspicious_gravel", map[string]string{"dusted": "2"})
	if id != want {
		t.Fatalf("dusted state %d want %d", id, want)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w, err = Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	for tick := 0; tick < 8; tick++ {
		w.AdvanceTime(1)
		if err := w.TickScheduled(eligible); err != nil {
			t.Fatal(err)
		}
	}
	id, err = w.BlockAt(p.X, p.Y, p.Z)
	if err != nil || id != want {
		t.Fatalf("reload reconstructed brushing from DUSTED: %d/%d %v", id, want, err)
	}
	if err := w.fallingEntityOperation(0, 0, func(_ *Region, ch *Chunk) error {
		if len(ch.brushableProgress) != 0 {
			t.Error("transient progress survived reload")
		}
		ch.brushableProgress = map[BlockPosition]worldgen.BrushableProgress{p: {Count: 4}}
		ch.SetBlock(8, 8, 64, StateStone)
		if len(ch.brushableProgress) != 0 {
			t.Error("replaced brushable retained live progress")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
