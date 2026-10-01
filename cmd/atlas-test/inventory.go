package main

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"bytes"
	"time"
)

func savedInventorySample() map[int16]item.Stack {
	stone, _ := item.ID("stone")
	sword, _ := item.ID("diamond_sword")
	helmet, _ := item.ID("diamond_helmet")
	shield, _ := item.ID("shield")
	potion, _ := item.ID("potion")
	apple, _ := item.ID("apple")
	name, _ := nbt.MarshalNetwork(nbt.Tag{Type: nbt.String, Value: "Atlas inventory probe"})
	return map[int16]item.Stack{
		9:  {ID: sword, Count: 1, Patch: proto.NewWriter().VarInt(2).VarInt(0).VarInt(3).VarInt(13).VarInt(5).Raw(name).Payload()},
		10: {ID: potion, Count: 1, Patch: proto.NewWriter().VarInt(1).VarInt(0).VarInt(31).Bool(false).Bool(true).I32(0x123456).VarInt(1).VarInt(0).VarInt(2).VarInt(600).Bool(true).Bool(false).Bool(true).Bool(false).Payload()},
		11: {ID: apple, Count: 3, Patch: proto.NewWriter().VarInt(1).VarInt(0).VarInt(20).VarInt(4).F32(.3).Bool(true).F32(.8).Bool(false).VarInt(1).VarInt(0).VarInt(0).VarInt(200).Bool(false).Bool(true).Bool(true).Bool(false).F32(.5).Payload()},
		5:  item.New(helmet, 1), 45: item.New(shield, 1), 38: item.New(stone, 31),
	}
}
func readInventoryContent(payload []byte) (item.Inventory, bool) {
	r := proto.NewReader(payload)
	window, e1 := r.U8()
	_, e2 := r.VarInt()
	count, e3 := r.VarInt()
	var v item.Inventory
	if e1 != nil || e2 != nil || e3 != nil || window != 0 || count != 46 {
		return v, false
	}
	for i := range v.Slots {
		stack, err := item.Read(r)
		if err != nil {
			return v, false
		}
		v.Slots[i] = stack
	}
	stack, err := item.Read(r)
	v.Carried = stack
	return v, err == nil && r.Remaining() == 0
}
func checkSavedInventory(v item.Inventory, selected int8) {
	for slot, stack := range savedInventorySample() {
		check("saved inventory slot "+itemSlotLabel(slot), bytes.Equal(v.Slots[slot].Bytes(), stack.Bytes()), "item/count/component mismatch")
	}
	check("saved selected hotbar slot", selected == 2, "")
}
func itemSlotLabel(slot int16) string {
	switch slot {
	case 9:
		return "main"
	case 10:
		return "potion"
	case 11:
		return "food"
	case 5:
		return "helmet"
	case 45:
		return "offhand"
	case 38:
		return "hotbar"
	}
	return "unknown"
}
func seedInventorySmoke(fc *proto.FramedConn) {
	for _, slot := range []int16{9, 10, 11, 5, 45, 38} {
		stack := savedInventorySample()[slot]
		w := proto.NewWriter().I16(slot)
		stack.Write(w)
		if err := fc.WritePacket(proto.CSetCreativeSlot, w.Payload()); err != nil {
			check("inventory update "+itemSlotLabel(slot), false, err.Error())
			return
		}
		ok := false
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) && !ok {
			id, payload, err := readWithTimeout(fc, 5*time.Second)
			if err != nil {
				break
			}
			if id == proto.SKeepAlive {
				fc.WritePacket(proto.CKeepAlive, payload)
			}
			if id != proto.SContainerSlot {
				continue
			}
			r := proto.NewReader(payload)
			r.U8()
			r.VarInt()
			index, err := r.I16()
			if err != nil || index != slot {
				continue
			}
			echo, err := item.Read(r)
			ok = err == nil && bytes.Equal(echo.Bytes(), stack.Bytes())
		}
		check("inventory update "+itemSlotLabel(slot), ok, "")
	}
	fc.WritePacket(proto.CHeldItemSlot, proto.NewWriter().I16(2).Payload())
	checkCommand(fc, "where", "Position:")
}
