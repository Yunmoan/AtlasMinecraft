package worldgen

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

func compileGeodePlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Blocks struct {
			Filling    json.RawMessage     `json:"filling_provider"`
			Inner      json.RawMessage     `json:"inner_layer_provider"`
			Alternate  json.RawMessage     `json:"alternate_inner_layer_provider"`
			Middle     json.RawMessage     `json:"middle_layer_provider"`
			Outer      json.RawMessage     `json:"outer_layer_provider"`
			Placements []SurfaceBlockState `json:"inner_placements"`
			Cannot     string              `json:"cannot_replace"`
			Invalid    string              `json:"invalid_blocks"`
		} `json:"blocks"`
		Layers struct {
			Filling float64 `json:"filling"`
			Inner   float64 `json:"inner_layer"`
			Middle  float64 `json:"middle_layer"`
			Outer   float64 `json:"outer_layer"`
		} `json:"layers"`
		Crack struct {
			Size   float64 `json:"base_crack_size"`
			Offset int32   `json:"crack_point_offset"`
			Chance float64 `json:"generate_crack_chance"`
		} `json:"crack"`
		Potential        float64         `json:"use_potential_placements_chance"`
		Alternate        float64         `json:"use_alternate_layer0_chance"`
		RequireAlternate bool            `json:"placements_require_layer0_alternate"`
		Distance         json.RawMessage `json:"outer_wall_distance"`
		Points           json.RawMessage `json:"distribution_points"`
		Offset           json.RawMessage `json:"point_offset"`
		Min              int32           `json:"min_gen_offset"`
		Max              int32           `json:"max_gen_offset"`
		Noise            float64         `json:"noise_multiplier"`
		InvalidThreshold int32           `json:"invalid_blocks_threshold"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Min > c.Max || c.Min < -64 || c.Max > 64 || len(c.Blocks.Placements) == 0 || c.InvalidThreshold < 0 {
		return nil, fmt.Errorf("invalid geode configuration")
	}
	cannot, ok := vegetationTags[strings.TrimPrefix(c.Blocks.Cannot, "#minecraft:")]
	if !ok {
		return nil, fmt.Errorf("unsupported geode replacement tag")
	}
	invalid, ok := vegetationTags[strings.TrimPrefix(c.Blocks.Invalid, "#minecraft:")]
	if !ok {
		return nil, fmt.Errorf("unsupported geode invalid tag")
	}
	var providers [5]featureStateProvider
	for i, raw := range []json.RawMessage{c.Blocks.Filling, c.Blocks.Inner, c.Blocks.Alternate, c.Blocks.Middle, c.Blocks.Outer} {
		p, err := decodeFeatureStateProvider(raw, func(SurfaceBlockState) error { return nil })
		if err != nil {
			return nil, err
		}
		providers[i] = p
	}
	distance, err := decodeFeatureIntProvider(c.Distance, 1, 20)
	if err != nil {
		return nil, err
	}
	points, err := decodeFeatureIntProvider(c.Points, 1, 20)
	if err != nil {
		return nil, err
	}
	offset, err := decodeFeatureIntProvider(c.Offset, 0, 20)
	if err != nil {
		return nil, err
	}
	var bounds struct {
		Max int32 `json:"max_inclusive"`
	}
	if err := json.Unmarshal(c.Distance, &bounds); err != nil {
		return nil, err
	}
	if bounds.Max < 1 {
		return nil, fmt.Errorf("invalid geode distance bound")
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		n := points(r)
		noise, err := newProviderNormalNoise(region.biomeSeed, -4, []float64{1})
		if err != nil {
			return false, err
		}
		ratio := float64(n) / float64(bounds.Max)
		filling, inner, middle, outer := 1/math.Sqrt(c.Layers.Filling), 1/math.Sqrt(c.Layers.Inner+ratio), 1/math.Sqrt(c.Layers.Middle+ratio), 1/math.Sqrt(c.Layers.Outer+ratio)
		crackExtra := float64(0)
		if n > 3 {
			crackExtra = ratio
		}
		crackLimit := 1 / math.Sqrt(c.Crack.Size+r.NextDouble()/2+crackExtra)
		crack := float64(r.NextFloat()) < c.Crack.Chance
		type geodePoint struct {
			p      FeaturePosition
			offset int32
		}
		var centers []geodePoint
		invalidCount := int32(0)
		for i := int32(0); i < n; i++ {
			x, y, z := distance(r), distance(r), distance(r)
			p := treeOffset(origin, x, y, z)
			s := e.block(p)
			if surfaceAir(s) || invalid[s.Name] {
				invalidCount++
				if invalidCount > c.InvalidThreshold {
					return false, e.err
				}
			}
			centers = append(centers, geodePoint{p, offset(r)})
		}
		var cracks []FeaturePosition
		if crack {
			x, z := int32(0), int32(0)
			choice := r.NextIntBound(4)
			size := n*2 + 1
			if choice == 0 || choice == 2 {
				x = size
			}
			if choice == 1 || choice == 2 {
				z = size
			}
			for _, y := range []int32{7, 5, 1} {
				cracks = append(cracks, treeOffset(origin, x, y, z))
			}
		}
		set := func(p FeaturePosition, s SurfaceBlockState) {
			if !cannot[e.block(p).Name] {
				e.set(p, s)
			}
		}
		squareDistance := func(a, b FeaturePosition) float64 {
			x, y, z := float64(a.X)-float64(b.X), float64(a.Y)-float64(b.Y), float64(a.Z)-float64(b.Z)
			return x*x + y*y + z*z
		}
		var buds []FeaturePosition
		for z := c.Min; z <= c.Max; z++ {
			for y := c.Min; y <= c.Max; y++ {
				for x := c.Min; x <= c.Max; x++ {
					p := treeOffset(origin, x, y, z)
					noiseValue := noise.Noise(float64(p.X), float64(p.Y), float64(p.Z)) * c.Noise
					value, crackValue := float64(0), float64(0)
					for _, center := range centers {
						value += 1/math.Sqrt(squareDistance(p, center.p)+float64(center.offset)) + noiseValue
					}
					for _, center := range cracks {
						crackValue += 1/math.Sqrt(squareDistance(p, center)+float64(c.Crack.Offset)) + noiseValue
					}
					if value < outer {
						continue
					}
					if crack && crackValue >= crackLimit && value < filling {
						set(p, SurfaceBlockState{Name: "minecraft:air"})
						for d := 0; d < 6; d++ {
							q := directionPosition(p, d)
							if featureFluid(e.block(q)) {
								if err := region.ScheduleFluidUpdate(q); err != nil {
									return false, err
								}
							}
						}
					} else if value >= filling {
						set(p, providers[0](r, p))
					} else if value >= inner {
						alternate := float64(r.NextFloat()) < c.Alternate
						i := 1
						if alternate {
							i = 2
						}
						set(p, providers[i](r, p))
						if (!c.RequireAlternate || alternate) && float64(r.NextFloat()) < c.Potential {
							buds = append(buds, p)
						}
					} else if value >= middle {
						set(p, providers[3](r, p))
					} else {
						set(p, providers[4](r, p))
					}
				}
			}
		}
		for _, p := range buds {
			s := c.Blocks.Placements[r.NextIntBound(int32(len(c.Blocks.Placements)))]
			for d := 0; d < 6; d++ {
				if _, ok := s.Properties["facing"]; ok {
					s = stateWithProperty(s, "facing", directionNames[d])
				}
				q := directionPosition(p, d)
				old := e.block(q)
				if _, ok := s.Properties["waterlogged"]; ok {
					s = stateWithProperty(s, "waterlogged", fmt.Sprint(hasSourceWater(old)))
				}
				if surfaceAir(old) || old.Name == "minecraft:water" && hasSourceWater(old) {
					set(q, s)
					break
				}
			}
		}
		return e.err == nil, e.err
	}, nil
}
