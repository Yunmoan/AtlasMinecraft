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
	p.SetPos(.5, 64, .5, 0, 0, true)
	p.ReserveChunk(10, 0)
	p.ReserveChunk(0, 1)
	q.enqueue(genJob{cx: 10}, []PacketSink{p}, false, true)
	q.enqueue(genJob{cz: 1}, []PacketSink{p}, false, true)
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
	p.ReserveChunk(0, 0)
	w.RequestChunk(p, 0, 0)
	<-started
	for x := int32(1); x < 100; x++ {
		p.ReserveChunk(x, 0)
		w.RequestChunk(p, x, 0)
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
