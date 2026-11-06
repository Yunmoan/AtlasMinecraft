package world

import (
	"atlas/internal/tickqueue"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

func checkFrozenActivationTicks(t *testing.T, w *World, cx, cz int32) {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/worldgen/native_generation_block_ticks_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var want []tickqueue.Saved[string]
	var x, z int32
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var label string
		if strings.HasPrefix(line, "native_block_chunk ") {
			var seed int64
			var count int
			if _, err := fmt.Sscan(line, &label, &seed, &x, &z, &count); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if x != cx || z != cz {
			continue
		}
		var tick tickqueue.Saved[string]
		var priority int32
		if _, err := fmt.Sscan(line, &label, &tick.Type, &tick.Position.X, &tick.Position.Y, &tick.Position.Z, &tick.Delay, &priority); err != nil {
			t.Fatal(err)
		}
		tick.Priority = tickqueue.ClampPriority(priority)
		want = append(want, tick)
	}
	w.scheduledTicks.mu.Lock()
	h := w.scheduledTicks.chunks[tickqueue.ChunkKey(cx, cz)]
	w.scheduledTicks.mu.Unlock()
	_, blocks, _, _, _ := h.snapshot()
	got := tickqueue.DecodeSavedNBT(blocks, cx, cz, func(name string) (string, bool) { return name, true })
	if !slices.Equal(got, want) {
		actual := map[tickqueue.Saved[string]]bool{}
		for _, row := range got {
			actual[row] = true
		}
		missing := 0
		for _, row := range want {
			if !actual[row] {
				if missing < 8 {
					t.Logf("missing native request %+v", row)
				}
				missing++
			}
			delete(actual, row)
		}
		t.Fatalf("activation %d,%d: got %d want %d, missing=%d extra=%v; ordered lists differ", cx, cz, len(got), len(want), missing, actual)
	}
	t.Logf("all %d ordered frozen activation block requests match", len(want))
}
