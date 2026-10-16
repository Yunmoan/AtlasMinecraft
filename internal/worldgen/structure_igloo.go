package worldgen

import (
	"atlas/internal/nbt"
	"embed"
	"fmt"
	"strings"
	"sync"
)

var iglooPivots = map[string]FeaturePosition{"minecraft:igloo/top": {3, 5, 5}, "minecraft:igloo/middle": {1, 3, 1}, "minecraft:igloo/bottom": {3, 6, 7}}
var iglooOffsets = map[string]FeaturePosition{"minecraft:igloo/top": {0, 0, 0}, "minecraft:igloo/middle": {2, -3, 4}, "minecraft:igloo/bottom": {0, -3, -2}}
var templateRotationNames = [...]string{"NONE", "CLOCKWISE_90", "CLOCKWISE_180", "COUNTERCLOCKWISE_90"}

//go:embed igloo_data/*.nbt
var iglooAssets embed.FS

var loadIglooTemplates = sync.OnceValues(func() (map[string]*StructureTemplate, error) {
	result := make(map[string]*StructureTemplate)
	for _, name := range []string{"top", "middle", "bottom"} {
		data, err := iglooAssets.ReadFile("igloo_data/" + name + ".nbt")
		if err != nil {
			return nil, err
		}
		template, err := DecodeStructureTemplate(data)
		if err != nil {
			return nil, fmt.Errorf("igloo %s: %w", name, err)
		}
		result["minecraft:igloo/"+name] = template
	}
	return result, nil
})

func LoadIglooTemplate(name string) (*StructureTemplate, error) {
	if !strings.Contains(name, ":") {
		name = "minecraft:igloo/" + strings.TrimPrefix(name, "igloo/")
	}
	templates, err := loadIglooTemplates()
	if err != nil {
		return nil, err
	}
	template := templates[name]
	if template == nil {
		return nil, fmt.Errorf("unsupported igloo template %q", name)
	}
	return template, nil
}

// GenerateIglooStructureStart reproduces the original random rotation,
// 50-percent laboratory, ladder count and saved template piece positions.
// A biome-rejected candidate returns nil. Block placement is a later stage.
func GenerateIglooStructureStart(seed int64, chunkX, chunkZ int32, biomeAtQuart StructureBiomeAtQuart,
	occupiedHeightAt StructureBaseOccupiedHeightAt) (*StructureStart, error) {
	if biomeAtQuart == nil || occupiedHeightAt == nil {
		return nil, fmt.Errorf("igloo start requires base height and biome sources")
	}
	x, z := chunkX*16+8, chunkZ*16+8
	y, err := occupiedHeightAt(x, z)
	if err != nil {
		return nil, err
	}
	biome, err := biomeAtQuart(x>>2, y>>2, z>>2)
	if err != nil {
		return nil, err
	}
	allowed, err := StructureBiomeNames("minecraft:igloo")
	if err != nil {
		return nil, err
	}
	valid := false
	for _, name := range allowed {
		valid = valid || biome == name
	}
	if !valid {
		return nil, nil
	}
	random := NewLegacyRandom(seed)
	xMultiplier, zMultiplier := random.NextLong(), random.NextLong()
	random = NewLegacyRandom(int64(uint64(int64(chunkX))*xMultiplier ^ uint64(int64(chunkZ))*zMultiplier ^ uint64(seed)))
	rotation := TemplateRotation(random.NextIntBound(4))
	start := &StructureStart{Structure: "minecraft:igloo", Chunk: StructureChunkPosition{chunkX, chunkZ}, TerrainAdjustment: "none"}
	appendPiece := func(name string, depth int32) error {
		template, err := LoadIglooTemplate(name)
		if err != nil {
			return err
		}
		offset := iglooOffsets[name]
		origin := FeaturePosition{chunkX*16 + offset.X, 90 + offset.Y - depth, chunkZ*16 + offset.Z}
		box, err := template.BoundingBox(origin, iglooPivots[name], TemplateMirrorNone, rotation)
		if err != nil {
			return err
		}
		start.Pieces = append(start.Pieces, StructurePiece{ID: "minecraft:iglu", Orientation: "north", Box: box, Data: nbt.Object{
			"Template": {Type: nbt.String, Value: name}, "Rot": {Type: nbt.String, Value: templateRotationNames[rotation]},
			"TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z},
		}})
		return nil
	}
	if random.NextDouble() < 0.5 {
		segments := random.NextIntBound(8) + 4
		if err := appendPiece("minecraft:igloo/bottom", segments*3); err != nil {
			return nil, err
		}
		for index := int32(0); index < segments-1; index++ {
			if err := appendPiece("minecraft:igloo/middle", index*3); err != nil {
				return nil, err
			}
		}
	}
	if err := appendPiece("minecraft:igloo/top", 0); err != nil {
		return nil, err
	}
	return start, nil
}
