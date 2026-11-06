package worldgen

import (
	"fmt"
	"strconv"
)

func kelpBlock(name string) bool { return name == "minecraft:kelp" || name == "minecraft:kelp_plant" }

func kelpSurvives(world interface {
	BlockAt(FeaturePosition) (SurfaceBlockState, error)
}, p FeaturePosition) (bool, error) {
	below, err := world.BlockAt(treeOffset(p, 0, -1, 0))
	if err != nil {
		return false, err
	}
	rules, err := iglooBlockRules(below)
	if err != nil {
		return false, err
	}
	return below.Name != "minecraft:magma_block" && (kelpBlock(below.Name) || rules.FaceSturdy[1]), nil
}

func tickKelp(world ScheduledBlockWorld, p FeaturePosition, state SurfaceBlockState) error {
	survives, err := kelpSurvives(world, p)
	if err != nil || survives {
		return err
	}
	destroy, ok := world.(interface {
		DestroyKelpBlock(FeaturePosition, SurfaceBlockState) error
	})
	if !ok {
		return fmt.Errorf("%w: kelp requires destruction and block drops", ErrUnsupportedScheduledBlock)
	}
	return destroy.DestroyKelpBlock(p, state)
}

// GrowingPlantHeadBlock/BodyBlock updateShape, with KelpBlock's support veto.
// Conversion returns before the water request. Support loss requests delay 1;
// it does not remove the plant during this callback.
func updateKelpNeighborShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	if direction == "down" {
		survives, err := kelpSurvives(world, p)
		if err != nil {
			return state, err
		}
		if !survives {
			scheduler, ok := world.(interface {
				ScheduleBlockTick(FeaturePosition, string, int32, int32) error
			})
			if !ok {
				return state, fmt.Errorf("kelp shape requires a block tick scheduler")
			}
			if err := scheduler.ScheduleBlockTick(p, state.Name, 1, 0); err != nil {
				return state, err
			}
		}
	}
	if direction == "up" {
		if state.Name == "minecraft:kelp" && kelpBlock(neighbor.Name) {
			return structureState("kelp_plant"), nil
		}
		if state.Name == "minecraft:kelp_plant" && !kelpBlock(neighbor.Name) {
			random, ok := world.(interface{ ShapeRandomInt(int32) (int32, error) })
			if !ok {
				return state, fmt.Errorf("kelp tip conversion requires world randomness")
			}
			age, err := random.ShapeRandomInt(25)
			if err != nil {
				return state, err
			}
			return stateWithProperty(structureState("kelp"), "age", strconv.Itoa(int(age))), nil
		}
	}
	return state, runtimePlantWaterTick(world, p)
}

func (region *OverworldFeatureRegion) ShapeRandomInt(bound int32) (int32, error) {
	if region.shapeRandom == nil {
		return 0, fmt.Errorf("shape callback requires the current generation region random")
	}
	return region.shapeRandom.NextIntBound(bound), nil
}
