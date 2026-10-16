package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

// These are the original compressed templates, not generated skeletons.
// This loader deliberately supports the full-cube, entity-free fossil subset
// of StructureTemplate; other structure types require their own implementation.
//
//go:embed fossil_data/*
var fossilAssets embed.FS

type fossilBlock struct {
	position FeaturePosition
	state    SurfaceBlockState
}

type fossilTemplate struct {
	size   FeaturePosition
	blocks []fossilBlock
}

var fossilTemplatesOnce sync.Once
var fossilTemplates map[string]*fossilTemplate
var fossilTemplatesError error

func fossilVector(tag nbt.Tag) (FeaturePosition, error) {
	items, ok := tag.Elements()
	if !ok || tag.ListType != nbt.Int || len(items) != 3 {
		return FeaturePosition{}, fmt.Errorf("fossil requires a three-int vector")
	}
	var v [3]int32
	for i, item := range items {
		n, ok := item.Number()
		if !ok || item.Type != nbt.Int {
			return FeaturePosition{}, fmt.Errorf("invalid fossil vector component")
		}
		v[i] = int32(n)
	}
	return FeaturePosition{v[0], v[1], v[2]}, nil
}

func decodeFossilTemplate(data []byte) (*fossilTemplate, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	data, err = io.ReadAll(io.LimitReader(gz, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return nil, fmt.Errorf("invalid compressed fossil template: %v", err)
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		return nil, err
	}
	obj, _ := root.Object()
	version, ok := obj.Number("DataVersion")
	if !ok || version != 3955 {
		return nil, fmt.Errorf("fossil template requires DataVersion 3955")
	}
	size, err := fossilVector(obj["size"])
	if err != nil || size.X < 1 || size.Y < 1 || size.Z < 1 || size.X > 48 || size.Y > 48 || size.Z > 48 {
		return nil, fmt.Errorf("invalid fossil template size %v", size)
	}
	entities, ok := obj["entities"].Elements()
	if !ok || len(entities) != 0 {
		return nil, fmt.Errorf("fossil entities are unsupported")
	}
	if _, ok := obj["palettes"]; ok {
		return nil, fmt.Errorf("multiple fossil palettes are unsupported")
	}
	palette, ok := obj["palette"].Elements()
	if !ok || len(palette) == 0 {
		return nil, fmt.Errorf("missing fossil palette")
	}
	states := make([]SurfaceBlockState, len(palette))
	for i, item := range palette {
		entry, ok := item.Object()
		if !ok {
			return nil, fmt.Errorf("invalid fossil palette")
		}
		states[i].Name, ok = entry.Text("Name")
		if !ok || (states[i].Name != "minecraft:bone_block" && states[i].Name != "minecraft:coal_ore") {
			return nil, fmt.Errorf("unsupported fossil state %s", states[i].Name)
		}
		if tag, exists := entry["Properties"]; exists {
			properties, ok := tag.Object()
			if !ok {
				return nil, fmt.Errorf("invalid fossil properties")
			}
			states[i].Properties = make(map[string]string, len(properties))
			for key, value := range properties {
				text, ok := value.Text()
				if !ok {
					return nil, fmt.Errorf("invalid fossil property %s", key)
				}
				states[i].Properties[key] = text
			}
		}
		if _, err := blockstate.BlockStateID(states[i].Name, states[i].Properties); err != nil {
			return nil, err
		}
	}
	blocks, ok := obj["blocks"].Elements()
	if !ok || len(blocks) == 0 {
		return nil, fmt.Errorf("missing fossil blocks")
	}
	result := &fossilTemplate{size: size}
	for _, tag := range blocks {
		entry, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("invalid fossil block")
		}
		if _, exists := entry["nbt"]; exists {
			return nil, fmt.Errorf("fossil block entities are unsupported")
		}
		p, err := fossilVector(entry["pos"])
		index, ok := entry.Number("state")
		if err != nil || !ok || index < 0 || index >= int64(len(states)) || p.X < 0 || p.Y < 0 || p.Z < 0 || p.X >= size.X || p.Y >= size.Y || p.Z >= size.Z {
			return nil, fmt.Errorf("invalid fossil block position or palette index")
		}
		result.blocks = append(result.blocks, fossilBlock{p, states[index]})
	}
	// StructureTemplate puts full cubes without NBT first and stably sorts
	// that group by Y, X, Z. All original fossil states belong to this group.
	sort.SliceStable(result.blocks, func(i, j int) bool {
		a, b := result.blocks[i].position, result.blocks[j].position
		if a.Y != b.Y {
			return a.Y < b.Y
		}
		if a.X != b.X {
			return a.X < b.X
		}
		return a.Z < b.Z
	})
	return result, nil
}

func loadFossilTemplate(name string) (*fossilTemplate, error) {
	fossilTemplatesOnce.Do(func() {
		fossilTemplates = make(map[string]*fossilTemplate)
		for _, kind := range []string{"spine", "skull"} {
			for i := 1; i <= 4; i++ {
				for _, suffix := range []string{"", "_coal"} {
					key := fmt.Sprintf("%s_%d%s", kind, i, suffix)
					data, err := fossilAssets.ReadFile("fossil_data/" + key + ".nbt")
					if err == nil {
						fossilTemplates["minecraft:fossil/"+key], err = decodeFossilTemplate(data)
					}
					if err != nil {
						fossilTemplatesError = fmt.Errorf("fossil %s: %w", key, err)
						return
					}
				}
			}
		}
	})
	if fossilTemplatesError != nil {
		return nil, fossilTemplatesError
	}
	template := fossilTemplates[name]
	if template == nil {
		return nil, fmt.Errorf("unsupported fossil template %s", name)
	}
	return template, nil
}

type fossilProcessor struct {
	integrity float32
	diamond   bool
}

func loadFossilProcessor(name string) (fossilProcessor, error) {
	if name != "minecraft:fossil_rot" && name != "minecraft:fossil_coal" && name != "minecraft:fossil_diamonds" {
		return fossilProcessor{}, fmt.Errorf("unsupported fossil processor list %s", name)
	}
	data, err := fossilAssets.ReadFile("fossil_data/" + strings.TrimPrefix(name, "minecraft:") + ".json")
	if err != nil {
		return fossilProcessor{}, err
	}
	var list struct {
		Processors []json.RawMessage `json:"processors"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return fossilProcessor{}, err
	}
	var result fossilProcessor
	for i, raw := range list.Processors {
		var p struct {
			Type      string  `json:"processor_type"`
			Integrity float32 `json:"integrity"`
			Value     string  `json:"value"`
			Rules     []struct {
				Input struct {
					Type  string `json:"predicate_type"`
					Block string `json:"block"`
				} `json:"input_predicate"`
				Location struct {
					Type string `json:"predicate_type"`
				} `json:"location_predicate"`
				Output SurfaceBlockState `json:"output_state"`
			} `json:"rules"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return result, err
		}
		switch {
		case i == 0 && p.Type == "minecraft:block_rot" && p.Integrity >= 0 && p.Integrity <= 1:
			result.integrity = p.Integrity
		case i == 1 && p.Type == "minecraft:rule" && len(list.Processors) == 3 && len(p.Rules) == 1:
			rule := p.Rules[0]
			if rule.Input.Type != "minecraft:block_match" || rule.Input.Block != "minecraft:coal_ore" || rule.Location.Type != "minecraft:always_true" || rule.Output.Name != "minecraft:deepslate_diamond_ore" || len(rule.Output.Properties) != 0 {
				return result, fmt.Errorf("unsupported fossil rule")
			}
			result.diamond = true
		case i == len(list.Processors)-1 && p.Type == "minecraft:protected_blocks" && p.Value == "#minecraft:features_cannot_replace":
		default:
			return result, fmt.Errorf("unsupported fossil processor %s", p.Type)
		}
	}
	if len(list.Processors) != 2 && len(list.Processors) != 3 {
		return result, fmt.Errorf("invalid fossil processors")
	}
	return result, nil
}

func rotateFossilPosition(p FeaturePosition, rotation int32) FeaturePosition {
	switch rotation {
	case 1:
		return FeaturePosition{-p.Z, p.Y, p.X}
	case 2:
		return FeaturePosition{-p.X, p.Y, -p.Z}
	case 3:
		return FeaturePosition{p.Z, p.Y, -p.X}
	default:
		return p
	}
}

func placeFossilTemplate(e *featureEditor, r *WorldgenRandom, template *fossilTemplate, processor fossilProcessor, start FeaturePosition, rotation, chunkX, chunkZ int32) {
	// getRandomPalette consumes a draw even with the only palette.
	r.NextIntBound(1)
	processed := make([]fossilBlock, 0, len(template.blocks))
	for _, block := range template.blocks {
		p := rotateFossilPosition(block.position, rotation)
		p = treeOffset(start, p.X, p.Y, p.Z)
		if r.NextFloat() > processor.integrity {
			continue
		}
		if vegetationTags["features_cannot_replace"][e.block(p).Name] {
			continue
		}
		state := block.state
		if processor.diamond && state.Name == "minecraft:coal_ore" {
			state = SurfaceBlockState{Name: "minecraft:deepslate_diamond_ore"}
		}
		if rotation&1 != 0 && state.Name == "minecraft:bone_block" {
			switch state.Properties["axis"] {
			case "x":
				state = stateWithProperty(state, "axis", "z")
			case "z":
				state = stateWithProperty(state, "axis", "x")
			}
		}
		processed = append(processed, fossilBlock{p, state})
	}
	// Process all blocks before clipping/placing; protected-block checks see
	// the world from before this pass, and overlay processing sees the base pass.
	for _, block := range processed {
		p := block.position
		if p.X >= chunkX*16-16 && p.X <= chunkX*16+31 && p.Z >= chunkZ*16-16 && p.Z <= chunkZ*16+31 && p.Y >= -64 && p.Y <= 320 {
			e.set(p, block.state)
		}
	}
}

func compileFossilPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config struct {
		Primary          []string `json:"fossil_structures"`
		Overlay          []string `json:"overlay_structures"`
		PrimaryProcessor string   `json:"fossil_processors"`
		OverlayProcessor string   `json:"overlay_processors"`
		MaxEmptyCorners  int      `json:"max_empty_corners_allowed"`
	}
	if err := json.Unmarshal(def.Config, &config); err != nil {
		return nil, err
	}
	if len(config.Primary) == 0 || len(config.Primary) != len(config.Overlay) || config.MaxEmptyCorners < 0 || config.MaxEmptyCorners > 7 {
		return nil, fmt.Errorf("invalid fossil configuration")
	}
	primary, overlay := make([]*fossilTemplate, len(config.Primary)), make([]*fossilTemplate, len(config.Primary))
	for i := range primary {
		var err error
		primary[i], err = loadFossilTemplate(config.Primary[i])
		if err != nil {
			return nil, err
		}
		overlay[i], err = loadFossilTemplate(config.Overlay[i])
		if err != nil {
			return nil, err
		}
	}
	baseProcessor, err := loadFossilProcessor(config.PrimaryProcessor)
	if err != nil {
		return nil, err
	}
	overlayProcessor, err := loadFossilProcessor(config.OverlayProcessor)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		rotation, index := r.NextIntBound(4), r.NextIntBound(int32(len(primary)))
		template := primary[index]
		width, depth := template.size.X, template.size.Z
		if rotation&1 != 0 {
			width, depth = depth, width
		}
		center := treeOffset(origin, -width/2, 0, -depth/2)
		height := origin.Y
		for x := int32(0); x < width; x++ {
			for z := int32(0); z < depth; z++ {
				y, err := region.OceanFloorHeightAt(center.X+x, center.Z+z)
				if err != nil {
					return false, err
				}
				height = min(height, y)
			}
		}
		center.Y = max(height-15-r.NextIntBound(10), -54)
		start := center
		// StructureTemplate.getZeroPositionWithTransform uses the inclusive
		// maximum local coordinate (size - 1), as confirmed by the 1.21.1
		// bytecode's unconditional width/depth decrements before mirroring.
		switch rotation {
		case 1:
			start.X += template.size.Z - 1
		case 2:
			start.X += template.size.X - 1
			start.Z += template.size.Z - 1
		case 3:
			start.Z += template.size.X - 1
		}
		a, b := rotateFossilPosition(FeaturePosition{}, rotation), rotateFossilPosition(treeOffset(template.size, -1, -1, -1), rotation)
		empty := 0
		for _, x := range []int32{min(a.X, b.X), max(a.X, b.X)} {
			for _, y := range []int32{min(a.Y, b.Y), max(a.Y, b.Y)} {
				for _, z := range []int32{min(a.Z, b.Z), max(a.Z, b.Z)} {
					state := e.block(treeOffset(start, x, y, z))
					if surfaceAir(state) || state.Name == "minecraft:water" || state.Name == "minecraft:lava" {
						empty++
					}
				}
			}
		}
		if e.err != nil {
			return false, e.err
		}
		if empty > config.MaxEmptyCorners {
			return false, nil
		}
		placeFossilTemplate(&e, r, template, baseProcessor, start, rotation, origin.X>>4, origin.Z>>4)
		placeFossilTemplate(&e, r, overlay[index], overlayProcessor, start, rotation, origin.X>>4, origin.Z>>4)
		return true, e.err
	}, nil
}
