package anvil

import (
	"atlas/internal/nbt"
	"os"
	"path/filepath"
	"testing"
)

// A populated Java region makes whole-file reads visible. The source is read
// only; every write benchmark uses its own temporary copy.
func BenchmarkReadPopulatedRegion(b *testing.B) {
	dir := filepath.Join("..", "..", "vanilla", "world", "region")
	if reference := os.Getenv("ATLAS_BENCH_REGION"); reference != "" {
		dir = reference
	}
	if _, err := os.Stat(regionPath(dir, 0, 0)); err != nil {
		b.Skip("vanilla sample region unavailable")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, _, exists, err := ReadChunk(dir, 0, 0); err != nil || !exists {
			b.Fatalf("chunk read: exists=%v err=%v", exists, err)
		}
	}
}

func BenchmarkWriteChunk(b *testing.B) {
	dir := b.TempDir()
	root := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"DataVersion": {Type: nbt.Int, Value: int32(3955)},
		"payload":     {Type: nbt.ByteArray, Value: make([]byte, 64<<10)},
	}}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := WriteChunk(dir, 0, 0, "", root); err != nil {
			b.Fatal(err)
		}
	}
}
