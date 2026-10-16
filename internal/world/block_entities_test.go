package world

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/worldgen"
	"bufio"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestBlockEntitySurvivesOwningBlockStateChange(t *testing.T) {
	ch := NewChunk(-1, -2)
	state := func(facing string) uint32 {
		t.Helper()
		id, err := blockstate.BlockStateID("minecraft:chest", map[string]string{"facing": facing, "type": "single", "waterlogged": "false"})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	north, east := state("north"), state("east")
	ch.SetBlock(15, 0, -60, north)
	entity := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"id":            {Type: nbt.String, Value: "minecraft:chest"},
		"x":             {Type: nbt.Int, Value: int32(-1)},
		"y":             {Type: nbt.Int, Value: int32(-60)},
		"z":             {Type: nbt.Int, Value: int32(-32)},
		"LootTable":     {Type: nbt.String, Value: "minecraft:chests/simple_dungeon"},
		"LootTableSeed": {Type: nbt.Long, Value: int64(1234)},
	}}
	entities := nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{entity}}
	ch.anvilRoot = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"block_entities": entities}}
	for _, next := range []uint32{north, east} {
		ch.SetBlock(15, 0, -60, next)
		root, _ := ch.anvilRoot.Object()
		if !reflect.DeepEqual(root["block_entities"], entities) {
			t.Fatal("same-block state change discarded or altered the chest entity")
		}
	}
	ch.SetBlock(15, 0, -60, StateStone)
	root, _ := ch.anvilRoot.Object()
	remaining, _ := root["block_entities"].Elements()
	if len(remaining) != 0 {
		t.Fatal("replacement retained the chest entity at negative world coordinates")
	}
}

func testGeneratedFeatureChunk(t *testing.T, kind string) *Chunk {
	t.Helper()
	terrain := &worldgen.NoiseChunkResult{X: 0, Z: 0, MinY: MinY, Height: WorldHeight, Blocks: make([]worldgen.NoiseMaterial, WorldHeight*256), SurfaceBlocks: make(map[int]worldgen.SurfaceBlockState)}
	for y := int32(MinY); y < MaxY; y++ {
		for z := int32(0); z < 16; z++ {
			for x := int32(0); x < 16; x++ {
				i, _ := terrain.Index(x, y, z)
				if kind == "well" {
					if y < 40 {
						terrain.Blocks[i] = worldgen.NoiseStone
					} else if y == 40 {
						terrain.SurfaceBlocks[i] = worldgen.SurfaceBlockState{Name: "minecraft:sand"}
					}
				} else if !(x >= 11 && z == 6 && y >= 41 && y <= 42) {
					terrain.Blocks[i] = worldgen.NoiseStone
				}
			}
		}
	}
	region, err := worldgen.NewOverworldFeatureRegion(terrain)
	if err != nil {
		t.Fatal(err)
	}
	r := worldgen.NewWorldgenRandom(0)
	place := worldgen.PlaceDesertWellFeature
	if kind == "room" {
		place = worldgen.PlaceMonsterRoomFeature
	}
	if placed, err := place(region, r, worldgen.FeaturePosition{X: 8, Y: 41, Z: 8}); err != nil || !placed {
		t.Fatalf("feature %s failed: %v", kind, err)
	}
	generated := &worldgen.GeneratedOverworldChunk{Terrain: terrain, BlockEntities: region.BlockEntitiesInChunk(0, 0)}
	for si := range generated.Biomes {
		for i := range generated.Biomes[si] {
			generated.Biomes[si][i] = "minecraft:desert"
		}
	}
	chunk, err := chunkFromGeneratedOverworld(generated)
	if err != nil {
		t.Fatal(err)
	}
	return chunk
}

func TestGeneratedBlockEntitiesPersistAndReplacementRemovesThem(t *testing.T) {
	for _, kind := range []string{"well", "room"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			seed := int64(0)
			if err := CreateNormalWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			save, err := openVanilla(dir)
			if err != nil {
				t.Fatal(err)
			}
			chunk := testGeneratedFeatureChunk(t, kind)
			root, _ := chunk.anvilRoot.Object()
			entities, _ := root["block_entities"].Elements()
			if len(entities) < 2 {
				t.Fatal("feature entities missing from bridge")
			}
			if err := save.persistChunk(chunk); err != nil {
				t.Fatal(err)
			}
			name, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), 0, 0)
			if err != nil || !exists {
				t.Fatal(err)
			}
			loaded, err := chunkFromAnvil(name, tag, 0, 0)
			if err != nil {
				t.Fatal(err)
			}
			loadedRoot, _ := loaded.anvilRoot.Object()
			if !reflect.DeepEqual(loadedRoot["block_entities"], root["block_entities"]) {
				t.Fatal("block entities or loot seeds changed after Anvil roundtrip")
			}
			entity, _ := entities[0].Object()
			x, _ := entity.Number("x")
			y, _ := entity.Number("y")
			z, _ := entity.Number("z")
			loaded.SetBlock(int(x&15), int(z&15), int32(y), StateStone)
			if err := save.persistChunk(loaded); err != nil {
				t.Fatal(err)
			}
			_, tag, _, err = anvil.ReadChunk(filepath.Join(dir, "region"), 0, 0)
			if err != nil {
				t.Fatal(err)
			}
			savedRoot, _ := tag.Object()
			remaining, _ := savedRoot["block_entities"].Elements()
			if len(remaining) != len(entities)-1 {
				t.Fatal("replaced entity was retained or unrelated entities removed")
			}
		})
	}
}

func TestGeneratedBlockEntityPacketsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/block_entity_packets_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	types, packets := map[string]int32{}, map[string][]byte{}
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		f := strings.Fields(scan.Text())
		if f[0] == "entity_type" {
			value, err := strconv.Atoi(f[2])
			if err != nil {
				t.Fatal(err)
			}
			types[f[1]] = int32(value)
		} else if f[0] == "entity_packet" {
			packets[f[1]], err = hex.DecodeString(f[2])
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	for id, typeID := range generatedBlockEntityTypes {
		if types[id] != typeID {
			t.Fatalf("registry ID mismatch for %s", id)
		}
		entity := nbt.Object{"id": {Type: nbt.String, Value: id}, "x": {Type: nbt.Int, Value: int32(-1)}, "y": {Type: nbt.Int, Value: int32(-60)}, "z": {Type: nbt.Int, Value: int32(-16)}, "LootTable": {Type: nbt.String, Value: "minecraft:test/private_loot"}}
		if id == "minecraft:mob_spawner" {
			chunk := testGeneratedFeatureChunk(t, "room")
			root, _ := chunk.anvilRoot.Object()
			values, _ := root["block_entities"].Elements()
			for _, tag := range values {
				o, _ := tag.Object()
				if name, _ := o.Text("id"); name == id {
					for name, tag := range o {
						if name != "x" && name != "y" && name != "z" {
							entity[name] = tag
						}
					}
				}
			}
			entity["SpawnData"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"entity": {Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:skeleton"}}}}}
		}
		ch := NewChunk(-1, -1)
		ch.anvilRoot = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"block_entities": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: entity}}}}}
		w := proto.NewWriter()
		writeChunkBlockEntities(w, ch)
		got := w.Payload()
		if got[0] != 1 || len(got) < 5 || !reflect.DeepEqual(got[1:5], packets[id][:4]) {
			t.Fatalf("entity header mismatch for %s: %x", id, got)
		}
		actual, n, err := nbt.ParseNetwork(got[5:])
		if err != nil || n != len(got)-5 {
			t.Fatalf("entity network tag invalid: %v", err)
		}
		expected, _, err := nbt.ParseNetwork(packets[id][4:])
		if err != nil || !reflect.DeepEqual(actual, expected) {
			t.Fatalf("entity update tag differs from JAR for %s: actual=%v expected=%v err=%v", id, actual, expected, err)
		}
		if strings.Contains(string(got), "private_loot") || strings.Contains(string(got), "SpawnPotentials") {
			t.Fatal("chunk packet exposed server-only data")
		}
	}
}
