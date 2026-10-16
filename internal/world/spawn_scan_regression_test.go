package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSafeSpawnColumnRejectsHazardsAndObstructions(t *testing.T) {
	state := func(name string, properties map[string]string) uint32 {
		t.Helper()
		id, err := BlockStateID(name, properties)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	grass := state("minecraft:short_grass", nil)
	water := state("minecraft:water", nil)
	lava := state("minecraft:lava", nil)
	magma := state("minecraft:magma_block", nil)
	powder := state("minecraft:powder_snow", nil)
	fire := state("minecraft:fire", nil)
	soulFire := state("minecraft:soul_fire", nil)
	berries := state("minecraft:sweet_berry_bush", map[string]string{"age": "3"})
	witherRose := state("minecraft:wither_rose", nil)
	slab := state("minecraft:stone_slab", map[string]string{"type": "bottom", "waterlogged": "false"})
	waterloggedSlab := state("minecraft:stone_slab", map[string]string{"type": "top", "waterlogged": "true"})
	tests := []struct {
		name   string
		ground uint32
		y      int32
		feet   uint32
		head   uint32
		want   bool
	}{
		{name: "dry ground", ground: StateDirt, y: 63, want: true},
		{name: "harmless short grass", ground: StateDirt, y: 63, feet: grass, want: true},
		{name: "water surface", ground: StateDirt, y: 63, feet: water},
		{name: "lava surface", ground: StateDirt, y: 63, feet: lava},
		{name: "void", y: 63},
		{name: "magma support", ground: magma, y: 63},
		{name: "powder snow", ground: StateDirt, y: 63, feet: powder},
		{name: "fire", ground: StateDirt, y: 63, feet: fire},
		{name: "soul fire", ground: StateDirt, y: 63, feet: soulFire},
		{name: "mature berry bush", ground: StateDirt, y: 63, feet: berries},
		{name: "wither rose", ground: StateDirt, y: 63, feet: witherRose},
		{name: "collision at feet", ground: StateDirt, y: 63, feet: slab},
		{name: "collision at head", ground: StateDirt, y: 63, head: slab},
		{name: "fluid at head", ground: StateDirt, y: 63, head: water},
		{name: "waterlogged support", ground: waterloggedSlab, y: 63},
		{name: "build ceiling", ground: StateDirt, y: MaxY - 1},
		{name: "head above build ceiling", ground: StateDirt, y: MaxY - 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := NewChunk(0, 0)
			ch.SetBlock(0, 0, tt.y, tt.ground)
			ch.SetBlock(0, 0, tt.y+1, tt.feet)
			ch.SetBlock(0, 0, tt.y+2, tt.head)
			y, ok := safeSpawnY(ch, 0, 0)
			if ok != tt.want || ok && y != tt.y+1 {
				t.Fatalf("safeSpawnY = (%d, %v), want (%d, %v)", y, ok, tt.y+1, tt.want)
			}
		})
	}
}

// Synthetic terrain isolates spawn selection from the still-incomplete normal
// world generator, while using a real level.dat and Anvil persistence path.
func spawnRegressionWorld(t *testing.T) (*World, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "normal")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	w.lighting = nil
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	})
	return w, dir
}

func spawnRegressionChunk(cx, cz int32, safe bool) (*Chunk, error) {
	ch := NewChunk(cx, cz)
	if safe {
		ch.SetBlock(3, 6, 70, StateDirt)
	}
	if err := ch.initializeAnvilRoot(); err != nil {
		return nil, err
	}
	return ch, nil
}

func spawnRegressionLevel(t *testing.T, dir string) nbt.Object {
	t.Helper()
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
	return level
}

func spawnRegressionReady(t *testing.T, dir string) int32 {
	t.Helper()
	marker, ok := spawnRegressionLevel(t, dir).Object("AtlasWorldgen")
	if !ok {
		t.Fatal("missing AtlasWorldgen marker")
	}
	ready, ok := marker.Number("SpawnReady")
	if !ok {
		t.Fatal("missing SpawnReady marker")
	}
	return int32(ready)
}

func TestInitialSpawnScansBeyondPreloadRadiusAndRecenters(t *testing.T) {
	for _, radius := range []int{0, 1} {
		t.Run(fmt.Sprintf("radius %d", radius), func(t *testing.T) {
			w, dir := spawnRegressionWorld(t)
			x, z := w.SpawnXZ()
			cx, cz := x>>4, z>>4
			var mu sync.Mutex
			var loaded []chunkKey
			w.loadChunk = func(x, z int32) (*Chunk, error) {
				mu.Lock()
				loaded = append(loaded, chunkKey{x, z})
				mu.Unlock()
				return spawnRegressionChunk(x, z, x == cx+1 && z == cz)
			}
			if err := w.PrepareSpawn(context.Background(), cx, cz, radius, nil); err != nil {
				t.Fatal(err)
			}
			wantX, wantY, wantZ := (cx+1)*16+3, int32(71), cz*16+6
			if x, y, z, _ := w.SpawnPoint(); x != wantX || y != wantY || z != wantZ {
				t.Fatalf("spawn = %d %d %d, want %d %d %d", x, y, z, wantX, wantY, wantZ)
			}
			mu.Lock()
			order := append([]chunkKey(nil), loaded...)
			mu.Unlock()
			if len(order) < 2 || order[0] != (chunkKey{cx, cz}) || order[1] != (chunkKey{cx + 1, cz}) {
				t.Fatalf("initial search order = %v, want center then east", order)
			}
			for dx := -radius; dx <= radius; dx++ {
				for dz := -radius; dz <= radius; dz++ {
					x, z := cx+1+int32(dx), cz+int32(dz)
					if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), x, z); err != nil || !exists {
						t.Fatalf("final spawn neighborhood chunk %d,%d not saved: exists=%v err=%v", x, z, exists, err)
					}
				}
			}
			level := spawnRegressionLevel(t, dir)
			for name, want := range map[string]int32{"SpawnX": wantX, "SpawnY": wantY, "SpawnZ": wantZ} {
				if got, _ := level.Number(name); int32(got) != want {
					t.Fatalf("level.dat %s = %v, want %d", name, got, want)
				}
			}
			if ready := spawnRegressionReady(t, dir); ready != 1 {
				t.Fatalf("SpawnReady = %d, want 1", ready)
			}
		})
	}
}

func TestInitialSpawnFailureLeavesInitializationPending(t *testing.T) {
	t.Run("all 121 chunks unsafe", func(t *testing.T) {
		w, dir := spawnRegressionWorld(t)
		x, y, z, angle := w.SpawnPoint()
		var mu sync.Mutex
		visited := make(map[chunkKey]bool)
		w.loadChunk = func(cx, cz int32) (*Chunk, error) {
			mu.Lock()
			visited[chunkKey{cx, cz}] = true
			mu.Unlock()
			return spawnRegressionChunk(cx, cz, false)
		}
		err := w.PrepareSpawn(context.Background(), x>>4, z>>4, 0, nil)
		if err == nil || !strings.Contains(err.Error(), "no safe initial spawn") {
			t.Fatalf("unsafe search error = %v", err)
		}
		mu.Lock()
		count := len(visited)
		for key := range visited {
			if key.x < x>>4-5 || key.x > x>>4+5 || key.z < z>>4-5 || key.z > z>>4+5 {
				mu.Unlock()
				t.Fatalf("searched chunk %v outside vanilla search square", key)
			}
		}
		mu.Unlock()
		if count != 121 {
			t.Fatalf("searched %d chunks, want 121", count)
		}
		if gx, gy, gz, ga := w.SpawnPoint(); gx != x || gy != y || gz != z || ga != angle {
			t.Fatalf("failed search moved placeholder to %d %d %d %v", gx, gy, gz, ga)
		}
		if ready := spawnRegressionReady(t, dir); ready != 0 {
			t.Fatalf("failed search marked SpawnReady = %d", ready)
		}
	})
	t.Run("load failure", func(t *testing.T) {
		w, dir := spawnRegressionWorld(t)
		x, z := w.SpawnXZ()
		failure := errors.New("fixture load failure")
		w.loadChunk = func(int32, int32) (*Chunk, error) { return nil, failure }
		if err := w.PrepareSpawn(context.Background(), x>>4, z>>4, 0, nil); !errors.Is(err, failure) {
			t.Fatalf("load error = %v, want %v", err, failure)
		}
		if ready := spawnRegressionReady(t, dir); ready != 0 {
			t.Fatalf("load failure marked SpawnReady = %d", ready)
		}
	})
	t.Run("cancellation during search", func(t *testing.T) {
		w, dir := spawnRegressionWorld(t)
		x, z := w.SpawnXZ()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		w.loadChunk = func(cx, cz int32) (*Chunk, error) {
			cancel()
			return spawnRegressionChunk(cx, cz, false)
		}
		if err := w.PrepareSpawn(ctx, x>>4, z>>4, 0, nil); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled search error = %v", err)
		}
		if ready := spawnRegressionReady(t, dir); ready != 0 {
			t.Fatalf("cancellation marked SpawnReady = %d", ready)
		}
	})
}

func TestExplicitUnchangedSpawnCompletesInitialization(t *testing.T) {
	w, dir := spawnRegressionWorld(t)
	x, y, z, angle := w.SpawnPoint()
	changed, err := w.SetSpawn(x, y, z, angle)
	if err != nil || changed {
		t.Fatalf("unchanged explicit spawn: changed=%v err=%v", changed, err)
	}
	if ready := spawnRegressionReady(t, dir); ready != 1 {
		t.Fatalf("explicit unchanged spawn left SpawnReady = %d", ready)
	}
	// The supplied center is empty. Rescanning it would move the spawn or fail.
	w.loadChunk = func(cx, cz int32) (*Chunk, error) { return spawnRegressionChunk(cx, cz, false) }
	if err := w.PrepareSpawn(context.Background(), x>>4, z>>4, 0, nil); err != nil {
		t.Fatal(err)
	}
	if gx, gy, gz, ga := w.SpawnPoint(); gx != x || gy != y || gz != z || ga != angle {
		t.Fatalf("initial scan overrode explicit unchanged spawn: %d %d %d %v", gx, gy, gz, ga)
	}
}

func TestExistingSaveWithoutSpawnReadyDoesNotRescan(t *testing.T) {
	w, dir := spawnRegressionWorld(t)
	x, y, z, angle := w.SpawnPoint()
	levelRoot, _ := w.store.vanilla.root.Object()
	level, _ := levelRoot.Object("Data")
	marker, _ := level.Object("AtlasWorldgen")
	delete(marker, "SpawnReady")
	if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), w.store.vanilla.name, w.store.vanilla.root); err != nil {
		t.Fatal(err)
	}
	w.loadChunk = func(cx, cz int32) (*Chunk, error) { return spawnRegressionChunk(cx, cz, false) }
	if err := w.PrepareSpawn(context.Background(), x>>4, z>>4, 0, nil); err != nil {
		t.Fatal(err)
	}
	if gx, gy, gz, ga := w.SpawnPoint(); gx != x || gy != y || gz != z || ga != angle {
		t.Fatalf("initial scan moved an existing save: %d %d %d %v", gx, gy, gz, ga)
	}
	storedMarker, _ := spawnRegressionLevel(t, dir).Object("AtlasWorldgen")
	if _, ok := storedMarker.Number("SpawnReady"); ok {
		t.Fatal("initial scan added SpawnReady to an existing save")
	}
}
