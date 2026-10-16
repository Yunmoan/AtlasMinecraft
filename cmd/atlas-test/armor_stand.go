package main

import (
	"atlas/internal/item"
	"atlas/internal/proto"
	"crypto/md5"
	"fmt"
	"strings"
	"time"
)

// The dedicated fixture is imported original-JAR actor NBT in a temporary flat
// save, not evidence that this actor was produced by normal world generation.
var armorFixtureUUID = [16]byte{9}

type armorTCPClient struct {
	d             *dimensionClient
	name          string
	stand         int32
	spawnCount    int
	equipment     [7]item.Stack
	equipmentSeen bool
	inventory     item.Inventory
	inventorySeen bool
	position      [3]float64
	errorText     string
}

func newArmorTCPClient(name string) *armorTCPClient {
	return &armorTCPClient{name: name, stand: -1, d: &dimensionClient{
		fc: dial(), dimension: "minecraft:overworld",
		crystals: make(map[[16]byte]dimensionCrystal), mobs: make(map[[16]byte]dimensionMob),
		crystalMetadata: make(map[int32]bool), receivedChunks: make(map[[2]int32]bool),
	}}
}

func (c *armorTCPClient) consume(id int32, payload []byte) {
	switch id {
	case proto.SKickDisconnect:
		c.errorText = "server disconnected " + string(payload)
	case proto.SPositionSync:
		position, err := c.d.confirmPosition(payload)
		if err != nil {
			c.errorText = err.Error()
		} else {
			c.position = position
		}
		return // the position confirm must be written once
	case proto.SContainerContent:
		inventory, ok := readInventoryContent(payload)
		if !ok {
			c.errorText = "invalid player inventory packet"
		} else {
			c.inventory, c.inventorySeen = inventory, true
		}
	case proto.SSpawnEntity:
		r := proto.NewReader(payload)
		eid, err := r.VarInt()
		if err != nil {
			c.errorText = err.Error()
			break
		}
		uuid, err := r.UUID()
		if err != nil || uuid != armorFixtureUUID {
			break
		}
		kind, err := r.VarInt()
		var position [3]float64
		for i := range position {
			position[i], err = r.F64()
			if err != nil {
				c.errorText = err.Error()
				return
			}
		}
		if kind != 3 || position != [3]float64{-.5, -60, .5} || !c.d.receivedChunks[[2]int32{-1, 0}] {
			c.errorText = fmt.Sprintf("fixture stand kind/position/terrain order: %d %v", kind, position)
		}
		c.stand = eid
		c.spawnCount++
	case proto.SSetEquipment:
		r := proto.NewReader(payload)
		eid, err := r.VarInt()
		if err != nil || eid != c.stand {
			break
		}
		var slots [7]item.Stack
		for i := range slots {
			index, err := r.U8()
			if err != nil || int(index&0x7f) != i || (index&0x80 != 0) != (i < 6) {
				c.errorText = "invalid complete armor-stand equipment slot list"
				return
			}
			slots[i], err = item.Read(r)
			if err != nil {
				c.errorText = err.Error()
				return
			}
		}
		if r.Remaining() != 0 {
			c.errorText = "trailing equipment bytes"
		}
		c.equipment, c.equipmentSeen = slots, true
	}
	c.d.control(id, payload)
}

func (c *armorTCPClient) wait(label string, ready func() bool) bool {
	deadline := time.Now().Add(15 * time.Second)
	for !ready() && c.errorText == "" && time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(c.d.fc, 5*time.Second)
		if err != nil {
			c.errorText = err.Error()
			break
		}
		c.consume(id, payload)
	}
	ok := ready() && c.errorText == ""
	check(label+" ("+c.name+")", ok, c.errorText)
	return ok
}

// Named startup retains the initial inventory and actor packets which the
// dimension-only login helper does not track; framing/control is shared.
func loginArmorTCPClient(name string) *armorTCPClient {
	c := newArmorTCPClient(name)
	failed := true
	defer func() {
		if failed {
			c.d.fc.Conn.Close()
		}
	}()
	fc := c.d.fc
	handshake(fc, 2)
	uuid := md5.Sum([]byte("OfflinePlayer:" + name))
	uuid[6], uuid[8] = uuid[6]&15|0x30, uuid[8]&63|0x80
	if err := fc.WritePacket(proto.CLoginStart, proto.NewWriter().String(name).UUID(uuid).Payload()); err != nil {
		check("armor TCP login "+name, false, err.Error())
		return nil
	}
	success := false
	for i := 0; i < 5 && !success; i++ {
		id, payload, err := readWithTimeout(fc, 10*time.Second)
		if err != nil {
			check("armor TCP login "+name, false, err.Error())
			return nil
		}
		switch id {
		case proto.SLoginSetCompress:
			threshold, err := proto.NewReader(payload).VarInt()
			if err != nil {
				return nil
			}
			fc.EnableCompression(threshold)
		case proto.SLoginSuccess:
			success = true
		case proto.SLoginDisconnect:
			check("armor TCP login "+name, false, string(payload))
			return nil
		}
	}
	check("armor TCP login "+name, success, "")
	if !success || fc.WritePacket(proto.CLoginAcknowledged, nil) != nil {
		return nil
	}
	finished, registries := false, 0
	for i := 0; i < 30 && !finished; i++ {
		id, payload, err := readWithTimeout(fc, 10*time.Second)
		if err != nil {
			check("armor TCP configuration "+name, false, err.Error())
			return nil
		}
		switch id {
		case proto.SCfgKeepAlive:
			fc.WritePacket(proto.CCfgKeepAlive, payload)
		case proto.SCfgRegistryData:
			registries++
		case proto.SCfgFinish:
			finished = true
		}
	}
	check("armor TCP configuration "+name, finished && registries == 11, fmt.Sprint(registries))
	if !finished || registries != 11 {
		return nil
	}
	fc.WritePacket(proto.CCfgClientSettings, clientSettingsPayload(2))
	fc.WritePacket(proto.CCfgFinishAck, nil)
	sawLogin, sawPosition, sawChunk, sawBatch := false, false, false, false
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) && !(sawLogin && sawPosition && sawChunk && sawBatch && c.inventorySeen) {
		id, payload, err := readWithTimeout(fc, 5*time.Second)
		if err != nil {
			c.errorText = err.Error()
			break
		}
		switch id {
		case proto.SLogin:
			r := proto.NewReader(payload)
			r.I32()
			r.Bool()
			count, _ := r.VarInt()
			for i := int32(0); i < count; i++ {
				r.String(256)
			}
			for range 3 {
				r.VarInt()
			}
			for range 3 {
				r.Bool()
			}
			name, seed, err := readDimensionSpawnInfo(r, false)
			sawLogin = err == nil && name == "minecraft:overworld"
			c.d.dimension, c.d.seed = name, seed
		case proto.SPositionSync:
			sawPosition = true
		case proto.SChunkData:
			if !sawLogin || !sawPosition {
				c.errorText = "terrain preceded play login or position"
			}
			sawChunk = true
		case proto.SChunkBatchFinished:
			sawBatch = true
		}
		c.consume(id, payload)
		if c.errorText != "" {
			break
		}
	}
	ok := sawLogin && sawPosition && sawChunk && sawBatch && c.inventorySeen && c.errorText == ""
	check("armor TCP play/chunk/inventory "+c.name, ok, c.errorText)
	if !ok {
		return nil
	}
	failed = false
	return c
}

func (c *armorTCPClient) command(command, want string) (string, bool) {
	return c.commandReply(command, want, true)
}

func (c *armorTCPClient) commandReply(command, want string, report bool) (string, bool) {
	if err := c.d.fc.WritePacket(proto.CChatCommand, proto.NewWriter().String(command).Payload()); err != nil {
		check("armor TCP /"+command, false, err.Error())
		return "", false
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(c.d.fc, 5*time.Second)
		if err != nil {
			check("armor TCP /"+command, false, err.Error())
			return "", false
		}
		c.consume(id, payload)
		if id == proto.SSystemChat && strings.Contains(string(payload), want) {
			if report || c.errorText != "" {
				check("armor TCP /"+command, c.errorText == "", c.errorText)
			}
			return string(payload), c.errorText == ""
		}
	}
	check("armor TCP /"+command, false, "missing command response")
	return "", false
}

func armorEmptyEquipment(slots [7]item.Stack) bool {
	for _, stack := range slots {
		if !stack.Empty() {
			return false
		}
	}
	return true
}

func armorBootCount(v item.Inventory) int32 {
	boots, _ := item.ID("iron_boots")
	n := int32(0)
	for _, stack := range v.Slots {
		if !stack.Empty() && stack.ID == boots {
			n += stack.Count
		}
	}
	if !v.Carried.Empty() && v.Carried.ID == boots {
		n += v.Carried.Count
	}
	return n
}

func armorClick(c *armorTCPClient) bool {
	err := c.d.fc.WritePacket(proto.CInteractEntity, proto.NewWriter().VarInt(c.stand).VarInt(2).
		F32(0).F32(.2).F32(0).VarInt(1).Bool(false).Payload())
	if err != nil {
		check("armor TCP interactAt", false, err.Error())
	}
	return err == nil
}

func testArmorStandTCP() {
	fmt.Println("== fixture armor stand / two TCP clients / equipment / reconnect ==")
	owner := loginArmorTCPClient("ArmorBot")
	if owner == nil {
		return
	}
	defer owner.d.fc.Conn.Close()
	viewer := loginArmorTCPClient("ArmorViewer")
	if viewer == nil {
		return
	}
	defer viewer.d.fc.Conn.Close()
	boots, _ := item.ID("iron_boots")
	for _, c := range []*armorTCPClient{owner, viewer} {
		if !c.wait("fixture stand spawn after owner chunk and metadata", func() bool { return c.stand >= 0 && c.d.crystalMetadata[c.stand] }) {
			return
		}
		check("fixture single UUID "+c.name, c.spawnCount == 1, fmt.Sprint(c.spawnCount))
	}
	check("two TCP clients share stand entity ID", owner.stand == viewer.stand, fmt.Sprintf("%d/%d", owner.stand, viewer.stand))
	if *armorStandResume {
		for _, c := range []*armorTCPClient{owner, viewer} {
			if _, ok := c.command("tps", "TPS (1m/5m/15m):"); !ok {
				return
			}
			check("restart fixture has no equipment "+c.name, armorEmptyEquipment(c.equipment), "stale item returned")
		}
		check("restart owner boots offhand exactly once", owner.inventory.Slots[45].ID == boots && owner.inventory.Slots[45].Count == 1 && armorBootCount(owner.inventory) == 1, "wrong saved inventory")
		check("restart viewer received no boots", armorBootCount(viewer.inventory) == 0, "duplicate equipment")
		return
	}
	for _, c := range []*armorTCPClient{owner, viewer} {
		if !c.wait("fixture original boots equipment", func() bool { return c.equipmentSeen }) {
			return
		}
		check("fixture original boots in feet slot "+c.name, c.equipment[2].ID == boots && c.equipment[2].Count == 1, "wrong original equipment")
		check("initial player has no boots "+c.name, armorBootCount(c.inventory) == 0, "save is not a fresh fixture")
	}
	owner.equipmentSeen, viewer.equipmentSeen = false, false
	if !armorClick(owner) {
		return
	}
	if !owner.wait("interactAt boots received in offhand and stand explicitly emptied", func() bool {
		return owner.equipmentSeen && armorEmptyEquipment(owner.equipment) && owner.inventory.Slots[45].ID == boots && owner.inventory.Slots[45].Count == 1
	}) {
		return
	}
	if !viewer.wait("other TCP viewer receives explicit seven empty equipment slots", func() bool { return viewer.equipmentSeen && armorEmptyEquipment(viewer.equipment) }) {
		return
	}
	if !armorClick(viewer) {
		return
	}
	if _, ok := viewer.command("tps", "TPS (1m/5m/15m):"); !ok {
		return
	}
	check("second client cannot retrieve the same boots", armorBootCount(viewer.inventory) == 0 && armorBootCount(owner.inventory) == 1, "equipment duplicated")
	if _, ok := owner.command("save-all flush", "World saved."); !ok {
		return
	}
	owner.d.fc.Conn.Close()
	left := false
	deadline := time.Now().Add(5 * time.Second)
	for !left && time.Now().Before(deadline) {
		text, ok := viewer.commandReply("list", "Players online", false)
		if !ok {
			return
		}
		left = !strings.Contains(text, "ArmorBot")
		if !left {
			// Pace polling; completion depends on the server's observed player
			// list, not on assuming a fixed logout delay.
			time.Sleep(10 * time.Millisecond)
		}
	}
	check("owner disconnected and player save completed", left, "old player remains online")
	if !left {
		return
	}
	owner = loginArmorTCPClient("ArmorBot")
	if owner == nil {
		return
	}
	defer owner.d.fc.Conn.Close()
	if !owner.wait("reconnect fixture stand spawn/metadata", func() bool { return owner.stand >= 0 && owner.d.crystalMetadata[owner.stand] }) {
		return
	}
	if _, ok := owner.command("tps", "TPS (1m/5m/15m):"); !ok {
		return
	}
	check("reconnect saved boots in offhand exactly once", owner.inventory.Slots[45].ID == boots && owner.inventory.Slots[45].Count == 1 && armorBootCount(owner.inventory) == 1, "wrong player save")
	check("reconnect stand equipment stays empty", armorEmptyEquipment(owner.equipment), "old equipment returned")
	check("reconnect retains UUID and remaining viewer entity ID", owner.spawnCount == 1 && owner.stand == viewer.stand, "duplicated actor or changed ID")
	owner.command("save-all flush", "World saved.")
}
