package worldgen

import (
	"atlas/internal/nbt"
	"reflect"
	"sync"
	"testing"
)

func pinnedOutpostRuntimeStart(t *testing.T, g *OverworldGenerator) StructureStart {
	t.Helper()
	eligible, err := g.structures.placement.IsStructureChunk("minecraft:pillager_outposts", -249, 327)
	if err != nil || !eligible {
		t.Fatal("pinned outpost placement rejected", err)
	}
	starts, err := g.structureStartsAt(-249, 327)
	if err != nil {
		t.Fatal(err)
	}
	for _, start := range starts {
		if start.Structure == "minecraft:pillager_outpost" {
			if len(start.Pieces) != 15 || start.TerrainAdjustment != "beard_thin" {
				t.Fatal("pinned outpost layout/terrain adjustment changed")
			}
			return start
		}
	}
	t.Fatal("actual biome rejected pinned outpost")
	return StructureStart{}
}

func TestOverworldOutpostStructurePipeline(t *testing.T) {
	g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
	if err != nil {
		t.Fatal(err)
	}
	// A start-height callback that re-enters base generation deadlocks with
	// this budget. Concurrent final chunks must share it without nesting it.
	g.SetTerrainWorkers(1)
	start := pinnedOutpostRuntimeStart(t, g)
	original, err := start.NBT()
	if err != nil {
		t.Fatal(err)
	}
	positions := []StructureChunkPosition{{-249, 327}, {-248, 328}, {-250, 328}, {-248, 326}}
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
	counts := make(map[string]int)
	flags := make(map[string]bool)
	writes, chests := 0, 0
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
			t.Fatal("outpost reference missing", pos)
		}
		if i == 0 {
			starts, _ := chunk.Structures.Object("starts")
			saved, ok := starts.Object(start.Structure)
			if !ok || !reflect.DeepEqual(saved, original) {
				t.Fatal("outpost lost its full original start NBT")
			}
		}
		cached, err := g.jigsawStructureBlocksFor(start, pos.X, pos.Z)
		if err != nil || len(cached.blocks) == 0 {
			t.Fatal("outpost clip placed no blocks", err)
		}
		matchedBlocks := 0
		for p, want := range cached.blocks {
			if p.X>>4 != pos.X || p.Z>>4 != pos.Z {
				continue
			}
			got, err := chunk.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(got, want) && want.Name != "minecraft:air" {
				matchedBlocks++
			}
		}
		if matchedBlocks < 30 {
			t.Fatal("outpost metadata without actual tower/cage blocks", pos, matchedBlocks)
		}
		writes += matchedBlocks
		for _, data := range chunk.BlockEntities {
			table, _ := data.Text("LootTable")
			if table != "minecraft:chests/pillager_outpost" {
				continue
			}
			chests++
			x, _ := data.Number("x")
			y, _ := data.Number("y")
			z, _ := data.Number("z")
			if x != -3974 || y != 109 || z != 5242 || data["LootTableSeed"].Type != nbt.Long {
				t.Fatal("outpost tower lost authored chest position/seed")
			}
		}
		regionRandom := NewStructureRegionRandom(g.seed, pos.X, pos.Z)
		for _, actor := range cached.pending {
			if actor.Position.X>>4 != pos.X || actor.Position.Z>>4 != pos.Z || actor.SavedFlag == "" || flags[actor.SavedFlag] {
				t.Fatal("outpost actor clip or stable identity invalid")
			}
			flags[actor.SavedFlag] = true
			counts[actor.ID]++
			// Compare against original authored NBT through the independently
			// JAR-verified load/finalize helper and the chunk's region stream.
			name := "minecraft:pillager_outpost/feature_cage_with_allays"
			if actor.ID == "minecraft:iron_golem" {
				name = "minecraft:pillager_outpost/feature_cage1"
			}
			template, err := LoadJigsawTemplate(name)
			if err != nil {
				t.Fatal(err)
			}
			ordinal := counts[actor.ID] - 1
			if actor.ID == "minecraft:allay" {
				ordinal %= 2
			}
			authored := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: template.Entities[ordinal].NBT}).Value.(nbt.Object)
			authored["Pos"], authored["Rotation"] = actor.Data["Pos"], actor.Data["Rotation"]
			want, err := FinalizeOutpostTemplateEntity(PendingStructureEntity{ID: actor.ID, Data: authored}, regionRandom)
			if err != nil || !reflect.DeepEqual(actor.Data, want) {
				t.Fatal("outpost actor lost full normalized NBT or original region RNG order", err)
			}
			matched := false
			for _, pending := range chunk.PendingEntities {
				matched = matched || reflect.DeepEqual(pending, actor)
			}
			if !matched {
				t.Fatal("outpost actor not installed in final normal chunk")
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
			t.Fatal("outpost cache changed after returned mutation or parallel generation", err)
		}
	}
	if chests != 1 || counts["minecraft:allay"] != 4 || counts["minecraft:iron_golem"] != 1 || len(counts) != 2 {
		t.Fatal("actual outpost chest/actor coverage changed", chests, counts)
	}
	// Terrain adjustment is evaluated in NOISE, before authored tower blocks.
	raw, err := g.noise.GenerateNoise(start.Chunk.X, start.Chunk.Z)
	if err != nil {
		t.Fatal(err)
	}
	adjusted, err := g.noise.GenerateNoiseWithStructures(start.Chunk.X, start.Chunk.Z, structureTerrainPieces([]StructureStart{start}))
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for i, material := range raw.Blocks {
		if material != adjusted.Blocks[i] {
			changed++
		}
	}
	if changed == 0 {
		t.Fatal("outpost beard_thin omitted from actual noise")
	}
	t.Logf("actual seed1234 owner(-249,327):15 pieces,4 parallel normal chunks with one terrain worker,%d retained structure blocks,one seeded chest,four allays,one golem;%d adjusted noise blocks", writes, changed)
}
