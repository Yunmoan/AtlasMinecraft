package worldgen

import (
	"atlas/internal/nbt"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type JigsawStartSources struct {
	MinY, Height, GenerationDepth int32
	BiomeAtQuart                  StructureBiomeAtQuart
	// First free base-noise Y for the requested original WG heightmap.
	BaseHeightAt func(x, z int32, heightmap string) (int32, error)
}

type jigsawPlacedPiece struct {
	element        JigsawPoolElement
	origin         FeaturePosition
	rotation       TemplateRotation
	box            StructureBoundingBox
	groundDelta    int32
	junctions      []nbt.Tag
	liquidSettings string
}
type jigsawConnector struct {
	position FeaturePosition
	state    SurfaceBlockState
	data     nbt.Object
}
type jigsawSpace struct {
	bounds   StructureBoundingBox
	occupied []StructureBoundingBox
}
type jigsawTask struct {
	piece           *jigsawPlacedPiece
	space           *jigsawSpace
	depth, priority int32
	order           int
}

func jigsawShuffle[T any](values []T, random *LegacyRandom) {
	for size := len(values); size > 1; size-- {
		j := int(random.NextIntBound(int32(size)))
		values[size-1], values[j] = values[j], values[size-1]
	}
}
func jigsawWeightedElements(pool JigsawTemplatePool) []JigsawPoolElement {
	var result []JigsawPoolElement
	for _, entry := range pool.Elements {
		for i := int32(0); i < entry.Weight; i++ {
			result = append(result, entry.Element)
		}
	}
	return result
}
func jigsawElementBox(element JigsawPoolElement, origin FeaturePosition, rotation TemplateRotation) (StructureBoundingBox, error) {
	switch element.Type {
	case "minecraft:single_pool_element", "minecraft:legacy_single_pool_element":
		template, err := LoadJigsawTemplate(element.Location)
		if err != nil {
			return StructureBoundingBox{}, err
		}
		return template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	case "minecraft:list_pool_element":
		if len(element.Elements) == 0 {
			return StructureBoundingBox{}, fmt.Errorf("empty Jigsaw list element")
		}
		box, err := jigsawElementBox(element.Elements[0], origin, rotation)
		if err != nil {
			return box, err
		}
		for _, child := range element.Elements[1:] {
			if child.Type == "minecraft:empty_pool_element" {
				continue
			}
			b, err := jigsawElementBox(child, origin, rotation)
			if err != nil {
				return box, err
			}
			box.MinX, box.MinY, box.MinZ = min(box.MinX, b.MinX), min(box.MinY, b.MinY), min(box.MinZ, b.MinZ)
			box.MaxX, box.MaxY, box.MaxZ = max(box.MaxX, b.MaxX), max(box.MaxY, b.MaxY), max(box.MaxZ, b.MaxZ)
		}
		return box, nil
	case "minecraft:feature_pool_element", "minecraft:empty_pool_element":
		return StructureBoundingBox{MinX: origin.X, MinY: origin.Y, MinZ: origin.Z, MaxX: origin.X, MaxY: origin.Y, MaxZ: origin.Z}, nil
	default:
		return StructureBoundingBox{}, fmt.Errorf("unsupported Jigsaw element type %q", element.Type)
	}
}
func jigsawConnectors(element JigsawPoolElement, origin FeaturePosition, rotation TemplateRotation, random *LegacyRandom) ([]jigsawConnector, error) {
	if element.Type == "minecraft:list_pool_element" {
		if len(element.Elements) == 0 {
			return nil, fmt.Errorf("empty Jigsaw list element")
		}
		return jigsawConnectors(element.Elements[0], origin, rotation, random)
	}
	if element.Type == "minecraft:empty_pool_element" {
		return nil, nil
	}
	if element.Type == "minecraft:feature_pool_element" {
		data := nbt.Object{}
		for key, value := range map[string]string{"name": "minecraft:bottom", "final_state": "minecraft:air", "pool": "minecraft:empty", "target": "minecraft:empty", "joint": "rollable"} {
			data[key] = nbt.Tag{Type: nbt.String, Value: value}
		}
		// FeaturePoolElement intentionally keeps this downward-facing synthetic
		// connector unchanged for every requested template rotation.
		return []jigsawConnector{{position: origin, state: SurfaceBlockState{Name: "minecraft:jigsaw", Properties: map[string]string{"orientation": "down_south"}}, data: data}}, nil
	}
	template, err := LoadJigsawTemplate(element.Location)
	if err != nil {
		return nil, err
	}
	if len(template.Palettes) == 0 {
		return nil, nil
	}
	palettes, err := jigsawRotatedConnectorPalettes(template, rotation)
	if err != nil {
		return nil, err
	}
	palette := palettes[NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z)).NextIntBound(int32(len(template.Palettes)))]
	result := make([]jigsawConnector, len(palette))
	for i, connector := range palette {
		connector.position = treeOffset(origin, connector.position.X, connector.position.Y, connector.position.Z)
		result[i] = connector
	}
	jigsawShuffle(result, random)
	sort.SliceStable(result, func(i, j int) bool {
		a, _ := result[i].data.Number("selection_priority")
		b, _ := result[j].data.Number("selection_priority")
		return a > b
	})
	return result, nil
}

func jigsawDirection(name string) FeaturePosition {
	switch name {
	case "east":
		return FeaturePosition{X: 1}
	case "west":
		return FeaturePosition{X: -1}
	case "south":
		return FeaturePosition{Z: 1}
	case "north":
		return FeaturePosition{Z: -1}
	case "up":
		return FeaturePosition{Y: 1}
	case "down":
		return FeaturePosition{Y: -1}
	}
	return FeaturePosition{}
}
func jigsawFacing(c jigsawConnector) (string, string) {
	parts := strings.Split(c.state.Properties["orientation"], "_")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
func jigsawMatches(parent, child jigsawConnector) bool {
	front, top := jigsawFacing(parent)
	childFront, childTop := jigsawFacing(child)
	if jigsawDirection(front) != (FeaturePosition{X: -jigsawDirection(childFront).X, Y: -jigsawDirection(childFront).Y, Z: -jigsawDirection(childFront).Z}) {
		return false
	}
	joint, _ := parent.data.Text("joint")
	if joint != "rollable" && joint != "aligned" {
		if front == "up" || front == "down" {
			joint = "rollable"
		} else {
			joint = "aligned"
		}
	}
	if joint != "rollable" && top != childTop {
		return false
	}
	target, _ := parent.data.Text("target")
	name, _ := child.data.Text("name")
	return target == name
}
func jigsawInside(box StructureBoundingBox, p FeaturePosition) bool {
	return p.X >= box.MinX && p.X <= box.MaxX && p.Y >= box.MinY && p.Y <= box.MaxY && p.Z >= box.MinZ && p.Z <= box.MaxZ
}
func (space *jigsawSpace) accepts(box StructureBoundingBox) bool {
	// VoxelShape uses continuous [min,max+1] boxes and a candidate deflated
	// by 0.25. Integer boxes therefore admit boundary contact, but no overlap.
	b := space.bounds
	if box.MinX < b.MinX || box.MinY < b.MinY || box.MinZ < b.MinZ || box.MaxX > b.MaxX || box.MaxY > b.MaxY || box.MaxZ > b.MaxZ {
		return false
	}
	for _, b := range space.occupied {
		if box.MinX <= b.MaxX && box.MaxX >= b.MinX && box.MinY <= b.MaxY && box.MaxY >= b.MinY && box.MinZ <= b.MaxZ && box.MaxZ >= b.MinZ {
			return false
		}
	}
	return true
}
func jigsawOffsetBox(box StructureBoundingBox, dy int32) StructureBoundingBox {
	box.MinY += dy
	box.MaxY += dy
	return box
}
func jigsawJunction(p FeaturePosition, y, delta int32, projection string) nbt.Tag {
	return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"source_x": {Type: nbt.Int, Value: p.X}, "source_ground_y": {Type: nbt.Int, Value: y}, "source_z": {Type: nbt.Int, Value: p.Z}, "delta_y": {Type: nbt.Int, Value: delta}, "dest_proj": {Type: nbt.String, Value: projection}}}
}
func (piece *jigsawPlacedPiece) saved() (StructurePiece, error) {
	element, err := piece.element.NBT()
	if err != nil {
		return StructurePiece{}, err
	}
	listType := byte(nbt.Compound)
	if len(piece.junctions) == 0 {
		listType = nbt.End
	}
	data := nbt.Object{
		"PosX": {Type: nbt.Int, Value: piece.origin.X}, "PosY": {Type: nbt.Int, Value: piece.origin.Y}, "PosZ": {Type: nbt.Int, Value: piece.origin.Z},
		"ground_level_delta": {Type: nbt.Int, Value: piece.groundDelta}, "pool_element": {Type: nbt.Compound, Value: element},
		"rotation": {Type: nbt.String, Value: templateRotationNames[piece.rotation]}, "junctions": {Type: nbt.List, ListType: listType, Value: append([]nbt.Tag{}, piece.junctions...)},
	}
	if piece.liquidSettings != "" && piece.liquidSettings != "apply_waterlogging" {
		data["liquid_settings"] = nbt.Tag{Type: nbt.String, Value: piece.liquidSettings}
	}
	return StructurePiece{ID: "minecraft:jigsaw", Box: piece.box, Data: data}, nil
}

// GenerateJigsawStructureStart performs original weighted pool expansion,
// connector matching, priority traversal, projection and collision clipping.
// Blocks and processor placement are separate from this saved start stage.
func GenerateJigsawStructureStart(seed int64, cx, cz int32, structure string, sources JigsawStartSources) (*StructureStart, error) {
	base := NewLegacyRandom(seed)
	xm, zm := base.NextLong(), base.NextLong()
	random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	return generateJigsawStructureStart(seed, cx, cz, structure, sources, random)
}

func generateJigsawStructureStart(seed int64, cx, cz int32, structure string, sources JigsawStartSources, random *LegacyRandom) (*StructureStart, error) {
	if sources.BiomeAtQuart == nil || sources.Height <= 0 || sources.GenerationDepth <= 0 {
		return nil, fmt.Errorf("Jigsaw start requires build/generation bounds and biomes")
	}
	definition, err := BuiltinStructureJSON(structure)
	if err != nil {
		return nil, err
	}
	var settings struct {
		Type              string            `json:"type"`
		StartPool         string            `json:"start_pool"`
		StartName         string            `json:"start_jigsaw_name"`
		Size              int32             `json:"size"`
		Height            json.RawMessage   `json:"start_height"`
		ProjectHeightmap  string            `json:"project_start_to_heightmap"`
		Expansion         bool              `json:"use_expansion_hack"`
		MaxDistance       int32             `json:"max_distance_from_center"`
		TerrainAdjustment string            `json:"terrain_adaptation"`
		Padding           json.RawMessage   `json:"dimension_padding"`
		Aliases           []json.RawMessage `json:"pool_aliases"`
		LiquidSettings    string            `json:"liquid_settings"`
	}
	if err := json.Unmarshal(definition, &settings); err != nil {
		return nil, err
	}
	if settings.TerrainAdjustment == "" {
		settings.TerrainAdjustment = "none"
	}
	if settings.Type != "minecraft:jigsaw" {
		return nil, fmt.Errorf("%s is not Jigsaw", structure)
	}
	registry, err := loadJigsawRegistry()
	if err != nil {
		return nil, err
	}
	var height struct {
		Type     string       `json:"type"`
		Absolute *int32       `json:"absolute"`
		Min      heightAnchor `json:"min_inclusive"`
		Max      heightAnchor `json:"max_inclusive"`
	}
	if err := json.Unmarshal(settings.Height, &height); err != nil {
		return nil, err
	}
	y := int32(0)
	if height.Absolute != nil {
		y = *height.Absolute
	} else if height.Type == "minecraft:uniform" {
		low, err := height.Min.resolve(sources.MinY, sources.GenerationDepth)
		if err != nil {
			return nil, err
		}
		high, err := height.Max.resolve(sources.MinY, sources.GenerationDepth)
		if err != nil {
			return nil, err
		}
		y = low
		if high >= low {
			y += random.NextIntBound(high - low + 1)
		}
	} else {
		return nil, fmt.Errorf("unsupported Jigsaw start height")
	}
	initial := FeaturePosition{cx * 16, y, cz * 16}
	aliases, err := buildJigsawAliases(seed, initial, settings.Aliases)
	if err != nil {
		return nil, err
	}
	resolvePool := func(name string) string {
		name = densityName(name)
		if target, ok := aliases[name]; ok {
			return target
		}
		return name
	}
	rotation := TemplateRotation(random.NextIntBound(4))
	pool, found := registry.Pools[resolvePool(settings.StartPool)]
	if !found {
		return nil, fmt.Errorf("unknown Jigsaw start pool %s", settings.StartPool)
	}
	elements := jigsawWeightedElements(pool)
	if len(elements) == 0 {
		return nil, nil
	}
	selected := elements[random.NextIntBound(int32(len(elements)))]
	if selected.Type == "minecraft:empty_pool_element" {
		return nil, nil
	}
	anchor := initial
	if settings.StartName != "" {
		connectors, err := jigsawConnectors(selected, initial, rotation, random)
		if err != nil {
			return nil, err
		}
		matched := false
		for _, connector := range connectors {
			name, _ := connector.data.Text("name")
			if name == settings.StartName {
				anchor = connector.position
				matched = true
				break
			}
		}
		if !matched {
			return nil, nil
		}
	}
	anchorDelta := FeaturePosition{anchor.X - initial.X, anchor.Y - initial.Y, anchor.Z - initial.Z}
	origin := FeaturePosition{initial.X - anchorDelta.X, initial.Y - anchorDelta.Y, initial.Z - anchorDelta.Z}
	box, err := jigsawElementBox(selected, origin, rotation)
	if err != nil {
		return nil, err
	}
	centerX, centerZ := (box.MinX+box.MaxX)/2, (box.MinZ+box.MaxZ)/2
	groundY := origin.Y
	if settings.ProjectHeightmap != "" {
		if sources.BaseHeightAt == nil {
			return nil, fmt.Errorf("Jigsaw heightmap projection requires base heights")
		}
		free, err := sources.BaseHeightAt(centerX, centerZ, settings.ProjectHeightmap)
		if err != nil {
			return nil, err
		}
		groundY = initial.Y + free
	}
	shift := groundY - (box.MinY + 1)
	origin.Y += shift
	box = jigsawOffsetBox(box, shift)
	stubPosition := FeaturePosition{centerX, groundY + anchorDelta.Y, centerZ}
	biome, err := sources.BiomeAtQuart(stubPosition.X>>2, stubPosition.Y>>2, stubPosition.Z>>2)
	if err != nil {
		return nil, err
	}
	allowed, err := StructureBiomeNames(structure)
	if err != nil {
		return nil, err
	}
	valid := false
	for _, name := range allowed {
		if name == biome {
			valid = true
			break
		}
	}
	if !valid {
		return nil, nil
	}
	root := &jigsawPlacedPiece{element: selected, origin: origin, rotation: rotation, box: box, groundDelta: 1, liquidSettings: settings.LiquidSettings}
	pieces := []*jigsawPlacedPiece{root}
	var bottomPadding, topPadding int32
	if len(settings.Padding) > 0 {
		if err := json.Unmarshal(settings.Padding, &bottomPadding); err == nil {
			topPadding = bottomPadding
		} else {
			var p struct {
				Bottom int32 `json:"bottom"`
				Top    int32 `json:"top"`
			}
			if err := json.Unmarshal(settings.Padding, &p); err != nil {
				return nil, err
			}
			bottomPadding, topPadding = p.Bottom, p.Top
		}
	}
	distance := settings.MaxDistance
	// The continuous max edge is center+distance+1, represented inclusively.
	space := &jigsawSpace{bounds: StructureBoundingBox{MinX: centerX - distance, MinY: max(stubPosition.Y-distance, sources.MinY+bottomPadding), MinZ: centerZ - distance, MaxX: centerX + distance, MaxY: min(stubPosition.Y+distance, sources.MinY+sources.Height-topPadding-1), MaxZ: centerZ + distance}, occupied: []StructureBoundingBox{box}}
	queue := []jigsawTask{{piece: root, space: space}}
	order := 0
	for len(queue) > 0 && settings.Size > 0 {
		index := 0
		for i := 1; i < len(queue); i++ {
			if queue[i].priority > queue[index].priority || queue[i].priority == queue[index].priority && queue[i].order < queue[index].order {
				index = i
			}
		}
		task := queue[index]
		queue = append(queue[:index], queue[index+1:]...)
		parent := task.piece
		connectors, err := jigsawConnectors(parent.element, parent.origin, parent.rotation, random)
		if err != nil {
			return nil, err
		}
		var internal *jigsawSpace
		for _, connector := range connectors {
			front, _ := jigsawFacing(connector)
			direction := jigsawDirection(front)
			neighbor := treeOffset(connector.position, direction.X, direction.Y, direction.Z)
			heightOffset := connector.position.Y - parent.box.MinY
			poolName, _ := connector.data.Text("pool")
			poolName = resolvePool(poolName)
			pool, found := registry.Pools[poolName]
			if !found {
				return nil, fmt.Errorf("unknown connector pool %s", poolName)
			}
			if len(pool.Elements) == 0 && poolName != "minecraft:empty" {
				continue
			}
			fallback, found := registry.Pools[pool.Fallback]
			if !found {
				return nil, fmt.Errorf("unknown fallback pool %s", pool.Fallback)
			}
			if len(fallback.Elements) == 0 && pool.Fallback != "minecraft:empty" {
				continue
			}
			available := task.space
			if jigsawInside(parent.box, neighbor) {
				if internal == nil {
					internal = &jigsawSpace{bounds: parent.box}
				}
				available = internal
			}
			var candidates []JigsawPoolElement
			if task.depth != settings.Size {
				values := jigsawWeightedElements(pool)
				jigsawShuffle(values, random)
				candidates = append(candidates, values...)
			}
			values := jigsawWeightedElements(fallback)
			jigsawShuffle(values, random)
			candidates = append(candidates, values...)
			priority, _ := connector.data.Number("placement_priority")
			placed := false
			var terrainY int32 = -1
			for _, candidate := range candidates {
				if candidate.Type == "minecraft:empty_pool_element" {
					break
				}
				rotations := []TemplateRotation{0, 1, 2, 3}
				jigsawShuffle(rotations, random)
				for _, rotation := range rotations {
					children, err := jigsawConnectors(candidate, FeaturePosition{}, rotation, random)
					if err != nil {
						return nil, err
					}
					candidateBox, err := jigsawElementBox(candidate, FeaturePosition{}, rotation)
					if err != nil {
						return nil, err
					}
					var expandHeight int32
					if settings.Expansion && candidateBox.MaxY-candidateBox.MinY+1 <= 16 {
						for _, child := range children {
							facing, _ := jigsawFacing(child)
							delta := jigsawDirection(facing)
							if !jigsawInside(candidateBox, treeOffset(child.position, delta.X, delta.Y, delta.Z)) {
								continue
							}
							name, _ := child.data.Text("pool")
							p, ok := registry.Pools[resolvePool(name)]
							if !ok {
								continue
							}
							for _, p := range []JigsawTemplatePool{p, registry.Pools[p.Fallback]} {
								for _, entry := range p.Elements {
									if entry.Element.Type == "minecraft:empty_pool_element" {
										continue
									}
									b, err := jigsawElementBox(entry.Element, FeaturePosition{}, TemplateRotationNone)
									if err != nil {
										return nil, err
									}
									expandHeight = max(expandHeight, b.MaxY-b.MinY+1)
								}
							}
						}
					}
					for _, child := range children {
						if !jigsawMatches(connector, child) {
							continue
						}
						childOrigin := FeaturePosition{neighbor.X - child.position.X, neighbor.Y - child.position.Y, neighbor.Z - child.position.Z}
						childBox, err := jigsawElementBox(candidate, childOrigin, rotation)
						if err != nil {
							return nil, err
						}
						deltaY := heightOffset - child.position.Y + direction.Y
						parentRigid, childRigid := parent.element.Projection == "rigid", candidate.Projection == "rigid"
						childMinY := parent.box.MinY + deltaY
						if !parentRigid || !childRigid {
							if sources.BaseHeightAt == nil {
								return nil, fmt.Errorf("terrain-matching Jigsaw element requires height source")
							}
							if terrainY == -1 {
								terrainY, err = sources.BaseHeightAt(connector.position.X, connector.position.Z, "WORLD_SURFACE_WG")
								if err != nil {
									return nil, err
								}
							}
							childMinY = terrainY - child.position.Y
						}
						shift := childMinY - childBox.MinY
						childOrigin.Y += shift
						childBox = jigsawOffsetBox(childBox, shift)
						if expandHeight > 0 {
							childBox.MaxY = max(childBox.MaxY, childBox.MinY+max(expandHeight+1, childBox.MaxY-childBox.MinY))
						}
						if !available.accepts(childBox) {
							continue
						}
						available.occupied = append(available.occupied, childBox)
						groundDelta := int32(1)
						if childRigid {
							groundDelta = parent.groundDelta - deltaY
						}
						piece := &jigsawPlacedPiece{element: candidate, origin: childOrigin, rotation: rotation, box: childBox, groundDelta: groundDelta, liquidSettings: settings.LiquidSettings}
						junctionY := parent.box.MinY + heightOffset
						if !parentRigid {
							if childRigid {
								junctionY = childMinY + child.position.Y
							} else {
								junctionY = terrainY + deltaY/2
							}
						}
						parent.junctions = append(parent.junctions, jigsawJunction(neighbor, junctionY-heightOffset+parent.groundDelta, deltaY, candidate.Projection))
						piece.junctions = append(piece.junctions, jigsawJunction(connector.position, junctionY-child.position.Y+groundDelta, -deltaY, parent.element.Projection))
						pieces = append(pieces, piece)
						if task.depth+1 <= settings.Size {
							order++
							queue = append(queue, jigsawTask{piece: piece, space: available, depth: task.depth + 1, priority: int32(priority), order: order})
						}
						placed = true
						break
					}
					if placed {
						break
					}
				}
				if placed {
					break
				}
			}
		}
	}
	start := &StructureStart{Structure: densityName(structure), Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: settings.TerrainAdjustment}
	for _, piece := range pieces {
		saved, err := piece.saved()
		if err != nil {
			return nil, err
		}
		start.Pieces = append(start.Pieces, saved)
	}
	return start, nil
}
