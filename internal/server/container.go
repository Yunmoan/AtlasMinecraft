package server

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"log/slog"
	"sync"
)

// A single server lock orders chest viewers and player saves. No code under
// this lock acquires entityMu; equipment publication happens after release.
type chestSessions struct {
	mu      sync.Mutex
	players map[*player.Player]*chestSession
	next    map[*player.Player]byte
}
type chestSession struct {
	window   byte
	state    int32
	history  int32 // how many preceding state IDs this menu has actually issued
	world    *world.World
	position world.BlockPosition
	contents world.ChestInventory
	drag     chestDrag
}

func chestOpenPayload(window byte, name string) []byte {
	title := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"translate": {Type: nbt.String, Value: "container.chest"}}}
	if name != "" {
		if parsed, err := item.SavedTextComponent(name); err == nil {
			title = parsed
		}
	}
	return proto.NewWriter().VarInt(int32(window)).VarInt(2).Raw(encodeText(title)).Payload()
}

func (s *Server) openChest(p *player.Player, position world.BlockPosition) {
	s.containers.mu.Lock()
	if s.containers.players == nil {
		s.containers.players = make(map[*player.Player]*chestSession)
		s.containers.next = make(map[*player.Player]byte)
	}
	if !canReachBlock(p, position) {
		s.containers.mu.Unlock()
		return
	}
	w := s.PlayerWorld(p)
	contents, err := w.ReadChest(position, 0)
	if err != nil {
		s.containers.mu.Unlock()
		s.Dbg("chest open rejected: %v", err)
		return
	}
	replaced := false
	if old := s.containers.players[p]; old != nil {
		s.closeChestLocked(p, old, true)
		replaced = true
	}
	next := s.containers.next[p]%100 + 1
	s.containers.next[p] = next
	session := &chestSession{window: next, world: w, position: position, contents: contents}
	s.containers.players[p] = session
	p.ResetInventoryDrag()
	p.SendPacket(proto.SOpenScreen, chestOpenPayload(next, contents.CustomName))
	s.sendChestLocked(p, session)
	s.containers.mu.Unlock()
	if replaced {
		s.publishEquipment(p)
	}
}

func chestMenuPlayerSlot(index int) int {
	if index >= 27 && index < 54 {
		return index - 18
	}
	if index >= 54 && index < 63 {
		return index - 18
	}
	return -1
}
func chestMenuSlots(chest *world.ChestInventory, inventory *item.Inventory) [63]item.Stack {
	var slots [63]item.Stack
	copy(slots[:27], chest.Slots[:])
	for i := 27; i < 63; i++ {
		slots[i] = inventory.Slots[chestMenuPlayerSlot(i)]
	}
	return slots
}
func storeChestMenu(slots [63]item.Stack, chest *world.ChestInventory, inventory *item.Inventory) {
	copy(chest.Slots[:], slots[:27])
	for i := 27; i < 63; i++ {
		inventory.Slots[chestMenuPlayerSlot(i)] = slots[i]
	}
}
func (s *Server) sendChestLocked(p *player.Player, session *chestSession) {
	inventory, _ := p.Inventory()
	slots := chestMenuSlots(&session.contents, &inventory)
	w := proto.NewWriter().U8(session.window).VarInt(session.state).VarInt(63)
	for _, stack := range slots {
		stack.Write(w)
	}
	inventory.Carried.Write(w)
	p.SendPacket(proto.SContainerContent, w.Payload())
}
func (s *Server) chestInRange(p *player.Player, session *chestSession) bool {
	if s.PlayerWorld(p) != session.world {
		return false
	}
	x, y, z, _, _, _ := p.Pos()
	dx, dy, dz := x-float64(session.position.X)-.5, y-float64(session.position.Y)-.5, z-float64(session.position.Z)-.5
	return dx*dx+dy*dy+dz*dz <= 64
}

func (s *Server) saveChestPlayerLocked(p *player.Player, inventory item.Inventory) {
	dimension, x, y, z, yaw, pitch, flying, cooldown := p.WorldPositionWithPortalCooldown()
	if err := s.World.SavePlayerState(p.UUID, world.PlayerPosition{Dimension: dimension, X: x, Y: y, Z: z, Yaw: yaw, Pitch: pitch, Flying: flying, PortalCooldown: cooldown}, inventory); err != nil {
		slog.Error("chest player save failed", "player", p.Name, "error", err)
	}
}

// Returns true whenever an open menu owns this packet, including malformed
// window-zero packets. Client changed stacks are parsed but never applied.
func (s *Server) handleChestClick(p *player.Player, payload []byte) bool {
	s.containers.mu.Lock()
	session := s.containers.players[p]
	if session == nil {
		s.containers.mu.Unlock()
		return false
	}
	defer func() { s.containers.mu.Unlock(); s.publishEquipment(p) }()
	if !s.chestInRange(p, session) {
		s.closeChestLocked(p, session, true)
		return true
	}
	r := proto.NewReader(payload)
	window, e1 := r.U8()
	state, e2 := r.VarInt()
	slot, e3 := r.I16()
	button, e4 := r.I8()
	mode, e5 := r.VarInt()
	changed, e6 := r.VarInt()
	// Vanilla replays stale clicks against its own inventory and resynchronizes
	// afterward. Permit recently issued states for pipelined GUI actions, never
	// a future state or a client-provided item snapshot.
	stateValid := state >= 0 && state <= 32767 && (session.state-state)&32767 <= session.history
	valid := e1 == nil && e2 == nil && e3 == nil && e4 == nil && e5 == nil && e6 == nil && window == session.window && stateValid && changed >= 0 && changed <= 63 && mode >= 0 && mode <= 6
	seen := [63]bool{}
	if valid {
		for i := int32(0); i < changed; i++ {
			index, err := r.I16()
			if err != nil || index < 0 || index >= 63 || seen[index] {
				valid = false
				break
			}
			seen[index] = true
			if _, err = item.Read(r); err != nil {
				valid = false
				break
			}
		}
	}
	if valid {
		_, err := item.Read(r)
		valid = err == nil && r.Remaining() == 0
	}
	if !valid {
		session.drag = chestDrag{}
		s.sendChestLocked(p, session)
		return true
	}
	p.SaveMu.Lock()
	inventory, _ := p.Inventory()
	beforeInventory := inventory.Clone()
	drag := session.drag.clone()
	contents, err := session.world.UpdateChest(session.position, 0, func(chest *world.ChestInventory) error {
		slots := chestMenuSlots(chest, &inventory)
		applyChestAction(&slots, &inventory.Carried, &inventory.Slots[45], &drag, int(slot), int(button), int(mode), true)
		storeChestMenu(slots, chest, &inventory)
		return nil
	})
	if err != nil {
		p.SaveMu.Unlock()
		if err == world.ErrNotSingleChest || err == world.ErrChestBlocked || err == world.ErrChestLocked {
			s.closeChestLocked(p, session, true)
		} else {
			s.sendChestLocked(p, session)
			s.Dbg("chest click rejected: %v", err)
		}
		return true
	}
	inventoryChanged := !chestStackEqual(inventory.Carried, beforeInventory.Carried) || !inventory.Carried.SameSavedFields(beforeInventory.Carried)
	for i, stack := range inventory.Slots {
		inventoryChanged = inventoryChanged || !chestStackEqual(stack, beforeInventory.Slots[i]) || !stack.SameSavedFields(beforeInventory.Slots[i])
	}
	if inventoryChanged {
		s.setInventory(p, inventory)
		s.saveChestPlayerLocked(p, inventory)
	}
	if !chestStackEqual(inventory.Slots[45], beforeInventory.Slots[45]) {
		_, state := p.Inventory()
		p.SendPacket(proto.SContainerSlot, chestOffhandPayload(state, inventory.Slots[45]))
	}
	p.SaveMu.Unlock()
	session.drag = drag
	for viewer, other := range s.containers.players {
		if other.world == session.world && other.position == session.position {
			changed := viewer == p && !chestStackEqual(inventory.Carried, beforeInventory.Carried)
			if viewer == p {
				for i, stack := range inventory.Slots {
					changed = changed || !chestStackEqual(stack, beforeInventory.Slots[i])
				}
			}
			for i, stack := range contents.Slots {
				changed = changed || !chestStackEqual(stack, other.contents.Slots[i])
			}
			other.contents = contents.Clone()
			if changed {
				other.state = (other.state + 1) & 32767
				other.history = min(64, other.history+1)
			}
			s.sendChestLocked(viewer, other)
		}
	}
	return true
}

func chestStackEqual(a, b item.Stack) bool {
	return a.Count == b.Count && (a.Empty() || a.SameItem(b))
}

func chestOffhandPayload(state int32, stack item.Stack) []byte {
	// PLAYER_INVENTORY (-2) targets the underlying Inventory index, not a
	// menu slot. Container0/slot45 is ignored by the vanilla client while a
	// chest menu is open; -2/index40 works without changing its carried item.
	packet := proto.NewWriter().I8(-2).VarInt(state).I16(40)
	stack.Write(packet)
	return packet.Payload()
}

func (s *Server) closeChestLocked(p *player.Player, session *chestSession, notify bool) {
	delete(s.containers.players, p)
	p.SaveMu.Lock()
	inventory, _ := p.Inventory()
	var targets []int
	for i := 9; i <= 44; i++ {
		targets = append(targets, i)
	}
	inventory.Carried = insertInventory(&inventory, inventory.Carried, targets)
	s.setInventory(p, inventory)
	s.saveChestPlayerLocked(p, inventory)
	p.SaveMu.Unlock()
	p.ResetInventoryDrag()
	if notify {
		p.SendPacket(proto.SCloseContainer, []byte{session.window})
	}
	s.sendInventory(p)
}
func (s *Server) closeChest(p *player.Player, window byte, notify bool) bool {
	s.containers.mu.Lock()
	session := s.containers.players[p]
	if session == nil {
		s.containers.mu.Unlock()
		return false
	}
	if window != session.window && window != 255 {
		s.sendChestLocked(p, session)
		s.containers.mu.Unlock()
		return true
	}
	s.closeChestLocked(p, session, notify)
	s.containers.mu.Unlock()
	s.publishEquipment(p)
	return true
}
func (s *Server) forgetChest(p *player.Player) {
	s.closeChest(p, 255, false)
	s.containers.mu.Lock()
	delete(s.containers.next, p)
	s.containers.mu.Unlock()
}
func (s *Server) validateChestSession(p *player.Player) {
	s.containers.mu.Lock()
	closed := false
	session := s.containers.players[p]
	if session != nil {
		var err error
		if s.chestInRange(p, session) {
			_, err = session.world.ReadChest(session.position, 0)
		}
		if !s.chestInRange(p, session) || err != nil {
			s.closeChestLocked(p, session, true)
			closed = true
		}
	}
	s.containers.mu.Unlock()
	if closed {
		s.publishEquipment(p)
	}
}

// GUI sessions own cursor/menu state. Other interaction handlers call this
// before taking MovementMu, entityMu or SaveMu, so no reverse lock is needed.
func (s *Server) rejectChestInteraction(p *player.Player) bool {
	s.containers.mu.Lock()
	defer s.containers.mu.Unlock()
	if session := s.containers.players[p]; session != nil {
		s.sendChestLocked(p, session)
		return true
	}
	return false
}
