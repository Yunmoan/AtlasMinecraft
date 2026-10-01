package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"testing"
)

func TestClientSkinPartsAndPlayerMetadata(t *testing.T) {
	payload := proto.NewWriter().String("zh_cn").U8(12).
		VarInt(0).Bool(true).U8(0x63).
		VarInt(1).Bool(false).Bool(true).VarInt(0).Payload()
	view, parts, err := clientSettings(payload)
	if err != nil || view != 12 || parts != 0x63 {
		t.Fatalf("client settings = view %d, skin parts %#x, error %v", view, parts, err)
	}
	if _, _, err := clientSettings(proto.NewWriter().String("zh_cn").U8(12).Payload()); err == nil {
		t.Fatal("truncated client settings accepted")
	}
	p := player.New("Tester", [16]byte{}, 42, nil)
	if !p.SetSkinParts(parts) || p.SkinParts() != parts {
		t.Fatal("skin parts were not stored")
	}
	r := proto.NewReader(playerSkinMetadata(p))
	eid, _ := r.VarInt()
	index, _ := r.U8()
	dataType, _ := r.VarInt()
	mask, _ := r.U8()
	end, _ := r.U8()
	if eid != 42 || index != 17 || dataType != 0 || mask != 0x63 || end != 0xff || r.Remaining() != 0 {
		t.Fatalf("invalid skin metadata: eid=%d index=%d type=%d mask=%#x end=%#x remaining=%d", eid, index, dataType, mask, end, r.Remaining())
	}
}

func TestVisiblePlayerSpawnIncludesSkinLayers(t *testing.T) {
	h := hub.New()
	s := New(Config{ViewDistance: 10}, h, nil)
	a := player.New("Alice", [16]byte{1}, 1, nil)
	b := player.New("Bob", [16]byte{2}, 2, nil)
	a.SetPos(0, 64, 0, 0, 0, true)
	b.SetPos(1, 64, 1, 0, 0, true)
	a.SetPlayReady(true)
	b.SetPlayReady(true)
	a.MarkChunkSent(0, 0)
	b.MarkChunkSent(0, 0)
	a.SetSkinParts(0x42)
	if !h.Join(a) || !h.Join(b) {
		t.Fatal("players did not join the hub")
	}
	s.updateVisibility(a)
	for _, test := range []struct {
		viewer *player.Player
		entity *player.Player
	}{{a, b}, {b, a}} {
		spawn := <-test.viewer.Out
		metadata := <-test.viewer.Out
		pose := <-test.viewer.Out
		if pose.ID != proto.SEntityMetadata {
			t.Fatal("missing pose metadata")
		}
		if spawn.ID != proto.SSpawnEntity || metadata.ID != proto.SEntityMetadata {
			t.Fatalf("%s saw packets %#x, %#x instead of spawn then metadata", test.viewer.Name, spawn.ID, metadata.ID)
		}
		r := proto.NewReader(metadata.Payload)
		eid, _ := r.VarInt()
		index, _ := r.U8()
		dataType, _ := r.VarInt()
		mask, _ := r.U8()
		if eid != test.entity.EID || index != 17 || dataType != 0 || mask != test.entity.SkinParts() {
			t.Fatalf("%s saw wrong skin layers for %s: eid=%d index=%d type=%d mask=%#x", test.viewer.Name, test.entity.Name, eid, index, dataType, mask)
		}
	}
}
