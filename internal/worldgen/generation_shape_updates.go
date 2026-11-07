package worldgen

import "fmt"

// Template edges touch neighboring terrain as well as palette states. Delayed
// callbacks belong at this shared boundary, before palette-specific shape
// adapters. The final Block.updateFromNeighbourShapes pass uses it too, retaining
// duplicate callbacks until the owning ProtoChunkTicks queue deduplicates them.
func updateTemplateShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState, updateShape structureTemplateShapeUpdater) (SurfaceBlockState, error) {
	w := generationShapeWorld{world}
	if kelpBlock(state.Name) {
		return updateKelpNeighborShape(w, p, state, direction, neighbor)
	}
	if runtimePlantShape(state.Name) {
		return updateRuntimePlantShape(w, p, state, direction, neighbor)
	}
	if brushableBlock(state.Name) {
		return state, NotifyBrushableBlockPlacement(w, p, state)
	}
	if fallingBlock(state.Name) {
		return updateFallingNeighborShape(w, p, state)
	}
	if blockTickHandlers[state.Name] == "dki" {
		return updateRuntimeLeafShape(w, p, state, neighbor)
	}
	if state.Name == "minecraft:magma_block" || state.Name == "minecraft:soul_sand" || state.Name == "minecraft:bubble_column" {
		face := map[string]int{"down": 0, "up": 1, "north": 2, "south": 3, "west": 4, "east": 5}[direction]
		return updateBubbleNeighborShape(w, p, state, face, neighbor)
	}
	if state.Name == "minecraft:water" || state.Name == "minecraft:lava" {
		return updateLiquidNeighborShape(w, p, state, neighbor)
	}
	return updateShape(world, p, state, direction, neighbor)
}

// LiquidBlock tests the actual FluidState of both blocks. A bubble column or
// waterlogged plant can supply a source just as a plain water block can.
func updateLiquidNeighborShape(world FeatureBlockWorld, p FeaturePosition, state, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	fluid, err := FluidStateForBlock(state)
	if err != nil {
		return state, err
	}
	next, err := FluidStateForBlock(neighbor)
	if err != nil {
		return state, err
	}
	if fluid.Source() || next.Source() {
		return state, (generationShapeWorld{world}).ScheduleFluidUpdate(p)
	}
	return state, nil
}

// Defined feature probes may intentionally omit scheduled callbacks. Real
// generation regions and structure recorders forward them to ProtoChunkTicks.
type generationShapeWorld struct{ FeatureBlockWorld }

func (w generationShapeWorld) ShapeRandomInt(bound int32) (int32, error) {
	if source, ok := w.FeatureBlockWorld.(interface{ ShapeRandomInt(int32) (int32, error) }); ok {
		return source.ShapeRandomInt(bound)
	}
	return 0, fmt.Errorf("generation shape requires world randomness")
}

func (w generationShapeWorld) ScheduleBlockTick(p FeaturePosition, name string, delay, priority int32) error {
	if scheduler, ok := w.FeatureBlockWorld.(interface {
		ScheduleBlockTick(FeaturePosition, string, int32, int32) error
	}); ok {
		return scheduler.ScheduleBlockTick(p, name, delay, priority)
	}
	return nil
}
func (w generationShapeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	if scheduler, ok := w.FeatureBlockWorld.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
		return scheduler.ScheduleFluidUpdate(p)
	}
	return nil
}
func (w generationShapeWorld) FeatureRawBrightness(p FeaturePosition, darkening int32) (int32, error) {
	return featureRawBrightness(w.FeatureBlockWorld, p, darkening)
}
