package world

import (
	"atlas/internal/anvil"
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPrepareSpawnPersistsSquareAndRetainsEdits(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "flat")
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 4, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	var progress []SpawnPreparationProgress
	if err := w.PrepareSpawn(context.Background(), -1, -1, 1, func(p SpawnPreparationProgress) { progress = append(progress, p) }); err != nil {
		t.Fatal(err)
	}
	if len(progress) != 10 || progress[0].Completed != 0 || progress[9].Completed != 9 {
		t.Fatalf("progress: %v", progress)
	}
	for i, p := range progress {
		if p.Total != 9 || p.Completed != i {
			t.Fatalf("progress not sequential: %v", progress)
		}
	}
	for x := int32(-2); x <= 0; x++ {
		for z := int32(-2); z <= 0; z++ {
			if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), x, z); err != nil || !exists {
				t.Fatalf("chunk %d,%d not saved before ready: %v", x, z, err)
			}
		}
	}
	if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), 1, 0); err != nil || exists {
		t.Fatalf("generated outside requested area: %v", err)
	}
	if err := w.SetBlock(-1, 1, -1, StateStone); err != nil {
		t.Fatal(err)
	}
	before, err := w.LoadChunk(-1, -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.PrepareSpawn(context.Background(), -1, -1, 1, nil); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.LoadChunk(0, 0); !errors.Is(err, ErrWorldClosed) {
		t.Fatalf("request after Close: %v", err)
	}
	reopened, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.PrepareSpawn(context.Background(), -1, -1, 1, nil); err != nil {
		t.Fatal(err)
	}
	after, err := reopened.LoadChunk(-1, -1)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("preload/restart replaced edits: %v", err)
	}
}

func TestConcurrentPreloadRequestsGenerateChunkOnce(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	w := testGenerationWorld(4, func(cx, cz int32) (*Chunk, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		return NewChunk(cx, cz), nil
	})
	defer w.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := w.ensureChunk(ctx, -1, -1); err != nil {
				t.Error(err)
			}
		}()
	}
	select {
	case <-started:
	case <-ctx.Done():
		close(release)
		t.Fatal("no generation started")
	}
	close(release)
	wg.Wait()
	if got := calls.Load(); got != 1 {
		t.Fatalf("concurrent requests generated the same chunk %d times", got)
	}
}

func testGenerationWorld(workers int, load func(int32, int32) (*Chunk, error)) *World {
	w := &World{regions: make(map[chunkKey]*Region), generation: newGenerationQueue(), hub: noViewers{}, done: make(chan struct{}), workers: workers, loadChunk: load}
	for i := 0; i < workers; i++ {
		w.wg.Add(1)
		go w.genWorker()
	}
	return w
}

func TestPrepareSpawnUsesParallelWorkersAndCoalesces(t *testing.T) {
	started := make(chan chunkKey, 4)
	release := make(chan struct{})
	var calls atomic.Int32
	w := testGenerationWorld(4, func(cx, cz int32) (*Chunk, error) {
		calls.Add(1)
		if cx != 0 || cz != 0 {
			select {
			case <-release:
			default:
				started <- chunkKey{cx, cz}
				<-release
			}
		}
		return NewChunk(cx, cz), nil
	})
	defer w.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- w.PrepareSpawn(ctx, 0, 0, 1, nil) }()
	for i := 0; i < 4; i++ {
		select {
		case <-started:
		case <-ctx.Done():
			close(release)
			t.Fatal("four generation workers did not overlap")
		}
	}
	close(release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != 9 {
		t.Fatalf("duplicate generation: %d calls, want 9", got)
	}
	// Repeated requests for the already loaded center never regenerate it.
	for i := 0; i < 20; i++ {
		if err := w.ensureChunk(ctx, 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	if got := calls.Load(); got != 9 {
		t.Fatalf("preload bypassed actor cache: %d calls", got)
	}
}

func TestPrepareDestinationStartsNeighborsWhileCenterIsBlocked(t *testing.T) {
	started, release := make(chan chunkKey, 4), make(chan struct{})
	var calls atomic.Int32
	w := testGenerationWorld(4, func(cx, cz int32) (*Chunk, error) {
		if calls.Add(1) <= 4 {
			started <- chunkKey{cx, cz}
			<-release
		}
		return NewChunk(cx, cz), nil
	})
	defer w.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- w.PrepareDestination(ctx, 625045, 624993) }()
	seen := map[chunkKey]bool{}
	for len(seen) < 4 {
		select {
		case p := <-started:
			seen[p] = true
		case <-ctx.Done():
			close(release)
			t.Fatal("destination preparation waited for the center instead of starting neighbors")
		}
	}
	close(release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 9 {
		t.Fatalf("destination generated %d chunks, want its immediate 3x3", calls.Load())
	}
	if err := w.PrepareDestination(ctx, 1875000, 0); err == nil {
		t.Fatal("accepted a destination outside generation bounds")
	}
	if err := w.PrepareDestination(ctx, 1874998, 0); err != nil {
		t.Fatalf("rejected an existing center near the world border: %v", err)
	}
}

func TestPrepareSpawnCancellationAndCloseWaitForActiveGeneration(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	w := testGenerationWorld(4, func(cx, cz int32) (*Chunk, error) {
		close(started)
		<-release
		return NewChunk(cx, cz), nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan error, 1)
	go func() { finished <- w.PrepareSpawn(ctx, 0, 0, 3, nil) }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("no generation started")
	}
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation: %v", err)
		}
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("preload did not cancel")
	}
	closed := make(chan error, 1)
	go func() { closed <- w.Close() }()
	select {
	case <-closed:
		close(release)
		t.Fatal("Close returned while a generator could still write")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not join workers")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareSpawnErrorAndValidation(t *testing.T) {
	failure := errors.New("disk unavailable")
	w := testGenerationWorld(2, func(cx, cz int32) (*Chunk, error) {
		if cx == 0 && cz == 0 {
			return NewChunk(cx, cz), nil
		}
		return nil, failure
	})
	defer w.Close()
	if err := w.PrepareSpawn(context.Background(), 0, 0, 1, nil); !errors.Is(err, failure) {
		t.Fatalf("lost generation error: %v", err)
	}
	for _, radius := range []int{-1, 17} {
		if err := w.PrepareSpawn(context.Background(), 0, 0, radius, nil); err == nil {
			t.Fatalf("accepted radius %d", radius)
		}
	}
	if err := w.PrepareSpawn(context.Background(), 1874997, 0, 1, nil); err == nil {
		t.Fatal("accepted out-of-bounds halo")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.PrepareSpawn(ctx, 0, 0, 0, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("already canceled: %v", err)
	}
}
