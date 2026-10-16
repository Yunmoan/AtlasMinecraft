package worldgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestSaplingSurvivalAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/sapling_survival_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	region := vegetationProbeRegion(t, 0, 0, 0)
	p := FeaturePosition{8, 41, 8}
	scanner := bufio.NewScanner(file)
	cases := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 6 || f[0] != "sapling_survival" {
			t.Fatalf("invalid sapling survival vector: %s", scanner.Text())
		}
		variant, err := strconv.Atoi(f[4])
		if err != nil {
			t.Fatal(err)
		}
		state := SurfaceBlockState{Name: f[1]}
		if variant > 0 {
			property := map[int]string{1: "waterlogged", 2: "hanging"}[variant]
			state.Properties = map[string]string{property: "true"}
		}
		for _, support := range []struct {
			delta int32
			name  string
		}{{-1, f[2]}, {1, f[3]}} {
			if err := region.SetBlock(treeOffset(p, 0, support.delta, 0), SurfaceBlockState{Name: support.name}); err != nil {
				t.Fatal(err)
			}
		}
		want := f[5] == "1"
		got, err := featurePlantSurvives(region, p, state)
		if err != nil || got != want {
			t.Fatalf("survival %v, %v; want %s", got, err, scanner.Text())
		}
		s, err := json.Marshal(state)
		if err != nil {
			t.Fatal(err)
		}
		predicate, err := compileFeaturePredicate(json.RawMessage(fmt.Sprintf(`{"type":"minecraft:would_survive","offset":[-2,-1,3],"state":%s}`, s)))
		if err != nil {
			t.Fatal(err)
		}
		got, err = predicate(region, treeOffset(p, 2, 1, -3))
		if err != nil || got != want {
			t.Fatalf("offset would_survive %v, %v; want %s", got, err, scanner.Text())
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 630 {
		t.Fatalf("got %d JAR vectors, want 630", cases)
	}
	t.Logf("matched %d actual Block.canSurvive vectors and offset placement predicates", cases)
}
