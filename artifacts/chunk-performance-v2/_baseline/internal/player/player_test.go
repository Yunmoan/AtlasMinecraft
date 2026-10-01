package player

import "testing"

func TestNegativeFractionalPositionUsesNegativeChunk(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	p.SetPos(-0.5, 70, -16.5, 0, 0, true)
	cx, cz := p.ChunkPos()
	if cx != -1 || cz != -2 {
		t.Fatalf("chunk = (%d,%d), want (-1,-2)", cx, cz)
	}
}
