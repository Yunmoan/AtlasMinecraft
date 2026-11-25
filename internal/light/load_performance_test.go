package light

import (
	"reflect"
	"testing"
)

// Replay loading against the original solver, including sparse trusted edges,
// completion of an imported neighborhood, untrusted terrain and a block edit.
func TestImportedLightFastPathMatchesSolver(t *testing.T) {
	for _, hasSky := range []bool{false, true} {
		fast, _ := New(-16, 32, hasSky)
		old, _ := New(-16, 32, hasSky)
		states := make([]uint16, 32*256)
		for i := 0; i < 16*256; i++ {
			states[i] = lightState(t, "stone", nil)
		}
		states[17*256+15] = lightState(t, "torch", nil)
		for x := int32(-2); x <= 2; x++ {
			for z := int32(-2); z <= 2; z++ {
				p := Position{x, z}
				var imported *Data
				if x != 2 {
					imported = &Data{Sky: make([][]byte, 4), Block: make([][]byte, 4), Preserved: true, Complete: z == -2}
				}
				updates, err := fast.Install(p, states, imported)
				if err != nil {
					t.Fatal(err)
				}
				// Supply exactly the same installed state, then call the original
				// solver directly to bypass the new no-output guard.
				copy := fast.columns[p]
				old.columns[p] = &column{states: append([]uint16(nil), states...), light: imported, trusted: copy.trusted}
				want := old.solve(p)
				if !reflect.DeepEqual(updates, want) {
					t.Fatalf("install %v sky=%t changed updates", p, hasSky)
				}
				for position, a := range fast.columns {
					if !reflect.DeepEqual(a.light, old.columns[position].light) {
						t.Fatalf("install %v changed snapshot %v", p, position)
					}
				}
			}
		}
		changes := []Change{{X: 15, Y: 1, Z: 0, State: 0}}
		got, errA := fast.Edit(Position{}, changes)
		want, errB := old.Edit(Position{}, changes)
		if errA != nil || errB != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("edit differed: %v/%v", errA, errB)
		}
	}
}

func TestImportedLightFastPathAvoidsScratchGrid(t *testing.T) {
	e, _ := New(0, 384, true)
	data := &Data{Sky: make([][]byte, 26), Block: make([][]byte, 26), Preserved: true}
	updates, err := e.Install(Position{}, make([]uint16, 384*256), data)
	if err != nil || len(updates) != 0 {
		t.Fatalf("install: %v, %v", updates, err)
	}
	if e.work != nil {
		t.Fatal("trusted incomplete load allocated an unused solver grid")
	}
	if e.revision != 1 {
		t.Fatal("skipped solve changed revision ordering")
	}
}
