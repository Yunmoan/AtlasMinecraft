package server

import (
	"atlas/internal/hub"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func spawnPositionPackets(p *player.Player) []int64 {
	var positions []int64
	for len(p.Out) > 0 {
		packet := <-p.Out
		if packet.ID == proto.SSpawnPosition {
			position, err := proto.NewReader(packet.Payload).I64()
			if err != nil {
				panic(err)
			}
			positions = append(positions, position)
		}
	}
	return positions
}

func TestSetWorldSpawnCommand(t *testing.T) {
	s, source := interactionServer(t)
	s.Cfg.Ops = map[string]bool{strings.ToLower(source.Name): true}
	source.SetPlayReady(true)
	source.SetPos(12.5, -60, -7.5, 0, 0, true)

	if got := s.executeCommand(nil, "setworldspawn 12 -60 -8"); got != "Set the world spawn point to 12, -60, -8 [0.0]" {
		t.Fatal(got)
	}
	if x, y, z, angle := s.World.SpawnPoint(); x != 12 || y != -60 || z != -8 || angle != 0 {
		t.Fatalf("world spawn: %d %d %d %v", x, y, z, angle)
	}
	if sent := spawnPositionPackets(source); len(sent) != 1 || sent[0] != proto.PackPosition(12, -60, -8) {
		t.Fatalf("players must receive the new spawn position once: %v", sent)
	}
	// An identical spawn is stored but not broadcast again.
	if got := s.executeCommand(source, "setworldspawn 12 -60 -8"); got == "" {
		t.Fatal("repeated setworldspawn failed")
	}
	if sent := spawnPositionPackets(source); len(sent) != 0 {
		t.Fatalf("unchanged spawn was broadcast: %v", sent)
	}
	// Angle arguments wrap like vanilla and accept relative yaws.
	if got := s.executeCommand(source, "setworldspawn ~ ~ ~ -90"); got != "Set the world spawn point to 12, -60, -8 [-90.0]" {
		t.Fatal(got)
	}
	source.SetPos(12.5, -60, -7.5, 130, 0, true)
	if got := s.executeCommand(source, "setworldspawn ~ ~ ~ ~200"); got != "Set the world spawn point to 12, -60, -8 [-30.0]" {
		t.Fatal(got)
	}
	if _, _, _, angle := s.World.SpawnPoint(); angle != -30 {
		t.Fatalf("relative angle: %v", angle)
	}
	spawnPositionPackets(source)
	if got := s.executeCommand(source, "setworldspawn 5"); got != "Usage: /setworldspawn [x y z [angle]]" {
		t.Fatal(got)
	}
	if got := s.executeCommand(source, "setworldspawn 30000000 -60 0"); !strings.Contains(got, "X/Z must be >= -30000000 and < 30000000") {
		t.Fatal(got)
	}
	if got := s.executeCommand(source, "setworldspawn 0 -60 0 abc"); !strings.Contains(got, "Invalid angle") {
		t.Fatal(got)
	}
	s.Cfg.Ops = nil
	if got := s.executeCommand(source, "setworldspawn 0 -60 0"); got != "You do not have permission to use this command." {
		t.Fatal(got)
	}
}

func TestSpawnpointCommand(t *testing.T) {
	dir, seed := t.TempDir(), int64(0)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(h, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	s := New(Config{}, h, w)
	source := player.New("Builder", [16]byte{1}, 1, nil)
	target := player.New("Target", [16]byte{2}, 2, nil)
	if !h.Join(source) || !h.Join(target) {
		t.Fatal("join")
	}
	s.Cfg.Ops = map[string]bool{"builder": true}
	source.SetPos(1.5, -60, 2.5, 0, 0, true)
	target.SetPos(50.5, -60, 60.5, 0, 0, true)

	if got := s.executeCommand(source, "spawnpoint Target 3 -60 4 90"); got != "Set spawn point to 3, -60, 4 [90.0] in minecraft:overworld for Target" {
		t.Fatal(got)
	}
	assertPlayerSpawnTags(t, dir, target.UUID, 3, -60, 4, 90)
	// No position targets the source's current block position, not the
	// target's, matching vanilla SetSpawnCommand defaults.
	if got := s.executeCommand(source, "spawnpoint Target"); got != "Set spawn point to 1, -60, 2 [0.0] in minecraft:overworld for Target" {
		t.Fatal(got)
	}
	assertPlayerSpawnTags(t, dir, target.UUID, 1, -60, 2, 0)
	// Without a target the source's own position is used.
	if got := s.executeCommand(source, "spawnpoint 7 -60 9"); got != "Set spawn point to 7, -60, 9 [0.0] in minecraft:overworld for Builder" {
		t.Fatal(got)
	}
	assertPlayerSpawnTags(t, dir, source.UUID, 7, -60, 9, 0)
	if got := s.executeCommand(source, "spawnpoint"); got != "Set spawn point to 1, -60, 2 [0.0] in minecraft:overworld for Builder" {
		t.Fatal(got)
	}
	if got := s.executeCommand(nil, "spawnpoint"); got != "Usage: /spawnpoint <player> [x y z [angle]]" {
		t.Fatal(got)
	}
	if got := s.executeCommand(source, "spawnpoint Ghost"); got != "Player not found; use /list." {
		t.Fatal(got)
	}
	if got := s.executeCommand(source, "spawnpoint Target ~ ~2 ~ -200"); got != "Set spawn point to 1, -58, 2 [160.0] in minecraft:overworld for Target" {
		t.Fatal(got)
	}
	assertPlayerSpawnTags(t, dir, target.UUID, 1, -58, 2, 160)
}

func assertPlayerSpawnTags(t *testing.T, dir string, uuid [16]byte, x, y, z int32, angle float32) {
	t.Helper()
	path := filepath.Join(dir, "playerdata", fmt.Sprintf("%x-%x-%x-%x-%x.dat",
		uuid[:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	_, root, err := nbt.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := root.Object()
	number := func(key string) int64 {
		value, ok := o[key].Number()
		if !ok {
			t.Fatalf("missing %s tag: %v", key, o[key])
		}
		return value
	}
	if int32(number("SpawnX")) != x || int32(number("SpawnY")) != y || int32(number("SpawnZ")) != z {
		t.Fatalf("spawn position tags: %v", o)
	}
	storedAngle, ok := o["SpawnAngle"].Value.(float32)
	if !ok || storedAngle != angle {
		t.Fatalf("spawn angle tag: %v", o["SpawnAngle"])
	}
	if forced, _ := o["SpawnForced"].Number(); forced != 1 {
		t.Fatalf("spawn forced tag: %v", o["SpawnForced"])
	}
	if dimension, _ := o["SpawnDimension"].Text(); dimension != "minecraft:overworld" {
		t.Fatalf("spawn dimension tag: %q", dimension)
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
