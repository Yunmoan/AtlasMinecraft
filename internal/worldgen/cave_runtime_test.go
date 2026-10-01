package worldgen

import (
	"fmt"
	"testing"
)

func TestCaveAndFrozenPlansEnabled(t *testing.T) {
	g, err := NewOverworldGenerator(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"large_dripstone", "dripstone_cluster", "pointed_dripstone", "underwater_magma", "glow_lichen", "cave_vines", "lush_caves_ceiling_vegetation", "lush_caves_clay", "lush_caves_vegetation", "spore_blossom", "rooted_azalea_tree", "classic_vines_cave_feature", "lake_lava_underground", "lake_lava_surface", "amethyst_geode", "spring_water", "spring_lava", "spring_lava_frozen", "ice_spike", "freeze_top_layer"} {
		if _, ok := g.plans["minecraft:"+name]; !ok {
			t.Fatalf("validated feature %s is not enabled", name)
		}
	}
}

func TestDecoratedCaveBiomesGenerate(t *testing.T) {
	for _, c := range []struct {
		seed int64
		x, z int32
	}{{1234, -208, -64}, {0, -248, -96}, {1234, -256, -192}, {0, -224, -40}, {1234, -184, -48}, {0, -24, -256}} {
		t.Run(fmt.Sprintf("%d/%d,%d", c.seed, c.x, c.z), func(t *testing.T) {
			g, err := NewOverworldGenerator(c.seed)
			if err != nil {
				t.Fatal(err)
			}
			result, err := g.Generate(c.x, c.z)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range result.SkippedFeatures {
				if name == "minecraft:large_dripstone" || name == "minecraft:dripstone_cluster" || name == "minecraft:pointed_dripstone" || name == "minecraft:glow_lichen" || name == "minecraft:lush_caves_vegetation" {
					t.Fatalf("supported cave feature skipped: %s", name)
				}
			}
		})
	}
}
