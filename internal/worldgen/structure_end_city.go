package worldgen

import (
	"atlas/internal/nbt"
	"embed"
	"fmt"
	"strings"
	"sync"
)

//go:embed end_city_data/*.nbt
var endCityAssets embed.FS

var loadEndCityTemplates = sync.OnceValues(func() (map[string]*StructureTemplate, error) {
	files, err := endCityAssets.ReadDir("end_city_data")
	if err != nil {
		return nil, err
	}
	result := make(map[string]*StructureTemplate)
	for _, file := range files {
		data, err := endCityAssets.ReadFile("end_city_data/" + file.Name())
		if err != nil {
			return nil, err
		}
		template, err := DecodeStructureTemplate(data)
		if err != nil {
			return nil, err
		}
		result[strings.TrimSuffix(file.Name(), ".nbt")] = template
	}
	if len(result) != 20 {
		return nil, fmt.Errorf("expected20 original End City templates")
	}
	return result, nil
})

func LoadEndCityTemplate(name string) (*StructureTemplate, error) {
	name = strings.TrimPrefix(strings.TrimPrefix(name, "minecraft:"), "end_city/")
	templates, err := loadEndCityTemplates()
	if err != nil {
		return nil, err
	}
	template := templates[name]
	if template == nil {
		return nil, fmt.Errorf("unknown End City template%s", name)
	}
	return template, nil
}

func NewEndCityTemplatePiece(name string, origin FeaturePosition, rotation TemplateRotation, overwrite bool) (*StructurePiece, error) {
	template, err := LoadEndCityTemplate(name)
	if err != nil {
		return nil, err
	}
	box, err := template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	if err != nil {
		return nil, err
	}
	flag := int8(0)
	if overwrite {
		flag = 1
	}
	return &StructurePiece{ID: "minecraft:ecp", Box: box, Orientation: "north", Data: nbt.Object{
		"Template": {Type: nbt.String, Value: name}, "Rot": {Type: nbt.String, Value: templateRotationNames[rotation]}, "OW": {Type: nbt.Byte, Value: flag},
		"TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z},
	}}, nil
}

type endCityAssembler struct {
	random      *LegacyRandom
	shipCreated bool
	err         error
}

func endCityOrigin(piece *StructurePiece) FeaturePosition {
	x, _ := piece.Data.Number("TPX")
	y, _ := piece.Data.Number("TPY")
	z, _ := piece.Data.Number("TPZ")
	return FeaturePosition{int32(x), int32(y), int32(z)}
}
func endCityRotation(piece *StructurePiece) TemplateRotation {
	text, _ := piece.Data.Text("Rot")
	for index, name := range templateRotationNames {
		if text == name {
			return TemplateRotation(index)
		}
	}
	return TemplateRotationNone
}
func (a *endCityAssembler) add(list *[]*StructurePiece, parent *StructurePiece, offset FeaturePosition, name string, rotation TemplateRotation, overwrite bool) *StructurePiece {
	if a.err != nil {
		return parent
	}
	position := offset
	if parent != nil {
		rotated, _ := TransformTemplatePosition(offset, FeaturePosition{}, TemplateMirrorNone, endCityRotation(parent))
		base := endCityOrigin(parent)
		position = treeOffset(base, rotated.X, rotated.Y, rotated.Z)
	}
	piece, err := NewEndCityTemplatePiece(name, position, rotation, overwrite)
	if err != nil {
		a.err = err
		return parent
	}
	*list = append(*list, piece)
	return piece
}

func (a *endCityAssembler) recursive(kind string, depth int32, parent *StructurePiece, offset FeaturePosition, list *[]*StructurePiece) bool {
	if depth > 8 || a.err != nil {
		return false
	}
	var group []*StructurePiece
	if !a.section(kind, depth, parent, offset, &group) {
		return false
	}
	marker := a.random.NextInt()
	for _, piece := range group {
		piece.GenerationDepth = marker
		// Vanilla tests the first colliding previous piece, in insertion order.
		for _, previous := range *list {
			if endCityBoxesIntersect(previous.Box, piece.Box) {
				if previous.GenerationDepth != parent.GenerationDepth {
					return false
				}
				break
			}
		}
	}
	*list = append(*list, group...)
	return true
}

func endCityBoxesIntersect(a, b StructureBoundingBox) bool {
	return a.MaxX >= b.MinX && a.MinX <= b.MaxX && a.MaxY >= b.MinY && a.MinY <= b.MaxY && a.MaxZ >= b.MinZ && a.MinZ <= b.MaxZ
}

type endCityBridge struct {
	rotation TemplateRotation
	offset   FeaturePosition
}

var endCityTowerBridges = []endCityBridge{
	{TemplateRotationNone, FeaturePosition{1, -1, 0}},
	{TemplateRotationClockwise90, FeaturePosition{6, -1, 1}},
	{TemplateRotationCounterclockwise90, FeaturePosition{0, -1, 5}},
	{TemplateRotationClockwise180, FeaturePosition{5, -1, 6}},
}
var endCityFatBridges = []endCityBridge{
	{TemplateRotationNone, FeaturePosition{4, -1, 0}},
	{TemplateRotationClockwise90, FeaturePosition{12, -1, 4}},
	{TemplateRotationCounterclockwise90, FeaturePosition{0, -1, 8}},
	{TemplateRotationClockwise180, FeaturePosition{8, -1, 12}},
}

func (a *endCityAssembler) section(kind string, depth int32, parent *StructurePiece, offset FeaturePosition, list *[]*StructurePiece) bool {
	random := a.random
	rotation := endCityRotation(parent)
	piece := parent
	switch kind {
	case "house":
		if depth > 8 {
			return false
		}
		piece = a.add(list, parent, offset, "base_floor", rotation, true)
		switch random.NextIntBound(3) {
		case 0:
			a.add(list, piece, FeaturePosition{-1, 4, -1}, "base_roof", rotation, true)
		case 1:
			piece = a.add(list, piece, FeaturePosition{-1, 0, -1}, "second_floor_2", rotation, false)
			piece = a.add(list, piece, FeaturePosition{-1, 8, -1}, "second_roof", rotation, false)
			a.recursive("tower", depth+1, piece, FeaturePosition{}, list)
		case 2:
			piece = a.add(list, piece, FeaturePosition{-1, 0, -1}, "second_floor_2", rotation, false)
			piece = a.add(list, piece, FeaturePosition{-1, 4, -1}, "third_floor_2", rotation, false)
			piece = a.add(list, piece, FeaturePosition{-1, 8, -1}, "third_roof", rotation, true)
			a.recursive("tower", depth+1, piece, FeaturePosition{}, list)
		}
	case "tower":
		piece = a.add(list, piece, FeaturePosition{3 + random.NextIntBound(2), -3, 3 + random.NextIntBound(2)}, "tower_base", rotation, true)
		piece = a.add(list, piece, FeaturePosition{0, 7, 0}, "tower_piece", rotation, true)
		var anchor *StructurePiece
		if random.NextIntBound(3) == 0 {
			anchor = piece
		}
		count := 1 + random.NextIntBound(3)
		for index := int32(0); index < count; index++ {
			piece = a.add(list, piece, FeaturePosition{0, 4, 0}, "tower_piece", rotation, true)
			if index < count-1 && random.NextBoolean() {
				anchor = piece
			}
		}
		if anchor != nil {
			for _, bridge := range endCityTowerBridges {
				if random.NextBoolean() {
					end := a.add(list, anchor, bridge.offset, "bridge_end", TemplateRotation((uint8(rotation)+uint8(bridge.rotation))&3), true)
					a.recursive("bridge", depth+1, end, FeaturePosition{}, list)
				}
			}
			a.add(list, piece, FeaturePosition{-1, 4, -1}, "tower_top", rotation, true)
		} else if depth == 7 {
			a.add(list, piece, FeaturePosition{-1, 4, -1}, "tower_top", rotation, true)
		} else {
			return a.recursive("fat", depth+1, piece, FeaturePosition{}, list)
		}
	case "bridge":
		count := random.NextIntBound(4) + 1
		piece = a.add(list, piece, FeaturePosition{0, 0, -4}, "bridge_piece", rotation, true)
		piece.GenerationDepth = -1
		y := int32(0)
		for index := int32(0); index < count; index++ {
			if random.NextBoolean() {
				piece = a.add(list, piece, FeaturePosition{0, y, -4}, "bridge_piece", rotation, true)
				y = 0
				continue
			}
			if random.NextBoolean() {
				piece = a.add(list, piece, FeaturePosition{0, y, -4}, "bridge_steep_stairs", rotation, true)
			} else {
				piece = a.add(list, piece, FeaturePosition{0, y, -8}, "bridge_gentle_stairs", rotation, true)
			}
			y = 4
		}
		if a.shipCreated || random.NextIntBound(10-depth) != 0 {
			if !a.recursive("house", depth+1, piece, FeaturePosition{-3, y + 1, -11}, list) {
				return false
			}
		} else {
			a.add(list, piece, FeaturePosition{-8 + random.NextIntBound(8), y, -70 + random.NextIntBound(10)}, "ship", rotation, true)
			a.shipCreated = true
		}
		piece = a.add(list, piece, FeaturePosition{4, y, 0}, "bridge_end", TemplateRotation((uint8(rotation)+2)&3), true)
		piece.GenerationDepth = -1
	case "fat":
		piece = a.add(list, piece, FeaturePosition{-3, 4, -3}, "fat_tower_base", rotation, true)
		piece = a.add(list, piece, FeaturePosition{0, 4, 0}, "fat_tower_middle", rotation, true)
		for index := 0; index < 2 && random.NextIntBound(3) != 0; index++ {
			piece = a.add(list, piece, FeaturePosition{0, 8, 0}, "fat_tower_middle", rotation, true)
			for _, bridge := range endCityFatBridges {
				if random.NextBoolean() {
					end := a.add(list, piece, bridge.offset, "bridge_end", TemplateRotation((uint8(rotation)+uint8(bridge.rotation))&3), true)
					a.recursive("bridge", depth+1, end, FeaturePosition{}, list)
				}
			}
		}
		a.add(list, piece, FeaturePosition{-2, 8, -2}, "fat_tower_top", rotation, true)
	default:
		a.err = fmt.Errorf("unknown End City section%s", kind)
		return false
	}
	return a.err == nil
}

func GenerateEndCityPieces(origin FeaturePosition, rotation TemplateRotation, random *LegacyRandom) ([]StructurePiece, error) {
	if random == nil || rotation > TemplateRotationCounterclockwise90 {
		return nil, fmt.Errorf("End City needs a valid rotation and structure random")
	}
	assembler := &endCityAssembler{random: random}
	var pieces []*StructurePiece
	piece := assembler.add(&pieces, nil, origin, "base_floor", rotation, true)
	piece = assembler.add(&pieces, piece, FeaturePosition{-1, 0, -1}, "second_floor_1", rotation, false)
	piece = assembler.add(&pieces, piece, FeaturePosition{-1, 4, -1}, "third_floor_1", rotation, false)
	piece = assembler.add(&pieces, piece, FeaturePosition{-1, 8, -1}, "third_roof", rotation, true)
	assembler.recursive("tower", 1, piece, FeaturePosition{}, &pieces)
	if assembler.err != nil {
		return nil, assembler.err
	}
	result := make([]StructurePiece, len(pieces))
	for index, piece := range pieces {
		result[index] = *piece
	}
	return result, nil
}

func GenerateEndCityStructureStart(seed int64, cx, cz int32, biomeAt StructureBiomeAtQuart, heightAt StructureBaseOccupiedHeightAt) (*StructureStart, error) {
	base := NewLegacyRandom(seed)
	xm, zm := base.NextLong(), base.NextLong()
	random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	return generateEndCityStructureStart(cx, cz, biomeAt, heightAt, random)
}

func generateEndCityStructureStart(cx, cz int32, biomeAt StructureBiomeAtQuart, heightAt StructureBaseOccupiedHeightAt, random *LegacyRandom) (*StructureStart, error) {
	if biomeAt == nil || heightAt == nil {
		return nil, fmt.Errorf("End City requires biomes and base-noise heights")
	}
	rotation := TemplateRotation(random.NextIntBound(4))
	x, z := cx*16+7, cz*16+7
	dx, dz := int32(5), int32(5)
	switch rotation {
	case TemplateRotationClockwise90:
		dx = -5
	case TemplateRotationClockwise180:
		dx, dz = -5, -5
	case TemplateRotationCounterclockwise90:
		dz = -5
	}
	y := int32(2147483647)
	for _, p := range []FeaturePosition{{X: x, Z: z}, {X: x, Z: z + dz}, {X: x + dx, Z: z}, {X: x + dx, Z: z + dz}} {
		height, err := heightAt(p.X, p.Z)
		if err != nil {
			return nil, err
		}
		y = min(y, height)
	}
	if y < 60 {
		return nil, nil
	}
	biome, err := biomeAt(x>>2, y>>2, z>>2)
	if err != nil {
		return nil, err
	}
	names, err := StructureBiomeNames("minecraft:end_city")
	if err != nil {
		return nil, err
	}
	valid := false
	for _, name := range names {
		valid = valid || name == biome
	}
	if !valid {
		return nil, nil
	}
	pieces, err := GenerateEndCityPieces(FeaturePosition{x, y, z}, rotation, random)
	if err != nil {
		return nil, err
	}
	return &StructureStart{Structure: "minecraft:end_city", Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: "none", Pieces: pieces}, nil
}
