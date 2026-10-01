package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestOverworldPreBiomeCandidatesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/feature_placement_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type key struct {
		seed        int64
		chunkX      int32
		chunkZ      int32
		step, index int
		name        string
	}
	expected := make(map[key][]FeaturePosition)
	type summary struct {
		count int
		next  int64
	}
	counts := make(map[key]summary)
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 10 && len(fields) != 9 {
			t.Fatalf("invalid JAR placement vector: %q", scanner.Text())
		}
		parse := func(field string) int64 {
			value, err := strconv.ParseInt(field, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		feature := key{parse(fields[1]), int32(parse(fields[2])), int32(parse(fields[3])),
			int(parse(fields[4])), int(parse(fields[5])), fields[6]}
		switch fields[0] {
		case "candidate_count":
			if len(fields) != 9 {
				t.Fatalf("invalid JAR count vector: %q", scanner.Text())
			}
			counts[feature] = summary{int(parse(fields[7])), parse(fields[8])}
		case "candidate":
			if len(fields) != 10 {
				t.Fatalf("invalid JAR candidate vector: %q", scanner.Text())
			}
			position := FeaturePosition{int32(parse(fields[7])), int32(parse(fields[8])), int32(parse(fields[9]))}
			expected[feature] = append(expected[feature], position)
			count++
		default:
			t.Fatalf("unknown JAR placement vector: %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 992 || len(counts) != 60 {
		t.Fatalf("got %d candidates in %d groups, want 992 in 60", count, len(counts))
	}
	for target, want := range counts {
		t.Run(fmt.Sprintf("%s/%d/%d,%d", target.name, target.seed, target.chunkX, target.chunkZ), func(t *testing.T) {
			feature := PlacedFeature{Step: target.step, Index: target.index, Name: target.name}
			got, random, err := overworldPreBiomeCandidates(target.seed, target.chunkX, target.chunkZ, feature, nil)
			if err != nil {
				t.Fatal(err)
			}
			wantPositions := expected[target]
			if len(got) != want.count || len(wantPositions) != want.count {
				t.Fatalf("got %d candidates, want %d; JAR listed %d", len(got), want.count, len(wantPositions))
			}
			if next := int64(random.NextLong()); next != want.next {
				t.Fatalf("next random long: got %d, want %d", next, want.next)
			}
			for index := range wantPositions {
				if got[index] != wantPositions[index] {
					t.Fatalf("candidate %d: got %+v, want %+v", index, got[index], wantPositions[index])
				}
			}
		})
	}
}

func TestOverworldPreBiomeCandidatesRejectsUnsupportedModifier(t *testing.T) {
	feature := PlacedFeature{Step: 9, Index: 0, Name: "minecraft:glow_lichen"}
	if _, err := OverworldPreBiomeCandidates(0, 0, 0, feature); err == nil {
		t.Fatal("unsupported modifier should not be silently ignored")
	}
}

func TestOverworldFeaturePositionsBiomeMembership(t *testing.T) {
	feature := PlacedFeature{Step: 6, Index: 33, Name: "minecraft:ore_emerald"}
	stonyPeaks := func(int32, int32, int32) (string, error) { return "minecraft:stony_peaks", nil }
	got, err := OverworldFeaturePositions(0, 0, 0, feature, stonyPeaks)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 {
		t.Fatalf("stony peaks has %d emerald candidates, want 100", len(got))
	}
	plains := func(int32, int32, int32) (string, error) { return "minecraft:plains", nil }
	got, err = OverworldFeaturePositions(0, 0, 0, feature, plains)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("plains has %d emerald candidates, want 0", len(got))
	}
}
