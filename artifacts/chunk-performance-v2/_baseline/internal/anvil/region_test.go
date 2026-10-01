package anvil

import (
	"atlas/internal/nbt"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAndRewriteVanillaRegion(t *testing.T) {
	source := filepath.Join("..", "..", "vanilla", "world", "region")
	if _, err := os.Stat(filepath.Join(source, "r.0.0.mca")); os.IsNotExist(err) {
		t.Skip("vanilla sample region unavailable")
	} else if err != nil {
		t.Fatal(err)
	}
	name, root, exists, err := ReadChunk(source, 0, 0)
	if os.IsNotExist(err) {
		t.Skip(err)
	}
	if err != nil || !exists {
		t.Fatalf("read vanilla chunk: exists=%v err=%v", exists, err)
	}
	o, ok := root.Object()
	if !ok {
		t.Fatal("chunk is not compound")
	}
	if version, ok := o.Number("DataVersion"); !ok || version != 3955 {
		t.Fatalf("DataVersion=%d", version)
	}
	if _, ok := o["sections"].Elements(); !ok {
		t.Fatal("sections missing")
	}
	dir := t.TempDir()
	regionName := "r.0.0.mca"
	b, err := os.ReadFile(filepath.Join(source, regionName))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, regionName), b, 0644); err != nil {
		t.Fatal(err)
	}
	o["AtlasTest"] = nbt.Tag{Type: nbt.String, Value: "roundtrip"}
	if err := WriteChunk(dir, 0, 0, name, root); err != nil {
		t.Fatal(err)
	}
	_, edited, exists, err := ReadChunk(dir, 0, 0)
	if err != nil || !exists {
		t.Fatalf("read edited chunk: %v", err)
	}
	editedObject, _ := edited.Object()
	if got, _ := editedObject.Text("AtlasTest"); got != "roundtrip" {
		t.Fatalf("edit=%q", got)
	}
	_, neighbor, exists, err := ReadChunk(dir, 1, 0)
	if err != nil || !exists {
		t.Fatalf("neighbor lost: %v", err)
	}
	neighborObject, _ := neighbor.Object()
	if x, _ := neighborObject.Number("xPos"); x != 1 {
		t.Fatalf("neighbor xPos=%d", x)
	}
}
