package server

import "atlas/internal/item"

type chestDrag struct {
	active bool
	kind   int
	slots  []int
}

func (v chestDrag) clone() chestDrag { v.slots = append([]int(nil), v.slots...); return v }
func chestInsert(slots *[63]item.Stack, stack item.Stack, first, last, step int) item.Stack {
	for phase := 0; phase < 2; phase++ {
		for i := first; i != last; i += step {
			existing := slots[i]
			if phase == 0 && !existing.SameItem(stack) || phase == 1 && !existing.Empty() {
				continue
			}
			count := min(stack.Count, stack.MaxCount()-existing.Count)
			if count > 0 {
				slots[i] = stack.WithCount(existing.Count + count)
				stack = stack.WithCount(stack.Count - count)
			}
			if stack.Empty() {
				return stack
			}
		}
	}
	return stack
}

// These generic-menu operations preserve complete component patches. Client
// predictions never enter this function. Throw/outside drop requests retain
// items until item-entity drops have an authoritative runtime implementation.
func applyChestAction(slots *[63]item.Stack, carried, offhand *item.Stack, drag *chestDrag, slot, button, mode int, creative bool) {
	if mode == 5 {
		applyChestDrag(slots, carried, offhand, drag, slot, button, creative)
		return
	}
	if drag.active {
		*drag = chestDrag{}
		return
	}
	if slot < 0 || slot >= 63 {
		return
	}
	stack := slots[slot]
	switch mode {
	case 0:
		if button < 0 || button > 1 {
			return
		}
		if carried.Empty() {
			count := stack.Count
			if button == 1 {
				count = (count + 1) / 2
			}
			*carried = stack.WithCount(count)
			slots[slot] = stack.WithCount(stack.Count - count)
		} else if stack.Empty() || stack.SameItem(*carried) {
			count := min(carried.Count, carried.MaxCount()-stack.Count)
			if button == 1 {
				count = min(count, 1)
			}
			if count > 0 {
				slots[slot] = carried.WithCount(stack.Count + count)
				*carried = carried.WithCount(carried.Count - count)
			}
		} else if carried.Count <= carried.MaxCount() {
			slots[slot], *carried = carried.Clone(), stack.Clone()
		}
	case 1:
		if button < 0 || button > 1 || stack.Empty() {
			return
		}
		if slot < 27 {
			slots[slot] = chestInsert(slots, stack, 62, 26, -1)
		} else {
			slots[slot] = chestInsert(slots, stack, 0, 27, 1)
		}
	case 2:
		var target *item.Stack
		if button == 40 {
			target = offhand
		} else if button >= 0 && button <= 8 {
			target = &slots[54+button]
		} else {
			return
		}
		// Generic chest/main/hotbar slots have the same limit for a stack.
		if target == &slots[slot] {
			return
		}
		slots[slot], *target = target.Clone(), stack.Clone()
	case 3:
		if creative && carried.Empty() && !stack.Empty() {
			*carried = stack.WithCount(stack.MaxCount())
		}
	case 6:
		if carried.Empty() || !stack.Empty() {
			return
		}
		first, last, step := 0, 63, 1
		if button != 0 {
			first, last, step = 62, -1, -1
		}
		for pass := 0; pass < 2; pass++ {
			for i := first; i != last && carried.Count < carried.MaxCount(); i += step {
				candidate := slots[i]
				if !candidate.SameItem(*carried) || pass == 0 && candidate.Count == candidate.MaxCount() {
					continue
				}
				count := min(candidate.Count, carried.MaxCount()-carried.Count)
				*carried = carried.WithCount(carried.Count + count)
				slots[i] = candidate.WithCount(candidate.Count - count)
			}
		}
	}
}
func applyChestDrag(slots *[63]item.Stack, carried, offhand *item.Stack, drag *chestDrag, slot, button int, creative bool) {
	phase, kind := button&3, (button>>2)&3
	if kind > 2 || kind == 2 && !creative || carried.Empty() {
		*drag = chestDrag{}
		return
	}
	switch phase {
	case 0:
		if drag.active {
			*drag = chestDrag{}
			return
		}
		*drag = chestDrag{active: true, kind: kind}
	case 1:
		if !drag.active || kind != drag.kind || slot < 0 || slot >= 63 {
			*drag = chestDrag{}
			return
		}
		existing := slots[slot]
		if !existing.Empty() && !existing.SameItem(*carried) || existing.Count > carried.MaxCount() || kind != 2 && carried.Count <= int32(len(drag.slots)) {
			return
		}
		for _, i := range drag.slots {
			if i == slot {
				return
			}
		}
		drag.slots = append(drag.slots, slot)
	case 2:
		if !drag.active || kind != drag.kind {
			*drag = chestDrag{}
			return
		}
		indices := drag.slots
		*drag = chestDrag{}
		if len(indices) == 1 {
			applyChestAction(slots, carried, offhand, drag, indices[0], kind, 0, creative)
			return
		}
		if len(indices) == 0 {
			return
		}
		original := carried.Clone()
		remaining := carried.Count
		per := original.Count / int32(len(indices))
		if kind == 1 {
			per = 1
		} else if kind == 2 {
			per = original.MaxCount()
		}
		for _, i := range indices {
			existing := slots[i]
			if !existing.Empty() && !existing.SameItem(original) || existing.Count > original.MaxCount() || kind != 2 && original.Count < int32(len(indices)) {
				continue
			}
			count := min(existing.Count+per, original.MaxCount())
			remaining -= count - existing.Count
			slots[i] = original.WithCount(count)
		}
		*carried = original.WithCount(remaining)
	default:
		*drag = chestDrag{}
	}
}
