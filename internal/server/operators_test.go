package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOperatorCommandsPersistAndRefreshOnlineTree(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.json")
	ops, err := OpenOperators(path)
	if err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w := world.New(h, 1)
	s := New(Config{MaxPlayers: 20, Operators: ops}, h, w)
	p := player.New("Alice", offlineUUID("Alice"), 1, nil)
	if !h.Join(p) {
		t.Fatal("player join failed")
	}
	if s.isOperator(p) {
		t.Fatal("new player is unexpectedly an operator")
	}
	if got := s.executeCommand(nil, "op Alice"); !strings.Contains(got, "Made Alice") || !s.isOperator(p) {
		t.Fatalf("op failed: %s", got)
	}
	if packet := <-p.Out; packet.ID != proto.SDeclareCommands {
		t.Fatalf("operator command tree not refreshed: %+v", packet)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved []OperatorEntry
	if err := json.Unmarshal(b, &saved); err != nil || len(saved) != 1 || saved[0].Name != "Alice" || saved[0].Level != 4 || saved[0].UUID != formatUUID(offlineUUID("Alice")) {
		t.Fatalf("invalid persisted ops.json: %s (%v)", b, err)
	}
	reopened, err := OpenOperators(path)
	if err != nil || !reopened.Has("alice") {
		t.Fatalf("operator did not survive restart: %v", err)
	}
	if got := s.executeCommand(nil, "deop Alice"); !strings.Contains(got, "Removed Alice") || s.isOperator(p) {
		t.Fatalf("deop failed: %s", got)
	}
	if packet := <-p.Out; packet.ID != proto.SDeclareCommands {
		t.Fatalf("nonoperator command tree not refreshed: %+v", packet)
	}
	if _, err := OpenOperators(path); err != nil {
		t.Fatalf("deop left invalid ops.json: %v", err)
	}
}

func TestOperatorLevelTwoCannotChangeOperators(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.json")
	entries := []OperatorEntry{{UUID: formatUUID(offlineUUID("Alice")), Name: "Alice", Level: 2}}
	b, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
	ops, err := OpenOperators(path)
	if err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	s := New(Config{Operators: ops}, h, world.New(h, 1))
	p := player.New("Alice", offlineUUID("Alice"), 1, nil)
	if got := s.executeCommand(p, "seed"); got != "Seed: 0" {
		t.Fatalf("level 2 cannot read seed: %s", got)
	}
	if got := s.executeCommand(p, "op Bob"); !strings.Contains(got, "do not have permission") || ops.Has("Bob") {
		t.Fatalf("level 2 changed operator list: %s", got)
	}
	spoofed := player.New("Alice", [16]byte{9}, 2, nil)
	if s.operatorLevel(spoofed) != 0 {
		t.Fatal("operator granted to a mismatched UUID")
	}
	if bytes := declareCommandsPayloadForLevel(2); strings.Contains(string(bytes), "deop") {
		t.Fatal("level 2 command tree exposes deop")
	}
}

func TestOnlineOperatorUsesVerifiedUUID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.json")
	ops, err := OpenOperators(path)
	if err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	verified := [16]byte{0x80, 0x72, 0xfe, 0xc8, 0x68, 0x8f, 0x40, 0x06, 0x99, 0xea, 0x4e, 0x74, 0xde, 0xec, 0x5d, 0xd7}
	p := player.New("Yun_mo_an", verified, 1, nil)
	if !h.Join(p) {
		t.Fatal("could not join verified player")
	}
	s := New(Config{MaxPlayers: 20, OnlineMode: true, Operators: ops}, h, nil)
	if got := s.executeCommand(nil, "op OfflinePlayer"); !strings.Contains(got, "must be connected") {
		t.Fatalf("offline target accepted: %s", got)
	}
	if got := s.executeCommand(nil, "op Yun_mo_an"); !strings.Contains(got, "Made Yun_mo_an") || !s.isOperator(p) {
		t.Fatalf("verified player op failed: %s", got)
	}
	entry := ops.entries["yun_mo_an"]
	if entry.UUID != formatUUID(verified) || entry.UUID == formatUUID(offlineUUID(p.Name)) {
		t.Fatalf("operator UUID is not the verified UUID: %+v", entry)
	}
	spoofed := player.New(p.Name, offlineUUID(p.Name), 2, nil)
	if s.isOperator(spoofed) {
		t.Fatal("offline UUID received online-mode operator permission")
	}
}
