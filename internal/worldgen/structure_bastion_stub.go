package worldgen

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed bastion_start_data
var bastionStartAssets embed.FS

type bastionStartElement struct {
	Template string
	Weight   int32
	Size     FeaturePosition
}

var loadBastionStartElements = sync.OnceValues(func() ([]bastionStartElement, error) {
	data, err := bastionStartAssets.ReadFile("bastion_start_data/starts.json")
	if err != nil {
		return nil, err
	}
	var pool struct {
		Elements []struct {
			Element struct {
				Type       string `json:"element_type"`
				Location   string `json:"location"`
				Projection string `json:"projection"`
			} `json:"element"`
			Weight int32 `json:"weight"`
		} `json:"elements"`
	}
	if err := json.Unmarshal(data, &pool); err != nil {
		return nil, err
	}
	if len(pool.Elements) != 4 {
		return nil, fmt.Errorf("expected four original Bastion start elements")
	}
	result := make([]bastionStartElement, 0, len(pool.Elements))
	for _, entry := range pool.Elements {
		if entry.Element.Type != "minecraft:single_pool_element" || entry.Element.Projection != "rigid" || entry.Weight <= 0 {
			return nil, fmt.Errorf("unsupported Bastion start pool element")
		}
		data, err := bastionStartAssets.ReadFile("bastion_start_data/" + strings.TrimPrefix(entry.Element.Location, "minecraft:") + ".nbt")
		if err != nil {
			return nil, err
		}
		template, err := DecodeStructureTemplate(data)
		if err != nil {
			return nil, err
		}
		result = append(result, bastionStartElement{Template: entry.Element.Location, Weight: entry.Weight, Size: template.Size})
	}
	return result, nil
})

// BastionGenerationStub is the original pre-assembly generation point. It
// deliberately carries no StructureStart: pool expansion and blocks must be
// implemented before a Bastion may be persisted as a generated structure.
type BastionGenerationStub struct {
	Template string
	Rotation TemplateRotation
	Origin   FeaturePosition
	Box      StructureBoundingBox
	Position FeaturePosition
}

func findBastionGenerationStub(seed int64, cx, cz int32) (*BastionGenerationStub, error) {
	elements, err := loadBastionStartElements()
	if err != nil {
		return nil, err
	}
	base := NewLegacyRandom(seed)
	xm, zm := base.NextLong(), base.NextLong()
	random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	rotation := TemplateRotation(random.NextIntBound(4))
	var total int32
	for _, element := range elements {
		total += element.Weight
	}
	index := random.NextIntBound(total)
	selected := elements[len(elements)-1]
	for _, element := range elements {
		index -= element.Weight
		if index < 0 {
			selected = element
			break
		}
	}
	// The single-pool element has groundLevelDelta=1. Its initial base is Y=33,
	// then JigsawPlacement moves the piece down one block before biome lookup.
	origin := FeaturePosition{cx * 16, 32, cz * 16}
	template := StructureTemplate{Size: selected.Size}
	box, err := template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	if err != nil {
		return nil, err
	}
	return &BastionGenerationStub{
		Template: selected.Template, Rotation: rotation, Origin: origin, Box: box,
		// Java integer division truncates toward zero, including negative centers.
		Position: FeaturePosition{(box.MinX + box.MaxX) / 2, 33, (box.MinZ + box.MaxZ) / 2},
	}, nil
}

// GenerateBastionGenerationStub reproduces Bastion's rotation, weighted start
// template, transformed center and biome gate without assembling a Bastion.
// A nil result is a genuine biome rejection, so the shared nether_complexes
// set can retry Fortress only when the original Bastion would reject it.
func GenerateBastionGenerationStub(seed int64, cx, cz int32, biomeAt StructureBiomeAtQuart) (*BastionGenerationStub, error) {
	if biomeAt == nil {
		return nil, fmt.Errorf("Bastion stub requires a biome source")
	}
	stub, err := findBastionGenerationStub(seed, cx, cz)
	if err != nil {
		return nil, err
	}
	biome, err := biomeAt(stub.Position.X>>2, stub.Position.Y>>2, stub.Position.Z>>2)
	if err != nil {
		return nil, err
	}
	allowed, err := StructureBiomeNames("minecraft:bastion_remnant")
	if err != nil {
		return nil, err
	}
	for _, name := range allowed {
		if biome == name {
			return stub, nil
		}
	}
	return nil, nil
}
