package worldgen

// UpdateFluidNeighborShape applies immediate plant callbacks, schedules leaf
// and falling-block updates, and reuses the existing structure block shapes.
// Redstone and other delayed families still need their scheduled handlers.
func UpdateFluidNeighborShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, face int, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	direction := [6]string{"down", "up", "north", "south", "west", "east"}[face]
	if kelpBlock(state.Name) {
		return updateKelpNeighborShape(world, p, state, direction, neighbor)
	}
	if brushableBlock(state.Name) {
		return state, NotifyBrushableBlockPlacement(world, p, state)
	}
	if fallingBlock(state.Name) {
		return updateFallingNeighborShape(world, p, state)
	}
	if state.Name == "minecraft:magma_block" || state.Name == "minecraft:soul_sand" || state.Name == "minecraft:bubble_column" {
		return updateBubbleNeighborShape(world, p, state, face, neighbor)
	}
	if blockTickHandlers[state.Name] == "dki" {
		return updateRuntimeLeafShape(world, p, state, neighbor)
	}
	if runtimePlantShape(state.Name) {
		return updateRuntimePlantShape(world, p, state, direction, neighbor)
	}
	return mansionTemplateUpdateShape(world, p, state, direction, neighbor)
}
