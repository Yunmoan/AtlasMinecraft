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

func TestImprovedNoiseMatchesVanilla1211(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "improved_noise_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	points := [][3]float64{
		{0, 0, 0}, {1.25, -2.75, 4.5}, {-17.125, 68.875, 33.5},
		{30000000.5, -63.25, -29999999.75}, {0.001, 0.999, -0.001},
	}
	noises := make(map[int64]*ImprovedNoise)
	field := map[string]func(*ImprovedNoise) float64{
		"a": func(n *ImprovedNoise) float64 { return n.XOffset },
		"b": func(n *ImprovedNoise) float64 { return n.YOffset },
		"c": func(n *ImprovedNoise) float64 { return n.ZOffset },
	}
	checks := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 4 {
			t.Fatalf("invalid fixture row %q", scanner.Text())
		}
		seed, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		want, err := strconv.ParseUint(parts[3], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		n := noises[seed]
		if n == nil {
			n = NewImprovedNoise(NewXoroshiro(seed))
			noises[seed] = n
		}
		var got float64
		switch parts[0] {
		case "offset":
			get, ok := field[parts[2]]
			if !ok {
				t.Fatalf("unknown offset %q", parts[2])
			}
			got = get(n)
		case "sample", "shifted":
			i, err := strconv.Atoi(parts[2])
			if err != nil || i < 0 || i >= len(points) {
				t.Fatalf("bad point %q", parts[2])
			}
			p := points[i]
			if parts[0] == "sample" {
				got = n.Noise(p[0], p[1], p[2])
			} else {
				got = n.NoiseShifted(p[0], p[1], p[2], 0.25, 0.5)
			}
		default:
			t.Fatalf("unknown fixture row %q", scanner.Text())
		}
		if math.Float64bits(got) != want {
			t.Fatalf("%s: got bits %d, vanilla %d (got %.17g, vanilla %.17g)", scanner.Text(), math.Float64bits(got), want, got, math.Float64frombits(want))
		}
		checks++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if checks != 52 {
		t.Fatalf("got %d reference checks, want 52", checks)
	}
}
