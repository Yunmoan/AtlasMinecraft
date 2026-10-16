package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"atlas/internal/worldgen"
	"bytes"
	"math"
	"path/filepath"
	"sort"
	"testing"
)

// This exercises the production chunk packet sink and actor publication with
// real DIM1 generation. It does not substitute for a graphical client check.
func TestEndCityActorsPacketSinkLifecycle(t *testing.T) {
	dir, seed := filepath.Join(t.TempDir(), "world"), int64(1234)
	if err := world.CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	overworld, err := world.Open(hub.DimensionView{Hub: h, Dimension: "minecraft:overworld"}, 4, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { overworld.Close() })
	end, err := world.OpenDimension(hub.DimensionView{Hub: h, Dimension: "minecraft:the_end"}, 4, dir, "end")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { end.Close() })
	s := New(Config{ViewDistance: 2, MaxPlayers: 10}, h, overworld)
	s.nextEID.Store(100)
	if err := s.AddWorld(end); err != nil {
		t.Fatal(err)
	}
	noise, err := worldgen.NewDimensionNoiseGenerator(seed, "end")
	if err != nil {
		t.Fatal(err)
	}
	biomes, err := worldgen.NewDimensionBiomeSampler(seed, "end")
	if err != nil {
		t.Fatal(err)
	}
	start, err := worldgen.GenerateEndCityStructureStart(seed, 145, -77, biomes.BiomeAtQuart, func(x, z int32) (int32, error) {
		column, err := noise.BaseStructureColumnAt(x, z)
		if err != nil {
			return 0, err
		}
		return column.OccupiedHeight(false), nil
	})
	if err != nil || start == nil || len(start.Pieces) != 94 {
		t.Fatal("pinned normal End City unavailable", err)
	}
	positions := map[[2]int32]bool{{145, -77}: true}
	for _, piece := range start.Pieces {
		if name, _ := piece.Data.Text("Template"); name == "ship" {
			for x := piece.Box.MinX >> 4; x <= piece.Box.MaxX>>4; x++ {
				for z := piece.Box.MinZ >> 4; z <= piece.Box.MaxZ>>4; z++ {
					positions[[2]int32{x, z}] = true
				}
			}
		}
	}
	var chunks [][2]int32
	for pos := range positions {
		chunks = append(chunks, pos)
	}
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i][0] != chunks[j][0] {
			return chunks[i][0] < chunks[j][0]
		}
		return chunks[i][1] < chunks[j][1]
	})
	actors := map[[16]byte]world.WorldEntity{}
	counts := map[string]int{}
	for _, pos := range chunks {
		if _, err := end.LoadChunk(pos[0], pos[1]); err != nil {
			t.Fatal(err)
		}
		entities, err := end.EntitiesInChunk(pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		for _, entity := range entities {
			if entity.Type != "minecraft:shulker" && entity.Type != "minecraft:item_frame" {
				t.Fatal("unexpected actor in pinned End City", entity.Type)
			}
			if _, duplicate := actors[entity.UUID]; duplicate {
				t.Fatal("cross-chunk duplicate actor UUID")
			}
			actors[entity.UUID] = entity
			counts[entity.Type]++
		}
	}
	if len(chunks) != 7 || counts["minecraft:shulker"] != 4 || counts["minecraft:item_frame"] != 1 {
		t.Fatal("pinned ship actor coverage changed", len(chunks), counts)
	}
	newViewer := func(name string, uuid [16]byte, id int32) *player.Player {
		p := player.New(name, uuid, id, nil)
		p.SetInitialDimension("minecraft:the_end")
		p.SetPos(2212.5, 140, -1242.5, 0, 0, true)
		p.SetPlayReady(true)
		if !h.Join(p) {
			t.Fatal("viewer join failed")
		}
		return p
	}
	p := newViewer("EndCityA", [16]byte{1}, 1)
	q := newViewer("EndCityB", [16]byte{2}, 2)
	a, b := newEndCityPacketView(), newEndCityPacketView()
	for _, viewer := range []*player.Player{p, q} {
		s.updateGeneratedEntityVisibility(viewer)
		if len(viewer.Out) != 0 || len(s.generatedEntities) != 0 {
			t.Fatal("actor published before chunk delivery")
		}
	}
	deliver := func(viewer *player.Player, view *endCityPacketView) {
		for _, pos := range chunks {
			if !viewer.ReserveChunk(pos[0], pos[1]) {
				t.Fatal("chunk reservation failed", pos)
			}
			if err := end.SpawnChunk(viewer.WorldPacketSink(), pos[0], pos[1], func(int32, int32) {}); err != nil {
				t.Fatal(err)
			}
			s.updateGeneratedEntityVisibility(viewer)
			view.consume(t, viewer, actors)
		}
		if len(view.spawned) != 5 || len(view.metadata) != 5 || len(view.chunks) != 7 {
			t.Fatal("incomplete delivered actors", len(view.spawned), len(view.metadata), len(view.chunks))
		}
	}
	deliver(p, a)
	deliver(q, b)
	assertShared := func() {
		for uuid, id := range a.spawned {
			entry := s.generatedEntities[generatedEntityKey{"minecraft:the_end", uuid}]
			if b.spawned[uuid] != id || entry == nil || entry.id != id || entry.viewers != 2 || !p.IsVisible(id) || !q.IsVisible(id) {
				t.Fatal("two viewers disagree on actor identity", uuid)
			}
		}
	}
	assertShared()
	s.updateGeneratedEntityVisibility(p)
	s.updateGeneratedEntityVisibility(q)
	if len(p.Out) != 0 || len(q.Out) != 0 {
		t.Fatal("replayed chunk visibility duplicated spawn or metadata")
	}
	oldSink := p.WorldPacketSink()
	// These are the lifecycle primitives used by TeleportToDimension. Avoid
	// requesting unrelated chunks around a destination in this focused test.
	s.entityMu.Lock()
	s.clearGeneratedEntityView(p, true)
	p.ResetWorld("minecraft:overworld", [3]float64{.5, -60, .5}, 0, 0, s.respawnPayload(p, overworld))
	s.entityMu.Unlock()
	a.consume(t, p, actors)
	if a.destroyed != 5 || len(a.spawned) != 0 || oldSink.Current() {
		t.Fatal("dimension exit retained End City actors")
	}
	oldSink.SendPacket(proto.SEntityMetadata, []byte{0xde})
	if len(p.Out) != 0 || len(q.Out) != 0 {
		t.Fatal("stale world epoch published actors or affected other viewer")
	}
	for _, entry := range s.generatedEntities {
		if entry.viewers != 1 || !q.IsVisible(entry.id) {
			t.Fatal("remaining viewer lost End City actor")
		}
	}
	p.ResetWorld("minecraft:the_end", [3]float64{2212.5, 140, -1242.5}, 0, 0, s.respawnPayload(p, end))
	a.consume(t, p, actors)
	deliver(p, a)
	assertShared()
	// Disconnect and create a fresh play-phase player while the other client
	// remains present. UUIDs and runtime IDs must still match its actor view.
	s.entityMu.Lock()
	s.clearGeneratedEntityView(p, false)
	s.entityMu.Unlock()
	h.Leave(p)
	p.Disconnect()
	p = newViewer("EndCityA", [16]byte{1}, 3)
	a = newEndCityPacketView()
	deliver(p, a)
	assertShared()
	for _, viewer := range []*player.Player{p, q} {
		s.entityMu.Lock()
		s.clearGeneratedEntityView(viewer, true)
		s.entityMu.Unlock()
	}
	a.consume(t, p, actors)
	b.consume(t, q, actors)
	if len(s.generatedEntities) != 0 || len(s.generatedViews) != 0 || len(a.spawned) != 0 || len(b.spawned) != 0 {
		t.Fatal("final viewer leaked End City actor tracking")
	}
	t.Log("normal DIM1 seed1234: 7 chunks, 4 shulkers and 1 elytra frame decoded through two packet sinks; chunk-before-spawn, original metadata/IDs, deduplication, dimension exit/reentry, reconnect and final release passed")
}

type endCityPacketView struct {
	chunks    map[[2]int32]bool
	spawned   map[[16]byte]int32
	metadata  map[[16]byte]bool
	destroyed int
}

func newEndCityPacketView() *endCityPacketView {
	return &endCityPacketView{chunks: map[[2]int32]bool{}, spawned: map[[16]byte]int32{}, metadata: map[[16]byte]bool{}}
}

func (v *endCityPacketView) consume(t *testing.T, p *player.Player, actors map[[16]byte]world.WorldEntity) {
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
			if v.chunks[[2]int32{x, z}] {
				t.Fatal("duplicate delivered chunk")
			}
			v.chunks[[2]int32{x, z}] = true
		case proto.SSpawnEntity:
			id, err := r.VarInt()
			check(err)
			uuid, err := r.UUID()
			check(err)
			typeID, err := r.VarInt()
			check(err)
			actor, exists := actors[uuid]
			if !exists || v.spawned[uuid] != 0 {
				t.Fatal("unknown or duplicate spawned actor", uuid)
			}
			var pos [3]float64
			for i := range pos {
				pos[i], err = r.F64()
				check(err)
			}
			pitch, err := r.I8()
			check(err)
			yaw, err := r.I8()
			check(err)
			headYaw, err := r.I8()
			check(err)
			data, err := r.VarInt()
			check(err)
			for range 3 {
				velocity, err := r.I16()
				if err != nil || velocity != 0 {
					t.Fatal("invalid structure actor velocity", velocity, err)
				}
			}
			if r.Remaining() != 0 || pitch != 0 {
				t.Fatal("invalid actor spawn framing or pitch")
			}
			if actor.Type == "minecraft:shulker" {
				if typeID != 88 || pos != actor.Position || data != 0 || yaw != 0 || headYaw != 0 {
					t.Fatal("shulker differs from original spawn defaults")
				}
			} else {
				for i, key := range []string{"TileX", "TileY", "TileZ"} {
					anchor, ok := actor.Data.Number(key)
					if !ok || pos[i] != float64(anchor) {
						t.Fatal("item frame spawn used bbox center instead of hanging anchor")
					}
				}
				facing, _ := actor.Data.Number("Facing")
				wantYaw, ok := map[int64]int8{2: -128, 3: 0, 4: 64, 5: -64}[facing]
				if typeID != 60 || data != int32(facing) || headYaw != 0 || !ok || yaw != wantYaw {
					t.Fatal("item frame differs from original Facing/yaw/head yaw", facing, yaw, headYaw)
				}
			}
			chunk := [2]int32{int32(math.Floor(pos[0] / 16)), int32(math.Floor(pos[2] / 16))}
			if !v.chunks[chunk] {
				t.Fatal("actor packet preceded its delivered chunk", chunk)
			}
			v.spawned[uuid] = id
		case proto.SEntityMetadata:
			id, err := r.VarInt()
			check(err)
			var uuid [16]byte
			for candidate, eid := range v.spawned {
				if id == eid {
					uuid = candidate
				}
			}
			actor, exists := actors[uuid]
			if !exists || v.metadata[uuid] {
				t.Fatal("metadata preceded spawn or duplicated initial metadata")
			}
			// Original JAR serializer output after the entity ID: Shulker
			// Health30, or one Elytra(id773) with no component overrides.
			body := []byte{9, 3, 0x41, 0xf0, 0, 0, 0xff}
			if actor.Type == "minecraft:item_frame" {
				body = []byte{8, 7, 1, 0x85, 0x06, 0, 0, 0xff}
			}
			if !bytes.Equal(packet.Payload[len(packet.Payload)-r.Remaining():], body) {
				t.Fatalf("actual generated %s metadata differs from JAR: %x", actor.Type, packet.Payload)
			}
			v.metadata[uuid] = true
		case proto.SEntityDestroy:
			count, err := r.VarInt()
			check(err)
			for range count {
				id, err := r.VarInt()
				check(err)
				found := false
				for uuid, eid := range v.spawned {
					if id == eid {
						delete(v.spawned, uuid)
						delete(v.metadata, uuid)
						v.destroyed++
						found = true
						break
					}
				}
				if !found {
					t.Fatal("destroy packet references an unspawned actor", id)
				}
			}
			if r.Remaining() != 0 {
				t.Fatal("destroy packet trailing bytes")
			}
		case proto.SRespawn:
			if len(v.spawned) != 0 {
				t.Fatal("respawn preceded old actor destruction")
			}
			clear(v.chunks)
		}
	}
}
