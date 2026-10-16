package worldgen

import (
	"atlas/internal/nbt"
	"sync"
	"testing"
)

func TestOverworldJigsawStructurePipeline(t *testing.T) {
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct{ set, name string }{
		{"minecraft:trail_ruins", "minecraft:trail_ruins"},
		{"minecraft:trial_chambers", "minecraft:trial_chambers"},
		{"minecraft:ancient_cities", "minecraft:ancient_city"},
	} {
		t.Run(entry.name, func(t *testing.T) {
			start := supportedIntegrationStart(t, g, entry.set, entry.name)
			t.Logf("accepted%d,%d with%d pieces", start.Chunk.X, start.Chunk.Z, len(start.Pieces))
			original, err := start.NBT()
			if err != nil {
				t.Fatal(err)
			}
			originalDigest := iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: original})
			// Exercise ownership and a remote template clip without eagerly
			// generating every chunk of a large city or chamber assembly.
			positions := []StructureChunkPosition{start.Chunk}
			for index := len(start.Pieces) - 1; index >= 0; index-- {
				box := start.Pieces[index].Box
				position := StructureChunkPosition{(box.MinX + box.MaxX) / 2 >> 4, (box.MinZ + box.MaxZ) / 2 >> 4}
				if position != start.Chunk {
					positions = append(positions, position)
					break
				}
			}
			results := make([]*GeneratedOverworldChunk, len(positions))
			errors := make([]error, len(positions))
			var wg sync.WaitGroup
			for index, position := range positions {
				wg.Add(1)
				go func(index int, position StructureChunkPosition) {
					defer wg.Done()
					results[index], errors[index] = g.Generate(position.X, position.Z)
				}(index, position)
			}
			wg.Wait()
			for index, chunk := range results {
				if errors[index] != nil {
					t.Fatal(errors[index])
				}
				position := positions[index]
				refs, _ := chunk.Structures.Object("References")
				values, _ := refs[entry.name].Longs()
				found := false
				for _, value := range values {
					found = found || value == int64(uint64(uint32(start.Chunk.X))|uint64(uint32(start.Chunk.Z))<<32)
				}
				if !found {
					t.Fatalf("missing%s reference in%d,%d", entry.name, position.X, position.Z)
				}
				if index == 0 {
					starts, _ := chunk.Structures.Object("starts")
					owned, ok := starts.Object(entry.name)
					if !ok || iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: owned}) != originalDigest {
						t.Fatal("Jigsaw owned start lost original full saved NBT")
					}
				}
				cached, err := g.jigsawStructureBlocksFor(start, position.X, position.Z)
				if err != nil {
					t.Fatal(err)
				}
				if len(cached.blocks) == 0 {
					t.Fatal("Jigsaw clip placed no blocks")
				}
				// Retained block-entity NBT must match the actual chunk, allowing
				// original later features to replace an archaeology/container block.
				for p, entity := range cached.blockEntities {
					if p.X>>4 != position.X || p.Z>>4 != position.Z {
						continue
					}
					state, err := chunk.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
					if err != nil {
						t.Fatal(err)
					}
					if state.Name != entity.block {
						continue
					}
					want := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
					want["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
					want["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
					want["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
					matched := false
					for _, actual := range chunk.BlockEntities {
						matched = matched || iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: actual}) == iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: want})
					}
					if !matched {
						t.Fatalf("Jigsaw block entity lost NBT at%v", p)
					}
				}
				digest := generatedStructureDigest(t, chunk)
				chunk.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
				chunk.Terrain.Blocks[0] = NoiseLava
				again, err := g.Generate(position.X, position.Z)
				if err != nil {
					t.Fatal(err)
				}
				if generatedStructureDigest(t, again) != digest {
					t.Fatal("Jigsaw parallel replay or returned mutation changed cached chunk")
				}
				t.Logf("clip%d,%d: %d structure writes,%d block entities", position.X, position.Z, len(cached.blocks), len(cached.blockEntities))
			}
		})
	}
}
