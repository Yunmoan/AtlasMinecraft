package worldgen

import (
	"math/rand"
	"testing"
)

func TestBiomeIndexMatchesExhaustiveSearch(t *testing.T) {
	points, err := OverworldBiomePoints()
	if err != nil {
		t.Fatal(err)
	}
	index, err := overworldBiomeIndex()
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewSource(76121))
	previous := -1
	for n := 0; n < 12000; n++ {
		var climate [6]int64
		for axis := range climate {
			climate[axis] = int64(random.Intn(40001) - 20000)
		}
		// Exercise interval boundaries and duplicate leaves as well as random
		// queries, checking point identity, not just the displayed biome name.
		if n < len(points) {
			for axis, bounds := range points[n].Ranges {
				climate[axis] = bounds[n%2]
			}
		}
		if n%3 == 0 {
			previous = random.Intn(len(points))
		}
		want := nearestBiomePoint(points, climate, previous)
		got := index.nearest(points, climate, previous)
		if got != want {
			t.Fatalf("query %d climate=%v previous=%d: got %d, want %d", n, climate, previous, got, want)
		}
		previous = got
	}
}

func TestBiomeIndexEqualDistancePreference(t *testing.T) {
	points := make([]BiomeParameterPoint, 32)
	for i := range points {
		points[i].Ranges[0] = [2]int64{int64(i%2*2 - 1), int64(i%2*2 - 1)}
	}
	indices := make([]int, len(points))
	for i := range indices {
		indices[i] = i
	}
	index := buildBiomeIndex(points, indices)
	for _, previous := range []int{-1, 0, 1, 23, 31, 32} {
		if got, want := index.nearest(points, [6]int64{}, previous), nearestBiomePoint(points, [6]int64{}, previous); got != want {
			t.Fatalf("tie with previous %d: got %d, want %d", previous, got, want)
		}
	}
}
