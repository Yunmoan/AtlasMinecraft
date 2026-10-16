package world

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEndCityStructureAnvilRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	metadata, err := os.ReadFile(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	noise, err := worldgen.NewDimensionNoiseGenerator(seed, "end")
	if err != nil {
		t.Fatal(err)
	}
	biomes, err := worldgen.NewDimensionBiomeSampler(seed, "end")
	if err != nil {
		t.Fatal(err)
	}
	start, err := worldgen.GenerateEndCityStructureStart(seed, 145, -77, biomes.BiomeAtQuart, func(x, z int32) (int32, error) {
		column, err := noise.BaseStructureColumnAt(x, z)
		if err != nil {
			return 0, err
		}
		return column.OccupiedHeight(false), nil
	})
	if err != nil || start == nil || len(start.Pieces) != 94 {
		t.Fatal("pinned actual End City rejected", err)
	}
	positions := map[[2]int32]bool{{145, -77}: true}
	var cityBlock worldgen.FeaturePosition
	for _, piece := range start.Pieces {
		name, _ := piece.Data.Text("Template")
		if name == "ship" {
			for x := piece.Box.MinX >> 4; x <= piece.Box.MaxX>>4; x++ {
				for z := piece.Box.MinZ >> 4; z <= piece.Box.MaxZ>>4; z++ {
					positions[[2]int32{x, z}] = true
				}
			}
			template, err := worldgen.LoadEndCityTemplate(name)
			if err != nil {
				t.Fatal(err)
			}
			x, _ := piece.Data.Number("TPX")
			y, _ := piece.Data.Number("TPY")
			z, _ := piece.Data.Number("TPZ")
			rot, _ := piece.Data.Text("Rot")
			rotation := worldgen.TemplateRotationNone
			for i, text := range []string{"NONE", "CLOCKWISE_90", "CLOCKWISE_180", "COUNTERCLOCKWISE_90"} {
				if text == rot {
					rotation = worldgen.TemplateRotation(i)
				}
			}
			for _, block := range template.Palettes[0] {
				if block.State.Name == "minecraft:purpur_block" {
					p, err := worldgen.TransformTemplatePosition(block.Position, worldgen.FeaturePosition{}, worldgen.TemplateMirrorNone, rotation)
					if err != nil {
						t.Fatal(err)
					}
					cityBlock = worldgen.FeaturePosition{X: int32(x) + p.X, Y: int32(y) + p.Y, Z: int32(z) + p.Z}
					break
				}
			}
		}
	}
	w, err := OpenDimension(noViewers{}, 4, dir, "end")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	beforeActors := make(map[[16]byte]WorldEntity)
	actorCounts := make(map[[2]int32]int)
	shulkers, frames := 0, 0
	for pos := range positions {
		if _, err := w.LoadChunk(pos[0], pos[1]); err != nil {
			t.Fatal(err)
		}
		actors, err := w.EntitiesInChunk(pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		for _, actor := range actors {
			if actor.Type == "minecraft:shulker" {
				shulkers++
				health, _ := actor.Data["Health"].Value.(float32)
				if health != 30 {
					t.Fatal("generated shulker lost original Health30")
				}
			}
			if actor.Type == "minecraft:item_frame" {
				frames++
				item, _ := actor.Data.Object("Item")
				id, _ := item.Text("id")
				count, _ := item.Number("count")
				if id != "minecraft:elytra" || count != 1 {
					t.Fatal("End ship frame lost elytra")
				}
				if _, ok := actor.Data["TileX"]; !ok {
					t.Fatal("End ship frame lost hanging anchor")
				}
			}
			if actor.Type != "minecraft:shulker" && actor.Type != "minecraft:item_frame" {
				continue
			}
			if _, exists := beforeActors[actor.UUID]; exists {
				t.Fatal("duplicate End City actor UUID")
			}
			beforeActors[actor.UUID] = actor
			actorCounts[pos]++
		}
	}
	if shulkers != 4 || frames != 1 {
		t.Fatalf("actual End City actors %d shulkers %d frames", shulkers, frames)
	}
	id, err := w.BlockAt(cityBlock.X, cityBlock.Y, cityBlock.Z)
	if err != nil {
		t.Fatal(err)
	}
	state, err := blockstate.BlockStateByID(id)
	if err != nil || state.Name != "minecraft:purpur_block" {
		t.Fatal("End City metadata without ship blocks", state, err)
	}
	editX, editZ := int32(145*16+8), int32(-77*16+8)
	if err := w.SetBlock(editX, 240, editZ, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(dir, "level.dat"))
	if err != nil || !bytes.Equal(metadata, after) {
		t.Fatal("End City rewrote root level.dat", err)
	}
	saved := make(map[[2]int32]nbt.Object)
	chests := 0
	for pos := range positions {
		_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "DIM1", "region"), pos[0], pos[1])
		if err != nil || !exists {
			t.Fatal("End City chunk not saved", pos, err)
		}
		root, _ := tag.Object()
		saved[pos] = root
		structures, _ := root.Object("structures")
		refs, _ := structures.Object("References")
		if refs["minecraft:end_city"].Type != nbt.LongArray {
			t.Fatal("End City refs not saved")
		}
		if pos == [2]int32{145, -77} {
			starts, _ := structures.Object("starts")
			city, ok := starts.Object("minecraft:end_city")
			children, _ := city["Children"].Elements()
			if !ok || len(children) != 94 {
				t.Fatal("End City start Children not saved")
			}
		}
		entities, _ := root["block_entities"].Elements()
		for _, tag := range entities {
			data, _ := tag.Object()
			table, _ := data.Text("LootTable")
			if table == "minecraft:chests/end_city_treasure" {
				chests++
				if _, ok := data["LootTableSeed"]; !ok {
					t.Fatal("End City chest lost seed")
				}
			}
		}
		_, entityTag, exists, err := anvil.ReadChunk(filepath.Join(dir, "DIM1", "entities"), pos[0], pos[1])
		if err != nil || !exists && actorCounts[pos] > 0 {
			t.Fatal("End City entity region missing", err)
		}
		entityRoot, _ := entityTag.Object()
		actorTags, _ := entityRoot["Entities"].Elements()
		if len(actorTags) != actorCounts[pos] {
			t.Fatal("End City entity region actor count changed", pos)
		}
		for _, tag := range actorTags {
			data, _ := tag.Object()
			if _, ok := data["UUID"]; !ok {
				t.Fatal("End City saved actor lost UUID")
			}
		}
	}
	if chests != 2 {
		t.Fatalf("End City saved loot chests %d want2", chests)
	}
	reopened, err := OpenDimension(noViewers{}, 4, dir, "end")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	seen := make(map[[16]byte]bool)
	for pos := range positions {
		if _, err := reopened.LoadChunk(pos[0], pos[1]); err != nil {
			t.Fatal(err)
		}
		chunk, err := reopened.store.vanilla.loadBlockChunk(pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := chunk.anvilRoot.Object()
		if !reflect.DeepEqual(root["structures"], saved[pos]["structures"]) || !reflect.DeepEqual(root["block_entities"], saved[pos]["block_entities"]) {
			t.Fatal("End City NBT changed after restart", pos)
		}
		actors, err := reopened.EntitiesInChunk(pos[0], pos[1])
		if err != nil {
			t.Fatal(err)
		}
		for _, actor := range actors {
			if actor.Type == "minecraft:shulker" || actor.Type == "minecraft:item_frame" {
				if seen[actor.UUID] || !reflect.DeepEqual(actor, loadedEntitySnapshot(beforeActors[actor.UUID])) {
					t.Fatal("End City actor changed/duplicated on restart")
				}
				seen[actor.UUID] = true
			}
		}
	}
	if len(seen) != 5 {
		t.Fatal("End City restart lost actors")
	}
	id, err = reopened.BlockAt(editX, 240, editZ)
	if err != nil || id != StateStone {
		t.Fatal("End City restart regenerated player edit", err)
	}
	t.Logf("saved/reopened %d actual DIM1 chunks, 94 city pieces, two seeded loot chests, four shulkers and one elytra frame with stable UUID/NBT", len(positions))
}
