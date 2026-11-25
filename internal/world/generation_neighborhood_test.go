package world

import (
	"atlas/internal/anvil"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalWorldLoggedChunkFailuresGenerateAndReload(t *testing.T) {
	dir := t.TempDir()
	seed := int64(-8981361456740854275)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	positions := []chunkKey{{1, -32}, {11, -63}, {11, -62}, {11, -61}}
	w, err := Open(noViewers{}, 4, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := w.prepareChunkSet(ctx, positions, nil); err != nil {
		t.Fatal(err)
	}
	for _, p := range positions {
		if packet, err := w.LoadChunk(p.x, p.z); err != nil || len(packet) == 0 {
			t.Fatalf("logged chunk %v unavailable: %v", p, err)
		}
	}
	// Generation can read these distant scratch neighbors but must not save
	// them as new full chunks or overwrite an existing player's neighbor.
	for _, p := range []chunkKey{{-2, -34}, {14, -62}} {
		if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), p.x, p.z); err != nil || exists {
			t.Fatalf("scratch neighbor %v persisted: exists=%t err=%v", p, exists, err)
		}
	}
	if err := w.SetBlock(17, 100, -511, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(noViewers{}, 4, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.prepareChunkSet(ctx, positions, nil); err != nil {
		t.Fatal(err)
	}
	for _, p := range positions {
		if packet, err := reopened.LoadChunk(p.x, p.z); err != nil || len(packet) == 0 {
			t.Fatalf("reopened logged chunk %v unavailable: %v", p, err)
		}
		name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), p.x, p.z)
		if err != nil || !exists {
			t.Fatalf("logged chunk %v not saved: %v", p, err)
		}
		chunk, err := chunkFromAnvil(name, root, p.x, p.z)
		if err != nil {
			t.Fatal(err)
		}
		if p.x == 1 && p.z == -32 && chunk.GetBlock(1, 1, 100) != StateStone {
			t.Fatal("saved edit changed on reload")
		}
	}
}
