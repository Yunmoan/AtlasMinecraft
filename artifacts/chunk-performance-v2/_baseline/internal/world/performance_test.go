package world

import (
	"context"
	"path/filepath"
	"testing"
)

func BenchmarkLoadedChunkPacket(b *testing.B) {
	seed := int64(1234)
	dir := filepath.Join(b.TempDir(), "world")
	if err := CreateFlatWorld(dir, &seed); err != nil {
		b.Fatal(err)
	}
	w, err := Open(noViewers{}, 4, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()
	if _, err := w.LoadChunk(-1, -1); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := w.LoadChunk(-1, -1); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func BenchmarkPrepareSavedSpawn(b *testing.B) {
	seed := int64(1234)
	dir := filepath.Join(b.TempDir(), "world")
	if err := CreateFlatWorld(dir, &seed); err != nil {
		b.Fatal(err)
	}
	w, err := Open(noViewers{}, 4, dir)
	if err != nil {
		b.Fatal(err)
	}
	if err := w.PrepareSpawn(context.Background(), 0, 0, 3, nil); err != nil {
		b.Fatal(err)
	}
	if err := w.Close(); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		w, err := Open(noViewers{}, 4, dir)
		if err != nil {
			b.Fatal(err)
		}
		err = w.PrepareSpawn(context.Background(), 0, 0, 3, nil)
		closeErr := w.Close()
		if err != nil || closeErr != nil {
			b.Fatalf("prepare=%v close=%v", err, closeErr)
		}
	}
}
