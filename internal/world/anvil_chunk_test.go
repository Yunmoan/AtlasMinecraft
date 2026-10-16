package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/registry"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewChunkAnvilBiomePaletteRoundTrip(t *testing.T) {
	chunk := NewChunk(-2, 3)
	forest, err := registry.BiomeID("minecraft:forest")
	if err != nil {
		t.Fatal(err)
	}
	desert, err := registry.BiomeID("minecraft:desert")
	if err != nil {
		t.Fatal(err)
	}
	for section := range chunk.biomes {
		for index := range chunk.biomes[section] {
			if index%3 == 0 {
				chunk.biomes[section][index] = forest
			} else if index%3 == 1 {
				chunk.biomes[section][index] = desert
			}
		}
	}
	if err := chunk.initializeAnvilRoot(); err != nil {
		t.Fatal(err)
	}
	name, root, err := chunk.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := anvil.WriteChunk(dir, -2, 3, name, root); err != nil {
		t.Fatal(err)
	}
	reloadedName, reloadedRoot, exists, err := anvil.ReadChunk(dir, -2, 3)
	if err != nil || !exists {
		t.Fatalf("read generated chunk: %v", err)
	}
	reloaded, err := chunkFromAnvil(reloadedName, reloadedRoot, -2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reloaded.biomes, chunk.biomes) {
		t.Fatal("quart biomes changed during Anvil round trip")
	}
}

func TestVanillaNoiseSaveChunkAvailability(t *testing.T) {
	dir := copyVanillaWorldForGenerationTest(t)
	if _, err := os.Stat(filepath.Join(dir, "level.dat")); os.IsNotExist(err) {
		t.Skip("vanilla reference save unavailable")
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.flat != nil {
		t.Skip("reference save is not a noise world")
	}
	if _, err := v.loadChunk(0, 1); err != nil {
		t.Fatalf("existing spawn chunk: %v", err)
	}
	chunk, err := v.loadChunk(32, 32)
	if errors.Is(err, ErrVanillaGenerationUnavailable) {
		t.Log("missing noise chunk 32,32 cannot yet be generated")
		return
	}
	if err != nil || chunk == nil {
		t.Fatalf("missing noise chunk 32,32: chunk=%v error=%v", chunk, err)
	}
}

func TestVanillaNoiseSaveRejectsUnsupportedPreset(t *testing.T) {
	dir := t.TempDir()
	root := newWorldLevel("large-biomes", 1234, false)
	object, _ := root.Object()
	data, _ := object.Object("Data")
	settings, _ := data.Object("WorldGenSettings")
	dimensions, _ := settings.Object("dimensions")
	overworld, _ := dimensions.Object("minecraft:overworld")
	generator, _ := overworld.Object("generator")
	generator["settings"] = nbt.Tag{Type: nbt.String, Value: "minecraft:large_biomes"}
	if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), "", root); err != nil {
		t.Fatal(err)
	}
	save, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	if save.flat != nil {
		t.Skip("reference save is not a noise world")
	}
	if _, err := save.loadChunk(100000, 100000); !errors.Is(err, ErrVanillaGenerationUnavailable) {
		t.Fatalf("unsupported noise preset must stay disabled, got %v", err)
	}
}

func copyVanillaWorldForGenerationTest(t *testing.T) string {
	t.Helper()
	source := filepath.Join("..", "..", "vanilla", "world")
	dir := t.TempDir()
	level, err := os.ReadFile(filepath.Join(source, "level.dat"))
	if os.IsNotExist(err) {
		t.Skip("vanilla reference save unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "level.dat"), level, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "region"), 0755); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(source, "region", "*.mca"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "region", filepath.Base(path)), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestAllSampleOverworldChunksParse(t *testing.T) {
	source := filepath.Join("..", "..", "vanilla", "world", "region")
	files, err := filepath.Glob(filepath.Join(source, "r.*.*.mca"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("sample world unavailable")
	}
	count := 0
	for _, path := range files {
		var rx, rz int32
		if _, err := fmt.Sscanf(filepath.Base(path), "r.%d.%d.mca", &rx, &rz); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 1024; i++ {
			if binary.BigEndian.Uint32(b[i*4:i*4+4]) == 0 {
				continue
			}
			cx, cz := rx*32+int32(i%32), rz*32+int32(i/32)
			name, root, exists, err := anvil.ReadChunk(source, cx, cz)
			if err != nil || !exists {
				t.Fatalf("read chunk %d,%d: %v", cx, cz, err)
			}
			if _, err := chunkFromAnvil(name, root, cx, cz); err != nil {
				t.Fatalf("decode chunk %d,%d: %v", cx, cz, err)
			}
			count++
		}
	}
	t.Logf("parsed %d vanilla overworld chunks", count)
}

func TestVanillaChunkPaletteAndEditRoundTrip(t *testing.T) {
	requireLegacyFlatSample(t)
	source := filepath.Join("..", "..", "vanilla", "world", "region")
	if _, err := os.Stat(filepath.Join(source, "r.0.0.mca")); os.IsNotExist(err) {
		t.Skip("vanilla sample region unavailable")
	} else if err != nil {
		t.Fatal(err)
	}
	name, root, exists, err := anvil.ReadChunk(source, 0, 0)
	if os.IsNotExist(err) {
		t.Skip(err)
	}
	if err != nil || !exists {
		t.Fatalf("read source: %v", err)
	}
	chunk, err := chunkFromAnvil(name, root, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := chunk.GetBlock(0, 0, -64); got != StateBedrock {
		t.Fatalf("bedrock=%d", got)
	}
	if got := chunk.GetBlock(0, 0, -61); got != StateGrass {
		t.Fatalf("grass=%d", got)
	}
	if got := chunk.GetBlock(0, 0, -60); got != StateAir {
		t.Fatalf("air=%d", got)
	}
	if got := chunk.Height(0, 0); got != -60 {
		t.Fatalf("surface=%d", got)
	}
	chunk.SetBlock(0, 0, -60, StateStone)
	editedName, editedRoot, err := chunk.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := anvil.WriteChunk(dir, 0, 0, editedName, editedRoot); err != nil {
		t.Fatal(err)
	}
	_, reloadedRoot, exists, err := anvil.ReadChunk(dir, 0, 0)
	if err != nil || !exists {
		t.Fatalf("read edit: %v", err)
	}
	reloaded, err := chunkFromAnvil(editedName, reloadedRoot, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.GetBlock(0, 0, -60); got != StateStone {
		t.Fatalf("edited block=%d", got)
	}
	o, _ := reloadedRoot.Object()
	if _, ok := o["structures"]; !ok {
		t.Fatal("vanilla structures tag lost")
	}
}

// Set ATLAS_FLAT_REFERENCE_WORLD to a fresh 1.21.1 flat world made by the
// supplied JAR with features, lakes, and structure_overrides disabled.
func TestFlatBlocksAndBiomesAgainstVanillaGenerator(t *testing.T) {
	dir := os.Getenv("ATLAS_FLAT_REFERENCE_WORLD")
	if dir == "" {
		t.Skip("no independent vanilla flat reference world")
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.flat == nil || !v.flat.exact {
		t.Fatal("reference is not a flat world with features and structures disabled")
	}
	actual, err := v.loadChunk(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := v.generateFlat(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	for y := int32(MinY); y < MaxY; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				if got, want := generated.GetBlock(x, z, y), actual.GetBlock(x, z, y); got != want {
					t.Fatalf("block %d,%d,%d: generated=%d vanilla=%d", x, y, z, got, want)
				}
			}
		}
	}
	if !reflect.DeepEqual(generated.biomes, actual.biomes) {
		t.Fatal("generated biome cells differ from vanilla")
	}
}
