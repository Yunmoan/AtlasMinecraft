package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestSetSpawnPersistsAndSurvivesRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "flat")
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := w.SetSpawn(3, -60, 4, 15)
	if err != nil || !changed {
		t.Fatalf("set spawn: changed=%v error=%v", changed, err)
	}
	if x, y, z, angle := w.SpawnPoint(); x != 3 || y != -60 || z != 4 || angle != 15 {
		t.Fatalf("spawn in memory: %d %d %d %v", x, y, z, angle)
	}
	// Storing the identical point again reports no change so callers can
	// skip the spawn-position broadcast.
	changed, err = w.SetSpawn(3, -60, 4, 15)
	if err != nil || changed {
		t.Fatalf("unchanged spawn: changed=%v error=%v", changed, err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if x, y, z, angle := reopened.SpawnPoint(); x != 3 || y != -60 || z != 4 || angle != 15 {
		t.Fatalf("spawn after restart: %d %d %d %v", x, y, z, angle)
	}
	if _, err := reopened.SetSpawn(-30000001, -60, 0, 0); err == nil {
		t.Fatal("accepted out-of-bounds X")
	}
	if _, err := reopened.SetSpawn(0, MaxY, 0, 0); err == nil {
		t.Fatal("accepted out-of-bounds Y")
	}
}

// A new normal save stores the climate-searched spawn chunk's center block at
// generator spawn height with an AtlasWorldgen SpawnReady=0 marker until the
// boot-time scan refines it. An explicit /setworldspawn must complete that
// marker so generation and the scan can never override the chosen position.
func TestSetSpawnOverridesSpawnHeightPlaceholder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "normal")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	searchX, searchZ, err := worldgen.FindSpawnPosition(seed)
	if err != nil {
		t.Fatal(err)
	}
	wantX, wantZ := searchX>>4*16+8, searchZ>>4*16+8
	if x, y, z, _ := w.SpawnPoint(); x != wantX || y != 64 || z != wantZ {
		t.Fatalf("unexpected placeholder spawn: %d %d %d, want %d 64 %d", x, y, z, wantX, wantZ)
	}
	if _, err := w.SetSpawn(3, -50, 4, 0); err != nil {
		t.Fatal(err)
	}
	if err := w.PreGenerateSpawn(0, 0); err != nil {
		t.Fatal(err)
	}
	if x, y, z, _ := w.SpawnPoint(); x != 3 || y != -50 || z != 4 {
		t.Fatalf("spawn chunk generation overwrote the explicit spawn: %d %d %d", x, y, z)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := readGzip(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	object, _ := root.Object()
	level, _ := object.Object("Data")
	if y, _ := level.Number("SpawnY"); y != -50 {
		t.Fatalf("level.dat SpawnY: %v", y)
	}
}

// The boot-time scan refines the climate-selected spawn chunk into the first
// safe column, exactly like vanilla's setInitialSpawn spiral, and persists
// the initialized flag so later boots do not rescan.
func TestPrepareSpawnInitializesScannedSpawnPoint(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "normal")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	spawnX, spawnZ := w.SpawnXZ()
	if err := w.PrepareSpawn(context.Background(), spawnX>>4, spawnZ>>4, 1, nil); err != nil {
		t.Fatal(err)
	}
	x, y, z, _ := w.SpawnPoint()
	if x>>4 != spawnX>>4-1 && x>>4 != spawnX>>4 && x>>4 != spawnX>>4+1 {
		t.Fatalf("scanned spawn %d %d %d left the prepared neighborhood", x, y, z)
	}
	// The scan guarantees ground below the spawn feet and no fluid at either
	// the feet or the supporting block.
	below, err := w.BlockAt(x, y-1, z)
	if err != nil {
		t.Fatal(err)
	}
	water, err := BlockStateID("minecraft:water", nil)
	if err != nil {
		t.Fatal(err)
	}
	if below == StateAir || below == water {
		t.Fatalf("spawn %d %d %d stands on state %d", x, y, z, below)
	}
	feet, err := w.BlockAt(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if feet == water {
		t.Fatalf("spawn feet at %d %d %d stand in water", x, y, z)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := readGzip(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	object, _ := root.Object()
	level, _ := object.Object("Data")
	storedX, _ := level.Number("SpawnX")
	storedY, _ := level.Number("SpawnY")
	if int32(storedX) != x || int32(storedY) != y {
		t.Fatalf("level.dat spawn not persisted: %v %v", storedX, storedY)
	}
	marker, ok := level.Object("AtlasWorldgen")
	if !ok {
		t.Fatal("AtlasWorldgen marker missing")
	}
	if ready, _ := marker.Number("SpawnReady"); ready != 1 {
		t.Fatalf("SpawnReady marker: %v", ready)
	}
	// A second boot must not rescan or move the spawn.
	reopened, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.PrepareSpawn(context.Background(), x>>4, z>>4, 0, nil); err != nil {
		t.Fatal(err)
	}
	if gx, gy, gz, _ := reopened.SpawnPoint(); gx != x || gy != y || gz != z {
		t.Fatalf("restart moved the spawn: %d %d %d", gx, gy, gz)
	}
}

func TestSetPlayerSpawnWritesVanillaTags(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "flat")
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	var uuid [16]byte
	uuid[0] = 7
	if err := w.SetPlayerSpawn(uuid, 3, -60, 4, -90); err != nil {
		t.Fatal(err)
	}
	// A later position save must retain the spawn tags.
	if err := w.SavePlayer(uuid, PlayerPosition{X: 1, Y: -60, Z: 2}); err != nil {
		t.Fatal(err)
	}
	level := readPlayerTags(t, dir, uuid)
	if x, _ := level.Number("SpawnX"); x != 3 {
		t.Fatalf("SpawnX: %v", x)
	}
	if y, _ := level.Number("SpawnY"); y != -60 {
		t.Fatalf("SpawnY: %v", y)
	}
	if z, _ := level.Number("SpawnZ"); z != 4 {
		t.Fatalf("SpawnZ: %v", z)
	}
	if angle, ok := level["SpawnAngle"].Value.(float32); !ok || angle != -90 {
		t.Fatalf("SpawnAngle: %v", level["SpawnAngle"])
	}
	if forced, _ := level.Number("SpawnForced"); forced != 1 {
		t.Fatalf("SpawnForced: %v", forced)
	}
	if dimension, _ := level.Text("SpawnDimension"); dimension != "minecraft:overworld" {
		t.Fatalf("SpawnDimension: %q", dimension)
	}
	pos, ok := level["Pos"].Elements()
	if !ok || len(pos) != 3 {
		t.Fatal("position save lost after setPlayerSpawn")
	}
}

func readPlayerTags(t *testing.T, dir string, uuid [16]byte) nbt.Object {
	t.Helper()
	name := fmt.Sprintf("%x-%x-%x-%x-%x.dat", uuid[:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
	data, err := readGzip(filepath.Join(dir, "playerdata", name))
	if err != nil {
		t.Fatal(err)
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	object, _ := root.Object()
	return object
}
