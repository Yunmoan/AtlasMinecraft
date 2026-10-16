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
	for _, name := range []string{"large_dripstone", "dripstone_cluster", "pointed_dripstone", "underwater_magma", "glow_lichen", "cave_vines", "lush_caves_ceiling_vegetation", "lush_caves_clay", "lush_caves_vegetation", "spore_blossom", "rooted_azalea_tree", "classic_vines_cave_feature", "lake_lava_underground", "lake_lava_surface", "amethyst_geode", "spring_water", "spring_lava", "spring_lava_frozen", "ice_spike", "freeze_top_layer", "iceberg_packed", "iceberg_blue", "seagrass_river", "seagrass_swamp", "seagrass_warm", "seagrass_normal", "seagrass_deep", "seagrass_cold", "seagrass_deep_cold", "seagrass_deep_warm", "seagrass_simple", "sea_pickle", "kelp_warm", "kelp_cold", "warm_ocean_vegetation", "bamboo", "bamboo_light", "patch_melon", "patch_melon_sparse"} {
		if _, ok := g.plans["minecraft:"+name]; !ok {
			t.Fatalf("validated feature %s is not enabled", name)
		}
	}
}

func TestDecoratedOceanAndJungleBiomesGenerate(t *testing.T) {
	g, err := NewOverworldGenerator(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []featureChunkPosition{{-256, 224}, {-256, 240}, {-232, -72}, {-48, -216}, {-32, -168}} {
		t.Run(fmt.Sprintf("%d,%d", p.chunkX, p.chunkZ), func(t *testing.T) {
			result, err := g.Generate(p.chunkX, p.chunkZ)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range result.SkippedFeatures {
				if name == "minecraft:iceberg_packed" || name == "minecraft:iceberg_blue" || name == "minecraft:warm_ocean_vegetation" || name == "minecraft:kelp_cold" || name == "minecraft:bamboo" || name == "minecraft:patch_melon" {
					t.Fatalf("validated feature skipped: %s", name)
				}
			}
		})
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

func TestDecoratedDeepDarkIncludesSculkVeins(t *testing.T) {
	g, err := NewOverworldGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(-256, 232)
	if err != nil {
		t.Fatal(err)
	}
	hasDeepDark := false
	for _, section := range result.Biomes {
		for _, biome := range section {
			hasDeepDark = hasDeepDark || biome == "minecraft:deep_dark"
		}
	}
	if !hasDeepDark {
		t.Fatal("integration case lost its deep-dark biome")
	}
	veins := 0
	for _, state := range result.Terrain.SurfaceBlocks {
		if state.Name == "minecraft:sculk_vein" {
			veins++
		}
	}
	if veins == 0 {
		t.Fatal("supported sculk veins did not reach the generated chunk")
	}
	for _, name := range result.SkippedFeatures {
		if name == "minecraft:sculk_vein" {
			t.Fatal("sculk-vein placement was skipped")
		}
	}
	t.Logf("deep-dark runtime generation contains %d sculk-vein blocks", veins)
}
