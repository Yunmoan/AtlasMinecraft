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

func TestStructureConcentricRingsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_rings_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type key struct {
		seed int64
		mode string
	}
	expected := make(map[key][]StructureChunkPosition)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 6 || f[0] != "ring" {
			t.Fatalf("invalid ring vector %q", scanner.Text())
		}
		seed, err := strconv.ParseInt(f[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		var values [3]int32
		for index := range values {
			number, err := strconv.ParseInt(f[3+index], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			values[index] = int32(number)
		}
		target := key{seed, f[2]}
		if int(values[0]) != len(expected[target]) {
			t.Fatalf("unexpected ring sequence in %q", scanner.Text())
		}
		expected[target] = append(expected[target], StructureChunkPosition{values[1], values[2]})
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(expected) != 12 {
		t.Fatalf("got %d ring scenarios, want 12", len(expected))
	}
	for target, want := range expected {
		t.Run(fmt.Sprintf("%d/%s", target.seed, target.mode), func(t *testing.T) {
			state, err := LoadStructurePlacementState(target.seed)
			if err != nil {
				t.Fatal(err)
			}
			var got []StructureChunkPosition
			if target.mode == "overworld" {
				err = state.PopulateOverworldStrongholdRings()
				got = state.RingPositions["minecraft:strongholds"]
			} else {
				got, err = GenerateConcentricRingPositions(target.seed, state.Sets["minecraft:strongholds"].Placement, nil, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 128 || len(want) != 128 {
				t.Fatalf("got %d positions, want %d", len(got), len(want))
			}
			for index := range want {
				if got[index] != want[index] {
					t.Fatalf("ring %d: got %v, want %v", index, got[index], want[index])
				}
			}
		})
	}
}
