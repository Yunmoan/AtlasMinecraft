package worldgen

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestSparseDecorationMatchesPrivateChunkAcrossBorder(t *testing.T) {
	left, right := newFlatForestRockChunk(-1, 0), newFlatForestRockChunk(0, 0)
	region, _ := NewOverworldFeatureRegion(left, right)
	region.borrowed = map[featureChunkPosition]bool{{-1, 0}: true, {0, 0}: true}
	private, _ := NewOverworldFeatureRegion(cloneTerrain(left), cloneTerrain(right))
	beforeLeft, beforeRight := cloneTerrain(left), cloneTerrain(right)
	random := rand.New(rand.NewSource(767))
	states := []SurfaceBlockState{{Name: "minecraft:air"}, {Name: "minecraft:water"}, {Name: "minecraft:stone"},
		{Name: "minecraft:oak_leaves", Properties: map[string]string{"distance": "1", "persistent": "false", "waterlogged": "false"}}}
	for i := 0; i < 200; i++ {
		p := FeaturePosition{int32(random.Intn(4) - 2), int32(random.Intn(110) - 64), int32(random.Intn(16))}
		state := states[random.Intn(len(states))]
		for _, r := range []*OverworldFeatureRegion{region, private} {
			if err := r.SetBlock(p, state); err != nil {
				t.Fatal(err)
			}
			if i%5 == 0 {
				if err := r.MarkPostProcessing(p); err != nil {
					t.Fatal(err)
				}
			}
		}
		for _, kind := range []string{"OCEAN_FLOOR", "WORLD_SURFACE", "MOTION_BLOCKING"} {
			got, err := region.HeightmapAt(kind, p.X, p.Z)
			want, otherErr := private.HeightmapAt(kind, p.X, p.Z)
			if err != nil || otherErr != nil || got != want {
				t.Fatalf("%s overlay=%d private=%d errors=%v/%v", kind, got, want, err, otherErr)
			}
		}
	}
	for _, x := range []int32{-1, 0} {
		got := region.cloneChunk(x, 0)
		want := private.chunks[featureChunkPosition{x, 0}]
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("sparse decoration changed chunk %d,0", x)
		}
		for _, state := range got.SurfaceBlocks {
			if state.Properties != nil {
				state.Properties["poison"] = "true"
			}
		}
	}
	if !reflect.DeepEqual(cloneTerrain(left), beforeLeft) || !reflect.DeepEqual(cloneTerrain(right), beforeRight) {
		t.Fatal("decoration mutated borrowed terrain")
	}
	for _, state := range states {
		if state.Properties["poison"] != "" {
			t.Fatal("returned state properties alias a feature template")
		}
	}
}
