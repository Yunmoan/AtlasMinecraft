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
			return PlaceTreeFeature(region, random, p, config)
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
	if _, err := bindPlacement(nil, definition, nil, nil); err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
		operations, err := bindPlacement(region, definition, nil, nil)
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
