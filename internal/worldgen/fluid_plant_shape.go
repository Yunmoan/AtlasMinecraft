package worldgen

import "strings"

// Only families whose updateShape and canSurvive are ported here. In particular
// cactus, cane, growing vines, bamboo stalks and leaves schedule block ticks;
// checking their support and deleting them immediately would change vanilla.
func runtimePlantShape(name string) bool {
	if SupportedPlant(name) {
		return true
	}
	switch name {
	case "minecraft:oak_sapling", "minecraft:spruce_sapling", "minecraft:birch_sapling", "minecraft:jungle_sapling", "minecraft:acacia_sapling", "minecraft:dark_oak_sapling", "minecraft:cherry_sapling",
		"minecraft:mangrove_propagule", "minecraft:azalea", "minecraft:flowering_azalea",
		"minecraft:crimson_roots", "minecraft:warped_roots", "minecraft:crimson_fungus", "minecraft:warped_fungus", "minecraft:nether_sprouts", "minecraft:nether_wart",
		"minecraft:brown_mushroom", "minecraft:red_mushroom", "minecraft:torchflower", "minecraft:wither_rose",
		"minecraft:wheat", "minecraft:carrots", "minecraft:potatoes", "minecraft:beetroots", "minecraft:pumpkin_stem", "minecraft:melon_stem", "minecraft:attached_pumpkin_stem", "minecraft:attached_melon_stem",
		"minecraft:seagrass", "minecraft:tall_seagrass", "minecraft:lily_pad", "minecraft:small_dripleaf", "minecraft:hanging_roots", "minecraft:spore_blossom", "minecraft:bamboo_sapling":
		return true
	}
	return false
}

func runtimePlantWaterTick(world FeatureBlockWorld, p FeaturePosition) error {
	if scheduler, ok := world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
		return scheduler.ScheduleFluidUpdate(p)
	}
	return nil
}

func updateRuntimePlantShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	air := structureState("air")
	// These two subclasses schedule before DoublePlantBlock/BushBlock removes
	// an unsupported plant. HangingRootsBlock instead removes before scheduling.
	if (state.Name == "minecraft:small_dripleaf" || state.Name == "minecraft:mangrove_propagule") && state.Properties["waterlogged"] == "true" {
		if err := runtimePlantWaterTick(world, p); err != nil {
			return state, err
		}
	}
	if state.Name == "minecraft:hanging_roots" || state.Name == "minecraft:spore_blossom" {
		if direction == "up" {
			ok, err := featurePlantSurvives(world, p, state)
			if err != nil {
				return state, err
			}
			if !ok {
				return air, nil
			}
		}
		if state.Properties["waterlogged"] == "true" {
			return state, runtimePlantWaterTick(world, p)
		}
		return state, nil
	}
	if strings.HasPrefix(state.Name, "minecraft:attached_") {
		fruit := strings.TrimSuffix(strings.TrimPrefix(state.Name, "minecraft:attached_"), "_stem")
		if direction == state.Properties["facing"] && neighbor.Name != "minecraft:"+fruit {
			return stateWithProperty(structureState(fruit+"_stem"), "age", "7"), nil
		}
	}
	double := DoublePlant(state.Name) || state.Name == "minecraft:tall_seagrass"
	if double {
		lower := state.Properties["half"] == "lower"
		if direction == "up" && lower || direction == "down" && !lower {
			if neighbor.Name != state.Name || neighbor.Properties["half"] == state.Properties["half"] {
				return air, nil
			}
		}
	}
	survives, err := runtimePlantSurvives(world, p, state, double)
	if err != nil {
		return state, err
	}
	if !survives {
		return air, nil
	}
	if state.Name == "minecraft:seagrass" {
		return state, runtimePlantWaterTick(world, p)
	}
	if state.Name == "minecraft:bamboo_sapling" && direction == "up" && neighbor.Name == "minecraft:bamboo" {
		// Original callback writes a stalk as a side effect, but returns the
		// old sapling state. NeighborUpdater subsequently observes that write.
		if err := world.SetBlock(p, structureState("bamboo")); err != nil {
			return state, err
		}
	}
	return state, nil
}

func runtimePlantSurvives(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, double bool) (bool, error) {
	below, err := world.BlockAt(treeOffset(p, 0, -1, 0))
	if err != nil {
		return false, err
	}
	if double && state.Properties["half"] == "upper" {
		return below.Name == state.Name && below.Properties["half"] == "lower", nil
	}
	switch state.Name {
	case "minecraft:bamboo_sapling":
		return vegetationTags["bamboo_plantable_on"][below.Name], nil
	case "minecraft:nether_wart":
		return below.Name == "minecraft:soul_sand", nil
	case "minecraft:wheat", "minecraft:carrots", "minecraft:potatoes", "minecraft:beetroots":
		brightness, err := featureRawBrightness(world, p, 0)
		return brightness >= 8 && below.Name == "minecraft:farmland", err
	case "minecraft:pumpkin_stem", "minecraft:melon_stem", "minecraft:attached_pumpkin_stem", "minecraft:attached_melon_stem":
		return below.Name == "minecraft:farmland", nil
	case "minecraft:torchflower":
		return PlantGroundSurvives(state.Name, below.Name), nil
	case "minecraft:wither_rose":
		return PlantGroundSurvives(state.Name, below.Name) || below.Name == "minecraft:netherrack" || below.Name == "minecraft:soul_sand" || below.Name == "minecraft:soul_soil", nil
	case "minecraft:tall_seagrass":
		e := featureEditor{world: world}
		return oceanSupport(&e, p, false), e.err
	}
	return featurePlantSurvives(world, p, state)
}
