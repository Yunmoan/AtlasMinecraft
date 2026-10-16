package world

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

//go:embed armor_stand_item_slots_1.21.1.txt
var armorStandItemSlotData string

// Equipment ordinals are the original protocol slots. DisabledSlots uses a
// different order: offhand is bit 5, and the armor slots use bits 1..4.
var armorStandDisabledBits = [6]uint{0, 5, 1, 2, 3, 4}
var armorStandItemSlots = func() map[int32]int {
	result := map[int32]int{}
	for _, line := range strings.Split(armorStandItemSlotData, "\n") {
		f := strings.Fields(line)
		if len(f) != 2 {
			continue
		}
		id, e1 := strconv.ParseInt(f[0], 10, 32)
		slot, e2 := strconv.Atoi(f[1])
		if e1 != nil || e2 != nil || slot < 0 || slot > 6 {
			panic("invalid official armor stand item slot")
		}
		result[int32(id)] = slot
	}
	return result
}()

func ArmorStandItemSlot(stack item.Stack) int { return armorStandItemSlots[stack.ID] }

func armorStandEquipment(data nbt.Object) ([6]item.Stack, error) {
	var slots [6]item.Stack
	for _, field := range []struct {
		name          string
		offset, limit int
	}{{"HandItems", 0, 2}, {"ArmorItems", 2, 4}} {
		values, _ := data[field.name].Elements()
		for i, tag := range values {
			if i >= field.limit {
				break
			}
			o, ok := tag.Object()
			if !ok {
				return slots, fmt.Errorf("invalid %s equipment", field.name)
			}
			id, _ := o.Text("id")
			if len(o) == 0 || id == "minecraft:air" {
				continue
			}
			s, err := item.FromNBT(tag)
			if err != nil {
				return slots, err
			}
			slots[field.offset+i] = s
		}
	}
	return slots, nil
}

func setArmorStandSlot(data nbt.Object, slot int, stack item.Stack) error {
	field, offset, size := "HandItems", 0, 2
	if slot >= 2 {
		field, offset, size = "ArmorItems", 2, 4
	}
	old, _ := data[field].Elements()
	values := make([]nbt.Tag, size)
	for i := range values {
		values[i] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
		if i < len(old) {
			values[i] = old[i]
		}
	}
	tag := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
	if !stack.Empty() {
		var err error
		tag, err = stack.ToNBT()
		if err != nil {
			return err
		}
	}
	values[slot-offset] = tag
	data[field] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: values}
	return nil
}

// ArmorStandInteraction applies the actual ArmorStand.interactAt selection and
// swap rules to a caller-owned draft. Unknown NBT and item components survive.
// Normal armor stands have no generic.scale attribute in the original 1.21.1
// attribute supplier; the only age scale is Small (.5).
func ArmorStandInteraction(data nbt.Object, held item.Stack, y float64, creative, spectator bool) (item.Stack, string, bool, error) {
	marker, _ := data.Number("Marker")
	name, _ := item.Name(held.ID)
	if marker != 0 || !held.Empty() && name == "minecraft:name_tag" {
		return held, "PASS", false, nil
	}
	if spectator {
		return held, "SUCCESS", false, nil
	}
	slots, err := armorStandEquipment(data)
	if err != nil {
		return held, "PASS", false, err
	}
	disabled, _ := data.Number("DisabledSlots")
	arms, _ := data.Number("ShowArms")
	isDisabled := func(slot int) bool {
		return slot < 0 || slot >= 6 || uint32(disabled)&(1<<armorStandDisabledBits[slot]) != 0 || slot < 2 && arms == 0
	}
	slot := ArmorStandItemSlot(held)
	if held.Empty() {
		clicked := 0
		small, _ := data.Number("Small")
		if small != 0 {
			y /= .5
		}
		feetWindow, chestOffset, chestWindow, legsWindow := .45, 0.0, .7, .8
		if small != 0 {
			feetWindow, chestOffset, chestWindow, legsWindow = .8, .3, 1.0, 1.0
		}
		// Preserve Java's runtime additions: .4 + .8 rounds above 1.2,
		// so the exact 1.2 boundary can still select equipped leggings.
		feetMax, chestMin, chestMax, legsMax := .1+feetWindow, .9+chestOffset, .9+chestWindow, .4+legsWindow
		switch {
		case y >= .1 && y < feetMax && !slots[2].Empty():
			clicked = 2
		case y >= chestMin && y < chestMax && !slots[4].Empty():
			clicked = 4
		case y >= .4 && y < legsMax && !slots[3].Empty():
			clicked = 3
		case y >= 1.6 && !slots[5].Empty():
			clicked = 5
		case slots[0].Empty() && !slots[1].Empty():
			clicked = 1
		}
		if !isDisabled(clicked) {
			slot = clicked
		}
		if slot >= 6 || slots[slot].Empty() {
			return held, "PASS", false, nil
		}
	} else if isDisabled(slot) {
		return held, "FAIL", false, nil
	}
	old := slots[slot]
	bit := armorStandDisabledBits[slot]
	if !old.Empty() && uint32(disabled)&(1<<(bit+8)) != 0 || old.Empty() && uint32(disabled)&(1<<(bit+16)) != 0 {
		return held, "PASS", false, nil
	}
	next := held.Clone()
	switch {
	case creative && old.Empty() && !held.Empty():
		next = held.WithCount(1)
	case !held.Empty() && held.Count > 1:
		if !old.Empty() {
			return held, "PASS", false, nil
		}
		next = held.WithCount(1)
		held = held.WithCount(held.Count - 1)
	default:
		held = old.Clone()
	}
	if err := setArmorStandSlot(data, slot, next); err != nil {
		return held, "PASS", false, err
	}
	return held, "SUCCESS", true, nil
}

var ErrEntityNotLoaded = errors.New("entity is not in the loaded chunk")
var ErrEntityInteractionRange = errors.New("entity is outside interaction range")

type ArmorStandClick struct {
	CX, CZ   int32
	UUID     [16]byte
	Held     item.Stack
	Point    [3]float32
	Eye      [3]float64
	Creative bool
}
type ArmorStandSwap struct {
	Entity  WorldEntity
	Held    item.Stack
	Result  string
	Changed bool
}
type armorStandQuery struct {
	click  ArmorStandClick
	result chan armorStandReply
}
type armorStandReply struct {
	swap ArmorStandSwap
	err  error
}

// InteractLoadedArmorStand never generates a chunk for a client-supplied ID.
// The owning actor selects from its current equipment and writes Anvil before
// publishing the new snapshot. Concurrent players cannot take the same item.
func (w *World) InteractLoadedArmorStand(click ArmorStandClick) (ArmorStandSwap, error) {
	click.Held = click.Held.Clone()
	reply := make(chan armorStandReply, 1)
	if err := w.post(context.Background(), click.CX, click.CZ, armorStandQuery{click, reply}); err != nil {
		return ArmorStandSwap{}, err
	}
	select {
	case res := <-reply:
		return res.swap, res.err
	case <-w.done:
		return ArmorStandSwap{}, ErrWorldClosed
	}
}

func (r *Region) interactArmorStand(query armorStandQuery) {
	swap, err := r.armorStandSwap(query.click)
	query.result <- armorStandReply{swap, err}
}

func (r *Region) armorStandSwap(click ArmorStandClick) (ArmorStandSwap, error) {
	ch := r.chunks[chunkKey{click.CX, click.CZ}]
	if ch == nil {
		return ArmorStandSwap{}, ErrChunkNotLoaded
	}
	for i, e := range ch.entities {
		if e.UUID != click.UUID || e.Type != "minecraft:armor_stand" {
			continue
		}
		small, _ := e.Data.Number("Small")
		half, height := .25, float64(float32(1.975))
		if small != 0 {
			half, height = half*.5, height*.5
		}
		marker, _ := e.Data.Number("Marker")
		if marker != 0 {
			return ArmorStandSwap{Held: click.Held, Result: "PASS"}, nil
		}
		var distance float64
		for axis, value := range click.Eye {
			minimum, maximum := e.Position[axis]-half, e.Position[axis]+half
			if axis == 1 {
				minimum, maximum = e.Position[axis], e.Position[axis]+height
			}
			delta := max(minimum-value, max(0, value-maximum))
			distance += delta * delta
		}
		// Original creative entity range 5 + packet tolerance 1.
		if math.IsNaN(distance) || math.IsInf(distance, 0) || distance >= 36 {
			return ArmorStandSwap{}, ErrEntityInteractionRange
		}
		for _, v := range click.Point {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return ArmorStandSwap{}, ErrEntityInteractionRange
			}
		}
		if math.Abs(float64(click.Point[0])) > half+.05 || math.Abs(float64(click.Point[2])) > half+.05 || float64(click.Point[1]) < -.05 || float64(click.Point[1]) > height+.05 {
			return ArmorStandSwap{}, ErrEntityInteractionRange
		}
		data, err := cloneEntityData(e.Data)
		if err != nil {
			return ArmorStandSwap{}, err
		}
		held, result, changed, err := ArmorStandInteraction(data, click.Held, float64(click.Point[1]), click.Creative, false)
		if err != nil {
			return ArmorStandSwap{}, err
		}
		if changed {
			if r.world.store == nil {
				return ArmorStandSwap{}, ErrVanillaGenerationUnavailable
			}
			next := e
			next.Data = data
			next.Revision++
			name, root, _, err := r.world.store.vanilla.readEntityChunk(ch.cx, ch.cz)
			if err != nil {
				return ArmorStandSwap{}, err
			}
			list := make([]nbt.Tag, len(ch.entities))
			for j, entity := range ch.entities {
				if i == j {
					entity = next
				}
				list[j] = nbt.Tag{Type: nbt.Compound, Value: entity.Data}
			}
			root["Entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: list}
			if err := r.world.store.vanilla.regions.WriteChunk(r.world.store.vanilla.entityDirectory(), ch.cx, ch.cz, name, nbt.Tag{Type: nbt.Compound, Value: root}); err != nil {
				return ArmorStandSwap{}, err
			}
			ch.entities[i] = next
			e = next
		}
		copy, err := cloneEntityData(e.Data)
		if err != nil {
			return ArmorStandSwap{}, err
		}
		e.Data = copy
		ch.lastAccess = time.Now()
		return ArmorStandSwap{e, held, result, changed}, nil
	}
	return ArmorStandSwap{}, ErrEntityNotLoaded
}
