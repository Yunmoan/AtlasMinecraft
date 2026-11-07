package worldgen

import "fmt"

// FluidContainerWorld is the server-side level contract of placeLiquid.
// Candle smoke uses Level.addParticle, which is a no-op on ServerLevel.
type FluidContainerWorld interface {
	SetFluidBlock(FeaturePosition, SurfaceBlockState, int32) error
	ScheduleFluid(FeaturePosition, FluidType, int32) error
	FluidSound(FeaturePosition, string) error
	FluidCampfireDowse(FeaturePosition) error
	FluidBlockChangeEvent(FeaturePosition) error
}

// PlaceFluidInBlock executes the original container callback, including its
// acceptance result. It must receive the current state and an actual registered
// LiquidBlockContainer; flowing and source water are distinct fluid objects.
func PlaceFluidInBlock(world FluidContainerWorld, p FeaturePosition, state SurfaceBlockState, fluid FluidState) (bool, error) {
	rules, err := fluidRulesForBlock(state)
	if err != nil {
		return false, err
	}
	if !rules.container() {
		return false, fmt.Errorf("not a fluid container: %s", state.Name)
	}
	switch state.Name {
	case "minecraft:kelp", "minecraft:kelp_plant", "minecraft:seagrass", "minecraft:tall_seagrass":
		return false, nil
	}
	if state.Properties["type"] == "double" || state.Properties["waterlogged"] == "true" || fluid.Type != FluidWater {
		return false, nil
	}
	next := stateWithProperty(state, "waterlogged", "true")
	flags := int32(3)
	lit := state.Properties["lit"] == "true"
	campfire := state.Name == "minecraft:campfire" || state.Name == "minecraft:soul_campfire"
	if campfire {
		if lit {
			if err := world.FluidSound(p, "minecraft:entity.generic.extinguish_fire"); err != nil {
				return false, err
			}
			if err := world.FluidCampfireDowse(p); err != nil {
				return false, err
			}
			if err := world.FluidBlockChangeEvent(p); err != nil {
				return false, err
			}
		}
		next = stateWithProperty(next, "lit", "false")
	} else if lit {
		// The only remaining lit container class is CandleBlock. It calls
		// AbstractCandleBlock.extinguish with the already waterlogged state.
		next = stateWithProperty(next, "lit", "false")
		flags = 11
	}
	if err := world.SetFluidBlock(p, next, flags); err != nil {
		return false, err
	}
	if lit && !campfire {
		if err := world.FluidSound(p, "minecraft:block.candle.extinguish"); err != nil {
			return false, err
		}
		if err := world.FluidBlockChangeEvent(p); err != nil {
			return false, err
		}
	}
	if err := world.ScheduleFluid(p, fluid.Type, 5); err != nil {
		return false, err
	}
	return true, nil
}
