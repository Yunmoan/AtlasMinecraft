package server

import (
	"atlas/internal/hub"
	"atlas/internal/proto"
	"atlas/internal/world"
	"strings"
	"testing"
)

func TestDeclaredCommandsAndWorldCommands(t *testing.T) {
	for _, operator := range []bool{false, true} {
		r := proto.NewReader(declareCommandsPayload(operator))
		count, err := r.VarInt()
		if err != nil || count <= 1 {
			t.Fatalf("operator=%v command nodes=%d err=%v", operator, count, err)
		}
		if operator && count <= 7 {
			t.Fatalf("operator tree has only %d nodes", count)
		}
		var literals []string
		for i := int32(0); i < count; i++ {
			flags, err := r.U8()
			if err != nil {
				t.Fatal(err)
			}
			if i > 0 && flags&0x04 == 0 {
				t.Fatalf("node %d is not executable", i)
			}
			children, err := r.VarInt()
			if err != nil || children < 0 || children >= count {
				t.Fatalf("node %d has invalid child count %d: %v", i, children, err)
			}
			for j := int32(0); j < children; j++ {
				child, err := r.VarInt()
				if err != nil || child <= i || child >= count {
					t.Fatalf("node %d has invalid child %d: %v", i, child, err)
				}
			}
			switch flags & 0x03 {
			case 0:
				if i != 0 {
					t.Fatal("non-root node encoded as root")
				}
			case 1:
				name, err := r.String(256)
				if err != nil {
					t.Fatal(err)
				}
				literals = append(literals, name)
			case 2:
				name, err := r.String(256)
				if err != nil || name != "args" {
					t.Fatalf("unexpected argument %q: %v", name, err)
				}
				parser, _ := r.VarInt()
				kind, _ := r.VarInt()
				if parser != 5 || kind != 2 {
					t.Fatalf("unexpected argument parser %d/%d", parser, kind)
				}
			default:
				t.Fatal("unexpected node type")
			}
		}
		root, err := r.VarInt()
		if err != nil || root != 0 || r.Remaining() != 0 {
			t.Fatalf("trailing command bytes or wrong root: root=%d remaining=%d err=%v", root, r.Remaining(), err)
		}
		if operator && !strings.Contains(strings.Join(literals, ","), "weather") {
			t.Fatal("operator command missing")
		}
		if !operator && len(literals) != 3 {
			t.Fatalf("public command literals: %v", literals)
		}
	}
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
