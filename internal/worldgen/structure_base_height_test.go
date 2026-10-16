package worldgen

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestStructureBaseHeightsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type key struct {
		seed int64
		x, z int32
	}
	expected := make(map[key]int32)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if f[0] != "height" {
			continue
		}
		if len(f) != 8 {
			t.Fatalf("invalid official base-height vector %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		expected[key{parse(1), int32(parse(5)), int32(parse(6))}] = int32(parse(7))
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	generators := make(map[int64]*OverworldNoiseGenerator)
	for position, want := range expected {
		generator := generators[position.seed]
		if generator == nil {
			generator, err = NewOverworldNoiseGenerator(position.seed)
			if err != nil {
				t.Fatal(err)
			}
			generators[position.seed] = generator
		}
		got, err := generator.BaseOccupiedHeightAt(position.x, position.z)
		if err != nil || got != want {
			t.Fatalf("seed %d column %d,%d: got %d (%v), want %d", position.seed, position.x, position.z, got, err, want)
		}
	}
	if len(expected) != 384 {
		t.Fatalf("matched %d base columns, want 384", len(expected))
	}
	t.Logf("matched %d actual NoiseBasedChunkGenerator single-column heights", len(expected))
}
