package worldgen

import (
	"atlas/internal/nbt"
	"reflect"
	"sort"
	"testing"
)

func TestOverworldMansionPipeline(t *testing.T) {
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	start := supportedIntegrationStart(t, g, "minecraft:woodland_mansions", "minecraft:mansion")
	t.Logf("accepted mansion at chunk %d,%d with %d pieces", start.Chunk.X, start.Chunk.Z, len(start.Pieces))
	positions := map[featureChunkPosition]bool{{start.Chunk.X, start.Chunk.Z}: true}
	box, _ := start.BoundingBox()
	positions[featureChunkPosition{box.MinX >> 4, box.MinZ >> 4}] = true
	positions[featureChunkPosition{box.MaxX >> 4, box.MaxZ >> 4}] = true
	markers := make(map[string]bool)
	for _, piece := range start.Pieces {
		name, _ := piece.Data.Text("Template")
		template, err := LoadMansionTemplate(name)
		if err != nil {
			t.Fatal(err)
		}
		rot, _ := piece.Data.Text("Rot")
		mi, _ := piece.Data.Text("Mi")
		var rotation TemplateRotation
		var mirror TemplateMirror
		for i, n := range templateRotationNames {
			if n == rot {
				rotation = TemplateRotation(i)
			}
		}
		for i, n := range []string{"NONE", "LEFT_RIGHT", "FRONT_BACK"} {
			if n == mi {
				mirror = TemplateMirror(i)
			}
		}
		for _, block := range template.Palettes[0] {
			if block.NBT == nil || block.State.Name != "minecraft:structure_block" {
				continue
			}
			marker, _ := block.NBT.Text("metadata")
			if marker != "Mage" && marker != "Warrior" && marker != "Group of Allays" {
				continue
			}
			if markers[marker] {
				continue
			}
			markers[marker] = true
			p, err := TransformTemplatePosition(block.Position, FeaturePosition{}, mirror, rotation)
			if err != nil {
				t.Fatal(err)
			}
			p = treeOffset(endCityOrigin(&piece), p.X, p.Y, p.Z)
			positions[featureChunkPosition{p.X >> 4, p.Z >> 4}] = true
		}
	}
	var keys []featureChunkPosition
	for p := range positions {
		keys = append(keys, p)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].chunkX < keys[j].chunkX || keys[i].chunkX == keys[j].chunkX && keys[i].chunkZ < keys[j].chunkZ
	})
	hashes := make(map[featureChunkPosition][32]byte)
	actorCount, solids := 0, 0
	for _, p := range keys {
		cached, err := g.clippedStructureBlocksFor(start, p.chunkX, p.chunkZ)
		if err != nil {
			t.Fatal(err)
		}
		chunk, err := g.Generate(p.chunkX, p.chunkZ)
		if err != nil {
			t.Fatal(err)
		}
		refs, _ := chunk.Structures.Object("References")
		if refs["minecraft:mansion"].Type != nbt.LongArray {
			t.Fatal("missing mansion reference", p)
		}
		if p.chunkX == start.Chunk.X && p.chunkZ == start.Chunk.Z {
			starts, _ := chunk.Structures.Object("starts")
			if _, found := starts.Object("minecraft:mansion"); !found {
				t.Fatal("missing saved mansion start")
			}
		}
		for _, actor := range cached.pending {
			found := false
			for _, got := range chunk.PendingEntities {
				found = found || reflect.DeepEqual(got, actor)
			}
			if !found || actor.SavedFlag == "" {
				t.Fatal("missing mansion actor", actor.ID, actor.Position)
			}
			actorCount++
		}
		for pos, state := range cached.blocks {
			if pos.X>>4 != p.chunkX || pos.Z>>4 != p.chunkZ {
				continue
			}
			if state.Name != "minecraft:dark_oak_planks" && state.Name != "minecraft:dark_oak_log" {
				continue
			}
			got, err := chunk.Terrain.BlockStateAt(pos.X&15, pos.Y, pos.Z&15)
			if err != nil || !reflect.DeepEqual(got, state) {
				t.Fatalf("mansion block %v: got %v want %v (%v)", pos, got, state, err)
			}
			solids++
		}
		hashes[p] = generatedStructureDigest(t, chunk)
		chunk.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
		for i := range chunk.PendingEntities {
			chunk.PendingEntities[i].Data["Health"] = nbt.Tag{Type: nbt.Float, Value: float32(-1)}
		}
	}
	if actorCount == 0 || solids < 100 {
		t.Fatalf("incomplete mansion replay: actors=%d solids=%d", actorCount, solids)
	}
	for i := len(keys) - 1; i >= 0; i-- {
		p := keys[i]
		chunk, err := g.Generate(p.chunkX, p.chunkZ)
		if err != nil {
			t.Fatal(err)
		}
		if generatedStructureDigest(t, chunk) != hashes[p] {
			t.Fatal("mansion replay changed", p)
		}
	}
	t.Logf("verified %d mansion clips, %d actors, %d solid blocks and reverse replay", len(keys), actorCount, solids)
}
