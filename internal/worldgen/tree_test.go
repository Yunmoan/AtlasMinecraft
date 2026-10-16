package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type treeProbeWorld struct {
	changed  map[FeaturePosition]SurfaceBlockState
	origin   FeaturePosition
	scenario int
}

func (w *treeProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if !w.InWorld(p) {
		return SurfaceBlockState{Name: "minecraft:air"}, nil
	}
	if state, exists := w.changed[p]; exists {
		return state, nil
	}
	name := "minecraft:air"
	if p.Y == 40 {
		name = "minecraft:grass_block"
		if w.scenario == 3 {
			name = "minecraft:podzol"
		}
		if w.scenario == 4 {
			name = "minecraft:stone"
		}
	}
	if w.scenario == 5 && p.X == w.origin.X && p.Z == w.origin.Z && p.Y == 44 {
		name = "minecraft:stone"
	}
	if w.scenario == 6 && p.Y >= 41 && p.Y <= 43 {
		name = "minecraft:water"
	}
	if w.scenario == 10 && p.Y >= 41 && p.Y <= 44 {
		return SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "1"}}, nil
	}
	if w.scenario >= 15 && p.Y == 40 {
		name = "minecraft:mud"
	}
	if w.scenario >= 15 && p.Y <= 39 {
		name = "minecraft:dirt"
	}
	if w.scenario == 16 && p.Y >= 41 && p.Y <= 44 {
		name = "minecraft:water"
	}
	if w.scenario == 17 && p.Y >= 41 && p.Y <= 44 {
		return SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "1"}}, nil
	}
	if w.scenario == 18 && p.X == w.origin.X && p.Z == w.origin.Z && p.Y == 42 {
		name = "minecraft:stone"
	}
	if w.scenario == 7 && p.X == w.origin.X+1 && p.Z == w.origin.Z && p.Y >= 43 && p.Y <= 45 {
		name = "minecraft:oak_leaves"
		if p.Y == 43 {
			name = "minecraft:oak_log"
		}
	}
	id, err := blockstate.BlockStateID(name, nil)
	if err != nil {
		return SurfaceBlockState{}, err
	}
	state, err := blockstate.BlockStateByID(id)
	return SurfaceBlockState{Name: state.Name, Properties: state.Properties}, err
}

func (w *treeProbeWorld) SetBlock(p FeaturePosition, state SurfaceBlockState) error {
	w.changed[p] = state
	return nil
}
func (w *treeProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }

func TestTreeFeaturesAgainstVanillaJAR(t *testing.T) {
	testTreeFeaturesAgainstVanillaJAR(t, "tree_feature_1.21.1.txt", 600, false)
}

func TestDecoratedTreeFeaturesAgainstVanillaJAR(t *testing.T) {
	testTreeFeaturesAgainstVanillaJAR(t, "decorated_tree_feature_1.21.1.txt", 840, true)
}

func TestGiantTreeFeaturesAgainstVanillaJAR(t *testing.T) {
	testTreeFeaturesAgainstVanillaJAR(t, "giant_tree_feature_1.21.1.txt", 600, false)
}

func TestBranchingTreeFeaturesAgainstVanillaJAR(t *testing.T) {
	testTreeFeaturesAgainstVanillaJAR(t, "branching_tree_feature_1.21.1.txt", 240, false)
}

func TestCherryTreeFeaturesAgainstVanillaJAR(t *testing.T) {
	testTreeFeaturesAgainstVanillaJAR(t, "cherry_tree_feature_1.21.1.txt", 240, true)
}

func TestMangroveTreeFeaturesAgainstVanillaJAR(t *testing.T) {
	testTreeFeaturesAgainstVanillaJAR(t, "mangrove_tree_feature_1.21.1.txt", 304, true)
}

func testTreeFeaturesAgainstVanillaJAR(t *testing.T, fixture string, expected int, decorated bool) {
	t.Helper()
	file, err := os.Open("../../docs/worldgen/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cases := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 11 || fields[0] != "tree_case" {
			t.Fatalf("invalid tree vector: %s", scanner.Text())
		}
		parse := func(s string) int64 {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		config, err := LoadTreeConfiguration(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		origin := FeaturePosition{int32(parse(fields[4])), int32(parse(fields[5])), int32(parse(fields[6]))}
		if decorated {
			independent := NewLegacyRandom(parse(fields[2]) ^ 0x123456789abcdef0)
			config = config.WithBeehiveShuffle(func(p []FeaturePosition) {
				for i := len(p); i > 1; i-- {
					j := independent.NextIntBound(int32(i))
					p[i-1], p[j] = p[j], p[i-1]
				}
			})
		}
		w := &treeProbeWorld{changed: make(map[FeaturePosition]SurfaceBlockState), origin: origin, scenario: int(parse(fields[3]))}
		random := NewWorldgenRandom(parse(fields[2]))
		placed, err := PlaceTreeFeature(w, random, origin, config)
		if err != nil {
			t.Fatal(err)
		}
		var output []string
		for p, state := range w.changed {
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
			output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
		}
		sort.Strings(output)
		hash := sha256.Sum256([]byte(strings.Join(output, "\n")))
		next := int64(random.NextLong())
		if placed != (parse(fields[7]) != 0) || len(w.changed) != int(parse(fields[8])) || next != parse(fields[9]) || hex.EncodeToString(hash[:]) != fields[10] {
			t.Errorf("%s seed %s scenario %s: placed=%t count=%d next=%d hash=%x; want %s", fields[1], fields[2], fields[3], placed, len(w.changed), next, hash, scanner.Text())
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != expected {
		t.Fatalf("got %d tree cases, want %d", cases, expected)
	}
}

func TestUnsupportedTreeConfigurationsRejected(t *testing.T) {
	for _, name := range []string{"minecraft:crimson_fungus", "minecraft:missing_tree"} {
		if _, err := LoadTreeConfiguration(name); err == nil {
			t.Fatalf("unsupported tree %s was accepted", name)
		}
	}
}

func TestVanillaTreeConfigurationCatalog(t *testing.T) {
	definitions, err := loadFeatureDefinitions()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for name, definition := range definitions.Configured {
		if definition.Type != "minecraft:tree" {
			continue
		}
		if _, err := LoadTreeConfiguration(name); err != nil {
			t.Errorf("vanilla configured tree %s: %v", name, err)
		}
		count++
	}
	if count != 31 {
		t.Fatalf("got %d vanilla tree definitions, want 31", count)
	}
}
