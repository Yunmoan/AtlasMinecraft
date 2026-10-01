package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"math"
	"testing"
)

func TestRelativeMovementPreservesFixedPointPosition(t *testing.T) {
	from := [3]float64{-1.001, 64, .001}
	initial := from
	var sum [3]int64
	for i := 0; i < 1000; i++ {
		to := [3]float64{from[0] + .00037, from[1], from[2] - .00027}
		id, payload := entityMovementPacket(3, from, to, 0, 0, true, false, true)
		if id != proto.SRelEntityMove {
			t.Fatalf("ordinary move used %#x", id)
		}
		r := proto.NewReader(payload)
		eid, _ := r.VarInt()
		if eid != 3 {
			t.Fatal(eid)
		}
		for axis := range sum {
			d, err := r.I16()
			if err != nil {
				t.Fatal(err)
			}
			sum[axis] += int64(d)
		}
		if g, err := r.Bool(); err != nil || !g || r.Remaining() != 0 {
			t.Fatal("bad relative packet")
		}
		from = to
	}
	for axis := range sum {
		want := int64(math.Floor(from[axis]*4096)) - int64(math.Floor(initial[axis]*4096))
		if sum[axis] != want {
			t.Fatalf("axis %d drift %d want %d", axis, sum[axis], want)
		}
	}
	id, _ := entityMovementPacket(3, from, [3]float64{from[0] + 8, 64, 0}, 0, 0, true, false, false)
	if id != proto.SEntityTeleport {
		t.Fatal("out of range movement needs absolute teleport")
	}
	id, _ = entityMovementPacket(3, from, from, 45, 0, false, true, true)
	if id != proto.SEntityLook {
		t.Fatal("rotation-only move")
	}
	id, _ = entityMovementPacket(3, from, [3]float64{from[0] + 1, 64, 0}, 45, 0, true, true, true)
	if id != proto.SEntityMoveLook {
		t.Fatal("combined position and look")
	}
}

func TestEntityWaitsForDeliveredChunkAndDoesNotReceiveDuplicateDelta(t *testing.T) {
	h := hub.New()
	s := New(Config{}, h, nil)
	a := player.New("Alice", [16]byte{1}, 1, nil)
	b := player.New("Bob", [16]byte{2}, 2, nil)
	a.SetPlayReady(true)
	b.SetPlayReady(true)
	a.SetPos(.5, 64, .5, 0, 0, true)
	b.SetPos(16.5, 64, .5, 0, 0, true)
	h.Join(a)
	h.Join(b)
	a.ReserveChunk(1, 0)
	s.updateVisibility(a)
	if a.IsVisible(b.EID) || len(a.Out) != 0 {
		t.Fatal("spawn preceded chunk delivery")
	}
	if !a.SendRequestedChunk(1, 0, []byte{1}) {
		t.Fatal("delivery failed")
	}
	for len(a.Out) > 0 {
		<-a.Out
	}
	s.updateVisibility(a)
	if !a.IsVisible(b.EID) {
		t.Fatal("entity not spawned after delivery")
	}
	spawn := <-a.Out
	if spawn.ID != proto.SSpawnEntity {
		t.Fatal("spawn order")
	}
	for len(a.Out) > 0 {
		<-a.Out
	}
	s.entityMu.Lock()
	b.SetPos(16.75, 64, .5, 0, 0, true)
	id, payload := entityMovementPacket(b.EID, [3]float64{16.5, 64, .5}, [3]float64{16.75, 64, .5}, 0, 0, true, false, true)
	s.sendTracked(b, id, payload)
	s.updateVisibilityLocked(b)
	s.entityMu.Unlock()
	if len(a.Out) != 1 || (<-a.Out).ID != proto.SRelEntityMove {
		t.Fatal("existing viewer did not get exactly one move")
	}
	if len(b.Out) != 0 {
		t.Fatal("unspawned entity got movement")
	}
	a.PruneChunks(10, 0, 4)
	for len(a.Out) > 0 {
		<-a.Out
	}
	s.updateVisibility(a)
	if a.IsVisible(b.EID) {
		t.Fatal("unloaded entity remains visible")
	}
}

func TestTeleportConfirmGatesStaleMovement(t *testing.T) {
	s, p := interactionServer(t)
	s.syncPosition(p)
	packet := <-p.Out
	r := proto.NewReader(packet.Payload)
	for range 3 {
		r.F64()
	}
	for range 2 {
		r.F32()
	}
	r.U8()
	id, _ := r.VarInt()
	movement := proto.NewWriter().F64(1.5).F64(-60).F64(.5).Bool(true).Payload()
	s.handleMovement(p, proto.CPlayerPosition, movement)
	if x, _, _, _, _, _ := p.Pos(); x != .5 {
		t.Fatal("movement accepted before confirm")
	}
	s.handlePlayPacket(p, proto.CTeleportConfirm, proto.NewWriter().VarInt(id-1).Payload())
	if !p.TeleportPending() {
		t.Fatal("wrong confirmation released movement")
	}
	s.handlePlayPacket(p, proto.CTeleportConfirm, proto.NewWriter().VarInt(id).Payload())
	s.handleMovement(p, proto.CPlayerPosition, movement)
	if x, _, _, _, _, _ := p.Pos(); x != 1.5 {
		t.Fatal("movement not resumed after confirm")
	}
}
