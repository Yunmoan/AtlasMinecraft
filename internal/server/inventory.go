package server

import (
	"atlas/internal/item"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"strings"
)

func (s *Server) setInventory(p *player.Player, v item.Inventory) {
	var states [9]uint32
	for i := range states {
		if !v.Slots[36+i].Empty() {
			states[i] = world.ItemBlock[v.Slots[36+i].ID]
		}
	}
	p.SetInventory(v, states)
}
func inventoryPayload(p *player.Player) []byte {
	v, state := p.Inventory()
	w := proto.NewWriter().U8(0).VarInt(state).VarInt(46)
	for _, stack := range v.Slots {
		stack.Write(w)
	}
	v.Carried.Write(w)
	return w.Payload()
}
func (s *Server) sendInventory(p *player.Player) {
	p.SendPacket(proto.SContainerContent, inventoryPayload(p))
}
func equipmentPayload(p *player.Player) []byte {
	v, _ := p.Inventory()
	w := proto.NewWriter().VarInt(p.EID)
	for index, slot := range []int{36 + int(v.Selected), 45, 8, 7, 6, 5} {
		wireIndex := byte(index)
		if index < 5 {
			wireIndex |= 0x80
		}
		w.U8(wireIndex)
		v.Slots[slot].Write(w)
	}
	return w.Payload()
}
func (s *Server) publishEquipment(p *player.Player) {
	s.entityMu.Lock()
	defer s.entityMu.Unlock()
	s.sendTracked(p, proto.SSetEquipment, equipmentPayload(p))
}
func (s *Server) handleCreativeInventorySlot(p *player.Player, payload []byte) {
	if s.rejectChestInteraction(p) {
		return
	}
	r := proto.NewReader(payload)
	slot, err := r.I16()
	if err != nil {
		return
	}
	p.SaveMu.Lock()
	v, _ := p.Inventory()
	stack, err := item.Read(r)
	if err != nil || r.Remaining() != 0 || slot < 1 || slot >= 46 {
		// Keep authoritative items, correcting any client prediction.
		s.setInventory(p, v)
		s.sendInventory(p)
		p.SaveMu.Unlock()
		return
	}
	v.Slots[slot] = stack
	s.setInventory(p, v)
	// Creative inventory screens also predict edits; echo an authoritative
	// slot, not a full cursor reset on every creative catalog operation.
	_, state := p.Inventory()
	w := proto.NewWriter().U8(0).VarInt(state).I16(slot)
	stack.Write(w)
	p.SendPacket(proto.SContainerSlot, w.Payload())
	p.SaveMu.Unlock()
	s.publishEquipment(p)
}

func armorSlot(stack item.Stack) int {
	name, _ := item.Name(stack.ID)
	switch {
	case strings.HasSuffix(name, "_helmet") || name == "minecraft:turtle_helmet" || name == "minecraft:carved_pumpkin" || strings.HasSuffix(name, "_head") || strings.HasSuffix(name, "_skull"):
		return 5
	case strings.HasSuffix(name, "_chestplate") || name == "minecraft:elytra":
		return 6
	case strings.HasSuffix(name, "_leggings"):
		return 7
	case strings.HasSuffix(name, "_boots"):
		return 8
	}
	return -1
}
func slotAccepts(slot int, stack item.Stack) bool {
	if slot <= 0 || slot >= 46 {
		return false
	}
	return stack.Empty() || slot < 5 || slot > 8 || armorSlot(stack) == slot && stack.Count == 1
}
func slotLimit(slot int, stack item.Stack) int32 {
	if slot >= 5 && slot <= 8 {
		return 1
	}
	return stack.MaxCount()
}

// Never trust the client's changed-slot contents as inventory authority.
// Replay the menu action against our own state and send its resulting snapshot.
func (s *Server) handleContainerClick(p *player.Player, payload []byte) {
	if s.handleChestClick(p, payload) {
		return
	}
	r := proto.NewReader(payload)
	window, e1 := r.U8()
	_, e2 := r.VarInt()
	slot, e3 := r.I16()
	button, e4 := r.I8()
	mode, e5 := r.VarInt()
	changed, e6 := r.VarInt()
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || e6 != nil || window != 0 || changed < 0 || changed > 46 {
		s.sendInventory(p)
		return
	}
	seen := map[int16]bool{}
	for range changed {
		index, err := r.I16()
		if err != nil || index < 0 || index >= 46 || seen[index] {
			s.sendInventory(p)
			return
		}
		seen[index] = true
		if _, err = item.Read(r); err != nil {
			s.sendInventory(p)
			return
		}
	}
	if _, err := item.Read(r); err != nil || r.Remaining() != 0 {
		s.sendInventory(p)
		return
	}
	p.SaveMu.Lock()
	v, _ := p.Inventory()
	if mode == 5 {
		s.handleInventoryDrag(p, &v, int(slot), int(button))
	} else {
		p.ResetInventoryDrag()
		applyInventoryClick(&v, int(slot), int(button), int(mode))
	}
	s.setInventory(p, v)
	s.sendInventory(p)
	p.SaveMu.Unlock()
	s.publishEquipment(p)
}

func applyInventoryClick(v *item.Inventory, slot, button, mode int) {
	if slot < 1 || slot >= 46 {
		return
	}
	stack := v.Slots[slot]
	switch mode {
	case 0:
		if button < 0 || button > 1 {
			return
		}
		if v.Carried.Empty() {
			count := stack.Count
			if button == 1 {
				count = (count + 1) / 2
			}
			v.Carried = stack.WithCount(count)
			v.Slots[slot] = stack.WithCount(stack.Count - count)
			return
		}
		if !slotAccepts(slot, v.Carried.WithCount(1)) {
			return
		}
		if stack.Empty() || stack.SameItem(v.Carried) {
			count := min(v.Carried.Count, slotLimit(slot, v.Carried)-stack.Count)
			if button == 1 {
				count = min(count, 1)
			}
			if count > 0 {
				v.Slots[slot] = v.Carried.WithCount(stack.Count + count)
				v.Carried = v.Carried.WithCount(v.Carried.Count - count)
			}
			return
		}
		if slotAccepts(slot, v.Carried) && v.Carried.Count <= slotLimit(slot, v.Carried) {
			v.Slots[slot], v.Carried = v.Carried, stack
		}
	case 1:
		if stack.Empty() {
			return
		}
		if armor := armorSlot(stack); armor >= 0 && slot >= 9 && v.Slots[armor].Empty() {
			v.Slots[armor] = stack.WithCount(1)
			v.Slots[slot] = stack.WithCount(stack.Count - 1)
			return
		}
		var targets []int
		if slot >= 9 && slot <= 35 {
			for i := 36; i <= 44; i++ {
				targets = append(targets, i)
			}
		} else if slot >= 36 && slot <= 44 {
			for i := 9; i <= 35; i++ {
				targets = append(targets, i)
			}
		} else {
			for i := 9; i <= 44; i++ {
				targets = append(targets, i)
			}
		}
		v.Slots[slot] = insertInventory(v, stack, targets)
	case 2:
		target := 36 + button
		if button == 40 {
			target = 45
		} else if button < 0 || button > 8 {
			return
		}
		if target == slot || !slotAccepts(slot, v.Slots[target]) || !slotAccepts(target, stack) {
			return
		}
		v.Slots[slot], v.Slots[target] = v.Slots[target], stack
	case 3:
		if button == 2 && v.Carried.Empty() && !stack.Empty() {
			v.Carried = stack.WithCount(stack.MaxCount())
		}
	case 6:
		if v.Carried.Empty() {
			return
		}
		for i := 1; i < 46 && v.Carried.Count < v.Carried.MaxCount(); i++ {
			stack := v.Slots[i]
			if !stack.SameItem(v.Carried) {
				continue
			}
			count := min(stack.Count, v.Carried.MaxCount()-v.Carried.Count)
			v.Carried = v.Carried.WithCount(v.Carried.Count + count)
			v.Slots[i] = stack.WithCount(stack.Count - count)
		}
	}
}
func insertInventory(v *item.Inventory, stack item.Stack, targets []int) item.Stack {
	for _, index := range targets {
		existing := v.Slots[index]
		if !existing.SameItem(stack) {
			continue
		}
		n := min(stack.Count, existing.MaxCount()-existing.Count)
		if n > 0 {
			v.Slots[index] = existing.WithCount(existing.Count + n)
			stack = stack.WithCount(stack.Count - n)
		}
		if stack.Empty() {
			return stack
		}
	}
	for _, index := range targets {
		if !v.Slots[index].Empty() {
			continue
		}
		n := min(stack.Count, stack.MaxCount())
		v.Slots[index] = stack.WithCount(n)
		stack = stack.WithCount(stack.Count - n)
		if stack.Empty() {
			return stack
		}
	}
	return stack
}
func (s *Server) handleInventoryDrag(p *player.Player, v *item.Inventory, slot, button int) {
	phase, kind := button&3, (button>>2)&3
	if kind > 2 {
		p.ResetInventoryDrag()
		return
	}
	switch phase {
	case 0:
		p.StartInventoryDrag(kind)
	case 1:
		if slotAccepts(slot, v.Carried.WithCount(1)) && !v.Carried.Empty() && (v.Slots[slot].Empty() || v.Slots[slot].SameItem(v.Carried)) {
			p.AddInventoryDragSlot(kind, slot)
		}
	case 2:
		slots := p.FinishInventoryDrag(kind)
		if len(slots) == 0 || v.Carried.Empty() {
			return
		}
		per := v.Carried.Count / int32(len(slots))
		if kind == 1 {
			per = 1
		}
		for _, index := range slots {
			existing := v.Slots[index]
			if !slotAccepts(index, v.Carried.WithCount(1)) || !existing.Empty() && !existing.SameItem(v.Carried) {
				continue
			}
			count := min(per, slotLimit(index, v.Carried)-existing.Count)
			if kind == 2 {
				count = slotLimit(index, v.Carried) - existing.Count
			} else {
				count = min(count, v.Carried.Count)
			}
			if count <= 0 {
				continue
			}
			v.Slots[index] = v.Carried.WithCount(existing.Count + count)
			if kind != 2 {
				v.Carried = v.Carried.WithCount(v.Carried.Count - count)
			}
		}
	default:
		p.ResetInventoryDrag()
	}
}
func (s *Server) closeInventory(p *player.Player) {
	p.SaveMu.Lock()
	v, _ := p.Inventory()
	var targets []int
	for i := 9; i <= 44; i++ {
		targets = append(targets, i)
	}
	v.Carried = insertInventory(&v, v.Carried, targets)
	for i := 1; i <= 4; i++ {
		v.Slots[i] = insertInventory(&v, v.Slots[i], targets)
	}
	p.ResetInventoryDrag()
	s.setInventory(p, v)
	s.sendInventory(p)
	p.SaveMu.Unlock()
	s.publishEquipment(p)
}
