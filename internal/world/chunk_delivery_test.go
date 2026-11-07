package world

import (
	"atlas/internal/player"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// The production adapter must satisfy the whole request contract at compile time.
var _ ChunkReceiver = (*player.WorldSink)(nil)

func awaitChunkCondition(t *testing.T, condition func() bool) {
	t.Helper()
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for !condition() {
		select {
		case <-timer.C:
			t.Fatal("chunk lifecycle did not complete")
		case <-ticker.C:
		}
	}
}

func TestChunkRequestsReserveOnceAndDeliverThroughWorldSink(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	var calls atomic.Int32
	w := testGenerationWorld(1, func(cx, cz int32) (*Chunk, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		return NewChunk(cx, cz), nil
	})
	t.Cleanup(func() { unblock.Do(func() { close(release) }); w.Close() })
	p := player.New("viewer", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	if err := w.RequestChunk(sink, 0, 0); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("generation never started")
	}
	if err := w.RequestChunk(p.WorldPacketSink(), 0, 0); err != nil {
		t.Fatal(err)
	}
	unblock.Do(func() { close(release) })
	select {
	case <-p.ChunkUpdates:
	case <-time.After(3 * time.Second):
		t.Fatal("no delivery notification")
	}
	if !sink.IsChunkSent(0, 0) || sink.WantsChunk(0, 0) || calls.Load() != 1 || len(p.Out) != 3 {
		t.Fatal("reservation, generation or delivery was duplicated")
	}
	if err := w.RequestChunk(sink, 0, 0); err != nil || len(p.Out) != 3 {
		t.Fatal("delivered terrain was requested again", err)
	}
}

func TestChunkRequestFailuresReleaseReservationForRetry(t *testing.T) {
	for _, synchronous := range []bool{false, true} {
		name := "async"
		if synchronous {
			name = "spawn"
		}
		t.Run(name, func(t *testing.T) {
			failure := errors.New("read failed")
			var calls atomic.Int32
			w := testGenerationWorld(1, func(cx, cz int32) (*Chunk, error) {
				if calls.Add(1) == 1 {
					return nil, failure
				}
				return NewChunk(cx, cz), nil
			})
			t.Cleanup(func() { w.Close() })
			p := player.New("retry", [16]byte{}, 1, nil)
			sink := p.WorldPacketSink()
			if synchronous {
				if err := w.SpawnChunk(sink, 0, 0); !errors.Is(err, failure) {
					t.Fatal("spawn did not return loading failure", err)
				}
			} else if err := w.RequestChunk(sink, 0, 0); err != nil {
				t.Fatal(err)
			}
			awaitChunkCondition(t, func() bool { return !sink.WantsChunk(0, 0) })
			if sink.IsChunkSent(0, 0) || len(p.Out) != 0 {
				t.Fatal("failed load published terrain")
			}
			// Both entrances must release their own failed reservation.
			if err := w.SpawnChunk(sink, 0, 0); err != nil || !sink.IsChunkSent(0, 0) || calls.Load() != 2 {
				t.Fatal("failed reservation blocked retry", err, calls.Load())
			}
		})
	}
}

func TestClosedWorldRejectsRequestsWithoutLeakingReservations(t *testing.T) {
	w := testGenerationWorld(1, func(cx, cz int32) (*Chunk, error) { return NewChunk(cx, cz), nil })
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	p := player.New("rejected", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	if err := w.RequestChunk(sink, 1, 0); !errors.Is(err, ErrWorldClosed) {
		t.Fatal("request did not report closed world", err)
	}
	if err := w.SpawnChunk(sink, 2, 0); !errors.Is(err, ErrWorldClosed) {
		t.Fatal("spawn did not report closed world", err)
	}
	for _, cx := range []int32{1, 2} {
		if !sink.ReserveChunk(cx, 0) {
			t.Fatal("rejected request leaked reservation", cx)
		}
	}
}

func TestWorldCloseReleasesAcceptedChunkRequests(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var begin, unblock sync.Once
	w := testGenerationWorld(1, func(cx, cz int32) (*Chunk, error) {
		begin.Do(func() { close(started) })
		<-release
		return NewChunk(cx, cz), nil
	})
	t.Cleanup(func() { unblock.Do(func() { close(release) }); w.Close() })
	p := player.New("closing", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	if err := w.RequestChunk(sink, 0, 0); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("generation never started")
	}
	for cx := int32(1); cx < 96; cx++ {
		if err := w.RequestChunk(sink, cx, 0); err != nil {
			t.Fatal(err)
		}
	}
	closed := make(chan error, 1)
	go func() { closed <- w.Close() }()
	<-w.done
	unblock.Do(func() { close(release) })
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("world close did not complete")
	}
	for cx := int32(0); cx < 96; cx++ {
		if !sink.ReserveChunk(cx, 0) {
			t.Fatal("close retained accepted reservation", cx)
		}
	}
}

func TestConcurrentCloseAndChunkPostsReleaseReservations(t *testing.T) {
	w := testGenerationWorld(2, func(int32, int32) (*Chunk, error) { return nil, errors.New("read failed") })
	t.Cleanup(func() { w.Close() })
	p := player.New("concurrent", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	var posters sync.WaitGroup
	start := make(chan struct{})
	for i := int32(0); i < 64; i++ {
		posters.Add(1)
		go func(cx int32) {
			defer posters.Done()
			<-start
			if err := w.RequestChunk(sink, cx, 0); err != nil && !errors.Is(err, ErrWorldClosed) {
				t.Error(err)
			}
		}(i)
	}
	close(start)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	posters.Wait()
	for cx := int32(0); cx < 64; cx++ {
		if !sink.ReserveChunk(cx, 0) {
			t.Fatal("post racing close leaked reservation", cx)
		}
	}
}

func TestChunkRequestsDisconnectPreservesSharedAndBlockingWork(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	generated := make(chan int32, 8)
	w := testGenerationWorld(1, func(cx, cz int32) (*Chunk, error) {
		generated <- cx
		if cx == 0 {
			close(started)
			<-release
		}
		return NewChunk(cx, cz), nil
	})
	t.Cleanup(func() { unblock.Do(func() { close(release) }); w.Close() })
	gone := player.New("gone", [16]byte{}, 1, nil)
	here := player.New("here", [16]byte{1}, 2, nil)
	old, current := gone.WorldPacketSink(), here.WorldPacketSink()
	if err := w.RequestChunk(old, 0, 0); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight generation never started")
	}
	for _, request := range []struct {
		to ChunkReceiver
		cx int32
	}{{old, 2}, {old, 1}, {current, 1}} {
		if err := w.RequestChunk(request.to, request.cx, 0); err != nil {
			t.Fatal(err)
		}
	}
	gone.Disconnect()
	blocking := make(chan error, 1)
	go func() { _, err := w.LoadChunk(3, 0); blocking <- err }()
	unblock.Do(func() { close(release) })
	select {
	case err := <-blocking:
		if err != nil {
			t.Fatal("blocking load was canceled with viewer", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("blocking load did not complete")
	}
	select {
	case <-here.ChunkUpdates:
	case <-time.After(3 * time.Second):
		t.Fatal("other viewer lost shared generation")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if !current.IsChunkSent(1, 0) || len(gone.Out) != 0 {
		t.Fatal("shared result lost or in-flight result reached disconnected viewer")
	}
	close(generated)
	seen := map[int32]bool{}
	for cx := range generated {
		seen[cx] = true
	}
	if len(seen) != 3 || !seen[0] || !seen[1] || !seen[3] {
		t.Fatal("wrong work survived disconnect", seen)
	}
}
