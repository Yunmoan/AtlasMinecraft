package worldgen

import (
	"atlas/internal/nbt"
	"embed"
	"fmt"
	"strings"
	"sync"
)

//go:embed mansion_data/*.nbt
var mansionAssets embed.FS

var loadMansionTemplates = sync.OnceValues(func() (map[string]*StructureTemplate, error) {
	files, err := mansionAssets.ReadDir("mansion_data")
	if err != nil {
		return nil, err
	}
	result := make(map[string]*StructureTemplate, len(files))
	for _, file := range files {
		data, err := mansionAssets.ReadFile("mansion_data/" + file.Name())
		if err != nil {
			return nil, err
		}
		template, err := DecodeStructureTemplate(data)
		if err != nil {
			return nil, err
		}
		result[strings.TrimSuffix(file.Name(), ".nbt")] = template
	}
	if len(result) != 73 {
		return nil, fmt.Errorf("expected 73 original mansion templates")
	}
	return result, nil
})

func LoadMansionTemplate(name string) (*StructureTemplate, error) {
	name = strings.TrimPrefix(strings.TrimPrefix(name, "minecraft:"), "woodland_mansion/")
	templates, err := loadMansionTemplates()
	if err != nil {
		return nil, err
	}
	template := templates[name]
	if template == nil {
		return nil, fmt.Errorf("unknown mansion template %q", name)
	}
	return template, nil
}

func NewMansionTemplatePiece(name string, origin FeaturePosition, rotation TemplateRotation, mirror TemplateMirror) (*StructurePiece, error) {
	template, err := LoadMansionTemplate(name)
	if err != nil {
		return nil, err
	}
	box, err := template.BoundingBox(origin, FeaturePosition{}, mirror, rotation)
	if err != nil {
		return nil, err
	}
	return &StructurePiece{ID: "minecraft:wmp", Box: box, Orientation: "north", Data: nbt.Object{
		"Template": {Type: nbt.String, Value: name}, "Rot": {Type: nbt.String, Value: templateRotationNames[rotation]}, "Mi": {Type: nbt.String, Value: []string{"NONE", "LEFT_RIGHT", "FRONT_BACK"}[mirror]},
		"TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z},
	}}, nil
}
