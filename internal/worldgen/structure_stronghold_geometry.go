package worldgen

func placeStronghold_shs(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 4, 4, 6, true, random)
	e.door(e.entryDoor(), 1, 1, 0)
	e.door("OPENING", 1, 1, 6)
	v7 := fortressProperty(structureState("wall_torch"), "facing", "east")
	v8 := fortressProperty(structureState("wall_torch"), "facing", "west")
	e.maybe(random, 0.1, v7, 1, 2, 1)
	e.maybe(random, 0.1, v8, 3, 2, 1)
	e.maybe(random, 0.1, v7, 1, 2, 5)
	e.maybe(random, 0.1, v8, 3, 2, 5)
	if e.flag("Left") {
		e.fill(0, 1, 2, 0, 3, 4, structureState("cave_air"), structureState("cave_air"), false)
	}
	if e.flag("Right") {
		e.fill(4, 1, 2, 4, 3, 4, structureState("cave_air"), structureState("cave_air"), false)
	}
}

func placeStronghold_shph(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 8, 4, 10, true, random)
	e.door(e.entryDoor(), 1, 1, 0)
	e.fill(1, 1, 10, 3, 3, 10, structureState("cave_air"), structureState("cave_air"), false)
	e.randomBox(4, 1, 1, 4, 3, 1, false, random)
	e.randomBox(4, 1, 3, 4, 3, 3, false, random)
	e.randomBox(4, 1, 7, 4, 3, 7, false, random)
	e.randomBox(4, 1, 9, 4, 3, 9, false, random)
	for v7 := int32(1); v7 <= 3; v7++ {
		e.set(fortressProperty(fortressProperty(structureState("iron_bars"), "north", "true"), "south", "true"), 4, v7, 4)
		e.set(fortressProperty(fortressProperty(fortressProperty(structureState("iron_bars"), "north", "true"), "south", "true"), "east", "true"), 4, v7, 5)
		e.set(fortressProperty(fortressProperty(structureState("iron_bars"), "north", "true"), "south", "true"), 4, v7, 6)
		e.set(fortressProperty(fortressProperty(structureState("iron_bars"), "west", "true"), "east", "true"), 5, v7, 5)
		e.set(fortressProperty(fortressProperty(structureState("iron_bars"), "west", "true"), "east", "true"), 6, v7, 5)
		e.set(fortressProperty(fortressProperty(structureState("iron_bars"), "west", "true"), "east", "true"), 7, v7, 5)
	}
	e.set(fortressProperty(fortressProperty(structureState("iron_bars"), "north", "true"), "south", "true"), 4, 3, 2)
	e.set(fortressProperty(fortressProperty(structureState("iron_bars"), "north", "true"), "south", "true"), 4, 3, 8)
	v8 := fortressProperty(structureState("iron_door"), "facing", "west")
	v9 := fortressProperty(fortressProperty(structureState("iron_door"), "facing", "west"), "half", "upper")
	e.set(v8, 4, 1, 2)
	e.set(v9, 4, 2, 2)
	e.set(v8, 4, 1, 8)
	e.set(v9, 4, 2, 8)
}

func placeStronghold_shlt(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 4, 4, 4, true, random)
	e.door(e.entryDoor(), 1, 1, 0)
	v7 := e.piece.Orientation
	if v7 == "north" || v7 == "east" {
		e.fill(0, 1, 1, 0, 3, 3, structureState("cave_air"), structureState("cave_air"), false)
	} else {
		e.fill(4, 1, 1, 4, 3, 3, structureState("cave_air"), structureState("cave_air"), false)
	}
}

func placeStronghold_shrt(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 4, 4, 4, true, random)
	e.door(e.entryDoor(), 1, 1, 0)
	v7 := e.piece.Orientation
	if v7 == "north" || v7 == "east" {
		e.fill(4, 1, 1, 4, 3, 3, structureState("cave_air"), structureState("cave_air"), false)
	} else {
		e.fill(0, 1, 1, 0, 3, 3, structureState("cave_air"), structureState("cave_air"), false)
	}
}

func placeStronghold_shssd(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 4, 10, 7, true, random)
	e.door(e.entryDoor(), 1, 7, 0)
	e.door("OPENING", 1, 1, 7)
	v7 := fortressProperty(structureState("cobblestone_stairs"), "facing", "south")
	for v8 := int32(0); v8 < 6; v8++ {
		e.set(v7, 1, 6-v8, 1+v8)
		e.set(v7, 2, 6-v8, 1+v8)
		e.set(v7, 3, 6-v8, 1+v8)
		if v8 >= 5 {
			continue
		}
		e.set(structureState("stone_bricks"), 1, 5-v8, 1+v8)
		e.set(structureState("stone_bricks"), 2, 5-v8, 1+v8)
		e.set(structureState("stone_bricks"), 3, 5-v8, 1+v8)
	}
}

func placeStronghold_shsd(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 4, 10, 4, true, random)
	e.door(e.entryDoor(), 1, 7, 0)
	e.door("OPENING", 1, 1, 4)
	e.set(structureState("stone_bricks"), 2, 6, 1)
	e.set(structureState("stone_bricks"), 1, 5, 1)
	e.set(structureState("smooth_stone_slab"), 1, 6, 1)
	e.set(structureState("stone_bricks"), 1, 5, 2)
	e.set(structureState("stone_bricks"), 1, 4, 3)
	e.set(structureState("smooth_stone_slab"), 1, 5, 3)
	e.set(structureState("stone_bricks"), 2, 4, 3)
	e.set(structureState("stone_bricks"), 3, 3, 3)
	e.set(structureState("smooth_stone_slab"), 3, 4, 3)
	e.set(structureState("stone_bricks"), 3, 3, 2)
	e.set(structureState("stone_bricks"), 3, 2, 1)
	e.set(structureState("smooth_stone_slab"), 3, 3, 1)
	e.set(structureState("stone_bricks"), 2, 2, 1)
	e.set(structureState("stone_bricks"), 1, 1, 1)
	e.set(structureState("smooth_stone_slab"), 1, 2, 1)
	e.set(structureState("stone_bricks"), 1, 1, 2)
	e.set(structureState("smooth_stone_slab"), 1, 1, 3)
}

func placeStronghold_sh5c(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 9, 8, 10, true, random)
	e.door(e.entryDoor(), 4, 3, 0)
	if e.flag("leftLow") {
		e.fill(0, 3, 1, 0, 5, 3, structureState("cave_air"), structureState("cave_air"), false)
	}
	if e.flag("rightLow") {
		e.fill(9, 3, 1, 9, 5, 3, structureState("cave_air"), structureState("cave_air"), false)
	}
	if e.flag("leftHigh") {
		e.fill(0, 5, 7, 0, 7, 9, structureState("cave_air"), structureState("cave_air"), false)
	}
	if e.flag("rightHigh") {
		e.fill(9, 5, 7, 9, 7, 9, structureState("cave_air"), structureState("cave_air"), false)
	}
	e.fill(5, 1, 10, 7, 3, 10, structureState("cave_air"), structureState("cave_air"), false)
	e.randomBox(1, 2, 1, 8, 2, 6, false, random)
	e.randomBox(4, 1, 5, 4, 4, 9, false, random)
	e.randomBox(8, 1, 5, 8, 4, 9, false, random)
	e.randomBox(1, 4, 7, 3, 4, 9, false, random)
	e.randomBox(1, 3, 5, 3, 3, 6, false, random)
	e.fill(1, 3, 4, 3, 3, 4, structureState("smooth_stone_slab"), structureState("smooth_stone_slab"), false)
	e.fill(1, 4, 6, 3, 4, 6, structureState("smooth_stone_slab"), structureState("smooth_stone_slab"), false)
	e.randomBox(5, 1, 7, 7, 1, 8, false, random)
	e.fill(5, 1, 9, 7, 1, 9, structureState("smooth_stone_slab"), structureState("smooth_stone_slab"), false)
	e.fill(5, 2, 7, 7, 2, 7, structureState("smooth_stone_slab"), structureState("smooth_stone_slab"), false)
	e.fill(4, 5, 7, 4, 5, 9, structureState("smooth_stone_slab"), structureState("smooth_stone_slab"), false)
	e.fill(8, 5, 7, 8, 5, 9, structureState("smooth_stone_slab"), structureState("smooth_stone_slab"), false)
	e.fill(5, 5, 7, 7, 5, 9, fortressProperty(structureState("smooth_stone_slab"), "type", "double"), fortressProperty(structureState("smooth_stone_slab"), "type", "double"), false)
	e.set(fortressProperty(structureState("wall_torch"), "facing", "south"), 6, 5, 6)
}

func placeStronghold_shcc(e *strongholdPieceEditor, random StructurePieceRandom) {
	e.randomBox(0, 0, 0, 4, 4, 6, true, random)
	e.door(e.entryDoor(), 1, 1, 0)
	e.door("OPENING", 1, 1, 6)
	e.fill(3, 1, 2, 3, 1, 4, structureState("stone_bricks"), structureState("stone_bricks"), false)
	e.set(structureState("stone_brick_slab"), 3, 1, 1)
	e.set(structureState("stone_brick_slab"), 3, 1, 5)
	e.set(structureState("stone_brick_slab"), 3, 2, 2)
	e.set(structureState("stone_brick_slab"), 3, 2, 4)
	for v7 := int32(2); v7 <= 4; v7++ {
		e.set(structureState("stone_brick_slab"), 2, 1, v7)
	}
	e.flaggedChest(random, 3, 2, 3)
}

func placeStronghold_shli(e *strongholdPieceEditor, random StructurePieceRandom) {
	v7 := int32(11)
	if !e.flag("Tall") {
		v7 = 6
	}
	e.randomBox(0, 0, 0, 13, v7-1, 14, true, random)
	e.door(e.entryDoor(), 4, 1, 0)
	e.randomFill(random, 0.07, 2, 1, 1, 11, 4, 13, structureState("cobweb"), structureState("cobweb"))
	for v10 := int32(1); v10 <= 13; v10++ {
		if (v10-1)%4 == 0 {
			e.fill(1, 1, v10, 1, 4, v10, structureState("oak_planks"), structureState("oak_planks"), false)
			e.fill(12, 1, v10, 12, 4, v10, structureState("oak_planks"), structureState("oak_planks"), false)
			e.set(fortressProperty(structureState("wall_torch"), "facing", "east"), 2, 3, v10)
			e.set(fortressProperty(structureState("wall_torch"), "facing", "west"), 11, 3, v10)
			if !e.flag("Tall") {
				continue
			}
			e.fill(1, 6, v10, 1, 9, v10, structureState("oak_planks"), structureState("oak_planks"), false)
			e.fill(12, 6, v10, 12, 9, v10, structureState("oak_planks"), structureState("oak_planks"), false)
			continue
		}
		e.fill(1, 1, v10, 1, 4, v10, structureState("bookshelf"), structureState("bookshelf"), false)
		e.fill(12, 1, v10, 12, 4, v10, structureState("bookshelf"), structureState("bookshelf"), false)
		if !e.flag("Tall") {
			continue
		}
		e.fill(1, 6, v10, 1, 9, v10, structureState("bookshelf"), structureState("bookshelf"), false)
		e.fill(12, 6, v10, 12, 9, v10, structureState("bookshelf"), structureState("bookshelf"), false)
	}
	for v11 := int32(3); v11 < 12; v11 += 2 {
		e.fill(3, 1, v11, 4, 3, v11, structureState("bookshelf"), structureState("bookshelf"), false)
		e.fill(6, 1, v11, 7, 3, v11, structureState("bookshelf"), structureState("bookshelf"), false)
		e.fill(9, 1, v11, 10, 3, v11, structureState("bookshelf"), structureState("bookshelf"), false)
	}
	if e.flag("Tall") {
		e.fill(1, 5, 1, 3, 5, 13, structureState("oak_planks"), structureState("oak_planks"), false)
		e.fill(10, 5, 1, 12, 5, 13, structureState("oak_planks"), structureState("oak_planks"), false)
		e.fill(4, 5, 1, 9, 5, 2, structureState("oak_planks"), structureState("oak_planks"), false)
		e.fill(4, 5, 12, 9, 5, 13, structureState("oak_planks"), structureState("oak_planks"), false)
		e.set(structureState("oak_planks"), 9, 5, 11)
		e.set(structureState("oak_planks"), 8, 5, 11)
		e.set(structureState("oak_planks"), 9, 5, 10)
		v12 := fortressProperty(fortressProperty(structureState("oak_fence"), "west", "true"), "east", "true")
		v13 := fortressProperty(fortressProperty(structureState("oak_fence"), "north", "true"), "south", "true")
		e.fill(3, 6, 3, 3, 6, 11, v13, v13, false)
		e.fill(10, 6, 3, 10, 6, 9, v13, v13, false)
		e.fill(4, 6, 2, 9, 6, 2, v12, v12, false)
		e.fill(4, 6, 12, 7, 6, 12, v12, v12, false)
		e.set(fortressProperty(fortressProperty(structureState("oak_fence"), "north", "true"), "east", "true"), 3, 6, 2)
		e.set(fortressProperty(fortressProperty(structureState("oak_fence"), "south", "true"), "east", "true"), 3, 6, 12)
		e.set(fortressProperty(fortressProperty(structureState("oak_fence"), "north", "true"), "west", "true"), 10, 6, 2)
		for v14 := int32(0); v14 <= 2; v14++ {
			e.set(fortressProperty(fortressProperty(structureState("oak_fence"), "south", "true"), "west", "true"), 8+v14, 6, 12-v14)
			if v14 == 2 {
				continue
			}
			e.set(fortressProperty(fortressProperty(structureState("oak_fence"), "north", "true"), "east", "true"), 8+v14, 6, 11-v14)
		}
		v15 := fortressProperty(structureState("ladder"), "facing", "south")
		e.set(v15, 10, 1, 13)
		e.set(v15, 10, 2, 13)
		e.set(v15, 10, 3, 13)
		e.set(v15, 10, 4, 13)
		e.set(v15, 10, 5, 13)
		e.set(v15, 10, 6, 13)
		e.set(v15, 10, 7, 13)
		v18 := fortressProperty(structureState("oak_fence"), "east", "true")
		e.set(v18, 6, 9, 7)
		v19 := fortressProperty(structureState("oak_fence"), "west", "true")
		e.set(v19, 7, 9, 7)
		e.set(v18, 6, 8, 7)
		e.set(v19, 7, 8, 7)
		v20 := fortressProperty(fortressProperty(v13, "west", "true"), "east", "true")
		e.set(v20, 6, 7, 7)
		e.set(v20, 7, 7, 7)
		e.set(v18, 5, 7, 7)
		e.set(v19, 8, 7, 7)
		e.set(fortressProperty(v18, "north", "true"), 6, 7, 6)
		e.set(fortressProperty(v18, "south", "true"), 6, 7, 8)
		e.set(fortressProperty(v19, "north", "true"), 7, 7, 6)
		e.set(fortressProperty(v19, "south", "true"), 7, 7, 8)
		v21 := structureState("torch")
		e.set(v21, 5, 8, 7)
		e.set(v21, 8, 8, 7)
		e.set(v21, 6, 8, 6)
		e.set(v21, 6, 8, 8)
		e.set(v21, 7, 8, 6)
		e.set(v21, 7, 8, 8)
	}
	e.chest(random, "stronghold_library", 3, 3, 5)
	if e.flag("Tall") {
		e.set(structureState("cave_air"), 12, 9, 1)
		e.chest(random, "stronghold_library", 12, 8, 1)
	}
}

func placeStronghold_shfc(e *strongholdPieceEditor, random StructurePieceRandom) {
	for v7 := int32(0); v7 < e.number("Steps"); v7++ {
		e.set(structureState("stone_bricks"), 0, 0, v7)
		e.set(structureState("stone_bricks"), 1, 0, v7)
		e.set(structureState("stone_bricks"), 2, 0, v7)
		e.set(structureState("stone_bricks"), 3, 0, v7)
		e.set(structureState("stone_bricks"), 4, 0, v7)
		for v8 := int32(1); v8 <= 3; v8++ {
			e.set(structureState("stone_bricks"), 0, v8, v7)
			e.set(structureState("cave_air"), 1, v8, v7)
			e.set(structureState("cave_air"), 2, v8, v7)
			e.set(structureState("cave_air"), 3, v8, v7)
			e.set(structureState("stone_bricks"), 4, v8, v7)
		}
		e.set(structureState("stone_bricks"), 0, 4, v7)
		e.set(structureState("stone_bricks"), 1, 4, v7)
		e.set(structureState("stone_bricks"), 2, 4, v7)
		e.set(structureState("stone_bricks"), 3, 4, v7)
		e.set(structureState("stone_bricks"), 4, 4, v7)
	}
}
