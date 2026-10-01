package worldgen

import (
	"encoding/json"
	"fmt"
	"math"
)

type largeDripstoneConfig struct {
	Search           int32           `json:"floor_to_ceiling_search_range"`
	Radius           json.RawMessage `json:"column_radius"`
	Scale            json.RawMessage `json:"height_scale"`
	Ratio            float32         `json:"max_column_radius_to_cave_height_ratio"`
	CeilingBluntness json.RawMessage `json:"stalactite_bluntness"`
	FloorBluntness   json.RawMessage `json:"stalagmite_bluntness"`
	Wind             json.RawMessage `json:"wind_speed"`
	WindRadius       int32           `json:"min_radius_for_wind"`
	WindBluntness    float32         `json:"min_bluntness_for_wind"`
}

type largeDripstone struct {
	root             FeaturePosition
	up               bool
	radius           int32
	bluntness, scale float64
}
type dripstoneWind struct {
	active bool
	y      int32
	x, z   float64
}

func (w dripstoneWind) offset(p FeaturePosition) FeaturePosition {
	if !w.active {
		return p
	}
	d := float64(w.y - p.Y)
	return treeOffset(p, int32(math.Floor(w.x*d)), 0, int32(math.Floor(w.z*d)))
}

func (d *largeDripstone) height(distance float32) int32 {
	x := max(float64(distance), d.bluntness) / float64(d.radius) * 0.384
	v := d.scale * (0.75*math.Pow(x, 4.0/3.0) - math.Pow(x, 2.0/3.0) - (1.0/3.0)*math.Log(x))
	return int32(max(v, 0) / 0.384 * float64(d.radius))
}

func (d *largeDripstone) findRoot(e *featureEditor, w dripstoneWind) bool {
	for d.radius > 1 {
		p := d.root
		steps := min(int32(10), d.height(0))
		for i := int32(0); i < steps; i++ {
			if e.block(p).Name == "minecraft:lava" {
				return false
			}
			q := w.offset(p)
			space := func(p FeaturePosition) bool { s := e.block(p); return dripstoneSpace(s) || s.Name == "minecraft:lava" }
			valid := !space(q)
			for angle := float32(0); valid && angle < float32(math.Pi)*2; angle += 6 / float32(d.radius) {
				x, z := int32(carverCos(angle)*float32(d.radius)), int32(carverSin(angle)*float32(d.radius))
				if space(treeOffset(q, x, 0, z)) {
					valid = false
				}
			}
			if valid {
				d.root = p
				return true
			}
			if d.up {
				p.Y--
			} else {
				p.Y++
			}
		}
		d.radius /= 2
	}
	return false
}

func (d *largeDripstone) place(e *featureEditor, region *OverworldFeatureRegion, r *WorldgenRandom, w dripstoneWind) {
	for x := -d.radius; x <= d.radius; x++ {
		for z := -d.radius; z <= d.radius; z++ {
			distance := float32(math.Sqrt(float64(float32(x*x + z*z))))
			if distance > float32(d.radius) {
				continue
			}
			h := d.height(distance)
			if h <= 0 {
				continue
			}
			if float64(r.NextFloat()) < 0.2 {
				h = int32(float32(h) * (0.8 + r.NextFloat()*(1.0-float32(0.8))))
			}
			p := treeOffset(d.root, x, 0, z)
			limit := int32(math.MaxInt32)
			if d.up {
				var err error
				limit, err = region.HeightmapAt("WORLD_SURFACE_WG", p.X, p.Z)
				if err != nil {
					e.err = err
					return
				}
			}
			placed := false
			for i := int32(0); i < h && p.Y < limit; i++ {
				q := w.offset(p)
				s := e.block(q)
				if dripstoneSpace(s) || s.Name == "minecraft:lava" {
					placed = true
					e.set(q, SurfaceBlockState{Name: "minecraft:dripstone_block"})
				} else if placed && vegetationTags["dripstone_replaceable_blocks"][s.Name] {
					break
				}
				if d.up {
					p.Y++
				} else {
					p.Y--
				}
			}
		}
	}
}

func compileLargeDripstonePlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c largeDripstoneConfig
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Search < 1 || c.Search > 512 || c.Ratio < 0 || c.Ratio > 1 || c.WindRadius < 0 || c.WindBluntness < 0 {
		return nil, fmt.Errorf("invalid large dripstone configuration")
	}
	// The radius provider's declared bounds clamp the cave-height-based maximum.
	var bounds struct {
		Min int32 `json:"min_inclusive"`
		Max int32 `json:"max_inclusive"`
	}
	if err := json.Unmarshal(c.Radius, &bounds); err != nil {
		return nil, err
	}
	if bounds.Min < 1 || bounds.Max < bounds.Min || bounds.Max > 128 {
		return nil, fmt.Errorf("invalid large dripstone radius")
	}
	scale, err := decodeFeatureFloatProvider(c.Scale, 0, 20)
	if err != nil {
		return nil, err
	}
	ceiling, err := decodeFeatureFloatProvider(c.CeilingBluntness, 0.01, 10)
	if err != nil {
		return nil, err
	}
	floor, err := decodeFeatureFloatProvider(c.FloorBluntness, 0.01, 10)
	if err != nil {
		return nil, err
	}
	wind, err := decodeFeatureFloatProvider(c.Wind, 0, 10)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		col, ok := scanFeatureColumn(&e, p, c.Search, dripstoneSpace, func(s SurfaceBlockState) bool { return dripstoneSupport(s) || s.Name == "minecraft:lava" })
		if !ok || !col.hasFloor || !col.hasCeiling || col.ceiling-col.floor-1 < 4 {
			return false, e.err
		}
		maxRadius := max(bounds.Min, min(bounds.Max, int32(float32(col.ceiling-col.floor-1)*c.Ratio)))
		radius := randomBetweenInclusive(r, bounds.Min, maxRadius)
		top := largeDripstone{root: treeOffset(p, 0, col.ceiling-1-p.Y, 0), radius: radius, bluntness: float64(ceiling(r)), scale: float64(scale(r))}
		bottom := largeDripstone{root: treeOffset(p, 0, col.floor+1-p.Y, 0), up: true, radius: radius, bluntness: float64(floor(r)), scale: float64(scale(r))}
		var w dripstoneWind
		if radius >= c.WindRadius && top.bluntness >= float64(c.WindBluntness) && bottom.bluntness >= float64(c.WindBluntness) {
			speed := wind(r)
			angle := r.NextFloat() * float32(math.Pi)
			w = dripstoneWind{active: true, y: p.Y, x: float64(carverCos(angle) * speed), z: float64(carverSin(angle) * speed)}
		}
		topOK, bottomOK := top.findRoot(&e, w), bottom.findRoot(&e, w)
		if topOK {
			top.place(&e, region, r, w)
		}
		if bottomOK {
			bottom.place(&e, region, r, w)
		}
		return e.err == nil, e.err
	}, nil
}
