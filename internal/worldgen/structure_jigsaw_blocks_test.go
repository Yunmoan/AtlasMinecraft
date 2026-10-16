package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func jigsawBastionElements(t *testing.T) map[string]JigsawPoolElement {
	t.Helper()
	registry, err := loadJigsawRegistry()
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]JigsawPoolElement)
	for name, pool := range registry.Pools {
		if !strings.HasPrefix(name, "minecraft:bastion/") {
			continue
		}
		for _, entry := range pool.Elements {
			result[entry.Element.Location] = entry.Element
		}
	}
	return result
}

func TestBastionTemplateInventory(t *testing.T) {
	actors := 0
	blockEntities := make(map[string]int)
	elements := jigsawBastionElements(t)
	if len(elements) != 167 {
		t.Fatalf("Bastion original template inventory = %d", len(elements))
	}
	for name, element := range elements {
		template, err := LoadJigsawTemplate(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := loadJigsawProcessors(element.Processors); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, actor := range template.Entities {
			encoded, _ := json.Marshal(jigsawNBTJSON(nbt.Tag{Type: nbt.Compound, Value: actor.NBT}))
			t.Logf("actor %s block=%v precise=%v %s", name, actor.BlockPosition, actor.Position, encoded)
			actors++
		}
		for _, block := range template.Palettes[0] {
			if block.NBT != nil {
				blockEntities[block.State.Name]++
			}
		}
	}
	if actors != 5 {
		t.Fatalf("got %d original Bastion template actors", actors)
	}
	t.Logf("template block entities: %v", blockEntities)
}

func TestJigsawTemplateNBTInventory(t *testing.T) {
	registry, err := loadJigsawRegistry()
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]bool)
	var collect func(JigsawPoolElement)
	collect = func(element JigsawPoolElement) {
		if element.Location != "" {
			names[element.Location] = true
		}
		for _, child := range element.Elements {
			collect(child)
		}
	}
	for _, pool := range registry.Pools {
		for _, entry := range pool.Elements {
			collect(entry.Element)
		}
	}
	blockEntities, actors := make(map[string]int), make(map[string]int)
	for name := range names {
		template, err := LoadJigsawTemplate(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, actor := range template.Entities {
			id, _ := actor.NBT.Text("id")
			actors[id]++
		}
		for _, palette := range template.Palettes {
			for _, block := range palette {
				if block.NBT != nil {
					blockEntities[block.State.Name]++
				}
			}
		}
	}
	t.Logf("templates=%d actors=%v template block NBT=%v", len(names), actors, blockEntities)
}

func TestJigsawBastionBlocksAgainstVanillaJAR(t *testing.T) {
	processors, placements, _ := testJigsawBlocksFixture(t, "structure_jigsaw_blocks_1.21.1.txt", jigsawBastionElements(t), 0)
	if processors != 4008 || placements != 8016 {
		t.Fatalf("got %d processors/%d block placements", processors, placements)
	}
}

func TestJigsawOverworldBlocksAgainstVanillaJAR(t *testing.T) {
	processors, placements, elements := testJigsawBlocksFixture(t, "jigsaw_overworld_blocks_1.21.1.txt", make(map[string]JigsawPoolElement), 1234)
	expectedProcessors := 0
	for _, element := range elements {
		if element.Location == "" {
			continue
		}
		template, err := LoadJigsawTemplate(element.Location)
		if err != nil {
			t.Fatal(err)
		}
		if len(template.Palettes) != 0 {
			expectedProcessors += 24
		}
	}
	if len(elements) < 300 || processors != expectedProcessors || placements != 48*len(elements) {
		t.Fatalf("got %d elements/%d processors/%d placements, expected %d/%d", len(elements), processors, placements, expectedProcessors, 48*len(elements))
	}
}

func TestJigsawOutpostBlocksAgainstVanillaJAR(t *testing.T) {
	processors, placements, elements := testJigsawBlocksFixture(t, "jigsaw_outpost_blocks_1.21.1.txt", make(map[string]JigsawPoolElement), 1234)
	if len(elements) != 12 || processors != 264 || placements != 576 {
		t.Fatalf("got %d Outpost elements/%d processors/%d placements", len(elements), processors, placements)
	}
}

func testJigsawBlocksFixture(t *testing.T, filename string, elements map[string]JigsawPoolElement, seed int64) (int, int, map[string]JigsawPoolElement) {
	t.Helper()
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", filename))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	processors, placements := 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) == 0 {
			continue
		}
		if f[0] == "jigsaw_element" {
			if len(f) != 3 {
				t.Fatalf("invalid element header: %s", scanner.Text())
			}
			encoded, err := base64.StdEncoding.DecodeString(f[2])
			if err != nil {
				t.Fatal(err)
			}
			var element JigsawPoolElement
			if err := json.Unmarshal(encoded, &element); err != nil {
				t.Fatal(err)
			}
			elements[f[1]] = element
			continue
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		element, found := elements[f[1]]
		if !found {
			t.Fatalf("unknown template %s", f[1])
		}
		if f[0] == "jigsaw_processors" {
			template, err := LoadJigsawTemplate(element.Location)
			if err != nil {
				t.Fatal(err)
			}
			if len(f) != 9 {
				t.Fatalf("invalid processor vector: %s", scanner.Text())
			}
			origin := FeaturePosition{int32(parse(4)), int32(parse(5)), int32(parse(6))}
			pivot := FeaturePosition{origin.X + 5, 48, origin.Z - 7}
			world := &jigsawProcessorProbeWorld{ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(3)), edits: make(map[FeaturePosition]SurfaceBlockState)}}, seed: seed}
			blocks, err := ProcessJigsawTemplateBlocks(world, template, element, origin, pivot, TemplateRotation(parse(2)))
			if err != nil {
				t.Fatal(err)
			}
			var rows []string
			for _, block := range blocks {
				id, err := blockstate.BlockStateID(block.State.Name, block.State.Properties)
				if err != nil {
					t.Fatal(err)
				}
				data := "null"
				if block.NBT != nil {
					data = iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: block.NBT})
				}
				rows = append(rows, fmt.Sprintf("%d %d %d %d %s", block.Position.X, block.Position.Y, block.Position.Z, id, data))
			}
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
			if len(blocks) != int(parse(7)) || hash != f[8] {
				t.Fatalf("%s: got %d/%s", scanner.Text(), len(blocks), hash)
			}
			processors++
			continue
		}
		if f[0] != "jigsaw_blocks" || len(f) != 15 {
			t.Fatalf("invalid block vector: %s", scanner.Text())
		}
		origin := FeaturePosition{int32(parse(5)), int32(parse(6)), int32(parse(7))}
		pivot := FeaturePosition{origin.X + 5, 48, origin.Z - 7}
		box, err := jigsawElementBox(element, origin, TemplateRotation(parse(3)))
		if err != nil {
			t.Fatal(err)
		}
		piece, err := (&jigsawPlacedPiece{element: element, origin: origin, rotation: TemplateRotation(parse(3)), box: box}).saved()
		if err != nil {
			t.Fatal(err)
		}
		world := &jigsawProcessorProbeWorld{ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(4)), edits: make(map[FeaturePosition]SurfaceBlockState)}}, seed: seed}
		random := NewWorldgenRandom(parse(2))
		_, err = PostProcessJigsawBlocks(world, &piece, random, int32(parse(8)), int32(parse(9)), pivot)
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		hash, _, err := structureBlockDigest(world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		var rows []string
		for p, entity := range world.entities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
			for name, value := range map[string]int32{"x": p.X, "y": p.Y, "z": p.Z} {
				data[name] = nbt.Tag{Type: nbt.Int, Value: value}
			}
			rows = append(rows, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
		}
		sort.Strings(rows)
		entityHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
		next := int64(random.NextLong())
		if len(world.edits) != int(parse(10)) || hash != f[11] || len(world.entities) != int(parse(12)) || entityHash != f[13] || next != parse(14) {
			t.Fatalf("%s: got %d/%s entities=%d/%s next=%d", scanner.Text(), len(world.edits), hash, len(world.entities), entityHash, next)
		}
		placements++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("matched %d original processor chains and %d original clipped block/NBT/loot/RNG calls", processors, placements)
	return processors, placements, elements
}
