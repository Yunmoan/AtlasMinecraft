package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"bufio"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedSculkEntitiesAgainstVanillaAndAnvil(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/block_entity_packets_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	count := 0
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if fields[0] != "entity_default" || !strings.HasPrefix(fields[1], "minecraft:sculk_") {
			continue
		}
		t.Run(fields[1], func(t *testing.T) {
			data, err := hex.DecodeString(fields[2])
			if err != nil {
				t.Fatal(err)
			}
			expected, used, err := nbt.ParseNetwork(data)
			if err != nil || used != len(data) {
				t.Fatalf("invalid JAR NBT: %v", err)
			}
			state := worldgen.SurfaceBlockState{Name: fields[1]}
			terrain := &worldgen.NoiseChunkResult{X: -1, Z: -1, MinY: MinY, Height: WorldHeight, Blocks: make([]worldgen.NoiseMaterial, WorldHeight*256), SurfaceBlocks: make(map[int]worldgen.SurfaceBlockState)}
			region, err := worldgen.NewOverworldFeatureRegion(terrain)
			if err != nil {
				t.Fatal(err)
			}
			p := worldgen.FeaturePosition{X: -1, Y: -60, Z: -16}
			if err := region.SetBlock(p, state); err != nil {
				t.Fatal(err)
			}
			if err := region.SetBlockEntity(p, state.Name, worldgen.DefaultSculkBlockEntity(state)); err != nil {
				t.Fatal(err)
			}
			entities := region.BlockEntitiesInChunk(-1, -1)
			if len(entities) != 1 {
				t.Fatal("sculk entity missing from generation region")
			}
			actual := nbt.Tag{Type: nbt.Compound, Value: entities[0]}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("generated default differs from official entity:\nactual %v\nexpected %v", actual, expected)
			}
			generated := &worldgen.GeneratedOverworldChunk{Terrain: terrain, BlockEntities: entities}
			for si := range generated.Biomes {
				for i := range generated.Biomes[si] {
					generated.Biomes[si][i] = "minecraft:deep_dark"
				}
			}
			chunk, err := chunkFromGeneratedOverworld(generated)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			seed := int64(0)
			if err := CreateNormalWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			save, err := openVanilla(dir)
			if err != nil {
				t.Fatal(err)
			}
			if err := save.persistChunk(chunk); err != nil {
				t.Fatal(err)
			}
			name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), -1, -1)
			if err != nil || !exists {
				t.Fatal(err)
			}
			loaded, err := chunkFromAnvil(name, root, -1, -1)
			if err != nil {
				t.Fatal(err)
			}
			loadedRoot, _ := loaded.anvilRoot.Object()
			values, _ := loadedRoot["block_entities"].Elements()
			if len(values) != 1 || !reflect.DeepEqual(values[0], expected) {
				t.Fatal("sculk NBT changed during Anvil restart")
			}
			loaded.SetBlock(15, 0, -60, StateStone)
			if err := save.persistChunk(loaded); err != nil {
				t.Fatal(err)
			}
			_, root, _, err = anvil.ReadChunk(filepath.Join(dir, "region"), -1, -1)
			if err != nil {
				t.Fatal(err)
			}
			saved, _ := root.Object()
			values, _ = saved["block_entities"].Elements()
			if len(values) != 0 {
				t.Fatal("replaced sculk entity remains in Anvil")
			}
		})
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("sculk entity JAR coverage %d", count)
	}
}
