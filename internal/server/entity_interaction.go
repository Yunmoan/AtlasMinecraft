package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"log/slog"
	"math"
)

type entityInteraction struct {
	target, action, hand int32
	point                [3]float32
	secondary            bool
}

func readEntityInteraction(payload []byte) (entityInteraction, bool) {
	r := proto.NewReader(payload)
	var v entityInteraction
	var err error
	v.target, err = r.VarInt()
	if err != nil || v.target < 0 {
		return v, false
	}
	v.action, err = r.VarInt()
	if err != nil || v.action < 0 || v.action > 2 {
		return v, false
	}
	if v.action == 2 {
		for i := range v.point {
			v.point[i], err = r.F32()
			if err != nil || math.IsNaN(float64(v.point[i])) || math.IsInf(float64(v.point[i]), 0) {
				return v, false
			}
		}
	}
	if v.action != 1 {
		v.hand, err = r.VarInt()
		if err != nil || v.hand < 0 || v.hand > 1 {
			return v, false
		}
	}
	v.secondary, err = r.Bool()
	return v, err == nil && r.Remaining() == 0
}

func (s *Server) handleEntityInteraction(p *player.Player, payload []byte) {
	click, ok := readEntityInteraction(payload)
	// ArmorStand's equipment action is interactAt; generic interact and combat
	// remain separate mechanics. Never interpret them as an equipment exchange.
	if !ok || click.action != 2 {
		return
	}
	// A current container owns the player's inventory view. Check before
	// SaveMu so container resynchronization follows containers -> SaveMu order.
	if s.rejectChestInteraction(p) {
		return
	}
	p.MovementMu.Lock()
	defer p.MovementMu.Unlock()
	p.SaveMu.Lock()
	defer p.SaveMu.Unlock()
	if !p.PlayReady() || p.TeleportPending() {
		return
	}
	w := s.PlayerWorld(p)
	sink := p.WorldPacketSinkInDimension(w.DimensionName())
	if sink == nil || !sink.Current() {
		return
	}
	var key generatedEntityKey
	var snapshot world.WorldEntity
	found := false
	s.entityMu.Lock()
	for candidate, id := range s.generatedViews[p] {
		if id == click.target && candidate.dimension == w.DimensionName() {
			tracking := s.generatedEntities[candidate]
			if tracking != nil && tracking.snapshot.Type == "minecraft:armor_stand" {
				key, snapshot, found = candidate, tracking.snapshot, true
			}
			break
		}
	}
	s.entityMu.Unlock()
	if !found {
		return
	}
	cx, cz := int32(math.Floor(snapshot.Position[0]/16)), int32(math.Floor(snapshot.Position[2]/16))
	if !sink.IsChunkSent(cx, cz) {
		return
	}
	inventory, _ := p.Inventory()
	slot := 36 + int(inventory.Selected)
	if click.hand == 1 {
		slot = 45
	}
	x, y, z, _, _, _ := p.Pos()
	eye := 1.62
	if p.Crouching() {
		eye = 1.27
	}
	swapped, err := w.InteractLoadedArmorStand(world.ArmorStandClick{CX: cx, CZ: cz, UUID: key.uuid, Held: inventory.Slots[slot], Point: click.point, Eye: [3]float64{x, y + eye, z}, Creative: true})
	if err != nil {
		s.Dbg("armor stand interaction rejected: player=%s error=%v", p.Name, err)
		s.sendInventory(p)
		return
	}
	if !swapped.Changed {
		return
	}
	inventory.Slots[slot] = swapped.Held
	s.setInventory(p, inventory)
	dimension, px, py, pz, yaw, pitch, flying, cooldown := p.WorldPositionWithPortalCooldown()
	position := world.PlayerPosition{Dimension: dimension, X: px, Y: py, Z: pz, Yaw: yaw, Pitch: pitch, Flying: flying, PortalCooldown: cooldown}
	if err := s.World.SavePlayerState(p.UUID, position, inventory); err != nil {
		// Keep the committed authoritative exchange in memory, and let the normal
		// autosave retry its player write. Entity and player files are independent;
		// crash-atomic transactions between these two files are not yet provided.
		slog.Error("armor stand player save failed", "player", p.Name, "error", err)
	}
	s.sendInventory(p)
	s.publishEquipment(p)
	s.publishGeneratedEntityEquipment(key, swapped.Entity)
}

func (s *Server) publishGeneratedEntityEquipment(key generatedEntityKey, entity world.WorldEntity) {
	s.entityMu.Lock()
	defer s.entityMu.Unlock()
	tracking := s.generatedEntities[key]
	if tracking == nil || tracking.snapshot.Revision > entity.Revision {
		return
	}
	tracking.snapshot = entity
	payload, err := generatedEntityEquipmentSnapshot(tracking.id, entity, true)
	if err != nil {
		s.Logf("generated entity equipment failed: %v", err)
		return
	}
	for p, view := range s.generatedViews {
		if id, visible := view[key]; visible && id == tracking.id && p.PlayReady() {
			sink := p.WorldPacketSinkInDimension(key.dimension)
			cx, cz := int32(math.Floor(entity.Position[0]/16)), int32(math.Floor(entity.Position[2]/16))
			if sink != nil && sink.IsChunkSent(cx, cz) {
				sink.SendPacket(proto.SSetEquipment, payload)
			}
		}
	}
}
