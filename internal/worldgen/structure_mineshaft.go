package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

type MineshaftStartSources struct {
	MinY, SeaLevel int32
	BiomeAtQuart   StructureBiomeAtQuart
	// WORLD_SURFACE_WG base height is the first free Y, one above the
	// occupied-height helper used by other structure-start queries.
	WorldSurfaceHeightAt StructureBaseOccupiedHeightAt
}

// GenerateMineshaftStructureStart assembles the original room, corridors,
// crossings and stairs before checking the biome at the shifted generation stub.
func GenerateMineshaftStructureStart(seed int64, cx, cz int32, mesa bool, sources MineshaftStartSources) (*StructureStart, error) {
	if sources.BiomeAtQuart == nil || mesa && sources.WorldSurfaceHeightAt == nil {
		return nil, fmt.Errorf("mineshaft needs biomes and mesa base surface heights")
	}
	base := NewLegacyRandom(seed)
	xm, zm := base.NextLong(), base.NextLong()
	random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	pieces, shift, err := assembleMineshaft(random, cx*16+2, cz*16+2, mesa, sources.MinY, sources.SeaLevel, sources.WorldSurfaceHeightAt)
	if err != nil {
		return nil, err
	}
	biome, err := sources.BiomeAtQuart(cx*4, (50+shift)>>2, cz*4)
	if err != nil {
		return nil, err
	}
	name := "minecraft:mineshaft"
	if mesa {
		name = "minecraft:mineshaft_mesa"
	}
	allowed, err := StructureBiomeNames(name)
	if err != nil {
		return nil, err
	}
	for _, value := range allowed {
		if value == biome {
			return &StructureStart{Structure: name, Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: "none", Pieces: pieces}, nil
		}
	}
	return nil, nil
}

type mineshaftAssembler struct {
	random *LegacyRandom
	mesa   bool
	pieces []StructurePiece
}

func assembleMineshaft(random *LegacyRandom, x, z int32, mesa bool, minY, seaLevel int32, surfaceAt StructureBaseOccupiedHeightAt) ([]StructurePiece, int32, error) {
	random.NextDouble() // The original generation stub discards this draw.
	a := &mineshaftAssembler{random: random, mesa: mesa}
	box := StructureBoundingBox{x, 50, z, x + 7 + random.NextIntBound(6), 54 + random.NextIntBound(6), z + 7 + random.NextIntBound(6)}
	room := a.piece("msroom", box, "", 0)
	room.Data["Entrances"] = nbt.Tag{Type: nbt.List, ListType: nbt.IntArray, Value: []nbt.Tag{}}
	a.pieces = append(a.pieces, room)
	a.children(0)
	start := StructureStart{Pieces: a.pieces}
	whole, _ := start.BoundingBox()
	var shift int32
	if mesa {
		cx := whole.MinX + (whole.MaxX-whole.MinX+1)/2
		cy := whole.MinY + (whole.MaxY-whole.MinY+1)/2
		cz := whole.MinZ + (whole.MaxZ-whole.MinZ+1)/2
		height, err := surfaceAt(cx, cz)
		if err != nil {
			return nil, 0, err
		}
		target := seaLevel
		if height > seaLevel {
			target += random.NextIntBound(height - seaLevel + 1)
		}
		shift = target - cy
	} else {
		top := whole.MaxY - whole.MinY + 1 + minY + 1
		if top < seaLevel-10 {
			top += random.NextIntBound(seaLevel - 10 - top)
		}
		shift = top - whole.MaxY
	}
	for i := range a.pieces {
		piece := &a.pieces[i]
		piece.Box.MinY += shift
		piece.Box.MaxY += shift
		if piece.ID == "minecraft:msroom" {
			entrances, _ := piece.Data["Entrances"].Elements()
			for _, entry := range entrances {
				box := entry.Value.([]int32)
				box[1] += shift
				box[4] += shift
			}
		}
	}
	return a.pieces, shift, nil
}

func (a *mineshaftAssembler) piece(id string, box StructureBoundingBox, direction string, depth int32) StructurePiece {
	typeID := int32(0)
	if a.mesa {
		typeID = 1
	}
	return StructurePiece{ID: "minecraft:" + id, Box: box, Orientation: direction, GenerationDepth: depth, Data: nbt.Object{"MST": {Type: nbt.Int, Value: typeID}}}
}

func (a *mineshaftAssembler) intersects(box StructureBoundingBox) bool {
	for _, piece := range a.pieces {
		other := piece.Box
		if box.MaxX >= other.MinX && box.MinX <= other.MaxX && box.MaxY >= other.MinY && box.MinY <= other.MaxY && box.MaxZ >= other.MinZ && box.MinZ <= other.MaxZ {
			return true
		}
	}
	return false
}

func (a *mineshaftAssembler) expand(x, y, z int32, direction string, depth int32) *StructureBoundingBox {
	root := a.pieces[0].Box
	if depth > 8 || abs32(x-root.MinX) > 80 || abs32(z-root.MinZ) > 80 {
		return nil
	}
	choice := a.random.NextIntBound(100)
	var box StructureBoundingBox
	id := "mscorridor"
	if choice >= 80 {
		id = "mscrossing"
		height := int32(2)
		if a.random.NextIntBound(4) == 0 {
			height = 6
		}
		switch direction {
		case "south":
			box = StructureBoundingBox{x - 1, y, z, x + 3, y + height, z + 4}
		case "west":
			box = StructureBoundingBox{x - 4, y, z - 1, x, y + height, z + 3}
		case "east":
			box = StructureBoundingBox{x, y, z - 1, x + 4, y + height, z + 3}
		default:
			box = StructureBoundingBox{x - 1, y, z - 4, x + 3, y + height, z}
		}
		if a.intersects(box) {
			return nil
		}
	} else if choice >= 70 {
		id = "msstairs"
		switch direction {
		case "south":
			box = StructureBoundingBox{x, y - 5, z, x + 2, y + 2, z + 8}
		case "west":
			box = StructureBoundingBox{x - 8, y - 5, z, x, y + 2, z + 2}
		case "east":
			box = StructureBoundingBox{x, y - 5, z, x + 8, y + 2, z + 2}
		default:
			box = StructureBoundingBox{x, y - 5, z - 8, x + 2, y + 2, z}
		}
		if a.intersects(box) {
			return nil
		}
	} else {
		found := false
		for sections := a.random.NextIntBound(3) + 2; sections > 0; sections-- {
			length := sections * 5
			switch direction {
			case "south":
				box = StructureBoundingBox{x, y, z, x + 2, y + 2, z + length - 1}
			case "west":
				box = StructureBoundingBox{x - length + 1, y, z, x, y + 2, z + 2}
			case "east":
				box = StructureBoundingBox{x, y, z, x + length - 1, y + 2, z + 2}
			default:
				box = StructureBoundingBox{x, y, z - length + 1, x + 2, y + 2, z}
			}
			if !a.intersects(box) {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	piece := a.piece(id, box, direction, depth+1)
	if id == "mscrossing" {
		piece.Orientation = ""
		orientation := map[string]int32{"south": 0, "west": 1, "north": 2, "east": 3}[direction]
		piece.Data["D"] = nbt.Tag{Type: nbt.Int, Value: orientation}
		bit := int8(0)
		if box.MaxY-box.MinY+1 > 3 {
			bit = 1
		}
		piece.Data["tf"] = nbt.Tag{Type: nbt.Byte, Value: bit}
	} else if id == "mscorridor" {
		rail, spider := int8(0), int8(0)
		if a.random.NextIntBound(3) == 0 {
			rail = 1
		} else if a.random.NextIntBound(23) == 0 {
			spider = 1
		}
		length := box.MaxZ - box.MinZ + 1
		if direction == "east" || direction == "west" {
			length = box.MaxX - box.MinX + 1
		}
		piece.Data["hr"] = nbt.Tag{Type: nbt.Byte, Value: rail}
		piece.Data["sc"] = nbt.Tag{Type: nbt.Byte, Value: spider}
		piece.Data["hps"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		piece.Data["Num"] = nbt.Tag{Type: nbt.Int, Value: length / 5}
	}
	index := len(a.pieces)
	a.pieces = append(a.pieces, piece)
	a.children(index)
	return &box
}

func (a *mineshaftAssembler) children(index int) {
	piece := a.pieces[index]
	b, d := piece.Box, piece.GenerationDepth
	r := a.random
	switch piece.ID {
	case "minecraft:msroom":
		yRange := max(int32(1), b.MaxY-b.MinY+1-4)
		entrances := []nbt.Tag{}
		for _, direction := range []string{"north", "south", "west", "east"} {
			span := b.MaxX - b.MinX + 1
			if direction == "west" || direction == "east" {
				span = b.MaxZ - b.MinZ + 1
			}
			for offset := int32(0); offset < span; offset += 4 {
				offset += r.NextIntBound(span)
				if offset+3 > span {
					break
				}
				x, z := b.MinX+offset, b.MinZ-1
				switch direction {
				case "south":
					z = b.MaxZ + 1
				case "west":
					x, z = b.MinX-1, b.MinZ+offset
				case "east":
					x, z = b.MaxX+1, b.MinZ+offset
				}
				child := a.expand(x, b.MinY+r.NextIntBound(yRange)+1, z, direction, d)
				if child == nil {
					continue
				}
				entry := *child
				switch direction {
				case "north":
					entry.MinZ, entry.MaxZ = b.MinZ, b.MinZ+1
				case "south":
					entry.MinZ, entry.MaxZ = b.MaxZ-1, b.MaxZ
				case "west":
					entry.MinX, entry.MaxX = b.MinX, b.MinX+1
				case "east":
					entry.MinX, entry.MaxX = b.MaxX-1, b.MaxX
				}
				entrances = append(entrances, nbt.Tag{Type: nbt.IntArray, Value: []int32{entry.MinX, entry.MinY, entry.MinZ, entry.MaxX, entry.MaxY, entry.MaxZ}})
			}
		}
		a.pieces[index].Data["Entrances"] = nbt.Tag{Type: nbt.List, ListType: nbt.IntArray, Value: entrances}
	case "minecraft:msstairs":
		x, z := b.MinX, b.MinZ-1
		switch piece.Orientation {
		case "south":
			z = b.MaxZ + 1
		case "west":
			x, z = b.MinX-1, b.MinZ
		case "east":
			x, z = b.MaxX+1, b.MinZ
		}
		a.expand(x, b.MinY, z, piece.Orientation, d)
	case "minecraft:mscrossing":
		facing, _ := piece.Data.Number("D")
		switch facing {
		case 2:
			a.expand(b.MinX+1, b.MinY, b.MinZ-1, "north", d)
			a.expand(b.MinX-1, b.MinY, b.MinZ+1, "west", d)
			a.expand(b.MaxX+1, b.MinY, b.MinZ+1, "east", d)
		case 0:
			a.expand(b.MinX+1, b.MinY, b.MaxZ+1, "south", d)
			a.expand(b.MinX-1, b.MinY, b.MinZ+1, "west", d)
			a.expand(b.MaxX+1, b.MinY, b.MinZ+1, "east", d)
		case 1:
			a.expand(b.MinX+1, b.MinY, b.MinZ-1, "north", d)
			a.expand(b.MinX+1, b.MinY, b.MaxZ+1, "south", d)
			a.expand(b.MinX-1, b.MinY, b.MinZ+1, "west", d)
		case 3:
			a.expand(b.MinX+1, b.MinY, b.MinZ-1, "north", d)
			a.expand(b.MinX+1, b.MinY, b.MaxZ+1, "south", d)
			a.expand(b.MaxX+1, b.MinY, b.MinZ+1, "east", d)
		}
		if tall, _ := piece.Data.Number("tf"); tall != 0 {
			if r.NextBoolean() {
				a.expand(b.MinX+1, b.MinY+4, b.MinZ-1, "north", d)
			}
			if r.NextBoolean() {
				a.expand(b.MinX-1, b.MinY+4, b.MinZ+1, "west", d)
			}
			if r.NextBoolean() {
				a.expand(b.MaxX+1, b.MinY+4, b.MinZ+1, "east", d)
			}
			if r.NextBoolean() {
				a.expand(b.MinX+1, b.MinY+4, b.MaxZ+1, "south", d)
			}
		}
	case "minecraft:mscorridor":
		choice := r.NextIntBound(4)
		y := b.MinY - 1 + r.NextIntBound(3)
		switch piece.Orientation {
		case "north":
			if choice <= 1 {
				a.expand(b.MinX, y, b.MinZ-1, "north", d)
			} else if choice == 2 {
				a.expand(b.MinX-1, y, b.MinZ, "west", d)
			} else {
				a.expand(b.MaxX+1, y, b.MinZ, "east", d)
			}
		case "south":
			if choice <= 1 {
				a.expand(b.MinX, y, b.MaxZ+1, "south", d)
			} else if choice == 2 {
				a.expand(b.MinX-1, y, b.MaxZ-3, "west", d)
			} else {
				a.expand(b.MaxX+1, y, b.MaxZ-3, "east", d)
			}
		case "west":
			if choice <= 1 {
				a.expand(b.MinX-1, y, b.MinZ, "west", d)
			} else if choice == 2 {
				a.expand(b.MinX, y, b.MinZ-1, "north", d)
			} else {
				a.expand(b.MinX, y, b.MaxZ+1, "south", d)
			}
		case "east":
			if choice <= 1 {
				a.expand(b.MaxX+1, y, b.MinZ, "east", d)
			} else if choice == 2 {
				a.expand(b.MaxX-3, y, b.MinZ-1, "north", d)
			} else {
				a.expand(b.MaxX-3, y, b.MaxZ+1, "south", d)
			}
		}
		if d >= 8 {
			return
		}
		if piece.Orientation == "north" || piece.Orientation == "south" {
			for z := b.MinZ + 3; z+3 <= b.MaxZ; z += 5 {
				choice := r.NextIntBound(5)
				if choice == 0 {
					a.expand(b.MinX-1, b.MinY, z, "west", d+1)
				} else if choice == 1 {
					a.expand(b.MaxX+1, b.MinY, z, "east", d+1)
				}
			}
		} else {
			for x := b.MinX + 3; x+3 <= b.MaxX; x += 5 {
				choice := r.NextIntBound(5)
				if choice == 0 {
					a.expand(x, b.MinY, b.MinZ-1, "north", d+1)
				} else if choice == 1 {
					a.expand(x, b.MinY, b.MaxZ+1, "south", d+1)
				}
			}
		}
	}
}
