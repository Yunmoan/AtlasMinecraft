package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"reflect"
	"testing"
)

func pinnedStrongholdRuntimeStart(t testing.TB, g *OverworldGenerator) StructureStart {
	t.Helper()
	eligible, err := g.strongholdPlacementAt(-100, -125)
	if err != nil || !eligible {
		t.Fatal("original seed-1234 first ring placement rejected", err)
	}
	starts, err := g.structureStartsAt(-100, -125)
	if err != nil {
		t.Fatal(err)
	}
	for _, start := range starts {
		if start.Structure == "minecraft:stronghold" {
			return start
		}
	}
	t.Fatal("actual biome rejected pinned stronghold")
	return StructureStart{}
}

func strongholdRuntimeClips(t testing.TB, start StructureStart) ([]StructureChunkPosition, FeaturePosition, []FeaturePosition, map[FeaturePosition]string) {
	t.Helper()
	clips := []StructureChunkPosition{start.Chunk}
	add := func(p FeaturePosition) {
		pos := StructureChunkPosition{p.X >> 4, p.Z >> 4}
		for _, existing := range clips {
			if existing == pos {
				return
			}
		}
		clips = append(clips, pos)
	}
	var spawner FeaturePosition
	var frames []FeaturePosition
	chests := make(map[FeaturePosition]string)
	library, corridor, portal := false, false, false
	for _, piece := range start.Pieces {
		switch piece.ID {
		case "minecraft:shpr":
			portal = true
			spawner = piece.BlockPosition(5, 3, 6)
			add(spawner)
			for _, at := range [][2]int32{{4, 8}, {5, 8}, {6, 8}, {4, 12}, {5, 12}, {6, 12}, {3, 9}, {3, 10}, {3, 11}, {7, 9}, {7, 10}, {7, 11}} {
				p := piece.BlockPosition(at[0], 3, at[1])
				frames = append(frames, p)
				add(p)
			}
		case "minecraft:shli":
			if library {
				continue
			}
			library = true
			p := piece.BlockPosition(3, 3, 5)
			chests[p] = "minecraft:chests/stronghold_library"
			add(p)
			if strongholdFlag(piece, "Tall") {
				p := piece.BlockPosition(12, 8, 1)
				chests[p] = "minecraft:chests/stronghold_library"
				add(p)
			}
		case "minecraft:shcc":
			if corridor {
				continue
			}
			corridor = true
			p := piece.BlockPosition(3, 2, 3)
			chests[p] = "minecraft:chests/stronghold_corridor"
			add(p)
		}
	}
	if !portal || !library || !corridor {
		t.Fatal("pinned stronghold has no portal/library/corridor")
	}
	return clips, spawner, frames, chests
}

func TestOverworldStrongholdStructurePipeline(t *testing.T) {
	results := make([]map[StructureChunkPosition]*GeneratedOverworldChunk, 2)
	var snapshots []StructureStart
	for order := 0; order < 2; order++ {
		g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
		if err != nil {
			t.Fatal(err)
		}
		g.SetTerrainWorkers(1)
		start := pinnedStrongholdRuntimeStart(t, g)
		original := start.Clone()
		positions, spawner, frames, chests := strongholdRuntimeClips(t, start)
		if order == 0 {
			t.Logf("seed 1234 owner %v pieces %d clips %v spawner %v chests %v", start.Chunk, len(start.Pieces), positions, spawner, chests)
		} else {
			for left, right := 0, len(positions)-1; left < right; left, right = left+1, right-1 {
				positions[left], positions[right] = positions[right], positions[left]
			}
		}
		results[order] = make(map[StructureChunkPosition]*GeneratedOverworldChunk)
		retained := 0
		for _, position := range positions {
			chunk, err := g.Generate(position.X, position.Z)
			if err != nil {
				t.Fatal(err)
			}
			results[order][position] = chunk
			refs, _ := chunk.Structures.Object("References")
			owners, _ := refs[start.Structure].Longs()
			packed := int64(uint64(uint32(start.Chunk.X)) | uint64(uint32(start.Chunk.Z))<<32)
			found := false
			for _, value := range owners {
				found = found || value == packed
			}
			if !found {
				t.Fatal("actual stronghold chunk has no owner reference", position)
			}
			cached, err := g.strongholdStructureBlocksFor(start, position.X, position.Z)
			if err != nil {
				t.Fatal(err)
			}
			for p, want := range cached.blocks {
				got, err := chunk.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
				if err != nil {
					t.Fatal(err)
				}
				gotID, err := blockstate.BlockStateID(got.Name, got.Properties)
				if err != nil {
					t.Fatal(err)
				}
				wantID, err := blockstate.BlockStateID(want.Name, want.Properties)
				if err != nil {
					t.Fatal(err)
				}
				if gotID == wantID && !surfaceAir(want) {
					retained++
				}
			}
		}
		if retained < 500 {
			t.Fatal("stronghold metadata without actual retained walls/library/portal", retained)
		}
		for _, p := range frames {
			chunk := results[order][StructureChunkPosition{p.X >> 4, p.Z >> 4}]
			state, err := chunk.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
			if err != nil || state.Name != "minecraft:end_portal_frame" {
				t.Fatal("original portal frame missing in actual normal chunk", p, err)
			}
		}
		wantEntities := make(map[FeaturePosition]string, len(chests)+1)
		wantEntities[spawner] = "minecraft:mob_spawner"
		for p := range chests {
			wantEntities[p] = "minecraft:chest"
		}
		for p, id := range wantEntities {
			found := false
			for _, data := range results[order][StructureChunkPosition{p.X >> 4, p.Z >> 4}].BlockEntities {
				x, _ := data.Number("x")
				y, _ := data.Number("y")
				z, _ := data.Number("z")
				if x != int64(p.X) || y != int64(p.Y) || z != int64(p.Z) {
					continue
				}
				gotID, _ := data.Text("id")
				if gotID != id {
					t.Fatal("wrong actual block-entity identity", p, gotID)
				}
				if id == "minecraft:chest" {
					table, _ := data.Text("LootTable")
					if table != chests[p] {
						t.Fatal("stronghold chest lost its original loot table", p, table)
					}
					if data["LootTableSeed"].Type != nbt.Long {
						t.Fatal("stronghold chest has no original long loot seed", p)
					}
				} else {
					spawn, _ := data.Object("SpawnData")
					entity, _ := spawn.Object("entity")
					kind, _ := entity.Text("id")
					if kind != "minecraft:silverfish" {
						t.Fatal("portal spawner has no original silverfish SpawnData", kind)
					}
				}
				found = true
			}
			if !found {
				t.Fatal("normal stronghold entity missing", p, id)
			}
		}
		if !reflect.DeepEqual(start, original) {
			t.Fatal("clipped generation mutated immutable stronghold start")
		}
		saved, exists := g.StrongholdStartSnapshot(start.Chunk.X, start.Chunk.Z)
		if !exists {
			t.Fatal("stronghold has no final metadata snapshot")
		}
		for _, piece := range saved.Pieces {
			if piece.ID == "minecraft:shpr" && !strongholdFlag(piece, "Mob") {
				t.Fatal("remote portal clip flag was not merged into owner metadata")
			}
		}
		if order == 0 {
			snapshots = append(snapshots, saved.Clone())
		} else if !reflect.DeepEqual(saved, snapshots[0]) {
			t.Fatal("stronghold metadata depends on clip generation order")
		}
		// Snapshot properties and flags are detached; editing a returned copy
		// must not poison subsequent generated owner NBT or the cached layout.
		saved.Pieces[0].Data["poison"] = nbt.Tag{Type: nbt.String, Value: "bad"}
		owner, err := g.Generate(start.Chunk.X, start.Chunk.Z)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(owner.Terrain, results[order][start.Chunk].Terrain) || !reflect.DeepEqual(owner.BlockEntities, results[order][start.Chunk].BlockEntities) {
			t.Fatal("merged saved flags changed regenerated owner geometry or loot")
		}
		starts, _ := owner.Structures.Object("starts")
		actual, _ := starts.Object("minecraft:stronghold")
		wanted, err := snapshots[0].NBT()
		if err != nil || !reflect.DeepEqual(actual, wanted) {
			t.Fatal("owner metadata did not retain full final original start/flags", err)
		}
		t.Logf("order %d retained %d structure blocks across %d actual chunks, twelve frames, original loot/spawner NBT and coalesced owner flags", order, retained, len(positions))
	}
	for position, before := range results[0] {
		after := results[1][position]
		if !reflect.DeepEqual(before.Terrain, after.Terrain) || !reflect.DeepEqual(before.Biomes, after.Biomes) || !reflect.DeepEqual(before.BlockEntities, after.BlockEntities) {
			t.Fatal("normal stronghold terrain or block entities depend on generation order", position)
		}
	}
}
