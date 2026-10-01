package worldgen

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestOverworldBiomeSelectionAgainstVanilla(t *testing.T) {
	points, err := OverworldBiomePoints()
	if err != nil || len(points) != 7593 {
		t.Fatalf("parameter points: %d, %v", len(points), err)
	}
	queries := 0
	scanner := bufio.NewScanner(strings.NewReader(overworldBiomeData))
	for scanner.Scan() {
		var climate [6]int64
		var label, want string
		if !strings.HasPrefix(scanner.Text(), "query ") {
			continue
		}
		line := scanner.Text()
		if _, err := fmt.Sscan(line, &label, &climate[0], &climate[1], &climate[2], &climate[3], &climate[4], &climate[5], &want); err != nil {
			t.Fatalf("vanilla query %q: %v", line, err)
		}
		got, err := SelectOverworldBiome(climate)
		if err != nil || got != want {
			t.Errorf("%v: got %s, want %s (%v)", climate, got, want, err)
		}
		queries++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if queries != 262 {
		t.Fatalf("expected 262 vanilla query vectors, got %d", queries)
	}
}

func TestOverworldBiomeClimatePipeline(t *testing.T) {
	s, err := NewOverworldBiomeSampler(7122209284156038298)
	if err != nil {
		t.Fatal(err)
	}
	name, err := s.BiomeAtQuart(0, 16, 0)
	if err != nil || !strings.HasPrefix(name, "minecraft:") {
		t.Fatalf("biome sample: %q, %v", name, err)
	}
}

func TestOverworldBiomePreviousPointTieAgainstVanilla(t *testing.T) {
	points, err := OverworldBiomePoints()
	if err != nil {
		t.Fatal(err)
	}
	previous := -1
	rows := 0
	scanner := bufio.NewScanner(strings.NewReader(overworldBiomeData))
	for scanner.Scan() {
		if !strings.HasPrefix(scanner.Text(), "sequence ") {
			continue
		}
		var climate [6]int64
		var label, want string
		if _, err := fmt.Sscan(scanner.Text(), &label, &climate[0], &climate[1], &climate[2], &climate[3], &climate[4], &climate[5], &want); err != nil {
			t.Fatal(err)
		}
		previous = nearestBiomePoint(points, climate, previous)
		if got := points[previous].Biome; got != want {
			t.Errorf("sequence %d: got %s, want %s", rows, got, want)
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != 4 {
		t.Fatalf("expected 4 vanilla sequence vectors, got %d", rows)
	}
}
