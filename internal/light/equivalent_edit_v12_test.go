package light

import "testing"

func TestLightEquivalentEditsRetainSnapshotsAndUpdateOwnedStates(t *testing.T) {
	e, err := New(0, 32, true)
	if err != nil {
		t.Fatal(err)
	}
	stone, dirt := lightState(t, "stone", nil), lightState(t, "dirt", nil)
	glass := lightState(t, "glass", nil)
	states := make([]uint16, 32*256)
	states[4*256+8*16+8] = stone
	p := Position{-1, 0}
	if _, err := e.Install(p, states, nil); err != nil {
		t.Fatal(err)
	}
	before := e.Snapshot(p)
	for _, change := range []Change{{8, 4, 8, dirt}, {9, 4, 8, glass}, {9, 4, 8, 0}} {
		updates, err := e.Edit(p, []Change{change})
		if err != nil || len(updates) != 0 || e.Snapshot(p) != before {
			t.Fatal("equal original light rules relit or changed a published snapshot", err)
		}
		if got := e.columns[p].states[change.Y*256+change.Z*16+change.X]; got != change.State {
			t.Fatal("equivalent edit did not update private block state")
		}
	}
	// Validation precedes every write, including on the equivalent fast path.
	if _, err := e.Edit(p, []Change{{8, 4, 8, stone}, {16, 4, 8, stone}}); err == nil {
		t.Fatal("invalid edit batch accepted")
	}
	if e.columns[p].states[4*256+8*16+8] != dirt {
		t.Fatal("invalid batch partially applied an equivalent edit")
	}
}

func TestLightEditSignatureIncludesFacesOpacityAndEmission(t *testing.T) {
	e, err := New(0, 32, true)
	if err != nil {
		t.Fatal(err)
	}
	p := Position{}
	if _, err := e.Install(p, make([]uint16, 32*256), nil); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		state        uint16
		sky, emitted byte
	}{
		{lightState(t, "oak_slab", map[string]string{"type": "bottom"}), 15, 0},
		{lightState(t, "oak_slab", map[string]string{"type": "top"}), 14, 0},
		{lightState(t, "water", nil), 14, 0},
		{lightState(t, "glowstone", nil), 0, 15},
		{0, 15, 0},
	} {
		if _, err := e.Edit(p, []Change{{8, 8, 8, test.state}}); err != nil {
			t.Fatal(err)
		}
		data := e.Snapshot(p)
		if got := level(data, true, 8, 8, 8, 0); got != test.sky {
			t.Fatal("directional faces or opacity did not update skylight", test.state, got, test.sky)
		}
		if got := level(data, false, 8, 8, 8, 0); got != test.emitted {
			t.Fatal("emission change or removal was incorrectly skipped", test.state, got, test.emitted)
		}
	}
}

func TestEquivalentEditRetainsImportedTrustedLighting(t *testing.T) {
	e, err := New(0, 32, true)
	if err != nil {
		t.Fatal(err)
	}
	stone, dirt := lightState(t, "stone", nil), lightState(t, "dirt", nil)
	states := make([]uint16, 32*256)
	states[8*256] = stone
	imported := &Data{Sky: make([][]byte, 4), Block: make([][]byte, 4), Preserved: true}
	imported.Block[1] = make([]byte, 2048)
	imported.Block[1][0] = 0x73 // retained contribution from an unloaded neighbor
	p := Position{-1, -1}
	if _, err := e.Install(p, states, imported); err != nil {
		t.Fatal(err)
	}
	before := e.Snapshot(p)
	if !e.columns[p].trusted || before == imported || !before.Preserved {
		t.Fatal("imported lighting not independently owned and trusted")
	}
	if updates, err := e.Edit(p, []Change{{0, 8, 0, dirt}}); err != nil || len(updates) != 0 {
		t.Fatal("equivalent edit relit imported neighborhood", err)
	}
	if e.Snapshot(p) != before || !e.columns[p].trusted || e.columns[p].states[8*256] != dirt || before.Block[1][0] != 0x73 {
		t.Fatal("equivalent edit invalidated original snapshot or imported trust")
	}
}
