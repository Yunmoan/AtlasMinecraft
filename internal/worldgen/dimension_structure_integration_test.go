package worldgen

import (
	"atlas/internal/nbt"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestDimensionNetherComplexWeightedRejection(t *testing.T) {
	g, err := NewDimensionGeneratorWithOptions(1234, "nether", DimensionGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	biome, err := NewDimensionBiomeSampler(g.seed, "nether")
	if err != nil {
		t.Fatal(err)
	}
	set := g.structures.placement.Sets["minecraft:nether_complexes"]
	var acceptedBastion, basaltFallback bool
	selectedStop := errors.New("capture first weighted candidate")
	for radius := int32(0); radius < 8 && (!acceptedBastion || !basaltFallback); radius++ {
		for x := -radius; x <= radius; x++ {
			for z := -radius; z <= radius; z++ {
				if max(abs32(x), abs32(z)) != radius {
					continue
				}
				pos, err := set.Placement.PotentialChunk(g.seed, x*set.Placement.Spacing, z*set.Placement.Spacing)
				if err != nil {
					t.Fatal(err)
				}
				first := ""
				_, err = SelectStructureSetStart(g.seed, pos.X, pos.Z, set, func(name string) (*StructureStart, error) { first = name; return nil, selectedStop })
				if !errors.Is(err, selectedStop) {
					t.Fatal("selection did not stop on accepted unsupported candidate")
				}
				if first != "minecraft:bastion_remnant" {
					continue
				}
				stub, err := GenerateBastionGenerationStub(g.seed, pos.X, pos.Z, biome.BiomeAtQuart)
				if err != nil {
					t.Fatal(err)
				}
				starts, err := g.structureStartsAt(pos.X, pos.Z)
				if err != nil {
					t.Fatal(err)
				}
				fortress, bastion := false, false
				for _, start := range starts {
					if start.Structure == "minecraft:fortress" {
						fortress = true
					}
					if start.Structure == "minecraft:bastion_remnant" {
						bastion = len(start.Pieces) > 0
					}
				}
				if stub != nil {
					if fortress || !bastion {
						t.Fatal("accepted Bastion missing or redirected to Fortress")
					}
					acceptedBastion = true
				} else {
					if !fortress || bastion {
						t.Fatal("biome rejected Bastion did not retry Fortress")
					}
					basaltFallback = true
				}
			}
		}
	}
	if !acceptedBastion || !basaltFallback {
		t.Fatalf("missing weighted runtime cases: accepted=%v basalt-rejected=%v", acceptedBastion, basaltFallback)
	}
}

func acceptedDimensionIntegrationStart(t *testing.T, g *DimensionGenerator, set, name string) StructureStart {
	t.Helper()
	placement := g.structures.placement.Sets[set].Placement
	for radius := int32(0); radius < 16; radius++ {
		for rx := -radius; rx <= radius; rx++ {
			for rz := -radius; rz <= radius; rz++ {
				if max(abs32(rx), abs32(rz)) != radius {
					continue
				}
				pos, err := placement.PotentialChunk(g.seed, rx*placement.Spacing, rz*placement.Spacing)
				if err != nil {
					t.Fatal(err)
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
	t.Fatalf("no accepted %s start", name)
	return StructureStart{}
}

func TestDimensionSupportedStructurePipeline(t *testing.T) {
	g, err := NewDimensionGeneratorWithOptions(1234, "nether", DimensionGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range []struct{ set, name string }{{"minecraft:nether_complexes", "minecraft:fortress"}, {"minecraft:nether_fossils", "minecraft:nether_fossil"}, {"minecraft:ruined_portals", "minecraft:ruined_portal_nether"}} {
		t.Run(family.name, func(t *testing.T) {
			start := acceptedDimensionIntegrationStart(t, g, family.set, family.name)
			t.Logf("accepted %d,%d", start.Chunk.X, start.Chunk.Z)
			positions := [][2]int32{{start.Chunk.X, start.Chunk.Z}}
			var expectedChest, expectedSpawner *FeaturePosition
			if family.name == "minecraft:fortress" {
				for _, piece := range start.Pieces {
					if piece.ID == "minecraft:nemt" && expectedSpawner == nil {
						p := piece.BlockPosition(3, 5, 5)
						expectedSpawner = &p
						positions = append(positions, [2]int32{p.X >> 4, p.Z >> 4})
					}
					flag, _ := piece.Data["Chest"].Number()
					if flag != 0 && expectedChest == nil {
						x := int32(1)
						if piece.ID == "minecraft:nesclt" {
							x = 3
						}
						p := piece.BlockPosition(x, 2, 3)
						expectedChest = &p
						positions = append(positions, [2]int32{p.X >> 4, p.Z >> 4})
					}
				}
			} else {
				box, _ := start.BoundingBox()
				clipX, clipZ := box.MaxX>>4, box.MaxZ>>4
				if clipX != start.Chunk.X || clipZ != start.Chunk.Z {
					positions = append(positions, [2]int32{clipX, clipZ})
				}
			}
			var wg sync.WaitGroup
			chunks := make([]*GeneratedDimensionChunk, len(positions))
			errs := make([]error, len(positions))
			for i := range positions {
				wg.Add(1)
				go func(i int) { defer wg.Done(); chunks[i], errs[i] = g.Generate(positions[i][0], positions[i][1]) }(i)
			}
			wg.Wait()
			for i, chunk := range chunks {
				if errs[i] != nil {
					t.Fatal(errs[i])
				}
				refs, _ := chunk.Structures.Object("References")
				if refs[family.name].Type != nbt.LongArray {
					t.Fatalf("%s missing references at %v", family.name, positions[i])
				}
				if i == 0 {
					owned, _ := chunk.Structures.Object("starts")
					saved, found := owned.Object(family.name)
					if !found {
						t.Fatal("owned start missing")
					}
					children, _ := saved["Children"].Elements()
					if len(children) != len(start.Pieces) {
						t.Fatal("saved piece list incomplete")
					}
				}
				again, err := g.Generate(positions[i][0], positions[i][1])
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(chunk, again) {
					t.Fatal("parallel/neighbor replay changed result")
				}
				chunk.Terrain.Blocks[0] = NoiseLava
				chunk.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
				detached, err := g.Generate(positions[i][0], positions[i][1])
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(again, detached) {
					t.Fatal("caller mutation changed cached structure")
				}
				chunks[i] = again
			}
			for _, expectation := range []struct {
				position *FeaturePosition
				name     string
			}{{expectedChest, "minecraft:chest"}, {expectedSpawner, "minecraft:spawner"}} {
				if expectation.position == nil {
					continue
				}
				p := *expectation.position
				found := false
				for i, chunk := range chunks {
					if positions[i][0] != p.X>>4 || positions[i][1] != p.Z>>4 {
						continue
					}
					state, err := chunk.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
					if err != nil || state.Name != expectation.name {
						t.Fatalf("runtime block at %v: %v %v", p, state, err)
					}
					for _, entity := range chunk.BlockEntities {
						x, _ := entity.Number("x")
						y, _ := entity.Number("y")
						z, _ := entity.Number("z")
						if x == int64(p.X) && y == int64(p.Y) && z == int64(p.Z) {
							found = true
							if expectation.name == "minecraft:chest" {
								loot, _ := entity.Text("LootTable")
								if loot != "minecraft:chests/nether_bridge" {
									t.Fatal("fortress loot table missing")
								}
							}
						}
					}
				}
				if !found {
					t.Fatalf("runtime %s block entity missing at %v", expectation.name, p)
				}
			}
			if family.name == "minecraft:nether_fossil" {
				piece := start.Pieces[0]
				name, _ := piece.Data.Text("Template")
				template, err := LoadNetherFossilTemplate(name)
				if err != nil {
					t.Fatal(err)
				}
				want := 0
				for _, block := range template.Palettes[0] {
					if block.State.Name == "minecraft:bone_block" {
						want++
					}
				}
				got := 0
				for _, chunk := range chunks {
					for _, state := range chunk.Terrain.SurfaceBlocks {
						if state.Name == "minecraft:bone_block" {
							got++
						}
					}
				}
				if got == 0 || want == 0 {
					t.Fatal("Nether fossil metadata without bone blocks")
				}
			}
		})
	}
}

func TestDimensionStructuresOptionsAndStarts(t *testing.T) {
	disabled, err := NewDimensionGenerator(1234, "nether")
	if err != nil {
		t.Fatal(err)
	}
	if disabled.StructuresEnabled() {
		t.Fatal("convenience generator unexpectedly enables structures")
	}
	g, err := NewDimensionGeneratorWithOptions(1234, "nether", DimensionGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	for cx := int32(-2); cx <= 2; cx++ {
		for cz := int32(-2); cz <= 2; cz++ {
			starts, err := g.structureStartsAt(cx, cz)
			if err != nil {
				t.Fatal(err)
			}
			for _, start := range starts {
				if len(start.Pieces) == 0 {
					t.Fatal("saved structure without actual pieces")
				}
			}
		}
	}
	fossil := StructureStart{Structure: "minecraft:nether_fossil", TerrainAdjustment: "beard_thin", Pieces: []StructurePiece{{Box: StructureBoundingBox{0, 64, 0, 9, 71, 9}}}}
	with, err := g.noise.GenerateNoiseWithStructures(0, 0, structureTerrainPieces([]StructureStart{fossil}))
	if err != nil {
		t.Fatal(err)
	}
	without, err := g.noise.GenerateNoise(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(with.Blocks, without.Blocks) {
		t.Fatal("Nether fossil beard_thin never reached actual terrain density")
	}
}

func BenchmarkDimensionSupportedStructures(b *testing.B) {
	for _, entry := range []struct {
		name string
		x, z int32
	}{{"fortress", 14, 8}, {"fossil", 4, -4}, {"ruined_portal", 8, 20}} {
		b.Run(entry.name, func(b *testing.B) {
			warm, err := NewDimensionGeneratorWithOptions(1234, "nether", DimensionGeneratorOptions{GenerateStructures: true})
			if err != nil {
				b.Fatal(err)
			}
			if _, err := warm.Generate(entry.x, entry.z); err != nil {
				b.Fatal(err)
			}
			b.Run("cached", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					if _, err := warm.Generate(entry.x, entry.z); err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("cold_world_cache", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					g, err := NewDimensionGeneratorWithOptions(1234, "nether", DimensionGeneratorOptions{GenerateStructures: true})
					if err != nil {
						b.Fatal(err)
					}
					if _, err := g.Generate(entry.x, entry.z); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
