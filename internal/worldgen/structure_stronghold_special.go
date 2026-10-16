package worldgen

import "strconv"

func placeStronghold_shrc(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 10, 6, 10, true, random)
	e.door(e.entryDoor(), 4, 1, 0)
	air, brick := structureState("cave_air"), structureState("stone_bricks")
	e.fill(4, 1, 10, 6, 3, 10, air, air, false)
	e.fill(0, 1, 4, 0, 3, 6, air, air, false)
	e.fill(10, 1, 4, 10, 3, 6, air, air, false)
	switch e.number("Type") {
	case 0:
		for y := int32(1); y <= 3; y++ {
			e.set(brick, 5, y, 5)
		}
		for _, torch := range []struct {
			x, z   int32
			facing string
		}{{4, 5, "west"}, {6, 5, "east"}, {5, 4, "south"}, {5, 6, "north"}} {
			e.set(stateWithProperty(structureState("wall_torch"), "facing", torch.facing), torch.x, 3, torch.z)
		}
		for _, at := range [][2]int32{{4, 4}, {4, 5}, {4, 6}, {6, 4}, {6, 5}, {6, 6}, {5, 4}, {5, 6}} {
			e.set(structureState("smooth_stone_slab"), at[0], 1, at[1])
		}
	case 1:
		for offset := int32(0); offset < 5; offset++ {
			e.set(brick, 3, 1, 3+offset)
			e.set(brick, 7, 1, 3+offset)
			e.set(brick, 3+offset, 1, 3)
			e.set(brick, 3+offset, 1, 7)
		}
		for y := int32(1); y <= 3; y++ {
			e.set(brick, 5, y, 5)
		}
		e.set(structureState("water"), 5, 4, 5)
	case 2:
		cobble, planks := structureState("cobblestone"), structureState("oak_planks")
		for offset := int32(1); offset <= 9; offset++ {
			e.set(cobble, 1, 3, offset)
			e.set(cobble, 9, 3, offset)
		}
		for offset := int32(1); offset <= 9; offset++ {
			e.set(cobble, offset, 3, 1)
			e.set(cobble, offset, 3, 9)
		}
		for _, at := range [][3]int32{{5, 1, 4}, {5, 1, 6}, {5, 3, 4}, {5, 3, 6}, {4, 1, 5}, {6, 1, 5}, {4, 3, 5}, {6, 3, 5}} {
			e.set(cobble, at[0], at[1], at[2])
		}
		for y := int32(1); y <= 3; y++ {
			for _, at := range [][2]int32{{4, 4}, {6, 4}, {4, 6}, {6, 6}} {
				e.set(cobble, at[0], y, at[1])
			}
		}
		e.set(structureState("wall_torch"), 5, 3, 5)
		for z := int32(2); z <= 8; z++ {
			e.set(planks, 2, 3, z)
			e.set(planks, 3, 3, z)
			if z <= 3 || z >= 7 {
				e.set(planks, 4, 3, z)
				e.set(planks, 5, 3, z)
				e.set(planks, 6, 3, z)
			}
			e.set(planks, 7, 3, z)
			e.set(planks, 8, 3, z)
		}
		ladder := stateWithProperty(structureState("ladder"), "facing", "west")
		for y := int32(1); y <= 3; y++ {
			e.set(ladder, 9, y, 3)
		}
		e.chest(random, "stronghold_crossing", 3, 4, 8)
	}
}

func placeStronghold_shpr(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 10, 7, 15, false, random)
	e.door("GRATES", 4, 1, 0)
	e.randomBox(1, 6, 1, 1, 6, 14, false, random)
	e.randomBox(9, 6, 1, 9, 6, 14, false, random)
	e.randomBox(2, 6, 1, 8, 6, 2, false, random)
	e.randomBox(2, 6, 14, 8, 6, 14, false, random)
	e.randomBox(1, 1, 1, 2, 1, 4, false, random)
	e.randomBox(8, 1, 1, 9, 1, 4, false, random)
	lava := structureState("lava")
	e.fill(1, 1, 1, 1, 1, 3, lava, lava, false)
	e.fill(9, 1, 1, 9, 1, 3, lava, lava, false)
	e.randomBox(3, 1, 8, 7, 1, 12, false, random)
	e.fill(4, 1, 9, 6, 1, 11, lava, lava, false)
	northSouth := stateWithProperty(stateWithProperty(structureState("iron_bars"), "north", "true"), "south", "true")
	westEast := stateWithProperty(stateWithProperty(structureState("iron_bars"), "west", "true"), "east", "true")
	for z := int32(3); z < 14; z += 2 {
		e.fill(0, 3, z, 0, 4, z, northSouth, northSouth, false)
		e.fill(10, 3, z, 10, 4, z, northSouth, northSouth, false)
	}
	for x := int32(2); x < 9; x += 2 {
		e.fill(x, 3, 15, x, 4, 15, westEast, westEast, false)
	}
	stairs := stateWithProperty(structureState("stone_brick_stairs"), "facing", "north")
	e.randomBox(4, 1, 5, 6, 1, 7, false, random)
	e.randomBox(4, 2, 6, 6, 2, 7, false, random)
	e.randomBox(4, 3, 7, 6, 3, 7, false, random)
	for x := int32(4); x <= 6; x++ {
		e.set(stairs, x, 1, 4)
		e.set(stairs, x, 2, 5)
		e.set(stairs, x, 3, 6)
	}
	var eyes [12]bool
	complete := true
	for index := range eyes {
		eyes[index] = random.NextFloat() > float32(0.9)
		complete = complete && eyes[index]
	}
	for index, at := range [][2]int32{{4, 8}, {5, 8}, {6, 8}, {4, 12}, {5, 12}, {6, 12}, {3, 9}, {3, 10}, {3, 11}, {7, 9}, {7, 10}, {7, 11}} {
		facing := [...]string{"north", "south", "east", "west"}[index/3]
		state := stateWithProperty(structureState("end_portal_frame"), "facing", facing)
		state = stateWithProperty(state, "eye", strconv.FormatBool(eyes[index]))
		e.set(state, at[0], 3, at[1])
	}
	if complete {
		for z := int32(9); z <= 11; z++ {
			for x := int32(4); x <= 6; x++ {
				e.set(structureState("end_portal"), x, 3, z)
			}
		}
	}
	e.silverfishSpawner(5, 3, 6)
}
