package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"testing"
)

func TestJoinGameDistancesMatchServerAndClientSettings(t *testing.T) {
	h := hub.New()
	w := world.New(h, 1)
	defer w.Close()
	s := New(Config{MaxPlayers: 20, ViewDistance: 5, SimulationDistance: 3}, h, w)
	for _, tc := range []struct {
		client, wantView int32
	}{
		{12, 5},
		{3, 3},
	} {
		p := player.New("Tester", [16]byte{}, 1, nil)
		p.SetViewDistance(s.effectiveViewDistance(tc.client))
		r := proto.NewReader(s.joinGamePayload(p))
		if _, err := r.I32(); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Bool(); err != nil {
			t.Fatal(err)
		}
		count, err := r.VarInt()
		if err != nil || count != 1 {
			t.Fatalf("dimension count: %d, %v", count, err)
		}
		if _, err := r.String(0); err != nil {
			t.Fatal(err)
		}
		if _, err := r.VarInt(); err != nil { // max players
			t.Fatal(err)
		}
		view, err := r.VarInt()
		if err != nil {
			t.Fatal(err)
		}
		simulation, err := r.VarInt()
		if err != nil || view != tc.wantView || simulation != 3 {
			t.Fatalf("client=%d: view=%d simulation=%d error=%v", tc.client, view, simulation, err)
		}
	}
}
