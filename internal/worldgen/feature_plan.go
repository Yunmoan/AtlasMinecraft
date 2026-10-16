package worldgen

import (
	"encoding/json"
	"fmt"
)

type configuredFeaturePlacer func(*OverworldFeatureRegion, *WorldgenRandom, FeaturePosition) (bool, error)

// Compile the entire selector graph before enabling a top-level feature.
// Unsupported branches disable that feature, preserving an honest coverage
// report and avoiding distribution changes caused by substituting features.
func compileFeaturePlan(name string, depth int) (configuredFeaturePlacer, error) {
	definition, err := LoadConfiguredFeatureDefinition(name)
	if err != nil {
		return nil, err
	}
	return compileConfiguredFeaturePlan(name, definition, depth)
}

func compileConfiguredFeaturePlan(name string, definition ConfiguredFeatureDefinition, depth int) (configuredFeaturePlacer, error) {
	if depth > 32 {
		return nil, fmt.Errorf("feature nesting is too deep")
	}
	switch definition.Type {
	case "minecraft:fossil":
		return compileFossilPlan(definition)
	case "minecraft:sculk_patch":
		return compileSculkPatchPlan(definition)
	case "minecraft:nether_forest_vegetation":
		return compileNetherForestPlan(definition)
	case "minecraft:huge_fungus":
		return compileHugeFungusPlan(definition)
	case "minecraft:twisting_vines":
		return compileTwistingVinesPlan(definition)
	case "minecraft:weeping_vines":
		return compileWeepingVinesPlan(definition)
	case "minecraft:glowstone_blob":
		return compileGlowstonePlan(definition)
	case "minecraft:basalt_columns":
		return compileBasaltColumnsPlan(definition)
	case "minecraft:basalt_pillar":
		return compileBasaltPillarPlan(definition)
	case "minecraft:delta_feature":
		return compileDeltaPlan(definition)
	case "minecraft:netherrack_replace_blobs":
		return compileReplaceBlobsPlan(definition)
	case "minecraft:scattered_ore":
		return compileScatteredOrePlan(name, definition)
	case "minecraft:end_island":
		return compileEndIslandPlan(definition)
	case "minecraft:end_platform":
		return compileEndPlatformPlan(definition)
	case "minecraft:end_gateway":
		return compileEndGatewayPlan(definition)
	case "minecraft:chorus_plant":
		return compileChorusPlantPlan(definition)
	case "minecraft:end_spike":
		return compileEndSpikePlan(definition)
	case "minecraft:monster_room":
		return PlaceMonsterRoomFeature, nil
	case "minecraft:desert_well":
		return func(region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) (bool, error) {
			return PlaceDesertWellFeature(region, r, p)
		}, nil
	case "minecraft:bamboo":
		return compileBambooPlan(definition)
	case "minecraft:block_pile":
		return compileBlockPilePlan(definition)
	case "minecraft:iceberg":
		return compileIcebergPlan(definition)
	case "minecraft:seagrass", "minecraft:sea_pickle", "minecraft:kelp", "minecraft:coral_tree", "minecraft:coral_claw", "minecraft:coral_mushroom":
		return compileOceanPlan(definition)
	case "minecraft:geode":
		return compileGeodePlan(definition)
	case "minecraft:lake":
		return compileLakePlan(definition)
	case "minecraft:vines":
		return func(region *OverworldFeatureRegion, _ *WorldgenRandom, p FeaturePosition) (bool, error) {
			return PlaceVinesFeature(region, p)
		}, nil
	case "minecraft:root_system":
		return compileRootSystemPlan(definition, depth)
	case "minecraft:vegetation_patch", "minecraft:waterlogged_vegetation_patch":
		return compileVegetationPatchPlan(definition, depth)
	case "minecraft:multiface_growth":
		return compileMultifaceGrowthPlan(definition)
	case "minecraft:dripstone_cluster":
		return compileDripstoneClusterPlan(definition)
	case "minecraft:large_dripstone":
		return compileLargeDripstonePlan(definition)
	case "minecraft:pointed_dripstone":
		return compilePointedDripstonePlan(definition)
	case "minecraft:underwater_magma":
		return compileUnderwaterMagmaPlan(definition)
	case "minecraft:simple_random_selector":
		var config struct {
			Features []json.RawMessage `json:"features"`
		}
		if err := json.Unmarshal(definition.Config, &config); err != nil {
			return nil, err
		}
		if len(config.Features) == 0 {
			return nil, fmt.Errorf("empty random selector")
		}
		placers := make([]configuredFeaturePlacer, len(config.Features))
		for i, feature := range config.Features {
			var err error
			placers[i], err = compileNestedFeaturePlacement(feature, depth+1)
			if err != nil {
				return nil, err
			}
		}
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			return placers[random.NextIntBound(int32(len(placers)))](region, random, p)
		}, nil
	case "minecraft:ice_spike":
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			return PlaceIceSpikeFeature(region, random, p)
		}, nil
	case "minecraft:freeze_top_layer":
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			return PlaceSnowAndFreezeFeature(region, p)
		}, nil
	case "minecraft:spring_feature":
		config, err := LoadSpringConfiguration(name)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			return PlaceSpringFeature(region, p, config)
		}, nil
	case "minecraft:huge_brown_mushroom", "minecraft:huge_red_mushroom":
		return compileHugeMushroomPlan(definition)
	case "minecraft:random_boolean_selector":
		var config struct {
			True  json.RawMessage `json:"feature_true"`
			False json.RawMessage `json:"feature_false"`
		}
		if err := json.Unmarshal(definition.Config, &config); err != nil {
			return nil, err
		}
		yes, err := compileNestedFeaturePlacement(config.True, depth+1)
		if err != nil {
			return nil, err
		}
		no, err := compileNestedFeaturePlacement(config.False, depth+1)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			if random.NextBoolean() {
				return yes(region, random, p)
			}
			return no(region, random, p)
		}, nil
	case "minecraft:block_column":
		return compileBlockColumnPlan(definition)
	case "minecraft:tree":
		config, err := LoadTreeConfiguration(name)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			invocation := config
			if config.HasBeehiveDecorator() {
				if region.treeShuffle == nil {
					return false, fmt.Errorf("tree %s requires frozen independent beehive shuffle for neighboring decoration replay", name)
				}
				invocation = config.WithBeehiveShuffle(region.treeShuffle.forTree(name, p, random))
			}
			return PlaceTreeFeature(region, random, p, invocation)
		}, nil
	case "minecraft:random_selector":
		var selector struct {
			Features []struct {
				Chance  float32         `json:"chance"`
				Feature json.RawMessage `json:"feature"`
			} `json:"features"`
			Default json.RawMessage `json:"default"`
		}
		if err := json.Unmarshal(definition.Config, &selector); err != nil {
			return nil, err
		}
		var placers []configuredFeaturePlacer
		for _, entry := range selector.Features {
			if entry.Chance < 0 || entry.Chance > 1 {
				return nil, fmt.Errorf("invalid tree selector chance")
			}
			placer, err := compileNestedFeaturePlacement(entry.Feature, depth+1)
			if err != nil {
				return nil, err
			}
			placers = append(placers, placer)
		}
		fallback, err := compileNestedFeaturePlacement(selector.Default, depth+1)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			for i, entry := range selector.Features {
				if random.NextFloat() < entry.Chance {
					return placers[i](region, random, p)
				}
			}
			return fallback(region, random, p)
		}, nil
	default:
		return compileVegetationPlan(definition, depth)
	}
}

func compileNestedFeaturePlacement(raw json.RawMessage, depth int) (configuredFeaturePlacer, error) {
	if depth > 32 {
		return nil, fmt.Errorf("feature nesting is too deep")
	}
	var name string
	var definition PlacedFeatureDefinition
	var err error
	var placer configuredFeaturePlacer
	if json.Unmarshal(raw, &name) == nil {
		definition, err = LoadPlacedFeatureDefinition(name)
		if err == nil {
			placer, err = compileFeaturePlan(definition.Feature, depth)
		}
	} else {
		var inline struct {
			Feature   json.RawMessage   `json:"feature"`
			Placement []json.RawMessage `json:"placement"`
		}
		if err = json.Unmarshal(raw, &inline); err == nil {
			definition.Placement = inline.Placement
			if json.Unmarshal(inline.Feature, &name) == nil {
				definition.Feature = name
				placer, err = compileFeaturePlan(name, depth)
			} else {
				var configured ConfiguredFeatureDefinition
				if err = json.Unmarshal(inline.Feature, &configured); err == nil {
					definition.Feature = configured.Type
					placer, err = compileConfiguredFeaturePlan("", configured, depth)
				}
			}
		}
	}
	if err != nil {
		return nil, err
	}
	// Validate region-dependent operations too, without executing predicates.
	template, err := compilePlacementTemplate(definition, nil)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
		operations, err := template.bind(region, nil)
		if err != nil {
			return false, err
		}
		success := false
		err = walkPlacement(operations, random, p, func(position FeaturePosition, random *WorldgenRandom) error {
			placed, err := placer(region, random, position)
			success = success || placed
			return err
		})
		return success, err
	}, nil
}
