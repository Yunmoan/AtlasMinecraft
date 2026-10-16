package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"path/filepath"
	"reflect"
	"testing"
)

func TestOverworldMineshaftStructureAnvilRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	positions := [][2]int32{{-16, 22}, {-15, 19}, {-14, 22}, {-15, 17}, {-15, 20}, {-15, 22}, {-14, 21}, {-365, 369}}
	beforeActors := make(map[[16]byte]WorldEntity)
	for _, position := range positions {
		if _, err := w.LoadChunk(position[0], position[1]); err != nil {
			t.Fatal(err)
		}
		actors, err := w.EntitiesInChunk(position[0], position[1])
		if err != nil {
			t.Fatal(err)
		}
		for _, actor := range actors {
			if actor.Type != "minecraft:chest_minecart" {
				continue
			}
			table, _ := actor.Data.Text("LootTable")
			if table != "minecraft:chests/abandoned_mineshaft" {
				t.Fatal("generated minecart lost its loot table")
			}
			if _, ok := actor.Data["LootTableSeed"]; !ok {
				t.Fatal("generated minecart lost its original loot seed")
			}
			for _, value := range actor.Position {
				if value-float64(int64(value)) != 0.5 && value-float64(int64(value)) != -0.5 {
					t.Fatalf("minecart position lost block-center placement: %v", actor.Position)
				}
			}
			if _, duplicate := beforeActors[actor.UUID]; duplicate {
				t.Fatal("duplicate generated loot minecart")
			}
			beforeActors[actor.UUID] = actor
		}
	}
	if len(beforeActors) != 6 {
		t.Fatalf("generated loot minecarts=%d want6", len(beforeActors))
	}
	editX, editZ := int32(-16*16+8), int32(22*16+8)
	if err := w.SetBlock(editX, 280, editZ, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	saved := make(map[[2]int32]nbt.Object)
	spawners := 0
	for _, position := range positions {
		_, tag, found, err := anvil.ReadChunk(filepath.Join(dir, "region"), position[0], position[1])
		if err != nil || !found {
			t.Fatal("mineshaft chunk not saved", err)
		}
		root, _ := tag.Object()
		saved[position] = root
		entities, _ := root["block_entities"].Elements()
		for _, tag := range entities {
			entity, _ := tag.Object()
			if id, _ := entity.Text("id"); id != "minecraft:mob_spawner" {
				continue
			}
			spawn, _ := entity.Object("SpawnData")
			actor, _ := spawn.Object("entity")
			if id, _ := actor.Text("id"); id != "minecraft:cave_spider" {
				t.Fatal("mine spawner changed actor type")
			}
			spawners++
		}
	}
	if spawners != 2 {
		t.Fatalf("saved spider spawners=%d want2", spawners)
	}
	for _, fixture := range []struct {
		pos    [2]int32
		name   string
		pieces int
	}{{[2]int32{-16, 22}, "minecraft:mineshaft", 88}, {[2]int32{-365, 369}, "minecraft:mineshaft_mesa", 87}} {
		structures, _ := saved[fixture.pos].Object("structures")
		starts, _ := structures.Object("starts")
		start, found := starts.Object(fixture.name)
		pieces, _ := start["Children"].Elements()
		if !found || len(pieces) != fixture.pieces {
			t.Fatalf("saved %s lost children: %d", fixture.name, len(pieces))
		}
		if fixture.name == "minecraft:mineshaft" {
			flags := 0
			for _, tag := range pieces {
				piece, _ := tag.Object()
				hps, _ := piece.Number("hps")
				if hps != 0 {
					flags++
				}
			}
			if flags != 2 {
				t.Fatalf("saved unique spider spawner flags=%d want2", flags)
			}
		}
	}
	reopened, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	seen := make(map[[16]byte]bool)
	for _, position := range positions {
		if _, err := reopened.LoadChunk(position[0], position[1]); err != nil {
			t.Fatal(err)
		}
		chunk, err := reopened.store.vanilla.loadBlockChunk(position[0], position[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := chunk.anvilRoot.Object()
		if !reflect.DeepEqual(root["structures"], saved[position]["structures"]) || !reflect.DeepEqual(root["block_entities"], saved[position]["block_entities"]) {
			t.Fatalf("restart changed mine structures/spawners at %v", position)
		}
		actors, err := reopened.EntitiesInChunk(position[0], position[1])
		if err != nil {
			t.Fatal(err)
		}
		for _, actor := range actors {
			if actor.Type != "minecraft:chest_minecart" {
				continue
			}
			if seen[actor.UUID] || !reflect.DeepEqual(actor, loadedEntitySnapshot(beforeActors[actor.UUID])) {
				t.Fatal("restart duplicated or changed generated loot minecart")
			}
			seen[actor.UUID] = true
		}
	}
	if len(seen) != len(beforeActors) {
		t.Fatal("restart lost generated loot minecart")
	}
	id, err := reopened.BlockAt(editX, 280, editZ)
	if err != nil || id != StateStone {
		t.Fatal("restart regenerated edited mine chunk", err)
	}
	t.Logf("saved/reopened eight actual mine chunks, two unique spider spawners and six complete loot minecarts")
}
