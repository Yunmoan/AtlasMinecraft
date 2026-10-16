package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

type StrongholdStartSources struct {
	MinY, SeaLevel int32
	BiomeAtQuart   StructureBiomeAtQuart
}

// GenerateStrongholdStructureStart assembles the original weighted pieces,
// retrying the original seed+attempt stream until a portal room exists. The
// generation stub biome is queried at the chunk minimum, Y=0, before assembly.
func GenerateStrongholdStructureStart(seed int64, cx, cz int32, sources StrongholdStartSources) (*StructureStart, error) {
	if sources.BiomeAtQuart == nil {
		return nil, fmt.Errorf("stronghold start requires a biome source")
	}
	biome, err := sources.BiomeAtQuart(cx*4, 0, cz*4)
	if err != nil {
		return nil, err
	}
	allowed, err := StructureBiomeNames("minecraft:stronghold")
	if err != nil {
		return nil, err
	}
	for _, name := range allowed {
		if name == biome {
			pieces, _, _ := assembleStronghold(seed, cx, cz, sources.MinY, sources.SeaLevel)
			return &StructureStart{Structure: "minecraft:stronghold", Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: "none", Pieces: pieces}, nil
		}
	}
	return nil, nil
}

type strongholdPieceWeight struct {
	id                          string
	weight, limit, count, after int32
}

type strongholdAssembler struct {
	random   *LegacyRandom
	pieces   []StructurePiece
	pending  []int
	weights  []*strongholdPieceWeight
	previous *strongholdPieceWeight
	imposed  string
	portal   bool
}

func strongholdLargeFeatureRandom(seed int64, cx, cz int32) *LegacyRandom {
	random := NewLegacyRandom(seed)
	xm, zm := random.NextLong(), random.NextLong()
	return NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
}

func assembleStronghold(seed int64, cx, cz, minY, seaLevel int32) ([]StructurePiece, *LegacyRandom, int32) {
	for attempt := int32(0); ; attempt++ {
		a := &strongholdAssembler{random: strongholdLargeFeatureRandom(int64(uint64(seed)+uint64(attempt)), cx, cz)}
		a.weights = []*strongholdPieceWeight{
			{"shs", 40, 0, 0, -1}, {"shph", 5, 5, 0, -1}, {"shlt", 20, 0, 0, -1}, {"shrt", 20, 0, 0, -1},
			{"shrc", 10, 6, 0, -1}, {"shssd", 5, 5, 0, -1}, {"shsd", 5, 5, 0, -1}, {"sh5c", 5, 4, 0, -1},
			{"shcc", 5, 4, 0, -1}, {"shli", 10, 2, 0, 4}, {"shpr", 20, 1, 0, 5},
		}
		direction := [...]string{"north", "east", "south", "west"}[a.random.NextIntBound(4)]
		x, z := cx*16+2, cz*16+2
		a.pieces = []StructurePiece{{ID: "minecraft:shstart", Orientation: direction, Box: StructureBoundingBox{x, 64, z, x + 4, 74, z + 4}, Data: nbt.Object{"EntryDoor": {Type: nbt.String, Value: "OPENING"}, "Source": {Type: nbt.Byte, Value: int8(1)}}}}
		a.children(0)
		for len(a.pending) > 0 {
			at := int(a.random.NextIntBound(int32(len(a.pending))))
			index := a.pending[at]
			a.pending = append(a.pending[:at], a.pending[at+1:]...)
			a.children(index)
		}
		start := StructureStart{Pieces: a.pieces}
		box, _ := start.BoundingBox()
		top := box.MaxY - box.MinY + 1 + minY + 1
		if top < seaLevel-10 {
			top += a.random.NextIntBound(seaLevel - 10 - top)
		}
		shift := top - box.MaxY
		for index := range a.pieces {
			a.pieces[index].Box.MinY += shift
			a.pieces[index].Box.MaxY += shift
		}
		if a.portal {
			return a.pieces, a.random, attempt + 1
		}
	}
}

func strongholdBoxesIntersect(a, b StructureBoundingBox) bool {
	return a.MaxX >= b.MinX && a.MinX <= b.MaxX && a.MaxY >= b.MinY && a.MinY <= b.MaxY && a.MaxZ >= b.MinZ && a.MinZ <= b.MaxZ
}

func (a *strongholdAssembler) collision(box StructureBoundingBox) *StructurePiece {
	for index := range a.pieces {
		if strongholdBoxesIntersect(box, a.pieces[index].Box) {
			return &a.pieces[index]
		}
	}
	return nil
}

var strongholdPieceDimensions = map[string][6]int32{
	"shs": {-1, -1, 0, 5, 5, 7}, "shph": {-1, -1, 0, 9, 5, 11}, "shlt": {-1, -1, 0, 5, 5, 5}, "shrt": {-1, -1, 0, 5, 5, 5},
	"shrc": {-4, -1, 0, 11, 7, 11}, "shssd": {-1, -7, 0, 5, 11, 8}, "shsd": {-1, -7, 0, 5, 11, 5}, "sh5c": {-4, -3, 0, 10, 9, 11},
	"shcc": {-1, -1, 0, 5, 5, 7}, "shli": {-4, -1, 0, 14, 11, 15}, "shpr": {-4, -1, 0, 11, 8, 16},
}

func strongholdSetBool(data nbt.Object, name string, value bool) {
	flag := int8(0)
	if value {
		flag = 1
	}
	data[name] = nbt.Tag{Type: nbt.Byte, Value: flag}
}

func strongholdFlag(piece StructurePiece, name string) bool {
	value, _ := piece.Data.Number(name)
	return value != 0
}

func (a *strongholdAssembler) create(id string, x, y, z int32, direction string, depth int32) *StructurePiece {
	d := strongholdPieceDimensions[id]
	box := fortressOrientedBox(x, y, z, d[0], d[1], d[2], d[3], d[4], d[5], direction)
	if id == "shli" && (box.MinY <= 10 || a.collision(box) != nil) {
		box = fortressOrientedBox(x, y, z, -4, -1, 0, 14, 6, 15, direction)
	}
	if box.MinY <= 10 || a.collision(box) != nil {
		return nil
	}
	data := nbt.Object{"EntryDoor": {Type: nbt.String, Value: "OPENING"}}
	if id != "shpr" {
		door := "OPENING"
		switch a.random.NextIntBound(5) {
		case 2:
			door = "WOOD_DOOR"
		case 3:
			door = "GRATES"
		case 4:
			door = "IRON_DOOR"
		}
		data["EntryDoor"] = nbt.Tag{Type: nbt.String, Value: door}
	}
	switch id {
	case "shs":
		strongholdSetBool(data, "Left", a.random.NextIntBound(2) == 0)
		strongholdSetBool(data, "Right", a.random.NextIntBound(2) == 0)
	case "shrc":
		data["Type"] = nbt.Tag{Type: nbt.Int, Value: a.random.NextIntBound(5)}
	case "shsd":
		strongholdSetBool(data, "Source", false)
	case "sh5c":
		strongholdSetBool(data, "leftLow", a.random.NextBoolean())
		strongholdSetBool(data, "leftHigh", a.random.NextBoolean())
		strongholdSetBool(data, "rightLow", a.random.NextBoolean())
		strongholdSetBool(data, "rightHigh", a.random.NextIntBound(3) > 0)
	case "shcc":
		strongholdSetBool(data, "Chest", false)
	case "shli":
		strongholdSetBool(data, "Tall", box.MaxY-box.MinY+1 > 6)
	case "shpr":
		strongholdSetBool(data, "Mob", false)
	}
	return &StructurePiece{ID: "minecraft:" + id, Box: box, Orientation: direction, GenerationDepth: depth, Data: data}
}

func (a *strongholdAssembler) filler(x, y, z int32, direction string, depth int32) *StructurePiece {
	box := fortressOrientedBox(x, y, z, -1, -1, 0, 5, 5, 4, direction)
	other := a.collision(box)
	if other == nil || other.Box.MinY != box.MinY {
		return nil
	}
	for length := int32(2); length >= 1; length-- {
		box = fortressOrientedBox(x, y, z, -1, -1, 0, 5, 5, length, direction)
		if strongholdBoxesIntersect(other.Box, box) {
			continue
		}
		box = fortressOrientedBox(x, y, z, -1, -1, 0, 5, 5, length+1, direction)
		if box.MinY <= 1 {
			return nil
		}
		return &StructurePiece{ID: "minecraft:shfc", Box: box, Orientation: direction, GenerationDepth: depth, Data: nbt.Object{"EntryDoor": {Type: nbt.String, Value: "OPENING"}, "Steps": {Type: nbt.Int, Value: length + 1}}}
	}
	return nil
}

func (a *strongholdAssembler) selectPiece(x, y, z int32, direction string, depth int32) *StructurePiece {
	total, available := int32(0), false
	for _, w := range a.weights {
		total += w.weight
		available = available || w.limit > 0 && w.count < w.limit
	}
	if !available {
		return nil
	}
	if a.imposed != "" {
		piece := a.create(a.imposed, x, y, z, direction, depth)
		a.imposed = ""
		if piece != nil {
			return piece
		}
	}
attempts:
	for attempt := 0; attempt < 5; attempt++ {
		choice := a.random.NextIntBound(total)
		for index, w := range a.weights {
			choice -= w.weight
			if choice >= 0 {
				continue
			}
			if w.limit > 0 && w.count >= w.limit || depth <= w.after || w == a.previous {
				continue attempts
			}
			piece := a.create(w.id, x, y, z, direction, depth)
			if piece == nil {
				continue
			}
			w.count++
			a.previous = w
			if w.limit > 0 && w.count >= w.limit {
				a.weights = append(a.weights[:index], a.weights[index+1:]...)
			}
			return piece
		}
	}
	return a.filler(x, y, z, direction, depth)
}

func (a *strongholdAssembler) generatePiece(x, y, z int32, direction string, depth int32) {
	root := a.pieces[0].Box
	if depth > 50 || int64(x)-int64(root.MinX) > 112 || int64(root.MinX)-int64(x) > 112 || int64(z)-int64(root.MinZ) > 112 || int64(root.MinZ)-int64(z) > 112 {
		return
	}
	if piece := a.selectPiece(x, y, z, direction, depth+1); piece != nil {
		a.pending = append(a.pending, len(a.pieces))
		a.pieces = append(a.pieces, *piece)
	}
}

func (a *strongholdAssembler) attach(piece StructurePiece, side string, first, second int32) {
	b, direction := piece.Box, piece.Orientation
	x, y, z := int32(0), int32(0), int32(0)
	switch side {
	case "forward":
		y = b.MinY + second
		switch direction {
		case "north":
			x, z = b.MinX+first, b.MinZ-1
		case "south":
			x, z = b.MinX+first, b.MaxZ+1
		case "west":
			x, z = b.MinX-1, b.MinZ+first
		case "east":
			x, z = b.MaxX+1, b.MinZ+first
		}
	case "left", "right":
		y = b.MinY + first
		if direction == "north" || direction == "south" {
			z = b.MinZ + second
			x, direction = b.MinX-1, "west"
			if side == "right" {
				x, direction = b.MaxX+1, "east"
			}
		} else {
			x = b.MinX + second
			z, direction = b.MinZ-1, "north"
			if side == "right" {
				z, direction = b.MaxZ+1, "south"
			}
		}
	}
	a.generatePiece(x, y, z, direction, piece.GenerationDepth)
}

func (a *strongholdAssembler) children(index int) {
	piece := a.pieces[index]
	switch piece.ID {
	case "minecraft:shstart", "minecraft:shsd":
		if strongholdFlag(piece, "Source") {
			a.imposed = "sh5c"
		}
		a.attach(piece, "forward", 1, 1)
	case "minecraft:shs":
		a.attach(piece, "forward", 1, 1)
		if strongholdFlag(piece, "Left") {
			a.attach(piece, "left", 1, 2)
		}
		if strongholdFlag(piece, "Right") {
			a.attach(piece, "right", 1, 2)
		}
	case "minecraft:shph", "minecraft:shssd", "minecraft:shcc":
		a.attach(piece, "forward", 1, 1)
	case "minecraft:shlt", "minecraft:shrt":
		side := "right"
		if piece.Orientation == "north" || piece.Orientation == "east" {
			side = "left"
		}
		if piece.ID == "minecraft:shrt" {
			if side == "left" {
				side = "right"
			} else {
				side = "left"
			}
		}
		a.attach(piece, side, 1, 1)
	case "minecraft:shrc":
		a.attach(piece, "forward", 4, 1)
		a.attach(piece, "left", 1, 4)
		a.attach(piece, "right", 1, 4)
	case "minecraft:sh5c":
		low, high := int32(3), int32(5)
		if piece.Orientation == "west" || piece.Orientation == "north" {
			low, high = 8-low, 8-high
		}
		a.attach(piece, "forward", 5, 1)
		for _, child := range []struct {
			flag, side    string
			first, second int32
		}{{"leftLow", "left", low, 1}, {"leftHigh", "left", high, 7}, {"rightLow", "right", low, 1}, {"rightHigh", "right", high, 7}} {
			if strongholdFlag(piece, child.flag) {
				a.attach(piece, child.side, child.first, child.second)
			}
		}
	case "minecraft:shpr":
		a.portal = true
	}
}
