package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sort"
)

// StructureTemplate keeps original palette-specific placement order, block
// entities and entities. A placer must not silently discard unsupported NBT.
type StructureTemplate struct {
	Size     FeaturePosition
	Palettes [][]StructureTemplateBlock
	Entities []StructureTemplateEntity
}

type StructureTemplateBlock struct {
	Position FeaturePosition
	State    SurfaceBlockState
	NBT      nbt.Object
}

type StructureTemplateEntity struct {
	Position      [3]float64
	BlockPosition FeaturePosition
	NBT           nbt.Object
}

// DecodeStructureTemplate loads the original compressed NBT. Only current
// 1.21.1 assets are accepted; upgrading old templates requires an explicit
// data-fix port rather than pretending their states already have current IDs.
func DecodeStructureTemplate(compressed []byte) (*StructureTemplate, error) {
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	data, err := io.ReadAll(io.LimitReader(gz, (32<<20)+1))
	if err != nil || len(data) > 32<<20 {
		return nil, fmt.Errorf("invalid compressed structure template: %v", err)
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		return nil, err
	}
	object, _ := root.Object()
	version, found := object.Number("DataVersion")
	if !found || version != 3955 {
		return nil, fmt.Errorf("structure template requires DataVersion 3955, got %d", version)
	}
	size, err := fossilVector(object["size"])
	if err != nil || size.X < 1 || size.Y < 1 || size.Z < 1 || size.X > 256 || size.Y > 384 || size.Z > 256 {
		return nil, fmt.Errorf("invalid template size %v", size)
	}
	result := &StructureTemplate{Size: size}
	var palettes []nbt.Tag
	if multiple, exists := object["palettes"]; exists {
		var ok bool
		palettes, ok = multiple.Elements()
		if !ok || len(palettes) == 0 {
			return nil, fmt.Errorf("invalid multiple structure palettes")
		}
	} else {
		palettes = []nbt.Tag{object["palette"]}
	}
	blocks, ok := object["blocks"].Elements()
	if !ok {
		return nil, fmt.Errorf("missing template block list")
	}
	// Positions, palette indices and block NBT are shared by all palettes.
	// Decode them once, then allocate one exact-size ordered block array for
	// each palette instead of growing three groups and a second merged array.
	type sourceBlock struct {
		position FeaturePosition
		state    int64
		data     nbt.Object
	}
	sources := make([]sourceBlock, len(blocks))
	for index, tag := range blocks {
		entry, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("invalid structure block entry")
		}
		position, err := fossilVector(entry["pos"])
		state, ok := entry.Number("state")
		if err != nil || !ok || state < 0 {
			return nil, fmt.Errorf("invalid structure block position/state")
		}
		sources[index] = sourceBlock{position: position, state: state}
		if data, exists := entry["nbt"]; exists {
			sources[index].data, ok = data.Object()
			if !ok {
				return nil, fmt.Errorf("invalid template block entity NBT")
			}
		}
	}
	for _, paletteTag := range palettes {
		palette, ok := paletteTag.Elements()
		if !ok || len(palette) == 0 {
			return nil, fmt.Errorf("missing structure palette states")
		}
		states := make([]SurfaceBlockState, len(palette))
		full := make([]bool, len(palette))
		for index, item := range palette {
			entry, ok := item.Object()
			if !ok {
				return nil, fmt.Errorf("invalid structure palette entry")
			}
			states[index].Name, ok = entry.Text("Name")
			if !ok {
				return nil, fmt.Errorf("missing structure state name")
			}
			if propertiesTag, exists := entry["Properties"]; exists {
				properties, ok := propertiesTag.Object()
				if !ok {
					return nil, fmt.Errorf("invalid template properties")
				}
				states[index].Properties = make(map[string]string, len(properties))
				for name, tag := range properties {
					value, ok := tag.Text()
					if !ok {
						return nil, fmt.Errorf("invalid template property %s", name)
					}
					states[index].Properties[name] = value
				}
			}
			if _, err := blockstate.BlockStateID(states[index].Name, states[index].Properties); err != nil {
				return nil, err
			}
			full[index], err = templateFullCube(states[index])
			if err != nil {
				return nil, err
			}
		}
		var counts [3]int
		groupFor := func(source sourceBlock) int {
			if source.data != nil {
				return 2
			}
			if full[source.state] {
				return 0
			}
			return 1
		}
		for _, source := range sources {
			if source.state >= int64(len(states)) {
				return nil, fmt.Errorf("invalid structure block position/state")
			}
			counts[groupFor(source)]++
		}
		sorted := make([]StructureTemplateBlock, len(sources))
		starts := [4]int{0, counts[0], counts[0] + counts[1], len(sources)}
		next := starts
		for _, source := range sources {
			group := groupFor(source)
			sorted[next[group]] = StructureTemplateBlock{Position: source.position, State: states[source.state], NBT: source.data}
			next[group]++
		}
		for index := range counts {
			group := sorted[starts[index]:starts[index+1]]
			sort.SliceStable(group, func(i, j int) bool {
				a, b := group[i].Position, group[j].Position
				if a.Y != b.Y {
					return a.Y < b.Y
				}
				if a.X != b.X {
					return a.X < b.X
				}
				return a.Z < b.Z
			})
		}
		result.Palettes = append(result.Palettes, sorted)
	}
	entities, ok := object["entities"].Elements()
	if !ok {
		return nil, fmt.Errorf("missing template entity list")
	}
	for _, tag := range entities {
		entry, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("invalid template entity")
		}
		blockPosition, err := fossilVector(entry["blockPos"])
		if err != nil {
			return nil, err
		}
		coordinates, ok := entry["pos"].Elements()
		if !ok || len(coordinates) != 3 {
			return nil, fmt.Errorf("invalid template entity coordinates")
		}
		entity := StructureTemplateEntity{BlockPosition: blockPosition}
		for axis, tag := range coordinates {
			value, ok := tag.Value.(float64)
			if !ok || tag.Type != nbt.Double {
				return nil, fmt.Errorf("invalid template entity coordinate")
			}
			entity.Position[axis] = value
		}
		entity.NBT, ok = entry.Object("nbt")
		if !ok {
			return nil, fmt.Errorf("missing template entity NBT")
		}
		result.Entities = append(result.Entities, entity)
	}
	return result, nil
}

func (template *StructureTemplate) BoundingBox(origin, pivot FeaturePosition, mirror TemplateMirror, rotation TemplateRotation) (StructureBoundingBox, error) {
	if template == nil {
		return StructureBoundingBox{}, fmt.Errorf("missing structure template")
	}
	var box StructureBoundingBox
	first := true
	for _, x := range []int32{0, template.Size.X - 1} {
		for _, y := range []int32{0, template.Size.Y - 1} {
			for _, z := range []int32{0, template.Size.Z - 1} {
				p, err := TransformTemplatePosition(FeaturePosition{x, y, z}, pivot, mirror, rotation)
				if err != nil {
					return box, err
				}
				p = treeOffset(origin, p.X, p.Y, p.Z)
				if first {
					box = StructureBoundingBox{p.X, p.Y, p.Z, p.X, p.Y, p.Z}
					first = false
				} else {
					box.MinX = min(box.MinX, p.X)
					box.MinY = min(box.MinY, p.Y)
					box.MinZ = min(box.MinZ, p.Z)
					box.MaxX = max(box.MaxX, p.X)
					box.MaxY = max(box.MaxY, p.Y)
					box.MaxZ = max(box.MaxZ, p.Z)
				}
			}
		}
	}
	return box, nil
}
