package worldgen

// The five grids retain vanilla's bit fields. Room metadata is separate from
// the footprint: combining them changes corridor traversal and random draws.
type mansionGrid [11][11]int32

func (g *mansionGrid) get(x, z int) int32 {
	if x < 0 || x >= 11 || z < 0 || z >= 11 {
		return 5
	}
	return g[x][z]
}
func (g *mansionGrid) set(x, z int, value int32) {
	if x >= 0 && x < 11 && z >= 0 && z < 11 {
		g[x][z] = value
	}
}
func (g *mansionGrid) fill(x1, z1, x2, z2 int, value int32) {
	for z := z1; z <= z2; z++ {
		for x := x1; x <= x2; x++ {
			g.set(x, z, value)
		}
	}
}
func (g *mansionGrid) replace(x, z int, old, value int32) {
	if g.get(x, z) == old {
		g.set(x, z, value)
	}
}
func (g *mansionGrid) adjacent(x, z int, value int32) bool {
	return g.get(x-1, z) == value || g.get(x+1, z) == value || g.get(x, z+1) == value || g.get(x, z-1) == value
}
func (g *mansionGrid) house(x, z int) bool { v := g.get(x, z); return v >= 1 && v <= 4 }

// Clockwise order also matches Direction.Plane.HORIZONTAL iteration. The
// random 2D direction lookup below has its distinct south/west/north/east order.
type mansionDirection int

const (
	mansionNorth mansionDirection = iota
	mansionEast
	mansionSouth
	mansionWest
	mansionUp
	mansionNoDirection
)

func (d mansionDirection) dx() int                            { return [4]int{0, 1, 0, -1}[d] }
func (d mansionDirection) dz() int                            { return [4]int{-1, 0, 1, 0}[d] }
func (d mansionDirection) clockwise() mansionDirection        { return (d + 1) & 3 }
func (d mansionDirection) counterclockwise() mansionDirection { return (d + 3) & 3 }
func (d mansionDirection) opposite() mansionDirection         { return (d + 2) & 3 }

type mansionFloorPlan struct {
	base, third mansionGrid
	rooms       [3]mansionGrid
	random      *LegacyRandom
}

func newMansionFloorPlan(random *LegacyRandom) *mansionFloorPlan {
	p := &mansionFloorPlan{random: random}
	g := &p.base
	g.fill(7, 4, 8, 5, 3)
	g.fill(6, 4, 6, 5, 2)
	g.fill(9, 2, 10, 7, 5)
	g.fill(8, 2, 8, 3, 1)
	g.fill(8, 6, 8, 7, 1)
	g.set(6, 3, 1)
	g.set(6, 6, 1)
	g.fill(0, 0, 11, 1, 5)
	g.fill(0, 9, 11, 11, 5)
	p.corridor(g, 7, 2, mansionWest, 6)
	p.corridor(g, 7, 7, mansionWest, 6)
	p.corridor(g, 5, 3, mansionWest, 3)
	p.corridor(g, 5, 6, mansionWest, 3)
	for p.cleanEdges(g) {
	}
	p.identifyRooms(g, &p.rooms[0])
	p.identifyRooms(g, &p.rooms[1])
	p.rooms[0].fill(8, 4, 8, 5, 0x800000)
	p.rooms[1].fill(8, 4, 8, 5, 0x800000)
	p.createThirdFloor()
	p.identifyRooms(&p.third, &p.rooms[2])
	return p
}

func (p *mansionFloorPlan) roomDirection(x, z, floor int, id int32) mansionDirection {
	for d := mansionNorth; d <= mansionWest; d++ {
		if p.rooms[floor].get(x+d.dx(), z+d.dz())&0xffff == id {
			return d
		}
	}
	return mansionNoDirection
}

func (p *mansionFloorPlan) corridor(g *mansionGrid, x, z int, d mansionDirection, depth int) {
	if depth <= 0 {
		return
	}
	g.set(x, z, 1)
	g.replace(x+d.dx(), z+d.dz(), 0, 1)
	for i := 0; i < 8; i++ {
		next := mansionDirection((p.random.NextIntBound(4) + 2) & 3)
		if next == d.opposite() || next == mansionEast && p.random.NextBoolean() {
			continue
		}
		nx, nz := x+d.dx(), z+d.dz()
		if g.get(nx+next.dx(), nz+next.dz()) != 0 || g.get(nx+next.dx()*2, nz+next.dz()*2) != 0 {
			continue
		}
		p.corridor(g, nx+next.dx(), nz+next.dz(), next, depth-1)
		break
	}
	cw, ccw := d.clockwise(), d.counterclockwise()
	for _, offset := range [][2]int{{cw.dx(), cw.dz()}, {ccw.dx(), ccw.dz()}, {d.dx() + cw.dx(), d.dz() + cw.dz()}, {d.dx() + ccw.dx(), d.dz() + ccw.dz()}, {d.dx() * 2, d.dz() * 2}, {cw.dx() * 2, cw.dz() * 2}, {ccw.dx() * 2, ccw.dz() * 2}} {
		g.replace(x+offset[0], z+offset[1], 0, 2)
	}
}

func (p *mansionFloorPlan) cleanEdges(g *mansionGrid) bool {
	changed := false
	for z := 0; z < 11; z++ {
		for x := 0; x < 11; x++ {
			if g.get(x, z) != 0 {
				continue
			}
			neighbors := 0
			for _, o := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if g.house(x+o[0], z+o[1]) {
					neighbors++
				}
			}
			fill := neighbors >= 3
			if neighbors == 2 {
				diagonal := 0
				for _, o := range [][2]int{{1, 1}, {-1, 1}, {1, -1}, {-1, -1}} {
					if g.house(x+o[0], z+o[1]) {
						diagonal++
					}
				}
				fill = diagonal <= 1
			}
			if fill {
				g.set(x, z, 2)
				changed = true
			}
		}
	}
	return changed
}

func (p *mansionFloorPlan) createThirdFloor() {
	var candidates [][2]int
	rooms, g := &p.rooms[1], &p.third
	for z := 0; z < 11; z++ {
		for x := 0; x < 11; x++ {
			value := rooms.get(x, z)
			if value&0xf0000 == 0x20000 && value&0x200000 != 0 {
				candidates = append(candidates, [2]int{x, z})
			}
		}
	}
	if len(candidates) == 0 {
		g.fill(0, 0, 11, 11, 5)
		return
	}
	chosen := candidates[p.random.NextIntBound(int32(len(candidates)))]
	x, z := chosen[0], chosen[1]
	value := rooms.get(x, z)
	rooms.set(x, z, value|0x400000)
	d := p.roomDirection(x, z, 1, value&0xffff)
	nx, nz := x+d.dx(), z+d.dz()
	for zz := 0; zz < 11; zz++ {
		for xx := 0; xx < 11; xx++ {
			if !p.base.house(xx, zz) {
				g.set(xx, zz, 5)
			} else if xx == x && zz == z {
				g.set(xx, zz, 3)
			} else if xx == nx && zz == nz {
				g.set(xx, zz, 3)
				p.rooms[2].set(xx, zz, 0x800000)
			}
		}
	}
	var directions []mansionDirection
	for dir := mansionNorth; dir <= mansionWest; dir++ {
		if g.get(nx+dir.dx(), nz+dir.dz()) == 0 {
			directions = append(directions, dir)
		}
	}
	if len(directions) == 0 {
		g.fill(0, 0, 11, 11, 5)
		rooms.set(x, z, value)
		return
	}
	d = directions[p.random.NextIntBound(int32(len(directions)))]
	p.corridor(g, nx+d.dx(), nz+d.dz(), d, 4)
	for p.cleanEdges(g) {
	}
}

func (p *mansionFloorPlan) identifyRooms(g, rooms *mansionGrid) {
	var candidates [][2]int
	for z := 0; z < 11; z++ {
		for x := 0; x < 11; x++ {
			if g.get(x, z) == 2 {
				candidates = append(candidates, [2]int{x, z})
			}
		}
	}
	for n := len(candidates); n > 1; n-- {
		i := int(p.random.NextIntBound(int32(n)))
		candidates[n-1], candidates[i] = candidates[i], candidates[n-1]
	}
	id := int32(10)
	for _, pos := range candidates {
		x, z := pos[0], pos[1]
		if rooms.get(x, z) != 0 {
			continue
		}
		x1, x2, z1, z2 := x, x, z, z
		size := int32(0x10000)
		free := func(dx, dz int) bool { return rooms.get(x+dx, z+dz) == 0 && g.get(x+dx, z+dz) == 2 }
		switch {
		case free(1, 0) && free(0, 1) && free(1, 1):
			x2++
			z2++
			size = 0x40000
		case free(-1, 0) && free(0, 1) && free(-1, 1):
			x1--
			z2++
			size = 0x40000
		case free(-1, 0) && free(0, -1) && free(-1, -1):
			x1--
			z1--
			size = 0x40000
		case free(1, 0):
			x2++
			size = 0x20000
		case free(0, 1):
			z2++
			size = 0x20000
		case free(-1, 0):
			x1--
			size = 0x20000
		case free(0, -1):
			z1--
			size = 0x20000
		}
		doorX, doorZ := x2, z2
		if p.random.NextBoolean() {
			doorX = x1
		}
		if p.random.NextBoolean() {
			doorZ = z1
		}
		flag := int32(0x200000)
		flip := func(value, lo, hi int) int {
			if value == lo {
				return hi
			}
			return lo
		}
		if !g.adjacent(doorX, doorZ, 1) {
			doorX = flip(doorX, x1, x2)
			doorZ = flip(doorZ, z1, z2)
			if !g.adjacent(doorX, doorZ, 1) {
				doorZ = flip(doorZ, z1, z2)
				if !g.adjacent(doorX, doorZ, 1) {
					doorX = flip(doorX, x1, x2)
					doorZ = flip(doorZ, z1, z2)
					if !g.adjacent(doorX, doorZ, 1) {
						flag = 0
						doorX = x1
						doorZ = z1
					}
				}
			}
		}
		for zz := z1; zz <= z2; zz++ {
			for xx := x1; xx <= x2; xx++ {
				value := size | id
				if xx == doorX && zz == doorZ {
					value |= 0x100000 | flag
				}
				rooms.set(xx, zz, value)
			}
		}
		id++
	}
}
