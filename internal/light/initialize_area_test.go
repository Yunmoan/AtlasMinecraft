package light

import (
	"bytes"
	"testing"
)

func TestInitializeAreaMatchesSequentialCenterLighting(t *testing.T) {
	stone := lightState(t, "stone", nil)
	torch := lightState(t, "torch", nil)
	leaves := lightState(t, "oak_leaves", map[string]string{"distance": "1", "persistent": "false", "waterlogged": "false"})
	states := map[Position][]uint16{}
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			v := make([]uint16, 32*256)
			for y := 0; y < 8; y++ {
				for at := 0; at < 256; at++ {
					v[y*256+at] = stone
				}
			}
			for at := 0; at < 256; at++ {
				v[24*256+at] = leaves
			}
			for at := 0; at < 256; at++ {
				if at%17 != 0 {
					v[16*256+at] = stone
				}
			}
			v[9*256+8*16+15] = torch
			states[Position{x, z}] = v
		}
	}
	a, err := New(0, 32, true)
	if err != nil {
		t.Fatal(err)
	}
	b, err := New(0, 32, true)
	if err != nil {
		t.Fatal(err)
	}
	// Install the center last so the baseline also solves a complete area.
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			p := Position{x, z}
			if p == (Position{}) {
				continue
			}
			if _, err := a.Install(p, states[p], nil); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := a.Install(Position{}, states[Position{}], nil); err != nil {
		t.Fatal(err)
	}
	if _, err := b.InitializeArea(Position{}, states); err != nil {
		t.Fatal(err)
	}
	want, got := a.Snapshot(Position{}), b.Snapshot(Position{})
	for i := range want.Sky {
		if !bytes.Equal(want.Sky[i], got.Sky[i]) || !bytes.Equal(want.Block[i], got.Block[i]) {
			t.Fatalf("layer%d differs", i)
		}
	}
	if b.LoadedColumns() != 9 {
		t.Fatal(b.LoadedColumns())
	}
	if _, err := b.InitializeArea(Position{}, states); err == nil {
		t.Fatal("nonfresh engine accepted")
	}
}
