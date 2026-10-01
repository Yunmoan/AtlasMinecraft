package server

import (
	"atlas/internal/hub"
	"atlas/internal/logging"
	"atlas/internal/player"
	"atlas/internal/proto"
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestPlayerChatBroadcastsAndAppearsInReadableLog(t *testing.T) {
	var log bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(logging.NewVanillaHandler(&log, slog.LevelInfo, false)))
	defer slog.SetDefault(previous)
	h := hub.New()
	a := player.New("Alice", [16]byte{1}, 1, nil)
	b := player.New("Bob", [16]byte{2}, 2, nil)
	if !h.Join(a) || !h.Join(b) {
		t.Fatal("could not join test players")
	}
	s := New(Config{}, h, nil)
	if !s.handlePlayPacket(a, proto.CChatMessage, proto.NewWriter().String("hello world").Payload()) {
		t.Fatal("chat packet disconnected player")
	}
	for _, p := range []*player.Player{a, b} {
		select {
		case packet := <-p.Out:
			if packet.ID != proto.SSystemChat || chatPlainText(t, packet.Payload) != "<Alice> hello world" {
				t.Fatalf("unexpected chat packet: %+v", packet)
			}
		default:
			t.Fatalf("%s received no chat", p.Name)
		}
	}
	if !strings.Contains(log.String(), "[Server thread/INFO]: <Alice> hello world") {
		t.Fatalf("chat missing from server log: %s", log.String())
	}
}
