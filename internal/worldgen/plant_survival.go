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
	case "minecraft:fire":
		return featureFireSurvives(world, p, below)
	case "minecraft:soul_fire":
		return below.Name == "minecraft:soul_sand" || below.Name == "minecraft:soul_soil", nil
	case "minecraft:crimson_roots", "minecraft:warped_roots", "minecraft:crimson_fungus", "minecraft:warped_fungus", "minecraft:nether_sprouts":
		return featureNetherPlantSurvives(state.Name, below), nil
	case "minecraft:brown_mushroom", "minecraft:red_mushroom":
		return featureMushroomSurvives(world, p, below)
	case "minecraft:seagrass":
		e := featureEditor{world: world}
		return oceanSupport(&e, p, false), e.err
	case "minecraft:hanging_roots":
		above, err := world.BlockAt(treeOffset(p, 0, 1, 0))
		if err != nil {
			return false, err
		}
		id, err := blockstate.BlockStateID(above.Name, above.Properties)
		if err != nil {
			return false, err
		}
		rule, err := blockrules.Lookup(id)
		return rule.FaceSturdy[0], err
	case "minecraft:azalea", "minecraft:flowering_azalea":
		return PlantGroundSurvives("minecraft:short_grass", below.Name) || below.Name == "minecraft:clay", nil
	case "minecraft:moss_carpet":
		return !surfaceAir(below), nil
	case "minecraft:small_dripleaf":
		current, err := world.BlockAt(p)
		return vegetationTags["small_dripleaf_placeable"][below.Name] || hasSourceWater(current) && PlantGroundSurvives("minecraft:short_grass", below.Name), err
	case "minecraft:spore_blossom":
		above, err := world.BlockAt(treeOffset(p, 0, 1, 0))
		if err != nil {
			return false, err
		}
		id, err := blockstate.BlockStateID(above.Name, above.Properties)
		if err != nil {
			return false, err
		}
		rule, err := blockrules.Lookup(id)
		if err != nil {
			return false, err
		}
		current, err := world.BlockAt(p)
		return rule.BottomCenter && !featureFluid(current), err
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
	case "minecraft:oak_sapling", "minecraft:birch_sapling", "minecraft:spruce_sapling", "minecraft:acacia_sapling", "minecraft:jungle_sapling", "minecraft:dark_oak_sapling", "minecraft:cherry_sapling":
		return PlantGroundSurvives(state.Name, below.Name), nil
	case "minecraft:mangrove_propagule":
		if state.Properties["hanging"] == "true" {
			above, err := world.BlockAt(treeOffset(p, 0, 1, 0))
			return above.Name == "minecraft:mangrove_leaves", err
		}
		return PlantGroundSurvives(state.Name, below.Name) || below.Name == "minecraft:clay", nil
	default:
		if SupportedPlant(state.Name) {
			return PlantGroundSurvives(state.Name, below.Name), nil
		}
		if state.Name == "minecraft:pumpkin" || state.Name == "minecraft:melon" {
			return true, nil
		}
		return false, fmt.Errorf("unsupported survival check for %s", state.Name)
	}
}
