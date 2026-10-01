package player

import "testing"

func TestChunkReservationDeliveryAndCancellation(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	if !p.ReserveChunk(1, 0) || p.ReserveChunk(1, 0) || p.IsChunkSent(1, 0) {
		t.Fatal("reservation marked delivery or duplicated")
	}
	if !p.SendRequestedChunk(1, 0, []byte{1}) || !p.IsChunkSent(1, 0) {
		t.Fatal("delivery not marked")
	}
	if p.SendRequestedChunk(1, 0, []byte{1}) {
		t.Fatal("duplicate delivery")
	}
	checkChunkBatches(t, p, 1)
	p.ReserveChunk(-1, 0)
	p.PruneChunks(10, 0, 4)
	if p.SendRequestedChunk(-1, 0, []byte{2}) || p.IsChunkSent(-1, 0) {
		t.Fatal("canceled generation reinserted unloaded chunk")
	}
}

func TestFarTeleportPrunesDeliveredAndPendingChunksWithoutDistanceOverflow(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	p.ReserveChunk(0, 0)
	p.SendRequestedChunk(0, 0, []byte{1})
	p.ReserveChunk(1, 0)
	p.ReserveChunk(625045, 624993)
	pruned := p.PruneChunks(625045, 624993, 144)
	if len(pruned) != 1 || pruned[0] != [2]int32{0, 0} || p.IsChunkSent(0, 0) || p.WantsChunk(1, 0) {
		t.Fatal("far teleport retained chunks at the old position")
	}
	if !p.WantsChunk(625045, 624993) || p.SendRequestedChunk(1, 0, []byte{1}) {
		t.Fatal("teleport lost target reservation or accepted canceled delivery")
	}
}

func TestPendingCancellationRetainsDeliveredGround(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	p.ReserveChunk(0, 0)
	p.SendRequestedChunk(0, 0, []byte{1})
	p.ReserveChunk(1, 0)
	p.CancelPendingChunks()
	if !p.IsChunkSent(0, 0) || p.WantsChunk(1, 0) || p.SendRequestedChunk(1, 0, []byte{1}) {
		t.Fatal("preparing a teleport removed delivered ground or accepted old generation")
	}
	if !p.ReserveChunk(1, 0) {
		t.Fatal("canceled request could not be restored after a failed teleport")
	}
}
