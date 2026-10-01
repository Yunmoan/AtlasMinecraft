package anvil

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// This fixture was written by lz4-java 1.8.0 from the supplied server.jar.
// The compressor used 256-byte blocks, exercising multiple compressed blocks.
func TestJavaLZ4BlockStream(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "lz4-java-1.8.0.bin"))
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Repeat([]byte("vanilla-lz4-block-"), 128)
	got, err := decompress(4, fixture)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("decoded %d bytes, want %d", len(got), len(want))
	}
	bad := bytes.Clone(fixture)
	bad[17] ^= 1
	if _, err := decompress(4, bad); err == nil {
		t.Fatal("corrupt LZ4 checksum was accepted")
	}
}
