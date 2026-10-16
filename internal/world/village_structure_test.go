package world

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestVillageStructureAnvilRestart(t *testing.T) {
	fixtures := []struct {
		name      string
		pieces    int
		positions [][2]int32
	}{
		{"plains", 230, [][2]int32{{-24, 47}, {-28, 45}, {-25, 46}, {-27, 47}, {-25, 47}}},
		{"savanna", 117, [][2]int32{{69, -23}, {68, -23}, {68, -22}, {69, -22}}},
		{"taiga", 113, [][2]int32{{-112, 93}, {-111, 89}, {-113, 93}, {-114, 92}, {-112, 92}}},
		{"snowy", 128, [][2]int32{{-167, 36}, {-171, 35}, {-168, 39}, {-165, 35}, {-168, 36}}},
		{"desert", 102, [][2]int32{{-403, 331}, {-405, 334}, {-403, 330}, {-402, 331}}},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "world")
			seed := int64(1234)
			if err := CreateNormalWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			w, err := OpenDimension(noViewers{}, 4, dir, "overworld")
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			w.store.vanilla.normal.SetTerrainWorkers(1)
			w.AdvanceTime(2053)
			w.SetDayTime(18000)
			if err := w.Save(); err != nil {
				t.Fatal(err)
			}
			metadataBytes, err := readGzip(filepath.Join(dir, "level.dat"))
			if err != nil {
				t.Fatal(err)
			}
			_, metadata, err := nbt.Parse(metadataBytes)
			if err != nil {
				t.Fatal(err)
			}
			name := "minecraft:village_" + fixture.name
			pending := make(map[[2]int32]nbt.Tag)
			wantActors := make(map[[16]byte]WorldEntity)
			wantBlockEntities := make(map[[2]int32]nbt.Tag)
			wantStructures := make(map[[2]int32]nbt.Tag)
			markers := make(map[[3]int32]uint32)
			loot, crops, camels, villageActors := 0, 0, 0, 0
			for _, pos := range fixture.positions {
				generated, err := w.store.vanilla.normal.Generate(pos[0], pos[1])
				if err != nil {
					t.Fatal(err)
				}
				chunk := newChunkBounds(pos[0], pos[1], -64, 384, false)
				if err := chunk.initializeAnvilRoot(); err != nil {
					t.Fatal(err)
				}
				installGeneratedStructures(chunk, generated)
				root, _ := chunk.anvilRoot.Object()
				pending[pos] = root["AtlasPendingStructureEntities"]
				wantStructures[pos] = root["structures"]
				requests, _ := pending[pos].Elements()
				for _, tag := range requests {
					request, _ := tag.Object()
					if flag, _ := request.Text("PieceFlag"); strings.HasPrefix(flag, "Jigsaw:"+name+":") {
						villageActors++
					}
					actor, supported, err := realizeGeneratedEntity(seed, "minecraft:overworld", request)
					if err != nil || !supported {
						t.Fatal("authored village actor unsupported by runtime", err)
					}
					if _, duplicate := wantActors[actor.UUID]; duplicate {
						t.Fatal("authored village actors aliased UUID")
					}
					if int32(math.Floor(actor.Position[0]))>>4 != pos[0] || int32(math.Floor(actor.Position[2]))>>4 != pos[1] {
						t.Fatal("generated actor queued in the wrong precise chunk", actor.Type, actor.Position, pos)
					}
					wantActors[actor.UUID] = actor
					if actor.Type == "minecraft:camel" {
						camels++
						pose, _ := actor.Data.Number("LastPoseTick")
						fire, _ := actor.Data.Number("Fire")
						if actor.Position != [3]float64{-6446, 67, 5291} || pos != [2]int32{-403, 330} || pose != 2000 || fire != 0 || actor.Data["Fire"].Type != nbt.Short {
							t.Fatal("native Camel lost precise Pos, GameTime-derived pose or authored Fire0", actor.Position, pose, fire)
						}
					}
				}
				blockEntities := make([]nbt.Tag, len(generated.BlockEntities))
				for i, data := range generated.BlockEntities {
					blockEntities[i] = nbt.Tag{Type: nbt.Compound, Value: data}
					if table, _ := data.Text("LootTable"); strings.HasPrefix(table, "minecraft:chests/village/") {
						loot++
						if data["LootTableSeed"].Type != nbt.Long {
							t.Fatal("village loot seed is not Long")
						}
					}
				}
				wantBlockEntities[pos] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: blockEntities}
				// Check meaningful runtime block positions without making a
				// mailbox round trip for every voxel of each clipped village.
				seenNames := make(map[string]bool)
				for index, state := range generated.Terrain.SurfaceBlocks {
					crop := state.Name == "minecraft:wheat" || state.Name == "minecraft:carrots" || state.Name == "minecraft:potatoes" || state.Name == "minecraft:beetroots"
					if crop {
						crops++
					}
					if seenNames[state.Name] || !(crop || state.Name == "minecraft:farmland" || state.Name == "minecraft:dirt_path" || state.Name == "minecraft:chest" || state.Name == "minecraft:smooth_sandstone") {
						continue
					}
					seenNames[state.Name] = true
					id, err := blockstate.BlockStateID(state.Name, state.Properties)
					if err != nil {
						t.Fatal(err)
					}
					p := [3]int32{pos[0]*16 + int32(index&15), int32(index/256) + generated.Terrain.MinY, pos[1]*16 + int32(index/16&15)}
					markers[p] = id
				}
			}
			if loot == 0 || crops == 0 || villageActors == 0 || (fixture.name == "desert" && camels != 1) {
				t.Fatal("native village fixture lost loot, crops or actors", loot, crops, len(wantActors), camels)
			}
			errors := make([]error, len(fixture.positions))
			var wg sync.WaitGroup
			for i, pos := range fixture.positions {
				wg.Add(1)
				go func(i int, pos [2]int32) { defer wg.Done(); _, errors[i] = w.LoadChunk(pos[0], pos[1]) }(i, pos)
			}
			wg.Wait()
			for _, err := range errors {
				if err != nil {
					t.Fatal(err)
				}
			}
			checkActors := func(current *World) {
				t.Helper()
				seen := make(map[[16]byte]bool)
				for _, pos := range fixture.positions {
					actors, err := current.EntitiesInChunk(pos[0], pos[1])
					if err != nil {
						t.Fatal(err)
					}
					for _, actor := range actors {
						want, tracked := wantActors[actor.UUID]
						if !tracked {
							for _, expected := range wantActors {
								if actor.Type == expected.Type && actor.Position == expected.Position {
									t.Fatal("village actor replay created another UUID at the authored position", actor.Type, pos)
								}
							}
							continue
						} // Natural biome animals have separate UUIDs.
						want.LoadedFromDisk = actor.LoadedFromDisk
						if seen[actor.UUID] || !reflect.DeepEqual(actor, want) {
							t.Fatal("village actor UUID/full NBT changed or duplicated", actor.Type, pos)
						}
						if int32(math.Floor(actor.Position[0]))>>4 != pos[0] || int32(math.Floor(actor.Position[2]))>>4 != pos[1] {
							t.Fatal("runtime village actor in wrong precise chunk")
						}
						seen[actor.UUID] = true
					}
				}
				if len(seen) != len(wantActors) {
					t.Fatal("village runtime actors missing", len(seen), len(wantActors))
				}
			}
			checkMarkers := func(current *World) {
				t.Helper()
				for p, want := range markers {
					if got, err := current.BlockAt(p[0], p[1], p[2]); err != nil || got != want {
						t.Fatal("native village runtime block changed", p, got, want, err)
					}
				}
			}
			checkActors(w)
			checkMarkers(w)
			owner := fixture.positions[0]
			edit := [3]int32{owner[0]*16 + 8, 280, owner[1]*16 + 8}
			if err := w.SetBlock(edit[0], edit[1], edit[2], StateStone); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			saved := make(map[[2]int32]nbt.Object)
			savedActors := make(map[[16]byte]bool)
			for _, pos := range fixture.positions {
				_, chunkTag, found, err := anvil.ReadChunk(filepath.Join(dir, "region"), pos[0], pos[1])
				if err != nil || !found {
					t.Fatal("village block chunk not saved", err)
				}
				root, _ := chunkTag.Object()
				saved[pos] = root
				if _, retained := root["AtlasPendingStructureEntities"]; retained {
					t.Fatal("realized village actors left pending")
				}
				if !reflect.DeepEqual(root["structures"], wantStructures[pos]) || !reflect.DeepEqual(root["block_entities"], wantBlockEntities[pos]) {
					t.Fatal("saved full village structure or loot NBT changed", pos)
				}
				structures, _ := root.Object("structures")
				refs, _ := structures.Object("References")
				packed, _ := refs[name].Longs()
				ownerPacked := int64(uint64(uint32(owner[0])) | uint64(uint32(owner[1]))<<32)
				present := false
				for _, value := range packed {
					present = present || value == ownerPacked
				}
				if !present {
					t.Fatal("saved remote village lost owner reference", pos)
				}
				if pos == owner {
					starts, _ := structures.Object("starts")
					start, found := starts.Object(name)
					children, _ := start["Children"].Elements()
					if !found || len(children) != fixture.pieces {
						t.Fatal("saved village start lost authored children")
					}
				}
				_, entityTag, found, err := anvil.ReadChunk(filepath.Join(dir, "entities"), pos[0], pos[1])
				if err != nil {
					t.Fatal(err)
				}
				if found {
					entityRoot, _ := entityTag.Object()
					actors, _ := entityRoot["Entities"].Elements()
					for _, tag := range actors {
						data, _ := tag.Object()
						actor, err := worldEntityFromNBT(data)
						if err != nil {
							t.Fatal(err)
						}
						want, tracked := wantActors[actor.UUID]
						if !tracked {
							for _, expected := range wantActors {
								if actor.Type == expected.Type && actor.Position == expected.Position {
									t.Fatal("saved village actor replay has another UUID at authored position", actor.Type, pos)
								}
							}
							continue
						}
						if savedActors[actor.UUID] || data["UUID"].Type != nbt.IntArray || !reflect.DeepEqual(data, want.Data) || int32(math.Floor(actor.Position[0]))>>4 != pos[0] || int32(math.Floor(actor.Position[2]))>>4 != pos[1] {
							t.Fatal("saved village actor UUID/full NBT or precise region ownership invalid", actor.Type, pos)
						}
						savedActors[actor.UUID] = true
					}
				}
				// Simulate actors saved before the block-region request ack.
				requests, _ := pending[pos].Elements()
				if len(requests) > 0 {
					root["AtlasPendingStructureEntities"] = pending[pos]
					if err := w.store.vanilla.regions.WriteChunk(filepath.Join(dir, "region"), pos[0], pos[1], "", chunkTag); err != nil {
						t.Fatal(err)
					}
				}
			}
			if len(savedActors) != len(wantActors) {
				t.Fatal("village entity-region actors missing", len(savedActors), len(wantActors))
			}
			reopened, err := OpenDimension(noViewers{}, 4, dir, "overworld")
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			checkActors(reopened)
			checkMarkers(reopened)
			for _, pos := range fixture.positions {
				chunk, err := reopened.store.vanilla.loadBlockChunk(pos[0], pos[1])
				if err != nil {
					t.Fatal(err)
				}
				root, _ := chunk.anvilRoot.Object()
				if _, retained := root["AtlasPendingStructureEntities"]; retained {
					t.Fatal("village crash replay not acknowledged")
				}
				for _, key := range []string{"structures", "block_entities", "sections"} {
					if !reflect.DeepEqual(root[key], saved[pos][key]) {
						t.Fatal("restart changed complete village Anvil field", key, pos)
					}
				}
			}
			if got, err := reopened.BlockAt(edit[0], edit[1], edit[2]); err != nil || got != StateStone {
				t.Fatal("restart regenerated player's edited village chunk", err)
			}
			afterBytes, err := readGzip(filepath.Join(dir, "level.dat"))
			if err != nil {
				t.Fatal(err)
			}
			_, after, err := nbt.Parse(afterBytes)
			if err != nil || !reflect.DeepEqual(metadata, after) {
				t.Fatal("village generation changed root level.dat tags", err)
			}
			t.Logf("seed1234 %s: saved/reopened %d native chunks, %d pieces, %d seeded loot, %d crops, %d village actors (%d Camel) plus %d nearby structure actors; full Anvil/NBT preserved, crash replay deduplicated, player edit retained", fixture.name, len(fixture.positions), fixture.pieces, loot, crops, villageActors, camels, len(wantActors)-villageActors)
		})
	}
}
