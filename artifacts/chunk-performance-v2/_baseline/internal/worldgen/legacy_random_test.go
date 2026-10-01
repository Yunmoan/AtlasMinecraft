package worldgen

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLegacyRandomMatchesVanilla1211(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "legacy_random_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	longs := make(map[int64]*LegacyRandom)
	bounds := make(map[int64]*LegacyRandom)
	fractions := make(map[int64]*LegacyRandom)
	checks := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 3 {
			t.Fatalf("invalid fixture row %q", scanner.Text())
		}
		seed, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		want, err := strconv.ParseUint(parts[len(parts)-1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		var got uint64
		switch parts[0] {
		case "long":
			if longs[seed] == nil {
				longs[seed] = NewLegacyRandom(seed)
			}
			got = longs[seed].NextLong()
		case "bound":
			if bounds[seed] == nil {
				bounds[seed] = NewLegacyRandom(seed)
			}
			bound, err := strconv.ParseInt(parts[2], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			got = uint64(bounds[seed].NextIntBound(int32(bound)))
		case "float":
			if fractions[seed] == nil {
				fractions[seed] = NewLegacyRandom(seed)
			}
			got = uint64(math.Float32bits(fractions[seed].NextFloat()))
		case "double":
			if fractions[seed] == nil {
				fractions[seed] = NewLegacyRandom(seed)
			}
			got = math.Float64bits(fractions[seed].NextDouble())
		case "position":
			got = NewLegacyRandom(seed).ForkPositional().At(-17, 68, 33).NextLong()
		case "hash":
			got = NewLegacyRandom(seed).ForkPositional().FromHashOf("minecraft:terrain").NextLong()
		default:
			t.Fatalf("unknown fixture row %q", scanner.Text())
		}
		if got != want {
			t.Fatalf("%s: got %d, vanilla %d", scanner.Text(), got, want)
		}
		checks++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if checks != 56 {
		t.Fatalf("got %d reference checks, want 56", checks)
	}
}
