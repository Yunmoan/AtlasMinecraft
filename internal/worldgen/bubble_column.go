package worldgen

import "fmt"

func scheduleBubbleBlock(world FeatureBlockWorld, p FeaturePosition, name string, delay int32) error {
	scheduler, ok := world.(interface {
		ScheduleBlockTick(FeaturePosition, string, int32, int32) error
	})
	if !ok {
		return fmt.Errorf("bubble update requires block tick scheduling")
	}
	return scheduler.ScheduleBlockTick(p, name, delay, 0)
}

// MagmaBlock and SoulSandBlock schedule unconditionally in onPlace.
func NotifyBubbleBlockPlacement(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState) error {
	if state.Name == "minecraft:magma_block" || state.Name == "minecraft:soul_sand" {
		return scheduleBubbleBlock(world, p, state.Name, 20)
	}
	return nil
}

func updateBubbleNeighborShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, face int, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	if state.Name != "minecraft:bubble_column" {
		if face == 1 && neighbor.Name == "minecraft:water" {
			return state, scheduleBubbleBlock(world, p, state.Name, 20)
		}
		return state, nil
	}
	if err := runtimePlantWaterTick(world, p); err != nil {
		return state, err
	}
	below, err := world.BlockAt(treeOffset(p, 0, -1, 0))
	if err != nil {
		return state, err
	}
	survives := below.Name == "minecraft:bubble_column" || below.Name == "minecraft:magma_block" || below.Name == "minecraft:soul_sand"
	if !survives || face == 0 || face == 1 && neighbor.Name != "minecraft:bubble_column" && bubbleColumnCanOccupy(neighbor) {
		return state, scheduleBubbleBlock(world, p, state.Name, 5)
	}
	return state, nil
}

// UpdateBubbleColumn reproduces BubbleColumnBlock.updateColumn's block-state
// changes when the magma/soul-sand or bubble-column tick runs. It does not
// simulate fluid flow, scheduled-tick timing or entity movement.
func UpdateBubbleColumn(world FeatureBlockWorld, position FeaturePosition, support SurfaceBlockState) (int, error) {
	if world == nil {
		return 0, fmt.Errorf("bubble column update requires a world")
	}
	if !world.InWorld(position) {
		return 0, nil
	}
	state, err := world.BlockAt(position)
	if err != nil || !bubbleColumnCanOccupy(state) {
		return 0, err
	}
	target := SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "0"}}
	switch support.Name {
	case "minecraft:magma_block":
		target = SurfaceBlockState{Name: "minecraft:bubble_column", Properties: map[string]string{"drag": "true"}}
	case "minecraft:soul_sand":
		target = SurfaceBlockState{Name: "minecraft:bubble_column", Properties: map[string]string{"drag": "false"}}
	case "minecraft:bubble_column":
		drag := support.Properties["drag"]
		if drag == "" {
			drag = "true"
		}
		target = SurfaceBlockState{Name: support.Name, Properties: map[string]string{"drag": drag}}
	}
	changed := 0
	for first := true; world.InWorld(position); first = false {
		if !first {
			state, err = world.BlockAt(position)
			if err != nil || !bubbleColumnCanOccupy(state) {
				return changed, err
			}
		}
		if bubbleColumnSameState(state, target) {
			// The first setBlock return value is ignored. Further up,
			// Level.setBlock returns false for an unchanged state and stops.
			if !first {
				break
			}
		} else {
			if err := world.SetBlock(position, cloneSurfaceState(target)); err != nil {
				return changed, err
			}
			changed++
		}
		position.Y++
	}
	return changed, nil
}

func bubbleColumnCanOccupy(state SurfaceBlockState) bool {
	return state.Name == "minecraft:bubble_column" || state.Name == "minecraft:water" &&
		(state.Properties["level"] == "" || state.Properties["level"] == "0")
}

func bubbleColumnSameState(a, b SurfaceBlockState) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Name == "minecraft:water" {
		return a.Properties["level"] == "" || a.Properties["level"] == "0"
	}
	drag := a.Properties["drag"]
	if drag == "" {
		drag = "true"
	}
	return drag == b.Properties["drag"]
}
