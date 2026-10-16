package worldgen

import (
	"atlas/internal/nbt"
	"reflect"
	"sync"
	"testing"
)

func endCityRuntimeFixture(t *testing.T, g *DimensionGenerator) StructureStart {
	t.Helper()
	p := g.structures.placement.Sets["minecraft:end_cities"].Placement
	for rx := int32(4); rx <= 14; rx++ {
		for rz := int32(-4); rz <= 4; rz++ {
			chunk, err := p.PotentialChunk(g.seed, rx*p.Spacing, rz*p.Spacing)
			if err != nil {
				t.Fatal(err)
			}
			starts, err := g.structureStartsAt(chunk.X, chunk.Z)
			if err != nil {
				t.Fatal(err)
			}
			for _, start := range starts {
				for _, piece := range start.Pieces {
					name, _ := piece.Data.Text("Template")
					if name == "ship" {
						t.Logf("End City runtime seed %d owner %d,%d pieces %d ship %v", g.seed, chunk.X, chunk.Z, len(start.Pieces), endCityOrigin(&piece))
						return start
					}
				}
			}
		}
	}
	t.Fatal("no accepted End City ship fixture")
	return StructureStart{}
}

func TestDimensionEndCityStructurePipeline(t *testing.T) {
	g, err := NewDimensionGeneratorWithOptions(1234, "end", DimensionGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	g.SetTerrainWorkers(1)
	coverage, err := g.StructureCoverage()
	if err != nil || len(coverage) != 1 || !coverage[0].Enabled {
		t.Fatalf("End City coverage %v %v", coverage, err)
	}
	start := endCityRuntimeFixture(t, g)
	positions := map[featureChunkPosition]bool{{start.Chunk.X, start.Chunk.Z}: true}
	var ship StructurePiece
	for _, piece := range start.Pieces {
		name, _ := piece.Data.Text("Template")
		if name == "ship" {
			ship = piece
			break
		}
	}
	for x := ship.Box.MinX >> 4; x <= ship.Box.MaxX>>4; x++ {
		for z := ship.Box.MinZ >> 4; z <= ship.Box.MaxZ>>4; z++ {
			positions[featureChunkPosition{x, z}] = true
		}
	}
	var keys []featureChunkPosition
	for p := range positions {
		keys = append(keys, p)
	}
	chunks := make([]*GeneratedDimensionChunk, len(keys))
	errs := make([]error, len(keys))
	var wg sync.WaitGroup
	for i := range keys {
		wg.Add(1)
		go func(i int) { defer wg.Done(); chunks[i], errs[i] = g.Generate(keys[i].chunkX, keys[i].chunkZ) }(i)
	}
	wg.Wait()
	shulkers, frames, chests, blocks := 0, 0, 0, 0
	for i, chunk := range chunks {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		refs, _ := chunk.Structures.Object("References")
		if refs["minecraft:end_city"].Type != nbt.LongArray {
			t.Fatal("End City references missing")
		}
		if keys[i].chunkX == start.Chunk.X && keys[i].chunkZ == start.Chunk.Z {
			starts, _ := chunk.Structures.Object("starts")
			saved, ok := starts.Object("minecraft:end_city")
			if !ok {
				t.Fatal("End City owned start missing")
			}
			children, _ := saved["Children"].Elements()
			if len(children) != len(start.Pieces) {
				t.Fatal("End City Children incomplete")
			}
		}
		for _, state := range chunk.Terrain.SurfaceBlocks {
			if state.Name == "minecraft:purpur_block" || state.Name == "minecraft:end_stone_bricks" {
				blocks++
			}
		}
		for _, entity := range chunk.BlockEntities {
			table, _ := entity.Text("LootTable")
			if table == "minecraft:chests/end_city_treasure" {
				chests++
			}
		}
		for _, actor := range chunk.PendingEntities {
			switch actor.ID {
			case "minecraft:shulker":
				shulkers++
			case "minecraft:item_frame":
				frames++
				item, _ := actor.Data.Object("Item")
				id, _ := item.Text("id")
				if id != "minecraft:elytra" {
					t.Fatal("End ship frame lacks elytra")
				}
			}
		}
		again, err := g.Generate(keys[i].chunkX, keys[i].chunkZ)
		if err != nil || !reflect.DeepEqual(chunk, again) {
			t.Fatalf("End City changed on cache/neighbor replay at %v: %v", keys[i], err)
		}
	}
	if blocks == 0 || chests < 2 || shulkers < 3 || frames != 1 {
		t.Fatalf("actual End City incomplete: %d blocks, %d loot chests, %d shulkers, %d frames", blocks, chests, shulkers, frames)
	}
	t.Logf("actual End City generated %d chunks, %d city blocks, %d loot chests, %d shulkers, %d elytra frame; one terrain worker and concurrent requests", len(keys), blocks, chests, shulkers, frames)
}
