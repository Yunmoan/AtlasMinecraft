package worldgen_test

import (
	"atlas/internal/worldgen"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBiomeZoomAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "biome_zoom_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		var label string
		var seed int64
		var x, y, z, expectedX, expectedY, expectedZ int32
		if _, err := fmt.Sscan(scanner.Text(), &label, &seed, &x, &y, &z,
			&expectedX, &expectedY, &expectedZ); err != nil || label != "zoom" {
			t.Fatalf("invalid zoom reference: %q (%v)", scanner.Text(), err)
		}
		qx, qy, qz := worldgen.BiomeQuartAtBlock(seed, x, y, z)
		if qx != expectedX || qy != expectedY || qz != expectedZ {
			t.Errorf("seed %d at %d,%d,%d: quart %d,%d,%d, want %d,%d,%d",
				seed, x, y, z, qx, qy, qz, expectedX, expectedY, expectedZ)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 60 {
		t.Fatalf("checked %d zoom positions, want 60", count)
	}
}
