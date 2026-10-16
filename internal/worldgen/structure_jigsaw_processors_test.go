package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type jigsawProcessorProbeWorld struct {
	*ruinedPortalProbeWorld
	seed int64
}

func (world *jigsawProcessorProbeWorld) WorldSeed() int64 { return world.seed }

func originalJigsawSyntheticPalette(t *testing.T) []StructureTemplateBlock {
	t.Helper()
	registry, err := loadJigsawRegistry()
	if err != nil {
		t.Fatal(err)
	}
	states := make(map[string]SurfaceBlockState)
	var collect func(any)
	collect = func(value any) {
		switch value := value.(type) {
		case []any:
			for _, item := range value {
				collect(item)
			}
		case map[string]any:
			if name, ok := value["Name"].(string); ok {
				encoded, _ := json.Marshal(value)
				state := SurfaceBlockState{Name: name}
				if properties, ok := value["Properties"].(map[string]any); ok {
					state.Properties = make(map[string]string)
					for key, item := range properties {
						state.Properties[key] = item.(string)
					}
				}
				states[string(encoded)] = state
			}
			if name, ok := value["block"].(string); ok {
				encoded, _ := json.Marshal(map[string]string{"Name": name})
				states[string(encoded)] = SurfaceBlockState{Name: name}
			}
			for _, item := range value {
				collect(item)
			}
		}
	}
	for _, encoded := range registry.Processors {
		var value any
		if err := json.Unmarshal(encoded, &value); err != nil {
			t.Fatal(err)
		}
		collect(value)
	}
	for _, name := range []string{"gravel", "white_wool", "red_wool", "white_carpet", "red_carpet", "stone", "bedrock", "obsidian", "air"} {
		name = "minecraft:" + name
		encoded, _ := json.Marshal(map[string]string{"Name": name})
		states[string(encoded)] = SurfaceBlockState{Name: name}
	}
	var names []string
	for name := range states {
		names = append(names, name)
	}
	sort.Strings(names)
	var result []StructureTemplateBlock
	index := 0
	for _, name := range names {
		for repeat := 0; repeat < 4; repeat++ {
			block := StructureTemplateBlock{Position: FeaturePosition{int32(index % 12), int32(index / 144 * 3), int32(index / 12 % 12)}, State: states[name]}
			if index%7 == 0 {
				block.NBT = nbt.Object{"atlas_marker": {Type: nbt.Int, Value: int32(index)}}
			}
			result = append(result, block)
			index++
		}
	}
	return result
}

func TestJigsawAllProcessorListsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "jigsaw_processors_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	palette := originalJigsawSyntheticPalette(t)
	template := &StructureTemplate{Size: FeaturePosition{12, 24, 12}, Palettes: [][]StructureTemplateBlock{palette}}
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 10 || f[0] != "jigsaw_rules" {
			t.Fatalf("invalid Jigsaw processor vector %q", scanner.Text())
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		origin := FeaturePosition{int32(parse(5)), int32(parse(6)), int32(parse(7))}
		pivot := FeaturePosition{origin.X + 5, 48, origin.Z - 7}
		encoded, _ := json.Marshal(f[1])
		element := JigsawPoolElement{Type: "minecraft:single_pool_element", Projection: "rigid", Processors: encoded}
		world := &jigsawProcessorProbeWorld{seed: parse(2), ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(4)), edits: make(map[FeaturePosition]SurfaceBlockState)}}}
		blocks, err := ProcessJigsawTemplateBlocks(world, template, element, origin, pivot, TemplateRotation(parse(3)))
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
		if len(blocks) != int(parse(8)) || hash != f[9] {
			t.Fatalf("%s: got %d/%s (synthetic=%d)", scanner.Text(), len(blocks), hash, len(palette))
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 4320 {
		t.Fatalf("got %d original processor list vectors", count)
	}
	t.Logf("matched %d actual original Rule/TagMatch/BlockRot/Protected/Capped/AppendLoot calls across all 40 processor lists", count)
}
