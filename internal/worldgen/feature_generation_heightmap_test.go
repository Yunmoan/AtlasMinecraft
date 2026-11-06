package worldgen

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGenerationHeightmapsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/generation_heightmaps_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(rows) != 56 {
		t.Fatalf("heightmap vectors: %d", len(rows))
	}
	for _, borrowed := range []bool{false, true} {
		for _, firstQuery := range []int{0, 6} {
			t.Run(fmt.Sprintf("borrowed=%t/first-query=%d", borrowed, firstQuery), func(t *testing.T) {
				chunk := &NoiseChunkResult{X: -1, Z: 2, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)}
				for x := int32(0); x < 7; x++ {
					chunk.setSurfaceState(x, 60+x, 0, structureState("stone"))
					chunk.setSurfaceState(x, 63+x, 0, structureState("water"))
				}
				region, err := NewOverworldFeatureRegion(chunk)
				if err != nil {
					t.Fatal(err)
				}
				region.borrowed = map[featureChunkPosition]bool{{-1, 2}: borrowed}
				region.BeginFeatureStage()
				put := func(x, y int32, name string) {
					t.Helper()
					if err := region.SetBlock(FeaturePosition{x - 16, y, 32}, structureState(name)); err != nil {
						t.Fatal(err)
					}
				}
				for row, line := range rows {
					fields := strings.Fields(line)
					if len(fields) != 7 || fields[0] != "generation_heightmap" {
						t.Fatal("invalid heightmap row", line)
					}
					phase, x := row/8, int32(row%8)
					if fields[1] != strconv.Itoa(phase) || fields[2] != strconv.Itoa(int(x)) {
						t.Fatal("heightmap row order", line)
					}
					switch phase {
					case 1:
						put(x, 75+x, "oak_log")
					case 2:
						put(x, 76+x, "oak_leaves")
					case 3:
						put(x, 77+x, "water")
					case 4:
						put(x, 78+x, "short_grass")
					case 5:
						for y := 75 + x; y <= 78+x; y++ {
							put(x, y, "air")
						}
					case 6:
						for y := 60 + x; y <= 63+x; y++ {
							put(x, y, "air")
						}
					}
					if phase < firstQuery {
						continue
					}
					for i, name := range []string{"OCEAN_FLOOR_WG", "WORLD_SURFACE_WG", "OCEAN_FLOOR", "WORLD_SURFACE"} {
						want, err := strconv.ParseInt(fields[3+i], 10, 32)
						if err != nil {
							t.Fatal(err)
						}
						got, err := region.HeightmapAt(name, x-16, 32)
						if err != nil || got != int32(want) {
							t.Fatalf("%s %s: got %d want %d (%v)", line, name, got, want, err)
						}
					}
				}
			})
		}
	}
}
