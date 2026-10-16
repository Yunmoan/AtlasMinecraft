package worldgen

import (
	"atlas/internal/nbt"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed jigsaw_data
var jigsawAssets embed.FS

type JigsawPoolElement struct {
	Type           string              `json:"element_type"`
	Location       string              `json:"location,omitempty"`
	Processors     json.RawMessage     `json:"processors,omitempty"`
	Projection     string              `json:"projection,omitempty"`
	Elements       []JigsawPoolElement `json:"elements,omitempty"`
	Feature        json.RawMessage     `json:"feature,omitempty"`
	LiquidSettings string              `json:"override_liquid_settings,omitempty"`
}

type JigsawTemplatePool struct {
	Fallback string `json:"fallback"`
	Elements []struct {
		Element JigsawPoolElement `json:"element"`
		Weight  int32             `json:"weight"`
	} `json:"elements"`
}

type jigsawRegistry struct {
	Version          int                           `json:"version"`
	Pools            map[string]JigsawTemplatePool `json:"pools"`
	Processors       map[string]json.RawMessage    `json:"processors"`
	MissingTemplates []string                      `json:"missing_templates"`
}

var loadJigsawRegistry = sync.OnceValues(func() (*jigsawRegistry, error) {
	data, err := jigsawAssets.ReadFile("jigsaw_data/registry.json")
	if err != nil {
		return nil, err
	}
	var registry jigsawRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}
	if registry.Version != 3955 || len(registry.Pools) != 186 || len(registry.Processors) != 40 {
		return nil, fmt.Errorf("incomplete original Jigsaw registry")
	}
	for name, pool := range registry.Pools {
		for index := range pool.Elements {
			pool.Elements[index].Element = normalizeJigsawElement(pool.Elements[index].Element)
		}
		registry.Pools[name] = pool
	}
	return &registry, nil
})

func normalizeJigsawElement(element JigsawPoolElement) JigsawPoolElement {
	if element.Type == "minecraft:list_pool_element" {
		for index, child := range element.Elements {
			child.Projection = element.Projection
			element.Elements[index] = normalizeJigsawElement(child)
		}
	}
	return element
}

type jigsawTemplateEntry struct {
	load func() (*StructureTemplate, error)
}

var jigsawTemplateCache sync.Map
var jigsawConnectorCache sync.Map
var jigsawRotatedConnectorCache sync.Map

type jigsawRotatedConnectorKey struct {
	template *StructureTemplate
	rotation TemplateRotation
}

type jigsawRotatedConnectorEntry struct {
	load func() ([][]jigsawConnector, error)
}

// Relative coordinates, rotated block properties and authored NBT are
// immutable. Calls copy the connector records before adding their own origin,
// shuffling and applying stable selection-priority ordering.
func jigsawRotatedConnectorPalettes(template *StructureTemplate, rotation TemplateRotation) ([][]jigsawConnector, error) {
	key := jigsawRotatedConnectorKey{template, rotation}
	if entry, found := jigsawRotatedConnectorCache.Load(key); found {
		return entry.(*jigsawRotatedConnectorEntry).load()
	}
	entry := &jigsawRotatedConnectorEntry{load: sync.OnceValues(func() ([][]jigsawConnector, error) {
		blocks := jigsawConnectorPalettes(template)
		result := make([][]jigsawConnector, len(blocks))
		for index, palette := range blocks {
			result[index] = make([]jigsawConnector, len(palette))
			for i, block := range palette {
				position, err := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
				if err != nil {
					return nil, err
				}
				state, err := TransformTemplateBlockState(block.State, TemplateMirrorNone, rotation)
				if err != nil {
					return nil, err
				}
				result[index][i] = jigsawConnector{position: position, state: state, data: block.NBT}
			}
		}
		return result, nil
	})}
	actual, _ := jigsawRotatedConnectorCache.LoadOrStore(key, entry)
	return actual.(*jigsawRotatedConnectorEntry).load()
}

func jigsawConnectorPalettes(template *StructureTemplate) [][]StructureTemplateBlock {
	if actual, ok := jigsawConnectorCache.Load(template); ok {
		return actual.([][]StructureTemplateBlock)
	}
	palettes := make([][]StructureTemplateBlock, len(template.Palettes))
	for index, palette := range template.Palettes {
		for _, block := range palette {
			if block.State.Name == "minecraft:jigsaw" {
				palettes[index] = append(palettes[index], block)
			}
		}
	}
	actual, _ := jigsawConnectorCache.LoadOrStore(template, palettes)
	return actual.([][]StructureTemplateBlock)
}

// LoadJigsawTemplate returns an immutable original template, including the
// original empty-template fallback for explicitly absent built-in resources.
func LoadJigsawTemplate(name string) (*StructureTemplate, error) {
	name = densityName(name)
	if actual, found := jigsawTemplateCache.Load(name); found {
		return actual.(*jigsawTemplateEntry).load()
	}
	entry := &jigsawTemplateEntry{load: sync.OnceValues(func() (*StructureTemplate, error) {
		registry, err := loadJigsawRegistry()
		if err != nil {
			return nil, err
		}
		for _, missing := range registry.MissingTemplates {
			if name == missing {
				return &StructureTemplate{}, nil
			}
		}
		data, err := jigsawAssets.ReadFile("jigsaw_data/templates/" + strings.TrimPrefix(name, "minecraft:") + ".nbt")
		if err != nil {
			return nil, err
		}
		return DecodeStructureTemplate(data)
	})}
	actual, _ := jigsawTemplateCache.LoadOrStore(name, entry)
	return actual.(*jigsawTemplateEntry).load()
}

func jigsawJSONTag(value any) (nbt.Tag, error) {
	switch value := value.(type) {
	case string:
		return nbt.Tag{Type: nbt.String, Value: value}, nil
	case float64:
		return nbt.Tag{Type: nbt.Int, Value: int32(value)}, nil
	case bool:
		var flag int8
		if value {
			flag = 1
		}
		return nbt.Tag{Type: nbt.Byte, Value: flag}, nil
	case map[string]any:
		object := make(nbt.Object, len(value))
		for key, item := range value {
			tag, err := jigsawJSONTag(item)
			if err != nil {
				return nbt.Tag{}, err
			}
			object[key] = tag
		}
		return nbt.Tag{Type: nbt.Compound, Value: object}, nil
	case []any:
		list := make([]nbt.Tag, 0, len(value))
		var kind byte
		for _, item := range value {
			tag, err := jigsawJSONTag(item)
			if err != nil {
				return nbt.Tag{}, err
			}
			if kind != 0 && kind != tag.Type {
				return nbt.Tag{}, fmt.Errorf("mixed Jigsaw codec list")
			}
			kind = tag.Type
			list = append(list, tag)
		}
		return nbt.Tag{Type: nbt.List, ListType: kind, Value: list}, nil
	default:
		return nbt.Tag{}, fmt.Errorf("unsupported Jigsaw codec value %T", value)
	}
}

func (element JigsawPoolElement) NBT() (nbt.Object, error) {
	data, err := json.Marshal(element)
	if err != nil {
		return nil, err
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	tag, err := jigsawJSONTag(value)
	if err != nil {
		return nil, err
	}
	return tag.Value.(nbt.Object), nil
}

func JigsawProcessorJSON(name string) ([]byte, error) {
	registry, err := loadJigsawRegistry()
	if err != nil {
		return nil, err
	}
	data := registry.Processors[densityName(name)]
	if data == nil {
		return nil, fmt.Errorf("unknown Jigsaw processor list %q", name)
	}
	return append([]byte(nil), data...), nil
}
