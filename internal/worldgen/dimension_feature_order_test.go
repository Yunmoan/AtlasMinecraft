package worldgen

import (
	"bufio"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestDimensionFeatureScheduleAgainstVanillaJAR(t *testing.T) {
	type key struct{ dimension, group string }
	expected := make(map[key][]PlacedFeature)
	scan := bufio.NewScanner(strings.NewReader(dimensionFeatureData))
	for scan.Scan() {
		f := strings.Fields(scan.Text())
		if len(f) == 0 || f[0] != "schedule" {
			continue
		}
		if len(f) != 6 {
			t.Fatal(scan.Text())
		}
		step, err := strconv.Atoi(f[3])
		if err != nil {
			t.Fatal(err)
		}
		index, err := strconv.Atoi(f[4])
		if err != nil {
			t.Fatal(err)
		}
		k := key{f[1], f[2]}
		expected[k] = append(expected[k], PlacedFeature{step, index, f[5]})
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if len(expected) != 10 {
		t.Fatalf("got %d nonempty schedule groups, want 10", len(expected))
	}
	for k, want := range expected {
		got, err := DimensionFeatureSchedule(k.dimension, strings.Split(k.group, ","))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s %s got %v, want %v", k.dimension, k.group, got, want)
		}
	}
	for _, dimension := range []string{"nether", "end"} {
		if _, err := DimensionFeatureSchedule(dimension, []string{"minecraft:plains"}); err == nil {
			t.Fatal("Overworld biome accepted in " + dimension)
		}
	}
}
