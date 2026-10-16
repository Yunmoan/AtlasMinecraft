package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"path/filepath"
	"strings"
	"testing"
)

func dimensionTestServer(t *testing.T) (*Server, *player.Player, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(hub.DimensionView{Hub: h, Dimension: "minecraft:overworld"}, 4, dir)
	if err != nil {
		t.Fatal(err)
	}
	s := New(Config{MaxPlayers: 20, ViewDistance: 2, Ops: map[string]bool{"builder": true}}, h, w)
	for _, name := range []string{"minecraft:the_nether", "minecraft:the_end"} {
		dimension, err := world.OpenDimension(hub.DimensionView{Hub: h, Dimension: name}, 4, dir, name)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.AddWorld(dimension); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for _, w := range s.Worlds() {
			if err := w.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	p := player.New("Builder", [16]byte{1}, 1, nil)
	p.SetViewDistance(2)
	p.SetPos(.5, -60, .5, 0, 0, true)
	p.SetPlayReady(true)
	if !h.Join(p) {
		t.Fatal("join failed")
	}
	return s, p, dir
}

func TestRuntimeDimensionsTeleportRoutingAndPlayerRestart(t *testing.T) {
	s, p, dir := dimensionTestServer(t)
	oldSink := p.WorldPacketSink()
	viewer := player.New("Overworld", [16]byte{2}, 2, nil)
	viewer.SetPos(.5, -60, .5, 0, 0, true)
	viewer.SetPlayReady(true)
	viewer.MarkChunkSent(0, 0)
	viewer.MarkVisible(p.EID)
	s.Hub.Join(viewer)
	if got := s.executeCommand(p, "execute in minecraft:the_nether run tp @s 0 200 0"); !strings.HasPrefix(got, "Teleported Builder") {
		t.Fatal(got)
	}
	if p.Dimension() != "minecraft:the_nether" || s.PlayerWorld(p).MaxBuildHeight() != 256 {
		t.Fatal("dimension did not change")
	}
	if viewer.IsVisible(p.EID) {
		t.Fatal("old dimension still tracks player entity")
	}
	p.ReserveChunk(17, -4)
	oldSink.UnmarkChunkSent(17, -4)
	if !p.WantsChunk(17, -4) || oldSink.ReserveChunk(18, -4) {
		t.Fatal("stale sink changed new dimension reservations")
	}
	currentSink := p.WorldPacketSink()
	currentSink.UnmarkChunkSent(17, -4)
	if !currentSink.ReserveChunk(17, -4) {
		t.Fatal("failed current-world request could not be retried")
	}
	if oldSink.SendRequestedChunk(17, -4, []byte{1}) {
		t.Fatal("old epoch delivered into new world's same-coordinate request")
	}
	for len(p.Out) > 0 {
		<-p.Out
	}
	oldSink.SendPacket(proto.SBlockChange, []byte{0xDE})
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID == proto.SBlockChange && len(packet.Payload) == 1 && packet.Payload[0] == 0xDE {
			t.Fatal("old light/block packets survived respawn")
		}
	}
	if got := s.executeCommand(p, "setblock 0 200 0 minecraft:stone"); !strings.HasPrefix(got, "Set block") {
		t.Fatal(got)
	}
	if got := s.executeCommand(p, "setblock 0 201 0 minecraft:torch"); !strings.HasPrefix(got, "Set block") {
		t.Fatal(got)
	}
	id, err := s.PlayerWorld(p).BlockAt(0, 201, 0)
	if err != nil {
		t.Fatal(err)
	}
	torch, _ := world.BlockStateID("minecraft:torch", nil)
	if id != torch {
		t.Fatal("block edit routed to wrong world")
	}
	if len(s.Hub.ViewersInDimension("minecraft:overworld", 0, 0)) != 1 || len(s.Hub.ViewersInDimension("minecraft:the_nether", 0, 0)) != 1 {
		t.Fatal("dimension viewer routing incorrect")
	}
	s.savePlayer(p)
	saved, exists, err := s.World.LoadPlayer(p.UUID)
	if err != nil || !exists || saved.Dimension != "minecraft:the_nether" || saved.Y != 200 {
		t.Fatal("cross-dimension player save failed", saved, err)
	}
	for _, w := range s.Worlds() {
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	}
	reopened, err := world.OpenDimension(hub.DimensionView{Hub: s.Hub, Dimension: "minecraft:the_nether"}, 2, dir, "nether")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	saved, exists, err = reopened.LoadPlayer(p.UUID)
	if err != nil || !exists || saved.Dimension != "minecraft:the_nether" {
		t.Fatal("restart lost player's dimension", err)
	}
	if _, err := reopened.LoadChunk(0, 0); err != nil {
		t.Fatal(err)
	}
	id, err = reopened.BlockAt(0, 201, 0)
	if err != nil || id != torch {
		t.Fatal("restart lost dimension edit", id, err)
	}
}

func TestDimensionTeleportPreservesDestinationRotationAndSameWorld(t *testing.T) {
	s, p, _ := dimensionTestServer(t)
	p.SetInitialDimension("minecraft:the_end")
	destination := player.New("Destination", [16]byte{3}, 3, nil)
	destination.SetPos(.5, -60, .5, 135, -25, true)
	s.Hub.Join(destination)
	if result := s.commandTeleport(p, []string{"Destination"}); !strings.HasPrefix(result, "Teleported") {
		t.Fatal(result)
	}
	_, _, _, yaw, pitch, _ := p.Pos()
	if p.Dimension() != "minecraft:overworld" || yaw != 135 || pitch != -25 {
		t.Fatal("cross-world teleport discarded destination rotation")
	}
	for len(p.Out) > 0 {
		<-p.Out
	}
	if err := s.TeleportToDimension(p, "minecraft:overworld", [3]float64{.5, -60, .5}); err != nil {
		t.Fatal(err)
	}
	for len(p.Out) > 0 {
		if packet := <-p.Out; packet.ID == proto.SRespawn {
			t.Fatal("same-world teleport discarded the client world")
		}
	}
}

func TestDimensionJoinAndRespawnPayloads(t *testing.T) {
	s, p, _ := dimensionTestServer(t)
	for _, entry := range []struct {
		name string
		id   int32
	}{{"minecraft:overworld", 0}, {"minecraft:the_end", 2}, {"minecraft:the_nether", 3}} {
		p.SetInitialDimension(entry.name)
		r := proto.NewReader(s.joinGamePayload(p))
		r.I32()
		r.Bool()
		count, err := r.VarInt()
		if err != nil || count != 3 {
			t.Fatal("login dimension list missing", count, err)
		}
		for i := int32(0); i < count; i++ {
			r.String(128)
		}
		r.VarInt()
		r.VarInt()
		r.VarInt()
		r.Bool()
		r.Bool()
		r.Bool()
		id, err := r.VarInt()
		name, nameErr := r.String(128)
		if err != nil || nameErr != nil || id != entry.id || name != entry.name {
			t.Fatalf("join dimension %s=%d/%s", entry.name, id, name)
		}
		w, _ := s.WorldInDimension(entry.name)
		r = proto.NewReader(s.respawnPayload(p, w))
		id, err = r.VarInt()
		name, nameErr = r.String(128)
		if err != nil || nameErr != nil || id != entry.id || name != entry.name {
			t.Fatal("respawn dimension mismatch")
		}
		r.I64()
		r.I8()
		r.I8()
		r.Bool()
		r.Bool()
		r.Bool()
		r.VarInt()
		flags, err := r.U8()
		if err != nil || flags != 3 || r.Remaining() != 0 {
			t.Fatal("respawn preservation flags/trailing bytes", flags, err)
		}
	}
}
