package item

import (
	"atlas/internal/nbt"
	"errors"
	"fmt"
)

// Slots use the player-menu layout: crafting 0..4, armor 5..8, main 9..35,
// hotbar 36..44 and offhand 45. Selected is a hotbar index, not a menu slot.
type Inventory struct {
	Slots    [46]Stack
	Carried  Stack
	Selected int32
}

func (v Inventory) Clone() Inventory {
	for i := range v.Slots {
		v.Slots[i] = v.Slots[i].Clone()
	}
	v.Carried = v.Carried.Clone()
	return v
}
func menuSlot(saved int) (int, bool) {
	switch {
	case saved >= 0 && saved <= 8:
		return saved + 36, true
	case saved >= 9 && saved <= 35:
		return saved, true
	case saved >= 100 && saved <= 103:
		return 108 - saved, true
	case saved == 150:
		return 45, true
	}
	return 0, false
}
func saveSlot(menu int) (int, bool) {
	switch {
	case menu >= 36 && menu <= 44:
		return menu - 36, true
	case menu >= 9 && menu <= 35:
		return menu, true
	case menu >= 5 && menu <= 8:
		return 108 - menu, true
	case menu == 45:
		return 150, true
	}
	return 0, false
}
func InventoryFromNBT(o nbt.Object) (Inventory, error) {
	var v Inventory
	if selected, ok := o.Number("SelectedItemSlot"); ok {
		if selected < 0 || selected > 8 {
			return v, errors.New("invalid selected item slot")
		}
		v.Selected = int32(selected)
	}
	if tag, exists := o["Inventory"]; exists {
		entries, ok := tag.Elements()
		if !ok {
			return v, errors.New("invalid Inventory list")
		}
		seen := map[int]bool{}
		for _, tag := range entries {
			entry, ok := tag.Object()
			if !ok {
				return v, errors.New("invalid inventory entry")
			}
			slot, ok := entry.Number("Slot")
			if !ok {
				return v, errors.New("missing inventory slot")
			}
			index, ok := menuSlot(int(uint8(slot)))
			if !ok {
				return v, fmt.Errorf("unknown player inventory slot %d", slot)
			}
			if seen[index] {
				return v, errors.New("duplicate player inventory slot")
			}
			seen[index] = true
			stack, err := FromNBT(tag)
			if err != nil {
				return v, fmt.Errorf("inventory slot %d: %w", slot, err)
			}
			v.Slots[index] = stack
		}
	}
	if tag, exists := o["AtlasCarriedItem"]; exists {
		stack, err := FromNBT(tag)
		if err != nil {
			return v, err
		}
		v.Carried = stack
	}
	if tag, exists := o["AtlasCraftingItems"]; exists {
		entries, ok := tag.Elements()
		if !ok {
			return v, errors.New("invalid saved crafting items")
		}
		for _, tag := range entries {
			entry, ok := tag.Object()
			if !ok {
				return v, errors.New("invalid crafting entry")
			}
			slot, ok := entry.Number("Slot")
			if !ok || slot < 1 || slot > 4 {
				return v, errors.New("invalid crafting slot")
			}
			stack, err := FromNBT(tag)
			if err != nil {
				return v, err
			}
			v.Slots[slot] = stack
		}
	}
	return v, nil
}
func (v Inventory) SaveNBT(o nbt.Object) error {
	entries := []nbt.Tag{}
	crafting := []nbt.Tag{}
	for menu, s := range v.Slots {
		if s.Empty() || menu == 0 {
			continue
		}
		tag, err := s.ToNBT()
		if err != nil {
			return err
		}
		entry, _ := tag.Object()
		if slot, ok := saveSlot(menu); ok {
			entry["Slot"] = nbt.Tag{Type: nbt.Byte, Value: int8(uint8(slot))}
			entries = append(entries, tag)
		} else {
			entry["Slot"] = nbt.Tag{Type: nbt.Byte, Value: int8(menu)}
			crafting = append(crafting, tag)
		}
	}
	o["Inventory"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: entries}
	o["SelectedItemSlot"] = integer(v.Selected)
	delete(o, "AtlasCarriedItem")
	delete(o, "AtlasCraftingItems")
	if !v.Carried.Empty() {
		tag, err := v.Carried.ToNBT()
		if err != nil {
			return err
		}
		o["AtlasCarriedItem"] = tag
	}
	if len(crafting) != 0 {
		o["AtlasCraftingItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: crafting}
	}
	return nil
}
