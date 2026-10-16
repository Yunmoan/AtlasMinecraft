package server

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/world"
	"fmt"
)

// Generated actors retain their authored/finalized equipment in saved NBT.
// Sending this snapshot after spawn also renders a fox's held item and the
// weapons and armor carried by structure actors.
func generatedEntityEquipment(id int32, entity world.WorldEntity) ([]byte, error) {
	return generatedEntityEquipmentSnapshot(id, entity, false)
}

// An equipment removal must explicitly send empty slots, including when the
// final item is taken; spawn-only encoding may omit an all-empty snapshot.
func generatedEntityEquipmentSnapshot(id int32, entity world.WorldEntity, force bool) ([]byte, error) {
	var slots [7]item.Stack
	populated := false
	read := func(index int, tag nbt.Tag) error {
		o, ok := tag.Object()
		if !ok || len(o) == 0 {
			return nil
		}
		name, _ := o.Text("id")
		if name == "" || name == "minecraft:air" {
			return nil
		}
		stack, err := item.FromNBT(tag)
		if err != nil {
			return fmt.Errorf("generated %s equipment slot%d: %w", entity.Type, index, err)
		}
		slots[index] = stack
		populated = populated || !stack.Empty()
		return nil
	}
	for _, field := range []struct {
		name   string
		offset int
	}{{"HandItems", 0}, {"ArmorItems", 2}} {
		values, _ := entity.Data[field.name].Elements()
		for i, tag := range values {
			if field.offset+i >= 6 {
				break
			}
			if err := read(field.offset+i, tag); err != nil {
				return nil, err
			}
		}
	}
	if err := read(6, entity.Data["body_armor_item"]); err != nil {
		return nil, err
	}
	if !populated && !force {
		return nil, nil
	}
	w := proto.NewWriter().VarInt(id)
	for index, stack := range slots {
		slot := byte(index)
		if index < len(slots)-1 {
			slot |= 0x80
		}
		w.U8(slot)
		stack.Write(w)
	}
	return w.Payload(), nil
}
