package nbt

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestVanillaLevelRoundTripPreservesUnknownTags(t *testing.T) {
	path := filepath.Join("..", "..", "vanilla", "world", "level.dat")
	f, err := os.Open(path)
	if err != nil {
		t.Skip(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if _, err := b.ReadFrom(gz); err != nil {
		t.Fatal(err)
	}
	gz.Close()
	name, root, err := Parse(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	data, ok := root.Object()
	if !ok {
		t.Fatal("missing root compound")
	}
	level, ok := data.Object("Data")
	if !ok {
		t.Fatal("missing Data compound")
	}
	version, ok := level.Number("DataVersion")
	if !ok || version != 3955 {
		t.Fatalf("DataVersion=%d", version)
	}
	reencoded, err := Marshal(name, root)
	if err != nil {
		t.Fatal(err)
	}
	_, again, err := Parse(reencoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(root, again) {
		t.Fatal("round-trip changed NBT tree")
	}
}

func TestRejectTruncatedNBT(t *testing.T) {
	for _, b := range [][]byte{{}, {Compound, 0, 0}, {Compound, 0, 0, Byte, 0, 1, 'x'}} {
		if _, _, err := Parse(b); err == nil {
			t.Fatalf("accepted truncated %v", b)
		}
	}
}
