package worldgen

import (
	"encoding/json"
	"fmt"
)

type featureIntProvider func(*WorldgenRandom) int32

// decodeFeatureIntProvider preserves draws for singleton uniform ranges and
// weighted lists. Constant providers consume no random values.
func decodeFeatureIntProvider(raw json.RawMessage, lower, upper int32) (featureIntProvider, error) {
	var constant int32
	if json.Unmarshal(raw, &constant) == nil && string(raw) != "null" {
		if constant < lower || constant > upper {
			return nil, fmt.Errorf("integer provider %d outside [%d,%d]", constant, lower, upper)
		}
		return func(*WorldgenRandom) int32 { return constant }, nil
	}
	var config struct {
		Type         string          `json:"type"`
		Value        int32           `json:"value"`
		Min          int32           `json:"min_inclusive"`
		Max          int32           `json:"max_inclusive"`
		Mean         float32         `json:"mean"`
		Deviation    float32         `json:"deviation"`
		Source       json.RawMessage `json:"source"`
		Distribution []struct {
			Data   json.RawMessage `json:"data"`
			Weight int32           `json:"weight"`
		} `json:"distribution"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	switch config.Type {
	case "minecraft:clamped":
		if config.Min < lower || config.Max > upper || config.Min > config.Max {
			return nil, fmt.Errorf("invalid clamped integer provider")
		}
		source, err := decodeFeatureIntProvider(config.Source, -2147483648, 2147483647)
		if err != nil {
			return nil, err
		}
		return func(r *WorldgenRandom) int32 { return max(config.Min, min(config.Max, source(r))) }, nil
	case "minecraft:clamped_normal":
		if config.Min < lower || config.Max > upper || config.Min > config.Max || config.Deviation < 0 {
			return nil, fmt.Errorf("invalid normal integer provider")
		}
		return func(r *WorldgenRandom) int32 {
			return int32(max(float32(config.Min), min(float32(config.Max), config.Mean+float32(r.NextGaussian())*config.Deviation)))
		}, nil
	case "minecraft:constant":
		return decodeFeatureIntProvider(json.RawMessage(fmt.Sprint(config.Value)), lower, upper)
	case "minecraft:uniform", "minecraft:biased_to_bottom":
		if config.Min < lower || config.Max > upper || config.Min > config.Max {
			return nil, fmt.Errorf("invalid uniform integer provider %s", raw)
		}
		return func(random *WorldgenRandom) int32 {
			if config.Type == "minecraft:biased_to_bottom" {
				return config.Min + random.NextIntBound(random.NextIntBound(config.Max-config.Min+1)+1)
			}
			return randomBetweenInclusive(random, config.Min, config.Max)
		}, nil
	case "minecraft:weighted_list":
		var providers []featureIntProvider
		var weights []int32
		var total int32
		for _, entry := range config.Distribution {
			provider, err := decodeFeatureIntProvider(entry.Data, lower, upper)
			if err != nil {
				return nil, err
			}
			if entry.Weight < 0 || int64(total)+int64(entry.Weight) > 2147483647 {
				return nil, fmt.Errorf("invalid integer provider weight %d", entry.Weight)
			}
			providers = append(providers, provider)
			weights = append(weights, entry.Weight)
			total += entry.Weight
		}
		if total == 0 {
			return nil, fmt.Errorf("empty weighted integer provider")
		}
		return func(random *WorldgenRandom) int32 {
			choice := random.NextIntBound(total)
			for i, weight := range weights {
				choice -= weight
				if choice < 0 {
					return providers[i](random)
				}
			}
			panic("worldgen: invalid weighted integer provider")
		}, nil
	default:
		return nil, fmt.Errorf("unsupported integer provider %q", config.Type)
	}
}
