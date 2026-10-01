package worldgen

import (
	"atlas/internal/blockstate"
	"encoding/json"
)

func compileLakePlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Fluid   json.RawMessage `json:"fluid"`
		Barrier json.RawMessage `json:"barrier"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	fluid, err := decodeFeatureStateProvider(c.Fluid, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return nil, err
	}
	barrier, err := decodeFeatureStateProvider(c.Barrier, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		if origin.Y <= -60 {
			return false, nil
		}
		origin.Y -= 4
		e := featureEditor{world: region}
		var mask [2048]bool
		index := func(x, y, z int32) int { return int((x*16+z)*8 + y) }
		n := r.NextIntBound(4) + 4
		for i := int32(0); i < n; i++ {
			rx, ry, rz := r.NextDouble()*6+3, r.NextDouble()*4+2, r.NextDouble()*6+3
			cx, cy, cz := r.NextDouble()*(16-rx-2)+1+rx/2, r.NextDouble()*(8-ry-4)+2+ry/2, r.NextDouble()*(16-rz-2)+1+rz/2
			for x := int32(1); x < 15; x++ {
				for z := int32(1); z < 15; z++ {
					for y := int32(1); y < 7; y++ {
						dx, dy, dz := (float64(x)-cx)/(rx/2), (float64(y)-cy)/(ry/2), (float64(z)-cz)/(rz/2)
						if dx*dx+dy*dy+dz*dz < 1 {
							mask[index(x, y, z)] = true
						}
					}
				}
			}
		}
		edge := func(x, y, z int32) bool {
			return !mask[index(x, y, z)] && (x < 15 && mask[index(x+1, y, z)] || x > 0 && mask[index(x-1, y, z)] || z < 15 && mask[index(x, y, z+1)] || z > 0 && mask[index(x, y, z-1)] || y < 7 && mask[index(x, y+1, z)] || y > 0 && mask[index(x, y-1, z)])
		}
		state := fluid(r, origin)
		fluidID, err := blockstate.BlockStateID(state.Name, state.Properties)
		if err != nil {
			return false, err
		}
		for x := int32(0); x < 16; x++ {
			for z := int32(0); z < 16; z++ {
				for y := int32(0); y < 8; y++ {
					if !edge(x, y, z) {
						continue
					}
					old := e.block(treeOffset(origin, x, y, z))
					// isLiquid is a block property: waterlogged solids are not liquid blocks.
					if y >= 4 && (old.Name == "minecraft:water" || old.Name == "minecraft:lava") {
						return false, e.err
					}
					if y < 4 {
						solid, err := featureSolid(old)
						if err != nil {
							return false, err
						}
						id, err := blockstate.BlockStateID(old.Name, old.Properties)
						if err != nil {
							return false, err
						}
						if !solid && id != fluidID {
							return false, e.err
						}
					}
				}
			}
		}
		for x := int32(0); x < 16; x++ {
			for z := int32(0); z < 16; z++ {
				for y := int32(0); y < 8; y++ {
					if !mask[index(x, y, z)] {
						continue
					}
					p := treeOffset(origin, x, y, z)
					if vegetationTags["features_cannot_replace"][e.block(p).Name] {
						continue
					}
					s := state
					if y >= 4 {
						s = SurfaceBlockState{Name: "minecraft:cave_air"}
					}
					e.set(p, s)
					if y >= 4 {
						if err := markFeatureAbove(region, p); err != nil {
							return false, err
						}
					}
				}
			}
		}
		wall := barrier(r, origin)
		if !surfaceAir(wall) {
			for x := int32(0); x < 16; x++ {
				for z := int32(0); z < 16; z++ {
					for y := int32(0); y < 8; y++ {
						if !edge(x, y, z) || y >= 4 && r.NextIntBound(2) == 0 {
							continue
						}
						p := treeOffset(origin, x, y, z)
						old := e.block(p)
						solid, err := featureSolid(old)
						if err != nil {
							return false, err
						}
						if solid && !vegetationTags["lava_pool_stone_cannot_replace"][old.Name] {
							e.set(p, wall)
							if err := markFeatureAbove(region, p); err != nil {
								return false, err
							}
						}
					}
				}
			}
		}
		if state.Name == "minecraft:water" && region.biomeAtQuart != nil {
			for x := int32(0); x < 16; x++ {
				for z := int32(0); z < 16; z++ {
					p := treeOffset(origin, x, 4, z)
					qx, qy, qz := BiomeQuartAtBlock(region.biomeSeed, p.X, p.Y, p.Z)
					biome, err := region.biomeAtQuart(qx, qy, qz)
					if err != nil {
						return false, err
					}
					cold, err := BiomeColdAt(biome, p.X, p.Y, p.Z)
					if err != nil {
						return false, err
					}
					old := e.block(p)
					if cold && p.Y >= -64 && p.Y < 320 && old.Name == "minecraft:water" && hasSourceWater(old) && !vegetationTags["features_cannot_replace"][old.Name] {
						e.set(p, SurfaceBlockState{Name: "minecraft:ice"})
					}
				}
			}
		}
		return e.err == nil, e.err
	}, nil
}

// Feature.markAboveForPostProcessing marks up to two contiguous non-air
// positions above the edit.
func markFeatureAbove(region *OverworldFeatureRegion, p FeaturePosition) error {
	for i := 0; i < 2; i++ {
		p.Y++
		s, err := region.BlockAt(p)
		if err != nil {
			return err
		}
		if surfaceAir(s) {
			return nil
		}
		if err := region.MarkPostProcessing(p); err != nil {
			return err
		}
	}
	return nil
}
