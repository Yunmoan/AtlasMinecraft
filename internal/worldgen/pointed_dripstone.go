package worldgen

import (
	"encoding/json"
	"fmt"
)

type pointedDripstoneConfiguration struct {
	Taller  float32 `json:"chance_of_taller_dripstone"`
	Spread  float32 `json:"chance_of_directional_spread"`
	Radius2 float32 `json:"chance_of_spread_radius2"`
	Radius3 float32 `json:"chance_of_spread_radius3"`
}

var featureDirections = [...]FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}}
var featureHorizontalDirections = [...]FeaturePosition{{0, 0, -1}, {1, 0, 0}, {0, 0, 1}, {-1, 0, 0}}

func dripstoneSupport(state SurfaceBlockState) bool {
	return state.Name == "minecraft:dripstone_block" || vegetationTags["dripstone_replaceable_blocks"][state.Name]
}

func compilePointedDripstonePlan(definition ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config pointedDripstoneConfiguration
	if err := json.Unmarshal(definition.Config, &config); err != nil {
		return nil, err
	}
	for _, value := range []float32{config.Taller, config.Spread, config.Radius2, config.Radius3} {
		if value < 0 || value > 1 {
			return nil, fmt.Errorf("invalid pointed dripstone probability")
		}
	}
	return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
		return placePointedDripstone(region, random, p, config)
	}, nil
}

func placePointedDripstone(world FeatureBlockWorld, random *WorldgenRandom, origin FeaturePosition, config pointedDripstoneConfiguration) (bool, error) {
	e := featureEditor{world: world}
	above, below := dripstoneSupport(e.block(treeOffset(origin, 0, 1, 0))), dripstoneSupport(e.block(treeOffset(origin, 0, -1, 0)))
	if e.err != nil {
		return false, e.err
	}
	if !above && !below {
		return false, nil
	}
	dy := int32(1)
	if above {
		dy = -1
	}
	if above && below {
		if random.NextBoolean() {
			dy = -1
		} else {
			dy = 1
		}
	}
	root := treeOffset(origin, 0, -dy, 0)
	replace := func(p FeaturePosition) {
		if vegetationTags["dripstone_replaceable_blocks"][e.block(p).Name] {
			e.set(p, SurfaceBlockState{Name: "minecraft:dripstone_block"})
		}
	}
	replace(root)
	for _, direction := range featureHorizontalDirections {
		if random.NextFloat() > config.Spread {
			continue
		}
		p := treeOffset(root, direction.X, 0, direction.Z)
		replace(p)
		if random.NextFloat() > config.Radius2 {
			continue
		}
		direction = featureDirections[random.NextIntBound(6)]
		p = treeOffset(p, direction.X, direction.Y, direction.Z)
		replace(p)
		if random.NextFloat() > config.Radius3 {
			continue
		}
		direction = featureDirections[random.NextIntBound(6)]
		p = treeOffset(p, direction.X, direction.Y, direction.Z)
		replace(p)
	}
	height := int32(1)
	if random.NextFloat() < config.Taller {
		state := e.block(treeOffset(origin, 0, dy, 0))
		if surfaceAir(state) || state.Name == "minecraft:water" {
			height = 2
		}
	}
	if dripstoneSupport(e.block(root)) {
		for y := int32(0); y < height; y++ {
			p := treeOffset(origin, 0, dy*y, 0)
			old := e.block(p)
			thickness := "tip"
			if y < height-1 {
				thickness = "frustum"
			}
			direction := "up"
			if dy < 0 {
				direction = "down"
			}
			waterlogged := "false"
			if old.Name == "minecraft:water" || old.Properties["waterlogged"] == "true" {
				waterlogged = "true"
			}
			e.set(p, SurfaceBlockState{Name: "minecraft:pointed_dripstone", Properties: map[string]string{"vertical_direction": direction, "thickness": thickness, "waterlogged": waterlogged}})
		}
	}
	return e.err == nil, e.err
}

func compileUnderwaterMagmaPlan(definition ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config struct {
		Search      int32   `json:"floor_search_range"`
		Radius      int32   `json:"placement_radius_around_floor"`
		Probability float32 `json:"placement_probability_per_valid_position"`
	}
	if err := json.Unmarshal(definition.Config, &config); err != nil {
		return nil, err
	}
	if config.Search < 1 || config.Search > 512 || config.Radius < 0 || config.Radius > 16 || config.Probability < 0 || config.Probability > 1 {
		return nil, fmt.Errorf("invalid underwater magma configuration")
	}
	return func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if e.block(origin).Name != "minecraft:water" {
			return false, e.err
		}
		floor := origin
		found := false
		for i := int32(1); i < config.Search; i++ {
			floor.Y--
			if e.block(floor).Name != "minecraft:water" {
				found = true
				break
			}
		}
		if !found {
			return false, e.err
		}
		placed := false
		// BlockPos.betweenClosed iterates X fastest, then Y, then Z.
		for z := -config.Radius; z <= config.Radius; z++ {
			for y := -config.Radius; y <= config.Radius; y++ {
				for x := -config.Radius; x <= config.Radius; x++ {
					if random.NextFloat() >= config.Probability {
						continue
					}
					p := treeOffset(floor, x, y, z)
					invalid := func(p FeaturePosition) bool {
						state := e.block(p)
						return surfaceAir(state) || state.Name == "minecraft:water"
					}
					if invalid(p) || invalid(treeOffset(p, 0, -1, 0)) {
						continue
					}
					blocked := false
					for _, direction := range featureHorizontalDirections {
						if invalid(treeOffset(p, direction.X, 0, direction.Z)) {
							blocked = true
							break
						}
					}
					if !blocked {
						e.set(p, SurfaceBlockState{Name: "minecraft:magma_block"})
						placed = true
					}
				}
			}
		}
		return placed && e.err == nil, e.err
	}, nil
}
