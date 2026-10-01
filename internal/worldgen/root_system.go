package worldgen

import (
	"encoding/json"
	"fmt"
	"strings"
)

func compileRootSystemPlan(def ConfiguredFeatureDefinition, depth int) (configuredFeaturePlacer, error) {
	var c struct {
		Tree            json.RawMessage `json:"feature"`
		Space           int32           `json:"required_vertical_space_for_tree"`
		RootRadius      int32           `json:"root_radius"`
		Replaceable     string          `json:"root_replaceable"`
		RootState       json.RawMessage `json:"root_state_provider"`
		RootAttempts    int32           `json:"root_placement_attempts"`
		Height          int32           `json:"root_column_max_height"`
		HangingRadius   int32           `json:"hanging_root_radius"`
		HangingSpan     int32           `json:"hanging_roots_vertical_span"`
		HangingState    json.RawMessage `json:"hanging_root_state_provider"`
		HangingAttempts int32           `json:"hanging_root_placement_attempts"`
		Water           int32           `json:"allowed_vertical_water_for_tree"`
		Allowed         json.RawMessage `json:"allowed_tree_position"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Space < 1 || c.RootRadius < 1 || c.HangingRadius < 1 || c.HangingSpan < 1 || c.Height < 1 || c.Height > 4096 || c.RootAttempts < 0 || c.HangingAttempts < 0 {
		return nil, fmt.Errorf("invalid root system configuration")
	}
	replaceable, ok := vegetationTags[strings.TrimPrefix(c.Replaceable, "#minecraft:")]
	if !ok {
		return nil, fmt.Errorf("unsupported root system tag")
	}
	tree, err := compileNestedFeaturePlacement(c.Tree, depth+1)
	if err != nil {
		return nil, err
	}
	root, err := decodeFeatureStateProvider(c.RootState, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return nil, err
	}
	hanging, err := decodeFeatureStateProvider(c.HangingState, validateVegetationState)
	if err != nil {
		return nil, err
	}
	allowed, err := compileFeaturePredicate(c.Allowed)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !surfaceAir(e.block(origin)) {
			return false, e.err
		}
		p := origin
		found := false
		for i := int32(0); i < c.Height; i++ {
			p.Y++
			valid, err := allowed(region, p)
			if err != nil {
				return false, err
			}
			if !valid {
				continue
			}
			for j := int32(1); j <= c.Space; j++ {
				s := e.block(treeOffset(p, 0, j, 0))
				if !surfaceAir(s) && !(j+1 <= c.Water && (s.Name == "minecraft:water" || s.Properties["waterlogged"] == "true")) {
					valid = false
					break
				}
			}
			if !valid {
				continue
			}
			below := e.block(treeOffset(p, 0, -1, 0))
			solid, err := featureSolid(below)
			if err != nil {
				return false, err
			}
			if below.Name == "minecraft:lava" || !solid {
				break
			}
			placed, err := tree(region, r, p)
			if err != nil {
				return false, err
			}
			if !placed {
				continue
			}
			for y := origin.Y; y < origin.Y+i; y++ {
				for n := int32(0); n < c.RootAttempts; n++ {
					x := r.NextIntBound(c.RootRadius) - r.NextIntBound(c.RootRadius)
					z := r.NextIntBound(c.RootRadius) - r.NextIntBound(c.RootRadius)
					q := treeOffset(origin, x, y-origin.Y, z)
					if replaceable[e.block(q).Name] {
						e.set(q, root(r, q))
					}
				}
			}
			found = true
			break
		}
		if found {
			for n := int32(0); n < c.HangingAttempts; n++ {
				x := r.NextIntBound(c.HangingRadius) - r.NextIntBound(c.HangingRadius)
				y := r.NextIntBound(c.HangingSpan) - r.NextIntBound(c.HangingSpan)
				z := r.NextIntBound(c.HangingRadius) - r.NextIntBound(c.HangingRadius)
				q := treeOffset(origin, x, y, z)
				if !surfaceAir(e.block(q)) {
					continue
				}
				state := hanging(r, q)
				survives, err := featurePlantSurvives(region, q, state)
				if err != nil {
					return false, err
				}
				if survives && featureSturdy(&e, treeOffset(q, 0, 1, 0), 0) {
					e.set(q, state)
				}
			}
		}
		return e.err == nil, e.err
	}, nil
}
