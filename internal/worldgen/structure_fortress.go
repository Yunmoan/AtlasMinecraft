package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

type fortressPieceWeight struct {
	id                   string
	weight, limit, count int32
	repeat               bool
}
type fortressAssembler struct {
	random         *LegacyRandom
	pieces         []StructurePiece
	pending        []int
	bridge, castle []*fortressPieceWeight
	previous       *fortressPieceWeight
}

// GenerateNetherFortressStructureStart ports the original piece assembler. The
// biome is checked at the original generation stub (chunk minimum X/Z, Y=64).
// Blocks are placed later, in each intersected chunk's structure feature step.
func GenerateNetherFortressStructureStart(seed int64, chunkX, chunkZ int32, biomeAtQuart StructureBiomeAtQuart) (*StructureStart, error) {
	if biomeAtQuart == nil {
		return nil, fmt.Errorf("fortress start requires a biome source")
	}
	biome, err := biomeAtQuart(chunkX*4, 16, chunkZ*4)
	if err != nil {
		return nil, err
	}
	allowed, err := StructureBiomeNames("minecraft:fortress")
	if err != nil {
		return nil, err
	}
	valid := false
	for _, name := range allowed {
		valid = valid || name == biome
	}
	if !valid {
		return nil, nil
	}
	random := NewLegacyRandom(seed)
	mx, mz := random.NextLong(), random.NextLong()
	random = NewLegacyRandom(int64(uint64(int64(chunkX))*mx ^ uint64(int64(chunkZ))*mz ^ uint64(seed)))
	pieces := assembleNetherFortress(random, chunkX*16+2, chunkZ*16+2)
	return &StructureStart{Structure: "minecraft:fortress", Chunk: StructureChunkPosition{X: chunkX, Z: chunkZ}, TerrainAdjustment: "none", Pieces: pieces}, nil
}

func assembleNetherFortress(random *LegacyRandom, x, z int32) []StructurePiece {
	a := &fortressAssembler{random: random}
	a.bridge = []*fortressPieceWeight{{"nebs", 30, 0, 0, true}, {"nebcr", 10, 4, 0, false}, {"nerc", 10, 4, 0, false}, {"nesr", 10, 3, 0, false}, {"nemt", 5, 2, 0, false}, {"nece", 5, 1, 0, false}}
	a.castle = []*fortressPieceWeight{{"nesc", 25, 0, 0, true}, {"nescsc", 15, 5, 0, false}, {"nescrt", 5, 10, 0, false}, {"nesclt", 5, 10, 0, false}, {"neccs", 10, 3, 0, true}, {"nectb", 7, 2, 0, false}, {"necsr", 5, 2, 0, false}}
	orientation := []string{"north", "east", "south", "west"}[random.NextIntBound(4)]
	// The live StartPiece constructor uses the crossing identity; the separate
	// nestart ID is used only by its NBT-loading constructor in the original JAR.
	a.pieces = []StructurePiece{{ID: "minecraft:nebcr", Orientation: orientation, Box: StructureBoundingBox{MinX: x, MinY: 64, MinZ: z, MaxX: x + 18, MaxY: 73, MaxZ: z + 18}, Data: make(nbt.Object)}}
	a.children(0)
	for len(a.pending) > 0 {
		i := int(random.NextIntBound(int32(len(a.pending))))
		index := a.pending[i]
		a.pending = append(a.pending[:i], a.pending[i+1:]...)
		a.children(index)
	}
	minY, maxY := a.pieces[0].Box.MinY, a.pieces[0].Box.MaxY
	for _, piece := range a.pieces[1:] {
		minY = min(minY, piece.Box.MinY)
		maxY = max(maxY, piece.Box.MaxY)
	}
	rangeY := int32(70-48+1) - (maxY - minY + 1)
	targetY := int32(48)
	if rangeY > 1 {
		targetY += random.NextIntBound(rangeY)
	}
	offset := targetY - minY
	for i := range a.pieces {
		a.pieces[i].Box.MinY += offset
		a.pieces[i].Box.MaxY += offset
	}
	return a.pieces
}

func fortressOrientedBox(x, y, z, ox, oy, oz, sx, sy, sz int32, direction string) StructureBoundingBox {
	switch direction {
	case "north":
		return StructureBoundingBox{x + ox, y + oy, z - sz + 1 + oz, x + sx - 1 + ox, y + sy - 1 + oy, z + oz}
	case "west":
		return StructureBoundingBox{x - sz + 1 + oz, y + oy, z + ox, x + oz, y + sy - 1 + oy, z + sx - 1 + ox}
	case "east":
		return StructureBoundingBox{x + oz, y + oy, z + ox, x + sz - 1 + oz, y + sy - 1 + oy, z + sx - 1 + ox}
	default:
		return StructureBoundingBox{x + ox, y + oy, z + oz, x + sx - 1 + ox, y + sy - 1 + oy, z + sz - 1 + oz}
	}
}

var fortressPieceDimensions = map[string][6]int32{
	"nebs": {-1, -3, 0, 5, 10, 19}, "nebcr": {-8, -3, 0, 19, 10, 19}, "nerc": {-2, 0, 0, 7, 9, 7}, "nesr": {-2, 0, 0, 7, 11, 7}, "nemt": {-2, 0, 0, 7, 8, 9}, "nece": {-5, -3, 0, 13, 14, 13},
	"nesc": {-1, 0, 0, 5, 7, 5}, "nescsc": {-1, 0, 0, 5, 7, 5}, "nescrt": {-1, 0, 0, 5, 7, 5}, "nesclt": {-1, 0, 0, 5, 7, 5}, "neccs": {-1, -7, 0, 5, 14, 10}, "nectb": {-3, 0, 0, 9, 7, 9}, "necsr": {-5, -3, 0, 13, 14, 13}, "nebef": {-1, -3, 0, 5, 10, 8},
}

func (a *fortressAssembler) create(id string, x, y, z int32, orientation string, depth int32) *StructurePiece {
	d := fortressPieceDimensions[id]
	box := fortressOrientedBox(x, y, z, d[0], d[1], d[2], d[3], d[4], d[5], orientation)
	if box.MinY <= 10 {
		return nil
	}
	for _, piece := range a.pieces {
		b := piece.Box
		if box.MaxX >= b.MinX && box.MinX <= b.MaxX && box.MaxY >= b.MinY && box.MinY <= b.MaxY && box.MaxZ >= b.MinZ && box.MinZ <= b.MaxZ {
			return nil
		}
	}
	piece := &StructurePiece{ID: "minecraft:" + id, Box: box, Orientation: orientation, GenerationDepth: depth, Data: make(nbt.Object)}
	switch id {
	case "nebef":
		piece.Data["Seed"] = nbt.Tag{Type: nbt.Int, Value: a.random.NextInt()}
	case "nescrt", "nesclt":
		piece.Data["Chest"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		if a.random.NextIntBound(3) == 0 {
			piece.Data["Chest"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
		}
	case "nemt":
		piece.Data["Mob"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	}
	return piece
}

func (a *fortressAssembler) selectPiece(weights *[]*fortressPieceWeight, x, y, z int32, orientation string, depth int32) *StructurePiece {
	total := int32(0)
	available := false
	for _, w := range *weights {
		total += w.weight
		available = available || (w.limit > 0 && w.count < w.limit)
	}
attempts:
	for attempt := 0; attempt < 5 && available && total > 0 && depth <= 30; attempt++ {
		choice := a.random.NextIntBound(total)
		for index, w := range *weights {
			choice -= w.weight
			if choice >= 0 {
				continue
			}
			if (w.limit > 0 && w.count >= w.limit) || (w == a.previous && !w.repeat) {
				continue attempts
			}
			piece := a.create(w.id, x, y, z, orientation, depth)
			if piece == nil {
				continue
			}
			w.count++
			a.previous = w
			if w.limit > 0 && w.count >= w.limit {
				*weights = append((*weights)[:index], (*weights)[index+1:]...)
			}
			return piece
		}
	}
	return a.create("nebef", x, y, z, orientation, depth)
}

func (a *fortressAssembler) attach(piece StructurePiece, side string, horizontal, vertical int32, castle bool) {
	b := piece.Box
	orientation := piece.Orientation
	x, y, z := int32(0), int32(0), int32(0)
	switch side {
	case "forward":
		y = b.MinY + vertical
		switch orientation {
		case "north":
			x, z = b.MinX+horizontal, b.MinZ-1
		case "south":
			x, z = b.MinX+horizontal, b.MaxZ+1
		case "west":
			x, z = b.MinX-1, b.MinZ+horizontal
		case "east":
			x, z = b.MaxX+1, b.MinZ+horizontal
		}
	case "left", "right":
		y = b.MinY + horizontal
		if orientation == "north" || orientation == "south" {
			z = b.MinZ + vertical
			if side == "left" {
				x, orientation = b.MinX-1, "west"
			} else {
				x, orientation = b.MaxX+1, "east"
			}
		}
		if piece.Orientation == "west" || piece.Orientation == "east" {
			x = b.MinX + vertical
			if side == "left" {
				z, orientation = b.MinZ-1, "north"
			} else {
				z, orientation = b.MaxZ+1, "south"
			}
		}
	}
	origin := a.pieces[0].Box
	if fortressAbs(x-origin.MinX) > 112 || fortressAbs(z-origin.MinZ) > 112 {
		a.create("nebef", x, y, z, orientation, piece.GenerationDepth)
		return
	}
	weights := &a.bridge
	if castle {
		weights = &a.castle
	}
	child := a.selectPiece(weights, x, y, z, orientation, piece.GenerationDepth+1)
	if child != nil {
		a.pending = append(a.pending, len(a.pieces))
		a.pieces = append(a.pieces, *child)
	}
}

func fortressAbs(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}

func (a *fortressAssembler) children(index int) {
	piece := a.pieces[index]
	attach := func(side string, h, v int32, castle bool) { a.attach(piece, side, h, v, castle) }
	switch piece.ID {
	case "minecraft:nebcr":
		attach("forward", 8, 3, false)
		attach("left", 3, 8, false)
		attach("right", 3, 8, false)
	case "minecraft:nebs":
		attach("forward", 1, 3, false)
	case "minecraft:nerc":
		attach("forward", 2, 0, false)
		attach("left", 0, 2, false)
		attach("right", 0, 2, false)
	case "minecraft:nesr":
		attach("right", 6, 2, false)
	case "minecraft:nece":
		attach("forward", 5, 3, true)
	case "minecraft:nesc", "minecraft:neccs":
		attach("forward", 1, 0, true)
	case "minecraft:nescsc":
		attach("forward", 1, 0, true)
		attach("left", 0, 1, true)
		attach("right", 0, 1, true)
	case "minecraft:nescrt":
		attach("right", 0, 1, true)
	case "minecraft:nesclt":
		attach("left", 0, 1, true)
	case "minecraft:nectb":
		offset := int32(1)
		if piece.Orientation == "west" || piece.Orientation == "north" {
			offset = 5
		}
		attach("left", 0, offset, a.random.NextIntBound(8) > 0)
		attach("right", 0, offset, a.random.NextIntBound(8) > 0)
	case "minecraft:necsr":
		attach("forward", 5, 3, true)
		attach("forward", 5, 11, true)
	}
}
