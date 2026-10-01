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

func TestBiomeTemperaturesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "biome_temperature_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		var label, biome string
		var x, y, z int32
		var expectedBits uint32
		var expectedCold bool
		if _, err := fmt.Sscan(scanner.Text(), &label, &biome, &x, &y, &z, &expectedBits, &expectedCold); err != nil || label != "temperature" {
			t.Fatalf("invalid temperature reference: %q (%v)", scanner.Text(), err)
		}
		value, err := worldgen.BiomeTemperatureAt(biome, x, y, z)
		if err != nil || math.Float32bits(value) != expectedBits {
			t.Errorf("%s at %d,%d,%d: bits %d, want %d (%v)", biome, x, y, z, math.Float32bits(value), expectedBits, err)
		}
		cold, err := worldgen.BiomeColdAt(biome, x, y, z)
		if err != nil || cold != expectedCold {
			t.Errorf("%s at %d,%d,%d: cold %v, want %v (%v)", biome, x, y, z, cold, expectedCold, err)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 640 {
		t.Fatalf("checked %d temperature samples, want 640", count)
	}
}
