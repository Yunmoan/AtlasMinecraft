package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"fmt"
)

func featureSolid(state SurfaceBlockState) (bool, error) {
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return false, err
	}
	rule, err := blockrules.Lookup(id)
	return rule.Solid, err
}

func featureFluid(state SurfaceBlockState) bool {
	return state.Properties["waterlogged"] == "true" || state.Name == "minecraft:water" || state.Name == "minecraft:lava"
}

func featurePlantSurvives(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState) (bool, error) {
	below, err := world.BlockAt(treeOffset(p, 0, -1, 0))
	if err != nil {
		return false, err
	}
	switch state.Name {
	case "minecraft:lily_pad":
		current, err := world.BlockAt(p)
		if err != nil {
			return false, err
		}
		return (hasSourceWater(below) || below.Name == "minecraft:ice" || below.Name == "minecraft:frosted_ice") && !featureFluid(current), nil
	case "minecraft:cactus":
		for _, offset := range [][2]int32{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
			neighbor, err := world.BlockAt(treeOffset(p, offset[0], 0, offset[1]))
			if err != nil {
				return false, err
			}
			solid, err := featureSolid(neighbor)
			if err != nil {
				return false, err
			}
			if solid || neighbor.Name == "minecraft:lava" {
				return false, nil
			}
		}
		above, err := world.BlockAt(treeOffset(p, 0, 1, 0))
		if err != nil {
			return false, err
		}
		return (below.Name == "minecraft:cactus" || below.Name == "minecraft:sand" || below.Name == "minecraft:red_sand") && above.Name != "minecraft:water" && above.Name != "minecraft:lava", nil
	case "minecraft:sugar_cane":
		if below.Name == "minecraft:sugar_cane" {
			return true, nil
		}
		if !vegetationTags["dirt"][below.Name] && below.Name != "minecraft:sand" && below.Name != "minecraft:red_sand" {
			return false, nil
		}
		for _, offset := range [][2]int32{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
			neighbor, err := world.BlockAt(treeOffset(p, offset[0], -1, offset[1]))
			if err != nil {
				return false, err
			}
			if neighbor.Name == "minecraft:water" || hasSourceWater(neighbor) || neighbor.Name == "minecraft:frosted_ice" {
				return true, nil
			}
		}
		return false, nil
	case "minecraft:oak_sapling", "minecraft:birch_sapling", "minecraft:spruce_sapling":
		return forestRockDirt[below.Name] || below.Name == "minecraft:farmland", nil
	default:
		if SupportedPlant(state.Name) {
			return PlantGroundSurvives(state.Name, below.Name), nil
		}
		if state.Name == "minecraft:pumpkin" {
			return true, nil
		}
		return false, fmt.Errorf("unsupported survival check for %s", state.Name)
	}
}
