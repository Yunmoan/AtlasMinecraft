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

func TestOverworldCarverStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "carver_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type target struct {
		seed int64
		x, z int32
	}
	expected := make(map[target][]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 9 || fields[0] != "start" || fields[7] != "1" {
			t.Fatalf("invalid carver reference: %q", scanner.Text())
		}
		seed, seedErr := strconv.ParseInt(fields[1], 10, 64)
		x, xErr := strconv.ParseInt(fields[2], 10, 32)
		z, zErr := strconv.ParseInt(fields[3], 10, 32)
		if seedErr != nil || xErr != nil || zErr != nil {
			t.Fatalf("invalid carver target: %q", scanner.Text())
		}
		key := target{seed, int32(x), int32(z)}
		expected[key] = append(expected[key], scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(expected) != 9 {
		t.Fatalf("got %d carver targets, want 9", len(expected))
	}
	for key, want := range expected {
		t.Run(fmt.Sprintf("%d/%d,%d", key.seed, key.x, key.z), func(t *testing.T) {
			var got []string
			for _, start := range OverworldCarverStarts(key.seed, key.x, key.z) {
				random := carverRandom(key.seed, start.SourceX, start.SourceZ, start.Index)
				random.NextFloat()
				got = append(got, fmt.Sprintf("start %d %d %d %d %d %d 1 %d", key.seed, key.x, key.z,
					start.SourceX, start.SourceZ, start.Index, int64(random.NextLong())))
			}
			if len(got) != len(want) {
				t.Fatalf("got %d carver starts, want %d", len(got), len(want))
			}
			for index := range want {
				if got[index] != want[index] {
					t.Fatalf("carver start %d: got %s, want %s", index, got[index], want[index])
				}
			}
		})
	}
}
