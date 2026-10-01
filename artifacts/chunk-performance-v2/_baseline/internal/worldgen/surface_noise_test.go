package worldgen_test

import (
	"atlas/internal/worldgen"
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestSurfaceSpecialNoisesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "surface_noise_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		var label, name string
		var seed int64
		var x, z int32
		var expected uint64
		if _, err := fmt.Sscan(scanner.Text(), &label, &seed, &name, &x, &z, &expected); err != nil || label != "surface_noise" {
			t.Fatalf("invalid special noise vector %q (%v)", scanner.Text(), err)
		}
		noise, err := worldgen.NewNamedNormalNoise(seed, name)
		if err != nil {
			t.Fatal(err)
		}
		scale := 1.0
		switch name {
		case "badlands_pillar":
			scale = 0.2
		case "badlands_pillar_roof":
			scale = 0.75
		case "iceberg_pillar":
			scale = 1.28
		case "iceberg_pillar_roof":
			scale = 1.17
		}
		got := noise.Noise(float64(x)*scale, 0, float64(z)*scale)
		if math.Float64bits(got) != expected {
			t.Errorf("seed %d %s at %d,%d: bits %d, want %d", seed, name, x, z, math.Float64bits(got), expected)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 144 {
		t.Fatalf("checked %d special noise values, want 144", count)
	}
}
