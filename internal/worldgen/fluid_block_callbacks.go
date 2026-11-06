package worldgen

// FluidTickDelay is the registered fluid's ordinary tick delay. Lava's
// randomized height-increase delay is a separate FlowingFluid.tick rule.
func FluidTickDelay(kind FluidType, ultraWarm bool) int32 {
	if kind == FluidEmpty {
		return 0
	}
	if kind.same(FluidLava) {
		if ultraWarm {
			return 10
		}
		return 30
	}
	return 5
}

// NotifyLiquidBlock implements LiquidBlock.onPlace and neighborChanged. The
// live adapter calls onPlace for every actual state change, including another
// liquid level of the same block type (LevelChunk.setBlockState).
func NotifyLiquidBlock(world FluidTickWorld, p FeaturePosition, state SurfaceBlockState, settings FluidSettings) error {
	if state.Name != "minecraft:water" && state.Name != "minecraft:lava" {
		return nil
	}
	e := &fluidStep{world: world, settings: settings}
	fluid := e.rules(state).state
	if fluid.Type.same(FluidLava) {
		soulSoil := e.block(fluidNeighbor(p, 0)).Name == "minecraft:soul_soil"
		// LiquidBlock.POSSIBLE_FLOW_DIRECTIONS is DOWN,SOUTH,NORTH,EAST,
		// WEST; shouldSpreadLiquid queries their opposites in that order.
		for _, face := range []int{1, 2, 3, 4, 5} {
			q := fluidNeighbor(p, face)
			other := e.block(q)
			if e.rules(other).state.Type.same(FluidWater) {
				name := "cobblestone"
				if e.rules(e.block(p)).state.Source() {
					name = "obsidian"
				}
				e.set(p, structureState(name), 3)
				if e.err == nil {
					e.err = world.FluidLevelEvent(p, 1501, 0)
				}
				return e.err
			}
			if soulSoil && other.Name == "minecraft:blue_ice" {
				e.set(p, structureState("basalt"), 3)
				if e.err == nil {
					e.err = world.FluidLevelEvent(p, 1501, 0)
				}
				return e.err
			}
		}
	}
	e.schedule(p, fluid.Type, FluidTickDelay(fluid.Type, settings.UltraWarm))
	return e.err
}

// UpdateLiquidShape schedules the current liquid if either it or the changed
// neighbor is a source. Neighbor family does not matter in this predicate.
func UpdateLiquidShape(world FluidTickWorld, p FeaturePosition, state, neighbor SurfaceBlockState, settings FluidSettings) error {
	if state.Name != "minecraft:water" && state.Name != "minecraft:lava" {
		return nil
	}
	self, err := FluidStateForBlock(state)
	if err != nil {
		return err
	}
	other, err := FluidStateForBlock(neighbor)
	if err != nil {
		return err
	}
	if self.Source() || other.Source() {
		return world.ScheduleFluid(p, self.Type, FluidTickDelay(self.Type, settings.UltraWarm))
	}
	return nil
}
