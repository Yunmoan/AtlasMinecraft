package world

import "context"

// ChunkReceiver owns the lifecycle of one viewer's chunk reservations. Unlike
// PacketSink (also used for broadcasts), every request receiver must support
// deduplication, current-view priority, guarded delivery and failure release.
// Implementations must be safe for concurrent region and generation workers.
type ChunkReceiver interface {
	ReserveChunk(cx, cz int32) bool
	// ChunkPriority returns squared distance from the current position and
	// whether this reservation still belongs to a live, current view.
	ChunkPriority(cx, cz int32) (distance int64, wanted bool)
	// SendRequestedChunk queues a complete batch only if still reserved.
	// Success means queued on the server, not acknowledged by the client.
	SendRequestedChunk(cx, cz int32, payload []byte) bool
	// ReleaseChunk releases only an outstanding reservation in this view;
	// it must not forget delivered terrain or affect a newer world view.
	ReleaseChunk(cx, cz int32)
}

// RequestChunk reserves and asynchronously delivers a chunk. Duplicate or
// obsolete requests are no-ops. Rejected posts release their reservation;
// accepted requests are completed or released by the owning region.
func (w *World) RequestChunk(to ChunkReceiver, cx, cz int32) error {
	if !to.ReserveChunk(cx, cz) {
		return nil
	}
	if err := w.post(context.Background(), cx, cz, chunkRequest{cx: cx, cz: cz, to: to}); err != nil {
		to.ReleaseChunk(cx, cz)
		return err
	}
	return nil
}

// SpawnChunk reserves and synchronously loads the player's current chunk so it
// can precede surrounding terrain. Delivery and failure use the same receiver
// contract as asynchronous requests; the receiver records successful delivery.
func (w *World) SpawnChunk(to ChunkReceiver, cx, cz int32) error {
	if !to.ReserveChunk(cx, cz) {
		return nil
	}
	payload, err := w.LoadChunk(cx, cz)
	if err != nil {
		to.ReleaseChunk(cx, cz)
		return err
	}
	to.SendRequestedChunk(cx, cz, payload)
	return nil
}

// releaseChunkRequests runs after all actors and admitted posts have stopped.
// Requests may still be pending generation or sitting unread in a mailbox.
func (r *Region) releaseChunkRequests() {
	for key, receivers := range r.pending {
		for _, to := range receivers {
			to.ReleaseChunk(key.x, key.z)
		}
	}
	clear(r.pending)
	for {
		select {
		case message := <-r.mailbox:
			if request, ok := message.(chunkRequest); ok {
				request.to.ReleaseChunk(request.cx, request.cz)
			}
		default:
			return
		}
	}
}
