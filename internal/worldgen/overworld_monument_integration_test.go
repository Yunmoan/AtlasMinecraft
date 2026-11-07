package worldgen

import (
	"atlas/internal/nbt"
	"reflect"
	"sort"
	"testing"
)

func TestOverworldMonumentPipeline(t *testing.T) {
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	start := supportedIntegrationStart(t, g, "minecraft:ocean_monuments", "minecraft:monument")
	t.Logf("accepted monument at chunk %d,%d", start.Chunk.X, start.Chunk.Z)
	cached, err := g.structureBlocksFor(start)
	if err != nil {
		t.Fatal(err)
	}
	if len(cached.pending) != 3 {
		t.Fatalf("got %d guardians, want 3", len(cached.pending))
	}
	for _, actor := range cached.pending {
		if actor.ID != "minecraft:elder_guardian" || actor.SavedFlag == "" {
			t.Fatal("invalid guardian request")
		}
		health, _ := actor.Data["Health"].Value.(float32)
		persistent, _ := actor.Data.Number("PersistenceRequired")
		if health != 80 || persistent != 1 {
			t.Fatalf("guardian defaults: health=%v persistent=%d", health, persistent)
		}
	}
	positions := map[featureChunkPosition]bool{{start.Chunk.X, start.Chunk.Z}: true}
	for _, actor := range cached.pending {
		positions[featureChunkPosition{actor.Position.X >> 4, actor.Position.Z >> 4}] = true
	}
	box, _ := start.BoundingBox()
	positions[featureChunkPosition{box.MinX >> 4, box.MinZ >> 4}] = true
	positions[featureChunkPosition{box.MaxX >> 4, box.MaxZ >> 4}] = true
	var keys []featureChunkPosition
	for p := range positions {
		keys = append(keys, p)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].chunkX < keys[j].chunkX || keys[i].chunkX == keys[j].chunkX && keys[i].chunkZ < keys[j].chunkZ
	})
	hashes := make(map[featureChunkPosition][32]byte)
	guardians, solids, magma := 0, 0, 0
	for _, p := range keys {
		chunk, err := g.Generate(p.chunkX, p.chunkZ)
		if err != nil {
			t.Fatal(err)
		}
		refs, _ := chunk.Structures.Object("References")
		if refs["minecraft:monument"].Type != nbt.LongArray {
			t.Fatal("missing monument reference", p)
		}
		if p.chunkX == start.Chunk.X && p.chunkZ == start.Chunk.Z {
			starts, _ := chunk.Structures.Object("starts")
			if _, found := starts.Object("minecraft:monument"); !found {
				t.Fatal("missing saved monument building")
			}
		}
		for _, actor := range chunk.PendingEntities {
			if actor.ID != "minecraft:elder_guardian" {
				continue
			}
			matched := false
			for _, expected := range cached.pending {
				matched = matched || reflect.DeepEqual(actor, expected)
			}
			if !matched {
				t.Fatal("guardian request changed during replay")
			}
			guardians++
		}
		for pos, state := range cached.blocks {
			if pos.X>>4 != p.chunkX || pos.Z>>4 != p.chunkZ {
				continue
			}
			switch state.Name {
			case "minecraft:prismarine", "minecraft:prismarine_bricks", "minecraft:dark_prismarine", "minecraft:gold_block", "minecraft:sea_lantern":
			default:
				continue
			}
			got, err := chunk.Terrain.BlockStateAt(pos.X&15, pos.Y, pos.Z&15)
			if err != nil {
				t.Fatal(err)
			}
			// Vanilla UnderwaterMagmaFeature runs after monuments and accepts
			// any supported non-air/non-water block, including prismarine.
			if got.Name == "minecraft:magma_block" {
				magma++
				continue
			}
			if got.Name != state.Name {
				t.Fatalf("monument block %v got %v want %v (%v)", pos, got, state, err)
			}
			solids++
		}
		hashes[p] = generatedStructureDigest(t, chunk)
		chunk.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
		for i := range chunk.PendingEntities {
			chunk.PendingEntities[i].Data["Health"] = nbt.Tag{Type: nbt.Float, Value: float32(-1)}
		}
	}
	if guardians != 3 || solids < 1000 {
		t.Fatalf("incomplete replay: guardians=%d monument blocks=%d", guardians, solids)
	}
	for i := len(keys) - 1; i >= 0; i-- {
		p := keys[i]
		chunk, err := g.Generate(p.chunkX, p.chunkZ)
		if err != nil {
			t.Fatal(err)
		}
		if got := generatedStructureDigest(t, chunk); got != hashes[p] {
			t.Fatal("reverse replay or returned mutation changed monument", p)
		}
	}
	coverage, err := g.StructureCoverage()
	if err != nil {
		t.Fatal(err)
	}
	enabled := 0
	for _, entry := range coverage {
		if entry.Enabled {
			enabled++
		} else {
			t.Fatal("unexpected missing structure", entry.Name)
		}
	}
	if enabled != 29 {
		t.Fatal("unexpected Overworld coverage", enabled)
	}
	t.Logf("verified %d runtime chunks, %d monument blocks, %d later magma replacements, 3 guardians, saved start/references and reverse replay", len(keys), solids, magma)
}
