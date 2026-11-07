package world

import (
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"errors"
	"testing"
)

func TestScheduledBlockFailureRetainsFluidPhase(t *testing.T) {
	w, _ := mechanicsWorld(t)
	p := worldgen.FeaturePosition{X: 8, Y: 64, Z: 8}
	farmland, err := BlockStateID("minecraft:farmland", nil)
	if err != nil {
		t.Fatal(err)
	}
	w.blockUpdateMu.Lock()
	_, err = w.editBlock(setBlockMsg{x: p.X, y: p.Y, z: p.Z, state: farmland, raw: true, flags: 2})
	if err == nil {
		err = w.scheduledTicks.scheduleBlock(p, "minecraft:farmland", 0, 0)
	}
	if err == nil {
		err = w.scheduledTicks.schedule(worldgen.FeaturePosition{X: 9, Y: 64, Z: 8}, worldgen.FluidWater, 0)
	}
	w.blockUpdateMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	eligible := func(x, z int32) bool { return x == 0 && z == 0 }
	if err := w.TickScheduled(eligible); !errors.Is(err, worldgen.ErrUnsupportedScheduledBlock) {
		t.Fatal("unsupported block tick did not fail explicitly", err)
	}
	o := w.scheduledTicks
	o.mu.Lock()
	blocks, fluids := o.blocks.Len(), o.level.Len()
	o.mu.Unlock()
	if blocks != 1 || fluids != 1 {
		t.Fatal("failure consumed block/fluid work", blocks, fluids)
	}
	// Replacing the unsupported block makes its request stale; the next pass
	// then drains both stale entries without invoking a different block's tick.
	w.blockUpdateMu.Lock()
	_, err = w.editBlock(setBlockMsg{x: p.X, y: p.Y, z: p.Z, state: StateStone, raw: true, flags: 2})
	w.blockUpdateMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err := w.TickScheduled(eligible); err != nil {
		t.Fatal(err)
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.blocks.Len() != 0 || o.level.Len() != 0 || o.chunks[tickqueue.ChunkKey(0, 0)].blocks.Len() != 0 {
		t.Fatal("stale scheduled entries survived retry")
	}
}
