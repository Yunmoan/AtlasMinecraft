package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"fmt"
	"strings"
)

func oceanRuinTemplateUpdateShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	if strings.HasSuffix(state.Name, "_wall") {
		if state.Properties["waterlogged"] == "true" {
			if scheduler, ok := world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
				if err := scheduler.ScheduleFluidUpdate(p); err != nil {
					return state, err
				}
			}
		}
		return ruinedPortalTemplateUpdateShape(world, p, state, direction, neighbor)
	}
	return shipwreckTemplateUpdateShape(world, p, state, direction, neighbor)
}

// PostProcessOceanRuinBlocks generates the exact supported block stage and
// returns persistent drowned requests for the separate runtime actor stage.
func PostProcessOceanRuinBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, worldSeed int64, cx, cz int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:orp" || random == nil {
		return result, fmt.Errorf("ocean ruin requires piece, world and decoration random")
	}
	name, ok := piece.Data.Text("Template")
	if !ok {
		return result, fmt.Errorf("ocean ruin lacks Template")
	}
	template, err := LoadOceanRuinTemplate(name)
	if err != nil {
		return result, err
	}
	rotationText, _ := piece.Data.Text("Rot")
	rotation := TemplateRotation(255)
	for index, text := range templateRotationNames {
		if text == rotationText {
			rotation = TemplateRotation(index)
		}
	}
	if rotation > TemplateRotationCounterclockwise90 {
		return result, fmt.Errorf("invalid ocean ruin rotation")
	}
	integrity, ok := piece.Data["Integrity"].Value.(float32)
	if !ok {
		return result, fmt.Errorf("ocean ruin lacks Integrity")
	}
	kind, ok := piece.Data.Text("BiomeType")
	if !ok || kind != "COLD" && kind != "WARM" {
		return result, fmt.Errorf("invalid ocean ruin BiomeType")
	}
	large, _ := piece.Data.Number("IsLarge")
	var origin FeaturePosition
	for index, key := range []string{"TPX", "TPY", "TPZ"} {
		value, ok := piece.Data.Number(key)
		if !ok {
			return result, fmt.Errorf("ocean ruin lacks %s", key)
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
	origin.Y, err = structureTemplateHeight(world, "OCEAN_FLOOR_WG", origin.X, origin.Z)
	if err != nil {
		return result, err
	}
	corner, _ := TransformTemplatePosition(FeaturePosition{template.Size.X - 1, 0, template.Size.Z - 1}, FeaturePosition{}, TemplateMirrorNone, rotation)
	corner = treeOffset(origin, corner.X, corner.Y, corner.Z)
	minimum, floor, deep := int32(512), origin.Y-1, int32(0)
	minY := int32(-64)
	if bounds, ok := world.(interface{ MinBuildHeight() int32 }); ok {
		minY = bounds.MinBuildHeight()
	}
	for x := min(origin.X, corner.X); x <= max(origin.X, corner.X); x++ {
		for z := min(origin.Z, corner.Z); z <= max(origin.Z, corner.Z); z++ {
			y := origin.Y - 1
			for {
				state, err := world.BlockAt(FeaturePosition{x, y, z})
				if err != nil {
					return result, err
				}
				water := state.Name == "minecraft:water" || state.Properties["waterlogged"] == "true"
				ice := state.Name == "minecraft:ice" || state.Name == "minecraft:packed_ice" || state.Name == "minecraft:blue_ice" || state.Name == "minecraft:frosted_ice"
				if !surfaceAir(state) && !water && !ice || y <= minY+1 {
					break
				}
				y--
			}
			minimum = min(minimum, y)
			if y < floor-2 {
				deep++
			}
		}
	}
	if floor-minimum > 2 && deep > abs32(origin.X-corner.X)-2 {
		origin.Y = minimum + 1
	}
	piece.Data["TPY"] = nbt.Tag{Type: nbt.Int, Value: origin.Y}
	piece.Box, err = template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	if err != nil {
		return result, err
	}
	blocks, err := ProcessOceanRuinTemplateBlocks(template, origin, rotation, integrity, kind == "COLD", worldSeed)
	if err != nil {
		return result, err
	}
	clip := StructureBoundingBox{cx * 16, -64, cz * 16, cx*16 + 15, 319, cz*16 + 15}
	var placed, pending []FeaturePosition
	for _, block := range blocks {
		p := block.Position
		if !structureContains(clip, p) || !world.InWorld(p) {
			continue
		}
		old, err := world.BlockAt(p)
		if err != nil {
			return result, err
		}
		state, err := TransformTemplateBlockState(block.State, TemplateMirrorNone, rotation)
		if err != nil {
			return result, err
		}
		id, err := blockstate.BlockStateID(state.Name, state.Properties)
		if err != nil {
			return result, err
		}
		full, err := blockstate.BlockStateByID(id)
		if err != nil {
			return result, err
		}
		_, container := full.Properties["waterlogged"]
		container = container && full.Properties["type"] != "double"
		if container && (old.Name == "minecraft:water" || old.Properties["waterlogged"] == "true") {
			state = stateWithProperty(state, "waterlogged", "true")
		}
		if err := world.SetBlock(p, state); err != nil {
			return result, err
		}
		placed = append(placed, p)
		if container {
			if _, source := ruinedPortalFluid(state); !source {
				pending = append(pending, p)
			}
		}
		if block.NBT != nil {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object)
			switch state.Name {
			case "minecraft:suspicious_sand", "minecraft:suspicious_gravel":
				data["id"] = nbt.Tag{Type: nbt.String, Value: "minecraft:brushable_block"}
			case "minecraft:chest":
				data["id"] = nbt.Tag{Type: nbt.String, Value: "minecraft:chest"}
				data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(random.NextLong())}
				if _, ok := data["LootTable"]; ok {
					delete(data, "Items")
				}
			}
			if err := world.SetBlockEntity(p, state.Name, data); err != nil {
				return result, err
			}
		}
	}
	for changed := true; changed && len(pending) > 0; {
		changed = false
		remaining := pending[:0]
		for _, p := range pending {
			state, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			kind, source := ruinedPortalFluid(state)
			for _, direction := range []string{"up", "north", "east", "south", "west"} {
				if source {
					break
				}
				neighbor, err := world.BlockAt(iglooDirectionOffset(p, direction))
				if err != nil {
					return result, err
				}
				fluidKind, fluidSource := ruinedPortalFluid(neighbor)
				if fluidSource {
					kind, source = fluidKind, true
				}
			}
			if source {
				if kind == "water" && state.Properties["waterlogged"] != "true" {
					if err := world.SetBlock(p, stateWithProperty(state, "waterlogged", "true")); err != nil {
						return result, err
					}
				}
				changed = true
			} else {
				remaining = append(remaining, p)
			}
		}
		pending = remaining
	}
	if err := finalizeStructureTemplateShape(world, placed, oceanRuinTemplateUpdateShape); err != nil {
		return result, err
	}
	paletteRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	palette := template.Palettes[paletteRandom.NextIntBound(int32(len(template.Palettes)))]
	seaLevel := int32(63)
	if source, ok := world.(interface{ GenerationSeaLevel() int32 }); ok {
		seaLevel = source.GenerationSeaLevel()
	}
	for _, block := range palette {
		if block.State.Name != "minecraft:structure_block" || block.NBT == nil {
			continue
		}
		mode, _ := block.NBT.Text("mode")
		if mode != "DATA" {
			continue
		}
		metadata, _ := block.NBT.Text("metadata")
		p, _ := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		if !structureContains(clip, p) {
			continue
		}
		switch metadata {
		case "chest":
			old, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			state := structureState("chest")
			if old.Name == "minecraft:water" || old.Properties["waterlogged"] == "true" {
				state = stateWithProperty(state, "waterlogged", "true")
			}
			if err := world.SetBlock(p, state); err != nil {
				return result, err
			}
			loot := "minecraft:chests/underwater_ruin_small"
			if large != 0 {
				loot = "minecraft:chests/underwater_ruin_big"
			}
			if err := world.SetBlockEntity(p, state.Name, nbt.Object{"id": {Type: nbt.String, Value: "minecraft:chest"}, "LootTable": {Type: nbt.String, Value: loot}, "LootTableSeed": {Type: nbt.Long, Value: int64(random.NextLong())}}); err != nil {
				return result, err
			}
		case "drowned":
			data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:drowned"}, "PersistenceRequired": {Type: nbt.Byte, Value: int8(1)}}
			result.PendingEntities = append(result.PendingEntities, PendingStructureEntity{ID: "minecraft:drowned", Position: p, Persistent: true, Data: data})
			state := structureState("water")
			if p.Y > seaLevel {
				state = structureState("air")
			}
			if err := world.SetBlock(p, state); err != nil {
				return result, err
			}
		}
	}
	return result, nil
}
