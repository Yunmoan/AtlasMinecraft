package server

import (
	"atlas/internal/anvil"
	"atlas/internal/hub"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type villagePacketActor struct {
	data      nbt.Object
	metadata  []byte
	equipment []byte
	yaw       int8
	typeID    int32
}

// Use actual authored STRUCTURE snapshots and original loaded packet codecs.
// The lightweight fixture world exercises production Anvil reads, chunk packet
// sinks and actor publication without regenerating entire village layouts.
func TestVillageActorsPacketSinkLifecycle(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/structure_village_actors_1.21.1.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	raw, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	actors := map[[16]byte]*villagePacketActor{}
	var selected bool
	var current *villagePacketActor
	decode := func(value string) []byte {
		t.Helper()
		data, err := hex.DecodeString(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	packetBody := func(value string) []byte {
		t.Helper()
		data := decode(value)
		r := proto.NewReader(data)
		id, err := r.VarInt()
		if err != nil || id != 77 {
			t.Fatal("original actor packet ID", id, err)
		}
		return data[len(data)-r.Remaining():]
	}
	for _, line := range strings.Split(string(raw), "\n") {
		f := strings.Fields(line)
		if len(f) > 0 && f[0] == "village_actor_case" {
			selected = f[4] == "0" && f[6] == "1" && (f[1] == "minecraft:village/desert/camel_spawn" || f[1] == "minecraft:village/taiga/houses/taiga_armorer_2")
			current = nil
		}
		if !selected || len(f) != 4 {
			continue
		}
		switch f[0] {
		case "village_actor_finalized":
			tag, _, err := nbt.ParseNetwork(decode(f[3]))
			if err != nil {
				t.Fatal(err)
			}
			data, _ := tag.Object()
			index := int32(len(actors) + 1)
			data["UUID"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{index, 0, 0, 0}}
			uuid := [16]byte{3: byte(index)}
			current = &villagePacketActor{data: data, yaw: 64, typeID: 14}
			if f[1] == "minecraft:armor_stand" {
				current.typeID = 3
				angles, _ := data["Rotation"].Elements()
				if angles[0].Value == float32(135) {
					current.yaw = 96
				} else if angles[0].Value == float32(180) {
					current.yaw = -128
				} else {
					t.Fatal("pinned armor stand original yaw changed")
				}
			}
			actors[uuid] = current
		case "village_actor_loaded_metadata":
			current.metadata = packetBody(f[3])
		case "village_actor_equipment":
			current.equipment = packetBody(f[3])
		}
	}
	if len(actors) != 3 {
		t.Fatal("missing original camel and two authored armor stands", len(actors))
	}
	var entities []nbt.Tag
	for _, actor := range actors {
		if len(actor.metadata) == 0 || actor.typeID == 3 && len(actor.equipment) == 0 {
			t.Fatal("incomplete original loaded packets")
		}
		entities = append(entities, nbt.Tag{Type: nbt.Compound, Value: actor.data})
	}
	dir, seed := filepath.Join(t.TempDir(), "world"), int64(1234)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	root := nbt.Object{
		"DataVersion": {Type: nbt.Int, Value: int32(3955)}, "Position": {Type: nbt.IntArray, Value: []int32{-1, 1}},
		"Entities": {Type: nbt.List, ListType: nbt.Compound, Value: entities},
	}
	if err := anvil.WriteChunk(filepath.Join(dir, "entities"), -1, 1, "", nbt.Tag{Type: nbt.Compound, Value: root}); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(hub.DimensionView{Hub: h, Dimension: "minecraft:overworld"}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	loaded, err := w.EntitiesInChunk(-1, 1)
	if err != nil || len(loaded) != 3 {
		t.Fatal("original actors did not reach production Anvil reader", len(loaded), err)
	}
	for _, entity := range loaded {
		if !entity.LoadedFromDisk || actors[entity.UUID] == nil {
			t.Fatal("Anvil actor lost loaded state or identity")
		}
	}
	s := New(Config{ViewDistance: 2, MaxPlayers: 10}, h, w)
	s.nextEID.Store(100)
	newViewer := func(id int32) *player.Player {
		p := player.New("VillageViewer"+strconv.Itoa(int(id)), [16]byte{byte(id)}, id, nil)
		p.SetInitialDimension("minecraft:overworld")
		p.SetPos(-15.5, 70, 16.5, 0, 0, true)
		p.SetPlayReady(true)
		if !h.Join(p) {
			t.Fatal("viewer join failed")
		}
		return p
	}
	p, q := newViewer(1), newViewer(2)
	a, b := newVillagePacketView(), newVillagePacketView()
	for _, viewer := range []*player.Player{p, q} {
		s.updateGeneratedEntityVisibility(viewer)
		if len(viewer.Out) != 0 || len(s.generatedEntities) != 0 {
			t.Fatal("village actor preceded terrain delivery")
		}
	}
	deliver := func(viewer *player.Player, view *villagePacketView) {
		if !viewer.ReserveChunk(-1, 1) {
			t.Fatal("chunk reservation failed")
		}
		if err := w.SpawnChunk(viewer.WorldPacketSink(), -1, 1, func(int32, int32) {}); err != nil {
			t.Fatal(err)
		}
		s.updateGeneratedEntityVisibility(viewer)
		view.consume(t, viewer, actors)
		if len(view.spawned) != 3 || len(view.metadata) != 3 || len(view.equipment) != 2 {
			t.Fatal("missing delivered village actor packets")
		}
	}
	deliver(p, a)
	deliver(q, b)
	shared := func() {
		for uuid, id := range a.spawned {
			entry := s.generatedEntities[generatedEntityKey{"minecraft:overworld", uuid}]
			if b.spawned[uuid] != id || entry == nil || entry.id != id || entry.viewers != 2 || !p.IsVisible(id) || !q.IsVisible(id) {
				t.Fatal("two viewers disagree on generated actor identity")
			}
		}
	}
	shared()
	s.updateGeneratedEntityVisibility(p)
	s.updateGeneratedEntityVisibility(q)
	if len(p.Out) != 0 || len(q.Out) != 0 {
		t.Fatal("repeated visibility duplicated spawn, metadata or equipment")
	}
	oldSink := p.WorldPacketSink()
	s.entityMu.Lock()
	s.clearGeneratedEntityView(p, true)
	p.ResetWorld("minecraft:overworld", [3]float64{1000.5, 70, 1000.5}, 0, 0, s.respawnPayload(p, w))
	s.entityMu.Unlock()
	a.consume(t, p, actors)
	if a.destroyed != 3 || len(a.spawned) != 0 || oldSink.Current() {
		t.Fatal("village exit retained actors or a live old packet sink")
	}
	oldSink.SendPacket(proto.SEntityMetadata, []byte{0xde})
	if len(p.Out) != 0 || len(q.Out) != 0 {
		t.Fatal("stale sink published an actor packet")
	}
	for _, entry := range s.generatedEntities {
		if entry.viewers != 1 || !q.IsVisible(entry.id) {
			t.Fatal("remaining viewer lost actors")
		}
	}
	p.ResetWorld("minecraft:overworld", [3]float64{-15.5, 70, 16.5}, 0, 0, s.respawnPayload(p, w))
	a.consume(t, p, actors)
	deliver(p, a)
	shared()
	s.entityMu.Lock()
	s.clearGeneratedEntityView(p, false)
	s.entityMu.Unlock()
	h.Leave(p)
	p.Disconnect()
	p, a = newViewer(3), newVillagePacketView()
	deliver(p, a)
	shared()
	for _, viewer := range []*player.Player{p, q} {
		s.entityMu.Lock()
		s.clearGeneratedEntityView(viewer, true)
		s.entityMu.Unlock()
	}
	a.consume(t, p, actors)
	b.consume(t, q, actors)
	if len(s.generatedEntities) != 0 || len(s.generatedViews) != 0 || len(a.spawned) != 0 || len(b.spawned) != 0 {
		t.Fatal("last village viewer leaked actor tracking")
	}
	t.Log("official camel and two equipped armor stands through production Anvil and two packet sinks: original loaded metadata/equipment, chunk-before-spawn, shared IDs, deduplication, exit/reentry, reconnect and final release passed")
}

type villagePacketView struct {
	chunk     bool
	spawned   map[[16]byte]int32
	metadata  map[[16]byte]bool
	equipment map[[16]byte]bool
	destroyed int
}

func newVillagePacketView() *villagePacketView {
	return &villagePacketView{spawned: map[[16]byte]int32{}, metadata: map[[16]byte]bool{}, equipment: map[[16]byte]bool{}}
}

func (v *villagePacketView) consume(t *testing.T, p *player.Player, actors map[[16]byte]*villagePacketActor) {
	t.Helper()
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	for len(p.Out) > 0 {
		packet := <-p.Out
		r := proto.NewReader(packet.Payload)
		switch packet.ID {
		case proto.SChunkData:
			x, err := r.I32()
			check(err)
			z, err := r.I32()
			check(err)
			if x != -1 || z != 1 || v.chunk {
				t.Fatal("unexpected or duplicate terrain packet")
			}
			v.chunk = true
		case proto.SSpawnEntity:
			id, err := r.VarInt()
			check(err)
			uuid, err := r.UUID()
			check(err)
			typeID, err := r.VarInt()
			check(err)
			actor := actors[uuid]
			if actor == nil || v.spawned[uuid] != 0 || !v.chunk || typeID != actor.typeID {
				t.Fatal("unknown, duplicate, premature or mistyped village actor")
			}
			for _, expected := range []float64{-15.5, 70, 16.5} {
				value, err := r.F64()
				if err != nil || value != expected {
					t.Fatal("original actor position changed", value, err)
				}
			}
			for _, expected := range []int8{0, actor.yaw, actor.yaw} {
				value, err := r.I8()
				if err != nil || value != expected {
					t.Fatal("original actor rotation changed", value, expected, err)
				}
			}
			data, err := r.VarInt()
			if err != nil || data != 0 {
				t.Fatal("unexpected actor spawn data", data, err)
			}
			for range 3 {
				velocity, err := r.I16()
				if err != nil || velocity != 0 {
					t.Fatal("unexpected initial actor velocity", velocity, err)
				}
			}
			if r.Remaining() != 0 {
				t.Fatal("trailing actor spawn bytes")
			}
			v.spawned[uuid] = id
		case proto.SEntityMetadata, proto.SSetEquipment:
			id, err := r.VarInt()
			check(err)
			var uuid [16]byte
			for candidate, spawned := range v.spawned {
				if spawned == id {
					uuid = candidate
				}
			}
			actor := actors[uuid]
			if actor == nil {
				t.Fatal("appearance packet preceded actor spawn")
			}
			want, seen := actor.metadata, v.metadata
			if packet.ID == proto.SSetEquipment {
				want, seen = actor.equipment, v.equipment
			}
			if seen[uuid] || !bytes.Equal(packet.Payload[len(packet.Payload)-r.Remaining():], want) {
				t.Fatalf("duplicate or non-original actor appearance packet: %x", packet.Payload)
			}
			seen[uuid] = true
		case proto.SEntityDestroy:
			count, err := r.VarInt()
			check(err)
			for range count {
				id, err := r.VarInt()
				check(err)
				found := false
				for uuid, spawned := range v.spawned {
					if id == spawned {
						delete(v.spawned, uuid)
						delete(v.metadata, uuid)
						delete(v.equipment, uuid)
						v.destroyed++
						found = true
						break
					}
				}
				if !found {
					t.Fatal("destroyed an actor absent from the viewer")
				}
			}
			if r.Remaining() != 0 {
				t.Fatal("trailing destroy bytes")
			}
		case proto.SRespawn:
			if len(v.spawned) != 0 {
				t.Fatal("respawn preceded actor destruction")
			}
			v.chunk = false
		}
	}
}
