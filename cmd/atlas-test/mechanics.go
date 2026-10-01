package main

import (
	"atlas/internal/blockstate"
	"atlas/internal/proto"
	"atlas/internal/world"
	"fmt"
	"time"
)

func sendCreativeBlock(fc *proto.FramedConn, name string) uint32 {
	state, err := blockstate.BlockStateID("minecraft:"+name, nil)
	if err != nil {
		panic(err)
	}
	for itemID, id := range world.ItemBlock {
		if id == state {
			fc.WritePacket(proto.CSetCreativeSlot, proto.NewWriter().I16(36).VarInt(64).VarInt(itemID).VarInt(0).VarInt(0).Payload())
			fc.WritePacket(proto.CHeldItemSlot, proto.NewWriter().I16(0).Payload())
			return state
		}
	}
	panic("no item for " + name)
}

func sendPlace(fc *proto.FramedConn, x, y, z, face, sequence int32) {
	fc.WritePacket(proto.CBlockPlace, proto.NewWriter().VarInt(0).I64(proto.PackPosition(x, y, z)).VarInt(face).F32(.5).F32(.5).F32(.5).Bool(false).VarInt(sequence).Payload())
}

func waitBlockUpdates(fc *proto.FramedConn, want map[[3]int32]uint32, sequence int32) bool {
	seen := make(map[[3]int32]bool)
	acked := sequence < 0
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && (!acked || len(seen) != len(want)) {
		id, payload, err := readWithTimeout(fc, 5*time.Second)
		if err != nil {
			return false
		}
		switch id {
		case proto.SBlockChange:
			r := proto.NewReader(payload)
			x, y, z, err := r.Position()
			if err != nil {
				return false
			}
			state, err := r.VarInt()
			if err != nil {
				return false
			}
			key := [3]int32{x, y, z}
			if expected, ok := want[key]; ok && uint32(state) == expected {
				seen[key] = true
			}
		case proto.SAckBlockChange:
			value, _ := proto.NewReader(payload).VarInt()
			if value == sequence {
				acked = true
			}
		case proto.SKeepAlive:
			value, _ := proto.NewReader(payload).I64()
			fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(value).Payload())
		}
	}
	return acked && len(seen) == len(want)
}

func checkMechanics(fc *proto.FramedConn) {
	stone := sendCreativeBlock(fc, "stone")
	sendPlace(fc, 0, 200, 0, 3, 8)
	check("creative placement + prediction ack", waitBlockUpdates(fc, map[[3]int32]uint32{{0, 200, 1}: stone}, 8), "")
	sendCreativeBlock(fc, "dirt")
	sendPlace(fc, 0, 200, 0, 3, 9)
	check("solid block overwrite rejected", waitBlockUpdates(fc, map[[3]int32]uint32{{0, 200, 1}: stone}, 9), "")
	sendCreativeBlock(fc, "stone")
	sendPlace(fc, 0, 200, 0, 1, 10)
	check("placement inside player rejected", waitBlockUpdates(fc, map[[3]int32]uint32{{0, 201, 0}: 0}, 10), "")
	checkCommand(fc, "setblock 10 200 0 stone", "Set block at 10 200 0")
	fc.WritePacket(proto.CBlockDig, proto.NewWriter().VarInt(0).I64(proto.PackPosition(10, 200, 0)).I8(1).VarInt(11).Payload())
	check("remote digging rejected", waitBlockUpdates(fc, map[[3]int32]uint32{{10, 200, 0}: stone}, 11), "")
	checkCommand(fc, "setblock 0 201 1 stone", "Set block at 0 201 1")
	checkCommand(fc, "setblock 0 202 1 stone", "Set block at 0 202 1")
	fc.WritePacket(proto.CPlayerAbilities, proto.NewWriter().U8(2).Payload())
	fc.WritePacket(proto.CPlayerPosition, proto.NewWriter().F64(.5).F64(201).F64(2.5).Bool(false).Payload())
	corrected := false
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && !corrected {
		id, payload, err := readWithTimeout(fc, 5*time.Second)
		if err != nil {
			break
		}
		if id == proto.SPositionSync {
			r := proto.NewReader(payload)
			x, _ := r.F64()
			y, _ := r.F64()
			z, _ := r.F64()
			r.F32()
			r.F32()
			r.U8()
			tid, _ := r.VarInt()
			fc.WritePacket(proto.CTeleportConfirm, proto.NewWriter().VarInt(tid).Payload())
			corrected = x == .5 && y == 201 && z == .5
		}
		if id == proto.SKeepAlive {
			value, _ := proto.NewReader(payload).I64()
			fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(value).Payload())
		}
	}
	check("flying wall crossing corrected", corrected, "")
	checkCommand(fc, "setblock 1 200 0 stone", "Set block at 1 200 0")
	checkCommand(fc, "setblock 1 201 0 air", "Set block at 1 201 0")
	sand, err := blockstate.BlockStateID("minecraft:sand", nil)
	if err != nil {
		panic(err)
	}
	fc.WritePacket(proto.CChatCommand, proto.NewWriter().String(fmt.Sprintf("setblock 1 210 0 %d", sand)).Payload())
	check("sand settles above support", waitBlockUpdates(fc, map[[3]int32]uint32{{1, 201, 0}: sand}, -1), "")
	checkCommand(fc, "setblock 2 200 0 grass_block", "Set block at 2 200 0")
	lower := sendCreativeBlock(fc, "tall_grass")
	upper, err := blockstate.BlockStateID("minecraft:tall_grass", map[string]string{"half": "upper"})
	if err != nil {
		panic(err)
	}
	sendPlace(fc, 2, 200, 0, 1, 12)
	check("double plant places both halves", waitBlockUpdates(fc, map[[3]int32]uint32{{2, 201, 0}: lower, {2, 202, 0}: upper}, 12), "")
	fc.WritePacket(proto.CBlockDig, proto.NewWriter().VarInt(0).I64(proto.PackPosition(2, 200, 0)).I8(1).VarInt(13).Payload())
	check("support removal updates both halves", waitBlockUpdates(fc, map[[3]int32]uint32{{2, 201, 0}: 0, {2, 202, 0}: 0}, 13), "")
}
