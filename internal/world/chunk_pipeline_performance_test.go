package world

import (
	"atlas/internal/nbt"
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

// One operation creates a normal world, then generates, lights and saves a
// square of new chunks through the production worker pool. It measures the
// whole exploration pipeline rather than terrain generation alone.
func BenchmarkGenerateNormalArea(b *testing.B) {
	for _, workers := range []int{1, 8} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				seed := int64(1234)
				dir := filepath.Join(b.TempDir(), "world")
				if err := CreateNormalWorld(dir, &seed); err != nil {
					b.Fatal(err)
				}
				pinChunkBenchmarkSalt(b, dir)
				w, err := Open(noViewers{}, workers, dir)
				if err != nil {
					b.Fatal(err)
				}
				chunks := spawnPreparationChunks(40, -40, 4)
				b.StartTimer()
				if err := w.prepareChunkSet(context.Background(), chunks, nil); err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				if err := w.Close(); err != nil {
					b.Fatal(err)
				}
				b.ReportMetric(float64(len(chunks)), "chunks/op")
			}
		})
	}
}

// One operation reopens a normal save and loads an already generated square,
// including Anvil decoding, entity loading and lighting installation.
func BenchmarkLoadSavedNormalArea(b *testing.B) {
	seed := int64(1234)
	dir := filepath.Join(b.TempDir(), "world")
	if err := CreateNormalWorld(dir, &seed); err != nil {
		b.Fatal(err)
	}
	pinChunkBenchmarkSalt(b, dir)
	chunks := spawnPreparationChunks(40, -40, 4)
	w, err := Open(noViewers{}, 8, dir)
	if err != nil {
		b.Fatal(err)
	}
	if err := w.prepareChunkSet(context.Background(), chunks, nil); err != nil {
		b.Fatal(err)
	}
	if err := w.Close(); err != nil {
		b.Fatal(err)
	}
	for _, workers := range []int{1, 8} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				w, err := Open(noViewers{}, workers, dir)
				if err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				if err := w.prepareChunkSet(context.Background(), chunks, nil); err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				if err := w.Close(); err != nil {
					b.Fatal(err)
				}
				b.ReportMetric(float64(len(chunks)), "chunks/op")
			}
		})
	}
}

// World creation normally uses an independent random beehive salt. Pin it in
// the temporary save, outside the timer, so repetitions decorate identical input.
func pinChunkBenchmarkSalt(b *testing.B, dir string) {
	b.Helper()
	path := filepath.Join(dir, "level.dat")
	data, err := readGzip(path)
	if err != nil {
		b.Fatal(err)
	}
	name, root, err := nbt.Parse(data)
	if err != nil {
		b.Fatal(err)
	}
	object, _ := root.Object()
	level, _ := object.Object("Data")
	marker, _ := level.Object("AtlasWorldgen")
	marker["BeehiveShuffleSalt"] = nbt.Tag{Type: nbt.Long, Value: int64(0)}
	if err := writeGzipNBTAtomic(path, name, root); err != nil {
		b.Fatal(err)
	}
}
