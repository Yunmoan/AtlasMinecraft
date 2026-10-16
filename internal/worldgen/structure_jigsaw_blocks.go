package worldgen

import (
	"atlas/internal/nbt"
	"encoding/json"
	"fmt"
	"strings"
)

func jigsawNBTJSON(tag nbt.Tag) any {
	if object, ok := tag.Object(); ok {
		result := make(map[string]any, len(object))
		for name, value := range object {
			result[name] = jigsawNBTJSON(value)
		}
		return result
	}
	if list, ok := tag.Elements(); ok {
		result := make([]any, len(list))
		for i, value := range list {
			result[i] = jigsawNBTJSON(value)
		}
		return result
	}
	return tag.Value
}

func jigsawSavedBlockSettings(piece *StructurePiece) (JigsawPoolElement, FeaturePosition, TemplateRotation, string, error) {
	var element JigsawPoolElement
	if piece == nil || piece.ID != "minecraft:jigsaw" {
		return element, FeaturePosition{}, 0, "", fmt.Errorf("Jigsaw requires a saved pool piece")
	}
	data, ok := piece.Data.Object("pool_element")
	if !ok {
		return element, FeaturePosition{}, 0, "", fmt.Errorf("Jigsaw piece lacks pool_element")
	}
	encoded, err := json.Marshal(jigsawNBTJSON(nbt.Tag{Type: nbt.Compound, Value: data}))
	if err != nil {
		return element, FeaturePosition{}, 0, "", err
	}
	if err = json.Unmarshal(encoded, &element); err != nil {
		return element, FeaturePosition{}, 0, "", err
	}
	origin, rotation, liquid, err := jigsawSavedOriginRotation(piece)
	return element, origin, rotation, liquid, err
}

func jigsawSavedOriginRotation(piece *StructurePiece) (FeaturePosition, TemplateRotation, string, error) {
	var origin FeaturePosition
	if piece == nil || piece.ID != "minecraft:jigsaw" {
		return origin, 0, "", fmt.Errorf("Jigsaw requires a saved pool piece")
	}
	for index, key := range []string{"PosX", "PosY", "PosZ"} {
		value, ok := piece.Data.Number(key)
		if !ok {
			return origin, 0, "", fmt.Errorf("Jigsaw piece lacks %s", key)
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
	rotationText, _ := piece.Data.Text("rotation")
	rotation := TemplateRotation(255)
	for index, text := range templateRotationNames {
		if text == rotationText {
			rotation = TemplateRotation(index)
		}
	}
	if rotation > TemplateRotationCounterclockwise90 {
		return origin, rotation, "", fmt.Errorf("invalid saved Jigsaw rotation %q", rotationText)
	}
	liquid, _ := piece.Data.Text("liquid_settings")
	if liquid == "" {
		liquid = "apply_waterlogging"
	}
	return origin, rotation, liquid, nil
}

// PostProcessJigsawBlocks places rigid/terrain-matching original single and
// list pool elements. pivot is StructureStart's first piece center X/Z and
// minimum Y, as used by original position-sensitive rules.
func PostProcessJigsawBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, cx, cz int32, pivot FeaturePosition) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || random == nil {
		return result, fmt.Errorf("Jigsaw requires world and decoration random")
	}
	element, origin, rotation, liquid, err := jigsawSavedBlockSettings(piece)
	if err != nil {
		return result, err
	}
	minY, height := int32(-64), int32(384)
	if bounds, ok := world.(interface {
		MinBuildHeight() int32
		MaxBuildHeight() int32
	}); ok {
		minY = bounds.MinBuildHeight()
		height = bounds.MaxBuildHeight() - minY
	}
	clip := StructureBoundingBox{cx * 16, minY, cz * 16, cx*16 + 15, minY + height - 1, cz*16 + 15}
	return placeJigsawElement(world, element, origin, pivot, rotation, liquid, random, clip)
}

func placeJigsawElement(world StructureBlockEntityWorld, element JigsawPoolElement, origin, pivot FeaturePosition, rotation TemplateRotation, liquid string, random StructurePieceRandom, clip StructureBoundingBox) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if element.Type == "minecraft:empty_pool_element" {
		return result, nil
	}
	if element.Type == "minecraft:list_pool_element" {
		for _, child := range element.Elements {
			placed, err := placeJigsawElement(world, child, origin, pivot, rotation, liquid, random, clip)
			if err != nil {
				return result, err
			}
			result.PendingEntities = append(result.PendingEntities, placed.PendingEntities...)
		}
		return result, nil
	}
	if element.Type == "minecraft:feature_pool_element" {
		placer, ok := world.(interface {
			PlaceJigsawFeature(json.RawMessage, StructurePieceRandom, FeaturePosition) (bool, error)
		})
		if !ok {
			return result, fmt.Errorf("Jigsaw feature element requires a feature region")
		}
		_, err := placer.PlaceJigsawFeature(element.Feature, random, origin)
		return result, err
	}
	if element.Type != "minecraft:single_pool_element" && element.Type != "minecraft:legacy_single_pool_element" {
		return result, fmt.Errorf("Jigsaw block placement for %q is not implemented", element.Type)
	}
	template, err := LoadJigsawTemplate(element.Location)
	if err != nil {
		return result, err
	}
	blocks, err := ProcessJigsawTemplateBlocks(world, template, element, origin, pivot, rotation)
	if err != nil {
		return result, err
	}
	if element.LiquidSettings != "" {
		liquid = element.LiquidSettings
	}
	if liquid != "apply_waterlogging" && liquid != "ignore_waterlogging" {
		return result, fmt.Errorf("unknown Jigsaw liquid setting %q", liquid)
	}
	var pendingLiquid []FeaturePosition
	templateSources := make(map[FeaturePosition]bool)
	for _, block := range blocks {
		p := block.Position
		if !structureContains(clip, p) || !world.InWorld(p) {
			continue
		}
		state, err := TransformTemplateBlockState(block.State, TemplateMirrorNone, rotation)
		if err != nil {
			return result, err
		}
		// TransformTemplateBlockState already resolves the full original state
		// and returns detached properties. Repeating ID lookup/load here only
		// copies the same property map a second time.
		_, container := state.Properties["waterlogged"]
		container = container && state.Properties["type"] != "double"
		if liquid == "apply_waterlogging" {
			old, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			if _, source := ruinedPortalFluid(state); source {
				templateSources[p] = true
			}
			if container && (old.Name == "minecraft:water" || old.Properties["waterlogged"] == "true") {
				state = jigsawWaterloggedState(state)
			}
			if container {
				if _, source := ruinedPortalFluid(state); !source {
					pendingLiquid = append(pendingLiquid, p)
				}
			}
		}
		var defaultEntity nbt.Object
		if block.NBT == nil && (state.Name == "minecraft:campfire" || state.Name == "minecraft:soul_campfire") {
			previous, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			// Level.setBlock creates a new entity when its block changes. A
			// Jigsaw replacement can create a campfire without template NBT;
			// retaining the same block must preserve its existing saved data.
			if previous.Name != state.Name {
				defaultEntity = jigsawDefaultCampfireBlockEntity()
			}
		}
		if block.NBT != nil {
			// StructureTemplate clears previous block entities with a barrier
			// before installing/loading a new one, even when types match.
			if err := world.SetBlock(p, structureState("barrier")); err != nil {
				return result, err
			}
		}
		if err := world.SetBlock(p, state); err != nil {
			return result, err
		}
		if block.NBT != nil {
			data, err := jigsawPlacedBlockEntity(state, block.NBT, random)
			if err != nil {
				return result, err
			}
			if data != nil {
				if err := world.SetBlockEntity(p, state.Name, data); err != nil {
					return result, err
				}
			}
		} else if defaultEntity != nil {
			if err := world.SetBlockEntity(p, state.Name, defaultEntity); err != nil {
				return result, err
			}
		}
	}
	for changed := true; changed && len(pendingLiquid) > 0; {
		changed = false
		remaining := pendingLiquid[:0]
		for _, p := range pendingLiquid {
			state, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			kind, source := ruinedPortalFluid(state)
			for _, direction := range []string{"up", "north", "east", "south", "west"} {
				if source {
					break
				}
				q := iglooDirectionOffset(p, direction)
				neighbor, err := world.BlockAt(q)
				if err != nil {
					return result, err
				}
				neighborKind, neighborSource := ruinedPortalFluid(neighbor)
				if neighborSource && !templateSources[q] {
					kind, source = neighborKind, true
				}
			}
			if source && kind == "water" {
				if err := world.SetBlock(p, jigsawWaterloggedState(state)); err != nil {
					return result, err
				}
				changed = true
			} else {
				remaining = append(remaining, p)
			}
		}
		pendingLiquid = remaining
	}
	// SinglePoolElement sets knownShape=true. The template preserves authored
	// connections and does not run its boundary/interior update-shape loops.
	for _, actor := range template.Entities {
		position, _ := TransformTemplatePosition(actor.BlockPosition, FeaturePosition{}, TemplateMirrorNone, rotation)
		position = treeOffset(origin, position.X, position.Y, position.Z)
		if !structureContains(clip, position) {
			continue
		}
		data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: actor.NBT}).Value.(nbt.Object)
		delete(data, "UUID")
		precise := transformTemplateEntityPosition(actor.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
		for axis, value := range []int32{origin.X, origin.Y, origin.Z} {
			precise[axis] += float64(value)
		}
		values := make([]nbt.Tag, 3)
		for axis, value := range precise {
			values[axis] = nbt.Tag{Type: nbt.Double, Value: value}
		}
		data["Pos"] = nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: values}
		// Entity.load initializes missing Rotation to zero before the template
		// rotates it. The original camel_spawn actor contains only its id.
		if _, present := data["Rotation"]; !present {
			data["Rotation"] = nbt.Tag{Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{{Type: nbt.Float, Value: float32(0)}, {Type: nbt.Float, Value: float32(0)}}}
		}
		if angles, ok := data["Rotation"].Elements(); ok && len(angles) == 2 {
			if yaw, ok := angles[0].Value.(float32); ok {
				angles[0] = nbt.Tag{Type: nbt.Float, Value: transformTemplateEntityYaw(yaw, TemplateMirrorNone, rotation)}
			}
		}
		id, _ := data.Text("id")
		persistent, _ := data.Number("PersistenceRequired")
		result.PendingEntities = append(result.PendingEntities, PendingStructureEntity{ID: id, Position: position, Persistent: persistent != 0, Data: data})
	}
	return result, nil
}

// Native CampfireBlockEntity constructor/save defaults. Both campfire block
// types share this entity type; no template load or decoration RNG is used.
func jigsawDefaultCampfireBlockEntity() nbt.Object {
	return nbt.Object{
		"id":                {Type: nbt.String, Value: "minecraft:campfire"},
		"Items":             {Type: nbt.List, ListType: nbt.End, Value: []nbt.Tag{}},
		"CookingTimes":      {Type: nbt.IntArray, Value: []int32{0, 0, 0, 0}},
		"CookingTotalTimes": {Type: nbt.IntArray, Value: []int32{0, 0, 0, 0}},
	}
}

func jigsawPlacedBlockEntity(state SurfaceBlockState, input nbt.Object, random StructurePieceRandom) (nbt.Object, error) {
	return jigsawNormalizeBlockEntity(state, input, random)
}

func jigsawWaterloggedState(state SurfaceBlockState) SurfaceBlockState {
	if state.Properties["waterlogged"] == "true" {
		return state
	}
	state = stateWithProperty(state, "waterlogged", "true")
	// Original CandleBlock and CampfireBlock extinguish when placeLiquid
	// accepts water; this is separate from template neighbor-shape updates.
	if state.Name == "minecraft:candle" || strings.HasSuffix(state.Name, "_candle") || state.Name == "minecraft:campfire" || state.Name == "minecraft:soul_campfire" {
		state = stateWithProperty(state, "lit", "false")
	}
	return state
}
