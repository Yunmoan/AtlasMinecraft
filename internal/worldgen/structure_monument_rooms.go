package worldgen

// Original 1.21.1 monument room geometry. Room graph and variants are supplied
// by assembleMonument; block calls retain the original clipping and RNG order.

func (e *monumentPieceEditor) placePenthouse(random StructurePieceRandom) {
	e.fill(2, -1, 2, 11, -1, 11, monumentBricks, monumentBricks, false)
	e.fill(0, -1, 0, 1, -1, 11, monumentPrismarine, monumentPrismarine, false)
	e.fill(12, -1, 0, 13, -1, 11, monumentPrismarine, monumentPrismarine, false)
	e.fill(2, -1, 0, 11, -1, 1, monumentPrismarine, monumentPrismarine, false)
	e.fill(2, -1, 12, 11, -1, 13, monumentPrismarine, monumentPrismarine, false)
	e.fill(0, 0, 0, 0, 0, 13, monumentBricks, monumentBricks, false)
	e.fill(13, 0, 0, 13, 0, 13, monumentBricks, monumentBricks, false)
	e.fill(1, 0, 0, 12, 0, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 0, 13, 12, 0, 13, monumentBricks, monumentBricks, false)
	for v7 := int32(2); v7 <= 11; v7 += 3 {
		e.set(monumentLantern, 0, 0, v7)
		e.set(monumentLantern, 13, 0, v7)
		e.set(monumentLantern, v7, 0, 0)
	}
	e.fill(2, 0, 3, 4, 0, 9, monumentBricks, monumentBricks, false)
	e.fill(9, 0, 3, 11, 0, 9, monumentBricks, monumentBricks, false)
	e.fill(4, 0, 9, 9, 0, 11, monumentBricks, monumentBricks, false)
	e.set(monumentBricks, 5, 0, 8)
	e.set(monumentBricks, 8, 0, 8)
	e.set(monumentBricks, 10, 0, 10)
	e.set(monumentBricks, 3, 0, 10)
	e.fill(3, 0, 3, 3, 0, 7, monumentDark, monumentDark, false)
	e.fill(10, 0, 3, 10, 0, 7, monumentDark, monumentDark, false)
	e.fill(6, 0, 10, 7, 0, 10, monumentDark, monumentDark, false)
	v8 := int32(3)
	for v9 := int32(0); v9 < 2; v9++ {
		for v10 := int32(2); v10 <= 8; v10 += 3 {
			e.fill(v8, 0, v10, v8, 2, v10, monumentBricks, monumentBricks, false)
		}
		v8 = 10
	}
	e.fill(5, 0, 10, 5, 2, 10, monumentBricks, monumentBricks, false)
	e.fill(8, 0, 10, 8, 2, 10, monumentBricks, monumentBricks, false)
	e.fill(6, -1, 7, 7, -1, 8, monumentDark, monumentDark, false)
	e.waterBox(6, -1, 3, 7, -1, 4)
	e.spawnGuardian(6, 1, 6)
}

func (e *monumentPieceEditor) placeWing(random StructurePieceRandom) {
	if e.child.variant == 0 {
		for v7 := int32(0); v7 < 4; v7++ {
			e.fill(10-v7, 3-v7, 20-v7, 12+v7, 3-v7, 20, monumentBricks, monumentBricks, false)
		}
		e.fill(7, 0, 6, 15, 0, 16, monumentBricks, monumentBricks, false)
		e.fill(6, 0, 6, 6, 3, 20, monumentBricks, monumentBricks, false)
		e.fill(16, 0, 6, 16, 3, 20, monumentBricks, monumentBricks, false)
		e.fill(7, 1, 7, 7, 1, 20, monumentBricks, monumentBricks, false)
		e.fill(15, 1, 7, 15, 1, 20, monumentBricks, monumentBricks, false)
		e.fill(7, 1, 6, 9, 3, 6, monumentBricks, monumentBricks, false)
		e.fill(13, 1, 6, 15, 3, 6, monumentBricks, monumentBricks, false)
		e.fill(8, 1, 7, 9, 1, 7, monumentBricks, monumentBricks, false)
		e.fill(13, 1, 7, 14, 1, 7, monumentBricks, monumentBricks, false)
		e.fill(9, 0, 5, 13, 0, 5, monumentBricks, monumentBricks, false)
		e.fill(10, 0, 7, 12, 0, 7, monumentDark, monumentDark, false)
		e.fill(8, 0, 10, 8, 0, 12, monumentDark, monumentDark, false)
		e.fill(14, 0, 10, 14, 0, 12, monumentDark, monumentDark, false)
		for v8 := int32(18); v8 >= 7; v8 -= 3 {
			e.set(monumentLantern, 6, 3, v8)
			e.set(monumentLantern, 16, 3, v8)
		}
		e.set(monumentLantern, 10, 0, 10)
		e.set(monumentLantern, 12, 0, 10)
		e.set(monumentLantern, 10, 0, 12)
		e.set(monumentLantern, 12, 0, 12)
		e.set(monumentLantern, 8, 3, 6)
		e.set(monumentLantern, 14, 3, 6)
		e.set(monumentBricks, 4, 2, 4)
		e.set(monumentLantern, 4, 1, 4)
		e.set(monumentBricks, 4, 0, 4)
		e.set(monumentBricks, 18, 2, 4)
		e.set(monumentLantern, 18, 1, 4)
		e.set(monumentBricks, 18, 0, 4)
		e.set(monumentBricks, 4, 2, 18)
		e.set(monumentLantern, 4, 1, 18)
		e.set(monumentBricks, 4, 0, 18)
		e.set(monumentBricks, 18, 2, 18)
		e.set(monumentLantern, 18, 1, 18)
		e.set(monumentBricks, 18, 0, 18)
		e.set(monumentBricks, 9, 7, 20)
		e.set(monumentBricks, 13, 7, 20)
		e.fill(6, 0, 21, 7, 4, 21, monumentBricks, monumentBricks, false)
		e.fill(15, 0, 21, 16, 4, 21, monumentBricks, monumentBricks, false)
		e.spawnGuardian(11, 2, 16)
	} else if e.child.variant == 1 {
		e.fill(9, 3, 18, 13, 3, 20, monumentBricks, monumentBricks, false)
		e.fill(9, 0, 18, 9, 2, 18, monumentBricks, monumentBricks, false)
		e.fill(13, 0, 18, 13, 2, 18, monumentBricks, monumentBricks, false)
		v9 := int32(9)
		for v12 := int32(0); v12 < 2; v12++ {
			e.set(monumentBricks, v9, 6, 20)
			e.set(monumentLantern, v9, 5, 20)
			e.set(monumentBricks, v9, 4, 20)
			v9 = 13
		}
		e.fill(7, 3, 7, 15, 3, 14, monumentBricks, monumentBricks, false)
		v9 = 10
		for v13 := int32(0); v13 < 2; v13++ {
			e.fill(v9, 0, 10, v9, 6, 10, monumentBricks, monumentBricks, false)
			e.fill(v9, 0, 12, v9, 6, 12, monumentBricks, monumentBricks, false)
			e.set(monumentLantern, v9, 0, 10)
			e.set(monumentLantern, v9, 0, 12)
			e.set(monumentLantern, v9, 4, 10)
			e.set(monumentLantern, v9, 4, 12)
			v9 = 12
		}
		v9 = 8
		for v14 := int32(0); v14 < 2; v14++ {
			e.fill(v9, 0, 7, v9, 2, 7, monumentBricks, monumentBricks, false)
			e.fill(v9, 0, 14, v9, 2, 14, monumentBricks, monumentBricks, false)
			v9 = 14
		}
		e.fill(8, 3, 8, 8, 3, 13, monumentDark, monumentDark, false)
		e.fill(14, 3, 8, 14, 3, 13, monumentDark, monumentDark, false)
		e.spawnGuardian(11, 5, 13)
	}
}

func (e *monumentPieceEditor) placeCore(random StructurePieceRandom) {
	e.waterOnlyFill(1, 8, 0, 14, 8, 14, monumentPrismarine)
	v8 := monumentBricks
	e.fill(0, 7, 0, 0, 7, 15, v8, v8, false)
	e.fill(15, 7, 0, 15, 7, 15, v8, v8, false)
	e.fill(1, 7, 0, 15, 7, 0, v8, v8, false)
	e.fill(1, 7, 15, 14, 7, 15, v8, v8, false)
	for v9 := int32(1); v9 <= 6; v9++ {
		v10 := monumentBricks
		if v9 == 2 || v9 == 6 {
			v10 = monumentPrismarine
		}
		for v11 := int32(0); v11 <= 15; v11 += 15 {
			e.fill(v11, v9, 0, v11, v9, 1, v10, v10, false)
			e.fill(v11, v9, 6, v11, v9, 9, v10, v10, false)
			e.fill(v11, v9, 14, v11, v9, 15, v10, v10, false)
		}
		e.fill(1, v9, 0, 1, v9, 0, v10, v10, false)
		e.fill(6, v9, 0, 9, v9, 0, v10, v10, false)
		e.fill(14, v9, 0, 14, v9, 0, v10, v10, false)
		e.fill(1, v9, 15, 14, v9, 15, v10, v10, false)
	}
	e.fill(6, 3, 6, 9, 6, 9, monumentDark, monumentDark, false)
	e.fill(7, 4, 7, 8, 5, 8, structureState("gold_block"), structureState("gold_block"), false)
	for v12 := int32(3); v12 <= 6; v12 += 3 {
		for v13 := int32(6); v13 <= 9; v13 += 3 {
			e.set(monumentLantern, v13, v12, 6)
			e.set(monumentLantern, v13, v12, 9)
		}
	}
	e.fill(5, 1, 6, 5, 2, 6, monumentBricks, monumentBricks, false)
	e.fill(5, 1, 9, 5, 2, 9, monumentBricks, monumentBricks, false)
	e.fill(10, 1, 6, 10, 2, 6, monumentBricks, monumentBricks, false)
	e.fill(10, 1, 9, 10, 2, 9, monumentBricks, monumentBricks, false)
	e.fill(6, 1, 5, 6, 2, 5, monumentBricks, monumentBricks, false)
	e.fill(9, 1, 5, 9, 2, 5, monumentBricks, monumentBricks, false)
	e.fill(6, 1, 10, 6, 2, 10, monumentBricks, monumentBricks, false)
	e.fill(9, 1, 10, 9, 2, 10, monumentBricks, monumentBricks, false)
	e.fill(5, 2, 5, 5, 6, 5, monumentBricks, monumentBricks, false)
	e.fill(5, 2, 10, 5, 6, 10, monumentBricks, monumentBricks, false)
	e.fill(10, 2, 5, 10, 6, 5, monumentBricks, monumentBricks, false)
	e.fill(10, 2, 10, 10, 6, 10, monumentBricks, monumentBricks, false)
	e.fill(5, 7, 1, 5, 7, 6, monumentBricks, monumentBricks, false)
	e.fill(10, 7, 1, 10, 7, 6, monumentBricks, monumentBricks, false)
	e.fill(5, 7, 9, 5, 7, 14, monumentBricks, monumentBricks, false)
	e.fill(10, 7, 9, 10, 7, 14, monumentBricks, monumentBricks, false)
	e.fill(1, 7, 5, 6, 7, 5, monumentBricks, monumentBricks, false)
	e.fill(1, 7, 10, 6, 7, 10, monumentBricks, monumentBricks, false)
	e.fill(9, 7, 5, 14, 7, 5, monumentBricks, monumentBricks, false)
	e.fill(9, 7, 10, 14, 7, 10, monumentBricks, monumentBricks, false)
	e.fill(2, 1, 2, 2, 1, 3, monumentBricks, monumentBricks, false)
	e.fill(3, 1, 2, 3, 1, 2, monumentBricks, monumentBricks, false)
	e.fill(13, 1, 2, 13, 1, 3, monumentBricks, monumentBricks, false)
	e.fill(12, 1, 2, 12, 1, 2, monumentBricks, monumentBricks, false)
	e.fill(2, 1, 12, 2, 1, 13, monumentBricks, monumentBricks, false)
	e.fill(3, 1, 13, 3, 1, 13, monumentBricks, monumentBricks, false)
	e.fill(13, 1, 12, 13, 1, 13, monumentBricks, monumentBricks, false)
	e.fill(12, 1, 13, 12, 1, 13, monumentBricks, monumentBricks, false)
}

func (e *monumentPieceEditor) placeDoubleYZ(random StructurePieceRandom) {
	v7 := e.child.room.connections[2]
	v8 := e.child.room
	v9 := v7.connections[1]
	v10 := v8.connections[1]
	if e.child.room.id/25 > 0 {
		e.floor(0, 8, v7.open[0])
		e.floor(0, 0, v8.open[0])
	}
	if v10.connections[1] == nil {
		e.waterOnlyFill(1, 8, 1, 6, 8, 7, monumentPrismarine)
	}
	if v9.connections[1] == nil {
		e.waterOnlyFill(1, 8, 8, 6, 8, 14, monumentPrismarine)
	}
	for v11 := int32(1); v11 <= 7; v11++ {
		v12 := monumentBricks
		if v11 == 2 || v11 == 6 {
			v12 = monumentPrismarine
		}
		e.fill(0, v11, 0, 0, v11, 15, v12, v12, false)
		e.fill(7, v11, 0, 7, v11, 15, v12, v12, false)
		e.fill(1, v11, 0, 6, v11, 0, v12, v12, false)
		e.fill(1, v11, 15, 6, v11, 15, v12, v12, false)
	}
	for v13 := int32(1); v13 <= 7; v13++ {
		v14 := monumentDark
		if v13 == 2 || v13 == 6 {
			v14 = monumentLantern
		}
		e.fill(3, v13, 7, 4, v13, 8, v14, v14, false)
	}
	if v8.open[3] {
		e.waterBox(3, 1, 0, 4, 2, 0)
	}
	if v8.open[5] {
		e.waterBox(7, 1, 3, 7, 2, 4)
	}
	if v8.open[4] {
		e.waterBox(0, 1, 3, 0, 2, 4)
	}
	if v7.open[2] {
		e.waterBox(3, 1, 15, 4, 2, 15)
	}
	if v7.open[4] {
		e.waterBox(0, 1, 11, 0, 2, 12)
	}
	if v7.open[5] {
		e.waterBox(7, 1, 11, 7, 2, 12)
	}
	if v10.open[3] {
		e.waterBox(3, 5, 0, 4, 6, 0)
	}
	if v10.open[5] {
		e.waterBox(7, 5, 3, 7, 6, 4)
		e.fill(5, 4, 2, 6, 4, 5, monumentBricks, monumentBricks, false)
		e.fill(6, 1, 2, 6, 3, 2, monumentBricks, monumentBricks, false)
		e.fill(6, 1, 5, 6, 3, 5, monumentBricks, monumentBricks, false)
	}
	if v10.open[4] {
		e.waterBox(0, 5, 3, 0, 6, 4)
		e.fill(1, 4, 2, 2, 4, 5, monumentBricks, monumentBricks, false)
		e.fill(1, 1, 2, 1, 3, 2, monumentBricks, monumentBricks, false)
		e.fill(1, 1, 5, 1, 3, 5, monumentBricks, monumentBricks, false)
	}
	if v9.open[2] {
		e.waterBox(3, 5, 15, 4, 6, 15)
	}
	if v9.open[4] {
		e.waterBox(0, 5, 11, 0, 6, 12)
		e.fill(1, 4, 10, 2, 4, 13, monumentBricks, monumentBricks, false)
		e.fill(1, 1, 10, 1, 3, 10, monumentBricks, monumentBricks, false)
		e.fill(1, 1, 13, 1, 3, 13, monumentBricks, monumentBricks, false)
	}
	if v9.open[5] {
		e.waterBox(7, 5, 11, 7, 6, 12)
		e.fill(5, 4, 10, 6, 4, 13, monumentBricks, monumentBricks, false)
		e.fill(6, 1, 10, 6, 3, 10, monumentBricks, monumentBricks, false)
		e.fill(6, 1, 13, 6, 3, 13, monumentBricks, monumentBricks, false)
	}
}

func (e *monumentPieceEditor) placeDoubleXY(random StructurePieceRandom) {
	v7 := e.child.room.connections[5]
	v8 := e.child.room
	v9 := v8.connections[1]
	v10 := v7.connections[1]
	if e.child.room.id/25 > 0 {
		e.floor(8, 0, v7.open[0])
		e.floor(0, 0, v8.open[0])
	}
	if v9.connections[1] == nil {
		e.waterOnlyFill(1, 8, 1, 7, 8, 6, monumentPrismarine)
	}
	if v10.connections[1] == nil {
		e.waterOnlyFill(8, 8, 1, 14, 8, 6, monumentPrismarine)
	}
	for v11 := int32(1); v11 <= 7; v11++ {
		v12 := monumentBricks
		if v11 == 2 || v11 == 6 {
			v12 = monumentPrismarine
		}
		e.fill(0, v11, 0, 0, v11, 7, v12, v12, false)
		e.fill(15, v11, 0, 15, v11, 7, v12, v12, false)
		e.fill(1, v11, 0, 15, v11, 0, v12, v12, false)
		e.fill(1, v11, 7, 14, v11, 7, v12, v12, false)
	}
	e.fill(2, 1, 3, 2, 7, 4, monumentBricks, monumentBricks, false)
	e.fill(3, 1, 2, 4, 7, 2, monumentBricks, monumentBricks, false)
	e.fill(3, 1, 5, 4, 7, 5, monumentBricks, monumentBricks, false)
	e.fill(13, 1, 3, 13, 7, 4, monumentBricks, monumentBricks, false)
	e.fill(11, 1, 2, 12, 7, 2, monumentBricks, monumentBricks, false)
	e.fill(11, 1, 5, 12, 7, 5, monumentBricks, monumentBricks, false)
	e.fill(5, 1, 3, 5, 3, 4, monumentBricks, monumentBricks, false)
	e.fill(10, 1, 3, 10, 3, 4, monumentBricks, monumentBricks, false)
	e.fill(5, 7, 2, 10, 7, 5, monumentBricks, monumentBricks, false)
	e.fill(5, 5, 2, 5, 7, 2, monumentBricks, monumentBricks, false)
	e.fill(10, 5, 2, 10, 7, 2, monumentBricks, monumentBricks, false)
	e.fill(5, 5, 5, 5, 7, 5, monumentBricks, monumentBricks, false)
	e.fill(10, 5, 5, 10, 7, 5, monumentBricks, monumentBricks, false)
	e.set(monumentBricks, 6, 6, 2)
	e.set(monumentBricks, 9, 6, 2)
	e.set(monumentBricks, 6, 6, 5)
	e.set(monumentBricks, 9, 6, 5)
	e.fill(5, 4, 3, 6, 4, 4, monumentBricks, monumentBricks, false)
	e.fill(9, 4, 3, 10, 4, 4, monumentBricks, monumentBricks, false)
	e.set(monumentLantern, 5, 4, 2)
	e.set(monumentLantern, 5, 4, 5)
	e.set(monumentLantern, 10, 4, 2)
	e.set(monumentLantern, 10, 4, 5)
	if v8.open[3] {
		e.waterBox(3, 1, 0, 4, 2, 0)
	}
	if v8.open[2] {
		e.waterBox(3, 1, 7, 4, 2, 7)
	}
	if v8.open[4] {
		e.waterBox(0, 1, 3, 0, 2, 4)
	}
	if v7.open[3] {
		e.waterBox(11, 1, 0, 12, 2, 0)
	}
	if v7.open[2] {
		e.waterBox(11, 1, 7, 12, 2, 7)
	}
	if v7.open[5] {
		e.waterBox(15, 1, 3, 15, 2, 4)
	}
	if v9.open[3] {
		e.waterBox(3, 5, 0, 4, 6, 0)
	}
	if v9.open[2] {
		e.waterBox(3, 5, 7, 4, 6, 7)
	}
	if v9.open[4] {
		e.waterBox(0, 5, 3, 0, 6, 4)
	}
	if v10.open[3] {
		e.waterBox(11, 5, 0, 12, 6, 0)
	}
	if v10.open[2] {
		e.waterBox(11, 5, 7, 12, 6, 7)
	}
	if v10.open[5] {
		e.waterBox(15, 5, 3, 15, 6, 4)
	}
}

func (e *monumentPieceEditor) placeDoubleZ(random StructurePieceRandom) {
	v7 := e.child.room.connections[2]
	v8 := e.child.room
	if e.child.room.id/25 > 0 {
		e.floor(0, 8, v7.open[0])
		e.floor(0, 0, v8.open[0])
	}
	if v8.connections[1] == nil {
		e.waterOnlyFill(1, 4, 1, 6, 4, 7, monumentPrismarine)
	}
	if v7.connections[1] == nil {
		e.waterOnlyFill(1, 4, 8, 6, 4, 14, monumentPrismarine)
	}
	e.fill(0, 3, 0, 0, 3, 15, monumentBricks, monumentBricks, false)
	e.fill(7, 3, 0, 7, 3, 15, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 0, 7, 3, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 15, 6, 3, 15, monumentBricks, monumentBricks, false)
	e.fill(0, 2, 0, 0, 2, 15, monumentPrismarine, monumentPrismarine, false)
	e.fill(7, 2, 0, 7, 2, 15, monumentPrismarine, monumentPrismarine, false)
	e.fill(1, 2, 0, 7, 2, 0, monumentPrismarine, monumentPrismarine, false)
	e.fill(1, 2, 15, 6, 2, 15, monumentPrismarine, monumentPrismarine, false)
	e.fill(0, 1, 0, 0, 1, 15, monumentBricks, monumentBricks, false)
	e.fill(7, 1, 0, 7, 1, 15, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 0, 7, 1, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 15, 6, 1, 15, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 1, 1, 1, 2, monumentBricks, monumentBricks, false)
	e.fill(6, 1, 1, 6, 1, 2, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 1, 1, 3, 2, monumentBricks, monumentBricks, false)
	e.fill(6, 3, 1, 6, 3, 2, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 13, 1, 1, 14, monumentBricks, monumentBricks, false)
	e.fill(6, 1, 13, 6, 1, 14, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 13, 1, 3, 14, monumentBricks, monumentBricks, false)
	e.fill(6, 3, 13, 6, 3, 14, monumentBricks, monumentBricks, false)
	e.fill(2, 1, 6, 2, 3, 6, monumentBricks, monumentBricks, false)
	e.fill(5, 1, 6, 5, 3, 6, monumentBricks, monumentBricks, false)
	e.fill(2, 1, 9, 2, 3, 9, monumentBricks, monumentBricks, false)
	e.fill(5, 1, 9, 5, 3, 9, monumentBricks, monumentBricks, false)
	e.fill(3, 2, 6, 4, 2, 6, monumentBricks, monumentBricks, false)
	e.fill(3, 2, 9, 4, 2, 9, monumentBricks, monumentBricks, false)
	e.fill(2, 2, 7, 2, 2, 8, monumentBricks, monumentBricks, false)
	e.fill(5, 2, 7, 5, 2, 8, monumentBricks, monumentBricks, false)
	e.set(monumentLantern, 2, 2, 5)
	e.set(monumentLantern, 5, 2, 5)
	e.set(monumentLantern, 2, 2, 10)
	e.set(monumentLantern, 5, 2, 10)
	e.set(monumentBricks, 2, 3, 5)
	e.set(monumentBricks, 5, 3, 5)
	e.set(monumentBricks, 2, 3, 10)
	e.set(monumentBricks, 5, 3, 10)
	if v8.open[3] {
		e.waterBox(3, 1, 0, 4, 2, 0)
	}
	if v8.open[5] {
		e.waterBox(7, 1, 3, 7, 2, 4)
	}
	if v8.open[4] {
		e.waterBox(0, 1, 3, 0, 2, 4)
	}
	if v7.open[2] {
		e.waterBox(3, 1, 15, 4, 2, 15)
	}
	if v7.open[4] {
		e.waterBox(0, 1, 11, 0, 2, 12)
	}
	if v7.open[5] {
		e.waterBox(7, 1, 11, 7, 2, 12)
	}
}

func (e *monumentPieceEditor) placeDoubleX(random StructurePieceRandom) {
	v7 := e.child.room.connections[5]
	v8 := e.child.room
	if e.child.room.id/25 > 0 {
		e.floor(8, 0, v7.open[0])
		e.floor(0, 0, v8.open[0])
	}
	if v8.connections[1] == nil {
		e.waterOnlyFill(1, 4, 1, 7, 4, 6, monumentPrismarine)
	}
	if v7.connections[1] == nil {
		e.waterOnlyFill(8, 4, 1, 14, 4, 6, monumentPrismarine)
	}
	e.fill(0, 3, 0, 0, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(15, 3, 0, 15, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 0, 15, 3, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 7, 14, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(0, 2, 0, 0, 2, 7, monumentPrismarine, monumentPrismarine, false)
	e.fill(15, 2, 0, 15, 2, 7, monumentPrismarine, monumentPrismarine, false)
	e.fill(1, 2, 0, 15, 2, 0, monumentPrismarine, monumentPrismarine, false)
	e.fill(1, 2, 7, 14, 2, 7, monumentPrismarine, monumentPrismarine, false)
	e.fill(0, 1, 0, 0, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(15, 1, 0, 15, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 0, 15, 1, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 7, 14, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(5, 1, 0, 10, 1, 4, monumentBricks, monumentBricks, false)
	e.fill(6, 2, 0, 9, 2, 3, monumentPrismarine, monumentPrismarine, false)
	e.fill(5, 3, 0, 10, 3, 4, monumentBricks, monumentBricks, false)
	e.set(monumentLantern, 6, 2, 3)
	e.set(monumentLantern, 9, 2, 3)
	if v8.open[3] {
		e.waterBox(3, 1, 0, 4, 2, 0)
	}
	if v8.open[2] {
		e.waterBox(3, 1, 7, 4, 2, 7)
	}
	if v8.open[4] {
		e.waterBox(0, 1, 3, 0, 2, 4)
	}
	if v7.open[3] {
		e.waterBox(11, 1, 0, 12, 2, 0)
	}
	if v7.open[2] {
		e.waterBox(11, 1, 7, 12, 2, 7)
	}
	if v7.open[5] {
		e.waterBox(15, 1, 3, 15, 2, 4)
	}
}

func (e *monumentPieceEditor) placeDoubleY(random StructurePieceRandom) {
	if e.child.room.id/25 > 0 {
		e.floor(0, 0, e.child.room.open[0])
	}
	v7 := e.child.room.connections[1]
	if v7.connections[1] == nil {
		e.waterOnlyFill(1, 8, 1, 6, 8, 6, monumentPrismarine)
	}
	e.fill(0, 4, 0, 0, 4, 7, monumentBricks, monumentBricks, false)
	e.fill(7, 4, 0, 7, 4, 7, monumentBricks, monumentBricks, false)
	e.fill(1, 4, 0, 6, 4, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 4, 7, 6, 4, 7, monumentBricks, monumentBricks, false)
	e.fill(2, 4, 1, 2, 4, 2, monumentBricks, monumentBricks, false)
	e.fill(1, 4, 2, 1, 4, 2, monumentBricks, monumentBricks, false)
	e.fill(5, 4, 1, 5, 4, 2, monumentBricks, monumentBricks, false)
	e.fill(6, 4, 2, 6, 4, 2, monumentBricks, monumentBricks, false)
	e.fill(2, 4, 5, 2, 4, 6, monumentBricks, monumentBricks, false)
	e.fill(1, 4, 5, 1, 4, 5, monumentBricks, monumentBricks, false)
	e.fill(5, 4, 5, 5, 4, 6, monumentBricks, monumentBricks, false)
	e.fill(6, 4, 5, 6, 4, 5, monumentBricks, monumentBricks, false)
	v8 := e.child.room
	for v9 := int32(1); v9 <= 5; v9 += 4 {
		v10 := int32(0)
		if v8.open[3] {
			e.fill(2, v9, v10, 2, v9+2, v10, monumentBricks, monumentBricks, false)
			e.fill(5, v9, v10, 5, v9+2, v10, monumentBricks, monumentBricks, false)
			e.fill(3, v9+2, v10, 4, v9+2, v10, monumentBricks, monumentBricks, false)
		} else {
			e.fill(0, v9, v10, 7, v9+2, v10, monumentBricks, monumentBricks, false)
			e.fill(0, v9+1, v10, 7, v9+1, v10, monumentPrismarine, monumentPrismarine, false)
		}
		v10 = 7
		if v8.open[2] {
			e.fill(2, v9, v10, 2, v9+2, v10, monumentBricks, monumentBricks, false)
			e.fill(5, v9, v10, 5, v9+2, v10, monumentBricks, monumentBricks, false)
			e.fill(3, v9+2, v10, 4, v9+2, v10, monumentBricks, monumentBricks, false)
		} else {
			e.fill(0, v9, v10, 7, v9+2, v10, monumentBricks, monumentBricks, false)
			e.fill(0, v9+1, v10, 7, v9+1, v10, monumentPrismarine, monumentPrismarine, false)
		}
		v11 := int32(0)
		if v8.open[4] {
			e.fill(v11, v9, 2, v11, v9+2, 2, monumentBricks, monumentBricks, false)
			e.fill(v11, v9, 5, v11, v9+2, 5, monumentBricks, monumentBricks, false)
			e.fill(v11, v9+2, 3, v11, v9+2, 4, monumentBricks, monumentBricks, false)
		} else {
			e.fill(v11, v9, 0, v11, v9+2, 7, monumentBricks, monumentBricks, false)
			e.fill(v11, v9+1, 0, v11, v9+1, 7, monumentPrismarine, monumentPrismarine, false)
		}
		v11 = 7
		if v8.open[5] {
			e.fill(v11, v9, 2, v11, v9+2, 2, monumentBricks, monumentBricks, false)
			e.fill(v11, v9, 5, v11, v9+2, 5, monumentBricks, monumentBricks, false)
			e.fill(v11, v9+2, 3, v11, v9+2, 4, monumentBricks, monumentBricks, false)
		} else {
			e.fill(v11, v9, 0, v11, v9+2, 7, monumentBricks, monumentBricks, false)
			e.fill(v11, v9+1, 0, v11, v9+1, 7, monumentPrismarine, monumentPrismarine, false)
		}
		v8 = v7
	}
}

func (e *monumentPieceEditor) placeSimpleTop(random StructurePieceRandom) {
	if e.child.room.id/25 > 0 {
		e.floor(0, 0, e.child.room.open[0])
	}
	if e.child.room.connections[1] == nil {
		e.waterOnlyFill(1, 4, 1, 6, 4, 6, monumentPrismarine)
	}
	for v7 := int32(1); v7 <= 6; v7++ {
		for v8 := int32(1); v8 <= 6; v8++ {
			if random.NextIntBound(3) == 0 {
				continue
			}
			v9 := int32(2)
			if random.NextIntBound(4) != 0 {
				v9++
			}
			v10 := structureState("wet_sponge")
			e.fill(v7, v9, v8, v7, 3, v8, v10, v10, false)
		}
	}
	e.fill(0, 1, 0, 0, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(7, 1, 0, 7, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 0, 6, 1, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 7, 6, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(0, 2, 0, 0, 2, 7, monumentDark, monumentDark, false)
	e.fill(7, 2, 0, 7, 2, 7, monumentDark, monumentDark, false)
	e.fill(1, 2, 0, 6, 2, 0, monumentDark, monumentDark, false)
	e.fill(1, 2, 7, 6, 2, 7, monumentDark, monumentDark, false)
	e.fill(0, 3, 0, 0, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(7, 3, 0, 7, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 0, 6, 3, 0, monumentBricks, monumentBricks, false)
	e.fill(1, 3, 7, 6, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(0, 1, 3, 0, 2, 4, monumentDark, monumentDark, false)
	e.fill(7, 1, 3, 7, 2, 4, monumentDark, monumentDark, false)
	e.fill(3, 1, 0, 4, 2, 0, monumentDark, monumentDark, false)
	e.fill(3, 1, 7, 4, 2, 7, monumentDark, monumentDark, false)
	if e.child.room.open[3] {
		e.waterBox(3, 1, 0, 4, 2, 0)
	}
}

func (e *monumentPieceEditor) placeSimple(random StructurePieceRandom) {
	if e.child.room.id/25 > 0 {
		e.floor(0, 0, e.child.room.open[0])
	}
	if e.child.room.connections[1] == nil {
		e.waterOnlyFill(1, 4, 1, 6, 4, 6, monumentPrismarine)
	}
	v7 := e.child.variant != 0 && random.NextBoolean() && !e.child.room.open[0] && !e.child.room.open[1] && e.child.room.openingCount() > 1
	if e.child.variant == 0 {
		e.fill(0, 1, 0, 2, 1, 2, monumentBricks, monumentBricks, false)
		e.fill(0, 3, 0, 2, 3, 2, monumentBricks, monumentBricks, false)
		e.fill(0, 2, 0, 0, 2, 2, monumentPrismarine, monumentPrismarine, false)
		e.fill(1, 2, 0, 2, 2, 0, monumentPrismarine, monumentPrismarine, false)
		e.set(monumentLantern, 1, 2, 1)
		e.fill(5, 1, 0, 7, 1, 2, monumentBricks, monumentBricks, false)
		e.fill(5, 3, 0, 7, 3, 2, monumentBricks, monumentBricks, false)
		e.fill(7, 2, 0, 7, 2, 2, monumentPrismarine, monumentPrismarine, false)
		e.fill(5, 2, 0, 6, 2, 0, monumentPrismarine, monumentPrismarine, false)
		e.set(monumentLantern, 6, 2, 1)
		e.fill(0, 1, 5, 2, 1, 7, monumentBricks, monumentBricks, false)
		e.fill(0, 3, 5, 2, 3, 7, monumentBricks, monumentBricks, false)
		e.fill(0, 2, 5, 0, 2, 7, monumentPrismarine, monumentPrismarine, false)
		e.fill(1, 2, 7, 2, 2, 7, monumentPrismarine, monumentPrismarine, false)
		e.set(monumentLantern, 1, 2, 6)
		e.fill(5, 1, 5, 7, 1, 7, monumentBricks, monumentBricks, false)
		e.fill(5, 3, 5, 7, 3, 7, monumentBricks, monumentBricks, false)
		e.fill(7, 2, 5, 7, 2, 7, monumentPrismarine, monumentPrismarine, false)
		e.fill(5, 2, 7, 6, 2, 7, monumentPrismarine, monumentPrismarine, false)
		e.set(monumentLantern, 6, 2, 6)
		if e.child.room.open[3] {
			e.fill(3, 3, 0, 4, 3, 0, monumentBricks, monumentBricks, false)
		} else {
			e.fill(3, 3, 0, 4, 3, 1, monumentBricks, monumentBricks, false)
			e.fill(3, 2, 0, 4, 2, 0, monumentPrismarine, monumentPrismarine, false)
			e.fill(3, 1, 0, 4, 1, 1, monumentBricks, monumentBricks, false)
		}
		if e.child.room.open[2] {
			e.fill(3, 3, 7, 4, 3, 7, monumentBricks, monumentBricks, false)
		} else {
			e.fill(3, 3, 6, 4, 3, 7, monumentBricks, monumentBricks, false)
			e.fill(3, 2, 7, 4, 2, 7, monumentPrismarine, monumentPrismarine, false)
			e.fill(3, 1, 6, 4, 1, 7, monumentBricks, monumentBricks, false)
		}
		if e.child.room.open[4] {
			e.fill(0, 3, 3, 0, 3, 4, monumentBricks, monumentBricks, false)
		} else {
			e.fill(0, 3, 3, 1, 3, 4, monumentBricks, monumentBricks, false)
			e.fill(0, 2, 3, 0, 2, 4, monumentPrismarine, monumentPrismarine, false)
			e.fill(0, 1, 3, 1, 1, 4, monumentBricks, monumentBricks, false)
		}
		if e.child.room.open[5] {
			e.fill(7, 3, 3, 7, 3, 4, monumentBricks, monumentBricks, false)
		} else {
			e.fill(6, 3, 3, 7, 3, 4, monumentBricks, monumentBricks, false)
			e.fill(7, 2, 3, 7, 2, 4, monumentPrismarine, monumentPrismarine, false)
			e.fill(6, 1, 3, 7, 1, 4, monumentBricks, monumentBricks, false)
		}
	} else if e.child.variant == 1 {
		e.fill(2, 1, 2, 2, 3, 2, monumentBricks, monumentBricks, false)
		e.fill(2, 1, 5, 2, 3, 5, monumentBricks, monumentBricks, false)
		e.fill(5, 1, 5, 5, 3, 5, monumentBricks, monumentBricks, false)
		e.fill(5, 1, 2, 5, 3, 2, monumentBricks, monumentBricks, false)
		e.set(monumentLantern, 2, 2, 2)
		e.set(monumentLantern, 2, 2, 5)
		e.set(monumentLantern, 5, 2, 5)
		e.set(monumentLantern, 5, 2, 2)
		e.fill(0, 1, 0, 1, 3, 0, monumentBricks, monumentBricks, false)
		e.fill(0, 1, 1, 0, 3, 1, monumentBricks, monumentBricks, false)
		e.fill(0, 1, 7, 1, 3, 7, monumentBricks, monumentBricks, false)
		e.fill(0, 1, 6, 0, 3, 6, monumentBricks, monumentBricks, false)
		e.fill(6, 1, 7, 7, 3, 7, monumentBricks, monumentBricks, false)
		e.fill(7, 1, 6, 7, 3, 6, monumentBricks, monumentBricks, false)
		e.fill(6, 1, 0, 7, 3, 0, monumentBricks, monumentBricks, false)
		e.fill(7, 1, 1, 7, 3, 1, monumentBricks, monumentBricks, false)
		e.set(monumentPrismarine, 1, 2, 0)
		e.set(monumentPrismarine, 0, 2, 1)
		e.set(monumentPrismarine, 1, 2, 7)
		e.set(monumentPrismarine, 0, 2, 6)
		e.set(monumentPrismarine, 6, 2, 7)
		e.set(monumentPrismarine, 7, 2, 6)
		e.set(monumentPrismarine, 6, 2, 0)
		e.set(monumentPrismarine, 7, 2, 1)
		if !e.child.room.open[3] {
			e.fill(1, 3, 0, 6, 3, 0, monumentBricks, monumentBricks, false)
			e.fill(1, 2, 0, 6, 2, 0, monumentPrismarine, monumentPrismarine, false)
			e.fill(1, 1, 0, 6, 1, 0, monumentBricks, monumentBricks, false)
		}
		if !e.child.room.open[2] {
			e.fill(1, 3, 7, 6, 3, 7, monumentBricks, monumentBricks, false)
			e.fill(1, 2, 7, 6, 2, 7, monumentPrismarine, monumentPrismarine, false)
			e.fill(1, 1, 7, 6, 1, 7, monumentBricks, monumentBricks, false)
		}
		if !e.child.room.open[4] {
			e.fill(0, 3, 1, 0, 3, 6, monumentBricks, monumentBricks, false)
			e.fill(0, 2, 1, 0, 2, 6, monumentPrismarine, monumentPrismarine, false)
			e.fill(0, 1, 1, 0, 1, 6, monumentBricks, monumentBricks, false)
		}
		if !e.child.room.open[5] {
			e.fill(7, 3, 1, 7, 3, 6, monumentBricks, monumentBricks, false)
			e.fill(7, 2, 1, 7, 2, 6, monumentPrismarine, monumentPrismarine, false)
			e.fill(7, 1, 1, 7, 1, 6, monumentBricks, monumentBricks, false)
		}
	} else if e.child.variant == 2 {
		e.fill(0, 1, 0, 0, 1, 7, monumentBricks, monumentBricks, false)
		e.fill(7, 1, 0, 7, 1, 7, monumentBricks, monumentBricks, false)
		e.fill(1, 1, 0, 6, 1, 0, monumentBricks, monumentBricks, false)
		e.fill(1, 1, 7, 6, 1, 7, monumentBricks, monumentBricks, false)
		e.fill(0, 2, 0, 0, 2, 7, monumentDark, monumentDark, false)
		e.fill(7, 2, 0, 7, 2, 7, monumentDark, monumentDark, false)
		e.fill(1, 2, 0, 6, 2, 0, monumentDark, monumentDark, false)
		e.fill(1, 2, 7, 6, 2, 7, monumentDark, monumentDark, false)
		e.fill(0, 3, 0, 0, 3, 7, monumentBricks, monumentBricks, false)
		e.fill(7, 3, 0, 7, 3, 7, monumentBricks, monumentBricks, false)
		e.fill(1, 3, 0, 6, 3, 0, monumentBricks, monumentBricks, false)
		e.fill(1, 3, 7, 6, 3, 7, monumentBricks, monumentBricks, false)
		e.fill(0, 1, 3, 0, 2, 4, monumentDark, monumentDark, false)
		e.fill(7, 1, 3, 7, 2, 4, monumentDark, monumentDark, false)
		e.fill(3, 1, 0, 4, 2, 0, monumentDark, monumentDark, false)
		e.fill(3, 1, 7, 4, 2, 7, monumentDark, monumentDark, false)
		if e.child.room.open[3] {
			e.waterBox(3, 1, 0, 4, 2, 0)
		}
		if e.child.room.open[2] {
			e.waterBox(3, 1, 7, 4, 2, 7)
		}
		if e.child.room.open[4] {
			e.waterBox(0, 1, 3, 0, 2, 4)
		}
		if e.child.room.open[5] {
			e.waterBox(7, 1, 3, 7, 2, 4)
		}
	}
	if v7 {
		e.fill(3, 1, 3, 4, 1, 4, monumentBricks, monumentBricks, false)
		e.fill(3, 2, 3, 4, 2, 4, monumentPrismarine, monumentPrismarine, false)
		e.fill(3, 3, 3, 4, 3, 4, monumentBricks, monumentBricks, false)
	}
}

func (e *monumentPieceEditor) placeEntry(random StructurePieceRandom) {
	e.fill(0, 3, 0, 2, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(5, 3, 0, 7, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(0, 2, 0, 1, 2, 7, monumentBricks, monumentBricks, false)
	e.fill(6, 2, 0, 7, 2, 7, monumentBricks, monumentBricks, false)
	e.fill(0, 1, 0, 0, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(7, 1, 0, 7, 1, 7, monumentBricks, monumentBricks, false)
	e.fill(0, 1, 7, 7, 3, 7, monumentBricks, monumentBricks, false)
	e.fill(1, 1, 0, 2, 3, 0, monumentBricks, monumentBricks, false)
	e.fill(5, 1, 0, 6, 3, 0, monumentBricks, monumentBricks, false)
	if e.child.room.open[2] {
		e.waterBox(3, 1, 7, 4, 2, 7)
	}
	if e.child.room.open[4] {
		e.waterBox(0, 1, 3, 1, 2, 4)
	}
	if e.child.room.open[5] {
		e.waterBox(6, 1, 3, 7, 2, 4)
	}
}
