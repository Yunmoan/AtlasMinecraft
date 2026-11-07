package world

import (
	"atlas/internal/player"
	"context"
	"sync"
	"testing"
	"time"
)

func TestGenerationQueuePrioritizesCurrentViewAndBlockingLoads(t *testing.T) {
	q := newGenerationQueue()
	p := player.New("test", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	p.SetPos(.5, 64, .5, 0, 0, true)
	sink.ReserveChunk(10, 0)
	sink.ReserveChunk(0, 1)
	q.enqueue(genJob{cx: 10}, []ChunkReceiver{sink}, false, true)
	q.enqueue(genJob{cz: 1}, []ChunkReceiver{sink}, false, true)
	done := make(chan struct{})
	j, ok := q.next(done)
	if !ok || j.cz != 1 || j.canceled {
		t.Fatal("distant request overtook nearby terrain")
	}
	q.enqueue(genJob{cx: 625045, cz: 624993}, nil, true, true)
	j, ok = q.next(done)
	if !ok || j.cx != 625045 || j.canceled {
		t.Fatal("teleport destination waited behind old view")
	}
	p.SetPos(10000722, 89, 9999902, 0, 0, false)
	p.PruneChunks(625045, 624993, 144)
	j, ok = q.next(done)
	if !ok || !j.canceled {
		t.Fatal("obsolete request still consumes generation")
	}
	close(done)
	if _, ok = q.next(done); ok {
		t.Fatal("queue survived shutdown")
	}
}

func TestTeleportCancelsOldQueueWithoutGeneratingIt(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var mu sync.Mutex
	var generated []chunkKey
	w := testGenerationWorld(1, func(cx, cz int32) (*Chunk, error) {
		if cx == 0 && cz == 0 {
			close(started)
			<-release
		}
		mu.Lock()
		generated = append(generated, chunkKey{cx, cz})
		mu.Unlock()
		return NewChunk(cx, cz), nil
	})
	defer w.Close()
	p := player.New("test", [16]byte{}, 1, nil)
	p.SetPos(.5, 64, .5, 0, 0, true)
	sink := p.WorldPacketSink()
	w.RequestChunk(sink, 0, 0)
	<-started
	for x := int32(1); x < 100; x++ {
		w.RequestChunk(sink, x, 0)
	}
	p.SetPos(10000722, 89, 9999902, 0, 0, false)
	p.PruneChunks(625045, 624993, 144)
	target := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() { target <- w.ensureChunk(ctx, 625045, 624993) }()
	close(release)
	if err := <-target; err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(generated) != 2 || generated[1] != (chunkKey{625045, 624993}) {
		t.Fatalf("obsolete chunks generated: %v", generated)
	}
}

func TestGenerationQueueTracksMovingWorldSink(t *testing.T) {
	q := newGenerationQueue()
	p := player.New("moving", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	for _, cx := range []int32{2, 8} {
		sink.ReserveChunk(cx, 0)
		q.enqueue(genJob{cx: cx}, []ChunkReceiver{sink}, false, true)
	}
	p.SetPos(128.5, 64, .5, 0, 0, true)
	job, ok := q.next(make(chan struct{}))
	if !ok || job.canceled || job.cx != 8 {
		t.Fatalf("queue retained old position's order: %+v", job)
	}
}

func TestGenerationQueueDisconnectKeepsSharedAndRequiredTasks(t *testing.T) {
	for _, scenario := range []struct {
		name             string
		shared, required bool
	}{
		{name: "disconnected-only"},
		{name: "another-viewer", shared: true},
		{name: "blocking-load", required: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			q := newGenerationQueue()
			p := player.New("gone", [16]byte{}, 1, nil)
			sink := p.WorldPacketSink()
			sink.ReserveChunk(1, 0)
			receivers := []ChunkReceiver{sink}
			if scenario.shared {
				other := player.New("here", [16]byte{1}, 2, nil).WorldPacketSink()
				other.ReserveChunk(1, 0)
				receivers = append(receivers, other)
			}
			q.enqueue(genJob{cx: 1}, receivers, scenario.required, true)
			p.Disconnect()
			job, ok := q.next(make(chan struct{}))
			wantCanceled := !scenario.shared && !scenario.required
			if !ok || job.canceled != wantCanceled {
				t.Fatalf("wrong disconnected/shared task selection: %+v", job)
			}
		})
	}
}
