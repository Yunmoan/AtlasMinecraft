package server

import (
	"atlas/internal/anvil"
	"atlas/internal/hub"
	"atlas/internal/item"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestEntityInteractionPacketsAgainstVanillaJAR(t *testing.T) {
	source, err := os.ReadFile("../../docs/worldgen/armor_stand_interaction_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := 0
	for _, line := range strings.Split(string(source), "\n") {
		f := strings.Fields(line)
		if len(f) != 5 || f[0] != "armor_interact_packet" {
			continue
		}
		action, _ := strconv.Atoi(f[1])
		hand, _ := strconv.Atoi(f[2])
		secondary, _ := strconv.ParseBool(f[3])
		raw, err := hex.DecodeString(f[4])
		if err != nil {
			t.Fatal(err)
		}
		parsed, ok := readEntityInteraction(raw)
		if !ok || parsed.target != 77 || int(parsed.action) != action || parsed.secondary != secondary || action != 1 && int(parsed.hand) != hand {
			t.Fatal("official packet decode differs", line)
		}
		if action == 2 && parsed.point != [3]float32{.1, 1.2, -.2} {
			t.Fatal(parsed.point)
		}
		w := proto.NewWriter().VarInt(77).VarInt(int32(action))
		if action == 2 {
			w.F32(.1).F32(1.2).F32(-.2)
		}
		if action != 1 {
			w.VarInt(int32(hand))
		}
		w.Bool(secondary)
		if !bytes.Equal(w.Payload(), raw) {
			t.Fatal("packet codec differs", line)
		}
		cases++
	}
	if cases != 12 || proto.CInteractEntity != 0x16 {
		t.Fatal("missing original packet vectors", cases)
	}
	for _, payload := range [][]byte{{}, proto.NewWriter().VarInt(77).VarInt(7).Payload(), proto.NewWriter().VarInt(77).VarInt(2).F32(0).F32(1).F32(0).VarInt(2).Bool(false).Payload()} {
		if _, ok := readEntityInteraction(payload); ok {
			t.Fatal("malformed entity action accepted")
		}
	}
}

func armorStandServer(t *testing.T) (*Server, *player.Player, *player.Player, world.WorldEntity, string) {
	t.Helper()
	dir := t.TempDir()
	seed := int64(0)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("../world/village_actor_defaults_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var data nbt.Object
	for _, line := range strings.Split(string(source), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && f[0] == "minecraft:armor_stand" {
			raw, _ := hex.DecodeString(f[1])
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			data, _ = tag.Object()
			break
		}
	}
	if data == nil {
		t.Fatal("no official armor stand default")
	}
	uuid := [16]byte{9}
	data["UUID"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{0x09000000, 0, 0, 0}}
	data["Pos"] = nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: []nbt.Tag{{Type: nbt.Double, Value: -.5}, {Type: nbt.Double, Value: float64(-60)}, {Type: nbt.Double, Value: .5}}}
	data["ShowArms"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	armor := []nbt.Tag{{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:iron_boots"}, "count": {Type: nbt.Int, Value: int32(1)}, "ForeignItemField": {Type: nbt.String, Value: "kept"}}}}
	for i := 0; i < 3; i++ {
		armor = append(armor, nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}})
	}
	data["ArmorItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: armor}
	root := nbt.Object{"DataVersion": {Type: nbt.Int, Value: int32(3955)}, "Position": {Type: nbt.IntArray, Value: []int32{-1, 0}}, "Entities": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: data}}}}
	if err := anvil.WriteChunk(filepath.Join(dir, "entities"), -1, 0, "", nbt.Tag{Type: nbt.Compound, Value: root}); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(h, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	if _, err := w.LoadChunk(-1, 0); err != nil {
		t.Fatal(err)
	}
	s := New(Config{}, h, w)
	s.nextEID.Store(100)
	p := player.New("First", [16]byte{1}, 1, nil)
	other := player.New("Second", [16]byte{2}, 2, nil)
	for _, p := range []*player.Player{p, other} {
		p.SetPos(.5, -60, .5, 0, 0, true)
		p.SetPlayReady(true)
		p.MarkChunkSent(-1, 0)
		h.Join(p)
		s.updateGeneratedEntityVisibility(p)
		for len(p.Out) > 0 {
			<-p.Out
		}
	}
	entity := world.WorldEntity{UUID: uuid, Type: "minecraft:armor_stand", Position: [3]float64{-.5, -60, .5}, Data: data}
	return s, p, other, entity, dir
}

func armorStandClickPayload(id int32, hand int32, y float32) []byte {
	return proto.NewWriter().VarInt(id).VarInt(2).F32(0).F32(y).F32(0).VarInt(hand).Bool(false).Payload()
}

func assertEmptyArmorStandEquipment(t *testing.T, p *player.Player, id int32) {
	t.Helper()
	seen := false
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID != proto.SSetEquipment {
			continue
		}
		r := proto.NewReader(packet.Payload)
		entity, err := r.VarInt()
		if err != nil {
			t.Fatal(err)
		}
		if entity != id {
			continue
		}
		seen = true
		for i := 0; i < 7; i++ {
			index, err := r.U8()
			if err != nil || int(index&0x7f) != i || (index&0x80 != 0) != (i < 6) {
				t.Fatal("equipment slot list wrong", index, err)
			}
			s, err := item.Read(r)
			if err != nil || !s.Empty() {
				t.Fatal("empty equipment not synchronized", s, err)
			}
		}
		if r.Remaining() != 0 {
			t.Fatal("trailing equipment bytes")
		}
	}
	if !seen {
		t.Fatal("viewer did not receive final equipment removal")
	}
}

func TestArmorStandTwoPlayerExchangeEquipmentAndPlayerRestart(t *testing.T) {
	s, p, other, entity, dir := armorStandServer(t)
	key := generatedEntityKey{s.World.DimensionName(), entity.UUID}
	id := s.generatedViews[p][key]
	var wg sync.WaitGroup
	for _, p := range []*player.Player{p, other} {
		wg.Add(1)
		go func(p *player.Player) {
			defer wg.Done()
			s.handlePlayPacket(p, proto.CInteractEntity, armorStandClickPayload(id, 0, .2))
		}(p)
	}
	wg.Wait()
	taken := 0
	var owner *player.Player
	var stack item.Stack
	for _, p := range []*player.Player{p, other} {
		inventory, _ := p.Inventory()
		if !inventory.Slots[36].Empty() {
			taken++
			owner = p
			stack = inventory.Slots[36]
		}
		assertEmptyArmorStandEquipment(t, p, id)
	}
	if taken != 1 {
		t.Fatalf("%d players received the same stand item", taken)
	}
	loaded, err := s.World.LoadInventory(owner.UUID)
	_, originalNBT, originalErr := nbt.Parse(stack.Saved)
	_, savedNBT, savedErr := nbt.Parse(loaded.Slots[36].Saved)
	if err != nil || originalErr != nil || savedErr != nil || !stack.SameItem(loaded.Slots[36]) || !reflect.DeepEqual(originalNBT, savedNBT) {
		t.Fatal("player swap was not saved immediately", err)
	}
	if err := s.World.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := world.Open(nil, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err = reopened.LoadInventory(owner.UUID)
	if err != nil || !stack.SameItem(loaded.Slots[36]) {
		t.Fatal("equipment lost from player after restart", err)
	}
	actors, err := reopened.EntitiesInChunk(-1, 0)
	if err != nil || len(actors) != 1 {
		t.Fatal(actors, err)
	}
	if payload, err := generatedEntityEquipment(id, actors[0]); err != nil || len(payload) != 0 {
		t.Fatal("stand equipment returned after restart", err)
	}
}

func TestArmorStandPacketAuthorityOffhandAndPublicationOrder(t *testing.T) {
	s, p, other, entity, _ := armorStandServer(t)
	key := generatedEntityKey{s.World.DimensionName(), entity.UUID}
	id := s.generatedViews[p][key]
	// These packets must not mutate authoritative actor or player items.
	s.handlePlayPacket(p, proto.CInteractEntity, armorStandClickPayload(id+999, 0, .2))
	s.handlePlayPacket(p, proto.CInteractEntity, proto.NewWriter().VarInt(id).VarInt(0).VarInt(0).Bool(false).Payload())
	s.handlePlayPacket(p, proto.CInteractEntity, proto.NewWriter().VarInt(id).VarInt(1).Bool(false).Payload())
	p.SetPos(100, -60, .5, 0, 0, true)
	s.handlePlayPacket(p, proto.CInteractEntity, armorStandClickPayload(id, 0, .2))
	p.SetPos(.5, -60, .5, 0, 0, true)
	p.UnmarkChunkSent(-1, 0)
	s.handlePlayPacket(p, proto.CInteractEntity, armorStandClickPayload(id, 0, .2))
	p.MarkChunkSent(-1, 0)
	for len(p.Out) > 0 {
		<-p.Out
	}
	inventory, _ := p.Inventory()
	if !inventory.Slots[36].Empty() {
		t.Fatal("invalid interaction created item")
	}
	s.handlePlayPacket(p, proto.CInteractEntity, armorStandClickPayload(id, 1, .2))
	inventory, _ = p.Inventory()
	if inventory.Slots[45].Empty() || !inventory.Slots[36].Empty() {
		t.Fatal("offhand interaction affected wrong hand")
	}
	assertEmptyArmorStandEquipment(t, p, id)
	assertEmptyArmorStandEquipment(t, other, id)
	current := s.generatedEntities[key].snapshot
	s.publishGeneratedEntityEquipment(key, entity) // stale pre-swap publication
	if s.generatedEntities[key].snapshot.Revision != current.Revision || len(p.Out) != 0 || len(other.Out) != 0 {
		t.Fatal("old snapshot overwrote newer actor equipment")
	}
	viewer := player.New("Later", [16]byte{3}, 3, nil)
	viewer.SetPos(.5, -60, .5, 0, 0, true)
	viewer.SetPlayReady(true)
	viewer.MarkChunkSent(-1, 0)
	s.Hub.Join(viewer)
	s.updateGeneratedEntityVisibility(viewer)
	for len(viewer.Out) > 0 {
		packet := <-viewer.Out
		if packet.ID == proto.SSetEquipment {
			t.Fatal("new viewer received stale equipment")
		}
	}
}

func TestArmorStandInteractionWhileChestOpenResynchronizesMenu(t *testing.T) {
	s, p, other, entity, _ := armorStandServer(t)
	key := generatedEntityKey{s.World.DimensionName(), entity.UUID}
	id := s.generatedViews[p][key]
	s.containers.players = map[*player.Player]*chestSession{
		p: {window: 3, state: 7, world: s.World},
	}
	s.handlePlayPacket(p, proto.CInteractEntity, armorStandClickPayload(id, 1, .2))
	inventory, _ := p.Inventory()
	if !inventory.Slots[45].Empty() || s.generatedEntities[key].snapshot.Revision != 0 || len(other.Out) != 0 {
		t.Fatal("entity interaction changed equipment while a container owned the inventory view")
	}
	if len(p.Out) != 1 {
		t.Fatal("current menu was not resynchronized")
	}
	packet := <-p.Out
	r := proto.NewReader(packet.Payload)
	window, _ := r.U8()
	state, _ := r.VarInt()
	count, _ := r.VarInt()
	if packet.ID != proto.SContainerContent || window != 3 || state != 7 || count != 63 {
		t.Fatal("wrong menu resynchronization", packet.ID, window, state, count)
	}
	delete(s.containers.players, p)
	s.handlePlayPacket(p, proto.CInteractEntity, armorStandClickPayload(id, 1, .2))
	inventory, _ = p.Inventory()
	if inventory.Slots[45].Empty() {
		t.Fatal("closed menu continued blocking equipment interaction")
	}
}
