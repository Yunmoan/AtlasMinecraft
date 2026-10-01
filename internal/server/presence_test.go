package server

import (
	"atlas/internal/hub"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"context"
	"testing"
)

func presenceMatches(t *testing.T, packet player.OutPacket, name string, joined bool) bool {
	t.Helper()
	if packet.ID != proto.SSystemChat {
		return false
	}
	tag, consumed, err := nbt.ParseNetwork(packet.Payload)
	if err != nil || consumed+1 != len(packet.Payload) || packet.Payload[consumed] != 0 {
		t.Fatalf("invalid system chat: %v", err)
	}
	object, _ := tag.Object()
	key, _ := object.Text("translate")
	expected := "multiplayer.player.left"
	if joined {
		expected = "multiplayer.player.joined"
	}
	if key != expected {
		return false
	}
	color, _ := object.Text("color")
	args, _ := object["with"].Elements()
	if color != "yellow" || len(args) != 1 {
		t.Fatal("presence message lost its style or player argument")
	}
	return plainComponent(args[0]) == name
}

func TestPresenceBroadcastIncludesSelfAndSkipsConnectionsBeforePlay(t *testing.T) {
	h := hub.New()
	a := player.New("Alice", [16]byte{1}, 1, nil)
	b := player.New("Bob", [16]byte{2}, 2, nil)
	pending := player.New("Pending", [16]byte{3}, 3, nil)
	a.SetPlayReady(true)
	b.SetPlayReady(true)
	if !h.Join(a) || !h.Join(b) || !h.Join(pending) {
		t.Fatal("could not join test players")
	}
	s := New(Config{}, h, nil)
	s.broadcastPlayerPresence("Bob", true)
	for _, p := range []*player.Player{a, b} {
		select {
		case packet := <-p.Out:
			if !presenceMatches(t, packet, "Bob", true) {
				t.Fatalf("%s did not receive Bob's join message", p.Name)
			}
		default:
			t.Fatalf("%s received no join message", p.Name)
		}
	}
	if len(pending.Out) != 0 {
		t.Fatal("connection received chat before entering play")
	}
	h.Leave(b)
	s.broadcastPlayerPresence("Bob", false)
	if packet := <-a.Out; !presenceMatches(t, packet, "Bob", false) {
		t.Fatal("remaining player received no leave message")
	}
}

func TestThreeTCPClientsReceiveJoinAndLeaveBroadcastsAcrossDistance(t *testing.T) {
	dir, seed := t.TempDir(), int64(1234)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(h, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	if err := w.PrepareSpawn(context.Background(), 0, 0, 0, nil); err != nil {
		t.Fatal(err)
	}
	s := New(Config{MaxPlayers: 10, ViewDistance: 2, SimulationDistance: 2, Ops: map[string]bool{"farprobe": true}}, h, w)
	address := serveTrackingConnections(t, s, 3)
	a := connectTrackingClient(t, address, "NearProbe")
	a.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "NearProbe", true) })
	b := connectTrackingClient(t, address, "FarProbe")
	b.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "FarProbe", true) })
	a.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "FarProbe", true) })
	if err := b.send(proto.CChatCommand, proto.NewWriter().String("tp 10000000 1024 10000000").Payload()); err != nil {
		t.Fatal(err)
	}
	b.await(t, func(p player.OutPacket) bool {
		if p.ID != proto.SPositionSync {
			return false
		}
		r := proto.NewReader(p.Payload)
		x, _ := r.F64()
		y, _ := r.F64()
		z, _ := r.F64()
		return x == 10000000.5 && y == 1024 && z == 10000000.5
	})
	c := connectTrackingClient(t, address, "JoinProbe")
	c.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "JoinProbe", true) })
	a.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "JoinProbe", true) })
	b.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "JoinProbe", true) })
	c.conn.Conn.Close()
	a.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "JoinProbe", false) })
	b.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "JoinProbe", false) })
	b.conn.Conn.Close()
	a.await(t, func(p player.OutPacket) bool { return presenceMatches(t, p, "FarProbe", false) })
}
