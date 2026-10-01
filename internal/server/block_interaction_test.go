package server

import (
	"atlas/internal/hub"
	"atlas/internal/item"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"testing"
)

func blockAt(t *testing.T, w *world.World, x, y, z int32) world.BlockState {
	t.Helper()
	id, err := w.BlockAt(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	state, err := world.BlockStateByID(id)
	if err != nil {
		t.Fatal(err)
	}
	return state
}
func stateID(t *testing.T, name string, properties map[string]string) uint32 {
	t.Helper()
	id, err := world.BlockStateID(name, properties)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestDoorPlacementUseCollisionAndSupport(t *testing.T) {
	s, p := interactionServer(t)
	p.SetHotbarSlot(36, stateID(t, "oak_door", nil))
	s.handlePlace(p, placePayload(2, -61, 2, 1, 101, .5))
	for y, half := range map[int32]string{-60: "lower", -59: "upper"} {
		state := blockAt(t, s.World, 2, y, 2)
		if state.Name != "minecraft:oak_door" || state.Properties["half"] != half || state.Properties["facing"] != "south" {
			t.Fatalf("door half/orientation: %+v", state)
		}
	}
	from, to := [3]float64{2.5, -60, 1.5}, [3]float64{2.5, -60, 3.5}
	allowed, err := s.World.CanMovePlayer(from, to, 1.8, false)
	if err != nil || allowed {
		t.Fatal("walked through closed door", err)
	}
	p.SetHotbarSlot(36, world.StateAir)
	s.handlePlace(p, placePayload(2, -59, 2, 2, 102, .5))
	for _, y := range []int32{-60, -59} {
		if blockAt(t, s.World, 2, y, 2).Properties["open"] != "true" {
			t.Fatal("upper click did not open both halves")
		}
	}
	allowed, err = s.World.CanMovePlayer(from, to, 1.8, false)
	if err != nil || !allowed {
		t.Fatal("open door blocked the doorway", err)
	}
	if err := s.World.SetBlock(2, -61, 2, world.StateAir); err != nil {
		t.Fatal(err)
	}
	for _, y := range []int32{-60, -59} {
		assertInteractionState(t, s.World, 2, y, 2, world.StateAir)
	}
}

func TestDoorUpperHalfCannotIntersectAnotherPlayer(t *testing.T) {
	s, p := interactionServer(t)
	p.SetHotbarSlot(36, stateID(t, "oak_door", nil))
	o := player.New("Other", [16]byte{2}, 2, nil)
	o.SetPos(2.5, -59, 2.08, 0, 0, false)
	if !s.Hub.Join(o) {
		t.Fatal("join observer")
	}
	s.handlePlace(p, placePayload(2, -61, 2, 1, 110, .5))
	assertInteractionState(t, s.World, 2, -60, 2, world.StateAir)
	assertInteractionState(t, s.World, 2, -59, 2, world.StateAir)
}

func TestTorchWallOrientationAndCrossRegionSupportRemoval(t *testing.T) {
	s, p := interactionServer(t)
	p.SetHotbarSlot(36, stateID(t, "torch", nil))
	if err := s.World.SetBlock(2, -60, 2, world.StateStone); err != nil {
		t.Fatal(err)
	}
	s.handlePlace(p, placePayload(2, -60, 2, 5, 120, .5))
	state := blockAt(t, s.World, 3, -60, 2)
	if state.Name != "minecraft:wall_torch" || state.Properties["facing"] != "east" {
		t.Fatalf("wrong wall torch: %+v", state)
	}
	if err := s.World.SetBlock(2, -60, 2, world.StateAir); err != nil {
		t.Fatal(err)
	}
	assertInteractionState(t, s.World, 3, -60, 2, world.StateAir)
	// x=511/512 crosses a region boundary, not only a chunk boundary.
	for _, cx := range []int32{31, 32} {
		if _, err := s.World.LoadChunk(cx, 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.World.SetBlock(511, -60, 2, world.StateStone); err != nil {
		t.Fatal(err)
	}
	wallTorch := stateID(t, "wall_torch", map[string]string{"facing": "east"})
	if err := s.World.SetBlock(512, -60, 2, wallTorch); err != nil {
		t.Fatal(err)
	}
	if err := s.World.SetBlock(511, -60, 2, world.StateAir); err != nil {
		t.Fatal(err)
	}
	assertInteractionState(t, s.World, 512, -60, 2, world.StateAir)
}

func TestOpenablesOffhandAndFenceConnections(t *testing.T) {
	s, p := interactionServer(t)
	gate := stateID(t, "oak_fence_gate", nil)
	p.SetHotbarSlot(36, gate)
	s.handlePlace(p, placePayload(2, -61, 2, 1, 130, .5))
	if blockAt(t, s.World, 2, -60, 2).Properties["facing"] != "south" {
		t.Fatal("gate ignored player's facing")
	}
	p.SetHotbarSlot(36, world.StateAir)
	s.handlePlace(p, placePayload(2, -60, 2, 2, 131, .5))
	if blockAt(t, s.World, 2, -60, 2).Properties["open"] != "true" {
		t.Fatal("gate did not open")
	}
	stone, _ := item.ID("stone")
	v := item.Inventory{}
	v.Slots[45] = item.New(stone, 64)
	s.setInventory(p, v)
	packet := proto.NewWriter().VarInt(1).I64(proto.PackPosition(1, -61, 2)).VarInt(1).F32(.5).F32(.5).F32(.5).Bool(false).VarInt(132).Payload()
	s.handlePlace(p, packet)
	assertInteractionState(t, s.World, 1, -60, 2, world.StateStone)
	fence := stateID(t, "oak_fence", nil)
	for _, x := range []int32{3, 4} {
		if err := s.World.SetBlock(x, -60, 1, fence); err != nil {
			t.Fatal(err)
		}
	}
	if blockAt(t, s.World, 3, -60, 1).Properties["east"] != "true" || blockAt(t, s.World, 4, -60, 1).Properties["west"] != "true" {
		t.Fatal("adjacent fences did not connect")
	}
	if err := s.World.SetBlock(4, -60, 1, world.StateAir); err != nil {
		t.Fatal(err)
	}
	if blockAt(t, s.World, 3, -60, 1).Properties["east"] != "false" {
		t.Fatal("removed fence left a ghost collision arm")
	}
}

func TestEquipmentPacketHasSelectedHandArmorAndOffhand(t *testing.T) {
	p := player.New("equipped", [16]byte{}, 1, nil)
	stone, _ := item.ID("stone")
	helmet, _ := item.ID("diamond_helmet")
	shield, _ := item.ID("shield")
	v := item.Inventory{Selected: 2}
	v.Slots[38] = item.New(stone, 4)
	v.Slots[5] = item.New(helmet, 1)
	v.Slots[45] = item.New(shield, 1)
	s := New(Config{}, hub.New(), nil)
	s.setInventory(p, v)
	r := proto.NewReader(equipmentPayload(p))
	r.VarInt()
	for index, want := range []item.Stack{v.Slots[38], v.Slots[45], {}, {}, {}, v.Slots[5]} {
		slot, err := r.U8()
		if err != nil || int(slot&127) != index || (slot&128 != 0) != (index < 5) {
			t.Fatal("bad equipment continuation")
		}
		got, err := item.Read(r)
		if err != nil || got.ID != want.ID || got.Count != want.Count {
			t.Fatal("bad equipment item", err)
		}
	}
	if r.Remaining() != 0 {
		t.Fatal("extra equipment bytes")
	}
}

func TestHangingLanternAndFenceConnectionExceptions(t *testing.T) {
	s, p := interactionServer(t)
	fence := stateID(t, "oak_fence", nil)
	// A fence supports the center of its underside without a full sturdy face.
	if err := s.World.SetBlock(3, -58, 2, fence); err != nil {
		t.Fatal(err)
	}
	p.SetHotbarSlot(36, stateID(t, "lantern", nil))
	s.handlePlace(p, placePayload(3, -58, 2, 0, 140, .5))
	lantern := blockAt(t, s.World, 3, -59, 2)
	if lantern.Name != "minecraft:lantern" || lantern.Properties["hanging"] != "true" {
		t.Fatal("center-supported hanging lantern rejected")
	}
	if err := s.World.SetBlock(3, -58, 2, world.StateAir); err != nil {
		t.Fatal(err)
	}
	assertInteractionState(t, s.World, 3, -59, 2, world.StateAir)
	if err := s.World.SetBlock(3, -60, 2, fence); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"oak_leaves", "barrier", "pumpkin", "shulker_box"} {
		if err := s.World.SetBlock(4, -60, 2, stateID(t, name, nil)); err != nil {
			t.Fatal(err)
		}
		if blockAt(t, s.World, 3, -60, 2).Properties["east"] != "false" {
			t.Fatalf("fence incorrectly connects to %s", name)
		}
	}
	if err := s.World.SetBlock(4, -60, 2, world.StateStone); err != nil {
		t.Fatal(err)
	}
	if blockAt(t, s.World, 3, -60, 2).Properties["east"] != "true" {
		t.Fatal("fence lost ordinary sturdy connection")
	}
}
