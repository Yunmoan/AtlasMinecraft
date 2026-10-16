package worldgen

import (
	"atlas/internal/nbt"
	"embed"
	"fmt"
	"strings"
	"sync"
)

//go:embed shipwreck_data/*.nbt
var shipwreckAssets embed.FS

var shipwreckOceanNames = []string{"with_mast", "upsidedown_full", "upsidedown_fronthalf", "upsidedown_backhalf", "sideways_full", "sideways_fronthalf", "sideways_backhalf", "rightsideup_full", "rightsideup_fronthalf", "rightsideup_backhalf", "with_mast_degraded", "upsidedown_full_degraded", "upsidedown_fronthalf_degraded", "upsidedown_backhalf_degraded", "sideways_full_degraded", "sideways_fronthalf_degraded", "sideways_backhalf_degraded", "rightsideup_full_degraded", "rightsideup_fronthalf_degraded", "rightsideup_backhalf_degraded"}
var shipwreckBeachedNames = []string{"with_mast", "sideways_full", "sideways_fronthalf", "sideways_backhalf", "rightsideup_full", "rightsideup_fronthalf", "rightsideup_backhalf", "with_mast_degraded", "rightsideup_full_degraded", "rightsideup_fronthalf_degraded", "rightsideup_backhalf_degraded"}
var shipwreckPivot = FeaturePosition{4, 0, 15}

var loadShipwreckTemplates = sync.OnceValues(func() (map[string]*StructureTemplate, error) {
	result := make(map[string]*StructureTemplate)
	for _, name := range shipwreckOceanNames {
		data, err := shipwreckAssets.ReadFile("shipwreck_data/" + name + ".nbt")
		if err != nil {
			return nil, err
		}
		template, err := DecodeStructureTemplate(data)
		if err != nil {
			return nil, fmt.Errorf("shipwreck %s: %w", name, err)
		}
		result["minecraft:shipwreck/"+name] = template
	}
	return result, nil
})

func LoadShipwreckTemplate(name string) (*StructureTemplate, error) {
	if !strings.Contains(name, ":") {
		name = "minecraft:shipwreck/" + strings.TrimPrefix(name, "shipwreck/")
	}
	templates, err := loadShipwreckTemplates()
	if err != nil {
		return nil, err
	}
	template := templates[name]
	if template == nil {
		return nil, fmt.Errorf("unsupported shipwreck template %q", name)
	}
	return template, nil
}

func NewShipwreckTemplatePiece(name string, origin FeaturePosition, rotation TemplateRotation, beached bool) (*StructurePiece, error) {
	if rotation > TemplateRotationCounterclockwise90 {
		return nil, fmt.Errorf("invalid shipwreck rotation")
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:shipwreck/" + strings.TrimPrefix(name, "shipwreck/")
	}
	template, err := LoadShipwreckTemplate(name)
	if err != nil {
		return nil, err
	}
	box, err := template.BoundingBox(origin, shipwreckPivot, TemplateMirrorNone, rotation)
	if err != nil {
		return nil, err
	}
	bit := int8(0)
	if beached {
		bit = 1
	}
	return &StructurePiece{ID: "minecraft:shipwreck", Orientation: "north", Box: box, Data: nbt.Object{"Template": {Type: nbt.String, Value: name}, "TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z}, "Rot": {Type: nbt.String, Value: templateRotationNames[rotation]}, "isBeached": {Type: nbt.Byte, Value: bit}}}, nil
}

// GenerateShipwreckStructureStart creates original ocean/beached starts. All
// original templates use the later piece height adjustment, with initial Y=90.
func GenerateShipwreckStructureStart(seed int64, chunkX, chunkZ int32, beached bool, biomeAt StructureBiomeAtQuart, baseHeight StructureBaseOccupiedHeightAt) (*StructureStart, error) {
	if biomeAt == nil || baseHeight == nil {
		return nil, fmt.Errorf("shipwreck start requires biomes and base occupied heights")
	}
	height, err := baseHeight(chunkX*16+8, chunkZ*16+8)
	if err != nil {
		return nil, err
	}
	biome, err := biomeAt((chunkX*16+8)>>2, height>>2, (chunkZ*16+8)>>2)
	if err != nil {
		return nil, err
	}
	structure := "minecraft:shipwreck"
	names := shipwreckOceanNames
	if beached {
		structure = "minecraft:shipwreck_beached"
		names = shipwreckBeachedNames
	}
	allowed, err := StructureBiomeNames(structure)
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
	random := NewLegacyRandom(int64(uint64(int64(chunkX))*xm ^ uint64(int64(chunkZ))*zm ^ uint64(seed)))
	rotation := TemplateRotation(random.NextIntBound(4))
	name := names[random.NextIntBound(int32(len(names)))]
	piece, err := NewShipwreckTemplatePiece(name, FeaturePosition{chunkX * 16, 90, chunkZ * 16}, rotation, beached)
	if err != nil {
		return nil, err
	}
	template, err := LoadShipwreckTemplate(name)
	if err != nil {
		return nil, err
	}
	if template.Size.X > 32 || template.Size.Y > 32 {
		return nil, fmt.Errorf("shipwreck template requires large start-height handling")
	}
	return &StructureStart{Structure: structure, Chunk: StructureChunkPosition{chunkX, chunkZ}, TerrainAdjustment: "none", Pieces: []StructurePiece{*piece}}, nil
}
