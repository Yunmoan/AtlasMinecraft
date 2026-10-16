package worldgen

import (
	"atlas/internal/nbt"
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

func structureWorldSurfaceHeight(world FeatureBlockWorld, x, z int32) (int32, error) {
	if source, ok := world.(interface {
		HeightmapAt(string, int32, int32) (int32, error)
	}); ok {
		return source.HeightmapAt("WORLD_SURFACE_WG", x, z)
	}
	for y := int32(319); y >= -64; y-- {
		state, err := world.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return 0, err
		}
		if !surfaceAir(state) {
			return y + 1, nil
		}
	}
	return -64, nil
}

// PostProcessIglooBlocks places the exact original top/ladder/laboratory
// template block stage. Template actors are returned as explicit pending
// requests; callers must save and instantiate them before treating them as live.
func PostProcessIglooBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:iglu" || random == nil {
		return result, fmt.Errorf("igloo requires an iglu piece, world and decoration random")
	}
	name, ok := piece.Data.Text("Template")
	if !ok {
		return result, fmt.Errorf("igloo piece has no Template")
	}
	template, err := LoadIglooTemplate(name)
	if err != nil {
		return result, err
	}
	rotationText, _ := piece.Data.Text("Rot")
	rotation := TemplateRotation(255)
	for index, name := range templateRotationNames {
		if name == rotationText {
			rotation = TemplateRotation(index)
		}
	}
	if rotation > TemplateRotationCounterclockwise90 {
		return result, fmt.Errorf("invalid saved igloo rotation %q", rotationText)
	}
	values := []int32{}
	for _, field := range []string{"TPX", "TPY", "TPZ"} {
		value, ok := piece.Data.Number(field)
		if !ok {
			return result, fmt.Errorf("igloo piece lacks %s", field)
		}
		values = append(values, int32(value))
	}
	savedOrigin := FeaturePosition{values[0], values[1], values[2]}
	pivot, offset := iglooPivots[name], iglooOffsets[name]
	relative, _ := TransformTemplatePosition(FeaturePosition{3 - offset.X, 0, -offset.Z}, pivot, TemplateMirrorNone, rotation)
	height, err := structureWorldSurfaceHeight(world, savedOrigin.X+relative.X, savedOrigin.Z+relative.Z)
	if err != nil {
		return result, err
	}
	origin := treeOffset(savedOrigin, 0, height-91, 0)
	piece.Box, err = template.BoundingBox(origin, pivot, TemplateMirrorNone, rotation)
	if err != nil {
		return result, err
	}
	clip := StructureBoundingBox{chunkX * 16, -64, chunkZ * 16, chunkX*16 + 15, 319, chunkZ*16 + 15}
	positionRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	paletteIndex := positionRandom.NextIntBound(int32(len(template.Palettes)))
	palette := template.Palettes[paletteIndex]
	var placed []FeaturePosition
	entities := make(map[FeaturePosition]featureBlockEntity)
	for _, block := range palette {
		if block.State.Name == "minecraft:structure_block" {
			continue
		}
		position, _ := TransformTemplatePosition(block.Position, pivot, TemplateMirrorNone, rotation)
		position = treeOffset(origin, position.X, position.Y, position.Z)
		if !structureContains(clip, position) || !world.InWorld(position) {
			continue
		}
		state, err := TransformTemplateBlockState(block.State, TemplateMirrorNone, rotation)
		if err != nil {
			return result, err
		}
		if err := world.SetBlock(position, state); err != nil {
			return result, err
		}
		placed = append(placed, position)
		if block.NBT != nil {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object)
			if state.Name == "minecraft:oak_wall_sign" {
				if err := normalizeIglooSignNBT(data); err != nil {
					return result, err
				}
			}
			if state.Name == "minecraft:chest" {
				data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(random.NextLong())}
			}
			if err := world.SetBlockEntity(position, state.Name, data); err != nil {
				return result, err
			}
			entities[position] = featureBlockEntity{state.Name, data}
		}
	}
	if err := finalizeIglooTemplateShape(world, placed); err != nil {
		return result, err
	}
	// Igloo's data marker is ignored by the template processor, then explicitly
	// replaced with air; the chest one block below receives the original loot.
	for _, block := range palette {
		if block.State.Name != "minecraft:structure_block" || block.NBT == nil {
			continue
		}
		mode, _ := block.NBT.Text("mode")
		metadata, _ := block.NBT.Text("metadata")
		if mode != "DATA" || metadata != "chest" {
			continue
		}
		position, _ := TransformTemplatePosition(block.Position, pivot, TemplateMirrorNone, rotation)
		position = treeOffset(origin, position.X, position.Y, position.Z)
		if !structureContains(clip, position) {
			continue
		}
		if err := world.SetBlock(position, structureState("air")); err != nil {
			return result, err
		}
		chestPosition := treeOffset(position, 0, -1, 0)
		chest, err := world.BlockAt(chestPosition)
		if err != nil {
			return result, err
		}
		if chest.Name == "minecraft:chest" {
			data := entities[chestPosition].data
			if data == nil {
				data = nbt.Object{"id": {Type: nbt.String, Value: "minecraft:chest"}}
			}
			data["LootTable"] = nbt.Tag{Type: nbt.String, Value: "minecraft:chests/igloo_chest"}
			data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(random.NextLong())}
			// RandomizableContainerBlockEntity saves the loot table instead of
			// serializing Items until that table has been unpacked.
			delete(data, "Items")
			if err := world.SetBlockEntity(chestPosition, "minecraft:chest", data); err != nil {
				return result, err
			}
		}
	}
	if name == "minecraft:igloo/top" {
		position, _ := TransformTemplatePosition(FeaturePosition{3, 0, 5}, pivot, TemplateMirrorNone, rotation)
		position = treeOffset(origin, position.X, position.Y, position.Z)
		below, err := world.BlockAt(treeOffset(position, 0, -1, 0))
		if err != nil {
			return result, err
		}
		if !surfaceAir(below) && below.Name != "minecraft:ladder" {
			if err := world.SetBlock(position, structureState("snow_block")); err != nil {
				return result, err
			}
		}
	}
	for _, actor := range template.Entities {
		blockPosition, _ := TransformTemplatePosition(actor.BlockPosition, pivot, TemplateMirrorNone, rotation)
		blockPosition = treeOffset(origin, blockPosition.X, blockPosition.Y, blockPosition.Z)
		if !structureContains(clip, blockPosition) {
			continue
		}
		data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: actor.NBT}).Value.(nbt.Object)
		delete(data, "UUID")
		precise := transformTemplateEntityPosition(actor.Position, pivot, TemplateMirrorNone, rotation)
		for axis, value := range []int32{origin.X, origin.Y, origin.Z} {
			precise[axis] += float64(value)
		}
		positionTags := make([]nbt.Tag, 3)
		for axis, value := range precise {
			positionTags[axis] = nbt.Tag{Type: nbt.Double, Value: value}
		}
		data["Pos"] = nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: positionTags}
		if angles, ok := data["Rotation"].Elements(); ok && len(angles) == 2 {
			if yaw, ok := angles[0].Value.(float32); ok {
				angles[0] = nbt.Tag{Type: nbt.Float, Value: transformTemplateEntityYaw(yaw, TemplateMirrorNone, rotation)}
			}
		}
		id, _ := data.Text("id")
		persistent, _ := data.Number("PersistenceRequired")
		result.PendingEntities = append(result.PendingEntities, PendingStructureEntity{ID: id, Position: blockPosition, Persistent: persistent != 0, Data: data})
	}
	return result, nil
}

// The original SignBlockEntity loads text components from the template and
// saves their canonical JSON representation. The three igloo assets contain
// literal components only; preserve other JSON values if assets change.
func normalizeIglooSignNBT(data nbt.Object) error {
	for _, field := range []string{"front_text", "back_text"} {
		text, ok := data[field].Object()
		if !ok {
			continue
		}
		for _, listField := range []string{"messages", "filtered_messages"} {
			messages, ok := text[listField].Elements()
			if !ok {
				continue
			}
			for index, message := range messages {
				encoded, ok := message.Value.(string)
				if !ok {
					return fmt.Errorf("igloo sign has a non-string text component")
				}
				var component map[string]json.RawMessage
				if err := json.Unmarshal([]byte(encoded), &component); err != nil {
					return fmt.Errorf("decode igloo sign text: %w", err)
				}
				if raw, ok := component["text"]; ok && len(component) == 1 {
					var literal string
					if err := json.Unmarshal(raw, &literal); err != nil {
						return err
					}
					var canonical bytes.Buffer
					encoder := json.NewEncoder(&canonical)
					encoder.SetEscapeHTML(false)
					if err := encoder.Encode(literal); err != nil {
						return err
					}
					messages[index] = nbt.Tag{Type: nbt.String, Value: string(bytes.TrimSuffix(canonical.Bytes(), []byte{'\n'}))}
				}
			}
		}
		if reflect.DeepEqual(text["messages"], text["filtered_messages"]) {
			delete(text, "filtered_messages")
		}
	}
	if _, ok := data["is_waxed"]; !ok {
		data["is_waxed"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	}
	return nil
}
