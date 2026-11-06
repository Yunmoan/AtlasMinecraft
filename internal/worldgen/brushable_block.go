package worldgen

import (
	"fmt"
	"strconv"
)

// BrushableBlock has its own scheduled handler, distinct from FallingBlock.
// Placement and every shape callback request delay 2 without changing DUSTED.
// Its later tick must reset brushing progress before it can start falling.
func brushableBlock(name string) bool   { return blockTickHandlers[name] == "dgd" }
func IsBrushableBlock(name string) bool { return brushableBlock(name) }

func NotifyBrushableBlockPlacement(world any, p FeaturePosition, state SurfaceBlockState) error {
	if !brushableBlock(state.Name) {
		return nil
	}
	return scheduleFallingBlock(world, p, state.Name)
}

// BrushableProgress belongs to the loaded block entity, not its saved NBT.
// In particular DUSTED is not a substitute for Count after loading a chunk.
// HitDirection uses Direction's ordinal; -1 represents the original null.
type BrushableProgress struct {
	Count                int32
	ResetAt, CooldownEnd int64
	HitDirection         int32
}

// ResetBrushableProgress implements BrushableBlockEntity.checkReset. It compares
// stages derived from the old/new counts, not the supplied block's DUSTED value.
func ResetBrushableProgress(world ScheduledBlockWorld, p FeaturePosition, state SurfaceBlockState, before BrushableProgress, now int64) (BrushableProgress, error) {
	after := before
	stage := func(count int32) int32 {
		switch {
		case count == 0:
			return 0
		case count < 3:
			return 1
		case count < 6:
			return 2
		default:
			return 3
		}
	}
	if after.Count != 0 && now >= after.ResetAt {
		old := stage(after.Count)
		after.Count = max(0, after.Count-2)
		if next := stage(after.Count); next != old {
			if err := world.SetFluidBlock(p, stateWithProperty(state, "dusted", strconv.Itoa(int(next))), 3); err != nil {
				return before, err
			}
		}
		after.ResetAt = now + 4
	}
	if after.Count == 0 {
		after.HitDirection = -1
		after.ResetAt, after.CooldownEnd = 0, 0
	} else if err := scheduleFallingBlock(world, p, state.Name); err != nil {
		return before, err
	}
	return after, nil
}

func tickBrushableBlock(world ScheduledBlockWorld, p FeaturePosition, state SurfaceBlockState) error {
	reset, ok := world.(interface {
		ResetBrushableBlockProgress(FeaturePosition, SurfaceBlockState) error
	})
	if !ok {
		return fmt.Errorf("%w: brushable tick requires block-entity progress", ErrUnsupportedScheduledBlock)
	}
	if err := reset.ResetBrushableBlockProgress(p, state); err != nil {
		return err
	}
	// Native tick keeps its incoming state even when checkReset changed DUSTED.
	return tickFallingBlock(world, p, state)
}
