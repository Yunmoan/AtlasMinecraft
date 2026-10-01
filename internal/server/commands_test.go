package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"strings"
	"testing"
)

func TestDeclaredCommandsAndWorldCommands(t *testing.T) {
	readCommandTree(t, declareCommandsPayload(false))
	readCommandTree(t, declareCommandsPayload(true))
	h := hub.New()
	w := world.New(h, 1)
	s := New(Config{MaxPlayers: 20}, h, w)
	if got := s.executeCommand(nil, "time set night"); !strings.Contains(got, "13000") {
		t.Fatal(got)
	}
	if got := s.executeCommand(nil, "weather thunder"); got != "Weather set to thunder" {
		t.Fatal(got)
	}
	if w.Weather() != "thunder" {
		t.Fatal("weather was not changed")
	}
	if got := s.executeCommand(nil, "time query"); !strings.Contains(got, "13000") {
		t.Fatal(got)
	}
}

func TestTeleportToOnlinePlayerAndFromConsole(t *testing.T) {
	s, source := interactionServer(t)
	s.Cfg.Ops = map[string]bool{strings.ToLower(source.Name): true}
	destination := player.New("Target", [16]byte{2}, 2, nil)
	destination.SetPos(7.25, 1024, -4.5, 130, -30, false)
	if !s.Hub.Join(destination) {
		t.Fatal("could not join destination player")
	}
	for _, test := range []struct {
		sender  *player.Player
		command string
	}{{source, "tp Target"}, {nil, "tp Builder Target"}, {source, "minecraft:teleport @s Target"}} {
		source.SetPos(.5, -60, .5, 0, 0, true)
		if got := s.executeCommand(test.sender, test.command); !strings.HasPrefix(got, "Teleported Builder to 7.25 1024.00 -4.50") {
			t.Fatalf("%s: %s", test.command, got)
		}
		x, y, z, yaw, pitch, _ := source.Pos()
		if x != 7.25 || y != 1024 || z != -4.5 || yaw != 130 || pitch != -30 {
			t.Fatalf("player teleport lost destination position/rotation: %v", []float64{x, y, z, float64(yaw), float64(pitch)})
		}
	}
	position, yaw, pitch := [3]float64{7.25, 1024, -4.5}, float32(130), float32(-30)
	source.ReserveChunk(2, 0)
	s.RunCommand(source, "tp 1000000000 ~ 1000000000")
	var response string
	for len(source.Out) > 0 {
		packet := <-source.Out
		if packet.ID == proto.SSystemChat {
			response = chatPlainText(t, packet.Payload)
		}
	}
	if !strings.Contains(response, "X/Z must be >= -30000000 and < 30000000") {
		t.Fatalf("out-of-bounds teleport did not explain vanilla limits: %s", response)
	}
	x, y, z, gotYaw, gotPitch, _ := source.Pos()
	if [3]float64{x, y, z} != position || gotYaw != yaw || gotPitch != pitch || !source.WantsChunk(2, 0) {
		t.Fatal("rejected teleport changed position or canceled terrain requests")
	}
}

func TestTeleportRestoresViewRequestsWhenRemainingInSameChunk(t *testing.T) {
	for _, args := range [][]string{{".5", "-60", ".5"}, {"30000000", "-60", "30000000"}} {
		s, p := interactionServer(t)
		p.SetViewDistance(2)
		p.ReserveChunk(2, 0)
		result := s.commandTeleport(p, args)
		if !p.WantsChunk(2, 0) && !p.IsChunkSent(2, 0) {
			t.Fatalf("teleport lost still-visible chunk requests: %s", result)
		}
		x, _, z, _, _, _ := p.Pos()
		if x != .5 || z != .5 {
			t.Fatal("failed destination preparation moved the player")
		}
	}
}
