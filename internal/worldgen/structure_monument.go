package worldgen

import (
	"encoding/json"
	"fmt"
	"sync"
)

type OceanMonumentStartSources struct {
	SeaLevel     int32
	BiomeAtQuart StructureBiomeAtQuart
	// Highest solid base-noise block (OCEAN_FLOOR_WG), not the first air block.
	OceanFloorOccupiedHeightAt StructureBaseOccupiedHeightAt
}

type monumentBiomes struct{ surrounding, center map[string]bool }

var monumentBiomeRules = sync.OnceValues(func() (monumentBiomes, error) {
	var registry structureRegistry
	if err := json.Unmarshal(builtinStructureData, &registry); err != nil {
		return monumentBiomes{}, err
	}
	resolve := func(tag string) (map[string]bool, error) {
		names, err := resolveStructureBiomes(registry.BiomeTags, []string{tag})
		result := make(map[string]bool, len(names))
		for _, name := range names {
			result[name] = true
		}
		return result, err
	}
	surrounding, err := resolve("#minecraft:required_ocean_monument_surrounding")
	if err != nil {
		return monumentBiomes{}, err
	}
	center, err := resolve("#minecraft:has_structure/ocean_monument")
	return monumentBiomes{surrounding, center}, err
})

// GenerateOceanMonumentStructureStart checks the original three-dimensional
// biome neighborhood before fitting rooms. Placement-set eligibility belongs
// to the caller; the Overworld pipeline supplies those placement checks.
func GenerateOceanMonumentStructureStart(seed int64, cx, cz int32, sources OceanMonumentStartSources) (*StructureStart, error) {
	if sources.BiomeAtQuart == nil || sources.OceanFloorOccupiedHeightAt == nil {
		return nil, fmt.Errorf("monument start requires biomes and base ocean-floor height")
	}
	biomes, err := monumentBiomeRules()
	if err != nil {
		return nil, err
	}
	x, z := cx*16+9, cz*16+9
	valid := true
	// BiomeSource.getBiomesWithin uses Z/X/Y traversal with inclusive quart
	// bounds. Finish the scan even after rejection: the sampler can retain a
	// previous climate-tree leaf to break distance ties.
	for qz := (z - 29) >> 2; qz <= (z+29)>>2; qz++ {
		for qx := (x - 29) >> 2; qx <= (x+29)>>2; qx++ {
			for qy := (sources.SeaLevel - 29) >> 2; qy <= (sources.SeaLevel+29)>>2; qy++ {
				biome, err := sources.BiomeAtQuart(qx, qy, qz)
				if err != nil {
					return nil, err
				}
				valid = biomes.surrounding[biome] && valid
			}
		}
	}
	if !valid {
		return nil, nil
	}
	x, z = cx*16+8, cz*16+8
	y, err := sources.OceanFloorOccupiedHeightAt(x, z)
	if err != nil {
		return nil, err
	}
	biome, err := sources.BiomeAtQuart(x>>2, y>>2, z>>2)
	if err != nil {
		return nil, err
	}
	if !biomes.center[biome] {
		return nil, nil
	}
	random := NewLegacyRandom(seed)
	xm, zm := random.NextLong(), random.NextLong()
	random = NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	direction := [...]string{"north", "east", "south", "west"}[random.NextIntBound(4)]
	layout, err := assembleMonument(random, cx*16-29, cz*16-29, direction)
	if err != nil {
		return nil, err
	}
	return &StructureStart{Structure: "minecraft:monument", Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: "none", Pieces: []StructurePiece{layout.building}}, nil
}

// Monument rooms use Direction's 3D indices, not its horizontal random order.
const (
	monumentDown = iota
	monumentUp
	monumentNorth
	monumentSouth
	monumentWest
	monumentEast
)

type monumentRoom struct {
	id              int
	connections     [6]*monumentRoom
	open            [6]bool
	claimed, source bool
	visit           int
}

type monumentChild struct {
	piece   StructurePiece
	room    *monumentRoom
	variant int32
}

// Only Building is serialized by vanilla. The room graph, fitted children and
// variants are reconstructed from the world seed when a monument is loaded.
type monumentLayout struct {
	building StructurePiece
	children []monumentChild
	rooms    []*monumentRoom
}

func (room *monumentRoom) connect(direction int, other *monumentRoom) {
	room.connections[direction] = other
	other.connections[direction^1] = room
}

func (room *monumentRoom) updateOpenings() {
	for d, other := range room.connections {
		room.open[d] = other != nil
	}
}

func (room *monumentRoom) reachesSource(scan int) bool {
	if room.source {
		return true
	}
	room.visit = scan
	for d, other := range room.connections {
		if other != nil && room.open[d] && other.visit != scan && other.reachesSource(scan) {
			return true
		}
	}
	return false
}

func monumentRoomIndex(x, y, z int) int { return y*25 + z*5 + x }

func monumentRoomGraph(random *LegacyRandom) (rooms []*monumentRoom, source, core *monumentRoom) {
	var grid [75]*monumentRoom
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			for z := 0; z < 4; z++ {
				if y == 2 && (x == 0 || x == 4 || z >= 2) {
					continue
				}
				id := monumentRoomIndex(x, y, z)
				grid[id] = &monumentRoom{id: id}
			}
		}
	}
	steps := [6][3]int{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}}
	for x := 0; x < 5; x++ {
		for z := 0; z < 5; z++ {
			for y := 0; y < 3; y++ {
				room := grid[monumentRoomIndex(x, y, z)]
				if room == nil {
					continue
				}
				for d, step := range steps {
					nx, ny, nz := x+step[0], y+step[1], z+step[2]
					if nx < 0 || nx >= 5 || ny < 0 || ny >= 3 || nz < 0 || nz >= 5 {
						continue
					}
					other := grid[monumentRoomIndex(nx, ny, nz)]
					if other == nil {
						continue
					}
					// Graph north points toward increasing grid Z.
					if nz != z {
						d ^= 1
					}
					room.connect(d, other)
				}
			}
		}
	}
	top, left, right := &monumentRoom{id: 1003, claimed: true}, &monumentRoom{id: 1001, claimed: true}, &monumentRoom{id: 1002, claimed: true}
	grid[52].connect(monumentUp, top)
	grid[25].connect(monumentSouth, left)
	grid[29].connect(monumentSouth, right)
	source = grid[2]
	source.source = true
	core = grid[monumentRoomIndex(int(random.NextIntBound(4)), 0, 2)]
	for dx := 0; dx < 2; dx++ {
		for dy := 0; dy < 2; dy++ {
			for dz := 0; dz < 2; dz++ {
				grid[core.id+dx+dy*25+dz*5].claimed = true
			}
		}
	}
	for _, room := range grid {
		if room != nil {
			room.updateOpenings()
			rooms = append(rooms, room)
		}
	}
	// The wing sentinels intentionally retain false openings; only the top
	// sentinel initializes its reverse edge before the connectivity checks.
	top.updateOpenings()
	for n := len(rooms); n > 1; n-- {
		i := int(random.NextIntBound(int32(n)))
		rooms[n-1], rooms[i] = rooms[i], rooms[n-1]
	}
	scan := 1
	for _, room := range rooms {
		removed := 0
		for attempt := 0; removed < 2 && attempt < 5; attempt++ {
			d := int(random.NextIntBound(6))
			if !room.open[d] {
				continue
			}
			other := room.connections[d]
			room.open[d], other.open[d^1] = false, false
			connected := room.reachesSource(scan)
			scan++
			// Short-circuiting also controls the scan counter in vanilla.
			if connected {
				connected = other.reachesSource(scan)
				scan++
			}
			if connected {
				removed++
			} else {
				room.open[d], other.open[d^1] = true, true
			}
		}
	}
	rooms = append(rooms, top, left, right)
	return
}

func monumentRoomBox(room *monumentRoom, direction string, sx, sy, sz int32) StructureBoundingBox {
	x, z, y := int32(room.id%5), int32(room.id/5%5), int32(room.id/25)
	width, depth := sx*8, sz*8
	if direction == "east" || direction == "west" {
		width, depth = depth, width
	}
	box := StructureBoundingBox{0, 0, 0, width - 1, sy*4 - 1, depth - 1}
	var ox, oz int32
	switch direction {
	case "north":
		ox, oz = x*8, -(z+sz)*8+1
	case "south":
		ox, oz = x*8, z*8
	case "west":
		ox, oz = -(z+sz)*8+1, x*8
	case "east":
		ox, oz = z*8, x*8
	}
	box.MinX += ox
	box.MaxX += ox
	box.MinY += y * 4
	box.MaxY += y * 4
	box.MinZ += oz
	box.MaxZ += oz
	return box
}

func assembleMonument(random *LegacyRandom, x, z int32, direction string) (*monumentLayout, error) {
	if random == nil {
		return nil, fmt.Errorf("monument layout requires a random source")
	}
	if direction != "north" && direction != "south" && direction != "east" && direction != "west" {
		return nil, fmt.Errorf("invalid monument direction %q", direction)
	}
	layout := &monumentLayout{building: StructurePiece{ID: "minecraft:omb", Orientation: direction, Box: StructureBoundingBox{x, 39, z, x + 57, 61, z + 57}}}
	rooms, source, core := monumentRoomGraph(random)
	layout.rooms = rooms
	add := func(id string, room *monumentRoom, sx, sy, sz int32, variant int32) {
		layout.children = append(layout.children, monumentChild{StructurePiece{ID: "minecraft:" + id, Orientation: direction, GenerationDepth: 1, Box: monumentRoomBox(room, direction, sx, sy, sz)}, room, variant})
	}
	source.claimed = true
	add("omentry", source, 1, 1, 1, -1)
	add("omcr", core, 2, 2, 2, -1)
	available := func(room *monumentRoom, d int) bool { return room.open[d] && !room.connections[d].claimed }
	for _, room := range rooms {
		if room.claimed || room.id >= 75 {
			continue
		}
		switch {
		case available(room, monumentEast) && available(room, monumentUp) && available(room.connections[monumentEast], monumentUp):
			room.claimed = true
			room.connections[monumentEast].claimed = true
			room.connections[monumentUp].claimed = true
			room.connections[monumentEast].connections[monumentUp].claimed = true
			add("omdxyr", room, 2, 2, 1, -1)
		case available(room, monumentNorth) && available(room, monumentUp) && available(room.connections[monumentNorth], monumentUp):
			room.claimed = true
			room.connections[monumentNorth].claimed = true
			room.connections[monumentUp].claimed = true
			room.connections[monumentNorth].connections[monumentUp].claimed = true
			add("omdyzr", room, 1, 2, 2, -1)
		case available(room, monumentNorth):
			room.claimed = true
			room.connections[monumentNorth].claimed = true
			add("omdzr", room, 1, 1, 2, -1)
		case available(room, monumentEast):
			room.claimed = true
			room.connections[monumentEast].claimed = true
			add("omdxr", room, 2, 1, 1, -1)
		case available(room, monumentUp):
			room.claimed = true
			room.connections[monumentUp].claimed = true
			add("omdyr", room, 1, 2, 1, -1)
		case !room.open[monumentWest] && !room.open[monumentEast] && !room.open[monumentNorth] && !room.open[monumentSouth] && !room.open[monumentUp]:
			room.claimed = true
			add("omsimplet", room, 1, 1, 1, -1)
		default:
			room.claimed = true
			add("omsimple", room, 1, 1, 1, random.NextIntBound(3))
		}
	}
	offset := layout.building.BlockPosition(9, 0, 22)
	for i := range layout.children {
		box := &layout.children[i].piece.Box
		box.MinX += offset.X
		box.MaxX += offset.X
		box.MinY += offset.Y
		box.MaxY += offset.Y
		box.MinZ += offset.Z
		box.MaxZ += offset.Z
	}
	boxAt := func(x1, y1, z1, x2, y2, z2 int32) StructureBoundingBox {
		a, b := layout.building.BlockPosition(x1, y1, z1), layout.building.BlockPosition(x2, y2, z2)
		return StructureBoundingBox{min(a.X, b.X), min(a.Y, b.Y), min(a.Z, b.Z), max(a.X, b.X), max(a.Y, b.Y), max(a.Z, b.Z)}
	}
	wing := random.NextInt()
	for i, box := range []StructureBoundingBox{boxAt(1, 1, 1, 23, 8, 21), boxAt(34, 1, 1, 56, 8, 21), boxAt(22, 13, 22, 35, 17, 35)} {
		id, variant := "omwr", (wing+int32(i))&1
		if i == 2 {
			id, variant = "ompenthouse", -1
		}
		layout.children = append(layout.children, monumentChild{StructurePiece{ID: "minecraft:" + id, Orientation: direction, GenerationDepth: 1, Box: box}, nil, variant})
	}
	return layout, nil
}
