package worldgen

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMansionFloorPlansAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../testdata/worldgen/structure_mansion_grid_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cases := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 608 || fields[0] != "mansion_grid" {
			t.Fatal("invalid mansion grid fixture", len(fields))
		}
		seed, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		random := NewLegacyRandom(seed)
		plan := newMansionFloorPlan(random)
		i := 3
		for gi, g := range []*mansionGrid{&plan.base, &plan.third, &plan.rooms[0], &plan.rooms[1], &plan.rooms[2]} {
			for z := 0; z < 11; z++ {
				for x := 0; x < 11; x++ {
					want, err := strconv.ParseInt(fields[i], 10, 32)
					if err != nil {
						t.Fatal(err)
					}
					i++
					if got := g.get(x, z); got != int32(want) {
						t.Fatalf("seed %d grid %d cell %d,%d: got %d want %d", seed, gi, x, z, got, want)
					}
				}
			}
		}
		want, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if got := int64(random.NextLong()); got != want {
			t.Fatalf("seed %d continuation: got %d want %d", seed, got, want)
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 260 {
		t.Fatal("incomplete reference corpus", cases)
	}
	t.Logf("matched all five grids and RNG continuation over %d original mansion partitions", cases)
}
