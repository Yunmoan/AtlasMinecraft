package worldgen

import (
	"atlas/internal/nbt"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type villageRuntimeFixture struct {
	name   string
	owner  StructureChunkPosition
	pieces int
}

var villageRuntimeFixtures = []villageRuntimeFixture{
	{"plains", StructureChunkPosition{-24, 47}, 230},
	{"savanna", StructureChunkPosition{69, -23}, 117},
	{"taiga", StructureChunkPosition{-112, 93}, 113},
	{"snowy", StructureChunkPosition{-167, 36}, 128},
	{"desert", StructureChunkPosition{-403, 331}, 102},
}

func pinnedVillageRuntimeStart(t *testing.T, g *OverworldGenerator, fixture villageRuntimeFixture) StructureStart {
	t.Helper()
	eligible, err := g.structures.placement.IsStructureChunk("minecraft:villages", fixture.owner.X, fixture.owner.Z)
	if err != nil || !eligible {
		t.Fatal("pinned village placement rejected", err)
	}
	starts, err := g.structureStartsAt(fixture.owner.X, fixture.owner.Z)
	if err != nil {
		t.Fatal(err)
	}
	for _, start := range starts {
		if start.Structure == "minecraft:village_"+fixture.name {
			if len(start.Pieces) != fixture.pieces || start.TerrainAdjustment != "beard_thin" {
				t.Fatal("pinned village layout or beard terrain adjustment changed", len(start.Pieces), start.TerrainAdjustment)
			}
			return start
		}
	}
	t.Fatal("actual biome rejected pinned village", fixture)
	return StructureStart{}
}

// Select authored farm, loot and actor clips, plus a distant projected road.
// This never treats the presence of a start NBT as proof of placed blocks.
func villageRuntimeClipPositions(t *testing.T, start StructureStart) []StructureChunkPosition {
	t.Helper()
	positions := []StructureChunkPosition{start.Chunk}
	add := func(p StructureChunkPosition) {
		for _, old := range positions {
			if old == p {
				return
			}
		}
		positions = append(positions, p)
	}
	var road, farm, chest, actor, camel *StructureChunkPosition
	roadDistance := int32(-1)
	for _, piece := range start.Pieces {
		element, origin, rotation, _, err := jigsawSavedBlockSettings(&piece)
		if err != nil {
			t.Fatal(err)
		}
		if element.Location == "" {
			continue
		}
		if element.Projection == "terrain_matching" && strings.Contains(element.Location, "/streets/") {
			box := piece.Box
			p := StructureChunkPosition{(box.MinX + box.MaxX) / 2 >> 4, (box.MinZ + box.MaxZ) / 2 >> 4}
			distance := abs32(p.X-start.Chunk.X) + abs32(p.Z-start.Chunk.Z)
			if distance > roadDistance {
				road, roadDistance = &p, distance
			}
		}
		template, err := LoadJigsawTemplate(element.Location)
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range template.Palettes[0] {
			p, err := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
			if err != nil {
				t.Fatal(err)
			}
			p = treeOffset(origin, p.X, p.Y, p.Z)
			pos := StructureChunkPosition{p.X >> 4, p.Z >> 4}
			if villageCrop(block.State.Name) && farm == nil {
				farm = &pos
			}
			if table, _ := block.NBT.Text("LootTable"); strings.HasPrefix(table, "minecraft:chests/village/") && chest == nil {
				chest = &pos
			}
		}
		for _, entry := range template.Entities {
			precise := transformTemplateEntityPosition(entry.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
			p := StructureChunkPosition{int32(math.Floor(precise[0]+float64(origin.X))) >> 4, int32(math.Floor(precise[2]+float64(origin.Z))) >> 4}
			if actor == nil && p != start.Chunk {
				actor = &p
			}
			if id, _ := entry.NBT.Text("id"); id == "minecraft:camel" {
				camel = &p
			}
		}
	}
	if road == nil || *road == start.Chunk || farm == nil || chest == nil || actor == nil {
		t.Fatal("fixture lacks remote roads, farm, village loot or authored actor", start.Structure)
	}
	// This street's bbox center lies in an authored empty corner. Pin a
	// populated remote clip instead of asserting that every bbox is solid.
	if start.Structure == "minecraft:village_savanna" {
		p := StructureChunkPosition{68, -23}
		road = &p
	}
	for _, p := range []*StructureChunkPosition{road, farm, chest, actor, camel} {
		if p != nil {
			add(*p)
		}
	}
	return positions
}

func villageCrop(name string) bool {
	switch name {
	case "minecraft:wheat", "minecraft:carrots", "minecraft:potatoes", "minecraft:beetroots":
		return true
	default:
		return false
	}
}

func villageRoadColumns(t *testing.T, start StructureStart) map[[2]int32]bool {
	t.Helper()
	columns := make(map[[2]int32]bool)
	for _, piece := range start.Pieces {
		element, origin, rotation, _, err := jigsawSavedBlockSettings(&piece)
		if err != nil {
			t.Fatal(err)
		}
		if element.Projection != "terrain_matching" || !strings.Contains(element.Location, "/streets/") {
			continue
		}
		template, err := LoadJigsawTemplate(element.Location)
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range template.Palettes[0] {
			if block.State.Name == "minecraft:air" || block.State.Name == "minecraft:structure_void" {
				continue
			}
			p, err := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
			if err != nil {
				t.Fatal(err)
			}
			columns[[2]int32{origin.X + p.X, origin.Z + p.Z}] = true
		}
	}
	return columns
}

func TestOverworldVillageStructurePipeline(t *testing.T) {
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	g.SetTerrainWorkers(1)
	g.SetStructureGameTimeSource(func() int64 { return 2053 })
	for _, fixture := range villageRuntimeFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			start := pinnedVillageRuntimeStart(t, g, fixture)
			original, err := start.NBT()
			if err != nil {
				t.Fatal(err)
			}
			positions := villageRuntimeClipPositions(t, start)
			t.Logf("fixture %s clips=%v", fixture.name, positions)
			roadColumns := villageRoadColumns(t, start)
			chunks := make([]*GeneratedOverworldChunk, len(positions))
			errors := make([]error, len(positions))
			var wg sync.WaitGroup
			for i, pos := range positions {
				wg.Add(1)
				go func(i int, pos StructureChunkPosition) {
					defer wg.Done()
					chunks[i], errors[i] = g.Generate(pos.X, pos.Z)
				}(i, pos)
			}
			wg.Wait()
			writes, roads, crops, loot, actors, camels := 0, 0, 0, 0, 0, 0
			flags := make(map[string]bool)
			for i, chunk := range chunks {
				if errors[i] != nil {
					t.Fatal(errors[i])
				}
				pos := positions[i]
				refs, _ := chunk.Structures.Object("References")
				packed, _ := refs[start.Structure].Longs()
				owner := int64(uint64(uint32(start.Chunk.X)) | uint64(uint32(start.Chunk.Z))<<32)
				found := false
				for _, value := range packed {
					found = found || value == owner
				}
				if !found {
					t.Fatal("village reference missing", pos)
				}
				if i == 0 {
					starts, _ := chunk.Structures.Object("starts")
					saved, ok := starts.Object(start.Structure)
					if !ok || !reflect.DeepEqual(saved, original) {
						t.Fatal("village lost full owned start NBT")
					}
				}
				cached, err := g.jigsawStructureBlocksFor(start, pos.X, pos.Z)
				if err != nil {
					t.Fatal(err)
				}
				matched := 0
				for p, want := range cached.blocks {
					if p.X>>4 != pos.X || p.Z>>4 != pos.Z {
						continue
					}
					got, err := chunk.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, want) || want.Name == "minecraft:air" {
						continue
					}
					matched++
					if i == 1 && roadColumns[[2]int32{p.X, p.Z}] {
						roads++
					}
					if villageCrop(got.Name) {
						crops++
					}
				}
				if matched < 20 {
					t.Fatal("village start metadata without retained authored blocks", pos, matched)
				}
				writes += matched
				for p, expected := range cached.blockEntities {
					table, _ := expected.data.Text("LootTable")
					if p.X>>4 != pos.X || p.Z>>4 != pos.Z || !strings.HasPrefix(table, "minecraft:chests/village/") {
						continue
					}
					got, err := chunk.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
					if err != nil {
						t.Fatal(err)
					}
					if got.Name != expected.block {
						continue
					}
					want := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: expected.data}).Value.(nbt.Object)
					want["x"], want["y"], want["z"] = nbt.Tag{Type: nbt.Int, Value: p.X}, nbt.Tag{Type: nbt.Int, Value: p.Y}, nbt.Tag{Type: nbt.Int, Value: p.Z}
					retained := false
					for _, actual := range chunk.BlockEntities {
						retained = retained || reflect.DeepEqual(actual, want)
					}
					if !retained || want["LootTableSeed"].Type != nbt.Long {
						t.Fatal("village loot lost full NBT or Long seed", p)
					}
					loot++
				}
				for _, actor := range cached.pending {
					precise := pendingEntityBlockPosition(actor)
					if precise.X>>4 != pos.X || precise.Z>>4 != pos.Z || actor.SavedFlag == "" || flags[actor.SavedFlag] {
						t.Fatal("village actor precise clip or identity invalid", actor.ID, pos, precise)
					}
					flags[actor.SavedFlag] = true
					matched := false
					for _, actual := range chunk.PendingEntities {
						matched = matched || reflect.DeepEqual(actual, actor)
					}
					if !matched {
						t.Fatal("village authored actor missing from final normal chunk")
					}
					actors++
					if actor.ID == "minecraft:camel" {
						camels++
						if pose, _ := actor.Data.Number("LastPoseTick"); pose != 2000 {
							t.Fatal("Camel pose lost world age", pose)
						}
						if fire, _ := actor.Data.Number("Fire"); fire != 0 || actor.Data["Fire"].Type != nbt.Short {
							t.Fatal("Camel lost authored short Fire0")
						}
					}
				}
				before := generatedStructureDigest(t, chunk)
				chunk.Terrain.Blocks[0] = NoiseLava
				chunk.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
				for _, actor := range chunk.PendingEntities {
					actor.Data["id"] = nbt.Tag{Type: nbt.String, Value: "mutated"}
				}
				again, err := g.Generate(pos.X, pos.Z)
				if err != nil || generatedStructureDigest(t, again) != before {
					t.Fatal("village cache changed through returned mutation", err)
				}
			}
			if roads == 0 || crops == 0 || loot == 0 || actors == 0 || (fixture.name == "desert" && camels != 1) {
				t.Fatal("village actual road/crop/loot/actor coverage missing", roads, crops, loot, actors, camels)
			}
			t.Logf("seed1234 %s owner=%v pieces=%d clips=%v; retained blocks=%d roads=%d crops=%d seeded loot=%d actors=%d camels=%d", fixture.name, start.Chunk, len(start.Pieces), positions, writes, roads, crops, loot, actors, camels)
		})
	}
}

// Opt-in fixture discovery reads only seed-derived starts, never saved worlds.
func TestVillageRuntimeFixtureDiscovery(t *testing.T) {
	if os.Getenv("ATLAS_VILLAGE_DISCOVER") == "" {
		t.Skip("set ATLAS_VILLAGE_DISCOVER to find stable runtime village fixtures")
	}
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	biomes, err := newOverworldBiomeSampler(g.seed, g.noise.settings)
	if err != nil {
		t.Fatal(err)
	}
	placement := g.structures.placement.Sets["minecraft:villages"].Placement
	found := make(map[string]bool)
	for radius := int32(0); radius <= 20 && len(found) < 5; radius++ {
		for rx := -radius; rx <= radius && len(found) < 5; rx++ {
			for rz := -radius; rz <= radius && len(found) < 5; rz++ {
				if max(abs32(rx), abs32(rz)) != radius {
					continue
				}
				position, err := placement.PotentialChunk(g.seed, rx*placement.Spacing, rz*placement.Spacing)
				if err != nil {
					t.Fatal(err)
				}
				biome, err := biomes.BiomeAtQuart((position.X*16+8)>>2, 16, (position.Z*16+8)>>2)
				if err != nil {
					t.Fatal(err)
				}
				name := map[string]string{"plains": "plains", "sunflower_plains": "plains", "meadow": "plains", "desert": "desert", "savanna": "savanna", "snowy_plains": "snowy", "taiga": "taiga"}[strings.TrimPrefix(biome, "minecraft:")]
				if name == "" || found[name] {
					continue
				}
				starts, err := g.structureStartsAt(position.X, position.Z)
				if err != nil {
					t.Fatal(err)
				}
				for _, start := range starts {
					if start.Structure != "minecraft:village_"+name {
						continue
					}
					found[name] = true
					t.Logf("FIXTURE %s owner=%d,%d pieces=%d", start.Structure, position.X, position.Z, len(start.Pieces))
					for _, piece := range start.Pieces {
						element, origin, rotation, _, err := jigsawSavedBlockSettings(&piece)
						if err != nil {
							t.Fatal(err)
						}
						if element.Location == "" {
							continue
						}
						if element.Projection == "terrain_matching" || strings.Contains(element.Location, "farm") {
							box := piece.Box
							t.Logf("PIECE %s projection=%s chunk=%d,%d origin=%v", element.Location, element.Projection, (box.MinX+box.MaxX)/2>>4, (box.MinZ+box.MaxZ)/2>>4, origin)
						}
						template, err := LoadJigsawTemplate(element.Location)
						if err != nil {
							t.Fatal(err)
						}
						for _, actor := range template.Entities {
							id, _ := actor.NBT.Text("id")
							pos, _ := TransformTemplatePosition(actor.BlockPosition, FeaturePosition{}, TemplateMirrorNone, rotation)
							pos = treeOffset(origin, pos.X, pos.Y, pos.Z)
							precise := transformTemplateEntityPosition(actor.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
							for axis, value := range []int32{origin.X, origin.Y, origin.Z} {
								precise[axis] += float64(value)
							}
							t.Logf("ACTOR %s template=%s block=%v precise=%v blockchunk=%d,%d precisechunk=%d,%d", id, element.Location, pos, precise, pos.X>>4, pos.Z>>4, int32(math.Floor(precise[0]))>>4, int32(math.Floor(precise[2]))>>4)
						}
					}
				}
			}
		}
	}
	if len(found) != 5 {
		t.Fatalf("incomplete runtime village fixtures: %v", found)
	}
}
