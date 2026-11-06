package worldgen

import (
	"atlas/internal/blockstate"
	"fmt"
)

// PendingFluidTick is a ProtoChunkTicks entry. Generation ticks have delay
// zero when saved; runtime scheduling supplies time and sub-tick order later.
// Fluid identifies the registered type (source and flowing types are distinct).
type PendingFluidTick struct {
	Position FeaturePosition
	Fluid    string
	Priority int32
}

type generationFluidTickKey struct {
	position FeaturePosition
	fluid    string
}

type generationFluidTickQueue struct {
	ticks []PendingFluidTick
	seen  map[generationFluidTickKey]struct{}
}

func (q *generationFluidTickQueue) add(tick PendingFluidTick) {
	key := generationFluidTickKey{tick.Position, tick.Fluid}
	if q.seen == nil {
		q.seen = make(map[generationFluidTickKey]struct{})
	}
	if _, exists := q.seen[key]; exists {
		return
	}
	q.seen[key] = struct{}{}
	q.ticks = append(q.ticks, tick)
}

// generationFluidType accepts a canonical block state. These cases are
// exhaustively checked against BlockState.getFluidState for all 26,684 states.
func generationFluidType(state SurfaceBlockState) string {
	switch state.Name {
	case "minecraft:water", "minecraft:lava":
		if level := state.Properties["level"]; level != "" && level != "0" {
			return "minecraft:flowing_" + state.Name[len("minecraft:"):]
		}
		return state.Name
	case "minecraft:bubble_column", "minecraft:seagrass", "minecraft:tall_seagrass", "minecraft:kelp", "minecraft:kelp_plant":
		return "minecraft:water"
	}
	if state.Properties["waterlogged"] == "true" {
		return "minecraft:water"
	}
	return "minecraft:empty"
}

func (region *OverworldFeatureRegion) fluidTickAt(p FeaturePosition) (PendingFluidTick, error) {
	if !region.InWorld(p) {
		return PendingFluidTick{Position: p, Fluid: "minecraft:empty"}, nil
	}
	state, err := region.BlockAt(p)
	if err != nil {
		return PendingFluidTick{}, err
	}
	// Partial template states inherit the registry defaults, including the
	// default waterlogged=true on corals and sea pickles.
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return PendingFluidTick{}, err
	}
	full, err := blockstate.BlockStateByID(id)
	if err != nil {
		return PendingFluidTick{}, err
	}
	return PendingFluidTick{Position: p, Fluid: generationFluidType(SurfaceBlockState{Name: full.Name, Properties: full.Properties})}, nil
}

func (region *OverworldFeatureRegion) ScheduleFluidUpdate(p FeaturePosition) error {
	tick, err := region.fluidTickAt(p)
	if err != nil {
		return err
	}
	if tick.Fluid == "minecraft:empty" {
		return nil
	}
	return region.ScheduleFluidTick(tick)
}

// ScheduleFluidTick preserves the requested type even if a subsequent write
// changes the block. Native execution checks the type when the tick is due.
func (region *OverworldFeatureRegion) ScheduleFluidTick(tick PendingFluidTick) error {
	// WorldGenTickAccess routes typed requests by owning chunk, independently
	// of the bounds checks used for feature writes or fluid-state queries.
	if tick.Priority < -3 || tick.Priority > 3 {
		return fmt.Errorf("invalid fluid tick priority %d", tick.Priority)
	}
	switch tick.Fluid {
	case "minecraft:empty", "minecraft:water", "minecraft:flowing_water", "minecraft:lava", "minecraft:flowing_lava":
	default:
		return fmt.Errorf("unknown scheduled fluid %q", tick.Fluid)
	}
	chunk, err := region.chunkAt(tick.Position.X, tick.Position.Z)
	if err != nil {
		return err
	}
	key := featureChunkPosition{chunk.X, chunk.Z}
	if region.fluidTicks == nil {
		region.fluidTicks = make(map[featureChunkPosition]*generationFluidTickQueue)
	}
	queue := region.fluidTicks[key]
	if queue == nil {
		queue = &generationFluidTickQueue{}
		for _, previous := range chunk.FluidTicks {
			queue.add(previous)
		}
		region.fluidTicks[key] = queue
	}
	queue.add(tick)
	if recording := region.structureRecording; recording != nil {
		recording.fluidTicks.add(tick)
	}
	return nil
}

func (region *OverworldFeatureRegion) fluidTicksInChunk(cx, cz int32) []PendingFluidTick {
	key := featureChunkPosition{cx, cz}
	if queue := region.fluidTicks[key]; queue != nil {
		return append([]PendingFluidTick(nil), queue.ticks...)
	}
	return append([]PendingFluidTick(nil), region.chunks[key].FluidTicks...)
}
