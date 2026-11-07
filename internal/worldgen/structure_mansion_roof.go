package worldgen

func (a *mansionAssembler) roof(origin mansionPosition, r mansionRotation, g, above *mansionGrid) {
	covered := func(x, z int) bool { return above != nil && above.house(x, z) }
	// Preserve the original three passes: roof fronts, upper-floor parapets,
	// then corners. Insertion order affects overlapping template placement.
	for z := 0; z < 11; z++ {
		for x := 0; x < 11; x++ {
			if !g.house(x, z) || covered(x, z) {
				continue
			}
			p := mansionCell(origin, r, x, z)
			a.add("roof", p.up(3), r)
			if !g.house(x+1, z) {
				a.add("roof_front", p.move(r.direction(mansionEast), 6), r)
			}
			if !g.house(x-1, z) {
				a.add("roof_front", p.move(r.direction(mansionSouth), 7), r.turn(2))
			}
			if !g.house(x, z-1) {
				a.add("roof_front", p.move(r.direction(mansionWest), 1), r.turn(3))
			}
			if !g.house(x, z+1) {
				a.add("roof_front", p.move(r.direction(mansionEast), 6).move(r.direction(mansionSouth), 6), r.turn(1))
			}
		}
	}
	if above != nil {
		for z := 0; z < 11; z++ {
			for x := 0; x < 11; x++ {
				if !g.house(x, z) || !covered(x, z) {
					continue
				}
				p := mansionCell(origin, r, x, z)
				place := func(name string, east, south, turn int) {
					a.add(name, p.move(r.direction(mansionEast), east).move(r.direction(mansionSouth), south), r.turn(turn))
				}
				if !g.house(x+1, z) {
					place("small_wall", 7, 0, 0)
				}
				if !g.house(x-1, z) {
					place("small_wall", -1, 6, 2)
				}
				if !g.house(x, z-1) {
					place("small_wall", 0, -1, 3)
				}
				if !g.house(x, z+1) {
					place("small_wall", 6, 7, 1)
				}
				if !g.house(x+1, z) {
					if !g.house(x, z-1) {
						place("small_wall_corner", 7, -2, 0)
					}
					if !g.house(x, z+1) {
						place("small_wall_corner", 8, 7, 1)
					}
				}
				if !g.house(x-1, z) {
					if !g.house(x, z-1) {
						place("small_wall_corner", -2, -1, 3)
					}
					if !g.house(x, z+1) {
						place("small_wall_corner", -1, 8, 2)
					}
				}
			}
		}
	}
	for z := 0; z < 11; z++ {
		for x := 0; x < 11; x++ {
			if !g.house(x, z) || covered(x, z) {
				continue
			}
			p := mansionCell(origin, r, x, z)
			place := func(name string, east, south, turn int) {
				a.add(name, p.move(r.direction(mansionEast), east).move(r.direction(mansionSouth), south), r.turn(turn))
			}
			if !g.house(x+1, z) {
				if !g.house(x, z+1) {
					place("roof_corner", 6, 6, 0)
				} else if g.house(x+1, z+1) {
					place("roof_inner_corner", 6, 5, 0)
				}
				if !g.house(x, z-1) {
					place("roof_corner", 6, 0, 3)
				} else if g.house(x+1, z-1) {
					place("roof_inner_corner", 9, -2, 1)
				}
			}
			if !g.house(x-1, z) {
				if !g.house(x, z+1) {
					place("roof_corner", 0, 6, 1)
				} else if g.house(x-1, z+1) {
					place("roof_inner_corner", -3, 8, 3)
				}
				if !g.house(x, z-1) {
					place("roof_corner", 0, 0, 2)
				} else if g.house(x-1, z-1) {
					place("roof_inner_corner", 0, 1, 2)
				}
			}
		}
	}
}
