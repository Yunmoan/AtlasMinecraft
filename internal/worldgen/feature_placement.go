package worldgen

import (
	"encoding/json"
	"fmt"
)

type FeaturePosition struct {
	X, Y, Z int32
}

type FeatureHeightmapAt func(heightmap string, blockX, blockZ int32) (int32, error)

type heightAnchor struct {
	Absolute    *int32 `json:"absolute"`
	AboveBottom *int32 `json:"above_bottom"`
	BelowTop    *int32 `json:"below_top"`
}

func (anchor heightAnchor) resolve(minY, height int32) (int32, error) {
	switch {
	case anchor.Absolute != nil:
		return *anchor.Absolute, nil
	case anchor.AboveBottom != nil:
		return minY + *anchor.AboveBottom, nil
	case anchor.BelowTop != nil:
		return minY + height - 1 - *anchor.BelowTop, nil
	default:
		return 0, fmt.Errorf("unknown feature height anchor")
	}
}

func randomBetweenInclusive(random *WorldgenRandom, lower, upper int32) int32 {
	return lower + random.NextIntBound(upper-lower+1)
}

// OverworldPreBiomeCandidates applies supported vanilla placement modifiers
// through (but not including) the final biome filter. These positions are not
// placed features: the biome check and configured feature still must run.
func OverworldPreBiomeCandidates(worldSeed int64, chunkX, chunkZ int32, feature PlacedFeature) ([]FeaturePosition, error) {
	positions, _, err := overworldPreBiomeCandidates(worldSeed, chunkX, chunkZ, feature, nil)
	return positions, err
}

func OverworldPreBiomeCandidatesWithHeightmap(worldSeed int64, chunkX, chunkZ int32, feature PlacedFeature,
	heightmapAt FeatureHeightmapAt) ([]FeaturePosition, error) {
	positions, _, err := overworldPreBiomeCandidates(worldSeed, chunkX, chunkZ, feature, heightmapAt)
	return positions, err
}

func overworldPreBiomeCandidates(worldSeed int64, chunkX, chunkZ int32, feature PlacedFeature,
	heightmapAt FeatureHeightmapAt) ([]FeaturePosition, *WorldgenRandom, error) {
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		return nil, nil, err
	}
	registered := false
	for _, entry := range order.features {
		if entry == feature {
			registered = true
			break
		}
	}
	if !registered {
		return nil, nil, fmt.Errorf("feature %+v is not in the 1.21.1 overworld order", feature)
	}
	definition, err := LoadPlacedFeatureDefinition(feature.Name)
	if err != nil {
		return nil, nil, err
	}
	decorationSeed := OverworldDecorationSeed(worldSeed, chunkX, chunkZ)
	random := OverworldPlacedFeatureRandom(decorationSeed, feature)
	positions := []FeaturePosition{{X: chunkX * 16, Z: chunkZ * 16}}
	for modifierIndex, raw := range definition.Placement {
		var kind struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &kind); err != nil {
			return nil, nil, err
		}
		switch kind.Type {
		case "minecraft:count":
			var modifier struct {
				Count json.RawMessage `json:"count"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil {
				return nil, nil, fmt.Errorf("unsupported count for %s: %s", feature.Name, raw)
			}
			var fixedCount int32
			var provider struct {
				Type         string `json:"type"`
				MinInclusive int32  `json:"min_inclusive"`
				MaxInclusive int32  `json:"max_inclusive"`
			}
			uniform := json.Unmarshal(modifier.Count, &fixedCount) != nil
			if uniform {
				if err := json.Unmarshal(modifier.Count, &provider); err != nil || provider.Type != "minecraft:uniform" || provider.MinInclusive > provider.MaxInclusive {
					return nil, nil, fmt.Errorf("unsupported count provider for %s: %s", feature.Name, modifier.Count)
				}
			}
			var next []FeaturePosition
			for _, position := range positions {
				count := fixedCount
				if uniform {
					count = randomBetweenInclusive(random, provider.MinInclusive, provider.MaxInclusive)
				}
				if count < 0 || count > 256 {
					return nil, nil, fmt.Errorf("unsupported count for %s: %s", feature.Name, modifier.Count)
				}
				for range count {
					next = append(next, position)
				}
			}
			positions = next
		case "minecraft:rarity_filter":
			var modifier struct {
				Chance int32 `json:"chance"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil || modifier.Chance <= 0 {
				return nil, nil, fmt.Errorf("invalid rarity filter for %s: %s", feature.Name, raw)
			}
			var next []FeaturePosition
			for _, position := range positions {
				if random.NextFloat() < 1/float32(modifier.Chance) {
					next = append(next, position)
				}
			}
			positions = next
		case "minecraft:in_square":
			for index := range positions {
				positions[index].X += random.NextIntBound(16)
				positions[index].Z += random.NextIntBound(16)
			}
		case "minecraft:heightmap":
			var modifier struct {
				Heightmap string `json:"heightmap"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil || modifier.Heightmap == "" {
				return nil, nil, fmt.Errorf("invalid heightmap modifier for %s: %s", feature.Name, raw)
			}
			if heightmapAt == nil {
				return nil, nil, fmt.Errorf("heightmap %s is required for %s", modifier.Heightmap, feature.Name)
			}
			var next []FeaturePosition
			for _, position := range positions {
				height, err := heightmapAt(modifier.Heightmap, position.X, position.Z)
				if err != nil {
					return nil, nil, err
				}
				if height > -64 {
					position.Y = height
					next = append(next, position)
				}
			}
			positions = next
		case "minecraft:height_range":
			var modifier struct {
				Height json.RawMessage `json:"height"`
			}
			if err := json.Unmarshal(raw, &modifier); err != nil {
				return nil, nil, err
			}
			height, err := decodeFeatureHeightProvider(modifier.Height)
			if err != nil {
				return nil, nil, err
			}
			for index := range positions {
				positions[index].Y = height(random)
			}
		case "minecraft:biome":
			if modifierIndex != len(definition.Placement)-1 {
				return nil, nil, fmt.Errorf("biome filter is not last for %s", feature.Name)
			}
			return positions, random, nil
		default:
			return nil, nil, fmt.Errorf("unsupported placement modifier %q for %s", kind.Type, feature.Name)
		}
	}
	return positions, random, nil
}

// OverworldFeaturePositions also applies the vanilla biome membership filter.
// It does not run the configured feature or modify blocks.
func OverworldFeaturePositions(worldSeed int64, chunkX, chunkZ int32, feature PlacedFeature,
	biomeAtQuart func(int32, int32, int32) (string, error)) ([]FeaturePosition, error) {
	positions, _, err := overworldFeaturePositions(worldSeed, chunkX, chunkZ, feature, biomeAtQuart, nil)
	return positions, err
}

func OverworldFeaturePositionsWithHeightmap(worldSeed int64, chunkX, chunkZ int32, feature PlacedFeature,
	biomeAtQuart func(int32, int32, int32) (string, error), heightmapAt FeatureHeightmapAt) ([]FeaturePosition, error) {
	positions, _, err := overworldFeaturePositions(worldSeed, chunkX, chunkZ, feature, biomeAtQuart, heightmapAt)
	return positions, err
}

func overworldFeaturePositions(worldSeed int64, chunkX, chunkZ int32, feature PlacedFeature,
	biomeAtQuart func(int32, int32, int32) (string, error), heightmapAt FeatureHeightmapAt) ([]FeaturePosition, *WorldgenRandom, error) {
	if biomeAtQuart == nil {
		return nil, nil, fmt.Errorf("biome source is required for feature placement")
	}
	positions, random, err := overworldPreBiomeCandidates(worldSeed, chunkX, chunkZ, feature, heightmapAt)
	if err != nil {
		return nil, nil, err
	}
	definition, err := LoadPlacedFeatureDefinition(feature.Name)
	if err != nil {
		return nil, nil, err
	}
	if len(definition.Placement) == 0 || !jsonContainsBiomeFilter(definition.Placement[len(definition.Placement)-1]) {
		return positions, random, nil
	}
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		return nil, nil, err
	}
	var filtered []FeaturePosition
	for _, position := range positions {
		biome, err := biomeAtQuart(position.X>>2, position.Y>>2, position.Z>>2)
		if err != nil {
			return nil, nil, err
		}
		found := false
		for _, index := range order.biomes[biome][feature.Step] {
			if index == feature.Index {
				found = true
				break
			}
		}
		if found {
			filtered = append(filtered, position)
		}
	}
	return filtered, random, nil
}

func jsonContainsBiomeFilter(raw json.RawMessage) bool {
	var modifier struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(raw, &modifier) == nil && modifier.Type == "minecraft:biome"
}
