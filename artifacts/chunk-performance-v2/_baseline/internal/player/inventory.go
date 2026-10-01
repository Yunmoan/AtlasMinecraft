package player

import "atlas/internal/item"

func (p *Player) Inventory() (item.Inventory, int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.inventory.Clone(), p.inventoryState
}

// Inventory access is serialized with packet processing by the server. The
// snapshot is deep-copied so autosaves cannot race mutable component bytes.
func (p *Player) SetInventory(v item.Inventory, states [9]uint32) int32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.inventory = v.Clone()
	p.heldSlot = v.Selected
	p.hotbar = states
	p.inventoryState = (p.inventoryState + 1) & 0x7fffffff
	return p.inventoryState
}

func (p *Player) ResetInventoryDrag() { p.mu.Lock(); p.dragKind = -1; p.dragSlots = nil; p.mu.Unlock() }
func (p *Player) StartInventoryDrag(kind int) {
	p.mu.Lock()
	p.dragKind = kind
	p.dragSlots = []int{}
	p.mu.Unlock()
}
func (p *Player) AddInventoryDragSlot(kind, slot int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.dragKind != kind || p.dragSlots == nil {
		return
	}
	for _, i := range p.dragSlots {
		if i == slot {
			return
		}
	}
	p.dragSlots = append(p.dragSlots, slot)
}
func (p *Player) FinishInventoryDrag(kind int) []int {
	p.mu.Lock()
	defer p.mu.Unlock()
	var slots []int
	if p.dragKind == kind {
		slots = append(slots, p.dragSlots...)
	}
	p.dragKind = -1
	p.dragSlots = nil
	return slots
}
