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

func TestNormalNoiseMatchesVanilla1211(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "normal_noise_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	firstOctaves := []int{-4, -2, -10}
	amplitudes := [][]float64{{1, 1, 0, 1, 1}, {0.5, 0, 1}, {1, 1, 1, 1, 1}}
	points := [][3]float64{
		{0, 0, 0}, {1.25, -2.75, 4.5}, {-17.125, 68.875, 33.5},
		{30000000.5, -63.25, -29999999.75}, {0.001, 0.999, -0.001},
	}
	noises := make(map[[2]int64]*NormalNoise)
	checks := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 5 || parts[0] != "normal" {
			t.Fatalf("invalid fixture row %q", scanner.Text())
		}
		seed, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		c, err := strconv.Atoi(parts[2])
		if err != nil || c < 0 || c >= len(amplitudes) {
			t.Fatalf("invalid case %q", parts[2])
		}
		i, err := strconv.Atoi(parts[3])
		if err != nil || i < 0 || i >= len(points) {
			t.Fatalf("invalid point %q", parts[3])
		}
		want, err := strconv.ParseUint(parts[4], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		key := [2]int64{seed, int64(c)}
		noise := noises[key]
		if noise == nil {
			noise, err = NewNormalNoise(NewXoroshiro(seed), firstOctaves[c], amplitudes[c])
			if err != nil {
				t.Fatal(err)
			}
			noises[key] = noise
		}
		p := points[i]
		got := noise.Noise(p[0], p[1], p[2])
		if math.Float64bits(got) != want {
			t.Fatalf("%s: got bits %d, vanilla %d (got %.17g, vanilla %.17g)", scanner.Text(), math.Float64bits(got), want, got, math.Float64frombits(want))
		}
		checks++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if checks != 60 {
		t.Fatalf("got %d reference checks, want 60", checks)
	}
}
