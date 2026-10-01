package world

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWorldStatsDuringBlockedGenerationAndAfterReuse(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	w := testGenerationWorld(1, func(cx, cz int32) (*Chunk, error) {
		if cx == 0 && cz == 0 {
			close(started)
			<-release
		}
		return NewChunk(cx, cz), nil
	})
	defer w.Close()
	defer unblock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first, second := make(chan error, 1), make(chan error, 1)
	go func() { first <- w.ensureChunk(ctx, 0, 0) }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("generation did not start")
	}
	go func() { second <- w.ensureChunk(ctx, 32, 0) }()
	for {
		s := w.Stats()
		if s.QueuedGeneration == 1 {
			if s.ActiveGeneration != 1 || s.GenerationWorkers != 1 || s.LoadedChunks != 0 || s.Regions != 2 {
				t.Fatalf("wrong busy-world counters: %+v", s)
			}
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("queued generation not visible")
		case <-time.After(time.Millisecond):
		}
	}
	unblock()
	for _, ready := range []chan error{first, second} {
		if err := <-ready; err != nil {
			t.Fatal(err)
		}
	}
	if err := w.ensureChunk(ctx, 0, 0); err != nil {
		t.Fatal(err)
	}
	if s := w.Stats(); s.LoadedChunks != 2 || s.ActiveGeneration != 0 || s.QueuedGeneration != 0 {
		t.Fatalf("cache reuse or completion changed counters: %+v", s)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if s := w.Stats(); s.ActiveGeneration != 0 || s.LoadedChunks != 2 {
		t.Fatalf("closed-world counters: %+v", s)
	}
}
