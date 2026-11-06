package worldgen

// Monument exterior geometry; local coordinates are transformed by the saved
// building orientation before testing the current chunk clip.

func (e *monumentPieceEditor) placeWingShell(v0 bool, v1 int32) {
	if e.intersects(v1, 0, v1+23, 20) {
		e.fill(v1+0, 0, 0, v1+24, 0, 20, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(v1+0, 1, 0, v1+24, 10, 20)
		for v6 := int32(0); v6 < 4; v6++ {
			e.fill(v1+v6, v6+1, v6, v1+v6, v6+1, 20, monumentBricks, monumentBricks, false)
			e.fill(v1+v6+7, v6+5, v6+7, v1+v6+7, v6+5, 20, monumentBricks, monumentBricks, false)
			e.fill(v1+17-v6, v6+5, v6+7, v1+17-v6, v6+5, 20, monumentBricks, monumentBricks, false)
			e.fill(v1+24-v6, v6+1, v6, v1+24-v6, v6+1, 20, monumentBricks, monumentBricks, false)
			e.fill(v1+v6+1, v6+1, v6, v1+23-v6, v6+1, v6, monumentBricks, monumentBricks, false)
			e.fill(v1+v6+8, v6+5, v6+7, v1+16-v6, v6+5, v6+7, monumentBricks, monumentBricks, false)
		}
		e.fill(v1+4, 4, 4, v1+6, 4, 20, monumentPrismarine, monumentPrismarine, false)
		e.fill(v1+7, 4, 4, v1+17, 4, 6, monumentPrismarine, monumentPrismarine, false)
		e.fill(v1+18, 4, 4, v1+20, 4, 20, monumentPrismarine, monumentPrismarine, false)
		e.fill(v1+11, 8, 11, v1+13, 8, 20, monumentPrismarine, monumentPrismarine, false)
		e.set(monumentBricks, v1+12, 9, 12)
		e.set(monumentBricks, v1+12, 9, 15)
		e.set(monumentBricks, v1+12, 9, 18)
		v7 := v1 + 5
		if v0 {
			v7 = v1 + 19
		}
		v8 := v1 + 19
		if v0 {
			v8 = v1 + 5
		}
		for v9 := int32(20); v9 >= 5; v9 -= 3 {
			e.set(monumentBricks, v7, 5, v9)
		}
		for v10 := int32(19); v10 >= 7; v10 -= 3 {
			e.set(monumentBricks, v8, 5, v10)
		}
		for v11 := int32(0); v11 < 4; v11++ {
			v12 := v1 + 17 - v11*3
			if v0 {
				v12 = v1 + 24 - (17 - v11*3)
			}
			e.set(monumentBricks, v12, 5, 5)
		}
		e.set(monumentBricks, v8, 5, 5)
		e.fill(v1+11, 1, 12, v1+13, 7, 12, monumentPrismarine, monumentPrismarine, false)
		e.fill(v1+12, 1, 11, v1+12, 7, 13, monumentPrismarine, monumentPrismarine, false)
	}
}

func (e *monumentPieceEditor) placeEntranceArches() {
	if e.intersects(22, 5, 35, 17) {
		e.waterBox(25, 0, 0, 32, 8, 20)
		for v3 := int32(0); v3 < 4; v3++ {
			e.fill(24, 2, 5+v3*4, 24, 4, 5+v3*4, monumentBricks, monumentBricks, false)
			e.fill(22, 4, 5+v3*4, 23, 4, 5+v3*4, monumentBricks, monumentBricks, false)
			e.set(monumentBricks, 25, 5, 5+v3*4)
			e.set(monumentBricks, 26, 6, 5+v3*4)
			e.set(monumentLantern, 26, 5, 5+v3*4)
			e.fill(33, 2, 5+v3*4, 33, 4, 5+v3*4, monumentBricks, monumentBricks, false)
			e.fill(34, 4, 5+v3*4, 35, 4, 5+v3*4, monumentBricks, monumentBricks, false)
			e.set(monumentBricks, 32, 5, 5+v3*4)
			e.set(monumentBricks, 31, 6, 5+v3*4)
			e.set(monumentLantern, 31, 5, 5+v3*4)
			e.fill(27, 6, 5+v3*4, 30, 6, 5+v3*4, monumentPrismarine, monumentPrismarine, false)
		}
	}
}

func (e *monumentPieceEditor) placeEntranceWall() {
	if e.intersects(15, 20, 42, 21) {
		e.fill(15, 0, 21, 42, 0, 21, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(26, 1, 21, 31, 3, 21)
		e.fill(21, 12, 21, 36, 12, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(17, 11, 21, 40, 11, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(16, 10, 21, 41, 10, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(15, 7, 21, 42, 9, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(16, 6, 21, 41, 6, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(17, 5, 21, 40, 5, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(21, 4, 21, 36, 4, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(22, 3, 21, 26, 3, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(31, 3, 21, 35, 3, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(23, 2, 21, 25, 2, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(32, 2, 21, 34, 2, 21, monumentPrismarine, monumentPrismarine, false)
		e.fill(28, 4, 20, 29, 4, 21, monumentBricks, monumentBricks, false)
		e.set(monumentBricks, 27, 3, 21)
		e.set(monumentBricks, 30, 3, 21)
		e.set(monumentBricks, 26, 2, 21)
		e.set(monumentBricks, 31, 2, 21)
		e.set(monumentBricks, 25, 1, 21)
		e.set(monumentBricks, 32, 1, 21)
		for v3 := int32(0); v3 < 7; v3++ {
			e.set(monumentDark, 28-v3, 6+v3, 21)
			e.set(monumentDark, 29+v3, 6+v3, 21)
		}
		for v4 := int32(0); v4 < 4; v4++ {
			e.set(monumentDark, 28-v4, 9+v4, 21)
			e.set(monumentDark, 29+v4, 9+v4, 21)
		}
		e.set(monumentDark, 28, 12, 21)
		e.set(monumentDark, 29, 12, 21)
		for v5 := int32(0); v5 < 3; v5++ {
			e.set(monumentDark, 22-v5*2, 8, 21)
			e.set(monumentDark, 22-v5*2, 9, 21)
			e.set(monumentDark, 35+v5*2, 8, 21)
			e.set(monumentDark, 35+v5*2, 9, 21)
		}
		e.waterBox(15, 13, 21, 42, 15, 21)
		e.waterBox(15, 1, 21, 15, 6, 21)
		e.waterBox(16, 1, 21, 16, 5, 21)
		e.waterBox(17, 1, 21, 20, 4, 21)
		e.waterBox(21, 1, 21, 21, 3, 21)
		e.waterBox(22, 1, 21, 22, 2, 21)
		e.waterBox(23, 1, 21, 24, 1, 21)
		e.waterBox(42, 1, 21, 42, 6, 21)
		e.waterBox(41, 1, 21, 41, 5, 21)
		e.waterBox(37, 1, 21, 40, 4, 21)
		e.waterBox(36, 1, 21, 36, 3, 21)
		e.waterBox(33, 1, 21, 34, 1, 21)
		e.waterBox(35, 1, 21, 35, 2, 21)
	}
}

func (e *monumentPieceEditor) placeRoof() {
	if e.intersects(21, 21, 36, 36) {
		e.fill(21, 0, 22, 36, 0, 36, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(21, 1, 22, 36, 23, 36)
		for v3 := int32(0); v3 < 4; v3++ {
			e.fill(21+v3, 13+v3, 21+v3, 36-v3, 13+v3, 21+v3, monumentBricks, monumentBricks, false)
			e.fill(21+v3, 13+v3, 36-v3, 36-v3, 13+v3, 36-v3, monumentBricks, monumentBricks, false)
			e.fill(21+v3, 13+v3, 22+v3, 21+v3, 13+v3, 35-v3, monumentBricks, monumentBricks, false)
			e.fill(36-v3, 13+v3, 22+v3, 36-v3, 13+v3, 35-v3, monumentBricks, monumentBricks, false)
		}
		e.fill(25, 16, 25, 32, 16, 32, monumentPrismarine, monumentPrismarine, false)
		e.fill(25, 17, 25, 25, 19, 25, monumentBricks, monumentBricks, false)
		e.fill(32, 17, 25, 32, 19, 25, monumentBricks, monumentBricks, false)
		e.fill(25, 17, 32, 25, 19, 32, monumentBricks, monumentBricks, false)
		e.fill(32, 17, 32, 32, 19, 32, monumentBricks, monumentBricks, false)
		e.set(monumentBricks, 26, 20, 26)
		e.set(monumentBricks, 27, 21, 27)
		e.set(monumentLantern, 27, 20, 27)
		e.set(monumentBricks, 26, 20, 31)
		e.set(monumentBricks, 27, 21, 30)
		e.set(monumentLantern, 27, 20, 30)
		e.set(monumentBricks, 31, 20, 31)
		e.set(monumentBricks, 30, 21, 30)
		e.set(monumentLantern, 30, 20, 30)
		e.set(monumentBricks, 31, 20, 26)
		e.set(monumentBricks, 30, 21, 27)
		e.set(monumentLantern, 30, 20, 27)
		e.fill(28, 21, 27, 29, 21, 27, monumentPrismarine, monumentPrismarine, false)
		e.fill(27, 21, 28, 27, 21, 29, monumentPrismarine, monumentPrismarine, false)
		e.fill(28, 21, 30, 29, 21, 30, monumentPrismarine, monumentPrismarine, false)
		e.fill(30, 21, 28, 30, 21, 29, monumentPrismarine, monumentPrismarine, false)
	}
}

func (e *monumentPieceEditor) placeLowerWall() {
	if e.intersects(0, 21, 6, 58) {
		e.fill(0, 0, 21, 6, 0, 57, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(0, 1, 21, 6, 7, 57)
		e.fill(4, 4, 21, 6, 4, 53, monumentPrismarine, monumentPrismarine, false)
		for v3 := int32(0); v3 < 4; v3++ {
			e.fill(v3, v3+1, 21, v3, v3+1, 57-v3, monumentBricks, monumentBricks, false)
		}
		for v4 := int32(23); v4 < 53; v4 += 3 {
			e.set(monumentBricks, 5, 5, v4)
		}
		e.set(monumentBricks, 5, 5, 52)
		for v5 := int32(0); v5 < 4; v5++ {
			e.fill(v5, v5+1, 21, v5, v5+1, 57-v5, monumentBricks, monumentBricks, false)
		}
		e.fill(4, 1, 52, 6, 3, 52, monumentPrismarine, monumentPrismarine, false)
		e.fill(5, 1, 51, 5, 3, 53, monumentPrismarine, monumentPrismarine, false)
	}
	if e.intersects(51, 21, 58, 58) {
		e.fill(51, 0, 21, 57, 0, 57, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(51, 1, 21, 57, 7, 57)
		e.fill(51, 4, 21, 53, 4, 53, monumentPrismarine, monumentPrismarine, false)
		for v6 := int32(0); v6 < 4; v6++ {
			e.fill(57-v6, v6+1, 21, 57-v6, v6+1, 57-v6, monumentBricks, monumentBricks, false)
		}
		for v7 := int32(23); v7 < 53; v7 += 3 {
			e.set(monumentBricks, 52, 5, v7)
		}
		e.set(monumentBricks, 52, 5, 52)
		e.fill(51, 1, 52, 53, 3, 52, monumentPrismarine, monumentPrismarine, false)
		e.fill(52, 1, 51, 52, 3, 53, monumentPrismarine, monumentPrismarine, false)
	}
	if e.intersects(0, 51, 57, 57) {
		e.fill(7, 0, 51, 50, 0, 57, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(7, 1, 51, 50, 10, 57)
		for v8 := int32(0); v8 < 4; v8++ {
			e.fill(v8+1, v8+1, 57-v8, 56-v8, v8+1, 57-v8, monumentBricks, monumentBricks, false)
		}
	}
}

func (e *monumentPieceEditor) placeMiddleWall() {
	if e.intersects(7, 21, 13, 50) {
		e.fill(7, 0, 21, 13, 0, 50, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(7, 1, 21, 13, 10, 50)
		e.fill(11, 8, 21, 13, 8, 53, monumentPrismarine, monumentPrismarine, false)
		for v3 := int32(0); v3 < 4; v3++ {
			e.fill(v3+7, v3+5, 21, v3+7, v3+5, 54, monumentBricks, monumentBricks, false)
		}
		for v4 := int32(21); v4 <= 45; v4 += 3 {
			e.set(monumentBricks, 12, 9, v4)
		}
	}
	if e.intersects(44, 21, 50, 54) {
		e.fill(44, 0, 21, 50, 0, 50, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(44, 1, 21, 50, 10, 50)
		e.fill(44, 8, 21, 46, 8, 53, monumentPrismarine, monumentPrismarine, false)
		for v5 := int32(0); v5 < 4; v5++ {
			e.fill(50-v5, v5+5, 21, 50-v5, v5+5, 54, monumentBricks, monumentBricks, false)
		}
		for v6 := int32(21); v6 <= 45; v6 += 3 {
			e.set(monumentBricks, 45, 9, v6)
		}
	}
	if e.intersects(8, 44, 49, 54) {
		e.fill(14, 0, 44, 43, 0, 50, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(14, 1, 44, 43, 10, 50)
		for v7 := int32(12); v7 <= 45; v7 += 3 {
			e.set(monumentBricks, v7, 9, 45)
			e.set(monumentBricks, v7, 9, 52)
			if v7 != 12 && v7 != 18 && v7 != 24 && v7 != 33 && v7 != 39 && v7 != 45 {
				continue
			}
			e.set(monumentBricks, v7, 9, 47)
			e.set(monumentBricks, v7, 9, 50)
			e.set(monumentBricks, v7, 10, 45)
			e.set(monumentBricks, v7, 10, 46)
			e.set(monumentBricks, v7, 10, 51)
			e.set(monumentBricks, v7, 10, 52)
			e.set(monumentBricks, v7, 11, 47)
			e.set(monumentBricks, v7, 11, 50)
			e.set(monumentBricks, v7, 12, 48)
			e.set(monumentBricks, v7, 12, 49)
		}
		for v8 := int32(0); v8 < 3; v8++ {
			e.fill(8+v8, 5+v8, 54, 49-v8, 5+v8, 54, monumentPrismarine, monumentPrismarine, false)
		}
		e.fill(11, 8, 54, 46, 8, 54, monumentBricks, monumentBricks, false)
		e.fill(14, 8, 44, 43, 8, 53, monumentPrismarine, monumentPrismarine, false)
	}
}

func (e *monumentPieceEditor) placeUpperWall() {
	if e.intersects(14, 21, 20, 43) {
		e.fill(14, 0, 21, 20, 0, 43, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(14, 1, 22, 20, 14, 43)
		e.fill(18, 12, 22, 20, 12, 39, monumentPrismarine, monumentPrismarine, false)
		e.fill(18, 12, 21, 20, 12, 21, monumentBricks, monumentBricks, false)
		for v3 := int32(0); v3 < 4; v3++ {
			e.fill(v3+14, v3+9, 21, v3+14, v3+9, 43-v3, monumentBricks, monumentBricks, false)
		}
		for v4 := int32(23); v4 <= 39; v4 += 3 {
			e.set(monumentBricks, 19, 13, v4)
		}
	}
	if e.intersects(37, 21, 43, 43) {
		e.fill(37, 0, 21, 43, 0, 43, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(37, 1, 22, 43, 14, 43)
		e.fill(37, 12, 22, 39, 12, 39, monumentPrismarine, monumentPrismarine, false)
		e.fill(37, 12, 21, 39, 12, 21, monumentBricks, monumentBricks, false)
		for v5 := int32(0); v5 < 4; v5++ {
			e.fill(43-v5, v5+9, 21, 43-v5, v5+9, 43-v5, monumentBricks, monumentBricks, false)
		}
		for v6 := int32(23); v6 <= 39; v6 += 3 {
			e.set(monumentBricks, 38, 13, v6)
		}
	}
	if e.intersects(15, 37, 42, 43) {
		e.fill(21, 0, 37, 36, 0, 43, monumentPrismarine, monumentPrismarine, false)
		e.waterBox(21, 1, 37, 36, 14, 43)
		e.fill(21, 12, 37, 36, 12, 39, monumentPrismarine, monumentPrismarine, false)
		for v7 := int32(0); v7 < 4; v7++ {
			e.fill(15+v7, v7+9, 43-v7, 42-v7, v7+9, 43-v7, monumentBricks, monumentBricks, false)
		}
		for v8 := int32(21); v8 <= 36; v8 += 3 {
			e.set(monumentBricks, v8, 13, 38)
		}
	}
}
