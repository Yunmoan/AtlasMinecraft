package worldgen

import (
	"encoding/json"
	"fmt"
	"math"
)

type featureFloatProvider func(*WorldgenRandom) float32

func decodeFeatureFloatProvider(raw json.RawMessage, lower, upper float32) (featureFloatProvider, error) {
	var constant float32
	if json.Unmarshal(raw, &constant) == nil && string(raw) != "null" {
		if constant < lower || constant > upper {
			return nil, fmt.Errorf("float provider outside range")
		}
		return func(*WorldgenRandom) float32 { return constant }, nil
	}
	var c struct {
		Type      string  `json:"type"`
		Value     float32 `json:"value"`
		Min       float32 `json:"min_inclusive"`
		Max       float32 `json:"max_exclusive"`
		Mean      float32 `json:"mean"`
		Deviation float32 `json:"deviation"`
		ClampMin  float32 `json:"min"`
		ClampMax  float32 `json:"max"`
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	switch c.Type {
	case "minecraft:constant":
		return decodeFeatureFloatProvider(json.RawMessage(fmt.Sprint(c.Value)), lower, upper)
	case "minecraft:uniform":
		if c.Min < lower || c.Max > upper || c.Min >= c.Max {
			return nil, fmt.Errorf("invalid uniform float provider")
		}
		return func(r *WorldgenRandom) float32 { return c.Min + r.NextFloat()*(c.Max-c.Min) }, nil
	case "minecraft:clamped_normal":
		if c.ClampMin < lower || c.ClampMax > upper || c.ClampMin > c.ClampMax || math.IsNaN(float64(c.Mean)) || c.Deviation < 0 {
			return nil, fmt.Errorf("invalid normal float provider")
		}
		return func(r *WorldgenRandom) float32 {
			return max(c.ClampMin, min(c.ClampMax, c.Mean+float32(r.NextGaussian())*c.Deviation))
		}, nil
	default:
		return nil, fmt.Errorf("unsupported float provider %s", c.Type)
	}
}
