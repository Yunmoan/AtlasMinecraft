package worldgen

import (
	"encoding/json"
	"fmt"
)

type positionConsumer func(FeaturePosition) error
type placementOperation func(FeaturePosition, *WorldgenRandom, positionConsumer) error

// compilePlacement builds a lazy modifier pipeline. Placement and feature
// execution share one random stream in vanilla: collecting every candidate
// before placing blocks changes all subsequent candidates.
func compilePlacement(feature PlacedFeature) ([]placementOperation, error) {
	definition, err := LoadPlacedFeatureDefinition(feature.Name)
	if err != nil {
		return nil, err
	}
	return compilePlacementDefinition(feature.Name, definition)
}

func compilePlacementDefinition(name string, definition PlacedFeatureDefinition) ([]placementOperation, error) {
	var operations []placementOperation
	for _, raw := range definition.Placement {
		var kind struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &kind); err != nil {
			return nil, err
		}
		switch kind.Type {
		case "minecraft:random_offset":
			var modifier struct {
				XZ json.RawMessage `json:"xz_spread"`
				Y  json.RawMessage `json:"y_spread"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil {
				return nil, err
			}
			xz, err := decodeFeatureIntProvider(modifier.XZ, -16, 16)
			if err != nil {
				return nil, err
			}
			y, err := decodeFeatureIntProvider(modifier.Y, -384, 384)
			if err != nil {
				return nil, err
			}
			operations = append(operations, func(p FeaturePosition, random *WorldgenRandom, next positionConsumer) error {
				dx, dy, dz := xz(random), y(random), xz(random)
				return next(treeOffset(p, dx, dy, dz))
			})
		case "minecraft:count":
			var modifier struct {
				Count json.RawMessage `json:"count"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil {
				return nil, err
			}
			provider, err := decodeFeatureIntProvider(modifier.Count, 0, 256)
			if err != nil {
				return nil, fmt.Errorf("count for %s: %w", name, err)
			}
			operations = append(operations, func(position FeaturePosition, random *WorldgenRandom, next positionConsumer) error {
				n := provider(random)
				for range n {
					if err := next(position); err != nil {
						return err
					}
				}
				return nil
			})
		case "minecraft:noise_threshold_count":
			var modifier struct {
				Level float64 `json:"noise_level"`
				Below int32   `json:"below_noise"`
				Above int32   `json:"above_noise"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil {
				return nil, err
			}
			if modifier.Below < 0 || modifier.Below > 256 || modifier.Above < 0 || modifier.Above > 256 {
				return nil, fmt.Errorf("invalid noise threshold count for %s", name)
			}
			// Biome.BIOME_INFO_NOISE is a single unshifted simplex octave.
			noise := newSimplexNoise(NewLegacyRandom(2345))
			operations = append(operations, func(p FeaturePosition, random *WorldgenRandom, next positionConsumer) error {
				count := modifier.Above
				if noise.sample(float64(p.X)/200, float64(p.Z)/200) < modifier.Level {
					count = modifier.Below
				}
				for range count {
					if err := next(p); err != nil {
						return err
					}
				}
				return nil
			})
		case "minecraft:in_square":
			operations = append(operations, func(position FeaturePosition, random *WorldgenRandom, next positionConsumer) error {
				position.X += random.NextIntBound(16)
				position.Z += random.NextIntBound(16)
				return next(position)
			})
		case "minecraft:rarity_filter":
			var modifier struct {
				Chance int32 `json:"chance"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil || modifier.Chance <= 0 {
				return nil, fmt.Errorf("invalid rarity for %s", name)
			}
			operations = append(operations, func(position FeaturePosition, random *WorldgenRandom, next positionConsumer) error {
				if random.NextFloat() < 1/float32(modifier.Chance) {
					return next(position)
				}
				return nil
			})
		case "minecraft:height_range":
			var modifier struct {
				Height struct {
					Type    string       `json:"type"`
					Min     heightAnchor `json:"min_inclusive"`
					Max     heightAnchor `json:"max_inclusive"`
					Plateau int32        `json:"plateau"`
				} `json:"height"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil {
				return nil, err
			}
			lower, err := modifier.Height.Min.resolve(-64, 384)
			if err != nil {
				return nil, err
			}
			upper, err := modifier.Height.Max.resolve(-64, 384)
			if err != nil {
				return nil, err
			}
			if modifier.Height.Type != "minecraft:uniform" && modifier.Height.Type != "minecraft:trapezoid" {
				return nil, fmt.Errorf("unsupported height for %s", name)
			}
			operations = append(operations, func(position FeaturePosition, random *WorldgenRandom, next positionConsumer) error {
				if lower > upper {
					position.Y = lower
				} else if modifier.Height.Type == "minecraft:uniform" || modifier.Height.Plateau >= upper-lower {
					position.Y = randomBetweenInclusive(random, lower, upper)
				} else {
					first := (upper - lower - modifier.Height.Plateau) / 2
					position.Y = lower + randomBetweenInclusive(random, 0, upper-lower-first) + randomBetweenInclusive(random, 0, first)
				}
				return next(position)
			})
		case "minecraft:biome", "minecraft:heightmap", "minecraft:block_predicate_filter", "minecraft:surface_water_depth_filter", "minecraft:environment_scan":
			// These need the region or biome source and are bound during execution.
			operations = append(operations, nil)
		default:
			return nil, fmt.Errorf("unsupported placement modifier %s for %s", kind.Type, name)
		}
	}
	return operations, nil
}

func visitFeaturePositions(region *OverworldFeatureRegion, seed int64, cx, cz int32, feature PlacedFeature,
	biomeAtQuart func(int32, int32, int32) (string, error), place func(FeaturePosition, *WorldgenRandom) error) error {
	definition, err := LoadPlacedFeatureDefinition(feature.Name)
	if err != nil {
		return err
	}
	operations, err := bindPlacement(region, definition, &feature, biomeAtQuart)
	if err != nil {
		return err
	}
	random := OverworldPlacedFeatureRandom(OverworldDecorationSeed(seed, cx, cz), feature)
	return walkPlacement(operations, random, FeaturePosition{X: cx * 16, Z: cz * 16}, place)
}

func bindPlacement(region *OverworldFeatureRegion, definition PlacedFeatureDefinition, feature *PlacedFeature,
	biomeAtQuart func(int32, int32, int32) (string, error)) ([]placementOperation, error) {
	name := definition.Feature
	if feature != nil {
		name = feature.Name
	}
	operations, err := compilePlacementDefinition(name, definition)
	if err != nil {
		return nil, err
	}
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		return nil, err
	}
	for i, operation := range operations {
		if operation != nil {
			continue
		}
		var modifier struct {
			Type       string          `json:"type"`
			Heightmap  string          `json:"heightmap"`
			Predicate  json.RawMessage `json:"predicate"`
			WaterDepth int32           `json:"max_water_depth"`
			Direction  string          `json:"direction_of_search"`
			MaxSteps   int32           `json:"max_steps"`
			Target     json.RawMessage `json:"target_condition"`
			Allowed    json.RawMessage `json:"allowed_search_condition"`
		}
		if err := json.Unmarshal(definition.Placement[i], &modifier); err != nil {
			return nil, err
		}
		switch modifier.Type {
		case "minecraft:biome":
			if feature == nil || biomeAtQuart == nil {
				return nil, fmt.Errorf("biome filter without top-level feature context")
			}
			operations[i] = func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
				biome, err := biomeAtQuart(position.X>>2, position.Y>>2, position.Z>>2)
				if err != nil {
					return err
				}
				for _, index := range order.biomes[biome][feature.Step] {
					if index == feature.Index {
						return next(position)
					}
				}
				return nil
			}
		case "minecraft:heightmap":
			operations[i] = func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
				height, err := region.HeightmapAt(modifier.Heightmap, position.X, position.Z)
				if err != nil {
					return err
				}
				if height > -64 {
					position.Y = height
					return next(position)
				}
				return nil
			}
		case "minecraft:block_predicate_filter":
			predicate, err := compileFeaturePredicate(modifier.Predicate)
			if err != nil {
				return nil, err
			}
			operations[i] = func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
				matches, err := predicate(region, position)
				if err != nil {
					return err
				}
				if matches {
					return next(position)
				}
				return nil
			}
		case "minecraft:surface_water_depth_filter":
			operations[i] = func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
				floor, err := region.HeightmapAt("OCEAN_FLOOR", position.X, position.Z)
				if err != nil {
					return err
				}
				surface, err := region.HeightmapAt("WORLD_SURFACE", position.X, position.Z)
				if err != nil {
					return err
				}
				if surface-floor <= modifier.WaterDepth {
					return next(position)
				}
				return nil
			}
		case "minecraft:environment_scan":
			if modifier.MaxSteps < 1 || modifier.MaxSteps > 32 || (modifier.Direction != "up" && modifier.Direction != "down") {
				return nil, fmt.Errorf("unsupported environment scan for %s", name)
			}
			target, err := compileFeaturePredicate(modifier.Target)
			if err != nil {
				return nil, err
			}
			allowed := featurePredicate(func(FeatureBlockWorld, FeaturePosition) (bool, error) { return true, nil })
			if len(modifier.Allowed) != 0 {
				allowed, err = compileFeaturePredicate(modifier.Allowed)
				if err != nil {
					return nil, err
				}
			}
			operations[i] = func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
				ok, err := allowed(region, position)
				if err != nil || !ok {
					return err
				}
				delta := int32(1)
				if modifier.Direction == "down" {
					delta = -1
				}
				for range modifier.MaxSteps {
					ok, err = target(region, position)
					if err != nil {
						return err
					}
					if ok {
						return next(position)
					}
					position.Y += delta
					if position.Y < -64 || position.Y >= 320 {
						return nil
					}
					ok, err = allowed(region, position)
					if err != nil {
						return err
					}
					if !ok {
						break
					}
				}
				ok, err = target(region, position)
				if err != nil {
					return err
				}
				if ok {
					return next(position)
				}
				return nil
			}
		}
	}
	return operations, nil
}

func walkPlacement(operations []placementOperation, random *WorldgenRandom, origin FeaturePosition,
	place func(FeaturePosition, *WorldgenRandom) error) error {
	var visit func(int, FeaturePosition) error
	visit = func(index int, position FeaturePosition) error {
		if index == len(operations) {
			return place(position, random)
		}
		return operations[index](position, random, func(next FeaturePosition) error { return visit(index+1, next) })
	}
	return visit(0, origin)
}
