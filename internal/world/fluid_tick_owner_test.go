package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func diskFluidTicks(t *testing.T, dir string, x, z int32) []tickqueue.Saved[worldgen.FluidType] {
	t.Helper()
	_, tag, ok, err := anvil.ReadChunk(filepath.Join(dir, "region"), x, z)
	if err != nil || !ok {
		t.Fatal("read saved fluid queue", ok, err)
	}
	root, _ := tag.Object()
	return tickqueue.DecodeSavedNBT(root["fluid_ticks"], x, z, parseScheduledFluid)
}

func diskBlockTicks(t *testing.T, dir string, x, z int32) []tickqueue.Saved[string] {
	t.Helper()
	_, tag, ok, err := anvil.ReadChunk(filepath.Join(dir, "region"), x, z)
	if err != nil || !ok {
		t.Fatal("read saved block queue", ok, err)
	}
	root, _ := tag.Object()
	return tickqueue.DecodeSavedNBT(root["block_ticks"], x, z, parseScheduledBlock)
}

func TestFluidQueueActivationSaveUnloadAndRestart(t *testing.T) {
	dir := t.TempDir()
	seed := int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	c, err := v.loadChunk(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := c.anvilRoot.Object()
	request := tickqueue.Saved[worldgen.FluidType]{Type: worldgen.FluidWater, Position: tickqueue.Position{X: 3, Y: 64, Z: 7}, Delay: 30}
	root["fluid_ticks"] = tickqueue.EncodeSavedNBT([]tickqueue.Saved[worldgen.FluidType]{request}, func(k worldgen.FluidType) string { return k.String() })
	blockRequest := tickqueue.Saved[string]{Type: "minecraft:oak_leaves", Position: request.Position, Delay: 30, Priority: -2}
	root["block_ticks"] = tickqueue.EncodeSavedNBT([]tickqueue.Saved[string]{blockRequest}, func(name string) string { return name })
	if err := v.persistChunk(c); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	load := func(x, z int32) {
		t.Helper()
		if _, err := w.LoadChunk(x, z); err != nil {
			t.Fatal(err)
		}
	}
	eligible := func(x, z int32) bool { return x == 0 && z == 0 }
	activate := func() {
		t.Helper()
		w.blockUpdateMu.Lock()
		defer w.blockUpdateMu.Unlock()
		if err := w.scheduledTicks.activate(eligible, func(worldgen.FeaturePosition) error {
			t.Fatal("flat chunk unexpectedly needs post-processing")
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	load(0, 0)
	w.AdvanceTime(50)
	activate()
	if w.scheduledTicks.chunks[tickqueue.ChunkKey(0, 0)].active {
		t.Fatal("activated without full neighbor ring")
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	if got := diskFluidTicks(t, dir, 0, 0); !reflect.DeepEqual(got, []tickqueue.Saved[worldgen.FluidType]{request}) {
		t.Fatalf("loaded but inactive delay changed: %+v", got)
	}
	if got := diskBlockTicks(t, dir, 0, 0); !reflect.DeepEqual(got, []tickqueue.Saved[string]{blockRequest}) {
		t.Fatal("inactive block delay changed", got)
	}
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			load(x, z)
		}
	}
	activate()
	w.AdvanceTime(10)
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	request.Delay = 20
	blockRequest.Delay = 20
	if got := diskFluidTicks(t, dir, 0, 0); !reflect.DeepEqual(got, []tickqueue.Saved[worldgen.FluidType]{request}) {
		t.Fatalf("active delay not saved relative to time: %+v", got)
	}
	if got := diskBlockTicks(t, dir, 0, 0); !reflect.DeepEqual(got, []tickqueue.Saved[string]{blockRequest}) {
		t.Fatal("active block delay changed", got)
	}
	if count, err := w.UnloadIdle(0); err != nil || count != 9 {
		t.Fatal("unload", count, err)
	}
	if len(w.scheduledTicks.chunks) != 0 || len(w.scheduledTicks.loadOrder) != 0 {
		t.Fatal("unloaded scheduling handles retained")
	}
	w.AdvanceTime(100)
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			load(x, z)
		}
	}
	activate()
	w.AdvanceTime(19)
	tick := func() int {
		t.Helper()
		w.blockUpdateMu.Lock()
		defer w.blockUpdateMu.Unlock()
		n := 0
		blocks := 0
		if err := w.scheduledTicks.tickBlocks(65536, eligible, func(tickqueue.Scheduled[string]) error { blocks++; return nil }); err != nil {
			t.Fatal(err)
		}
		err := w.scheduledTicks.tick(65536, eligible, func(tickqueue.Scheduled[worldgen.FluidType]) error { n++; return nil })
		if err != nil {
			t.Fatal(err)
		}
		if n != blocks {
			t.Fatal("block and fluid activation/unload lifetime diverged", n, blocks)
		}
		return n
	}
	if tick() != 0 {
		t.Fatal("unloaded time consumed saved delay")
	}
	w.AdvanceTime(1)
	if tick() != 1 {
		t.Fatal("restored request missed its due tick")
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	if got := diskFluidTicks(t, dir, 0, 0); len(got) != 0 {
		t.Fatalf("consumed request retained on disk: %+v", got)
	}
	if got := diskBlockTicks(t, dir, 0, 0); len(got) != 0 {
		t.Fatal("consumed block request retained", got)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if _, err := reopened.LoadChunk(0, 0); err != nil {
		t.Fatal(err)
	}
	if reopened.scheduledTicks.level.Len() != 0 {
		t.Fatal("consumed request reappeared on restart")
	}
	if reopened.scheduledTicks.blocks.Len() != 0 {
		t.Fatal("consumed block request reappeared on restart")
	}
}

func TestScheduledQueuesSharedOrderCheckpointAndFailure(t *testing.T) {
	w, dir := mechanicsWorld(t)
	eligible := func(x, z int32) bool { return x == 0 && z == 0 }
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	o := w.scheduledTicks
	if err := o.activate(eligible, func(worldgen.FeaturePosition) error { return nil }); err != nil {
		t.Fatal(err)
	}
	block := func(x int32) {
		t.Helper()
		if err := o.scheduleBlock(worldgen.FeaturePosition{X: x, Y: 64, Z: 0}, "minecraft:oak_leaves", 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	fluid := func(x int32) {
		t.Helper()
		if err := o.schedule(worldgen.FeaturePosition{X: x, Y: 64, Z: 0}, worldgen.FluidWater, 0); err != nil {
			t.Fatal(err)
		}
	}
	block(0)
	fluid(0)
	block(0)
	block(1) // Duplicate still consumes the world's order 2.
	failure := errors.New("injected block callback failure")
	calls := 0
	err := o.tickBlocks(65536, eligible, func(s tickqueue.Scheduled[string]) error {
		calls++
		if s.Order != 0 {
			t.Fatal("wrong initial block order", s.Order)
		}
		block(2)
		fluid(2)
		if _, err := w.editBlock(setBlockMsg{x: 8, y: 64, z: 8, state: StateStone}); err != nil {
			return err
		}
		if got := diskBlockTicks(t, dir, 0, 0); len(got) != 3 {
			t.Fatal("block callback checkpoint lost pending requests", got)
		}
		if got := diskFluidTicks(t, dir, 0, 0); len(got) != 2 {
			t.Fatal("block callback checkpoint lost fluid requests", got)
		}
		return failure
	})
	if !errors.Is(err, failure) || calls != 1 {
		t.Fatal("block failure did not stop batch", err, calls)
	}
	var blockOrders, fluidOrders []int64
	if err := o.tickBlocks(65536, eligible, func(s tickqueue.Scheduled[string]) error { blockOrders = append(blockOrders, s.Order); return nil }); err != nil {
		t.Fatal(err)
	}
	if err := o.tick(65536, eligible, func(s tickqueue.Scheduled[worldgen.FluidType]) error {
		fluidOrders = append(fluidOrders, s.Order)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(blockOrders, []int64{0, 3, 4}) || !reflect.DeepEqual(fluidOrders, []int64{1, 5}) {
		t.Fatal("shared order/retry mismatch", blockOrders, fluidOrders)
	}
	result := make(chan error, 1)
	w.regions[chunkKey{0, 0}].mailbox <- saveRegion{result: result}
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("block queue actor save deadlocked")
	}
	if len(diskBlockTicks(t, dir, 0, 0)) != 0 || len(diskFluidTicks(t, dir, 0, 0)) != 0 {
		t.Fatal("completed queues retained on disk")
	}
}

func TestFluidQueueCallbackCheckpointAndFailureRetainPending(t *testing.T) {
	w, dir := mechanicsWorld(t)
	eligible := func(x, z int32) bool { return x == 0 && z == 0 }
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	if err := w.scheduledTicks.activate(eligible, func(worldgen.FeaturePosition) error { return nil }); err != nil {
		t.Fatal(err)
	}
	for i := int32(0); i < 3; i++ {
		if err := w.scheduledTicks.schedule(worldgen.FeaturePosition{X: i, Y: 64, Z: 0}, worldgen.FluidWater, 0); err != nil {
			t.Fatal(err)
		}
	}
	n := 0
	failure := errors.New("injected tick callback failure")
	err := w.scheduledTicks.tick(65536, eligible, func(tick tickqueue.Scheduled[worldgen.FluidType]) error {
		n++
		if _, err := w.editBlock(setBlockMsg{x: 8, y: 64, z: 8, state: StateStone}); err != nil {
			return err
		}
		// The actor saves while the scheduler's outer tick is in progress.
		// This must neither deadlock nor lose the remaining collected ticks.
		got := diskFluidTicks(t, dir, 0, 0)
		if len(got) != 3 {
			t.Fatalf("callback checkpoint lost collected work: %+v", got)
		}
		return failure
	})
	if !errors.Is(err, failure) || n != 1 {
		t.Fatalf("failed callback did not stop pass: n=%d err=%v", n, err)
	}
	if w.scheduledTicks.level.Len() != 3 {
		t.Fatal("failure discarded queued work")
	}
	var positions []int32
	if err := w.scheduledTicks.tick(65536, eligible, func(tick tickqueue.Scheduled[worldgen.FluidType]) error {
		positions = append(positions, tick.Position.X)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(positions, []int32{0, 1, 2}) {
		t.Fatalf("retry changed order: %v", positions)
	}
	// Save directly through an actor while holding the edit lock. This mirrors
	// the end-of-pass flush without recursively invoking World.Save's lock.
	result := make(chan error, 1)
	w.regions[chunkKey{0, 0}].mailbox <- saveRegion{result: result}
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("actor save deadlocked")
	}
	if got := diskFluidTicks(t, dir, 0, 0); len(got) != 0 {
		t.Fatalf("retry completion not persisted: %+v", got)
	}
	// The queue codec continues to produce a legal empty native NBT list.
	_, tag, _, err := anvil.ReadChunk(filepath.Join(dir, "region"), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := tag.Object()
	if root["fluid_ticks"].ListType != nbt.End {
		t.Fatal("non-native empty list marker")
	}
}

func TestFluidQueueCheckpointAcrossRegionBoundary(t *testing.T) {
	dir := t.TempDir()
	seed := int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	for x := int32(30); x <= 33; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	eligible := func(x, z int32) bool { return (x == 31 || x == 32) && z == 0 }
	w.blockUpdateMu.Lock()
	err = w.scheduledTicks.activate(eligible, func(worldgen.FeaturePosition) error { return nil })
	if err != nil {
		w.blockUpdateMu.Unlock()
		t.Fatal(err)
	}
	for _, x := range []int32{511, 512} {
		if err := w.scheduledTicks.schedule(worldgen.FeaturePosition{X: x, Y: 64, Z: 8}, worldgen.FluidWater, 0); err != nil {
			w.blockUpdateMu.Unlock()
			t.Fatal(err)
		}
	}
	var ran []int32
	err = w.scheduledTicks.tick(65536, eligible, func(tick tickqueue.Scheduled[worldgen.FluidType]) error {
		ran = append(ran, tick.Position.X)
		if tick.Position.X == 511 {
			if _, err := w.editBlock(setBlockMsg{x: 513, y: 64, z: 8, state: StateStone}); err != nil {
				return err
			}
			got := diskFluidTicks(t, dir, 32, 0)
			if len(got) != 1 || got[0].Position.X != 512 {
				t.Errorf("other region lost collected update: %+v", got)
			}
		}
		return nil
	})
	w.blockUpdateMu.Unlock()
	if err != nil || !reflect.DeepEqual(ran, []int32{511, 512}) {
		t.Fatalf("cross-region pass got %v, %v", ran, err)
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	for _, x := range []int32{31, 32} {
		if got := diskFluidTicks(t, dir, x, 0); len(got) != 0 {
			t.Fatalf("chunk %d did not persist completion: %+v", x, got)
		}
	}
}
