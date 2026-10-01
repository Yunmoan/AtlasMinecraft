package worldgen

import (
	"encoding/json"
	"fmt"
)

func decodeFeatureHeightProvider(raw json.RawMessage) (featureIntProvider, error) {
	var anchor heightAnchor
	if json.Unmarshal(raw, &anchor) == nil && (anchor.Absolute != nil || anchor.AboveBottom != nil || anchor.BelowTop != nil) {
		value, err := anchor.resolve(-64, 384)
		return func(*WorldgenRandom) int32 { return value }, err
	}
	var config struct {
		Type    string       `json:"type"`
		Min     heightAnchor `json:"min_inclusive"`
		Max     heightAnchor `json:"max_inclusive"`
		Value   heightAnchor `json:"value"`
		Plateau int32        `json:"plateau"`
		Inner   *int32       `json:"inner"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	if config.Type == "minecraft:constant" {
		value, err := config.Value.resolve(-64, 384)
		return func(*WorldgenRandom) int32 { return value }, err
	}
	lower, err := config.Min.resolve(-64, 384)
	if err != nil {
		return nil, err
	}
	upper, err := config.Max.resolve(-64, 384)
	if err != nil {
		return nil, err
	}
	inner := int32(1)
	if config.Inner != nil {
		inner = *config.Inner
	}
	switch config.Type {
	case "minecraft:uniform", "minecraft:trapezoid", "minecraft:biased_to_bottom", "minecraft:very_biased_to_bottom":
	default:
		return nil, fmt.Errorf("unsupported feature height provider %q", config.Type)
	}
	if inner < 1 || config.Plateau < 0 {
		return nil, fmt.Errorf("invalid height provider parameters")
	}
	return func(random *WorldgenRandom) int32 {
		if lower > upper {
			return lower
		}
		switch config.Type {
		case "minecraft:uniform":
			return randomBetweenInclusive(random, lower, upper)
		case "minecraft:trapezoid":
			if config.Plateau >= upper-lower {
				return randomBetweenInclusive(random, lower, upper)
			}
			first := (upper - lower - config.Plateau) / 2
			return lower + randomBetweenInclusive(random, 0, upper-lower-first) + randomBetweenInclusive(random, 0, first)
		case "minecraft:biased_to_bottom":
			if upper-lower-inner+1 <= 0 {
				return lower
			}
			first := random.NextIntBound(upper - lower - inner + 1)
			return lower + random.NextIntBound(first+inner)
		default:
			if upper-lower-inner+1 <= 0 {
				return lower
			}
			first := randomBetweenInclusive(random, lower+inner, upper)
			second := randomBetweenInclusive(random, lower, first-1)
			return randomBetweenInclusive(random, lower, second-1+inner)
		}
	}, nil
}
