package hub

import (
	"atlas/internal/player"
	"testing"
)

func TestViewersAtReportedFarCoordinates(t *testing.T) {
	h := New()
	origin := player.New("Origin", [16]byte{1}, 1, nil)
	remote := player.New("Remote", [16]byte{2}, 2, nil)
	origin.SetPos(.5, 64, .5, 0, 0, true)
	remote.SetPos(10000722, 89, 9999902, 0, 0, false)
	h.Join(origin)
	h.Join(remote)
	for _, position := range [][2]int32{{0, 0}, {10000722, 9999902}} {
		viewers := h.ViewersOf(position[0], position[1])
		want := origin
		if position[0] != 0 {
			want = remote
		}
		if len(viewers) != 1 || viewers[0] != want {
			t.Fatalf("viewers at %v include a player millions of blocks away", position)
		}
	}
}
