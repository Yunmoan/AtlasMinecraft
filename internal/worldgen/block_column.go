package worldgen

import (
	"encoding/json"
	"fmt"
)

func compileBlockColumnPlan(definition ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config struct {
		Direction     string          `json:"direction"`
		PrioritizeTip bool            `json:"prioritize_tip"`
		Allowed       json.RawMessage `json:"allowed_placement"`
		Layers        []struct {
			Height   json.RawMessage `json:"height"`
			Provider json.RawMessage `json:"provider"`
		} `json:"layers"`
	}
	if err := json.Unmarshal(definition.Config, &config); err != nil {
		return nil, err
	}
	directions := map[string][3]int32{"up": {0, 1, 0}, "down": {0, -1, 0}, "north": {0, 0, -1}, "south": {0, 0, 1}, "west": {-1, 0, 0}, "east": {1, 0, 0}}
	direction, ok := directions[config.Direction]
	if !ok || len(config.Layers) > 64 {
		return nil, fmt.Errorf("invalid block column configuration")
	}
	allowed, err := compileFeaturePredicate(config.Allowed)
	if err != nil {
		return nil, err
	}
	heights := make([]featureIntProvider, len(config.Layers))
	providers := make([]featureStateProvider, len(config.Layers))
	for i, layer := range config.Layers {
		heights[i], err = decodeFeatureIntProvider(layer.Height, 0, 128)
		if err != nil {
			return nil, err
		}
		providers[i], err = decodeFeatureStateProvider(layer.Provider, validateColumnState)
		if err != nil {
			return nil, err
		}
	}
	return func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
		lengths := make([]int32, len(heights))
		total := int32(0)
		for i, height := range heights {
			lengths[i] = height(random)
			total += lengths[i]
		}
		if total == 0 {
			return false, nil
		}
		p := treeOffset(origin, direction[0], direction[1], direction[2])
		for placed := int32(0); placed < total; placed++ {
			ok, err := allowed(region, p)
			if err != nil {
				return false, err
			}
			if !ok {
				remove := total - placed
				index, step := len(lengths)-1, -1
				if config.PrioritizeTip {
					index, step = 0, 1
				}
				for ; index >= 0 && index < len(lengths) && remove > 0; index += step {
					count := min(lengths[index], remove)
					lengths[index] -= count
					remove -= count
				}
				break
			}
			p = treeOffset(p, direction[0], direction[1], direction[2])
		}
		p = origin
		for i, count := range lengths {
			for j := int32(0); j < count; j++ {
				if err := setFeatureBlock(region, p, providers[i](random, p)); err != nil {
					return false, err
				}
				p = treeOffset(p, direction[0], direction[1], direction[2])
			}
		}
		return true, nil
	}, nil
}

func validateColumnState(s SurfaceBlockState) error {
	switch s.Name {
	case "minecraft:cave_vines", "minecraft:cave_vines_plant", "minecraft:big_dripleaf", "minecraft:big_dripleaf_stem":
		return nil
	default:
		return validateVegetationState(s)
	}
}
