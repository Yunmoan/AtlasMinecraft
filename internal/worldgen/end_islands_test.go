package worldgen

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestEndIslandsAgainstVanillaJAR(t *testing.T) {
	f, err := os.Open("../../docs/worldgen/end_islands_1.21.1.txt")
	if err != nil { t.Fatal(err) }
	defer f.Close()
	scan := bufio.NewScanner(f)
	islands := make(map[int64]*EndIslandDensity)
	simplex := make(map[int64]*simplexNoise)
	counts := make(map[string]int)
	for scan.Scan() {
		line := scan.Text()
		var kind string
		var seed int64
		var x, z float64
		var bits uint64
		if _, err := fmt.Sscan(line, &kind, &seed, &x, &z, &bits); err != nil { t.Fatal(err) }
		var got float64
		switch kind {
		case "end_islands":
			if islands[seed] == nil { islands[seed] = NewEndIslandDensity(seed) }
			got = islands[seed].Sample(int32(x), int32(z))
		case "simplex2d":
			if simplex[seed] == nil { simplex[seed] = newSimplexNoise(NewLegacyRandom(seed)) }
			got = simplex[seed].sample(x, z)
		default: t.Fatalf("unknown reference row %s", kind)
		}
		if math.Float64bits(got) != bits {
			t.Fatalf("%s: got %016x (%g), expected %016x (%g)", line, math.Float64bits(got), got, bits, math.Float64frombits(bits))
		}
		counts[kind]++
	}
	if err := scan.Err(); err != nil { t.Fatal(err) }
	if counts["end_islands"] != 5238 || counts["simplex2d"] != 30 { t.Fatalf("reference count %v", counts) }
}

func TestEndIslandsDensityGraph(t *testing.T) {
	settings, err := LoadDensitySettings("end")
	if err != nil { t.Fatal(err) }
	sampler := NewNoiseChunkDensitySampler(1234, settings, 64, -70)
	for _, p := range [][3]int32{{1024, 64, -1120}, {1031, 48, -1109}, {1039, 127, -1105}} {
		v, err := sampler.Sample(settings.Router["final_density"], p[0], p[1], p[2])
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) { t.Fatalf("End final density at %v: %g, %v", p, v, err) }
	}
	islandNode := &DensityNode{Kind: "end_islands"}
	got, err := sampler.Sample(islandNode, 1024, 0, -1120)
	if err != nil || got != NewEndIslandDensity(1234).Sample(1024, -1120) { t.Fatalf("End islands graph dispatch: %g %v", got, err) }
}
