package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"testing"
)

func interactionServer(t *testing.T) (*Server, *player.Player) {
	t.Helper()
	dir := t.TempDir()
	seed := int64(0)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(h, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	p := player.New("Builder", [16]byte{1}, 1, nil)
	p.SetPos(.5, -60, .5, 0, 0, true)
	if !h.Join(p) {
		t.Fatal("join")
	}
	return New(Config{}, h, w), p
}

func placePayload(x, y, z, face, sequence int32, cursorY float32) []byte {
	return proto.NewWriter().VarInt(0).I64(proto.PackPosition(x, y, z)).VarInt(face).F32(.5).F32(cursorY).F32(.5).Bool(false).VarInt(sequence).Payload()
}

func assertInteractionState(t *testing.T, w *world.World, x, y, z int32, state uint32) {
	t.Helper()
	got, err := w.BlockAt(x, y, z)
	if err != nil || got != state {
		t.Fatalf("%d %d %d: %d want %d: %v", x, y, z, got, state, err)
	}
}
func assertAck(t *testing.T, p *player.Player, sequence int32) {
	t.Helper()
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID == proto.SAckBlockChange {
			got, err := proto.NewReader(packet.Payload).VarInt()
			if err != nil || got != sequence {
				t.Fatalf("ack %d: %v", got, err)
			}
			return
		}
	}
	t.Fatalf("no prediction ack for %d", sequence)
}

func TestCreativePlacementValidationAndPredictionAck(t *testing.T) {
	s, p := interactionServer(t)
	p.SetHotbarSlot(36, world.StateStone)
	s.handlePlace(p, placePayload(1, -61, 1, 1, 1, .5))
	assertInteractionState(t, s.World, 1, -60, 1, world.StateStone)
	assertAck(t, p, 1)
	p.SetHotbarSlot(36, world.StateDirt)
	s.handlePlace(p, placePayload(1, -61, 1, 1, 2, .5))
	assertInteractionState(t, s.World, 1, -60, 1, world.StateStone)
	assertAck(t, p, 2)
	s.handlePlace(p, placePayload(0, -61, 0, 1, 3, .5))
	assertInteractionState(t, s.World, 0, -60, 0, world.StateAir)
	assertAck(t, p, 3)
	s.handlePlace(p, placePayload(13, -61, 13, 1, 4, .5))
	assertInteractionState(t, s.World, 13, -60, 13, world.StateAir)
	assertAck(t, p, 4)
	grass, err := world.BlockStateID("short_grass", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.World.SetBlock(2, -60, 1, grass); err != nil {
		t.Fatal(err)
	}
	s.handlePlace(p, placePayload(2, -60, 1, 1, 5, .5))
	assertInteractionState(t, s.World, 2, -60, 1, world.StateDirt)
	assertInteractionState(t, s.World, 2, -59, 1, world.StateAir)
	assertAck(t, p, 5)
	// A non-block stack must clear the previous held block.
	s.handleCreativeSlot(p, proto.NewWriter().I16(36).VarInt(64).VarInt(9999).VarInt(0).VarInt(0).Payload())
	if p.HeldBlock() != world.StateAir {
		t.Fatal("unsupported item left stale held block")
	}
	s.handlePlace(p, placePayload(3, -61, 1, 1, 6, .5))
	assertInteractionState(t, s.World, 3, -60, 1, world.StateAir)
	assertAck(t, p, 6)
}

func TestCreativeSlabLogAndDoublePlantPlacement(t *testing.T) {
	s, p := interactionServer(t)
	id := func(name string, properties map[string]string) uint32 {
		state, err := world.BlockStateID(name, properties)
		if err != nil {
			t.Fatal(err)
		}
		return state
	}
	slab := id("oak_slab", nil)
	p.SetHotbarSlot(36, slab)
	s.handlePlace(p, placePayload(1, -61, 2, 1, 10, .5))
	assertInteractionState(t, s.World, 1, -60, 2, id("oak_slab", map[string]string{"type": "bottom"}))
	assertAck(t, p, 10)
	s.handlePlace(p, placePayload(1, -60, 2, 1, 11, .5))
	assertInteractionState(t, s.World, 1, -60, 2, id("oak_slab", map[string]string{"type": "double"}))
	assertAck(t, p, 11)
	p.SetHotbarSlot(36, id("oak_log", nil))
	s.handlePlace(p, placePayload(1, -60, 2, 5, 12, .5))
	assertInteractionState(t, s.World, 2, -60, 2, id("oak_log", map[string]string{"axis": "x"}))
	assertAck(t, p, 12)
	lower, upper := id("tall_grass", map[string]string{"half": "lower"}), id("tall_grass", map[string]string{"half": "upper"})
	p.SetHotbarSlot(36, lower)
	s.handlePlace(p, placePayload(-1, -61, -1, 1, 13, .5))
	assertInteractionState(t, s.World, -1, -60, -1, lower)
	assertInteractionState(t, s.World, -1, -59, -1, upper)
	assertAck(t, p, 13)
	if err := s.World.SetBlock(-2, -59, -1, world.StateStone); err != nil {
		t.Fatal(err)
	}
	s.handlePlace(p, placePayload(-2, -61, -1, 1, 14, .5))
	assertInteractionState(t, s.World, -2, -60, -1, world.StateAir)
	assertAck(t, p, 14)
}

func TestCreativeDigDistanceCancelAndFinish(t *testing.T) {
	s, p := interactionServer(t)
	if err := s.World.SetBlock(1, -60, 1, world.StateStone); err != nil {
		t.Fatal(err)
	}
	if err := s.World.SetBlock(13, -60, 13, world.StateStone); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		status, x, z, seq int32
		want              uint32
	}{{0, 13, 13, 20, world.StateStone}, {1, 1, 1, 21, world.StateStone}, {2, 1, 1, 22, world.StateStone}, {0, 1, 1, 23, world.StateAir}} {
		s.handleDig(p, proto.NewWriter().VarInt(test.status).I64(proto.PackPosition(test.x, -60, test.z)).I8(1).VarInt(test.seq).Payload())
		assertInteractionState(t, s.World, test.x, -60, test.z, test.want)
		assertAck(t, p, test.seq)
	}
}

func TestMovementCorrectsWallCrossingEvenWhileFlying(t *testing.T) {
	s, p := interactionServer(t)
	if err := s.World.SetBlock(0, -60, 1, world.StateStone); err != nil {
		t.Fatal(err)
	}
	p.SetFlying(true)
	s.handleMovement(p, proto.CPlayerPosition, proto.NewWriter().F64(.5).F64(-60).F64(2.5).Bool(false).Payload())
	x, y, z, _, _, _ := p.Pos()
	if x != .5 || y != -60 || z != .5 {
		t.Fatalf("wall crossing accepted: %v %v %v", x, y, z)
	}
	corrected := false
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID == proto.SPositionSync {
			corrected = true
			r := proto.NewReader(packet.Payload)
			for range 3 {
				r.F64()
			}
			for range 2 {
				r.F32()
			}
			r.U8()
			id, _ := r.VarInt()
			s.handlePlayPacket(p, proto.CTeleportConfirm, proto.NewWriter().VarInt(id).Payload())
		}
	}
	if !corrected {
		t.Fatal("no authoritative correction")
	}
	p.SetPos(-1.5, -60, .5, 0, 0, true)
	s.handleMovement(p, proto.CPlayerPosition, proto.NewWriter().F64(-.5).F64(-60).F64(.5).Bool(true).Payload())
	x, _, _, _, _, _ = p.Pos()
	if x != -.5 {
		t.Fatal("clear negative-coordinate movement rejected")
	}
}
