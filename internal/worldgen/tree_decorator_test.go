package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type decoratorProbeWorld struct {
	changed, initial map[FeaturePosition]SurfaceBlockState
	entities         map[FeaturePosition]nbt.Object
	origin           FeaturePosition
	scenario         int
}

func probeTreeState(name string) SurfaceBlockState {
	id, err := blockstate.BlockStateID("minecraft:"+name, nil)
	if err != nil {
		panic(err)
	}
	s, err := blockstate.BlockStateByID(id)
	if err != nil {
		panic(err)
	}
	return SurfaceBlockState{Name: s.Name, Properties: s.Properties}
}

func (w *decoratorProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if s, ok := w.changed[p]; ok {
		return s, nil
	}
	if s, ok := w.initial[p]; ok {
		return s, nil
	}
	name := "air"
	if p.Y == 40 {
		name = "grass_block"
		if w.scenario == 1 {
			name = "stone"
		}
	}
	if w.scenario == 2 && p.Y >= 41 && p.Y <= 46 && abs32(p.X-w.origin.X)+abs32(p.Z-w.origin.Z) == 1 {
		name = "stone"
	}
	if w.scenario == 5 && p.Y == 40 && (p.X+p.Z)%3 == 0 {
		name = "air"
	}
	if w.scenario == 5 && p.Y == 38 {
		name = "dirt"
	}
	if w.scenario == 6 && p.Y == 43 && abs32(p.X-w.origin.X) == 3 && abs32(p.Z-w.origin.Z) == 3 {
		name = "vine"
	}
	return probeTreeState(name), nil
}
func (w *decoratorProbeWorld) SetBlock(p FeaturePosition, s SurfaceBlockState) error {
	w.changed[p] = s
	return nil
}
func (w *decoratorProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }
func (w *decoratorProbeWorld) SetBlockEntity(p FeaturePosition, _ string, data nbt.Object) error {
	w.entities[p] = data
	return nil
}

func decoratorReferenceConfig(t *testing.T, kind int) treeDecorator {
	t.Helper()
	names := []string{"minecraft:jungle_tree", "minecraft:jungle_tree", "minecraft:jungle_tree", "minecraft:oak_bees_005", "minecraft:mega_spruce", "minecraft:mangrove"}
	indices := []int{1, 2, 0, 0, 0, 1}
	definition, err := LoadConfiguredFeatureDefinition(names[kind])
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Decorators []json.RawMessage `json:"decorators"`
	}
	if err := json.Unmarshal(definition.Config, &raw); err != nil {
		t.Fatal(err)
	}
	d, err := decodeTreeDecorator(raw.Decorators[indices[kind]])
	if err != nil {
		t.Fatal(err)
	}
	d.probability = 1
	return d
}

func TestTreeDecoratorsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/tree_decorator_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cases := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 11 || fields[0] != "decorator_case" {
			t.Fatalf("invalid decorator vector %q", scanner.Text())
		}
		parse := func(s string) int64 {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		kind, seed, scenario := int(parse(fields[1])), parse(fields[2]), int(parse(fields[3]))
		origin := FeaturePosition{int32(parse(fields[4])), 41, int32(parse(fields[5]))}
		w := &decoratorProbeWorld{changed: make(map[FeaturePosition]SurfaceBlockState), initial: make(map[FeaturePosition]SurfaceBlockState), entities: make(map[FeaturePosition]nbt.Object), origin: origin, scenario: scenario}
		independent := NewLegacyRandom(seed ^ 0x123456789abcdef0)
		config := TreeConfiguration{decorators: []treeDecorator{decoratorReferenceConfig(t, kind)}, beehiveShuffle: func(p []FeaturePosition) {
			for i := len(p); i > 1; i-- {
				j := independent.NextIntBound(int32(i))
				p[i-1], p[j] = p[j], p[i-1]
			}
		}}
		tree := &treePlacement{world: w, random: NewWorldgenRandom(seed), config: config, trunks: make(map[FeaturePosition]bool), leaves: make(map[FeaturePosition]bool)}
		for y := int32(40); y <= 47; y++ {
			p := FeaturePosition{origin.X, y, origin.Z}
			tree.trunks[p] = true
			tree.trunkOrder.add(p)
			name := "oak_log"
			if y == 40 {
				name = "dirt"
			}
			w.initial[p] = probeTreeState(name)
		}
		for y := int32(45); y <= 48; y++ {
			for x := int32(-2); x <= 2; x++ {
				for z := int32(-2); z <= 2; z++ {
					if scenario == 3 || x == 0 && z == 0 && y < 48 || abs32(x) == 2 && abs32(z) == 2 {
						continue
					}
					p := FeaturePosition{origin.X + x, y, origin.Z + z}
					tree.leaves[p] = true
					tree.leafOrder.add(p)
					w.initial[p] = probeTreeState("oak_leaves")
				}
			}
		}
		if scenario == 8 || scenario == 9 {
			tree.roots = make(map[FeaturePosition]bool)
			for x := int32(3); x <= 4; x++ {
				y := int32(40)
				if scenario == 8 {
					y = 38
				}
				p := FeaturePosition{origin.X + x, y, origin.Z}
				tree.roots[p] = true
				tree.rootOrder.add(p)
				w.initial[p] = probeTreeState("mangrove_roots")
			}
		}
		tree.placeDecorators()
		if tree.err != nil {
			t.Fatal(tree.err)
		}
		var output, entities []string
		for p, s := range w.changed {
			id, err := blockstate.BlockStateID(s.Name, s.Properties)
			if err != nil {
				t.Fatal(err)
			}
			output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
		}
		for p, entity := range w.entities {
			var beeOutput []string
			bees, _ := entity["bees"].Elements()
			for _, bee := range bees {
				data, _ := bee.Object()
				entityData, _ := data.Object("entity_data")
				id, _ := entityData.Text("id")
				ticks, _ := data.Number("ticks_in_hive")
				minimum, _ := data.Number("min_ticks_in_hive")
				beeOutput = append(beeOutput, fmt.Sprintf("%s:%d:%d", id, ticks, minimum))
			}
			entities = append(entities, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, strings.Join(beeOutput, ",")))
		}
		sort.Strings(output)
		sort.Strings(entities)
		hash := sha256.Sum256([]byte(strings.Join(output, "\n")))
		entityHash := sha256.Sum256([]byte(strings.Join(entities, "\n")))
		next := int64(tree.random.NextLong())
		if len(w.changed) != int(parse(fields[6])) || next != parse(fields[7]) || hex.EncodeToString(hash[:]) != fields[8] || len(w.entities) != int(parse(fields[9])) || hex.EncodeToString(entityHash[:]) != fields[10] {
			t.Errorf("kind %d seed %d scenario %d: count=%d next=%d hash=%x entities=%d entityHash=%x; want %s", kind, seed, scenario, len(w.changed), next, hash, len(w.entities), entityHash, scanner.Text())
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 420 {
		t.Fatalf("got %d decorators, want 420", cases)
	}
}
