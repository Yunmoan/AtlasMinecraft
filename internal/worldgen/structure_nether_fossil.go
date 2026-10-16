package worldgen

import (
	"atlas/internal/nbt"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed nether_fossil_data/*.nbt
var netherFossilAssets embed.FS

var loadNetherFossilTemplates = sync.OnceValues(func() (map[string]*StructureTemplate, error) {
	templates := make(map[string]*StructureTemplate, 14)
	for index := 1; index <= 14; index++ {
		name := fmt.Sprintf("fossil_%d", index)
		data, err := netherFossilAssets.ReadFile("nether_fossil_data/" + name + ".nbt")
		if err != nil {
			return nil, err
		}
		template, err := DecodeStructureTemplate(data)
		if err != nil {
			return nil, err
		}
		if len(template.Entities) != 0 {
			return nil, fmt.Errorf("unexpected Nether fossil template actors")
		}
		templates["minecraft:nether_fossils/"+name] = template
	}
	return templates, nil
})

func LoadNetherFossilTemplate(name string) (*StructureTemplate, error) {
	if !strings.Contains(name, ":") {
		name = "minecraft:nether_fossils/" + strings.TrimPrefix(name, "nether_fossils/")
	}
	templates, err := loadNetherFossilTemplates()
	if err != nil {
		return nil, err
	}
	if template := templates[name]; template != nil {
		return template, nil
	}
	return nil, fmt.Errorf("unknown Nether fossil template %q", name)
}

func NewNetherFossilTemplatePiece(name string, origin FeaturePosition, rotation TemplateRotation) (*StructurePiece, error) {
	if rotation > TemplateRotationCounterclockwise90 {
		return nil, fmt.Errorf("invalid Nether fossil rotation")
	}
	template, err := LoadNetherFossilTemplate(name)
	if err != nil {
		return nil, err
	}
	box, err := template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:nether_fossils/" + strings.TrimPrefix(name, "nether_fossils/")
	}
	return &StructurePiece{ID: "minecraft:nefos", Orientation: "north", Box: box, Data: nbt.Object{
		"Template": {Type: nbt.String, Value: name}, "Rot": {Type: nbt.String, Value: templateRotationNames[rotation]},
		"TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z},
	}}, nil
}

type NetherFossilStartSources struct {
	MinY, GenerationDepth, SeaLevel int32
	BiomeAtQuart                    StructureBiomeAtQuart
	BaseColumnAt                    StructureBaseNoiseColumnAt
}

// The original start descends a base-noise column looking for air immediately
// above a supporting block, rejects a result at/below sea level, then checks
// the biome before consuming the rotation and template-selection draws.
func GenerateNetherFossilStructureStart(seed int64, cx, cz int32, sources NetherFossilStartSources) (*StructureStart, error) {
	if sources.GenerationDepth <= 0 || sources.BiomeAtQuart == nil || sources.BaseColumnAt == nil {
		return nil, fmt.Errorf("Nether fossil start requires generation bounds, biomes and columns")
	}
	definition, err := BuiltinStructureJSON("minecraft:nether_fossil")
	if err != nil {
		return nil, err
	}
	var settings struct {
		Height struct {
			Type string       `json:"type"`
			Min  heightAnchor `json:"min_inclusive"`
			Max  heightAnchor `json:"max_inclusive"`
		} `json:"height"`
	}
	if err := json.Unmarshal(definition, &settings); err != nil {
		return nil, err
	}
	if settings.Height.Type != "minecraft:uniform" {
		return nil, fmt.Errorf("unsupported Nether fossil height provider")
	}
	low, err := settings.Height.Min.resolve(sources.MinY, sources.GenerationDepth)
	if err != nil {
		return nil, err
	}
	high, err := settings.Height.Max.resolve(sources.MinY, sources.GenerationDepth)
	if err != nil {
		return nil, err
	}
	random := NewLegacyRandom(seed)
	xm, zm := random.NextLong(), random.NextLong()
	random = NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	x, z := cx*16+random.NextIntBound(16), cz*16+random.NextIntBound(16)
	y := low
	if high >= low {
		y += random.NextIntBound(high - low + 1)
	}
	column, err := sources.BaseColumnAt(x, z)
	if err != nil {
		return nil, err
	}
	if column == nil {
		return nil, fmt.Errorf("nil Nether fossil base column")
	}
	for y > sources.SeaLevel {
		above := column.MaterialAt(y)
		y--
		below := column.MaterialAt(y)
		if above == NoiseAir && below != NoiseAir && below != NoiseWater && below != NoiseLava {
			break
		}
	}
	if y <= sources.SeaLevel {
		return nil, nil
	}
	biome, err := sources.BiomeAtQuart(x>>2, y>>2, z>>2)
	if err != nil {
		return nil, err
	}
	if biome != "minecraft:soul_sand_valley" {
		return nil, nil
	}
	rotation := TemplateRotation(random.NextIntBound(4))
	name := fmt.Sprintf("minecraft:nether_fossils/fossil_%d", random.NextIntBound(14)+1)
	piece, err := NewNetherFossilTemplatePiece(name, FeaturePosition{x, y, z}, rotation)
	if err != nil {
		return nil, err
	}
	return &StructureStart{Structure: "minecraft:nether_fossil", Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: "beard_thin", Pieces: []StructurePiece{*piece}}, nil
}

// NetherFossilPiece expands the caller's clip to its entire template, so a
// single intersecting invocation places every bone block. Air and structure
// markers are ignored; the official assets contain no block entities.
func PostProcessNetherFossilBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, cx, cz int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:nefos" || random == nil {
		return result, fmt.Errorf("Nether fossil requires saved piece, world and random")
	}
	name, _ := piece.Data.Text("Template")
	rotationName, _ := piece.Data.Text("Rot")
	rotation := TemplateRotation(255)
	for index, name := range templateRotationNames {
		if name == rotationName {
			rotation = TemplateRotation(index)
		}
	}
	if rotation > TemplateRotationCounterclockwise90 {
		return result, fmt.Errorf("invalid saved Nether fossil rotation")
	}
	var origin FeaturePosition
	for index, key := range []string{"TPX", "TPY", "TPZ"} {
		value, ok := piece.Data.Number(key)
		if !ok {
			return result, fmt.Errorf("Nether fossil lacks %s", key)
		}
		switch index {
		case 0:
			origin.X = int32(value)
		case 1:
			origin.Y = int32(value)
		case 2:
			origin.Z = int32(value)
		}
	}
	template, err := LoadNetherFossilTemplate(name)
	if err != nil {
		return result, err
	}
	piece.Box, err = template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	if err != nil {
		return result, err
	}
	positionRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	palette := template.Palettes[positionRandom.NextIntBound(int32(len(template.Palettes)))]
	var placed []FeaturePosition
	for _, block := range palette {
		if surfaceAir(block.State) || block.State.Name == "minecraft:structure_block" {
			continue
		}
		if block.NBT != nil {
			return result, fmt.Errorf("unexpected Nether fossil block entity")
		}
		p, err := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
		if err != nil {
			return result, err
		}
		p = treeOffset(origin, p.X, p.Y, p.Z)
		state, err := TransformTemplateBlockState(block.State, TemplateMirrorNone, rotation)
		if err != nil {
			return result, err
		}
		if err := world.SetBlock(p, state); err != nil {
			return result, err
		}
		placed = append(placed, p)
	}
	return result, finalizeIglooTemplateShape(world, placed)
}
