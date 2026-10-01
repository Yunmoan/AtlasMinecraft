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

func TestNamedNoiseMatchesVanilla1211(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "named_noise_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	points := [][3]float64{{0, 0, 0}, {-17.125, 68.875, 33.5},
		{30000000.5, -63.25, -29999999.75}}
	noises := make(map[string]*NormalNoise)
	checks := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 5 || parts[0] != "named" {
			t.Fatalf("invalid fixture row %q", scanner.Text())
		}
		seed, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		point, err := strconv.Atoi(parts[3])
		if err != nil || point < 0 || point >= len(points) {
			t.Fatalf("invalid point %q", parts[3])
		}
		want, err := strconv.ParseUint(parts[4], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		key := parts[1] + ":" + parts[2]
		n := noises[key]
		if n == nil {
			n, err = NewNamedNormalNoise(seed, parts[2])
			if err != nil {
				t.Fatal(err)
			}
			noises[key] = n
		}
		p := points[point]
		got := n.Noise(p[0], p[1], p[2])
		if math.Float64bits(got) != want {
			t.Fatalf("%s: got bits %d, vanilla %d", scanner.Text(), math.Float64bits(got), want)
		}
		checks++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if checks != 720 {
		t.Fatalf("got %d reference checks, want 720", checks)
	}
}
