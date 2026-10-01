package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"encoding/json"
	"fmt"
	"strings"
)

func featureSturdy(e *featureEditor, p FeaturePosition, d int) bool {
	s := e.block(p)
	id, err := blockstate.BlockStateID(s.Name, s.Properties)
	if err != nil {
		e.err = err
		return false
	}
	rule, err := blockrules.Lookup(id)
	if err != nil {
		e.err = err
		return false
	}
	return rule.FaceSturdy[d]
}

func compileVegetationPatchPlan(def ConfiguredFeatureDefinition, depth int) (configuredFeaturePlacer, error) {
	var c struct {
		Replaceable  string          `json:"replaceable"`
		Ground       json.RawMessage `json:"ground_state"`
		Feature      json.RawMessage `json:"vegetation_feature"`
		Surface      string          `json:"surface"`
		Depth        json.RawMessage `json:"depth"`
		BottomChance float32         `json:"extra_bottom_block_chance"`
		Vertical     int32           `json:"vertical_range"`
		Chance       float32         `json:"vegetation_chance"`
		Radius       json.RawMessage `json:"xz_radius"`
		EdgeChance   float32         `json:"extra_edge_column_chance"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	tag := strings.TrimPrefix(c.Replaceable, "#minecraft:")
	replaceable, ok := vegetationTags[tag]
	if !ok || !strings.HasPrefix(c.Replaceable, "#minecraft:") || c.Vertical < 1 || c.Vertical > 256 || c.Surface != "floor" && c.Surface != "ceiling" {
		return nil, fmt.Errorf("unsupported vegetation patch configuration")
	}
	for _, p := range []float32{c.BottomChance, c.Chance, c.EdgeChance} {
		if p < 0 || p > 1 {
			return nil, fmt.Errorf("invalid vegetation patch probability")
		}
	}
	ground, err := decodeFeatureStateProvider(c.Ground, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return nil, err
	}
	vegetation, err := compileNestedFeaturePlacement(c.Feature, depth+1)
	if err != nil {
		return nil, err
	}
	layers, err := decodeFeatureIntProvider(c.Depth, 1, 128)
	if err != nil {
		return nil, err
	}
	radius, err := decodeFeatureIntProvider(c.Radius, 0, 64)
	if err != nil {
		return nil, err
	}
	dir := 0
	if c.Surface == "ceiling" {
		dir = 1
	}
	waterlogged := def.Type == "minecraft:waterlogged_vegetation_patch"
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		rx, rz := radius(r)+1, radius(r)+1
		var patches javaPositionSet
		for x := -rx; x <= rx; x++ {
			for z := -rz; z <= rz; z++ {
				xedge, zedge := x == -rx || x == rx, z == -rz || z == rz
				if xedge && zedge || (xedge || zedge) && (c.EdgeChance == 0 || r.NextFloat() > c.EdgeChance) {
					continue
				}
				p := treeOffset(origin, x, 0, z)
				for i := int32(0); surfaceAir(e.block(p)) && i < c.Vertical; i++ {
					p = directionPosition(p, dir)
				}
				for i := int32(0); !surfaceAir(e.block(p)) && i < c.Vertical; i++ {
					p = directionPosition(p, dir^1)
				}
				q := directionPosition(p, dir)
				if !surfaceAir(e.block(p)) || !featureSturdy(&e, q, dir^1) {
					continue
				}
				n := layers(r)
				if c.BottomChance > 0 && r.NextFloat() < c.BottomChance {
					n++
				}
				success := true
				root := q
				for i := int32(0); i < n; i++ {
					state := ground(r, q)
					old := e.block(q)
					if old.Name == state.Name {
						continue
					}
					if !replaceable[old.Name] {
						success = i != 0
						break
					}
					e.set(q, state)
					q = directionPosition(q, dir)
				}
				if success {
					patches.add(root)
				}
			}
		}
		if waterlogged {
			var enclosed javaPositionSet
			for _, p := range patches.positions() {
				valid := true
				for _, d := range []int{2, 5, 3, 4, 0} {
					if !featureSturdy(&e, directionPosition(p, d), d^1) {
						valid = false
						break
					}
				}
				if valid {
					enclosed.add(p)
				}
			}
			patches = enclosed
			for _, p := range patches.positions() {
				e.set(p, SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "0"}})
			}
		}
		for _, p := range patches.positions() {
			if c.Chance <= 0 || r.NextFloat() >= c.Chance {
				continue
			}
			q := directionPosition(p, dir^1)
			if waterlogged {
				q = p
			}
			placed, err := vegetation(region, r, q)
			if err != nil {
				return false, err
			}
			if waterlogged && placed {
				state := e.block(p)
				if state.Properties["waterlogged"] == "false" {
					e.set(p, stateWithProperty(state, "waterlogged", "true"))
				}
			}
		}
		return patches.size > 0 && e.err == nil, e.err
	}, nil
}
