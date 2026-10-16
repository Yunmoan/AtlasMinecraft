package worldgen

import (
	"atlas/internal/nbt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestOverworldMineshaftStructurePipeline(t *testing.T) {
	for _, fixture := range []struct {
		name string
		x, z int32
	}{
		{"minecraft:mineshaft", -16, 22},
		{"minecraft:mineshaft_mesa", -365, 369},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
			if err != nil {
				t.Fatal(err)
			}
			starts, err := g.structureStartsAt(fixture.x, fixture.z)
			if err != nil {
				t.Fatal(err)
			}
			var start *StructureStart
			for index := range starts {
				if starts[index].Structure == fixture.name {
					start = &starts[index]
					break
				}
			}
			if start == nil {
				t.Fatal("pinned mineshaft fixture was rejected")
			}
			initial, err := start.NBT()
			if err != nil {
				t.Fatal(err)
			}
			initialDigest := iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: initial})
			// Simultaneous requests must coalesce one shared, private replay.
			cold := time.Now()
			cacheResults := make([]*structureBlockResult, 4)
			errors := make([]error, 4)
			var wg sync.WaitGroup
			for index := range cacheResults {
				wg.Add(1)
				go func(index int) { defer wg.Done(); cacheResults[index], errors[index] = g.structureBlocksFor(*start) }(index)
			}
			wg.Wait()
			for index := range cacheResults {
				if errors[index] != nil {
					t.Fatal(errors[index])
				}
				if cacheResults[index] != cacheResults[0] {
					t.Fatal("concurrent mine cache request did not coalesce")
				}
			}
			cached := cacheResults[0]
			t.Logf("cold canonical replay: %s, %d pieces, %d writes, %d block entities, %d loot minecarts", time.Since(cold), len(start.Pieces), len(cached.blocks), len(cached.blockEntities), len(cached.pending))
			for p := range cached.blockEntities {
				t.Logf("generated mine spawner at %v", p)
			}
			for _, entity := range cached.pending {
				t.Logf("generated minecart at %v", entity.Position)
			}
			if len(cached.blocks) == 0 {
				t.Fatal("mineshaft placed no blocks")
			}
			spawners, crossClipSpiders := 0, 0
			for index, piece := range cached.start.Pieces {
				if piece.ID != "minecraft:mscorridor" {
					continue
				}
				sc, _ := piece.Data.Number("sc")
				hps, _ := piece.Data.Number("hps")
				if sc != 0 && (piece.Box.MinX>>4 != piece.Box.MaxX>>4 || piece.Box.MinZ>>4 != piece.Box.MaxZ>>4) {
					crossClipSpiders++
				}
				count := 0
				for p, entity := range cached.blockEntities {
					if entity.block == "minecraft:spawner" && structureContains(piece.Box, p) {
						count++
					}
				}
				if count > 1 || hps != 0 && count != 1 || hps == 0 && count != 0 {
					t.Fatalf("corridor %d hps=%d spawners=%d", index, hps, count)
				}
				spawners += count
			}
			if crossClipSpiders == 0 {
				t.Fatal("fixture lacks a cross-chunk spider corridor")
			}
			if spawners == 0 && fixture.name == "minecraft:mineshaft" {
				t.Fatal("fixture did not generate a cave-spider spawner")
			}
			// Generate every chunk receiving any structure record, including owner.
			wanted := map[StructureChunkPosition]bool{start.Chunk: true}
			for p := range cached.blocks {
				wanted[StructureChunkPosition{p.X >> 4, p.Z >> 4}] = true
			}
			for p := range cached.postProcessing {
				wanted[StructureChunkPosition{p.X >> 4, p.Z >> 4}] = true
			}
			for _, entity := range cached.pending {
				wanted[StructureChunkPosition{entity.Position.X >> 4, entity.Position.Z >> 4}] = true
			}
			positions := make([]StructureChunkPosition, 0, len(wanted))
			for position := range wanted {
				positions = append(positions, position)
			}
			sort.Slice(positions, func(i, j int) bool {
				return positions[i].X < positions[j].X || positions[i].X == positions[j].X && positions[i].Z < positions[j].Z
			})
			chunks := make([]*GeneratedOverworldChunk, len(positions))
			chunkErrors := make([]error, len(positions))
			jobs := make(chan int, len(positions))
			for index := range positions {
				jobs <- index
			}
			close(jobs)
			for worker := 0; worker < 4; worker++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for index := range jobs {
						p := positions[index]
						chunks[index], chunkErrors[index] = g.Generate(p.X, p.Z)
					}
				}()
			}
			wg.Wait()
			replayed := make(map[StructureChunkPosition]*GeneratedOverworldChunk)
			actualMinecarts := 0
			box, _ := start.BoundingBox()
			for index, ch := range chunks {
				if chunkErrors[index] != nil {
					t.Fatal(chunkErrors[index])
				}
				position := positions[index]
				replayed[position] = ch
				if position == start.Chunk {
					owned, _ := ch.Structures.Object("starts")
					actual, ok := owned.Object(start.Structure)
					want, err := cached.start.NBT()
					if err != nil || !ok || iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: actual}) != iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: want}) {
						t.Fatal("owned mine lost postprocess flags or complete saved pieces")
					}
				}
				if box.MaxX >= position.X*16 && box.MinX <= position.X*16+15 && box.MaxZ >= position.Z*16 && box.MinZ <= position.Z*16+15 {
					refs, _ := ch.Structures.Object("References")
					values, _ := refs[start.Structure].Longs()
					found := false
					for _, value := range values {
						found = found || value == int64(uint64(uint32(start.Chunk.X))|uint64(uint32(start.Chunk.Z))<<32)
					}
					if !found {
						t.Fatalf("missing mine reference in %v", position)
					}
				}
				for _, entity := range ch.PendingEntities {
					if entity.ID == "minecraft:chest_minecart" {
						actualMinecarts++
					}
				}
			}
			for p, entity := range cached.blockEntities {
				ch := replayed[StructureChunkPosition{p.X >> 4, p.Z >> 4}]
				if ch == nil {
					t.Fatalf("block entity chunk not replayed: %v", p)
				}
				state, err := ch.Terrain.BlockStateAt(p.X&15, p.Y, p.Z&15)
				if err != nil || state.Name != entity.block {
					t.Fatalf("mine spawner missing at %v: %v %v", p, state, err)
				}
				want := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
				want["x"], want["y"], want["z"] = nbt.Tag{Type: nbt.Int, Value: p.X}, nbt.Tag{Type: nbt.Int, Value: p.Y}, nbt.Tag{Type: nbt.Int, Value: p.Z}
				matched := false
				for _, actual := range ch.BlockEntities {
					matched = matched || iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: actual}) == iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: want})
				}
				if !matched {
					t.Fatalf("mine spawner NBT changed at %v", p)
				}
			}
			for p := range cached.postProcessing {
				ch := replayed[StructureChunkPosition{p.X >> 4, p.Z >> 4}]
				index, err := ch.Terrain.Index(p.X&15, p.Y, p.Z&15)
				if err != nil || index >= len(ch.Terrain.PostProcessing) || !ch.Terrain.PostProcessing[index] {
					t.Fatalf("mine shape update lost at %v", p)
				}
			}
			for _, want := range cached.pending {
				ch := replayed[StructureChunkPosition{want.Position.X >> 4, want.Position.Z >> 4}]
				matched := 0
				for _, actual := range ch.PendingEntities {
					if reflect.DeepEqual(actual, want) {
						matched++
					}
				}
				if matched != 1 {
					t.Fatalf("loot minecart request replayed %d times at %v", matched, want.Position)
				}
			}
			checkReturnedNBT := func(position StructureChunkPosition) {
				ch := replayed[position]
				digest := generatedStructureDigest(t, ch)
				if len(ch.BlockEntities) > 0 {
					ch.BlockEntities[0]["id"] = nbt.Tag{Type: nbt.String, Value: "mutated"}
				}
				for index := range ch.PendingEntities {
					if ch.PendingEntities[index].ID == "minecraft:chest_minecart" {
						ch.PendingEntities[index].Data["LootTable"] = nbt.Tag{Type: nbt.String, Value: "mutated"}
					}
				}
				second, err := g.Generate(position.X, position.Z)
				if err != nil || generatedStructureDigest(t, second) != digest {
					t.Fatalf("returned remote mine NBT mutation changed cache in %v: %v", position, err)
				}
				replayed[position] = second
			}
			for p := range cached.blockEntities {
				checkReturnedNBT(StructureChunkPosition{p.X >> 4, p.Z >> 4})
				break
			}
			if len(cached.pending) > 0 {
				p := cached.pending[0].Position
				checkReturnedNBT(StructureChunkPosition{p.X >> 4, p.Z >> 4})
			}
			// Mutable results never leak into the immutable start or record cache.
			owner := replayed[start.Chunk]
			digest := generatedStructureDigest(t, owner)
			owner.Terrain.Blocks[0] = NoiseLava
			owner.Structures["starts"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
			if len(owner.BlockEntities) > 0 {
				owner.BlockEntities[0]["id"] = nbt.Tag{Type: nbt.String, Value: "mutated"}
			}
			if len(owner.PendingEntities) > 0 {
				owner.PendingEntities[0].Data["id"] = nbt.Tag{Type: nbt.String, Value: "mutated"}
			}
			again, err := g.Generate(start.Chunk.X, start.Chunk.Z)
			if err != nil || generatedStructureDigest(t, again) != digest {
				t.Fatalf("returned mine mutation changed cache: %v", err)
			}
			unchanged, err := start.NBT()
			if err != nil || iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: unchanged}) != initialDigest {
				t.Fatal("postprocessing mutated cached initial start")
			}
			t.Logf("replayed %d chunks; unique spider spawners=%d; own loot requests=%d; total neighboring minecarts=%d", len(chunks), spawners, len(cached.pending), actualMinecarts)
		})
	}
}
