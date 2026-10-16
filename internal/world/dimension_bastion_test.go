package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

// Exercise production creation, chunk loading, actor realization, both Anvil
// directories, crash-window pending replay, and an actual close/reopen.
func TestDimensionBastionAnvilRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	metadata, err := os.ReadFile(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	biome, err := worldgen.NewDimensionBiomeSampler(seed, "nether")
	if err != nil {
		t.Fatal(err)
	}
	start, err := worldgen.GenerateJigsawStructureStart(seed, -26, -27, "minecraft:bastion_remnant", worldgen.JigsawStartSources{MinY: 0, Height: 256, GenerationDepth: 128, BiomeAtQuart: biome.BiomeAtQuart})
	if err != nil || start == nil || len(start.Pieces) != 88 {
		t.Fatal("pinned accepted Bastion changed", err)
	}
	positions := map[[2]int32]bool{{-26, -27}: true}
	var chest *worldgen.FeaturePosition
	authoredActors := 0
	for _, piece := range start.Pieces {
		element, _ := piece.Data.Object("pool_element")
		location, _ := element.Text("location")
		template, err := worldgen.LoadJigsawTemplate(location)
		if err != nil {
			t.Fatal(err)
		}
		x, _ := piece.Data.Number("PosX")
		y, _ := piece.Data.Number("PosY")
		z, _ := piece.Data.Number("PosZ")
		text, _ := piece.Data.Text("rotation")
		rotation, ok := map[string]worldgen.TemplateRotation{"NONE": worldgen.TemplateRotationNone, "CLOCKWISE_90": worldgen.TemplateRotationClockwise90, "CLOCKWISE_180": worldgen.TemplateRotationClockwise180, "COUNTERCLOCKWISE_90": worldgen.TemplateRotationCounterclockwise90}[text]
		if !ok {
			t.Fatalf("unexpected Bastion rotation %s", text)
		}
		transform := func(p worldgen.FeaturePosition) worldgen.FeaturePosition {
			p, _ = worldgen.TransformTemplatePosition(p, worldgen.FeaturePosition{}, worldgen.TemplateMirrorNone, rotation)
			return worldgen.FeaturePosition{X: p.X + int32(x), Y: p.Y + int32(y), Z: p.Z + int32(z)}
		}
		for _, block := range template.Palettes[0] {
			if block.State.Name == "minecraft:chest" && chest == nil {
				p := transform(block.Position)
				chest = &p
				positions[[2]int32{p.X >> 4, p.Z >> 4}] = true
			}
		}
		for _, actor := range template.Entities {
			p := transform(actor.BlockPosition)
			positions[[2]int32{p.X >> 4, p.Z >> 4}] = true
			authoredActors++
		}
	}
	if chest == nil || authoredActors == 0 {
		t.Fatal("Bastion lacks authored chest/actors")
	}
	w, err := OpenDimension(noViewers{}, 4, dir, "nether")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	want := make(map[[2]int32][]WorldEntity)
	pending := make(map[[2]int32]nbt.Tag)
	var clips [][2]int32
	for p := range positions {
		clips = append(clips, p)
		generated, err := w.store.vanilla.dimensionGenerator.Generate(p[0], p[1])
		if err != nil {
			t.Fatal(err)
		}
		chunk := newChunkBounds(p[0], p[1], 0, 256, false)
		if err := chunk.initializeAnvilRoot(); err != nil {
			t.Fatal(err)
		}
		installGeneratedStructures(chunk, &worldgen.GeneratedOverworldChunk{Structures: generated.Structures, PendingEntities: generated.PendingEntities})
		root, _ := chunk.anvilRoot.Object()
		pending[p] = root["AtlasPendingStructureEntities"]
		requests, _ := pending[p].Elements()
		for _, request := range requests {
			data, _ := request.Object()
			entity, supported, err := realizeGeneratedEntity(seed, "minecraft:the_nether", data)
			if err != nil || !supported {
				t.Fatal("Bastion actor unsupported by runtime", err)
			}
			want[p] = append(want[p], entity)
		}
	}
	errs := make([]error, len(clips))
	var wg sync.WaitGroup
	for i, p := range clips {
		wg.Add(1)
		go func(i int, p [2]int32) { defer wg.Done(); _, errs[i] = w.LoadChunk(p[0], p[1]) }(i, p)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	foundActors := 0
	for _, p := range clips {
		entities, err := w.EntitiesInChunk(p[0], p[1])
		if err != nil || !reflect.DeepEqual(entities, want[p]) && !(len(entities) == 0 && len(want[p]) == 0) {
			t.Fatalf("Bastion actor realization mismatch at %v: %v", p, err)
		}
		for _, entity := range entities {
			switch entity.Type {
			case "minecraft:piglin", "minecraft:piglin_brute", "minecraft:hoglin":
				persistent, _ := entity.Data.Number("PersistenceRequired")
				if persistent != 1 {
					t.Fatal("Bastion actor lost authored persistence")
				}
				foundActors++
			}
		}
	}
	if foundActors != authoredActors {
		t.Fatalf("realized %d Bastion actors, want %d", foundActors, authoredActors)
	}
	if err := w.SetBlock(-26*16+8, 200, -27*16+8, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	saved := make(map[[2]int32]nbt.Object)
	for _, p := range clips {
		_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "DIM-1", "region"), p[0], p[1])
		if err != nil || !exists {
			t.Fatal("Bastion chunk did not persist", err)
		}
		root, _ := tag.Object()
		saved[p] = root
		structures, _ := root.Object("structures")
		references, _ := structures.Object("References")
		if references["minecraft:bastion_remnant"].Type != nbt.LongArray {
			t.Fatalf("Bastion references missing at %v", p)
		}
		if _, ok := root["AtlasPendingStructureEntities"]; ok {
			t.Fatal("Bastion actor requests remained pending")
		}
		if len(want[p]) > 0 {
			_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "DIM-1", "entities"), p[0], p[1])
			if err != nil || !exists {
				t.Fatal("Bastion entity region missing", err)
			}
			root, _ := tag.Object()
			entities, _ := root["Entities"].Elements()
			if len(entities) != len(want[p]) {
				t.Fatal("saved Bastion actor count wrong")
			}
		}
	}
	structures, _ := saved[[2]int32{-26, -27}].Object("structures")
	starts, _ := structures.Object("starts")
	bastion, _ := starts.Object("minecraft:bastion_remnant")
	children, _ := bastion["Children"].Elements()
	if len(children) != len(start.Pieces) {
		t.Fatal("saved Bastion piece list incomplete")
	}
	blockEntities, _ := saved[[2]int32{chest.X >> 4, chest.Z >> 4}]["block_entities"].Elements()
	foundChest := false
	for _, tag := range blockEntities {
		entity, _ := tag.Object()
		x, _ := entity.Number("x")
		y, _ := entity.Number("y")
		z, _ := entity.Number("z")
		if x == int64(chest.X) && y == int64(chest.Y) && z == int64(chest.Z) {
			loot, _ := entity.Text("LootTable")
			foundChest = len(loot) > 0
		}
	}
	if !foundChest {
		t.Fatal("saved Bastion chest loot missing")
	}
	// Simulate a crash after the entity-region write but before the block
	// region clears its pending requests. UUID merge must prevent duplication.
	for p, requests := range pending {
		list, _ := requests.Elements()
		if len(list) == 0 {
			continue
		}
		_, tag, _, err := anvil.ReadChunk(filepath.Join(dir, "DIM-1", "region"), p[0], p[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := tag.Object()
		root["AtlasPendingStructureEntities"] = requests
		if err := w.store.vanilla.regions.WriteChunk(filepath.Join(dir, "DIM-1", "region"), p[0], p[1], "", tag); err != nil {
			t.Fatal(err)
		}
	}
	r, err := OpenDimension(noViewers{}, 4, dir, "nether")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, p := range clips {
		entities, err := r.EntitiesInChunk(p[0], p[1])
		if err != nil || !reflect.DeepEqual(entities, loadedEntitySnapshots(want[p])) && !(len(entities) == 0 && len(want[p]) == 0) {
			t.Fatalf("Bastion actors duplicated/changed after restart at %v: %v", p, err)
		}
		chunk, err := r.store.vanilla.loadBlockChunk(p[0], p[1])
		if err != nil {
			t.Fatal(err)
		}
		root, _ := chunk.anvilRoot.Object()
		if !reflect.DeepEqual(root["structures"], saved[p]["structures"]) || !reflect.DeepEqual(root["block_entities"], saved[p]["block_entities"]) {
			t.Fatal("restart changed Bastion structures/block entities")
		}
		if _, ok := root["AtlasPendingStructureEntities"]; ok {
			t.Fatal("pending replay was not acknowledged")
		}
	}
	id, err := r.BlockAt(-26*16+8, 200, -27*16+8)
	if err != nil || id != StateStone {
		t.Fatal("restart regenerated edited Bastion chunk", err)
	}
	after, err := os.ReadFile(filepath.Join(dir, "level.dat"))
	if err != nil || !bytes.Equal(metadata, after) {
		t.Fatal("Bastion dimension modified shared level.dat", err)
	}
	t.Logf("saved/reopened %d Bastion clips, %d authored actors, chest loot, complete starts/references; crash-window replay preserved actor UUIDs", len(clips), foundActors)
}
