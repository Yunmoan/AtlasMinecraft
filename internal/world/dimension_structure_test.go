package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDimensionStructureAnvilRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	metadata, err := os.ReadFile(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	biome, err := worldgen.NewDimensionBiomeSampler(seed, "nether")
	if err != nil {
		t.Fatal(err)
	}
	fortress, err := worldgen.GenerateNetherFortressStructureStart(seed, 14, 8, biome.BiomeAtQuart)
	if err != nil || fortress == nil {
		t.Fatal("pinned fortress rejected", err)
	}
	positions := map[[2]int32]bool{{14, 8}: true, {4, -4}: true, {8, 20}: true}
	var chests, spawners []worldgen.FeaturePosition
	for _, piece := range fortress.Pieces {
		if piece.ID == "minecraft:nemt" && len(spawners) == 0 {
			p := piece.BlockPosition(3, 5, 5)
			spawners = append(spawners, p)
			positions[[2]int32{p.X >> 4, p.Z >> 4}] = true
		}
		flag, _ := piece.Data["Chest"].Number()
		if flag != 0 && len(chests) == 0 {
			x := int32(1)
			if piece.ID == "minecraft:nesclt" {
				x = 3
			}
			p := piece.BlockPosition(x, 2, 3)
			chests = append(chests, p)
			positions[[2]int32{p.X >> 4, p.Z >> 4}] = true
		}
	}
	if len(chests) == 0 || len(spawners) == 0 {
		t.Fatal("fortress fixture lacks chest/spawner")
	}
	w, err := OpenDimension(noViewers{}, 4, dir, "nether")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if w.store.vanilla.dimensionGenerator == nil || !w.store.vanilla.dimensionGenerator.StructuresEnabled() {
		t.Fatal("saved generate_features was not applied to Nether generator")
	}
	for pos := range positions {
		if _, err := w.LoadChunk(pos[0], pos[1]); err != nil {
			t.Fatal(err)
		}
	}
	editX, editZ := int32(14*16+8), int32(8*16+8)
	if err := w.SetBlock(editX, 200, editZ, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(dir, "level.dat"))
	if err != nil || !bytes.Equal(metadata, after) {
		t.Fatal("Nether structures rewrote shared level.dat", err)
	}
	saved := make(map[[2]int32]nbt.Object)
	for pos := range positions {
		_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "DIM-1", "region"), pos[0], pos[1])
		if err != nil || !exists {
			t.Fatal("dimension chunk not saved", err)
		}
		root, _ := tag.Object()
		saved[pos] = root
		coverage, _ := root.Object("AtlasWorldgen")
		complete, _ := coverage.Number("Complete")
		if complete != 0 {
			t.Fatal("partial structure coverage advertised complete")
		}
	}
	for _, entry := range []struct {
		pos  [2]int32
		name string
	}{{[2]int32{14, 8}, "minecraft:fortress"}, {[2]int32{4, -4}, "minecraft:nether_fossil"}, {[2]int32{8, 20}, "minecraft:ruined_portal_nether"}} {
		structures, _ := saved[entry.pos].Object("structures")
		starts, _ := structures.Object("starts")
		start, found := starts.Object(entry.name)
		if !found {
			t.Fatalf("saved start missing %s", entry.name)
		}
		pieces, _ := start["Children"].Elements()
		if len(pieces) == 0 {
			t.Fatal("saved structure has no pieces")
		}
		refs, _ := structures.Object("References")
		if refs[entry.name].Type != nbt.LongArray {
			t.Fatal("saved structure references missing")
		}
	}
	for _, expectation := range []struct {
		positions []worldgen.FeaturePosition
		id        string
	}{{chests, "minecraft:chest"}, {spawners, "minecraft:mob_spawner"}} {
		for _, p := range expectation.positions {
			entities, _ := saved[[2]int32{p.X >> 4, p.Z >> 4}]["block_entities"].Elements()
			found := false
			for _, tag := range entities {
				entity, _ := tag.Object()
				x, _ := entity.Number("x")
				y, _ := entity.Number("y")
				z, _ := entity.Number("z")
				if x != int64(p.X) || y != int64(p.Y) || z != int64(p.Z) {
					continue
				}
				id, _ := entity.Text("id")
				if id != expectation.id {
					t.Fatal("saved block entity id mismatch")
				}
				if id == "minecraft:chest" {
					loot, _ := entity.Text("LootTable")
					if loot != "minecraft:chests/nether_bridge" {
						t.Fatal("saved fortress loot missing")
					}
				} else {
					spawn, _ := entity.Object("SpawnData")
					actor, _ := spawn.Object("entity")
					id, _ := actor.Text("id")
					if id != "minecraft:blaze" {
						t.Fatal("saved blaze spawner missing")
					}
				}
				found = true
			}
			if !found {
				t.Fatalf("saved %s missing %v", expectation.id, p)
			}
		}
	}
	reopened, err := OpenDimension(noViewers{}, 4, dir, "nether")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for pos := range positions {
		if _, err := reopened.LoadChunk(pos[0], pos[1]); err != nil {
			t.Fatal(err)
		}
		chunk, err := reopened.store.vanilla.loadBlockChunk(pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := chunk.anvilRoot.Object()
		if !reflect.DeepEqual(root["structures"], saved[pos]["structures"]) {
			t.Fatal("restart changed structure NBT")
		}
		if !reflect.DeepEqual(root["block_entities"], saved[pos]["block_entities"]) {
			t.Fatal("restart changed structure block entities")
		}
	}
	id, err := reopened.BlockAt(editX, 200, editZ)
	if err != nil || id != StateStone {
		t.Fatal("restart regenerated player-edited structure chunk", err)
	}
}
