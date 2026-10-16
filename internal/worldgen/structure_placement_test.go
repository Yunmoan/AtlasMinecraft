package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestStructurePlacementsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_placement_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	states := make(map[int64]*StructurePlacementState)
	counts := make(map[string]int)
	trueCount := make(map[string]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) < 6 {
			t.Fatalf("invalid structure vector %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatalf("invalid integer in %q: %v", scanner.Text(), err)
			}
			return value
		}
		seed := parse(1)
		state := states[seed]
		if state == nil {
			state, err = LoadStructurePlacementState(seed)
			if err != nil {
				t.Fatal(err)
			}
			states[seed] = state
		}
		counts[f[0]]++
		switch f[0] {
		case "potential":
			if len(f) != 7 {
				t.Fatalf("invalid potential vector %q", scanner.Text())
			}
			p := state.Sets[f[2]].Placement
			got, err := p.PotentialChunk(seed, int32(parse(3)), int32(parse(4)))
			want := StructureChunkPosition{int32(parse(5)), int32(parse(6))}
			if err != nil || got != want {
				t.Fatalf("%s: got %v (%v), want %v", scanner.Text(), got, err, want)
			}
		case "placement":
			if len(f) != 6 {
				t.Fatalf("invalid placement vector %q", scanner.Text())
			}
			got, err := state.IsStructureChunk(f[2], int32(parse(3)), int32(parse(4)))
			want := parse(5) == 1
			if err != nil || got != want {
				t.Fatalf("%s: got %v (%v), want %v", scanner.Text(), got, err, want)
			}
			if want {
				trueCount[f[2]]++
			}
		case "frequency":
			if len(f) != 8 {
				t.Fatalf("invalid frequency vector %q", scanner.Text())
			}
			frequency, err := strconv.ParseFloat(f[4], 32)
			if err != nil {
				t.Fatal(err)
			}
			p := StructurePlacement{Type: "minecraft:random_spread", Spacing: 34, Separation: 8,
				Frequency: float32(frequency), Salt: int32(parse(3)), FrequencyReductionMethod: f[2]}
			got, err := p.AllowsFrequency(seed, int32(parse(5)), int32(parse(6)))
			want := parse(7) == 1
			if err != nil || got != want {
				t.Fatalf("%s: got %v (%v), want %v", scanner.Text(), got, err, want)
			}
		default:
			t.Fatalf("unknown structure vector %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if counts["potential"] != 1254 || counts["placement"] != 52554 || counts["frequency"] != 1584 {
		t.Fatalf("unexpected vector coverage %v", counts)
	}
	for name := range states[0].Sets {
		if name != "minecraft:strongholds" && trueCount[name] == 0 {
			t.Errorf("no accepted placement tested for %s", name)
		}
	}
	t.Logf("matched %d official structure vectors", counts["potential"]+counts["placement"]+counts["frequency"])
}

func TestStructurePlacementPrerequisiteBoundaries(t *testing.T) {
	state, err := LoadStructurePlacementState(1234)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := state.IsStructureChunk("strongholds", 0, 0); err == nil {
		t.Fatal("strongholds accepted without biome-relocated ring positions")
	}
	state.RingPositions["minecraft:strongholds"] = []StructureChunkPosition{{19, -27}}
	if got, err := state.IsStructureChunk("strongholds", 19, -27); err != nil || !got {
		t.Fatalf("supplied stronghold rings: got %v, %v", got, err)
	}
	if got, err := state.IsStructureChunk("strongholds", 0, 0); err != nil || got {
		t.Fatalf("outside supplied stronghold rings: got %v, %v", got, err)
	}
	treasure := state.Sets["minecraft:buried_treasures"].Placement
	if got, want := treasure.LocatePosition(StructureChunkPosition{-2, 3}), (FeaturePosition{-23, 0, 57}); got != want {
		t.Fatalf("locate offset got %v, want %v", got, want)
	}
	for name, set := range state.Sets {
		for _, entry := range set.Structures {
			biomes, err := StructureBiomeNames(entry.Structure)
			if err != nil || len(biomes) == 0 {
				t.Fatalf("%s/%s biomes: %v, %v", name, entry.Structure, biomes, err)
			}
		}
	}
	invalid := []StructurePlacement{
		{Type: "minecraft:random_spread", Spacing: 8, Separation: 8, Frequency: 1},
		{Type: "minecraft:random_spread", Spacing: 8, Separation: -1, Frequency: 1},
		{Type: "minecraft:random_spread", Spacing: 8, Frequency: 1, SpreadType: "unknown"},
		{Type: "minecraft:random_spread", Spacing: 8, Frequency: 1, FrequencyReductionMethod: "unknown"},
	}
	for index, p := range invalid {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			if _, err := p.PotentialChunk(0, 0, 0); err == nil {
				t.Fatal("invalid structure placement accepted")
			}
		})
	}
}
