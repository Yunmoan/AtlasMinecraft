package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"fmt"
	"strings"
)

func iglooBlockRules(state SurfaceBlockState) (blockrules.Rules, error) {
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return blockrules.Rules{}, err
	}
	return blockrules.Lookup(id)
}

func iglooOpposite(direction string) string {
	return map[string]string{"down": "up", "up": "down", "north": "south", "south": "north", "west": "east", "east": "west"}[direction]
}
func iglooDirectionOffset(p FeaturePosition, direction string) FeaturePosition {
	switch direction {
	case "down":
		p.Y--
	case "up":
		p.Y++
	case "north":
		p.Z--
	case "south":
		p.Z++
	case "west":
		p.X--
	case "east":
		p.X++
	}
	return p
}
func iglooClockwise(direction string) string {
	return map[string]string{"north": "east", "east": "south", "south": "west", "west": "north"}[direction]
}
func iglooCounterclockwise(direction string) string { return iglooOpposite(iglooClockwise(direction)) }
func iglooFacingAxis(direction string) string {
	switch direction {
	case "north", "south":
		return "z"
	case "west", "east":
		return "x"
	default:
		return "y"
	}
}

func iglooStairsShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState) (string, error) {
	facing := state.Properties["facing"]
	checkSide := func(direction string) (bool, error) {
		neighbor, err := world.BlockAt(iglooDirectionOffset(p, direction))
		return !strings.HasSuffix(neighbor.Name, "_stairs") || neighbor.Properties["facing"] != facing || neighbor.Properties["half"] != state.Properties["half"], err
	}
	front, err := world.BlockAt(iglooDirectionOffset(p, facing))
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(front.Name, "_stairs") && front.Properties["half"] == state.Properties["half"] && iglooFacingAxis(front.Properties["facing"]) != iglooFacingAxis(facing) {
		clear, err := checkSide(iglooOpposite(front.Properties["facing"]))
		if err != nil {
			return "", err
		}
		if clear {
			if front.Properties["facing"] == iglooCounterclockwise(facing) {
				return "outer_left", nil
			}
			return "outer_right", nil
		}
	}
	back, err := world.BlockAt(iglooDirectionOffset(p, iglooOpposite(facing)))
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(back.Name, "_stairs") && back.Properties["half"] == state.Properties["half"] && iglooFacingAxis(back.Properties["facing"]) != iglooFacingAxis(facing) {
		clear, err := checkSide(back.Properties["facing"])
		if err != nil {
			return "", err
		}
		if clear {
			if back.Properties["facing"] == iglooCounterclockwise(facing) {
				return "inner_left", nil
			}
			return "inner_right", nil
		}
	}
	return "straight", nil
}

// Only block subclasses present in the original igloo palettes are handled
// here. General template processors and arbitrary template shape updates need
// their own ports; this is not a blanket implementation of every block type.
func iglooTemplateUpdateShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	air := structureState("air")
	facing := state.Properties["facing"]
	switch {
	case state.Name == "minecraft:iron_bars":
		if direction == "up" || direction == "down" {
			return state, nil
		}
		rules, err := iglooBlockRules(neighbor)
		if err != nil {
			return state, err
		}
		face := map[string]int{"north": 3, "south": 2, "west": 5, "east": 4}[direction]
		connect := !rules.ConnectionException && rules.FaceSturdy[face] || neighbor.Name == "minecraft:iron_bars" || strings.HasSuffix(neighbor.Name, "_glass_pane") || neighbor.Name == "minecraft:glass_pane" || strings.HasSuffix(neighbor.Name, "_wall")
		value := "false"
		if connect {
			value = "true"
		}
		return stateWithProperty(state, direction, value), nil
	case strings.HasSuffix(state.Name, "_stairs"):
		if direction == "up" || direction == "down" {
			return state, nil
		}
		shape, err := iglooStairsShape(world, p, state)
		return stateWithProperty(state, "shape", shape), err
	case strings.HasSuffix(state.Name, "_bed"):
		otherDirection := facing
		if state.Properties["part"] == "head" {
			otherDirection = iglooOpposite(facing)
		}
		if direction == otherDirection {
			if neighbor.Name == state.Name && neighbor.Properties["part"] != state.Properties["part"] {
				return stateWithProperty(state, "occupied", neighbor.Properties["occupied"]), nil
			}
			return air, nil
		}
	case strings.HasSuffix(state.Name, "_carpet"):
		below, err := world.BlockAt(iglooDirectionOffset(p, "down"))
		if err != nil {
			return state, err
		}
		if surfaceAir(below) {
			return air, nil
		}
	case state.Name == "minecraft:ladder" || strings.HasSuffix(state.Name, "_wall_torch") || state.Name == "minecraft:wall_torch" || strings.HasSuffix(state.Name, "_wall_sign"):
		if direction == iglooOpposite(facing) {
			rules, err := iglooBlockRules(neighbor)
			if err != nil {
				return state, err
			}
			supported := rules.FaceSturdy[map[string]int{"north": 2, "south": 3, "west": 4, "east": 5}[facing]]
			if strings.HasSuffix(state.Name, "_wall_sign") {
				supported = rules.Solid
			}
			if !supported {
				return air, nil
			}
		}
	case state.Name == "minecraft:torch" || state.Name == "minecraft:redstone_torch" || state.Name == "minecraft:soul_torch":
		if direction == "down" {
			rules, err := iglooBlockRules(neighbor)
			if err != nil {
				return state, err
			}
			if !rules.TopCenter {
				return air, nil
			}
		}
	}
	return state, nil
}

func iglooSetShapeIfChanged(world FeatureBlockWorld, p FeaturePosition, old, new SurfaceBlockState) error {
	oldID, err := blockstate.BlockStateID(old.Name, old.Properties)
	if err != nil {
		return err
	}
	newID, err := blockstate.BlockStateID(new.Name, new.Properties)
	if err != nil {
		return err
	}
	if oldID == newID {
		return nil
	}
	return world.SetBlock(p, new)
}

func finalizeIglooTemplateShape(world FeatureBlockWorld, placed []FeaturePosition) error {
	return finalizeStructureTemplateShape(world, placed, iglooTemplateUpdateShape)
}

type structureTemplateShapeUpdater func(FeatureBlockWorld, FeaturePosition, SurfaceBlockState, string, SurfaceBlockState) (SurfaceBlockState, error)

func finalizeStructureTemplateShape(world FeatureBlockWorld, placed []FeaturePosition, updateShape structureTemplateShapeUpdater) error {
	if len(placed) == 0 {
		return nil
	}
	set := make(map[FeaturePosition]bool, len(placed))
	minimum, maximum := placed[0], placed[0]
	for _, p := range placed {
		set[p] = true
		minimum.X = min(minimum.X, p.X)
		minimum.Y = min(minimum.Y, p.Y)
		minimum.Z = min(minimum.Z, p.Z)
		maximum.X = max(maximum.X, p.X)
		maximum.Y = max(maximum.Y, p.Y)
		maximum.Z = max(maximum.Z, p.Z)
	}
	updateEdge := func(p FeaturePosition, direction string) error {
		q := iglooDirectionOffset(p, direction)
		state, err := world.BlockAt(p)
		if err != nil {
			return err
		}
		neighbor, err := world.BlockAt(q)
		if err != nil {
			return err
		}
		updated, err := updateShape(world, p, state, direction, neighbor)
		if err != nil {
			return err
		}
		if err := iglooSetShapeIfChanged(world, p, state, updated); err != nil {
			return err
		}
		outside, err := updateShape(world, q, neighbor, iglooOpposite(direction), updated)
		if err != nil {
			return err
		}
		return iglooSetShapeIfChanged(world, q, neighbor, outside)
	}
	// DiscreteVoxelShape.forAllFaces: inverse axis-cycle NONE/BACKWARD/FORWARD.
	for axis := 0; axis < 3; axis++ {
		firstMin, firstMax, secondMin, secondMax, scanMin, scanMax := minimum.X, maximum.X, minimum.Y, maximum.Y, minimum.Z, maximum.Z
		negative, positive := "north", "south"
		if axis == 1 {
			firstMin, firstMax, secondMin, secondMax, scanMin, scanMax = minimum.Z, maximum.Z, minimum.X, maximum.X, minimum.Y, maximum.Y
			negative, positive = "down", "up"
		}
		if axis == 2 {
			firstMin, firstMax, secondMin, secondMax, scanMin, scanMax = minimum.Y, maximum.Y, minimum.Z, maximum.Z, minimum.X, maximum.X
			negative, positive = "west", "east"
		}
		for first := firstMin; first <= firstMax; first++ {
			for second := secondMin; second <= secondMax; second++ {
				previous := false
				for scan := scanMin; scan <= scanMax+1; scan++ {
					p := FeaturePosition{first, second, scan}
					if axis == 1 {
						p = FeaturePosition{second, scan, first}
					}
					if axis == 2 {
						p = FeaturePosition{scan, first, second}
					}
					current := scan <= scanMax && set[p]
					if !previous && current {
						if err := updateEdge(p, negative); err != nil {
							return err
						}
					}
					if previous && !current {
						p = iglooDirectionOffset(p, negative)
						if err := updateEdge(p, positive); err != nil {
							return err
						}
					}
					previous = current
				}
			}
		}
	}
	for _, p := range placed {
		state, err := world.BlockAt(p)
		if err != nil {
			return err
		}
		updated := state
		for _, direction := range []string{"west", "east", "north", "south", "down", "up"} {
			neighbor, err := world.BlockAt(iglooDirectionOffset(p, direction))
			if err != nil {
				return err
			}
			updated, err = updateShape(world, p, updated, direction, neighbor)
			if err != nil {
				return fmt.Errorf("template shape at %v: %w", p, err)
			}
		}
		if err := iglooSetShapeIfChanged(world, p, state, updated); err != nil {
			return err
		}
	}
	return nil
}
