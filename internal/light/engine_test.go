package light

import (
	"atlas/internal/blockstate"
	"testing"
)

func lightState(t *testing.T, name string, properties map[string]string) uint16 {
	t.Helper()
	id, err := blockstate.BlockStateID(name, properties)
	if err != nil {
		t.Fatal(err)
	}
	return uint16(id)
}

func level(data *Data, sky bool, x, y, z, minY int) byte {
	layers := data.Block
	if sky {
		layers = data.Sky
	}
	layer := layers[(y-minY+16)>>4]
	if layer == nil {
		return 0
	}
	i := ((y-minY)&15)*256 + z*16 + x
	return layer[i/2] >> uint((i&1)*4) & 15
}

func TestSkyRoofOpeningAndClosing(t *testing.T) {
	e, err := New(0, 16, true)
	if err != nil {
		t.Fatal(err)
	}
	states := make([]uint16, 16*256)
	stone := lightState(t, "stone", nil)
	for i := 8 * 256; i < 9*256; i++ {
		states[i] = stone
	}
	p := Position{0, 0}
	if _, err := e.Install(p, states, nil); err != nil {
		t.Fatal(err)
	}
	if got := level(e.Snapshot(p), true, 8, 7, 8, 0); got != 0 {
		t.Fatalf("closed roof sky=%d", got)
	}
	if got := level(e.Snapshot(p), true, 8, 9, 8, 0); got != 15 {
		t.Fatalf("open sky=%d", got)
	}
	if _, err := e.Edit(p, []Change{{8, 8, 8, 0}}); err != nil {
		t.Fatal(err)
	}
	if got := level(e.Snapshot(p), true, 8, 0, 8, 0); got != 15 {
		t.Fatalf("direct sun through opening=%d", got)
	}
	if got := level(e.Snapshot(p), true, 9, 0, 8, 0); got != 14 {
		t.Fatalf("sideways sky=%d", got)
	}
	if _, err := e.Edit(p, []Change{{8, 8, 8, stone}}); err != nil {
		t.Fatal(err)
	}
	if got := level(e.Snapshot(p), true, 8, 0, 8, 0); got != 0 {
		t.Fatalf("closed opening retained sky=%d", got)
	}
}

func TestBlockLightCrossesChunkBorderAndRemovalClearsIt(t *testing.T) {
	e, err := New(0, 16, true)
	if err != nil {
		t.Fatal(err)
	}
	p, neighbor := Position{0, 0}, Position{1, 0}
	states := make([]uint16, 16*256)
	states[4*256+8*16+15] = lightState(t, "torch", nil)
	if _, err := e.Install(p, states, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Install(neighbor, make([]uint16, len(states)), nil); err != nil {
		t.Fatal(err)
	}
	if got := level(e.Snapshot(p), false, 15, 4, 8, 0); got != 14 {
		t.Fatalf("torch source=%d", got)
	}
	if got := level(e.Snapshot(neighbor), false, 0, 4, 8, 0); got != 13 {
		t.Fatalf("border light=%d", got)
	}
	before := e.Snapshot(neighbor)
	if _, err := e.Edit(p, []Change{{15, 4, 8, 0}}); err != nil {
		t.Fatal(err)
	}
	if got := level(e.Snapshot(neighbor), false, 0, 4, 8, 0); got != 0 {
		t.Fatalf("removed source retained light=%d", got)
	}
	if got := level(before, false, 0, 4, 8, 0); got != 13 {
		t.Fatalf("published snapshot mutated=%d", got)
	}
}

func TestSlabFacesOccludeSkyAndOpaqueBlocksCanEmit(t *testing.T) {
	e, err := New(0, 16, true)
	if err != nil {
		t.Fatal(err)
	}
	states := make([]uint16, 16*256)
	bottom := lightState(t, "oak_slab", map[string]string{"type": "bottom"})
	top := lightState(t, "oak_slab", map[string]string{"type": "top"})
	for i := 0; i < 256; i++ {
		states[8*256+i] = bottom
		states[7*256+i] = top
	}
	states[4*256+8*16+8] = lightState(t, "glowstone", nil)
	p := Position{0, 0}
	if _, err := e.Install(p, states, nil); err != nil {
		t.Fatal(err)
	}
	if got := level(e.Snapshot(p), true, 8, 6, 8, 0); got != 0 {
		t.Fatalf("slab roof leaked sky=%d", got)
	}
	if got := level(e.Snapshot(p), false, 8, 4, 8, 0); got != 15 {
		t.Fatalf("opaque emitter=%d", got)
	}
	if got := level(e.Snapshot(p), false, 9, 4, 8, 0); got != 14 {
		t.Fatalf("opaque emission did not escape=%d", got)
	}
}

func TestNoSkyDimensionAndInputSnapshotIsolation(t *testing.T) {
	e, err := New(-16, 32, false)
	if err != nil {
		t.Fatal(err)
	}
	states := make([]uint16, 32*256)
	states[8*256] = lightState(t, "lava", nil)
	p := Position{-1, -1}
	if _, err := e.Install(p, states, nil); err != nil {
		t.Fatal(err)
	}
	states[8*256] = 0
	if got := level(e.Snapshot(p), true, 0, 0, 0, -16); got != 0 {
		t.Fatalf("no-sky dimension sky=%d", got)
	}
	if got := level(e.Snapshot(p), false, 0, -8, 0, -16); got != 15 {
		t.Fatalf("mutable caller poisoned emission=%d", got)
	}
}
