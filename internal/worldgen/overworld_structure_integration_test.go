package worldgen

import (
	"atlas/internal/nbt"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func supportedIntegrationStart(t *testing.T, g *OverworldGenerator, set, name string) StructureStart {
	t.Helper()
	// Pin accepted runtime locations so this integration test measures chunk
	// replay and persistence rather than repeatedly searching remote biomes.
	if g.seed == 1234 {
		locations := map[string]StructureChunkPosition{
			"minecraft:stronghold":     {-100, -125},
			"minecraft:desert_pyramid": {-154, 402}, "minecraft:swamp_hut": {-31, 119},
			"minecraft:jungle_pyramid": {131, -54}, "minecraft:igloo": {-79, 35},
			"minecraft:buried_treasure": {-23, -10}, "minecraft:ruined_portal": {-16, 51},
			"minecraft:ruined_portal_desert": {-140, 423}, "minecraft:ruined_portal_jungle": {136, -71},
			"minecraft:ruined_portal_mountain": {44, 59}, "minecraft:ruined_portal_ocean": {8, 20},
			"minecraft:ruined_portal_swamp": {-33, 171},
			"minecraft:shipwreck":           {4, 12}, "minecraft:shipwreck_beached": {72, -62},
			"minecraft:ocean_ruin_cold": {-30, -56}, "minecraft:ocean_ruin_warm": {62, -60},
			"minecraft:trail_ruins": {44, 38}, "minecraft:trial_chambers": {13, 21}, "minecraft:ancient_city": {13, 60},
		}
		if position, found := locations[name]; found {
			var eligible bool
			var err error
			if name == "minecraft:stronghold" {
				eligible, err = g.strongholdPlacementAt(position.X, position.Z)
			} else {
				eligible, err = g.structures.placement.IsStructureChunk(set, position.X, position.Z)
			}
			if err != nil || !eligible {
				t.Fatalf("fixture start is not eligible: %v", err)
			}
			starts, err := g.structureStartsAt(position.X, position.Z)
			if err != nil {
				t.Fatal(err)
			}
			for _, start := range starts {
				if start.Structure == name {
					return start
				}
			}
			t.Fatalf("pinned runtime %s start rejected at %d,%d", name, position.X, position.Z)
		}
	}
	placement := g.structures.placement.Sets[set].Placement
	limit := int32(16)
	if placement.Spacing == 1 {
		limit = 128
	}
	for radius := int32(0); radius < limit; radius++ {
		for rx := -radius; rx <= radius; rx++ {
			for rz := -radius; rz <= radius; rz++ {
				if max(abs32(rx), abs32(rz)) != radius {
					continue
				}
				pos, err := placement.PotentialChunk(g.seed, rx*placement.Spacing, rz*placement.Spacing)
				if err != nil {
					t.Fatal(err)
				}
				eligible, err := g.structures.placement.IsStructureChunk(set, pos.X, pos.Z)
				if err != nil {
					t.Fatal(err)
				}
				if !eligible {
					continue
				}
				starts, err := g.structureStartsAt(pos.X, pos.Z)
				if err != nil {
					t.Fatal(err)
				}
				for _, start := range starts {
					if start.Structure == name {
						return start
					}
				}
			}
		}
	}
	t.Fatalf("no accepted %s start for seed %d", name, g.seed)
	return StructureStart{}
}

func TestOverworldSupportedStructurePipeline(t *testing.T) {
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range supportedOverworldStructures {
		if isOverworldJigsawStructure(entry.name) || isOverworldMineshaftStructure(entry.name) || entry.name == "minecraft:stronghold" {
			continue // Large starts have focused multi-clip integration below.
		}
		t.Run(entry.name, func(t *testing.T) {
			start := supportedIntegrationStart(t, g, entry.set, entry.name)
			cached, err := g.structureBlocksFor(start)
			if err != nil {
				t.Fatal(err)
			}
			cx, cz := start.Chunk.X, start.Chunk.Z
			t.Logf("accepted start %d,%d", cx, cz)
			positions := [][2]int32{{cx, cz}}
			box, _ := start.BoundingBox()
			writeBox, _ := structureWriteBox(start)
			for x := writeBox.MinX >> 4; x <= writeBox.MaxX>>4; x++ {
				for z := writeBox.MinZ >> 4; z <= writeBox.MaxZ>>4; z++ {
					if x != cx || z != cz {
						positions = append(positions, [2]int32{x, z})
					}
				}
			}
			chunks := make([]*GeneratedOverworldChunk, len(positions))
			errors := make([]error, len(positions))
			var wg sync.WaitGroup
			for i := range chunks {
				wg.Add(1)
				go func(i int) { defer wg.Done(); chunks[i], errors[i] = g.Generate(positions[i][0], positions[i][1]) }(i)
			}
			wg.Wait()
			var chestCount, pendingCount, archaeologyCount int
			var targetPending []PendingStructureEntity
			for i, ch := range chunks {
				if errors[i] != nil {
					t.Fatal(errors[i])
				}
				for _, entity := range ch.BlockEntities {
					x, _ := entity.Number("x")
					y, _ := entity.Number("y")
					z, _ := entity.Number("z")
					expected, belongs := cached.blockEntities[FeaturePosition{int32(x), int32(y), int32(z)}]
					if !belongs {
						continue // Other enabled structures can share these chunks.
					}
					wantTable, _ := expected.data.Text("LootTable")
					if table, _ := entity.Text("LootTable"); table != wantTable {
						continue
					}
					if id, _ := entity.Text("id"); id == "minecraft:chest" {
						if table, _ := entity.Text("LootTable"); table == "minecraft:chests/desert_pyramid" || table == "minecraft:chests/igloo_chest" || table == "minecraft:chests/buried_treasure" || table == "minecraft:chests/ruined_portal" || strings.HasPrefix(table, "minecraft:chests/shipwreck_") || strings.HasPrefix(table, "minecraft:chests/underwater_ruin_") {
							chestCount++
						}
					}
					if id, _ := entity.Text("id"); id == "minecraft:brushable_block" {
						if table, _ := entity.Text("LootTable"); strings.HasPrefix(table, "minecraft:archaeology/ocean_ruin_") {
							archaeologyCount++
						}
					}
				}
				for _, actual := range ch.PendingEntities {
					for _, expected := range cached.pending {
						if reflect.DeepEqual(actual, expected) {
							pendingCount++
							targetPending = append(targetPending, actual)
							break
						}
					}
				}
				refs, _ := ch.Structures.Object("References")
				if i == 0 {
					starts, _ := ch.Structures.Object("starts")
					if _, ok := starts.Object(entry.name); !ok {
						t.Fatal("owned start missing from canonical Anvil NBT")
					}
				}
				px, pz := positions[i][0], positions[i][1]
				if box.MaxX >= px*16 && box.MinX <= px*16+15 && box.MaxZ >= pz*16 && box.MinZ <= pz*16+15 {
					if refs[entry.name].Type != nbt.LongArray {
						t.Fatal("start reference missing")
					}
				}
				before := generatedStructureDigest(t, ch)
				ch.Terrain.Blocks[0] = NoiseLava
				ch.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
				if len(ch.PendingEntities) > 0 {
					ch.PendingEntities[0].Data["id"] = nbt.Tag{Type: nbt.String, Value: "mutated"}
				}
				if len(ch.Terrain.PostProcessing) > 0 {
					ch.Terrain.PostProcessing[0] = !ch.Terrain.PostProcessing[0]
				}
				second, err := g.Generate(positions[i][0], positions[i][1])
				if err != nil {
					t.Fatal(err)
				}
				if before != generatedStructureDigest(t, second) {
					t.Fatal("parallel replay or returned mutation changed saved structure")
				}
				chunks[i] = second
			}
			if entry.name == "minecraft:desert_pyramid" && chestCount != 4 {
				t.Fatalf("pyramid loot chests=%d, want4", chestCount)
			}
			if entry.name == "minecraft:swamp_hut" && pendingCount != 2 {
				t.Fatalf("hut entity requests=%d, want2", pendingCount)
			}
			if entry.name == "minecraft:buried_treasure" || strings.HasPrefix(entry.name, "minecraft:ruined_portal") {
				if chestCount != 1 {
					t.Fatalf("loot chests=%d, want1", chestCount)
				}
			}
			if strings.HasPrefix(entry.name, "minecraft:shipwreck") {
				name, _ := start.Pieces[0].Data.Text("Template")
				template, err := LoadShipwreckTemplate(name)
				if err != nil {
					t.Fatal(err)
				}
				expected := 0
				for _, block := range template.Palettes[0] {
					if block.State.Name == "minecraft:chest" {
						expected++
					}
				}
				if chestCount != expected || pendingCount != 0 {
					t.Fatalf("shipwreck %s loot chests=%d, want %d; unexpected actors=%d", name, chestCount, expected, pendingCount)
				}
			}
			if entry.name == "minecraft:igloo" {
				hasBasement := false
				for _, piece := range start.Pieces {
					name, _ := piece.Data.Text("Template")
					hasBasement = hasBasement || name == "minecraft:igloo/bottom"
				}
				if hasBasement && (chestCount != 1 || pendingCount != 2) {
					t.Fatalf("igloo basement chests=%d entities=%d", chestCount, pendingCount)
				}
				if !hasBasement && (chestCount != 0 || pendingCount != 0) {
					t.Fatal("igloo without basement gained underground loot/entities")
				}
			}
			replayed := make(map[featureChunkPosition]*GeneratedOverworldChunk)
			for i, position := range positions {
				replayed[featureChunkPosition{position[0], position[1]}] = chunks[i]
			}
			for p := range cached.postProcessing {
				if p.X < writeBox.MinX || p.X > writeBox.MaxX || p.Z < writeBox.MinZ || p.Z > writeBox.MaxZ {
					t.Fatalf("postprocessing outside replay footprint: %+v", p)
				}
				ch := replayed[featureChunkPosition{p.X >> 4, p.Z >> 4}]
				if ch == nil {
					t.Fatalf("postprocessing chunk was not replayed: %+v", p)
				}
				index, err := ch.Terrain.Index(p.X&15, p.Y, p.Z&15)
				if err != nil || index >= len(ch.Terrain.PostProcessing) || !ch.Terrain.PostProcessing[index] {
					t.Fatalf("cached fluid/shape update missing from runtime chunk: %+v", p)
				}
			}
			if strings.HasPrefix(entry.name, "minecraft:ocean_ruin_") {
				expectedChests, expectedArchaeology := 0, 0
				overwritten := 0
				for p, entity := range cached.blockEntities {
					ch := replayed[featureChunkPosition{p.X >> 4, p.Z >> 4}]
					if ch == nil {
						t.Fatalf("block entity chunk was not replayed: %+v", p)
					}
					state, err := ch.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
					if err != nil {
						t.Fatal(err)
					}
					if state.Name != entity.block {
						// Later decoration, including underwater magma, can
						// replace archaeology blocks in the original step order.
						overwritten++
						continue
					}
					expected := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
					expected["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
					expected["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
					expected["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
					found := false
					for _, actual := range ch.BlockEntities {
						if iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: expected}) == iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: actual}) {
							found = true
							break
						}
					}
					if !found {
						t.Fatalf("surviving ocean ruin block entity lost NBT: %+v %v", p, expected)
					}
					id, _ := entity.data.Text("id")
					table, _ := entity.data.Text("LootTable")
					if id == "minecraft:chest" && strings.HasPrefix(table, "minecraft:chests/underwater_ruin_") {
						expectedChests++
					}
					if id == "minecraft:brushable_block" && strings.HasPrefix(table, "minecraft:archaeology/ocean_ruin_") {
						expectedArchaeology++
					}
				}
				if expectedChests == 0 || expectedArchaeology == 0 || chestCount != expectedChests || archaeologyCount != expectedArchaeology {
					t.Fatalf("ocean ruin loot/archaeology %d/%d, want %d/%d", chestCount, archaeologyCount, expectedChests, expectedArchaeology)
				}
				if pendingCount != len(cached.pending) {
					t.Fatalf("persistent drowned requests=%d, want %d", pendingCount, len(cached.pending))
				}
				for _, entity := range targetPending {
					persistence, _ := entity.Data.Number("PersistenceRequired")
					if entity.ID != "minecraft:drowned" || !entity.Persistent || persistence != 1 {
						t.Fatalf("ocean ruin lost persistent drowned request: %+v", entity)
					}
				}
				t.Logf("replayed %d pieces, %d loot chests, %d archaeology records (%d replaced by later decoration) and %d persistent drowned requests", len(cached.start.Pieces), chestCount, archaeologyCount, overwritten, pendingCount)
			}
			hpos, _ := cached.start.Pieces[0].Data.Number("HPos")
			if hpos < 0 {
				t.Fatal("postprocessed HPos not retained")
			}
			if entry.name == "minecraft:swamp_hut" {
				for _, flag := range []string{"Witch", "Cat"} {
					value, _ := cached.start.Pieces[0].Data.Number(flag)
					if value != 0 {
						t.Fatalf("unfulfilled entity flag %s marked complete", flag)
					}
				}
			}
			if strings.HasPrefix(entry.name, "minecraft:ruined_portal") {
				outside := 0
				for p := range cached.blocks {
					if p.X < box.MinX || p.X > box.MaxX || p.Z < box.MinZ || p.Z > box.MaxZ {
						outside++
					}
					if !structureContains(writeBox, p) {
						t.Fatalf("portal write outside replay footprint: %+v", p)
					}
				}
				t.Logf("replayed %d portal writes, including %d outside template bounds", len(cached.blocks), outside)
				if outside == 0 {
					t.Fatal("portal surrounding spread was omitted")
				}
			}
		})
	}
}

func generatedStructureDigest(t *testing.T, ch *GeneratedOverworldChunk) [32]byte {
	t.Helper()
	h := sha256.New()
	for y := ch.Terrain.MinY; y < ch.Terrain.MinY+ch.Terrain.Height; y++ {
		for z := int32(0); z < 16; z++ {
			for x := int32(0); x < 16; x++ {
				state, err := ch.Terrain.BlockStateAt(x, y, z)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Fprintln(h, state)
			}
		}
	}
	fmt.Fprintln(h, ch.Structures)
	for _, entity := range ch.BlockEntities {
		fmt.Fprintln(h, entity)
	}
	for index, marked := range ch.Terrain.PostProcessing {
		if marked {
			fmt.Fprintln(h, "postprocessing", index)
		}
	}
	for _, entity := range ch.PendingEntities {
		fmt.Fprintln(h, entity)
	}
	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return sum
}
