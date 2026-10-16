package server

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedPassengerVisibilityAndPacketOrdering(t *testing.T) {
	s, p, dir := dimensionTestServer(t)
	// The helper inserts players with explicit test EIDs instead of using the
	// runtime allocator; reserve them before allocating generated actor IDs.
	s.nextEID.Store(100)
	fixture, err := os.ReadFile("../../docs/worldgen/structure_drowned_actors_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var vehicle nbt.Tag
	for _, line := range strings.Split(string(fixture), "\n") {
		f := strings.Fields(line)
		if len(f) != 4 || f[0] != "drowned_actor_extra" {
			continue
		}
		raw, err := hex.DecodeString(f[3])
		if err != nil {
			t.Fatal(err)
		}
		vehicle, _, err = nbt.ParseNetwork(raw)
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	root, ok := vehicle.Object()
	if !ok {
		t.Fatal("missing actual saved jockey")
	}
	passengers, _ := root["Passengers"].Elements()
	child, _ := passengers[0].Object()
	child["UUID"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{9, 8, 7, 6}}
	entityRoot := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"DataVersion": {Type: nbt.Int, Value: int32(3955)}, "Position": {Type: nbt.IntArray, Value: []int32{-1, 1}}, "Entities": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{vehicle}}}}
	if err := anvil.WriteChunk(filepath.Join(dir, "entities"), -1, 1, "", entityRoot); err != nil {
		t.Fatal(err)
	}
	p.SetPos(-15.5, 50, 16.5, 0, 0, true)
	s.updateGeneratedEntityVisibility(p)
	if len(s.generatedEntities) != 0 {
		t.Fatal("jockey published before visible chunk")
	}
	p.MarkChunkSent(-1, 1)
	s.updateGeneratedEntityVisibility(p)
	if len(s.generatedEntities) != 2 {
		t.Fatal("vehicle/passenger tracking missing", len(s.generatedEntities))
	}
	seen := map[int32]bool{}
	mounts := 0
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID == proto.SSpawnEntity {
			id, err := proto.NewReader(packet.Payload).VarInt()
			if err != nil {
				t.Fatal(err)
			}
			seen[id] = true
		}
		if packet.ID == proto.SSetPassengers {
			r := proto.NewReader(packet.Payload)
			vehicle, err := r.VarInt()
			if err != nil {
				t.Fatal(err)
			}
			count, err := r.VarInt()
			if err != nil || count != 1 {
				t.Fatal(count, err)
			}
			child, err := r.VarInt()
			if err != nil || !seen[vehicle] || !seen[child] || r.Remaining() != 0 {
				t.Fatal("mount preceded actor spawn", seen, vehicle, child, err)
			}
			mounts++
		}
	}
	if mounts != 1 {
		t.Fatal("expected one mounted passenger packet", mounts)
	}
	q := player.New("JockeyViewer", [16]byte{4}, 4, nil)
	q.SetPlayReady(true)
	q.SetPos(-15.5, 50, 16.5, 0, 0, true)
	q.MarkChunkSent(-1, 1)
	s.Hub.Join(q)
	s.updateGeneratedEntityVisibility(q)
	jockeys := make(map[generatedEntityKey]*generatedEntityTracking)
	for key, entry := range s.generatedEntities {
		jockeys[key] = entry
		if entry.viewers != 2 || !p.IsVisible(entry.id) || !q.IsVisible(entry.id) {
			t.Fatal("viewers received inconsistent jockey IDs")
		}
	}
	s.updateGeneratedEntityVisibility(p)
	for len(p.Out) > 0 {
		if packet := <-p.Out; packet.ID == proto.SSetPassengers || packet.ID == proto.SSpawnEntity {
			t.Fatal("repeat chunk visibility duplicated jockey")
		}
	}
	if err := s.TeleportToDimension(p, "minecraft:the_end", [3]float64{.5, 100, .5}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range jockeys {
		if entry.viewers != 1 || p.IsVisible(entry.id) || !q.IsVisible(entry.id) {
			t.Fatalf("dimension transfer jockey id=%d viewers=%d old-visible=%v remaining-visible=%v", entry.id, entry.viewers, p.IsVisible(entry.id), q.IsVisible(entry.id))
		}
	}
	s.entityMu.Lock()
	s.clearGeneratedEntityView(q, true)
	s.entityMu.Unlock()
	for key := range jockeys {
		if s.generatedEntities[key] != nil {
			t.Fatal("final viewer retained jockey")
		}
	}
}
