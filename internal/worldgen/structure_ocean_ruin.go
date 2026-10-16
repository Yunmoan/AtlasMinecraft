package worldgen

import (
	"atlas/internal/nbt"
	"embed"
	"fmt"
	"strings"
	"sync"
)

//go:embed ocean_ruin_data/*.nbt
var oceanRuinAssets embed.FS

var loadOceanRuinTemplates = sync.OnceValues(func() (map[string]*StructureTemplate, error) {
	entries, err := oceanRuinAssets.ReadDir("ocean_ruin_data")
	if err != nil {
		return nil, err
	}
	result := make(map[string]*StructureTemplate)
	for _, entry := range entries {
		data, err := oceanRuinAssets.ReadFile("ocean_ruin_data/" + entry.Name())
		if err != nil {
			return nil, err
		}
		template, err := DecodeStructureTemplate(data)
		if err != nil {
			return nil, fmt.Errorf("ocean ruin %s: %w", entry.Name(), err)
		}
		result["minecraft:underwater_ruin/"+strings.TrimSuffix(entry.Name(), ".nbt")] = template
	}
	return result, nil
})

func LoadOceanRuinTemplate(name string) (*StructureTemplate, error) {
	if !strings.Contains(name, ":") {
		name = "minecraft:underwater_ruin/" + strings.TrimPrefix(name, "underwater_ruin/")
	}
	templates, err := loadOceanRuinTemplates()
	if err != nil {
		return nil, err
	}
	template := templates[name]
	if template == nil {
		return nil, fmt.Errorf("unsupported ocean ruin template %q", name)
	}
	return template, nil
}

func NewOceanRuinTemplatePiece(name string, origin FeaturePosition, rotation TemplateRotation, integrity float32, cold, large bool) (*StructurePiece, error) {
	if rotation > TemplateRotationCounterclockwise90 || integrity < 0 || integrity > 1 {
		return nil, fmt.Errorf("invalid ocean ruin rotation/integrity")
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:underwater_ruin/" + strings.TrimPrefix(name, "underwater_ruin/")
	}
	template, err := LoadOceanRuinTemplate(name)
	if err != nil {
		return nil, err
	}
	box, err := template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	if err != nil {
		return nil, err
	}
	biomeType := "WARM"
	if cold {
		biomeType = "COLD"
	}
	bit := int8(0)
	if large {
		bit = 1
	}
	return &StructurePiece{ID: "minecraft:orp", Orientation: "north", Box: box, Data: nbt.Object{"Template": {Type: nbt.String, Value: name}, "TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z}, "Rot": {Type: nbt.String, Value: templateRotationNames[rotation]}, "Integrity": {Type: nbt.Float, Value: integrity}, "BiomeType": {Type: nbt.String, Value: biomeType}, "IsLarge": {Type: nbt.Byte, Value: bit}}}, nil
}

func GenerateOceanRuinStructureStart(seed int64, cx, cz int32, cold bool, biomeAt StructureBiomeAtQuart, oceanFloorHeight StructureBaseOccupiedHeightAt) (*StructureStart, error) {
	if biomeAt == nil || oceanFloorHeight == nil {
		return nil, fmt.Errorf("ocean ruin requires biomes/base ocean floor height")
	}
	height, err := oceanFloorHeight(cx*16+8, cz*16+8)
	if err != nil {
		return nil, err
	}
	biome, err := biomeAt((cx*16+8)>>2, height>>2, (cz*16+8)>>2)
	if err != nil {
		return nil, err
	}
	name := "minecraft:ocean_ruin_warm"
	if cold {
		name = "minecraft:ocean_ruin_cold"
	}
	allowed, err := StructureBiomeNames(name)
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
	base := NewLegacyRandom(seed)
	xm, zm := base.NextLong(), base.NextLong()
	random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	rotation := TemplateRotation(random.NextIntBound(4))
	origin := FeaturePosition{cx * 16, 90, cz * 16}
	var pieces []StructurePiece
	appendRuin := func(origin FeaturePosition, rotation TemplateRotation, large bool, integrity float32) error {
		if !cold {
			count, offset, prefix := int32(8), int32(1), "warm_"
			if large {
				count, offset, prefix = 4, 4, "big_warm_"
			}
			index := random.NextIntBound(count) + offset
			piece, err := NewOceanRuinTemplatePiece(fmt.Sprintf("%s%d", prefix, index), origin, rotation, integrity, cold, large)
			if err != nil {
				return err
			}
			pieces = append(pieces, *piece)
			return nil
		}
		count := int32(8)
		if large {
			count = 4
		}
		index := random.NextIntBound(count) + 1
		if large && index == 4 {
			index = 8
		}
		prefix := ""
		if large {
			prefix = "big_"
		}
		for position, material := range []string{"brick", "cracked", "mossy"} {
			value := integrity
			if position == 1 {
				value = 0.7
			} else if position == 2 {
				value = 0.5
			}
			piece, err := NewOceanRuinTemplatePiece(fmt.Sprintf("%s%s_%d", prefix, material, index), origin, rotation, value, cold, large)
			if err != nil {
				return err
			}
			pieces = append(pieces, *piece)
		}
		return nil
	}
	large := random.NextFloat() <= 0.3
	integrity := float32(0.8)
	if large {
		integrity = 0.9
	}
	if err := appendRuin(origin, rotation, large, integrity); err != nil {
		return nil, err
	}
	if large && random.NextFloat() <= 0.9 {
		corner, _ := TransformTemplatePosition(FeaturePosition{15, 0, 15}, FeaturePosition{}, TemplateMirrorNone, rotation)
		corner = treeOffset(origin, corner.X, corner.Y, corner.Z)
		box := StructureBoundingBox{min(origin.X, corner.X), 90, min(origin.Z, corner.Z), max(origin.X, corner.X), 90, max(origin.Z, corner.Z)}
		minimum := FeaturePosition{box.MinX, 90, box.MinZ}
		between := func(low, high int32) int32 { return low + random.NextIntBound(high-low+1) }
		positions := []FeaturePosition{
			treeOffset(minimum, -16+between(1, 8), 0, 16+between(1, 7)), treeOffset(minimum, -16+between(1, 8), 0, between(1, 7)), treeOffset(minimum, -16+between(1, 8), 0, -16+between(4, 8)), treeOffset(minimum, between(1, 7), 0, 16+between(1, 7)), treeOffset(minimum, between(1, 7), 0, -16+between(4, 6)), treeOffset(minimum, 16+between(1, 7), 0, 16+between(3, 8)), treeOffset(minimum, 16+between(1, 7), 0, between(1, 7)), treeOffset(minimum, 16+between(1, 7), 0, -16+between(4, 8)),
		}
		count := between(4, 8)
		for i := int32(0); i < count && len(positions) > 0; i++ {
			index := random.NextIntBound(int32(len(positions)))
			position := positions[index]
			positions = append(positions[:index], positions[index+1:]...)
			rotation := TemplateRotation(random.NextIntBound(4))
			corner, _ := TransformTemplatePosition(FeaturePosition{5, 0, 6}, FeaturePosition{}, TemplateMirrorNone, rotation)
			corner = treeOffset(position, corner.X, corner.Y, corner.Z)
			candidate := StructureBoundingBox{min(position.X, corner.X), 90, min(position.Z, corner.Z), max(position.X, corner.X), 90, max(position.Z, corner.Z)}
			if candidate.MaxX >= box.MinX && candidate.MinX <= box.MaxX && candidate.MaxZ >= box.MinZ && candidate.MinZ <= box.MaxZ {
				continue
			}
			if err := appendRuin(position, rotation, false, 0.8); err != nil {
				return nil, err
			}
		}
	}
	return &StructureStart{Structure: name, Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: "none", Pieces: pieces}, nil
}
