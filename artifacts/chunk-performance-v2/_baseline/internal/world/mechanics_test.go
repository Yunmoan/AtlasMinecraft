package world

import (
	"errors"
	"testing"
)

func mechanicsWorld(t *testing.T) (*World, string) {
	t.Helper()
	dir := t.TempDir()
	seed := int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	})
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	return w, dir
}

func TestGravityAndDoublePlantUpdatesPersist(t *testing.T) {
	w, dir := mechanicsWorld(t)
	id := func(name string, properties map[string]string) uint32 {
		state, err := BlockStateID(name, properties)
		if err != nil {
			t.Fatal(err)
		}
		return state
	}
	set := func(x, y, z int32, state uint32) {
		if err := w.SetBlock(x, y, z, state); err != nil {
			t.Fatal(err)
		}
	}
	check := func(x, y, z int32, state uint32) {
		got, err := w.BlockAt(x, y, z)
		if err != nil || got != state {
			t.Fatalf("%d %d %d: state=%d want=%d error=%v", x, y, z, got, state, err)
		}
	}
	sand, gravel := id("sand", nil), id("gravel", nil)
	set(-1, -50, -1, StateStone)
	set(-1, -49, -1, sand)
	set(-1, -48, -1, gravel)
	set(-1, -50, -1, StateAir)
	check(-1, -60, -1, sand)
	check(-1, -59, -1, gravel)
	check(-1, -49, -1, StateAir)
	check(-1, -48, -1, StateAir)
	lower, upper := id("tall_grass", map[string]string{"half": "lower"}), id("tall_grass", map[string]string{"half": "upper"})
	if err := w.TryPlaceBlock(BlockPosition{-2, -60, -1}, StateAir, lower, &upper); err != nil {
		t.Fatal(err)
	}
	check(-2, -60, -1, lower)
	check(-2, -59, -1, upper)
	set(-2, -59, -1, StateAir)
	check(-2, -60, -1, StateAir)
	if err := w.TryPlaceBlock(BlockPosition{-2, -60, -1}, StateAir, lower, &upper); err != nil {
		t.Fatal(err)
	}
	set(-2, -61, -1, StateAir)
	check(-2, -60, -1, StateAir)
	check(-2, -59, -1, StateAir)
	if err := w.TryPlaceBlock(BlockPosition{-4, -60, -1}, StateAir, lower, &upper); err != nil {
		t.Fatal(err)
	}
	set(-4, -45, -1, sand)
	check(-4, -60, -1, sand)
	check(-4, -59, -1, StateAir)
	set(-3, -59, -1, StateStone)
	if err := w.TryPlaceBlock(BlockPosition{-3, -60, -1}, StateAir, lower, &upper); !errors.Is(err, ErrBlockChanged) {
		t.Fatalf("blocked upper half: %v", err)
	}
	check(-3, -60, -1, StateAir)
	if err := w.TryPlaceBlock(BlockPosition{-1, -60, -1}, StateAir, StateDirt, nil); !errors.Is(err, ErrBlockChanged) {
		t.Fatalf("stale expected state: %v", err)
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for _, point := range []struct {
		p     BlockPosition
		state uint32
	}{{BlockPosition{-1, -60, -1}, sand}, {BlockPosition{-1, -59, -1}, gravel}, {BlockPosition{-1, -49, -1}, StateAir}, {BlockPosition{-2, -60, -1}, StateAir}, {BlockPosition{-2, -59, -1}, StateAir}, {BlockPosition{-3, -59, -1}, StateStone}} {
		got, err := reopened.BlockAt(point.p.X, point.p.Y, point.p.Z)
		if err != nil || got != point.state {
			t.Fatalf("update lost on restart at %+v: %d %v", point.p, got, err)
		}
	}
}

func TestPlayerCollisionShapesAndSweptMovement(t *testing.T) {
	w, _ := mechanicsWorld(t)
	if err := w.SetBlock(0, -60, 1, StateStone); err != nil {
		t.Fatal(err)
	}
	from, to := [3]float64{.5, -60, .5}, [3]float64{.5, -60, 2.5}
	if allowed, err := w.CanMovePlayer(from, to, 1.8, true); err != nil || allowed {
		t.Fatalf("crossed a wall: %t %v", allowed, err)
	}
	slab, err := BlockStateID("oak_slab", map[string]string{"type": "bottom"})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(0, -60, 1, slab); err != nil {
		t.Fatal(err)
	}
	if allowed, err := w.CanMovePlayer(from, [3]float64{.5, -59.5, 1.5}, 1.8, true); err != nil || !allowed {
		t.Fatalf("step onto half slab rejected: %t %v", allowed, err)
	}
	if allowed, err := w.CanMovePlayer(from, [3]float64{.5, -60, 1.5}, 1.8, true); err != nil || allowed {
		t.Fatalf("walk through half slab: %t %v", allowed, err)
	}
	top, err := BlockStateID("oak_slab", map[string]string{"type": "top"})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(1, -59, 0, top); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		height  float64
		allowed bool
	}{{1.8, false}, {1.5, true}} {
		allowed, err := w.CanMovePlayer(from, [3]float64{1.5, -60, .5}, test.height, true)
		if err != nil || allowed != test.allowed {
			t.Fatalf("ceiling height %f: %t %v", test.height, allowed, err)
		}
	}
	if _, err := w.ReadLoadedBlocks([]BlockPosition{{1000, 0, 1000}}); !errors.Is(err, ErrChunkNotLoaded) {
		t.Fatalf("missing chunk lookup: %v", err)
	}
}
