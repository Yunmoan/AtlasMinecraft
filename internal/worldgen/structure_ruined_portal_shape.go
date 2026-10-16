package worldgen

import "strings"

func ruinedPortalTemplateUpdateShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	flags, err := ruinedPortalStateFlags(state)
	if err != nil {
		return state, err
	}
	if flags&8 == 0 {
		return iglooTemplateUpdateShape(world, p, state, direction, neighbor)
	}
	if direction == "down" {
		return state, nil
	}
	connections := make(map[string]bool, 4)
	for _, face := range []string{"north", "east", "south", "west"} {
		connections[face] = state.Properties[face] != "" && state.Properties[face] != "none"
	}
	if direction != "up" {
		rules, err := iglooBlockRules(neighbor)
		if err != nil {
			return state, err
		}
		neighborFlags, err := ruinedPortalStateFlags(neighbor)
		if err != nil {
			return state, err
		}
		oppositeFace := map[string]int{"north": 3, "south": 2, "west": 5, "east": 4}[direction]
		gate := strings.HasSuffix(neighbor.Name, "_fence_gate") && iglooFacingAxis(neighbor.Properties["facing"]) != iglooFacingAxis(direction)
		connections[direction] = neighborFlags&8 != 0 || !rules.ConnectionException && rules.FaceSturdy[oppositeFace] || neighbor.Name == "minecraft:iron_bars" || gate
	}
	above, err := world.BlockAt(treeOffset(p, 0, 1, 0))
	if err != nil {
		return state, err
	}
	aboveFlags, err := ruinedPortalStateFlags(above)
	if err != nil {
		return state, err
	}
	for _, face := range []string{"north", "east", "south", "west"} {
		side := "none"
		if connections[face] {
			side = "low"
			if aboveFlags&(1<<map[string]int{"north": 13, "south": 14, "west": 15, "east": 16}[face]) != 0 {
				side = "tall"
			}
		}
		state = stateWithProperty(state, face, side)
	}
	up := true
	if aboveFlags&8 == 0 || above.Properties["up"] != "true" {
		n, e, s, w := !connections["north"], !connections["east"], !connections["south"], !connections["west"]
		if !(n && e && s && w || n != s || e != w) {
			if state.Properties["north"] == "tall" && state.Properties["south"] == "tall" || state.Properties["east"] == "tall" && state.Properties["west"] == "tall" {
				up = false
			} else {
				up = aboveFlags&16 != 0 || aboveFlags&(1<<12) != 0
			}
		}
	}
	value := "false"
	if up {
		value = "true"
	}
	return stateWithProperty(state, "up", value), nil
}
