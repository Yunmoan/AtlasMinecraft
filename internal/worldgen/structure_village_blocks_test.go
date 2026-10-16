package worldgen

import (
	"atlas/internal/nbt"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestJigsawCampfireDefaultsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/village_campfire_defaults_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, row := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(row)
		if len(fields) != 3 || fields[0] != "campfire_default" {
			t.Fatalf("invalid native campfire default: %s", row)
		}
		encoded, err := base64.StdEncoding.DecodeString(fields[2])
		if err != nil {
			t.Fatal(err)
		}
		expected, n, err := nbt.ParseNetwork(encoded)
		if err != nil || n != len(encoded) {
			t.Fatalf("invalid native default NBT: %v", err)
		}
		got := nbt.Tag{Type: nbt.Compound, Value: jigsawDefaultCampfireBlockEntity()}
		if iglooCanonicalNBT(t, got) != iglooCanonicalNBT(t, expected) {
			t.Fatalf("%s constructor/save defaults differ", fields[1])
		}
		checked++
	}
	if checked != 2 {
		t.Fatalf("matched %d campfire defaults, want 2", checked)
	}
}

func TestJigsawVillageCampfireReplacementRecordsEntity(t *testing.T) {
	chunk := &NoiseChunkResult{X: 0, Z: 0, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
	region, err := NewOverworldFeatureRegion(chunk)
	if err != nil {
		t.Fatal(err)
	}
	result := &structureBlockResult{blocks: make(map[FeaturePosition]SurfaceBlockState), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
	world := &structureRecordingWorld{region: region, result: result}
	origin := FeaturePosition{1, 64, 1}
	element := JigsawPoolElement{Type: "minecraft:legacy_single_pool_element", Location: "minecraft:village/taiga/taiga_decoration_5", Projection: "rigid", Processors: json.RawMessage(`{"processors":[]}`)}
	box, err := jigsawElementBox(element, origin, TemplateRotationNone)
	if err != nil {
		t.Fatal(err)
	}
	piece, err := (&jigsawPlacedPiece{element: element, origin: origin, rotation: TemplateRotationNone, box: box}).saved()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PostProcessJigsawBlocks(world, &piece, NewWorldgenRandom(1234), 0, 0, origin); err != nil {
		t.Fatal(err)
	}
	entity, ok := result.blockEntities[origin]
	if !ok || entity.block != "minecraft:campfire" {
		t.Fatal("Jigsaw replacement omitted its default campfire block entity")
	}
	// A later same-type placement must not reset saved cooking progress.
	entity.data["CookingTimes"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{12, 0, 0, 0}}
	if err := world.SetBlockEntity(origin, entity.block, entity.data); err != nil {
		t.Fatal(err)
	}
	if _, err := PostProcessJigsawBlocks(world, &piece, NewWorldgenRandom(1234), 0, 0, origin); err != nil {
		t.Fatal(err)
	}
	if got := result.blockEntities[origin].data["CookingTimes"].Value.([]int32)[0]; got != 12 {
		t.Fatalf("same campfire placement reset cooking progress to %d", got)
	}
}

func TestVillageTemplateAssetsAndProcessors(t *testing.T) {
	registry, err := loadJigsawRegistry()
	if err != nil {
		t.Fatal(err)
	}
	names, processors := make(map[string]bool), make(map[string]bool)
	features, blockEntities := make(map[string]bool), make(map[string]int)
	pools := 0
	for name, pool := range registry.Pools {
		if !strings.HasPrefix(name, "minecraft:village/") {
			continue
		}
		pools++
		for _, weighted := range pool.Elements {
			element := weighted.Element
			if element.Type == "minecraft:feature_pool_element" {
				features[string(element.Feature)] = true
				continue
			}
			if element.Type == "minecraft:empty_pool_element" {
				continue
			}
			if element.Type != "minecraft:legacy_single_pool_element" {
				t.Fatalf("unverified village element %s", element.Type)
			}
			if _, err := loadJigsawProcessors(element.Processors); err != nil {
				t.Fatalf("%s: %v", element.Location, err)
			}
			var named string
			if json.Unmarshal(element.Processors, &named) == nil {
				processors[named] = true
			}
			names[element.Location] = true
		}
	}
	for name := range names {
		template, err := LoadJigsawTemplate(name)
		if err != nil || len(template.Palettes) == 0 {
			t.Fatalf("missing native village template %s: %v", name, err)
		}
		for _, palette := range template.Palettes {
			for _, block := range palette {
				if block.NBT == nil || block.State.Name == "minecraft:jigsaw" || block.State.Name == "minecraft:structure_block" {
					continue
				}
				if _, err := jigsawPlacedBlockEntity(block.State, cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object), NewWorldgenRandom(1234)); err != nil {
					t.Fatalf("unsupported village block-entity payload %s/%s: %v", name, block.State.Name, err)
				}
				blockEntities[block.State.Name]++
			}
		}
	}
	if pools != 62 || len(names) != 478 || len(processors) != 16 || len(features) != 13 {
		t.Fatalf("incomplete village assets: %d pools, %d templates, %d processor lists, %d feature elements", pools, len(names), len(processors), len(features))
	}
	t.Logf("verified %d pools, %d templates, %d processors, %d feature elements; original template block NBT=%v", pools, len(names), len(processors), len(features), blockEntities)
}

// All original legacy templates and processor combinations, including five
// biome families, zombie variants, farms, slopes and terrain-matching streets.
func TestJigsawVillageBlocksAgainstVanillaJAR(t *testing.T) {
	processors, placements, elements := testJigsawBlocksFixture(t, "jigsaw_village_blocks_1.21.1.txt", make(map[string]JigsawPoolElement), 1234)
	families := make(map[string]int)
	projected, common := 0, 0
	for _, element := range elements {
		parts := strings.Split(element.Location, "/")
		if len(parts) < 3 || parts[0] != "minecraft:village" {
			t.Fatalf("non-village element %s", element.Location)
		}
		if parts[1] == "common" {
			common++
		} else {
			families[parts[1]]++
		}
		if element.Projection == "terrain_matching" {
			projected++
		}
	}
	if len(families) != 5 || common != 21 || projected != 173 || len(elements) != 573 || processors != 18336 || placements != 36672 {
		t.Fatalf("village coverage: %d elements, %d processors, %d placements, %d projected, families=%v", len(elements), processors, placements, projected, families)
	}
	t.Logf("all five village families %v and %d common elements; %d terrain-matching elements", families, common, projected)
}
