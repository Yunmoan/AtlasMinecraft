package world

import (
	"atlas/internal/light"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestUnloadIdleReleasesLightingAndPacketsAndReloadsEdits(t *testing.T) {
	w, _ := mechanicsWorld(t)
	torch, err := BlockStateID("minecraft:torch", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(-1, -60, 8, torch); err != nil {
		t.Fatal(err)
	}
	before, err := w.LoadChunk(-1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := w.UnloadIdle(time.Minute); err != nil || count != 0 {
		t.Fatalf("recent direct accesses evicted: %d %v", count, err)
	}
	count, err := w.UnloadIdle(0)
	if err != nil || count != 9 {
		t.Fatalf("unload: %d %v", count, err)
	}
	if stats := w.Stats(); stats.LoadedChunks != 0 || stats.LightColumns != 0 {
		t.Fatalf("unload retained dense columns: %+v", stats)
	}
	// Actor maps are empty; no actor can access these cache entries now.
	for _, r := range w.regions {
		if len(r.packets.entries) != 0 || r.packets.bytes != 0 {
			t.Fatal("packet cache retained unloaded chunks")
		}
	}
	after, err := w.LoadChunk(-1, 0)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("saved blocks/light changed after unload and reload: %v", err)
	}
	state, err := w.BlockAt(-1, -60, 8)
	if err != nil || state != torch {
		t.Fatalf("edit lost after unload: %d %v", state, err)
	}
}

type unloadView struct{ visible atomic.Bool }

func (h *unloadView) ViewersOf(x, z int32) []PacketSink {
	if h.visible.Load() && x>>4 == -1 && z>>4 == 0 {
		return []PacketSink{unloadSink{}}
	}
	return nil
}

type unloadSink struct{}

func (unloadSink) SendPacket(int32, []byte) {}
func (unloadSink) SendChunkData([]byte)     {}

func TestUnloadRetainsViewerBorderAndActivePreparation(t *testing.T) {
	view := &unloadView{}
	view.visible.Store(true)
	dir := t.TempDir()
	seed := int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(view, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if err := w.PrepareSpawn(context.Background(), 0, 0, 1, nil); err != nil {
		t.Fatal(err)
	}
	if count, err := w.UnloadIdle(0); err != nil || count != 3 {
		t.Fatalf("view and lighting border not retained: %d %v", count, err)
	}
	if stats := w.Stats(); stats.LoadedChunks != 6 || stats.LightColumns != 6 {
		t.Fatalf("wrong visible cache: %+v", stats)
	}
	// A following load must coexist with maintenance and generate/load once.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ready := make(chan error, 1)
	go func() { ready <- w.ensureChunk(ctx, -2, 0) }()
	if _, err := w.UnloadIdle(time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := <-ready; err != nil {
		t.Fatal(err)
	}
	view.visible.Store(false)
	if count, err := w.UnloadIdle(0); err != nil || count != 7 {
		t.Fatalf("departed viewer retained chunks: %d %v", count, err)
	}
}

func TestUnloadSaveFailureRetainsChunkLightAndPacket(t *testing.T) {
	dir := t.TempDir()
	seed := int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	store, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	w := &World{store: store, hub: noViewers{}}
	w.lighting, err = light.New(MinY, WorldHeight, true)
	if err != nil {
		t.Fatal(err)
	}
	ch, err := store.vanilla.loadChunk(-1, 0)
	if err != nil {
		t.Fatal(err)
	}
	position := light.Position{X: -1, Z: 0}
	if _, err := w.lighting.Install(position, denseLightStates(ch), nil); err != nil {
		t.Fatal(err)
	}
	ch.lights = w.lighting.Snapshot(position)
	ch.dirty = true
	w.loadedChunks.Add(1)
	r := newRegion(w, -1, 0)
	r.chunks[chunkKey{-1, 0}] = ch
	r.packets.packet(ch)
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, []byte("blocked"), 0644); err != nil {
		t.Fatal(err)
	}
	store.vanilla.regionDir = filepath.Join(blocked, "region")
	if count, err := r.unloadUnused(time.Now()); err == nil || count != 0 {
		t.Fatalf("failed save discarded chunk: %d %v", count, err)
	}
	if r.chunks[chunkKey{-1, 0}] != ch || w.lighting.Snapshot(position) == nil || len(r.packets.entries) != 1 || w.loadedChunks.Load() != 1 {
		t.Fatal("failed save released live data")
	}
	store.vanilla.regionDir = filepath.Join(dir, "region")
	if count, err := r.unloadUnused(time.Now()); err != nil || count != 1 {
		t.Fatalf("retry did not save and release: %d %v", count, err)
	}
}
