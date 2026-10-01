package worldgen

import (
	"math/rand"
	"testing"
)

func TestFeatureHeightmapsRemainCorrectDuringUndergroundAndSurfaceWrites(t *testing.T) {
	chunk := newFlatForestRockChunk(-1, -1)
	region, err := NewOverworldFeatureRegion(chunk)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewSource(767))
	states := []SurfaceBlockState{
		{Name: "minecraft:air"}, {Name: "minecraft:stone"},
		{Name: "minecraft:water"}, {Name: "minecraft:short_grass"},
		{Name: "minecraft:oak_leaves"}, {Name: "minecraft:copper_ore"},
		{Name: "minecraft:ladder", Properties: map[string]string{"waterlogged": "true"}},
		{Name: "minecraft:ladder", Properties: map[string]string{"waterlogged": "false"}},
	}
	heightmaps := []string{"OCEAN_FLOOR", "WORLD_SURFACE", "MOTION_BLOCKING"}
	for step := 0; step < 180; step++ {
		if step > 0 {
			y := int32(random.Intn(95) - 64)
			if step%3 == 0 {
				y = int32(39 + random.Intn(12))
			}
			if err := region.SetBlock(FeaturePosition{-1, y, -1}, states[random.Intn(len(states))]); err != nil {
				t.Fatal(err)
			}
		}
		for _, kind := range heightmaps {
			want := chunk.MinY
			for y := chunk.MinY + chunk.Height - 1; y >= chunk.MinY; y-- {
				state := chunk.surfaceState(15, y, 15)
				matches := !surfaceAir(state)
				var err error
				switch kind {
				case "OCEAN_FLOOR":
					matches, err = oceanFloorSolid(state)
				case "MOTION_BLOCKING":
					matches, err = motionBlocking(state)
				}
				if err != nil {
					t.Fatal(err)
				}
				if matches {
					want = y + 1
					break
				}
			}
			got, err := region.HeightmapAt(kind, -1, -1)
			if err != nil || got != want {
				t.Fatalf("step %d %s: got %d, want %d, error=%v", step, kind, got, want, err)
			}
		}
	}
}
