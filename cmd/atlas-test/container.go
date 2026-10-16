package main

import (
	"atlas/internal/item"
	"atlas/internal/proto"
	"atlas/internal/world"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type chestTCPState struct {
	Position world.BlockPosition
	Chest    [27]item.Stack
	Player   item.Inventory
}
type chestTCPMenu struct {
	window  byte
	state   int32
	slots   [63]item.Stack
	carried item.Stack
}

func readChestTCPMenu(payload []byte) (chestTCPMenu, error) {
	r := proto.NewReader(payload)
	var result chestTCPMenu
	var err error
	result.window, err = r.U8()
	if err != nil {
		return result, err
	}
	result.state, err = r.VarInt()
	if err != nil {
		return result, err
	}
	count, err := r.VarInt()
	if err != nil || count != 63 {
		return result, fmt.Errorf("invalid 9x3 menu slot count%d", count)
	}
	for i := range result.slots {
		result.slots[i], err = item.Read(r)
		if err != nil {
			return result, err
		}
	}
	result.carried, err = item.Read(r)
	if err == nil && r.Remaining() != 0 {
		err = fmt.Errorf("trailing chest content bytes")
	}
	return result, err
}
func chestTCPRead(c *dimensionClient, idWanted int32, window byte) ([]byte, bool) {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(c.fc, 15*time.Second)
		if err != nil {
			check("chest TCP transition", false, err.Error())
			return nil, false
		}
		c.control(id, payload)
		if id == idWanted && (id != proto.SContainerContent || len(payload) > 0 && payload[0] == window) {
			return payload, true
		}
	}
	check("chest TCP transition", false, "packet deadline")
	return nil, false
}
func sameChestStack(a, b item.Stack) bool { return a.Count == b.Count && (a.Empty() || a.SameItem(b)) }
func testChestTCP() {
	fmt.Println("== native single chest / original loot / saved items ==")
	var position world.BlockPosition
	var expected chestTCPState
	if *chestResume {
		raw, err := os.ReadFile(*chestState)
		if err != nil {
			check("chest saved-state fixture", false, err.Error())
			return
		}
		if err := json.Unmarshal(raw, &expected); err != nil {
			check("chest saved-state fixture", false, err.Error())
			return
		}
		position = expected.Position
	} else {
		parts := strings.Split(*chestPosition, ",")
		if len(parts) != 3 {
			check("chest position", false, "use x,y,z")
			return
		}
		values := []*int32{&position.X, &position.Y, &position.Z}
		for i, part := range parts {
			number, err := strconv.ParseInt(strings.TrimSpace(part), 10, 32)
			if err != nil {
				check("chest position", false, err.Error())
				return
			}
			*values[i] = int32(number)
		}
	}
	c := loginDimensionClient("minecraft:overworld")
	if c == nil {
		return
	}
	defer c.fc.Conn.Close()
	if !c.command(fmt.Sprintf("tp %g %d %g", float64(position.X)+.5, position.Y+1, float64(position.Z)+.5), "Teleported TestBot to") {
		return
	}
	chunk := [2]int32{position.X >> 4, position.Z >> 4}
	deadline := time.Now().Add(60 * time.Second)
	for !c.receivedChunks[chunk] && time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(c.fc, 15*time.Second)
		if err != nil {
			check("chest terrain delivery", false, err.Error())
			return
		}
		c.control(id, payload)
	}
	if !c.receivedChunks[chunk] {
		check("chest terrain delivery", false, "chest chunk not received")
		return
	}
	check("chest terrain delivered before use", true, "")
	use := proto.NewWriter().VarInt(0).I64(proto.PackPosition(position.X, position.Y, position.Z)).VarInt(1).F32(.5).F32(.5).F32(.5).Bool(false).VarInt(301).Payload()
	if err := c.fc.WritePacket(proto.CBlockPlace, use); err != nil {
		check("native chest use", false, err.Error())
		return
	}
	open, ok := chestTCPRead(c, proto.SOpenScreen, 0)
	if !ok {
		return
	}
	r := proto.NewReader(open)
	window, e1 := r.VarInt()
	kind, e2 := r.VarInt()
	valid := e1 == nil && e2 == nil && window > 0 && window <= 100 && kind == 2
	check("native chest OpenScreen generic9x3", valid, fmt.Sprintf("window%d menu%d", window, kind))
	if !valid {
		return
	}
	content, ok := chestTCPRead(c, proto.SContainerContent, byte(window))
	if !ok {
		return
	}
	menu, err := readChestTCPMenu(content)
	if err != nil {
		check("native chest contents", false, err.Error())
		return
	}
	if *chestResume {
		valid := true
		for i, stack := range expected.Chest {
			valid = valid && sameChestStack(menu.slots[i], stack)
		}
		for i := 27; i < 63; i++ {
			valid = valid && sameChestStack(menu.slots[i], expected.Player.Slots[i-18])
		}
		valid = valid && sameChestStack(menu.carried, expected.Player.Carried)
		check("restart all27 chest slots and player items/components", valid, "")
		if !valid {
			return
		}
	} else {
		want, err := world.FillChestLoot(*chestLootTable, *chestLootSeed, 0)
		if err != nil {
			check("original chest loot arguments", false, err.Error())
			return
		}
		valid := true
		slot := -1
		for i, stack := range want {
			valid = valid && sameChestStack(menu.slots[i], stack)
			if slot < 0 && !stack.Empty() {
				slot = i
			}
		}
		check("native chest all27 exact original loot slots/count/components", valid, "")
		if !valid || slot < 0 {
			return
		}
		clicked := proto.NewWriter().U8(byte(window)).VarInt(menu.state).I16(int16(slot)).I8(0).VarInt(1).VarInt(1).I16(int16(slot))
		forged, _ := item.ID("diamond")
		item.New(forged, 64).Write(clicked)
		item.Stack{}.Write(clicked)
		if err := c.fc.WritePacket(proto.CContainerClick, clicked.Payload()); err != nil {
			check("chest shift click", false, err.Error())
			return
		}
		content, ok = chestTCPRead(c, proto.SContainerContent, byte(window))
		if !ok {
			return
		}
		after, err := readChestTCPMenu(content)
		if err != nil {
			check("chest after-click content", false, err.Error())
			return
		}
		beforeCount, afterCount := int32(0), int32(0)
		for i := 27; i < 63; i++ {
			if menu.slots[i].SameItem(want[slot]) {
				beforeCount += menu.slots[i].Count
			}
			if after.slots[i].SameItem(want[slot]) {
				afterCount += after.slots[i].Count
			}
		}
		valid = after.slots[slot].Empty() && afterCount-beforeCount == want[slot].Count
		for i, stack := range want {
			if i != slot {
				valid = valid && sameChestStack(after.slots[i], stack)
			}
		}
		check("authoritative shift transfer ignored forged diamond delta", valid, "")
		if !valid {
			return
		}
		menu = after
	}
	if err := c.fc.WritePacket(proto.CCloseContainer, []byte{byte(window)}); err != nil {
		check("chest close", false, err.Error())
		return
	}
	closed, ok := chestTCPRead(c, proto.SContainerContent, 0)
	if !ok {
		return
	}
	inventory, valid := readInventoryContent(closed)
	check("closed chest player inventory and cursor", valid && inventory.Carried.Empty(), "")
	if !valid {
		return
	}
	if !*chestResume {
		expected = chestTCPState{Position: position, Player: inventory}
		copy(expected.Chest[:], menu.slots[:27])
		raw, err := json.Marshal(expected)
		if err == nil {
			err = os.WriteFile(*chestState, raw, 0600)
		}
		check("wrote chest restart expectation", err == nil, fmt.Sprint(err))
		if err != nil {
			return
		}
	}
	if !c.command("save-all flush", "World saved.") {
		return
	}
	check("chest inventory save-all accepted", true, "")
}
