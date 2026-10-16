package world

import (
	"atlas/internal/proto"
	"sync"
	"testing"
)

type orderedLightSink struct {
	mu      sync.Mutex
	packets []int32
}

func (s *orderedLightSink) SendPacket(id int32, _ []byte) {
	s.mu.Lock()
	s.packets = append(s.packets, id)
	s.mu.Unlock()
}
func (*orderedLightSink) SendChunkData([]byte) {}
func (s *orderedLightSink) drain() []int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	packets := s.packets
	s.packets = nil
	return packets
}

type orderedLightView struct {
	world *World
	sink  orderedLightSink
	err   error
}

func (v *orderedLightView) ViewersOf(x, z int32) []PacketSink {
	if x == -1 && z == 8 {
		// The neighboring column is owned by a different region actor. This
		// read barrier forces it to handle any queued light notification before
		// the source block packet is sent. No sleeps or scheduler luck needed.
		_, v.err = v.world.ReadLoadedBlocks([]BlockPosition{{0, -60, 8}})
	}
	return []PacketSink{&v.sink}
}

func TestCrossRegionBlockPacketPrecedesEveryLightUpdate(t *testing.T) {
	dir, seed := t.TempDir(), int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	view := &orderedLightView{}
	w, err := Open(view, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	view.world = w
	defer w.Close()
	for x := int32(-1); x <= 0; x++ {
		if _, err := w.LoadChunk(x, 0); err != nil {
			t.Fatal(err)
		}
	}
	barrier := func() {
		t.Helper()
		if _, err := w.ReadLoadedBlocks([]BlockPosition{{-1, -60, 8}, {0, -60, 8}}); err != nil {
			t.Fatal(err)
		}
	}
	barrier()
	view.sink.drain()
	glowstone, err := BlockStateID("glowstone", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []uint32{glowstone, StateAir} {
		if err := w.SetBlock(-1, -60, 8, state); err != nil {
			t.Fatal(err)
		}
		barrier()
		if view.err != nil {
			t.Fatal(view.err)
		}
		packets := view.sink.drain()
		seenBlock, lights := false, 0
		for _, id := range packets {
			if id == proto.SBlockChange {
				seenBlock = true
			}
			if id == proto.SUpdateLight {
				if !seenBlock {
					t.Fatal("neighbor light packet arrived before source block modification", packets)
				}
				lights++
			}
		}
		if !seenBlock || lights != 2 {
			t.Fatal("missing source or cross-region light update", packets)
		}
	}
	// Equal original opacity/emission/face rules need the block update only.
	if err := w.SetBlock(-1, -61, 8, StateStone); err != nil {
		t.Fatal(err)
	}
	barrier()
	view.sink.drain()
	if err := w.SetBlock(-1, -61, 8, StateDirt); err != nil {
		t.Fatal(err)
	}
	barrier()
	if packets := view.sink.drain(); len(packets) != 1 || packets[0] != proto.SBlockChange {
		t.Fatal("equal light rules emitted redundant lighting packets", packets)
	}
}
