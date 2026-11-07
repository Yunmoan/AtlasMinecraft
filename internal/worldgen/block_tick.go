package worldgen

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

//go:embed block_tick_handlers_1.21.1.json
var blockTickHandlerData []byte

// Original method owners distinguish inherited no-op handlers from handlers
// not ported yet. Unknown behavior must not silently consume a saved request.
var blockTickHandlers = func() map[string]string {
	var groups map[string][]string
	if err := json.Unmarshal(blockTickHandlerData, &groups); err != nil {
		panic(err)
	}
	result := make(map[string]string)
	for owner, names := range groups {
		for _, name := range names {
			if result[name] != "" {
				panic("duplicate block tick handler")
			}
			result[name] = owner
		}
	}
	return result
}()

var ErrUnsupportedScheduledBlock = errors.New("scheduled block handler is not implemented")

type ScheduledBlockWorld interface {
	BlockAt(FeaturePosition) (SurfaceBlockState, error)
	SetFluidBlock(FeaturePosition, SurfaceBlockState, int32) error
}

// TickScheduledBlock runs BlockState.tick only when the current block matches
// the scheduled block type. Leaf decay is a separate random tick operation.
func TickScheduledBlock(world ScheduledBlockWorld, p FeaturePosition, name string) error {
	state, err := world.BlockAt(p)
	if err != nil {
		return err
	}
	if state.Name != name {
		return nil
	}
	switch blockTickHandlers[name] {
	case "dtb":
		return nil // BlockBehaviour.tick is empty.
	case "diu":
		return tickFallingBlock(world, p, state)
	case "dgd":
		return tickBrushableBlock(world, p, state)
	case "djl":
		if kelpBlock(name) {
			return tickKelp(world, p, state)
		}
		return fmt.Errorf("%w: %s (original owner djl)", ErrUnsupportedScheduledBlock, name)
	case "dkr", "dnl", "dge":
		w, ok := world.(FeatureBlockWorld)
		if !ok {
			return fmt.Errorf("bubble tick requires bounded block access")
		}
		position, support := treeOffset(p, 0, 1, 0), state
		if state.Name == "minecraft:bubble_column" {
			position = p
			support, err = world.BlockAt(treeOffset(p, 0, -1, 0))
			if err != nil {
				return err
			}
		}
		_, err := UpdateBubbleColumn(w, position, support)
		return err
	case "dki":
		distance := 7
		for _, d := range [6]FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}} {
			neighbor, err := world.BlockAt(treeOffset(p, d.X, d.Y, d.Z))
			if err != nil {
				return err
			}
			distance = min(distance, scheduledLeafDistance(neighbor)+1)
			if distance == 1 {
				break
			}
		}
		return world.SetFluidBlock(p, stateWithProperty(state, "distance", strconv.Itoa(distance)), 3)
	default:
		return fmt.Errorf("%w: %s (original owner %s)", ErrUnsupportedScheduledBlock, name, blockTickHandlers[name])
	}
}

func scheduledLeafDistance(state SurfaceBlockState) int {
	if treeTags["logs"][state.Name] {
		return 0
	}
	if blockTickHandlers[state.Name] == "dki" {
		if distance, err := strconv.Atoi(state.Properties["distance"]); err == nil {
			return distance
		}
	}
	return 7
}

func updateRuntimeLeafShape(world FeatureBlockWorld, p FeaturePosition, state, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	if state.Properties["waterlogged"] == "true" {
		if err := runtimePlantWaterTick(world, p); err != nil {
			return state, err
		}
	}
	distance := scheduledLeafDistance(neighbor) + 1
	if distance != 1 || scheduledLeafDistance(state) != distance {
		scheduler, ok := world.(interface {
			ScheduleBlockTick(FeaturePosition, string, int32, int32) error
		})
		if !ok {
			return state, fmt.Errorf("leaf shape requires a block tick scheduler")
		}
		return state, scheduler.ScheduleBlockTick(p, state.Name, 1, 0)
	}
	return state, nil
}
