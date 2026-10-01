package worldgen

import (
	"encoding/json"
	"fmt"
)

type positionConsumer func(FeaturePosition) error
type placementOperation func(FeaturePosition, *WorldgenRandom, positionConsumer) error

type placementBinding func(*OverworldFeatureRegion, func(int32, int32, int32) (string, error)) placementOperation

type placementTemplate struct {
	operations []placementOperation
	bindings   []placementBinding
	needsBiome bool
}

func (template *placementTemplate) bind(region *OverworldFeatureRegion, biomeAtQuart func(int32, int32, int32) (string, error)) ([]placementOperation, error) {
	if template.needsBiome && biomeAtQuart == nil {
		return nil, fmt.Errorf("biome filter without top-level feature context")
	}
	if len(template.bindings) == 0 {
		return template.operations, nil
	}
	operations := append([]placementOperation(nil), template.operations...)
	for i, binding := range template.bindings {
		if binding != nil {
			operations[i] = binding(region, biomeAtQuart)
		}
	}
	return operations, nil
}

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
				Height json.RawMessage `json:"height"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil {
				return nil, err
			}
			height, err := decodeFeatureHeightProvider(modifier.Height)
			if err != nil {
				return nil, err
			}
			operations = append(operations, func(position FeaturePosition, random *WorldgenRandom, next positionConsumer) error {
				position.Y = height(random)
				return next(position)
			})
		case "minecraft:biome", "minecraft:heightmap", "minecraft:block_predicate_filter", "minecraft:surface_water_depth_filter", "minecraft:environment_scan", "minecraft:surface_relative_threshold_filter":
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
	template, err := compilePlacementTemplate(definition, feature)
	if err != nil {
		return nil, err
	}
	return template.bind(region, biomeAtQuart)
}

// Compile providers and predicates once, then bind only the private region and
// biome callback for each replay. Templates capture no mutable random stream.
func compilePlacementTemplate(definition PlacedFeatureDefinition, feature *PlacedFeature) (*placementTemplate, error) {
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
	template := &placementTemplate{operations: operations}
	var bindings []placementBinding
	for i, operation := range operations {
		if operation != nil {
			continue
		}
		if bindings == nil {
			bindings = make([]placementBinding, len(operations))
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
			Min        *int32          `json:"min_inclusive"`
			Max        *int32          `json:"max_inclusive"`
		}
		if err := json.Unmarshal(definition.Placement[i], &modifier); err != nil {
			return nil, err
		}
		switch modifier.Type {
		case "minecraft:surface_relative_threshold_filter":
			lower, upper := int32(-2147483648), int32(2147483647)
			if modifier.Min != nil {
				lower = *modifier.Min
			}
			if modifier.Max != nil {
				upper = *modifier.Max
			}
			if lower > upper {
				return nil, fmt.Errorf("invalid surface threshold")
			}
			bindings[i] = func(region *OverworldFeatureRegion, _ func(int32, int32, int32) (string, error)) placementOperation {
				return func(p FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
					h, err := region.HeightmapAt(modifier.Heightmap, p.X, p.Z)
					if err != nil {
						return err
					}
					// Java promotes the additions to long, avoiding int overflow.
					if int64(p.Y) >= int64(h)+int64(lower) && int64(p.Y) <= int64(h)+int64(upper) {
						return next(p)
					}
					return nil
				}
			}
		case "minecraft:biome":
			template.needsBiome = true
			if feature == nil {
				return nil, fmt.Errorf("biome filter without top-level feature context")
			}
			bindings[i] = func(region *OverworldFeatureRegion, biomeAtQuart func(int32, int32, int32) (string, error)) placementOperation {
				return func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
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
			}
		case "minecraft:heightmap":
			bindings[i] = func(region *OverworldFeatureRegion, biomeAtQuart func(int32, int32, int32) (string, error)) placementOperation {
				return func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
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
			}
		case "minecraft:block_predicate_filter":
			predicate, err := compileFeaturePredicate(modifier.Predicate)
			if err != nil {
				return nil, err
			}
			bindings[i] = func(region *OverworldFeatureRegion, biomeAtQuart func(int32, int32, int32) (string, error)) placementOperation {
				return func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
					matches, err := predicate(region, position)
					if err != nil {
						return err
					}
					if matches {
						return next(position)
					}
					return nil
				}
			}
		case "minecraft:surface_water_depth_filter":
			bindings[i] = func(region *OverworldFeatureRegion, biomeAtQuart func(int32, int32, int32) (string, error)) placementOperation {
				return func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
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
			bindings[i] = func(region *OverworldFeatureRegion, biomeAtQuart func(int32, int32, int32) (string, error)) placementOperation {
				return func(position FeaturePosition, _ *WorldgenRandom, next positionConsumer) error {
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
	}
	template.bindings = bindings
	return template, nil
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
