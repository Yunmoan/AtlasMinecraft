package worldgen

import (
	"atlas/internal/nbt"
	"reflect"
	"sync"
	"testing"
)

func TestDimensionBastionPipeline(t *testing.T) {
	g, err := NewDimensionGeneratorWithOptions(1234, "nether", DimensionGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	start := acceptedDimensionIntegrationStart(t, g, "minecraft:nether_complexes", "minecraft:bastion_remnant")
	t.Logf("accepted Bastion %d,%d with %d pieces", start.Chunk.X, start.Chunk.Z, len(start.Pieces))
	positions := map[[2]int32]bool{{start.Chunk.X, start.Chunk.Z}: true}
	var chest *FeaturePosition
	actorPositions := make(map[FeaturePosition]string)
	for _, piece := range start.Pieces {
		element, origin, rotation, _, err := jigsawSavedBlockSettings(&piece)
		if err != nil {
			t.Fatal(err)
		}
		template, err := LoadJigsawTemplate(element.Location)
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range template.Palettes[0] {
			if block.State.Name == "minecraft:chest" && chest == nil {
				p, _ := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
				p = treeOffset(origin, p.X, p.Y, p.Z)
				chest = &p
				positions[[2]int32{p.X >> 4, p.Z >> 4}] = true
			}
		}
		for _, actor := range template.Entities {
			p, _ := TransformTemplatePosition(actor.BlockPosition, FeaturePosition{}, TemplateMirrorNone, rotation)
			p = treeOffset(origin, p.X, p.Y, p.Z)
			id, _ := actor.NBT.Text("id")
			actorPositions[p] = id
			positions[[2]int32{p.X >> 4, p.Z >> 4}] = true
		}
	}
	if chest == nil || len(actorPositions) == 0 {
		t.Fatal("Bastion fixture lacks chest/actors")
	}
	var clips [][2]int32
	for position := range positions {
		clips = append(clips, position)
	}
	chunks := make([]*GeneratedDimensionChunk, len(clips))
	errs := make([]error, len(clips))
	var wg sync.WaitGroup
	for i, position := range clips {
		wg.Add(1)
		go func(i int, p [2]int32) { defer wg.Done(); chunks[i], errs[i] = g.Generate(p[0], p[1]) }(i, position)
	}
	wg.Wait()
	chestFound := false
	for i, chunk := range chunks {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		refs, _ := chunk.Structures.Object("References")
		if refs[start.Structure].Type != nbt.LongArray {
			t.Fatal("Bastion references missing")
		}
		if clips[i] == [2]int32{start.Chunk.X, start.Chunk.Z} {
			starts, _ := chunk.Structures.Object("starts")
			saved, found := starts.Object(start.Structure)
			if !found {
				t.Fatal("Bastion owned start missing")
			}
			children, _ := saved["Children"].Elements()
			if len(children) != len(start.Pieces) {
				t.Fatal("Bastion saved Jigsaw piece list incomplete")
			}
		}
		if clips[i] == [2]int32{chest.X >> 4, chest.Z >> 4} {
			state, err := chunk.Terrain.BlockStateAt(chest.X&15, chest.Y, chest.Z&15)
			if err != nil || state.Name != "minecraft:chest" {
				t.Fatalf("Bastion chest missing at %v: %v %v", *chest, state, err)
			}
			for _, entity := range chunk.BlockEntities {
				x, _ := entity.Number("x")
				y, _ := entity.Number("y")
				z, _ := entity.Number("z")
				if x == int64(chest.X) && y == int64(chest.Y) && z == int64(chest.Z) {
					loot, _ := entity.Text("LootTable")
					if loot == "" {
						t.Fatal("Bastion chest loot missing")
					}
					chestFound = true
				}
			}
		}
		for _, entity := range chunk.PendingEntities {
			if want, found := actorPositions[entity.Position]; found && want == entity.ID {
				if !entity.Persistent {
					t.Fatal("authored Bastion mob lost persistence")
				}
				if _, found := entity.Data["Dimension"]; found {
					t.Fatal("obsolete authored Dimension leaked into finalized actor")
				}
				attributes, _ := entity.Data["attributes"].Elements()
				if len(attributes) == 0 {
					t.Fatal("Bastion mob STRUCTURE finalize missing attributes")
				}
				brain, _ := entity.Data.Object("Brain")
				memories, _ := brain.Object("memories")
				if entity.ID == "minecraft:piglin_brute" {
					home, _ := memories.Object("minecraft:home")
					value, _ := home.Object("value")
					dimension, _ := value.Text("dimension")
					if dimension != "minecraft:the_nether" {
						t.Fatal("Bastion brute home dimension wrong")
					}
				}
				delete(actorPositions, entity.Position)
			}
		}
		again, err := g.Generate(clips[i][0], clips[i][1])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(chunk, again) {
			t.Fatal("Bastion concurrent clip replay changed terrain/NBT/entities")
		}
		chunk.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
		if len(chunk.PendingEntities) > 0 {
			chunk.PendingEntities[0].Data["id"] = nbt.Tag{Type: nbt.String, Value: "changed"}
		}
		detached, err := g.Generate(clips[i][0], clips[i][1])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(detached, again) {
			t.Fatal("caller mutation contaminated cached Bastion")
		}
	}
	if !chestFound || len(actorPositions) > 0 {
		t.Fatalf("runtime Bastion chest=%v missing actors=%v", chestFound, actorPositions)
	}
}
