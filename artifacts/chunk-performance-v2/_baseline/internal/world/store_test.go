package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The user-supplied vanilla/world can be regenerated with a different world
// preset. Tests that assert the historical superflat layers need that preset.
func requireLegacyFlatSample(t *testing.T) {
	t.Helper()
	v, err := openVanilla(filepath.Join("..", "..", "vanilla", "world"))
	if os.IsNotExist(err) {
		t.Skip("vanilla sample unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	want := []flatLayer{{state: StateBedrock, height: 1}, {state: StateDirt, height: 2}, {state: StateGrass, height: 1}}
	if v.flat == nil || !reflect.DeepEqual(v.flat.layers, want) {
		t.Skip("vanilla sample is not the original four-layer superflat world")
	}
}

func TestVanillaPlayerDataRoundTrip(t *testing.T) {
	source := filepath.Join("..", "..", "vanilla", "world")
	const playerID = "167a8c6b-831e-32fb-9885-01128e28c2ec"
	level, err := os.ReadFile(filepath.Join(source, "level.dat"))
	if os.IsNotExist(err) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	playerData, err := os.ReadFile(filepath.Join(source, "playerdata", playerID+".dat"))
	if os.IsNotExist(err) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "playerdata"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "level.dat"), level, 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "playerdata", playerID+".dat")
	if err := os.WriteFile(path, playerData, 0644); err != nil {
		t.Fatal(err)
	}
	idBytes, err := hex.DecodeString(strings.ReplaceAll(playerID, "-", ""))
	if err != nil {
		t.Fatal(err)
	}
	var uuid [16]byte
	copy(uuid[:], idBytes)
	s, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, ok, err := s.LoadPlayer(uuid)
	if err != nil || !ok {
		t.Fatalf("load player: found=%v error=%v", ok, err)
	}
	p.X += 3.5
	p.Y += 1
	p.Flying = !p.Flying
	if err := s.SavePlayer(uuid, p); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LoadPlayer(uuid)
	if err != nil || !ok || got != p {
		t.Fatalf("saved player: %+v, found=%v, error=%v", got, ok, err)
	}
	backup, err := os.ReadFile(path + "_old")
	if err != nil || !bytes.Equal(backup, playerData) {
		t.Fatalf("player backup mismatch: %v", err)
	}
	before, err := readGzip(filepath.Join(source, "playerdata", playerID+".dat"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := readGzip(path)
	if err != nil {
		t.Fatal(err)
	}
	_, beforeRoot, err := nbt.Parse(before)
	if err != nil {
		t.Fatal(err)
	}
	_, afterRoot, err := nbt.Parse(after)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := beforeRoot.Object()
	a, _ := afterRoot.Object()
	if !reflect.DeepEqual(b["Inventory"], a["Inventory"]) || !reflect.DeepEqual(b["EnderItems"], a["EnderItems"]) {
		t.Fatal("unmodified player inventory tags changed")
	}
}

func TestRejectedVanillaEditDoesNotCreateChunk(t *testing.T) {
	dir := t.TempDir()
	// A structured flat save remains unsupported. This rejection must not
	// depend on whether the optional sample happens to be a normal world.
	root := newWorldLevel("unsupported", 1234, true)
	object, _ := root.Object()
	data, _ := object.Object("Data")
	settings, _ := data.Object("WorldGenSettings")
	dimensions, _ := settings.Object("dimensions")
	overworld, _ := dimensions.Object("minecraft:overworld")
	generator, _ := overworld.Object("generator")
	flat, _ := generator.Object("settings")
	flat["structure_overrides"] = nbt.Tag{Type: nbt.List, ListType: nbt.String, Value: []nbt.Tag{{Type: nbt.String, Value: "minecraft:villages"}}}
	if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), "", root); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(0, -60, 0, StateStone); err == nil {
		t.Fatal("missing structured chunk edit unexpectedly succeeded")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "blocks.log")); !os.IsNotExist(err) {
		t.Fatalf("legacy journal created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "region")); !os.IsNotExist(err) {
		t.Fatalf("rejected edit created a region: %v", err)
	}
}

func TestFlatWorldWithoutStructuresGeneratesAndSavesMissingChunk(t *testing.T) {
	requireLegacyFlatSample(t)
	source := filepath.Join("..", "..", "vanilla", "world", "level.dat")
	level, err := os.ReadFile(source)
	if os.IsNotExist(err) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "level.dat"), level, 0644); err != nil {
		t.Fatal(err)
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := v.root.Object()
	data, _ := root.Object("Data")
	settings, _ := data.Object("WorldGenSettings")
	dimensions, _ := settings.Object("dimensions")
	overworld, _ := dimensions.Object("minecraft:overworld")
	generator, _ := overworld.Object("generator")
	flat, _ := generator.Object("settings")
	flat["structure_overrides"] = nbt.Tag{Type: nbt.List, ListType: nbt.String, Value: []nbt.Tag{}}
	flat["features"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	flat["lakes"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), v.name, v.root); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	if warning := w.GenerationWarning(); warning != "" {
		t.Fatalf("unexpected warning: %s", warning)
	}
	if _, err := w.LoadChunk(40, 40); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	name, saved, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), 40, 40)
	if err != nil || !exists {
		t.Fatalf("generated Anvil chunk: %v", err)
	}
	c, err := chunkFromAnvil(name, saved, 40, 40)
	if err != nil {
		t.Fatal(err)
	}
	for _, pos := range [][2]int{{0, 0}, {15, 15}} {
		if c.GetBlock(pos[0], pos[1], -64) != StateBedrock || c.GetBlock(pos[0], pos[1], -61) != StateGrass || c.GetBlock(pos[0], pos[1], -60) != StateAir {
			t.Fatalf("wrong flat layers at %v", pos)
		}
	}
}

func TestVanillaWorldEditWritesAnvilWithoutTouchingSource(t *testing.T) {
	requireLegacyFlatSample(t)
	source := filepath.Join("..", "..", "vanilla", "world")
	level, err := os.ReadFile(filepath.Join(source, "level.dat"))
	if os.IsNotExist(err) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	region, err := os.ReadFile(filepath.Join(source, "region", "r.0.0.mca"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "region"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "level.dat"), level, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "region", "r.0.0.mca"), region, 0644); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.PreGenerateSpawn(0, 0); err != nil {
		t.Fatal(err)
	}
	surface, err := w.SurfaceY(0, 0)
	if err != nil || surface != -60 {
		t.Fatalf("surface=%d err=%v", surface, err)
	}
	if err := w.SetBlock(0, -60, 0, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), 0, 0)
	if err != nil || !exists {
		t.Fatalf("saved chunk: %v", err)
	}
	c, err := chunkFromAnvil(name, root, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.GetBlock(0, 0, -60); got != StateStone {
		t.Fatalf("saved block=%d", got)
	}
	originalAfter, err := os.ReadFile(filepath.Join(source, "region", "r.0.0.mca"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(region, originalAfter) {
		t.Fatal("source vanilla region was modified")
	}
}

type noViewers struct{}

func (noViewers) ViewersOf(x, z int32) []PacketSink { return nil }
