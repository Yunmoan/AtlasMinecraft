package worldgen

import (
	"encoding/json"
	"fmt"
)

type featureColumn struct {
	floor, ceiling       int32
	hasFloor, hasCeiling bool
}

// Column.scan checks the origin before walking at most searchRange-1 blocks.
func scanFeatureColumn(e *featureEditor, origin FeaturePosition, searchRange int32, pass, stop func(SurfaceBlockState) bool) (featureColumn, bool) {
	var c featureColumn
	if !pass(e.block(origin)) {
		return c, false
	}
	for _, dy := range []int32{1, -1} {
		p := origin
		for i := int32(1); i < searchRange && pass(e.block(p)); i++ {
			p.Y += dy
		}
		if stop(e.block(p)) {
			if dy > 0 {
				c.ceiling, c.hasCeiling = p.Y, true
			} else {
				c.floor, c.hasFloor = p.Y, true
			}
		}
	}
	return c, true
}

func dripstoneSpace(s SurfaceBlockState) bool { return surfaceAir(s) || s.Name == "minecraft:water" }

func growPointedDripstone(e *featureEditor, origin FeaturePosition, dy, height int32, merge bool) {
	if !dripstoneSupport(e.block(treeOffset(origin, 0, -dy, 0))) {
		return
	}
	for i := int32(0); i < height; i++ {
		p := treeOffset(origin, 0, dy*i, 0)
		thickness := "tip"
		if merge {
			thickness = "tip_merge"
		}
		if i < height-1 {
			thickness = "frustum"
		}
		if i < height-2 {
			thickness = "middle"
		}
		if i == 0 && height >= 3 {
			thickness = "base"
		}
		direction := "up"
		if dy < 0 {
			direction = "down"
		}
		waterlogged := "false"
		s := e.block(p)
		if s.Name == "minecraft:water" || s.Properties["waterlogged"] == "true" {
			waterlogged = "true"
		}
		e.set(p, SurfaceBlockState{Name: "minecraft:pointed_dripstone", Properties: map[string]string{"vertical_direction": direction, "thickness": thickness, "waterlogged": waterlogged}})
	}
}

func replaceDripstone(e *featureEditor, p FeaturePosition) bool {
	if !vegetationTags["dripstone_replaceable_blocks"][e.block(p).Name] {
		return false
	}
	e.set(p, SurfaceBlockState{Name: "minecraft:dripstone_block"})
	return true
}

type dripstoneClusterConfig struct {
	Search         int32           `json:"floor_to_ceiling_search_range"`
	Height         json.RawMessage `json:"height"`
	Radius         json.RawMessage `json:"radius"`
	Diff           int32           `json:"max_stalagmite_stalactite_height_diff"`
	Deviation      int32           `json:"height_deviation"`
	Layer          json.RawMessage `json:"dripstone_block_layer_thickness"`
	Density        json.RawMessage `json:"density"`
	Wetness        json.RawMessage `json:"wetness"`
	EdgeChance     float32         `json:"chance_of_dripstone_column_at_max_distance_from_center"`
	EdgeDistance   int32           `json:"max_distance_from_edge_affecting_chance_of_dripstone_column"`
	CenterDistance int32           `json:"max_distance_from_center_affecting_height_bias"`
}

func compileDripstoneClusterPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c dripstoneClusterConfig
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Search < 1 || c.Search > 512 || c.Diff < 0 || c.Deviation < 0 || c.EdgeChance < 0 || c.EdgeChance > 1 || c.EdgeDistance < 1 || c.CenterDistance < 1 {
		return nil, fmt.Errorf("invalid dripstone cluster configuration")
	}
	height, err := decodeFeatureIntProvider(c.Height, 1, 128)
	if err != nil {
		return nil, err
	}
	radius, err := decodeFeatureIntProvider(c.Radius, 1, 128)
	if err != nil {
		return nil, err
	}
	layer, err := decodeFeatureIntProvider(c.Layer, 0, 128)
	if err != nil {
		return nil, err
	}
	density, err := decodeFeatureFloatProvider(c.Density, 0, 1)
	if err != nil {
		return nil, err
	}
	wetness, err := decodeFeatureFloatProvider(c.Wetness, 0, 1)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !dripstoneSpace(e.block(p)) {
			return false, e.err
		}
		h, w, d := height(r), wetness(r), density(r)
		rx, rz := radius(r), radius(r)
		for x := -rx; x <= rx; x++ {
			for z := -rz; z <= rz; z++ {
				edge := min(rx-abs32(x), rz-abs32(z))
				// Mth.clampedMap(float) keeps the edge interpolation in float arithmetic.
				chance := c.EdgeChance + max(float32(0), min(float32(1), float32(edge)/float32(c.EdgeDistance)))*(1-c.EdgeChance)
				q := treeOffset(p, x, 0, z)
				col, ok := scanFeatureColumn(&e, q, c.Search, dripstoneSpace, func(s SurfaceBlockState) bool { return !dripstoneSpace(s) })
				if !ok || !col.hasFloor && !col.hasCeiling {
					continue
				}
				wet := r.NextFloat() < w
				if wet && col.hasFloor && canDripstonePool(&e, treeOffset(q, 0, col.floor-q.Y, 0)) {
					e.set(treeOffset(q, 0, col.floor-q.Y, 0), SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "0"}})
					col.floor--
				}
				columnHeight := func(limit int32) int32 {
					if r.NextFloat() > d {
						return 0
					}
					distance := abs32(x) + abs32(z)
					fraction := min(float64(1), float64(distance)/float64(c.CenterDistance))
					base := float64(limit) / 2
					bias := float32(base + fraction*(0-base))
					v := bias + float32(r.NextGaussian())*float32(c.Deviation)
					return int32(max(float32(0), min(float32(limit), v)))
				}
				addLayer := func(y, dy int32) {
					n := layer(r)
					for i := int32(0); i < n; i++ {
						if !replaceDripstone(&e, treeOffset(q, 0, y-q.Y+dy*i, 0)) {
							break
						}
					}
				}
				ch, fh := int32(0), int32(0)
				ceilingChance := r.NextDouble() < float64(chance)
				if col.hasCeiling && ceilingChance && e.block(treeOffset(q, 0, col.ceiling-q.Y, 0)).Name != "minecraft:lava" {
					addLayer(col.ceiling, 1)
					limit := h
					if col.hasFloor {
						limit = min(h, col.ceiling-col.floor)
					}
					ch = columnHeight(limit)
				}
				floorChance := r.NextDouble() < float64(chance)
				if col.hasFloor && floorChance && e.block(treeOffset(q, 0, col.floor-q.Y, 0)).Name != "minecraft:lava" {
					addLayer(col.floor, -1)
					if col.hasCeiling {
						fh = max(int32(0), ch+randomBetweenInclusive(r, -c.Diff, c.Diff))
					} else {
						fh = columnHeight(h)
					}
				}
				if col.hasFloor && col.hasCeiling && col.ceiling-ch <= col.floor+fh {
					split := randomBetweenInclusive(r, max(col.ceiling-ch, col.floor+1), min(col.floor+fh, col.ceiling-1)+1)
					ch, fh = col.ceiling-split, split-1-col.floor
				}
				merge := r.NextBoolean() && ch > 0 && fh > 0 && col.hasFloor && col.hasCeiling && ch+fh == col.ceiling-col.floor-1
				if col.hasCeiling {
					growPointedDripstone(&e, treeOffset(q, 0, col.ceiling-1-q.Y, 0), -1, ch, merge)
				}
				if col.hasFloor {
					growPointedDripstone(&e, treeOffset(q, 0, col.floor+1-q.Y, 0), 1, fh, merge)
				}
			}
		}
		return e.err == nil, e.err
	}, nil
}

func canDripstonePool(e *featureEditor, p FeaturePosition) bool {
	s := e.block(p)
	if s.Name == "minecraft:water" || s.Name == "minecraft:dripstone_block" || s.Name == "minecraft:pointed_dripstone" {
		return false
	}
	above := e.block(treeOffset(p, 0, 1, 0))
	if above.Name == "minecraft:water" || above.Properties["waterlogged"] == "true" {
		return false
	}
	valid := func(q FeaturePosition) bool {
		s := e.block(q)
		return vegetationTags["dripstone_replaceable_blocks"][s.Name] || s.Name == "minecraft:water" || s.Properties["waterlogged"] == "true"
	}
	for _, dir := range featureHorizontalDirections {
		if !valid(treeOffset(p, dir.X, 0, dir.Z)) {
			return false
		}
	}
	return valid(treeOffset(p, 0, -1, 0))
}
