package worldgen

import "fmt"

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

// PostProcessGeneratedBubbleColumns settles just the generated magma/soul-sand
// bubble ticks after decoration and before the detached final chunk is saved.
// Sources live in the sparse state map; no full-volume scan or neighbor chunk
// request is needed. A feature written across a chunk border is included in
// its destination's final state map. Only vertical columns are affected.
func PostProcessGeneratedBubbleColumns(chunk *NoiseChunkResult) (int, error) {
	if chunk == nil || chunk.Height <= 0 || len(chunk.Blocks) != int(chunk.Height)*256 {
		return 0, fmt.Errorf("bubble postprocessing requires a complete chunk")
	}
	world := generatedBubbleColumnWorld{chunk}
	changed := 0
	for index, support := range chunk.SurfaceBlocks {
		if support.Name != "minecraft:magma_block" && support.Name != "minecraft:soul_sand" {
			continue
		}
		if index < 0 || index >= len(chunk.Blocks) {
			return changed, fmt.Errorf("bubble source index %d outside chunk", index)
		}
		position := FeaturePosition{chunk.X*16 + int32(index&15), chunk.MinY + int32(index/256) + 1, chunk.Z*16 + int32((index>>4)&15)}
		n, err := UpdateBubbleColumn(world, position, support)
		if err != nil {
			return changed, err
		}
		changed += n
	}
	return changed, nil
}

type generatedBubbleColumnWorld struct{ chunk *NoiseChunkResult }

func (world generatedBubbleColumnWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	return world.chunk.BlockStateAt(p.X-world.chunk.X*16, p.Y, p.Z-world.chunk.Z*16)
}

func (world generatedBubbleColumnWorld) SetBlock(p FeaturePosition, state SurfaceBlockState) error {
	if !world.InWorld(p) {
		return fmt.Errorf("bubble column write outside generated chunk at %v", p)
	}
	world.chunk.setSurfaceState(p.X-world.chunk.X*16, p.Y, p.Z-world.chunk.Z*16, state)
	return nil
}

func (world generatedBubbleColumnWorld) InWorld(p FeaturePosition) bool {
	return p.X>>4 == world.chunk.X && p.Z>>4 == world.chunk.Z &&
		p.Y >= world.chunk.MinY && p.Y < world.chunk.MinY+world.chunk.Height
}
