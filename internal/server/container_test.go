package server

import (
	"atlas/internal/anvil"
	"atlas/internal/hub"
	"atlas/internal/item"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type chestWireSnapshot struct {
	window  byte
	state   int32
	slots   [63]item.Stack
	carried item.Stack
}

func readChestPacket(t *testing.T, packet player.OutPacket) chestWireSnapshot {
	t.Helper()
	r := proto.NewReader(packet.Payload)
	var value chestWireSnapshot
	var err error
	value.window, err = r.U8()
	if err != nil {
		t.Fatal(err)
	}
	value.state, err = r.VarInt()
	if err != nil {
		t.Fatal(err)
	}
	count, err := r.VarInt()
	if err != nil || count != 63 {
		t.Fatal("invalid chest content size", count, err)
	}
	for i := range value.slots {
		value.slots[i], err = item.Read(r)
		if err != nil {
			t.Fatal(err)
		}
	}
	value.carried, err = item.Read(r)
	if err != nil || r.Remaining() != 0 {
		t.Fatal("invalid chest cursor", err)
	}
	return value
}
func chestClickPacket(value chestWireSnapshot, slot, button, mode int, forged bool) []byte {
	w := proto.NewWriter().U8(value.window).VarInt(value.state).I16(int16(slot)).I8(int8(button)).VarInt(int32(mode))
	if forged {
		w.VarInt(1).I16(0)
		id, _ := item.ID("diamond")
		item.New(id, 64).Write(w)
	} else {
		w.VarInt(0)
	}
	item.Stack{}.Write(w)
	return w.Payload()
}
func drainChestPackets(t *testing.T, p *player.Player) chestWireSnapshot {
	t.Helper()
	var result chestWireSnapshot
	found := false
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID == proto.SContainerContent && len(packet.Payload) > 0 && packet.Payload[0] != 0 {
			result = readChestPacket(t, packet)
			found = true
		}
	}
	if !found {
		t.Fatal("no authoritative chest content")
	}
	return result
}
func chestSessionServer(t *testing.T) (*Server, *player.Player, *player.Player, world.BlockPosition, string) {
	t.Helper()
	dir := t.TempDir()
	seed := int64(1)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(h, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	p := world.BlockPosition{X: -1, Y: -60, Z: 1}
	state, err := world.BlockStateID("minecraft:chest", map[string]string{"type": "single", "facing": "north", "waterlogged": "false"})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(p.X, p.Y, p.Z, state); err != nil {
		t.Fatal(err)
	}
	s := New(Config{}, h, w)
	a := player.New("ChestOne", [16]byte{11}, 1, nil)
	b := player.New("ChestTwo", [16]byte{12}, 2, nil)
	for _, viewer := range []*player.Player{a, b} {
		viewer.SetPos(.5, -60, .5, 0, 0, true)
		h.Join(viewer)
	}
	return s, a, b, p, dir
}
func TestChestSessionsForgedStateAndConcurrentTake(t *testing.T) {
	s, a, b, p, _ := chestSessionServer(t)
	id, _ := item.ID("stone")
	if _, err := s.World.UpdateChest(p, 0, func(v *world.ChestInventory) error { v.Slots[0] = item.New(id, 1); return nil }); err != nil {
		t.Fatal(err)
	}
	// Exercise UseItemOn -> useBlock -> original 9x3 menu, not a direct open.
	s.handlePlace(a, placePayload(p.X, p.Y, p.Z, 1, 1, .5))
	s.handlePlace(b, placePayload(p.X, p.Y, p.Z, 1, 2, .5))
	one, two := drainChestPackets(t, a), drainChestPackets(t, b)
	bad := one
	bad.window = 0
	s.handleContainerClick(a, chestClickPacket(bad, 0, 0, 0, true))
	result := drainChestPackets(t, a)
	if result.slots[0].Count != 1 || !result.carried.Empty() {
		t.Fatal("wrong window trusted")
	}
	bad = one
	bad.state++
	s.handleContainerClick(a, chestClickPacket(bad, 0, 0, 0, true))
	result = drainChestPackets(t, a)
	if result.slots[0].Count != 1 || !result.carried.Empty() {
		t.Fatal("forged state trusted")
	}
	var wg sync.WaitGroup
	for i, viewer := range []*player.Player{a, b} {
		wg.Add(1)
		value := []chestWireSnapshot{one, two}[i]
		go func() { defer wg.Done(); s.handleContainerClick(viewer, chestClickPacket(value, 0, 0, 0, true)) }()
	}
	wg.Wait()
	first, second := drainChestPackets(t, a), drainChestPackets(t, b)
	if first.slots[0].Count != 0 || second.slots[0].Count != 0 || first.carried.Count+second.carried.Count != 1 {
		t.Fatal("two viewers duplicated stale chest items")
	}
	for _, viewer := range []*player.Player{a, b} {
		inventory, _ := viewer.Inventory()
		for _, stack := range inventory.Slots {
			if !stack.Empty() {
				name, _ := item.Name(stack.ID)
				if name == "minecraft:diamond" {
					t.Fatal("forged client delta imported")
				}
			}
		}
	}
	// Window-zero close while a menu is active must not return its cursor.
	s.handlePlayPacket(a, proto.CCloseContainer, []byte{0})
	after := drainChestPackets(t, a)
	if after.carried.Count != first.carried.Count {
		t.Fatal("wrong close window mutated cursor")
	}
	for _, viewer := range []*player.Player{a, b} {
		s.forgetChest(viewer)
		saved, err := s.World.LoadInventory(viewer.UUID)
		if err != nil {
			t.Fatal(err)
		}
		inventory, _ := viewer.Inventory()
		if saved.Carried.Count != inventory.Carried.Count {
			t.Fatal("close failed cursor persistence")
		}
	}
}
func TestChestSessionSaveFailureLeavesPlayerAndChestUntouched(t *testing.T) {
	s, p, _, position, dir := chestSessionServer(t)
	id, _ := item.ID("diamond_sword")
	if _, err := s.World.UpdateChest(position, 0, func(v *world.ChestInventory) error { v.Slots[0] = item.New(id, 1); return nil }); err != nil {
		t.Fatal(err)
	}
	s.openChest(p, position)
	snapshot := drainChestPackets(t, p)
	path := filepath.Join(dir, "region", "r.-1.0.mca")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	s.handleContainerClick(p, chestClickPacket(snapshot, 0, 0, 0, false))
	after := drainChestPackets(t, p)
	if !after.carried.Empty() || after.slots[0].Count != 1 {
		t.Fatal("failed chest save changed player/menu")
	}
	inventory, _ := p.Inventory()
	chest, err := s.World.ReadChest(position, 0)
	if err != nil || !inventory.Carried.Empty() || chest.Slots[0].Count != 1 {
		t.Fatal("failed commit changed authority", err)
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	s.forgetChest(p)
}
func TestChestSessionRangeBlockAndFullCursorClose(t *testing.T) {
	s, p, _, position, _ := chestSessionServer(t)
	id, _ := item.ID("stone")
	if _, err := s.World.UpdateChest(position, 0, func(v *world.ChestInventory) error { v.Slots[0] = item.New(id, 11); return nil }); err != nil {
		t.Fatal(err)
	}
	s.openChest(p, position)
	snapshot := drainChestPackets(t, p)
	s.handleContainerClick(p, chestClickPacket(snapshot, 0, 0, 0, false))
	drainChestPackets(t, p)
	inventory, _ := p.Inventory()
	for i := 9; i <= 44; i++ {
		inventory.Slots[i] = item.New(id, 64)
	}
	s.setInventory(p, inventory)
	if !s.closeChest(p, snapshot.window, false) {
		t.Fatal("menu did not close")
	}
	inventory, _ = p.Inventory()
	if inventory.Carried.Count != 11 {
		t.Fatal("full inventory close lost cursor")
	}
	saved, err := s.World.LoadInventory(p.UUID)
	if err != nil || saved.Carried.Count != 11 {
		t.Fatal("full cursor not saved", err)
	}
	s.openChest(p, position)
	snapshot = drainChestPackets(t, p)
	p.SetPos(40, -60, .5, 0, 0, true)
	s.handleContainerClick(p, chestClickPacket(snapshot, 0, 0, 0, false))
	s.containers.mu.Lock()
	active := s.containers.players[p]
	s.containers.mu.Unlock()
	if active != nil {
		t.Fatal("out of range chest remains open")
	}
	p.SetPos(.5, -60, .5, 0, 0, true)
	s.openChest(p, position)
	snapshot = drainChestPackets(t, p)
	if err := s.World.SetBlock(position.X, position.Y, position.Z, world.StateAir); err != nil {
		t.Fatal(err)
	}
	s.handleContainerClick(p, chestClickPacket(snapshot, 0, 0, 0, false))
	s.containers.mu.Lock()
	active = s.containers.players[p]
	s.containers.mu.Unlock()
	if active != nil {
		t.Fatal("removed chest remains open")
	}
}

func TestChestAutomaticCloseAndReplacementPublishHeldEquipment(t *testing.T) {
	s, p, observer, position, _ := chestSessionServer(t)
	stone, _ := item.ID("stone")
	sword, _ := item.ID("diamond_sword")
	inventory := item.Inventory{Carried: item.New(sword, 1)}
	for i := 9; i <= 44; i++ {
		inventory.Slots[i] = item.New(stone, 64)
	}
	inventory.Slots[36] = item.Stack{}
	s.setInventory(p, inventory)
	observer.SetPlayReady(true)
	observer.MarkVisible(p.EID)
	observer.MarkChunkSent(0, 0)
	s.openChest(p, position)
	drainChestPackets(t, p)
	if err := s.World.SetBlock(position.X, position.Y+1, position.Z, world.StateStone); err != nil {
		t.Fatal(err)
	}
	s.validateChestSession(p)
	assertHeld := func() {
		t.Helper()
		seen := false
		for len(observer.Out) > 0 {
			packet := <-observer.Out
			if packet.ID != proto.SSetEquipment {
				continue
			}
			r := proto.NewReader(packet.Payload)
			eid, _ := r.VarInt()
			index, _ := r.U8()
			stack, err := item.Read(r)
			if eid == p.EID && index == 0x80 && err == nil && stack.ID == sword && stack.Count == 1 {
				seen = true
			}
		}
		if !seen {
			t.Fatal("automatic cursor return omitted observer held equipment")
		}
	}
	assertHeld()
	if err := s.World.SetBlock(position.X, position.Y+1, position.Z, world.StateAir); err != nil {
		t.Fatal(err)
	}
	inventory.Slots[36] = item.Stack{}
	inventory.Carried = item.New(sword, 1)
	s.setInventory(p, inventory)
	s.openChest(p, position)
	drainChestPackets(t, p)
	s.openChest(p, position)
	drainChestPackets(t, p)
	assertHeld()
	s.forgetChest(p)
}

func TestChestBackgroundCloseSerializesInventoryAndSelected(t *testing.T) {
	s, p, _, position, _ := chestSessionServer(t)
	stone, _ := item.ID("stone")
	diamond, _ := item.ID("diamond")
	for iteration := 0; iteration < 32; iteration++ {
		for len(p.Out) > 0 {
			<-p.Out
		}
		p.SetPos(.5, -60, .5, 0, 0, true)
		s.setInventory(p, item.Inventory{Carried: item.New(stone, 1)})
		s.openChest(p, position)
		drainChestPackets(t, p)
		p.SetPos(40, -60, .5, 0, 0, true)
		packet := proto.NewWriter().I16(10)
		item.New(diamond, 7).Write(packet)
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(3)
		go func() { defer wg.Done(); <-start; s.validateChestSession(p) }()
		go func() { defer wg.Done(); <-start; s.handleCreativeInventorySlot(p, packet.Payload()) }()
		go func() {
			defer wg.Done()
			<-start
			s.handlePlayPacket(p, proto.CHeldItemSlot, proto.NewWriter().I16(3).Payload())
		}()
		close(start)
		wg.Wait()
		inventory, _ := p.Inventory()
		creativeAcknowledged := false
		for len(p.Out) > 0 {
			packet := <-p.Out
			if packet.ID == proto.SContainerSlot {
				r := proto.NewReader(packet.Payload)
				window, _ := r.U8()
				r.VarInt()
				slot, _ := r.I16()
				if window == 0 && slot == 10 {
					creativeAcknowledged = true
				}
			}
		}
		creativeOK := inventory.Slots[10].Empty()
		if creativeAcknowledged {
			creativeOK = inventory.Slots[10].ID == diamond && inventory.Slots[10].Count == 7
		}
		if !inventory.Carried.Empty() || inventory.Slots[9].ID != stone || inventory.Slots[9].Count != 1 || !creativeOK || inventory.Selected != 3 {
			t.Fatal("background close overwrote creative slot, selected slot or cursor", iteration)
		}
	}
}

func TestChestPipelinedDragStaleClickAndOffhandWire(t *testing.T) {
	s, p, _, position, _ := chestSessionServer(t)
	stone, _ := item.ID("stone")
	sword, _ := item.ID("diamond_sword")
	if _, err := s.World.UpdateChest(position, 0, func(v *world.ChestInventory) error { v.Slots[3] = item.New(stone, 5); return nil }); err != nil {
		t.Fatal(err)
	}
	s.setInventory(p, item.Inventory{Carried: item.New(stone, 12)})
	s.openChest(p, position)
	start := drainChestPackets(t, p)
	// A GUI may send the whole drag before any acknowledgement arrives.
	var after chestWireSnapshot
	for i, action := range [][2]int{{-999, 0}, {0, 1}, {1, 1}, {-999, 2}} {
		s.handleContainerClick(p, chestClickPacket(start, action[0], action[1], 5, false))
		after = drainChestPackets(t, p)
		if i < 3 && after.state != start.state {
			t.Fatal("no-op drag changed state for existing saved chest items")
		}
	}
	if after.slots[0].Count != 6 || after.slots[1].Count != 6 || !after.carried.Empty() {
		t.Fatal("pipelined drag rejected valid same-state phases")
	}
	// A stale click still replays against current server contents.
	s.handleContainerClick(p, chestClickPacket(start, 0, 0, 0, true))
	after = drainChestPackets(t, p)
	if !after.slots[0].Empty() || after.carried.Count != 6 {
		t.Fatal("legitimate stale click did not replay against current state")
	}
	s.handleContainerClick(p, chestClickPacket(after, 2, 0, 0, false))
	after = drainChestPackets(t, p)
	inventory, _ := p.Inventory()
	inventory.Slots[45] = item.New(sword, 1)
	s.setInventory(p, inventory)
	s.handleContainerClick(p, chestClickPacket(after, 1, 40, 2, false))
	seenOffhand := false
	var menu chestWireSnapshot
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID == proto.SContainerContent && packet.Payload[0] != 0 {
			menu = readChestPacket(t, packet)
		}
		if packet.ID == proto.SContainerSlot {
			r := proto.NewReader(packet.Payload)
			window, _ := r.U8()
			r.VarInt()
			slot, _ := r.I16()
			stack, err := item.Read(r)
			if window == 254 && slot == 40 && err == nil && stack.ID == stone && stack.Count == 6 {
				seenOffhand = true
			}
		}
	}
	if !seenOffhand || menu.slots[1].ID != sword || menu.slots[1].Count != 1 {
		t.Fatal("offhand swap omitted hidden authoritative player slot")
	}
	s.forgetChest(p)
}

func TestChestRejectsCreativeWindowZeroDuringMenu(t *testing.T) {
	s, p, _, position, _ := chestSessionServer(t)
	s.openChest(p, position)
	before := drainChestPackets(t, p)
	diamond, _ := item.ID("diamond")
	packet := proto.NewWriter().I16(9)
	item.New(diamond, 64).Write(packet)
	s.handleCreativeInventorySlot(p, packet.Payload())
	after := drainChestPackets(t, p)
	inventory, _ := p.Inventory()
	if after.state != before.state || !after.slots[27].Empty() || !inventory.Slots[9].Empty() {
		t.Fatal("forged creative inventory operation changed active menu")
	}
	s.closeChest(p, before.window, false)
	s.handleCreativeInventorySlot(p, packet.Payload())
	inventory, _ = p.Inventory()
	if inventory.Slots[9].ID != diamond || inventory.Slots[9].Count != 64 {
		t.Fatal("normal creative GUI operation rejected after menu close")
	}
}

func TestChestOpaqueSameItemSwapCommitsBothOwners(t *testing.T) {
	s, p, _, position, _ := chestSessionServer(t)
	stack := func(owner string) item.Stack {
		t.Helper()
		value, err := item.FromNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:stone"}, "count": {Type: nbt.Int, Value: int32(1)}, "ForeignItemField": {Type: nbt.String, Value: owner}}})
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	chestItem, playerItem := stack("chest-original"), stack("player-original")
	if !chestItem.SameItem(playerItem) {
		t.Fatal("fixture must have identical wire item identity")
	}
	if _, err := s.World.UpdateChest(position, 0, func(v *world.ChestInventory) error { v.Slots[0] = chestItem; return nil }); err != nil {
		t.Fatal(err)
	}
	inventory := item.Inventory{}
	inventory.Slots[36] = playerItem
	s.setInventory(p, inventory)
	s.openChest(p, position)
	before := drainChestPackets(t, p)
	s.handleContainerClick(p, chestClickPacket(before, 0, 0, 2, false))
	after := drainChestPackets(t, p)
	if after.state != before.state {
		t.Fatal("opaque save fields changed unchanged wire menu state")
	}
	owner := func(value item.Stack) string {
		t.Helper()
		tag, err := value.ToNBT()
		if err != nil {
			t.Fatal(err)
		}
		data, _ := tag.Object()
		name, _ := data.Text("ForeignItemField")
		return name
	}
	inventory, _ = p.Inventory()
	chest, err := s.World.ReadChest(position, 0)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := s.World.LoadInventory(p.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if owner(inventory.Slots[36]) != "chest-original" || owner(saved.Slots[36]) != "chest-original" || owner(chest.Slots[0]) != "player-original" {
		t.Fatal("same wire stack swap skipped one owner's opaque NBT commit")
	}
	s.forgetChest(p)
}

// Native structures and animals use many entity types. This reader verifies
// container packets without the player-only trackingClient.await assertions.
func awaitChestTCP(t *testing.T, c *trackingClient, match func(player.OutPacket) bool) player.OutPacket {
	t.Helper()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	for {
		select {
		case packet := <-c.packets:
			if packet.ID == proto.SChunkData {
				r := proto.NewReader(packet.Payload)
				x, _ := r.I32()
				z, _ := r.I32()
				c.delivered[[2]int32{x, z}] = true
			}
			if match(packet) {
				return packet
			}
		case err := <-c.errors:
			t.Fatal(err)
		case <-timer.C:
			t.Fatal("timed out waiting for native chest protocol transition")
		}
	}
}
func nativeGeneratedChest(t *testing.T, w *world.World, dir, prefix string, chunks [][2]int32) (world.BlockPosition, [27]item.Stack) {
	t.Helper()
	for _, chunk := range chunks {
		if _, err := w.LoadChunk(chunk[0], chunk[1]); err != nil {
			t.Fatal(err)
		}
		_, tag, found, err := anvil.ReadChunk(filepath.Join(dir, "region"), chunk[0], chunk[1])
		if err != nil || !found {
			t.Fatal("native generated chunk not saved", err)
		}
		root, _ := tag.Object()
		entities, _ := root["block_entities"].Elements()
		for _, tag := range entities {
			data, _ := tag.Object()
			table, _ := data.Text("LootTable")
			if !strings.HasPrefix(table, prefix) {
				continue
			}
			x, _ := data.Number("x")
			y, _ := data.Number("y")
			z, _ := data.Number("z")
			seed, _ := data.Number("LootTableSeed")
			p := world.BlockPosition{X: int32(x), Y: int32(y), Z: int32(z)}
			want, err := world.FillChestLoot(table, seed, 0)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("actual normal generator chest %s at %+v, LootTableSeed%d", table, p, seed)
			return p, want
		}
	}
	t.Fatal("native generator produced no pinned chest")
	return world.BlockPosition{}, [27]item.Stack{}
}
func TestTCPNativeVillageAndOutpostChestCloseLogoutRestart(t *testing.T) {
	fixtures := []struct {
		name, prefix string
		chunks       [][2]int32
	}{{"village", "minecraft:chests/village/", [][2]int32{{-24, 47}, {-28, 45}, {-25, 46}, {-27, 47}, {-25, 47}}}, {"outpost", "minecraft:chests/pillager_outpost", [][2]int32{{-249, 327}, {-248, 328}, {-250, 328}, {-248, 326}}}}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			dir := t.TempDir()
			seed := int64(1234)
			if err := world.CreateNormalWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			h := hub.New()
			w, err := world.Open(h, 2, dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { w.Close() })
			position, want := nativeGeneratedChest(t, w, dir, fixture.prefix, fixture.chunks)
			if _, err := w.SetSpawn(position.X+1, position.Y, position.Z, 0); err != nil {
				t.Fatal(err)
			}
			selected := -1
			for i, stack := range want {
				if !stack.Empty() {
					selected = i
					break
				}
			}
			if selected < 0 {
				t.Fatal("pinned chest rolled no loot")
			}
			taken := want[selected].Clone()
			for pass := 0; pass < 2; pass++ {
				if pass == 1 {
					h = hub.New()
					w, err = world.Open(h, 2, dir)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { w.Close() })
				}
				s := New(Config{MaxPlayers: 10, ViewDistance: 1, SimulationDistance: 1}, h, w)
				address := serveTrackingConnections(t, s, 1)
				client := connectTrackingClient(t, address, "NativeChestProbe")
				initial := readInventoryPacket(t, awaitChestTCP(t, client, func(p player.OutPacket) bool { return p.ID == proto.SContainerContent && p.Payload[0] == 0 }))
				awaitChestTCP(t, client, func(p player.OutPacket) bool { return p.ID == proto.SPositionSync })
				chestChunk := [2]int32{position.X >> 4, position.Z >> 4}
				if !client.delivered[chestChunk] {
					awaitChestTCP(t, client, func(p player.OutPacket) bool { return p.ID == proto.SChunkData && client.delivered[chestChunk] })
				}
				if pass == 1 {
					if initial.Slots[9].Count != taken.Count || !initial.Slots[9].SameItem(taken) {
						t.Fatal("restart lost actual native chest item/components")
					}
				}
				if err := client.send(proto.CBlockPlace, placePayload(position.X, position.Y, position.Z, 1, int32(pass+10), .5)); err != nil {
					t.Fatal(err)
				}
				open := awaitChestTCP(t, client, func(p player.OutPacket) bool { return p.ID == proto.SOpenScreen })
				r := proto.NewReader(open.Payload)
				window, _ := r.VarInt()
				kind, _ := r.VarInt()
				if window <= 0 || kind != 2 {
					t.Fatal("native chest opened wrong menu")
				}
				snapshot := readChestPacket(t, awaitChestTCP(t, client, func(p player.OutPacket) bool { return p.ID == proto.SContainerContent && p.Payload[0] == byte(window) }))
				for i, stack := range want {
					if snapshot.slots[i].Count != stack.Count || !stack.Empty() && !snapshot.slots[i].SameItem(stack) {
						t.Fatal("native chest contents mismatch", pass, i)
					}
				}
				if pass == 0 {
					if err := client.send(proto.CContainerClick, chestClickPacket(snapshot, selected, 0, 0, true)); err != nil {
						t.Fatal(err)
					}
					snapshot = readChestPacket(t, awaitChestTCP(t, client, func(p player.OutPacket) bool { return p.ID == proto.SContainerContent && p.Payload[0] == byte(window) }))
					if !snapshot.slots[selected].Empty() || snapshot.carried.Count != taken.Count || !snapshot.carried.SameItem(taken) {
						t.Fatal("native chest pickup ignored authoritative loot")
					}
					want[selected] = item.Stack{}
				}
				if err := client.send(proto.CCloseContainer, []byte{byte(window)}); err != nil {
					t.Fatal(err)
				}
				closed := readInventoryPacket(t, awaitChestTCP(t, client, func(p player.OutPacket) bool { return p.ID == proto.SContainerContent && p.Payload[0] == 0 }))
				if !closed.Carried.Empty() || closed.Slots[9].Count != taken.Count || !closed.Slots[9].SameItem(taken) {
					t.Fatal("close failed native item return")
				}
				client.conn.Conn.Close()
				deadline := time.Now().Add(10 * time.Second)
				for h.Count() != 0 && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
				if h.Count() != 0 {
					t.Fatal("native chest logout did not finish")
				}
				if err := w.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
