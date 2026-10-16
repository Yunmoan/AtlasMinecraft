package worldgen

import "fmt"

// PostProcessSwampHutBlocks ports the original hut's clipped geometry, height
// alignment, stair mirror/rotation and support pillars. Witch/cat requests are
// returned explicitly; their flags remain false until a runtime consumer
// actually fulfills the entity stage. It does not imply a complete structure.
func PostProcessSwampHutBlocks(world FeatureBlockWorld, piece *StructurePiece, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if piece == nil || piece.ID != "minecraft:tesh" {
		return result, fmt.Errorf("swamp hut postprocessing requires a tesh piece")
	}
	e, err := newScatteredPieceEditor(world, piece, chunkX, chunkZ)
	if err != nil {
		return result, err
	}
	ready, err := scatteredAverageHeight(world, piece, e.clip)
	if err != nil || !ready {
		return result, err
	}
	planks := SurfaceBlockState{Name: "minecraft:spruce_planks"}
	log := SurfaceBlockState{Name: "minecraft:oak_log"}
	air := SurfaceBlockState{Name: "minecraft:air"}
	fence := SurfaceBlockState{Name: "minecraft:oak_fence"}
	e.fill(1, 1, 1, 5, 1, 7, planks, planks, false)
	e.fill(1, 4, 2, 5, 4, 7, planks, planks, false)
	e.fill(2, 1, 0, 4, 1, 0, planks, planks, false)
	e.fill(2, 2, 2, 3, 3, 2, planks, planks, false)
	e.fill(1, 2, 3, 1, 3, 6, planks, planks, false)
	e.fill(5, 2, 3, 5, 3, 6, planks, planks, false)
	e.fill(2, 2, 7, 4, 3, 7, planks, planks, false)
	e.fill(1, 0, 2, 1, 3, 2, log, log, false)
	e.fill(5, 0, 2, 5, 3, 2, log, log, false)
	e.fill(1, 0, 7, 1, 3, 7, log, log, false)
	e.fill(5, 0, 7, 5, 3, 7, log, log, false)
	e.set(fence, 2, 3, 2)
	e.set(fence, 3, 3, 7)
	e.set(air, 1, 3, 4)
	e.set(air, 5, 3, 4)
	e.set(air, 5, 3, 5)
	e.set(SurfaceBlockState{Name: "minecraft:potted_red_mushroom"}, 1, 3, 5)
	e.set(SurfaceBlockState{Name: "minecraft:crafting_table"}, 3, 2, 6)
	e.set(SurfaceBlockState{Name: "minecraft:cauldron"}, 4, 2, 6)
	e.set(fence, 1, 2, 1)
	e.set(fence, 5, 2, 1)
	stairs := func(facing string) SurfaceBlockState {
		return SurfaceBlockState{Name: "minecraft:spruce_stairs", Properties: map[string]string{"facing": facing}}
	}
	north, east, west, south := stairs("north"), stairs("east"), stairs("west"), stairs("south")
	e.fill(0, 4, 1, 6, 4, 1, north, north, false)
	e.fill(0, 4, 2, 0, 4, 7, east, east, false)
	e.fill(6, 4, 2, 6, 4, 7, west, west, false)
	e.fill(0, 4, 8, 6, 4, 8, south, south, false)
	e.set(stateWithProperty(north, "shape", "outer_right"), 0, 4, 1)
	e.set(stateWithProperty(north, "shape", "outer_left"), 6, 4, 1)
	e.set(stateWithProperty(south, "shape", "outer_left"), 0, 4, 8)
	e.set(stateWithProperty(south, "shape", "outer_right"), 6, 4, 8)
	for z := int32(2); z <= 7; z += 5 {
		for x := int32(1); x <= 5; x += 4 {
			e.fillDown(log, x, -1, z)
		}
	}
	if e.err != nil {
		return result, e.err
	}
	position := piece.BlockPosition(2, 2, 5)
	if structureContains(e.clip, position) {
		for _, entity := range []struct{ id, flag string }{{"minecraft:witch", "Witch"}, {"minecraft:cat", "Cat"}} {
			flag, _ := piece.Data[entity.flag].Number()
			if flag == 0 {
				result.PendingEntities = append(result.PendingEntities, PendingStructureEntity{ID: entity.id, Position: position, Persistent: true, SavedFlag: entity.flag})
			}
		}
	}
	return result, nil
}
