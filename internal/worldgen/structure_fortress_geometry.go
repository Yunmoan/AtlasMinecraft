package worldgen

func placeFortress_nebs(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 3, 0, 4, 4, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 5, 0, 3, 7, 18, structureState("air"), structureState("air"), false)
	e.fill(0, 5, 0, 0, 5, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 5, 0, 4, 5, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 4, 2, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 13, 4, 2, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 0, 0, 4, 1, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 0, 15, 4, 1, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v7 := int32(0); v7 <= 4; v7++ {
		for v8 := int32(0); v8 <= 2; v8++ {
			e.fillDown(structureState("nether_bricks"), v7, -1, v8)
			e.fillDown(structureState("nether_bricks"), v7, -1, 18-v8)
		}
	}
	v9 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	v10 := fortressProperty(v9, "east", "true")
	v11 := fortressProperty(v9, "west", "true")
	e.fill(0, 1, 1, 0, 4, 1, v10, v10, false)
	e.fill(0, 3, 4, 0, 4, 4, v10, v10, false)
	e.fill(0, 3, 14, 0, 4, 14, v10, v10, false)
	e.fill(0, 1, 17, 0, 4, 17, v10, v10, false)
	e.fill(4, 1, 1, 4, 4, 1, v11, v11, false)
	e.fill(4, 3, 4, 4, 4, 4, v11, v11, false)
	e.fill(4, 3, 14, 4, 4, 14, v11, v11, false)
	e.fill(4, 1, 17, 4, 4, 17, v11, v11, false)
}

func placeFortress_nebcr(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(7, 3, 0, 11, 4, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 3, 7, 18, 4, 11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(8, 5, 0, 10, 7, 18, structureState("air"), structureState("air"), false)
	e.fill(0, 5, 8, 18, 7, 10, structureState("air"), structureState("air"), false)
	e.fill(7, 5, 0, 7, 5, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(7, 5, 11, 7, 5, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(11, 5, 0, 11, 5, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(11, 5, 11, 11, 5, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 5, 7, 7, 5, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(11, 5, 7, 18, 5, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 5, 11, 7, 5, 11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(11, 5, 11, 18, 5, 11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(7, 2, 0, 11, 2, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(7, 2, 13, 11, 2, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(7, 0, 0, 11, 1, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(7, 0, 15, 11, 1, 18, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v7 := int32(7); v7 <= 11; v7++ {
		for v8 := int32(0); v8 <= 2; v8++ {
			e.fillDown(structureState("nether_bricks"), v7, -1, v8)
			e.fillDown(structureState("nether_bricks"), v7, -1, 18-v8)
		}
	}
	e.fill(0, 2, 7, 5, 2, 11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(13, 2, 7, 18, 2, 11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 0, 7, 3, 1, 11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(15, 0, 7, 18, 1, 11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v9 := int32(0); v9 <= 2; v9++ {
		for v10 := int32(7); v10 <= 11; v10++ {
			e.fillDown(structureState("nether_bricks"), v9, -1, v10)
			e.fillDown(structureState("nether_bricks"), 18-v9, -1, v10)
		}
	}
}

func placeFortress_nerc(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 0, 0, 6, 1, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 6, 7, 6, structureState("air"), structureState("air"), false)
	e.fill(0, 2, 0, 1, 6, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 6, 1, 6, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 2, 0, 6, 6, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 2, 6, 6, 6, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 0, 6, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 5, 0, 6, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 2, 0, 6, 6, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 2, 5, 6, 6, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	e.fill(2, 6, 0, 4, 6, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 0, 4, 5, 0, v7, v7, false)
	e.fill(2, 6, 6, 4, 6, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 6, 4, 5, 6, v7, v7, false)
	e.fill(0, 6, 2, 0, 6, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 5, 2, 0, 5, 4, v8, v8, false)
	e.fill(6, 6, 2, 6, 6, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 5, 2, 6, 5, 4, v8, v8, false)
	for v9 := int32(0); v9 <= 6; v9++ {
		for v10 := int32(0); v10 <= 6; v10++ {
			e.fillDown(structureState("nether_bricks"), v9, -1, v10)
		}
	}
}

func placeFortress_nesr(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 0, 0, 6, 1, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 6, 10, 6, structureState("air"), structureState("air"), false)
	e.fill(0, 2, 0, 1, 8, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 2, 0, 6, 8, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 1, 0, 8, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 2, 1, 6, 8, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 2, 6, 5, 8, 6, structureState("nether_bricks"), structureState("nether_bricks"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	e.fill(0, 3, 2, 0, 5, 4, v8, v8, false)
	e.fill(6, 3, 2, 6, 5, 2, v8, v8, false)
	e.fill(6, 3, 4, 6, 5, 4, v8, v8, false)
	e.set(structureState("nether_bricks"), 5, 2, 5)
	e.fill(4, 2, 5, 4, 3, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(3, 2, 5, 3, 4, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 2, 5, 2, 5, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 2, 5, 1, 6, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 7, 1, 5, 7, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 8, 2, 6, 8, 4, structureState("air"), structureState("air"), false)
	e.fill(2, 6, 0, 4, 8, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 0, 4, 5, 0, v7, v7, false)
	for v9 := int32(0); v9 <= 6; v9++ {
		for v10 := int32(0); v10 <= 6; v10++ {
			e.fillDown(structureState("nether_bricks"), v9, -1, v10)
		}
	}
}

func placeFortress_nemt(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 2, 0, 6, 7, 7, structureState("air"), structureState("air"), false)
	e.fill(1, 0, 0, 5, 1, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 2, 1, 5, 2, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 3, 2, 5, 3, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 4, 3, 5, 4, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 2, 0, 1, 4, 2, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 2, 0, 5, 4, 2, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 5, 2, 1, 5, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 5, 2, 5, 5, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 5, 3, 0, 5, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 5, 3, 6, 5, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 5, 8, 5, 5, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	e.set(fortressProperty(structureState("nether_brick_fence"), "west", "true"), 1, 6, 3)
	e.set(fortressProperty(structureState("nether_brick_fence"), "east", "true"), 5, 6, 3)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "east", "true"), "north", "true"), 0, 6, 3)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "north", "true"), 6, 6, 3)
	e.fill(0, 6, 4, 0, 6, 7, v8, v8, false)
	e.fill(6, 6, 4, 6, 6, 7, v8, v8, false)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "east", "true"), "south", "true"), 0, 6, 8)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "south", "true"), 6, 6, 8)
	e.fill(1, 6, 8, 5, 6, 8, v7, v7, false)
	e.set(fortressProperty(structureState("nether_brick_fence"), "east", "true"), 1, 7, 8)
	e.fill(2, 7, 8, 4, 7, 8, v7, v7, false)
	e.set(fortressProperty(structureState("nether_brick_fence"), "west", "true"), 5, 7, 8)
	e.set(fortressProperty(structureState("nether_brick_fence"), "east", "true"), 2, 8, 8)
	e.set(v7, 3, 8, 8)
	e.set(fortressProperty(structureState("nether_brick_fence"), "west", "true"), 4, 8, 8)
	e.spawner(random, 3, 5, 5)
	for v12 := int32(0); v12 <= 6; v12++ {
		for v13 := int32(0); v13 <= 6; v13++ {
			e.fillDown(structureState("nether_bricks"), v12, -1, v13)
		}
	}
}

func placeFortress_nece(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 3, 0, 12, 4, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 5, 0, 12, 13, 12, structureState("air"), structureState("air"), false)
	e.fill(0, 5, 0, 1, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(11, 5, 0, 12, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 11, 4, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(8, 5, 11, 10, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 9, 11, 7, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 0, 4, 12, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(8, 5, 0, 10, 12, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 9, 0, 7, 12, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 11, 2, 10, 12, 10, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 8, 0, 7, 8, 0, structureState("nether_brick_fence"), structureState("nether_brick_fence"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	for v9 := int32(1); v9 <= 11; v9 += 2 {
		e.fill(v9, 10, 0, v9, 11, 0, v7, v7, false)
		e.fill(v9, 10, 12, v9, 11, 12, v7, v7, false)
		e.fill(0, 10, v9, 0, 11, v9, v8, v8, false)
		e.fill(12, 10, v9, 12, 11, v9, v8, v8, false)
		e.set(structureState("nether_bricks"), v9, 13, 0)
		e.set(structureState("nether_bricks"), v9, 13, 12)
		e.set(structureState("nether_bricks"), 0, 13, v9)
		e.set(structureState("nether_bricks"), 12, 13, v9)
		if v9 == 11 {
			continue
		}
		e.set(v7, v9+1, 13, 0)
		e.set(v7, v9+1, 13, 12)
		e.set(v8, 0, 13, v9+1)
		e.set(v8, 12, 13, v9+1)
	}
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "east", "true"), 0, 13, 0)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "south", "true"), "east", "true"), 0, 13, 12)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "south", "true"), "west", "true"), 12, 13, 12)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "west", "true"), 12, 13, 0)
	for v10 := int32(3); v10 <= 9; v10 += 2 {
		e.fill(1, 7, v10, 1, 8, v10, fortressProperty(v8, "west", "true"), fortressProperty(v8, "west", "true"), false)
		e.fill(11, 7, v10, 11, 8, v10, fortressProperty(v8, "east", "true"), fortressProperty(v8, "east", "true"), false)
	}
	e.fill(4, 2, 0, 8, 2, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 4, 12, 2, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 0, 0, 8, 1, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 0, 9, 8, 1, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 0, 4, 3, 1, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(9, 0, 4, 12, 1, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v11 := int32(4); v11 <= 8; v11++ {
		for v12 := int32(0); v12 <= 2; v12++ {
			e.fillDown(structureState("nether_bricks"), v11, -1, v12)
			e.fillDown(structureState("nether_bricks"), v11, -1, 12-v12)
		}
	}
	for v13 := int32(0); v13 <= 2; v13++ {
		for v14 := int32(4); v14 <= 8; v14++ {
			e.fillDown(structureState("nether_bricks"), v13, -1, v14)
			e.fillDown(structureState("nether_bricks"), 12-v13, -1, v14)
		}
	}
	e.fill(5, 5, 5, 7, 5, 7, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 1, 6, 6, 4, 6, structureState("air"), structureState("air"), false)
	e.set(structureState("nether_bricks"), 6, 0, 6)
	e.set(structureState("lava"), 6, 5, 6)
	e.lavaTick(6, 5, 6)
}

func placeFortress_nesc(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 0, 0, 4, 1, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 4, 5, 4, structureState("air"), structureState("air"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	e.fill(0, 2, 0, 0, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 2, 0, 4, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 3, 1, 0, 4, 1, v7, v7, false)
	e.fill(0, 3, 3, 0, 4, 3, v7, v7, false)
	e.fill(4, 3, 1, 4, 4, 1, v7, v7, false)
	e.fill(4, 3, 3, 4, 4, 3, v7, v7, false)
	e.fill(0, 6, 0, 4, 6, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v8 := int32(0); v8 <= 4; v8++ {
		for v9 := int32(0); v9 <= 4; v9++ {
			e.fillDown(structureState("nether_bricks"), v8, -1, v9)
		}
	}
}

func placeFortress_nescrt(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 0, 0, 4, 1, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 4, 5, 4, structureState("air"), structureState("air"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	e.fill(0, 2, 0, 0, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 3, 1, 0, 4, 1, v8, v8, false)
	e.fill(0, 3, 3, 0, 4, 3, v8, v8, false)
	e.fill(4, 2, 0, 4, 5, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 2, 4, 4, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 3, 4, 1, 4, 4, v7, v7, false)
	e.fill(3, 3, 4, 3, 4, 4, v7, v7, false)
	e.chest(random, 1, 2, 3)
	e.fill(0, 6, 0, 4, 6, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v9 := int32(0); v9 <= 4; v9++ {
		for v10 := int32(0); v10 <= 4; v10++ {
			e.fillDown(structureState("nether_bricks"), v9, -1, v10)
		}
	}
}

func placeFortress_nesclt(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 0, 0, 4, 1, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 4, 5, 4, structureState("air"), structureState("air"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	e.fill(4, 2, 0, 4, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 3, 1, 4, 4, 1, v8, v8, false)
	e.fill(4, 3, 3, 4, 4, 3, v8, v8, false)
	e.fill(0, 2, 0, 0, 5, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 4, 3, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 3, 4, 1, 4, 4, v7, v7, false)
	e.fill(3, 3, 4, 3, 4, 4, v7, v7, false)
	e.chest(random, 3, 2, 3)
	e.fill(0, 6, 0, 4, 6, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v9 := int32(0); v9 <= 4; v9++ {
		for v10 := int32(0); v10 <= 4; v10++ {
			e.fillDown(structureState("nether_bricks"), v9, -1, v10)
		}
	}
}

func placeFortress_neccs(e *fortressPieceEditor, random StructurePieceRandom) {
	v7 := fortressProperty(structureState("nether_brick_stairs"), "facing", "south")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	for v9 := int32(0); v9 <= 9; v9++ {
		v10 := int32(max(1, 7-v9))
		v11 := int32(min(max(v10+5, 14-v9), 13))
		v12 := int32(v9)
		e.fill(0, 0, v12, 4, v10, v12, structureState("nether_bricks"), structureState("nether_bricks"), false)
		e.fill(1, v10+1, v12, 3, v11-1, v12, structureState("air"), structureState("air"), false)
		if v9 <= 6 {
			e.set(v7, 1, v10+1, v12)
			e.set(v7, 2, v10+1, v12)
			e.set(v7, 3, v10+1, v12)
		}
		e.fill(0, v11, v12, 4, v11, v12, structureState("nether_bricks"), structureState("nether_bricks"), false)
		e.fill(0, v10+1, v12, 0, v11-1, v12, structureState("nether_bricks"), structureState("nether_bricks"), false)
		e.fill(4, v10+1, v12, 4, v11-1, v12, structureState("nether_bricks"), structureState("nether_bricks"), false)
		if (v9 & 1) == 0 {
			e.fill(0, v10+2, v12, 0, v10+3, v12, v8, v8, false)
			e.fill(4, v10+2, v12, 4, v10+3, v12, v8, v8, false)
		}
		for v13 := int32(0); v13 <= 4; v13++ {
			e.fillDown(structureState("nether_bricks"), v13, -1, v12)
		}
	}
}

func placeFortress_nectb(e *fortressPieceEditor, random StructurePieceRandom) {
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	e.fill(0, 0, 0, 8, 1, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 8, 5, 8, structureState("air"), structureState("air"), false)
	e.fill(0, 6, 0, 8, 6, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 2, 5, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 2, 0, 8, 5, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 3, 0, 1, 4, 0, v8, v8, false)
	e.fill(7, 3, 0, 7, 4, 0, v8, v8, false)
	e.fill(0, 2, 4, 8, 2, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 1, 4, 2, 2, 4, structureState("air"), structureState("air"), false)
	e.fill(6, 1, 4, 7, 2, 4, structureState("air"), structureState("air"), false)
	e.fill(1, 3, 8, 7, 3, 8, v8, v8, false)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "east", "true"), "south", "true"), 0, 3, 8)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "south", "true"), 8, 3, 8)
	e.fill(0, 3, 6, 0, 3, 7, v7, v7, false)
	e.fill(8, 3, 6, 8, 3, 7, v7, v7, false)
	e.fill(0, 3, 4, 0, 5, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(8, 3, 4, 8, 5, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 3, 5, 2, 5, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(6, 3, 5, 7, 5, 5, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(1, 4, 5, 1, 5, 5, v8, v8, false)
	e.fill(7, 4, 5, 7, 5, 5, v8, v8, false)
	for v9 := int32(0); v9 <= 5; v9++ {
		for v10 := int32(0); v10 <= 8; v10++ {
			e.fillDown(structureState("nether_bricks"), v10, -1, v9)
		}
	}
}

func placeFortress_nescsc(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 0, 0, 4, 1, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 0, 4, 5, 4, structureState("air"), structureState("air"), false)
	e.fill(0, 2, 0, 0, 5, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 2, 0, 4, 5, 0, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 4, 0, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 2, 4, 4, 5, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 6, 0, 4, 6, 4, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v7 := int32(0); v7 <= 4; v7++ {
		for v8 := int32(0); v8 <= 4; v8++ {
			e.fillDown(structureState("nether_bricks"), v7, -1, v8)
		}
	}
}

func placeFortress_necsr(e *fortressPieceEditor, random StructurePieceRandom) {
	e.fill(0, 3, 0, 12, 4, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 5, 0, 12, 13, 12, structureState("air"), structureState("air"), false)
	e.fill(0, 5, 0, 1, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(11, 5, 0, 12, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 11, 4, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(8, 5, 11, 10, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 9, 11, 7, 12, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 0, 4, 12, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(8, 5, 0, 10, 12, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(5, 9, 0, 7, 12, 1, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 11, 2, 10, 12, 10, structureState("nether_bricks"), structureState("nether_bricks"), false)
	v7 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "west", "true"), "east", "true")
	v8 := fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "south", "true")
	v9 := fortressProperty(v8, "west", "true")
	v10 := fortressProperty(v8, "east", "true")
	for v11 := int32(1); v11 <= 11; v11 += 2 {
		e.fill(v11, 10, 0, v11, 11, 0, v7, v7, false)
		e.fill(v11, 10, 12, v11, 11, 12, v7, v7, false)
		e.fill(0, 10, v11, 0, 11, v11, v8, v8, false)
		e.fill(12, 10, v11, 12, 11, v11, v8, v8, false)
		e.set(structureState("nether_bricks"), v11, 13, 0)
		e.set(structureState("nether_bricks"), v11, 13, 12)
		e.set(structureState("nether_bricks"), 0, 13, v11)
		e.set(structureState("nether_bricks"), 12, 13, v11)
		if v11 == 11 {
			continue
		}
		e.set(v7, v11+1, 13, 0)
		e.set(v7, v11+1, 13, 12)
		e.set(v8, 0, 13, v11+1)
		e.set(v8, 12, 13, v11+1)
	}
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "east", "true"), 0, 13, 0)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "south", "true"), "east", "true"), 0, 13, 12)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "south", "true"), "west", "true"), 12, 13, 12)
	e.set(fortressProperty(fortressProperty(structureState("nether_brick_fence"), "north", "true"), "west", "true"), 12, 13, 0)
	for v12 := int32(3); v12 <= 9; v12 += 2 {
		e.fill(1, 7, v12, 1, 8, v12, v9, v9, false)
		e.fill(11, 7, v12, 11, 8, v12, v10, v10, false)
	}
	v13 := fortressProperty(structureState("nether_brick_stairs"), "facing", "north")
	for v14 := int32(0); v14 <= 6; v14++ {
		v15 := int32(v14 + 4)
		for v16 := int32(5); v16 <= 7; v16++ {
			e.set(v13, v16, 5+v14, v15)
		}
		if v15 >= 5 && v15 <= 8 {
			e.fill(5, 5, v15, 7, v14+4, v15, structureState("nether_bricks"), structureState("nether_bricks"), false)
		} else if v15 >= 9 && v15 <= 10 {
			e.fill(5, 8, v15, 7, v14+4, v15, structureState("nether_bricks"), structureState("nether_bricks"), false)
		}
		if v14 < 1 {
			continue
		}
		e.fill(5, 6+v14, v15, 7, 9+v14, v15, structureState("air"), structureState("air"), false)
	}
	for v17 := int32(5); v17 <= 7; v17++ {
		e.set(v13, v17, 12, 11)
	}
	e.fill(5, 6, 7, 5, 7, 7, v10, v10, false)
	e.fill(7, 6, 7, 7, 7, 7, v9, v9, false)
	e.fill(5, 13, 12, 7, 13, 12, structureState("air"), structureState("air"), false)
	e.fill(2, 5, 2, 3, 5, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 9, 3, 5, 10, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(2, 5, 4, 2, 5, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(9, 5, 2, 10, 5, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(9, 5, 9, 10, 5, 10, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(10, 5, 4, 10, 5, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	v18 := fortressProperty(v13, "facing", "east")
	v19 := fortressProperty(v13, "facing", "west")
	e.set(v19, 4, 5, 2)
	e.set(v19, 4, 5, 3)
	e.set(v19, 4, 5, 9)
	e.set(v19, 4, 5, 10)
	e.set(v18, 8, 5, 2)
	e.set(v18, 8, 5, 3)
	e.set(v18, 8, 5, 9)
	e.set(v18, 8, 5, 10)
	e.fill(3, 4, 4, 4, 4, 8, structureState("soul_sand"), structureState("soul_sand"), false)
	e.fill(8, 4, 4, 9, 4, 8, structureState("soul_sand"), structureState("soul_sand"), false)
	e.fill(3, 5, 4, 4, 5, 8, structureState("nether_wart"), structureState("nether_wart"), false)
	e.fill(8, 5, 4, 9, 5, 8, structureState("nether_wart"), structureState("nether_wart"), false)
	e.fill(4, 2, 0, 8, 2, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 2, 4, 12, 2, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 0, 0, 8, 1, 3, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(4, 0, 9, 8, 1, 12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(0, 0, 4, 3, 1, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	e.fill(9, 0, 4, 12, 1, 8, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v20 := int32(4); v20 <= 8; v20++ {
		for v21 := int32(0); v21 <= 2; v21++ {
			e.fillDown(structureState("nether_bricks"), v20, -1, v21)
			e.fillDown(structureState("nether_bricks"), v20, -1, 12-v21)
		}
	}
	for v22 := int32(0); v22 <= 2; v22++ {
		for v23 := int32(4); v23 <= 8; v23++ {
			e.fillDown(structureState("nether_bricks"), v22, -1, v23)
			e.fillDown(structureState("nether_bricks"), 12-v22, -1, v23)
		}
	}
}

func placeFortress_nebef(e *fortressPieceEditor, random StructurePieceRandom) {
	seed, _ := e.piece.Data["Seed"].Number()
	fillerRandom := NewLegacyRandom(int64(int32(seed)))
	for v8 := int32(0); v8 <= 4; v8++ {
		for v9 := int32(3); v9 <= 4; v9++ {
			v10 := int32(fillerRandom.NextIntBound(8))
			e.fill(v8, v9, 0, v8, v9, v10, structureState("nether_bricks"), structureState("nether_bricks"), false)
		}
	}
	v11 := int32(fillerRandom.NextIntBound(8))
	e.fill(0, 5, 0, 0, 5, v11, structureState("nether_bricks"), structureState("nether_bricks"), false)
	v12 := int32(fillerRandom.NextIntBound(8))
	e.fill(4, 5, 0, 4, 5, v12, structureState("nether_bricks"), structureState("nether_bricks"), false)
	for v13 := int32(0); v13 <= 4; v13++ {
		v14 := int32(fillerRandom.NextIntBound(5))
		e.fill(v13, 2, 0, v13, 2, v14, structureState("nether_bricks"), structureState("nether_bricks"), false)
	}
	for v15 := int32(0); v15 <= 4; v15++ {
		for v16 := int32(0); v16 <= 1; v16++ {
			v17 := int32(fillerRandom.NextIntBound(3))
			e.fill(v15, v16, 0, v15, v16, v17, structureState("nether_bricks"), structureState("nether_bricks"), false)
		}
	}
}
