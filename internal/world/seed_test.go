package world

import (
	"atlas/internal/anvil"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFlatWorldGeneratesJavaSave(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "flat-world")
	seed := int64(7122209284156038298)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	v, err := openVanilla(dir)
	if err != nil || v.seed != seed || v.flat == nil || !v.flat.exact {
		t.Fatalf("new level.dat: save=%+v err=%v", v, err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	if w.Seed() != seed || !w.IsFlat() {
		t.Fatal("created world has wrong seed or generator")
	}
	if err := w.PreGenerateSpawn(0, 0); err != nil {
		t.Fatal(err)
	}
	if y, err := w.SurfaceY(0, 0); err != nil || y != -60 {
		t.Fatalf("spawn surface = %d, %v", y, err)
	}
	if _, err := w.LoadChunk(-33, 17); err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(-33*16, -60, 17*16, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), -33, 17)
	if err != nil || !exists {
		t.Fatalf("new Anvil chunk missing: %v", err)
	}
	chunk, err := chunkFromAnvil(name, root, -33, 17)
	if err != nil || chunk.GetBlock(0, 0, -60) != StateStone {
		t.Fatalf("reloaded generated chunk: %v", err)
	}
	reopened, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if reopened.Seed() != seed {
		t.Fatal("seed changed on reopen")
	}
	if _, err := os.Stat(filepath.Join(dir, "level.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy metadata created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "blocks.log")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy journal created: %v", err)
	}
}

func TestCreateFlatWorldRefusesExistingFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "level.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CreateFlatWorld(dir, nil); err == nil {
		t.Fatal("legacy save was overwritten")
	}
	if _, err := Open(noViewers{}, 1, dir); err == nil {
		t.Fatal("legacy save was loaded")
	}
}

func TestNormalWorldMetadataUsesVanillaNoisePreset(t *testing.T) {
	dir := t.TempDir()
	seed := int64(-42)
	root := newWorldLevel("normal-world", seed, false)
	if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), "", root); err != nil {
		t.Fatal(err)
	}
	save, err := openVanilla(dir)
	if err != nil || save.seed != seed || save.flat != nil {
		t.Fatalf("normal metadata: save=%+v err=%v", save, err)
	}
	levelRoot, _ := root.Object()
	data, _ := levelRoot.Object("Data")
	settings, _ := data.Object("WorldGenSettings")
	dimensions, _ := settings.Object("dimensions")
	overworld, _ := dimensions.Object("minecraft:overworld")
	generator, _ := overworld.Object("generator")
	biomeSource, _ := generator.Object("biome_source")
	if kind, _ := generator.Text("type"); kind != "minecraft:noise" {
		t.Fatalf("wrong generator: %s", kind)
	}
	if preset, _ := biomeSource.Text("preset"); preset != "minecraft:overworld" {
		t.Fatalf("wrong biome preset: %s", preset)
	}
	if save.normal == nil {
		t.Fatal("normal preset has no generator")
	}
	if structures, _ := settings.Number("generate_features"); structures != 1 {
		t.Fatal("new normal world's supported structures were disabled")
	}
	if marker, ok := data.Object("AtlasWorldgen"); !ok {
		t.Fatal("missing generation coverage marker")
	} else if complete, _ := marker.Number("Complete"); complete != 0 {
		t.Fatal("partial generator advertised as complete")
	}
}
