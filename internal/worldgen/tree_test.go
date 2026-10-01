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
	file, err := os.Open("../../docs/worldgen/tree_feature_1.21.1.txt")
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
	if cases != 600 {
		t.Fatalf("got %d tree cases, want 600", cases)
	}
}

func TestUnsupportedTreeConfigurationsRejected(t *testing.T) {
	for _, name := range []string{"minecraft:oak_bees_0002", "minecraft:dark_oak", "minecraft:mega_spruce", "minecraft:mangrove"} {
		if _, err := LoadTreeConfiguration(name); err == nil {
			t.Fatalf("unsupported tree %s was accepted", name)
		}
	}
}
