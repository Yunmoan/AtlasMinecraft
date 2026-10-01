package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"encoding/json"
	"fmt"
	"strconv"
)

func compileHugeMushroomPlan(definition ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config struct {
		Cap    json.RawMessage `json:"cap_provider"`
		Stem   json.RawMessage `json:"stem_provider"`
		Radius int32           `json:"foliage_radius"`
	}
	if err := json.Unmarshal(definition.Config, &config); err != nil {
		return nil, err
	}
	if config.Radius < 1 || config.Radius > 16 {
		return nil, fmt.Errorf("invalid mushroom foliage radius")
	}
	validate := func(state SurfaceBlockState) error {
		_, err := blockstate.BlockStateID(state.Name, state.Properties)
		return err
	}
	cap, err := decodeFeatureStateProvider(config.Cap, validate)
	if err != nil {
		return nil, err
	}
	stem, err := decodeFeatureStateProvider(config.Stem, validate)
	if err != nil {
		return nil, err
	}
	red := definition.Type == "minecraft:huge_red_mushroom"
	return func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
		height := random.NextIntBound(3) + 4
		if random.NextIntBound(12) == 0 {
			height *= 2
		}
		if origin.Y < -63 || origin.Y+height+1 >= 320 {
			return false, nil
		}
		ground, err := region.BlockAt(treeOffset(origin, 0, -1, 0))
		if err != nil {
			return false, err
		}
		if !forestRockDirt[ground.Name] {
			return false, nil
		}
		// Red's vanilla position check calls its radius function with height=-1,
		// so only the center column is checked. Brown expands above offset 3.
		for y := int32(0); y <= height; y++ {
			radius := int32(0)
			if !red && y > 3 {
				radius = config.Radius
			}
			for x := -radius; x <= radius; x++ {
				for z := -radius; z <= radius; z++ {
					s, err := region.BlockAt(treeOffset(origin, x, y, z))
					if err != nil {
						return false, err
					}
					if !surfaceAir(s) && !treeTags["leaves"][s.Name] {
						return false, nil
					}
				}
			}
		}
		opaque := func(p FeaturePosition) (bool, error) {
			state, err := region.BlockAt(p)
			if err != nil {
				return false, err
			}
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				return false, err
			}
			rules, err := blockrules.Lookup(id)
			return rules.SolidRender, err
		}
		first := height
		if red {
			first = height - 3
		}
		for y := first; y <= height; y++ {
			radius := config.Radius
			if red && y == height {
				radius--
			}
			for x := -radius; x <= radius; x++ {
				for z := -radius; z <= radius; z++ {
					west, east, north, south := x == -radius, x == radius, z == -radius, z == radius
					edgeX, edgeZ := west || east, north || south
					if !red && edgeX && edgeZ || red && y < height && edgeX == edgeZ {
						continue
					}
					p := treeOffset(origin, x, y, z)
					blocked, err := opaque(p)
					if err != nil {
						return false, err
					}
					if blocked {
						continue
					}
					state := cap(random, origin)
					_, w := state.Properties["west"]
					_, e := state.Properties["east"]
					_, n := state.Properties["north"]
					_, s := state.Properties["south"]
					_, u := state.Properties["up"]
					if w && e && n && s && (!red || u) {
						if red {
							bound := config.Radius - 2
							west, east, north, south = x < -bound, x > bound, z < -bound, z > bound
							state = stateWithProperty(state, "up", strconv.FormatBool(y >= height-1))
						} else {
							west = west || edgeZ && x == 1-radius
							east = east || edgeZ && x == radius-1
							north = north || edgeX && z == 1-radius
							south = south || edgeX && z == radius-1
						}
						state = stateWithProperty(state, "west", strconv.FormatBool(west))
						state = stateWithProperty(state, "east", strconv.FormatBool(east))
						state = stateWithProperty(state, "north", strconv.FormatBool(north))
						state = stateWithProperty(state, "south", strconv.FormatBool(south))
					}
					if err = setFeatureBlock(region, p, state); err != nil {
						return false, err
					}
				}
			}
		}
		for y := int32(0); y < height; y++ {
			p := treeOffset(origin, 0, y, 0)
			blocked, err := opaque(p)
			if err != nil {
				return false, err
			}
			if blocked {
				continue
			}
			if err = setFeatureBlock(region, p, stem(random, origin)); err != nil {
				return false, err
			}
		}
		return true, nil
	}, nil
}
