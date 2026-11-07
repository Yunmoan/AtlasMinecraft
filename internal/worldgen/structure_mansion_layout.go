package worldgen

import "fmt"

type mansionPosition FeaturePosition
type mansionRotation TemplateRotation

func (p mansionPosition) move(d mansionDirection, n int) mansionPosition {
	p.X += int32(d.dx() * n)
	p.Z += int32(d.dz() * n)
	return p
}
func (p mansionPosition) up(n int) mansionPosition   { p.Y += int32(n); return p }
func (r mansionRotation) turn(n int) mansionRotation { return (r + mansionRotation(n)) & 3 }
func (r mansionRotation) direction(d mansionDirection) mansionDirection {
	return (d + mansionDirection(r)) & 3
}

type mansionWallCursor struct {
	position mansionPosition
	rotation mansionRotation
	wall     string
}
type mansionAssembler struct {
	random *LegacyRandom
	pieces []StructurePiece
	err    error
}

func (a *mansionAssembler) add(name string, p mansionPosition, r mansionRotation, mirrors ...TemplateMirror) {
	if a.err != nil {
		return
	}
	mirror := TemplateMirrorNone
	if len(mirrors) > 0 {
		mirror = mirrors[0]
	}
	piece, err := NewMansionTemplatePiece(name, FeaturePosition(p), TemplateRotation(r), mirror)
	if err != nil {
		a.err = err
		return
	}
	a.pieces = append(a.pieces, *piece)
}

func assembleMansion(random *LegacyRandom, origin FeaturePosition, rotation TemplateRotation) ([]StructurePiece, error) {
	plan := newMansionFloorPlan(random)
	a := mansionAssembler{random: random}
	a.create(mansionPosition(origin), mansionRotation(rotation), plan)
	return a.pieces, a.err
}

func (a *mansionAssembler) wall(c *mansionWallCursor) {
	a.add(c.wall, c.position.move(c.rotation.direction(mansionEast), 7), c.rotation)
	c.position = c.position.move(c.rotation.direction(mansionSouth), 8)
}
func (a *mansionAssembler) outerTurn(c *mansionWallCursor) {
	c.position = c.position.move(c.rotation.direction(mansionSouth), -1)
	a.add("wall_corner", c.position, c.rotation)
	c.position = c.position.move(c.rotation.direction(mansionSouth), -7).move(c.rotation.direction(mansionWest), -6)
	c.rotation = c.rotation.turn(1)
}
func (a *mansionAssembler) innerTurn(c *mansionWallCursor) {
	c.position = c.position.move(c.rotation.direction(mansionSouth), 6).move(c.rotation.direction(mansionEast), 8)
	c.rotation = c.rotation.turn(3)
}
func (a *mansionAssembler) traverse(c *mansionWallCursor, g *mansionGrid, d mansionDirection, x, z, endX, endZ int) {
	startDirection := d
	for {
		if !g.house(x+d.dx(), z+d.dz()) {
			a.outerTurn(c)
			d = d.clockwise()
			if x != endX || z != endZ || startDirection != d {
				a.wall(c)
			}
		} else if g.house(x+d.dx()+d.counterclockwise().dx(), z+d.dz()+d.counterclockwise().dz()) {
			a.innerTurn(c)
			x += d.dx()
			z += d.dz()
			d = d.counterclockwise()
		} else {
			x += d.dx()
			z += d.dz()
			if x != endX || z != endZ || startDirection != d {
				a.wall(c)
			}
		}
		if x == endX && z == endZ && startDirection == d {
			return
		}
	}
}

func mansionCell(origin mansionPosition, r mansionRotation, x, z int) mansionPosition {
	return origin.move(r.direction(mansionSouth), 8+(z-5)*8).move(r.direction(mansionEast), (x-8)*8)
}

func (a *mansionAssembler) create(origin mansionPosition, r mansionRotation, plan *mansionFloorPlan) {
	first := mansionWallCursor{origin, r, "wall_flat"}
	a.add("entrance", first.position.move(r.direction(mansionWest), 9), r)
	first.position = first.position.move(r.direction(mansionSouth), 16)
	second := mansionWallCursor{first.position.up(8), first.rotation, "wall_window"}
	a.traverse(&first, &plan.base, mansionSouth, 8, 5, 8, 4)
	a.traverse(&second, &plan.base, mansionSouth, 8, 5, 8, 4)
	third := mansionWallCursor{first.position.up(19), first.rotation, "wall_window"}
	found := false
	for z := 0; z < 11 && !found; z++ {
		for x := 10; x >= 0 && !found; x-- {
			if !plan.third.house(x, z) {
				continue
			}
			third.position = mansionCell(third.position, r, x, z)
			a.wall(&third)
			a.traverse(&third, &plan.third, mansionSouth, x, z, x, z)
			found = true
		}
	}
	a.roof(origin.up(16), r, &plan.base, &plan.third)
	a.roof(origin.up(27), r, &plan.third, nil)
	for floor := 0; floor < 3; floor++ {
		height := floor * 8
		if floor == 2 {
			height += 3
		}
		base := origin.up(height)
		rooms := &plan.rooms[floor]
		g := &plan.base
		if floor == 2 {
			g = &plan.third
		}
		carpetSouth, carpetWest := "carpet_south_1", "carpet_west_1"
		wall, door := "indoors_wall_1", "indoors_door_1"
		if floor != 0 {
			carpetSouth, carpetWest = "carpet_south_2", "carpet_west_2"
			wall, door = "indoors_wall_2", "indoors_door_2"
		}
		for z := 0; z < 11; z++ {
			for x := 0; x < 11; x++ {
				if g.get(x, z) != 1 {
					continue
				}
				p := mansionCell(base, r, x, z)
				a.add("corridor_floor", p, r)
				if g.get(x, z-1) == 1 || rooms.get(x, z-1)&0x800000 != 0 {
					a.add("carpet_north", p.move(r.direction(mansionEast), 1).up(1), r)
				}
				if g.get(x+1, z) == 1 || rooms.get(x+1, z)&0x800000 != 0 {
					a.add("carpet_east", p.move(r.direction(mansionSouth), 1).move(r.direction(mansionEast), 5).up(1), r)
				}
				if g.get(x, z+1) == 1 || rooms.get(x, z+1)&0x800000 != 0 {
					a.add(carpetSouth, p.move(r.direction(mansionSouth), 5).move(r.direction(mansionWest), 1), r)
				}
				if g.get(x-1, z) == 1 || rooms.get(x-1, z)&0x800000 != 0 {
					a.add(carpetWest, p.move(r.direction(mansionWest), 1).move(r.direction(mansionNorth), 1), r)
				}
			}
		}
		for z := 0; z < 11; z++ {
			for x := 0; x < 11; x++ {
				stairs := floor == 2 && g.get(x, z) == 3
				if g.get(x, z) != 2 && !stairs {
					continue
				}
				value := rooms.get(x, z)
				size, id := value&0xf0000, value&0xffff
				stairs = stairs && value&0x800000 != 0
				var directions []mansionDirection
				if value&0x200000 != 0 {
					for d := mansionNorth; d <= mansionWest; d++ {
						if g.get(x+d.dx(), z+d.dz()) == 1 {
							directions = append(directions, d)
						}
					}
				}
				doorDirection := mansionNoDirection
				if len(directions) > 0 {
					doorDirection = directions[a.random.NextIntBound(int32(len(directions)))]
				} else if value&0x100000 != 0 {
					doorDirection = mansionUp
				}
				p := mansionCell(base, r, x, z).move(r.direction(mansionEast), -1)
				wallName := func(d mansionDirection) string {
					if doorDirection == d {
						return door
					}
					return wall
				}
				sameRoom := func(xx, zz int) bool { return rooms.get(xx, zz)&0xffff == id }
				if g.house(x-1, z) && !sameRoom(x-1, z) {
					a.add(wallName(mansionWest), p, r)
				}
				if g.get(x+1, z) == 1 && !stairs {
					a.add(wallName(mansionEast), p.move(r.direction(mansionEast), 8), r)
				}
				if g.house(x, z+1) && !sameRoom(x, z+1) {
					a.add(wallName(mansionSouth), p.move(r.direction(mansionSouth), 7).move(r.direction(mansionEast), 7), r.turn(1))
				}
				if g.get(x, z-1) == 1 && !stairs {
					a.add(wallName(mansionNorth), p.move(r.direction(mansionNorth), 1).move(r.direction(mansionEast), 7), r.turn(1))
				}
				switch {
				case size == 0x10000:
					a.room1x1(p, r, doorDirection, floor)
				case size == 0x20000 && doorDirection != mansionNoDirection:
					a.room1x2(p, r, plan.roomDirection(x, z, floor, id), doorDirection, floor, value&0x400000 != 0)
				case size == 0x40000 && doorDirection != mansionNoDirection && doorDirection != mansionUp:
					d := doorDirection.clockwise()
					if !sameRoom(x+d.dx(), z+d.dz()) {
						d = d.opposite()
					}
					a.room2x2(p, r, d, doorDirection, floor)
				case size == 0x40000 && doorDirection == mansionUp:
					a.add("2x2_s1", p.move(r.direction(mansionEast), 1), r)
				}
			}
		}
	}
}

func (a *mansionAssembler) roomName(kind string, floor int, stairs bool) string {
	prefix, count := "", int32(0)
	switch kind {
	case "1x1":
		prefix, count = "1x1_a", 5
		if floor > 0 {
			prefix, count = "1x1_b", 4
		}
	case "1x1_secret":
		prefix, count = "1x1_as", 4
	case "1x2_side":
		prefix, count = "1x2_a", 9
		if floor > 0 {
			if stairs {
				return "1x2_c_stairs"
			}
			prefix, count = "1x2_c", 4
		}
	case "1x2_front":
		prefix, count = "1x2_b", 5
		if floor > 0 {
			if stairs {
				return "1x2_d_stairs"
			}
			prefix = "1x2_d"
		}
	case "1x2_secret":
		prefix, count = "1x2_s", 2
		if floor > 0 {
			prefix, count = "1x2_se", 1
		}
	case "2x2":
		prefix, count = "2x2_a", 4
		if floor > 0 {
			prefix, count = "2x2_b", 5
		}
	default:
		panic("unknown mansion room family")
	}
	return fmt.Sprintf("%s%d", prefix, a.random.NextIntBound(count)+1)
}

func (a *mansionAssembler) room1x1(p mansionPosition, r mansionRotation, door mansionDirection, floor int) {
	local := mansionRotation(0)
	name := a.roomName("1x1", floor, false)
	switch door {
	case mansionEast:
	case mansionNorth:
		local = 3
	case mansionWest:
		local = 2
	case mansionSouth:
		local = 1
	default:
		name = a.roomName("1x1_secret", floor, false)
	}
	// StructureTemplate.getZeroPositionWithTransform(1,0,0,NONE,local,7,7).
	// Bytecode decrements both dimensions before its mirror branches; some
	// decompilers incorrectly move those decrements into the ternaries.
	x, z := int32(1), int32(0)
	switch local {
	case 1:
		x, z = 7, 0
	case 2:
		x, z = 7, 6
	case 3:
		x, z = 1, 6
	}
	offset, _ := TransformTemplatePosition(FeaturePosition{x, 0, z}, FeaturePosition{}, TemplateMirrorNone, TemplateRotation(r))
	p.X += offset.X
	p.Z += offset.Z
	a.add(name, p, r.turn(int(local)))
}

func (a *mansionAssembler) room1x2(p mansionPosition, r mansionRotation, extension, door mansionDirection, floor int, stairs bool) {
	x, z, turn := 0, 0, 0
	mirror := TemplateMirrorNone
	kind := "1x2_side"
	switch {
	case door == mansionEast && extension == mansionSouth:
		x = 1
	case door == mansionEast && extension == mansionNorth:
		x, z, mirror = 1, 6, TemplateMirrorLeftRight
	case door == mansionWest && extension == mansionNorth:
		x, z, turn = 7, 6, 2
	case door == mansionWest && extension == mansionSouth:
		x, mirror = 7, TemplateMirrorFrontBack
	case door == mansionSouth && extension == mansionEast:
		x, turn, mirror = 1, 1, TemplateMirrorLeftRight
	case door == mansionSouth && extension == mansionWest:
		x, turn = 7, 1
	case door == mansionNorth && extension == mansionWest:
		x, z, turn, mirror = 7, 6, 1, TemplateMirrorFrontBack
	case door == mansionNorth && extension == mansionEast:
		x, z, turn = 1, 6, 3
	case door == mansionSouth && extension == mansionNorth:
		x, z, kind = 1, -8, "1x2_front"
	case door == mansionNorth && extension == mansionSouth:
		x, z, turn, kind = 7, 14, 2, "1x2_front"
	case door == mansionWest && extension == mansionEast:
		x, turn, kind = 15, 1, "1x2_front"
	case door == mansionEast && extension == mansionWest:
		x, z, turn, kind = -7, 6, 3, "1x2_front"
	case door == mansionUp && extension == mansionEast:
		x, turn, kind = 15, 1, "1x2_secret"
	case door == mansionUp && extension == mansionSouth:
		x, kind = 1, "1x2_secret"
	default:
		return
	}
	p = p.move(r.direction(mansionEast), x).move(r.direction(mansionSouth), z)
	a.add(a.roomName(kind, floor, stairs), p, r.turn(turn), mirror)
}

func (a *mansionAssembler) room2x2(p mansionPosition, r mansionRotation, extension, door mansionDirection, floor int) {
	x, z, turn := 0, 0, 0
	mirror := TemplateMirrorNone
	switch {
	case door == mansionEast && extension == mansionSouth:
		x = -7
	case door == mansionEast && extension == mansionNorth:
		x, z, mirror = -7, 6, TemplateMirrorLeftRight
	case door == mansionNorth && extension == mansionEast:
		x, z, turn = 1, 14, 3
	case door == mansionNorth && extension == mansionWest:
		x, z, turn, mirror = 7, 14, 3, TemplateMirrorLeftRight
	case door == mansionSouth && extension == mansionWest:
		x, z, turn = 7, -8, 1
	case door == mansionSouth && extension == mansionEast:
		x, z, turn, mirror = 1, -8, 1, TemplateMirrorLeftRight
	case door == mansionWest && extension == mansionNorth:
		x, z, turn = 15, 6, 2
	case door == mansionWest && extension == mansionSouth:
		x, mirror = 15, TemplateMirrorFrontBack
	}
	p = p.move(r.direction(mansionEast), x).move(r.direction(mansionSouth), z)
	a.add(a.roomName("2x2", floor, false), p, r.turn(turn), mirror)
}
