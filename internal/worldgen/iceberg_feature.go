package worldgen

import (
	"encoding/json"
	"math"
)

func icebergState(s SurfaceBlockState) bool {
	return s.Name == "minecraft:packed_ice" || s.Name == "minecraft:snow_block" || s.Name == "minecraft:blue_ice"
}

func icebergRoundRadius(r *WorldgenRandom, y, height, radius int32) int32 {
	factor := float32(3.5) - r.NextFloat()
	f := (float32(1) - float32(y*y)/(float32(height)*factor)) * float32(radius)
	if height > 15+r.NextIntBound(5) {
		q := y
		if y < 3+r.NextIntBound(6) {
			q = y / 2
		}
		f = (float32(1) - float32(q)/(float32(height)*factor*float32(0.4))) * float32(radius)
	}
	return int32(math.Ceil(float64(f / float32(2))))
}

func icebergSteepRadius(r *WorldgenRandom, y, height, radius int32) int32 {
	factor := float32(1) + r.NextFloat()/float32(2)
	f := (float32(1) - float32(y)/(float32(height)*factor)) * float32(radius)
	return int32(math.Ceil(float64(f / float32(2))))
}

func icebergEllipseDistance(x, z, cx, cz, a, b int32, angle float64) float64 {
	dx, dz := float64(x-cx), float64(z-cz)
	sin, cos := featureSinCos(angle)
	u := (float64(dx*cos) - float64(dz*sin)) / float64(a)
	v := (float64(dx*sin) + float64(dz*cos)) / float64(b)
	return float64(u*u) + float64(v*v) - 1
}

func compileIcebergPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config struct {
		State SurfaceBlockState `json:"state"`
	}
	if err := json.Unmarshal(def.Config, &config); err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) (bool, error) {
		return PlaceIcebergFeature(region, r, p, config.State)
	}, nil
}

func PlaceIcebergFeature(world FeatureBlockWorld, r *WorldgenRandom, origin FeaturePosition, state SurfaceBlockState) (bool, error) {
	e := featureEditor{world: world}
	origin.Y = 63
	snow := r.NextDouble() > 0.7
	angle := r.NextDouble() * 2 * math.Pi
	ellipseRadius := 11 - r.NextIntBound(5)
	smallRadius := 3 + r.NextIntBound(3)
	ellipse := r.NextDouble() > 0.7
	var height int32
	if ellipse {
		height = 6 + r.NextIntBound(6)
	} else {
		height = 3 + r.NextIntBound(15)
	}
	if !ellipse && r.NextDouble() > 0.9 {
		height += 7 + r.NextIntBound(19)
	}
	depth := min(height+r.NextIntBound(11), int32(18))
	radius := min(height+r.NextIntBound(7)-r.NextIntBound(5), int32(11))
	width := int32(11)
	if ellipse {
		width = ellipseRadius
	}
	generate := func(x, y, z, layerRadius, layerWidth, totalHeight int32) {
		var distance float64
		if ellipse {
			b := smallRadius
			if y > 0 && totalHeight-y <= 3 {
				b -= 4 - (totalHeight - y)
			}
			distance = icebergEllipseDistance(x, z, 0, 0, layerWidth, b, angle)
		} else {
			f := float32(10) * max(float32(0.2), min(float32(0.8), r.NextFloat())) / float32(layerRadius)
			distance = float64(f) + float64(x*x) + float64(z*z) - float64(layerRadius*layerRadius)
		}
		if distance >= 0 || math.IsNaN(distance) {
			return
		}
		threshold := -0.5
		if !ellipse {
			threshold = float64(-6 - r.NextIntBound(3))
		}
		if distance > threshold && r.NextDouble() > 0.9 {
			return
		}
		p := treeOffset(origin, x, y, z)
		s := e.block(p)
		if surfaceAir(s) || s.Name == "minecraft:snow_block" || s.Name == "minecraft:ice" || s.Name == "minecraft:water" {
			useSnow := true
			if ellipse {
				useSnow = r.NextDouble() > 0.05
			}
			divisor := int32(2)
			if ellipse {
				divisor = 3
			}
			if snow && s.Name != "minecraft:water" && float64(totalHeight-y) <= float64(r.NextIntBound(max(int32(1), totalHeight/divisor)))+float64(totalHeight)*0.6 && useSnow {
				e.set(p, SurfaceBlockState{Name: "minecraft:snow_block"})
			} else {
				e.set(p, state)
			}
		}
	}
	for x := -width; x < width; x++ {
		for z := -width; z < width; z++ {
			for y := int32(0); y < height; y++ {
				var layer int32
				if ellipse {
					layer = int32(math.Ceil(float64((float32(1) - float32(y*y)/float32(height)) * float32(radius) / float32(2))))
				} else {
					layer = icebergRoundRadius(r, y, height, radius)
				}
				if !ellipse && x >= layer {
					continue
				}
				generate(x, y, z, layer, width, height)
			}
		}
	}
	// Vanilla smooths the above-water body before generating the submerged body.
	smoothRadius := radius / 2
	if ellipse {
		smoothRadius = ellipseRadius
	}
	for x := -smoothRadius; x <= smoothRadius; x++ {
		for z := -smoothRadius; z <= smoothRadius; z++ {
			for y := int32(0); y <= height; y++ {
				p := treeOffset(origin, x, y, z)
				s := e.block(p)
				if !icebergState(s) && s.Name != "minecraft:snow" {
					continue
				}
				if surfaceAir(e.block(treeOffset(p, 0, -1, 0))) {
					e.set(p, SurfaceBlockState{Name: "minecraft:air"})
					e.set(treeOffset(p, 0, 1, 0), SurfaceBlockState{Name: "minecraft:air"})
					continue
				}
				if icebergState(s) {
					neighbors := 0
					for _, d := range []int{2, 3, 4, 5} {
						if !icebergState(e.block(directionPosition(p, d))) {
							neighbors++
						}
					}
					if neighbors >= 3 {
						e.set(p, SurfaceBlockState{Name: "minecraft:air"})
					}
				}
			}
		}
	}
	for x := -width; x < width; x++ {
		for z := -width; z < width; z++ {
			for y := int32(-1); y > -depth; y-- {
				layerWidth := width
				if ellipse {
					layerWidth = int32(math.Ceil(float64(float32(width) * (float32(1) - float32(y*y)/(float32(depth)*float32(8))))))
				}
				layer := icebergSteepRadius(r, -y, depth, radius)
				if x >= layer {
					continue
				}
				generate(x, y, z, layer, layerWidth, depth)
			}
		}
	}
	cutout := r.NextDouble()
	threshold := 0.7
	if ellipse {
		threshold = 0.1
	}
	if cutout > threshold {
		sx, sz := int32(1), int32(1)
		if r.NextBoolean() {
			sx = -1
		}
		if r.NextBoolean() {
			sz = -1
		}
		cx := r.NextIntBound(max(radius/2-2, int32(1)))
		if r.NextBoolean() {
			cx = radius/2 + 1 - r.NextIntBound(max(radius-radius/2-1, int32(1)))
		}
		cz := r.NextIntBound(max(radius/2-2, int32(1)))
		if r.NextBoolean() {
			cz = radius/2 + 1 - r.NextIntBound(max(radius-radius/2-1, int32(1)))
		}
		if ellipse {
			cx = r.NextIntBound(max(ellipseRadius-5, int32(1)))
			cz = cx
		}
		cx *= sx
		cz *= sz
		cutAngle := angle + math.Pi/2
		if !ellipse {
			cutAngle = r.NextDouble() * 2 * math.Pi
		}
		carve := func(layer, y int32, water bool) {
			a := layer + 1 + ellipseRadius/3
			b := min(layer-3, int32(3)) + smallRadius/2 - 1
			for x := -a; x < a; x++ {
				for z := -a; z < a; z++ {
					if !(icebergEllipseDistance(x, z, cx, cz, a, b, cutAngle) < 0) {
						continue
					}
					p := treeOffset(origin, x, y, z)
					s := e.block(p)
					if !icebergState(s) && s.Name != "minecraft:snow_block" {
						continue
					}
					if water {
						e.set(p, SurfaceBlockState{Name: "minecraft:water"})
					} else {
						e.set(p, SurfaceBlockState{Name: "minecraft:air"})
						above := treeOffset(p, 0, 1, 0)
						if e.block(above).Name == "minecraft:snow" {
							e.set(above, SurfaceBlockState{Name: "minecraft:air"})
						}
					}
				}
			}
		}
		for y := int32(0); y < height-3; y++ {
			carve(icebergRoundRadius(r, y, height, radius), y, false)
		}
		for y := int32(-1); y > -height+r.NextIntBound(5); y-- {
			carve(icebergSteepRadius(r, -y, height, radius), y, true)
		}
	}
	return true, e.err
}
