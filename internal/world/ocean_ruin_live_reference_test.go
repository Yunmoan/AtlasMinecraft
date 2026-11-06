package world

import (
	"atlas/internal/nbt"
	"atlas/internal/tickqueue"
	"encoding/base64"
	"encoding/json"
	"os"
	"slices"
	"testing"
)

type nativeRuinDispatch struct {
	Time     int64    `json:"time"`
	Kind     string   `json:"kind"`
	Position [3]int32 `json:"position"`
	Type     string   `json:"type"`
}

// Observe the real production scheduler. Native callbacks were recorded at
// ServerLevel.tickBlock/tickFluid entry, before stale-type checks, at the same
// boundary as scheduledTickOwner's callback. No queue sorting is permitted.
func observeStructureRuntime(t *testing.T, w *World, x, z int32, fixture string) func() {
	t.Helper()
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var reference struct {
		Version string `json:"version"`
		Seed    int64  `json:"seed"`
		Ticks   int    `json:"ticks"`
		Chunks  []struct {
			X, Z           int32
			Initial, Final [2]string
			Dispatch       []nativeRuinDispatch
		} `json:"chunks"`
	}
	if err := json.Unmarshal(data, &reference); err != nil {
		t.Fatal(err)
	}
	if reference.Version != "1.21.1" || reference.Seed != 1234 || reference.Ticks != 100 || len(reference.Chunks) != 2 {
		t.Fatal("invalid native live reference metadata")
	}
	index := -1
	for i, ch := range reference.Chunks {
		if ch.X == x && ch.Z == z {
			if index >= 0 {
				t.Fatal("duplicate native live chunk")
			}
			index = i
		}
	}
	if index < 0 {
		t.Fatal("missing native live chunk")
	}
	want := reference.Chunks[index]
	checkQueues := func(encoded [2]string, phase string) {
		t.Helper()
		o := w.scheduledTicks
		o.mu.Lock()
		h := o.chunks[tickqueue.ChunkKey(x, z)]
		o.mu.Unlock()
		if h == nil {
			t.Fatal("missing live chunk owner")
		}
		fluid, block, _, _, _ := h.snapshot()
		for i, actual := range [2]nbt.Tag{block, fluid} {
			raw, err := base64.StdEncoding.DecodeString(encoded[i])
			if err != nil {
				t.Fatal(err)
			}
			tag, n, err := nbt.ParseNetwork(raw)
			if err != nil || n != len(raw) {
				t.Fatal("invalid native saved tick list", err)
			}
			list, ok := tag.Elements()
			if !ok {
				t.Fatal("native ticks are not a list")
			}
			decode := func(tag nbt.Tag) []tickqueue.Saved[string] {
				return tickqueue.DecodeSavedNBT(tag, x, z, func(name string) (string, bool) { return name, true })
			}
			got, expected := decode(actual), decode(tag)
			if len(expected) != len(list) {
				t.Fatal("native tick reference has invalid fields")
			}
			if !slices.Equal(got, expected) {
				first := 0
				for first < min(len(got), len(expected)) && got[first] == expected[first] {
					first++
				}
				t.Errorf("%s live %s queue differs: got %d want %d; first=%d got=%v want=%v", phase, [2]string{"block", "fluid"}[i], len(got), len(expected), first, got[max(0, first-1):min(len(got), first+4)], expected[max(0, first-1):min(len(expected), first+4)])
			}
		}
	}
	checkQueues(want.Initial, "initial")
	initialTime, _ := w.Time()
	var actual []nativeRuinDispatch
	w.blockUpdateMu.Lock()
	w.scheduledTicks.observeDispatch = func(time int64, kind string, p tickqueue.Position, name string) {
		if p.X>>4 == x && p.Z>>4 == z {
			actual = append(actual, nativeRuinDispatch{time - initialTime, kind, [3]int32{p.X, p.Y, p.Z}, name})
		}
	}
	w.blockUpdateMu.Unlock()
	return func() {
		t.Helper()
		checkQueues(want.Final, "final")
		if !slices.Equal(actual, want.Dispatch) {
			first := 0
			for first < min(len(actual), len(want.Dispatch)) && actual[first] == want.Dispatch[first] {
				first++
			}
			t.Errorf("ordered native dispatch differs: got %d want %d; first=%d got=%v want=%v", len(actual), len(want.Dispatch), first, actual[max(0, first-1):min(len(actual), first+3)], want.Dispatch[max(0, first-1):min(len(want.Dispatch), first+3)])
		} else {
			t.Logf("matched %d ordered native dispatches", len(actual))
		}
		w.blockUpdateMu.Lock()
		w.scheduledTicks.observeDispatch = nil
		w.blockUpdateMu.Unlock()
	}
}
