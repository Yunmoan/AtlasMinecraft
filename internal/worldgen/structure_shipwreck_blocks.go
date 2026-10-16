package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"fmt"
	"strings"
)

func shipwreckTemplateUpdateShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	schedule := state.Properties["waterlogged"] == "true"
	if state.Name == "minecraft:water" || state.Name == "minecraft:lava" {
		_, source := ruinedPortalFluid(state)
		_, neighborSource := ruinedPortalFluid(neighbor)
		schedule = source || neighborSource
	}
	if schedule {
		if scheduler, ok := world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
			if err := scheduler.ScheduleFluidUpdate(p); err != nil {
				return state, err
			}
		}
	}
	switch {
	case strings.HasSuffix(state.Name, "_fence"):
		if direction == "up" || direction == "down" {
			return state, nil
		}
		rules, err := iglooBlockRules(neighbor)
		if err != nil {
			return state, err
		}
		otherFence := strings.HasSuffix(neighbor.Name, "_fence") && (neighbor.Name == "minecraft:nether_brick_fence") == (state.Name == "minecraft:nether_brick_fence")
		gate := strings.HasSuffix(neighbor.Name, "_fence_gate") && iglooFacingAxis(neighbor.Properties["facing"]) != iglooFacingAxis(direction)
		connect := !rules.ConnectionException && rules.FaceSturdy[map[string]int{"north": 3, "south": 2, "west": 5, "east": 4}[direction]] || otherFence || gate
		value := "false"
		if connect {
			value = "true"
		}
		return stateWithProperty(state, direction, value), nil
	case strings.HasSuffix(state.Name, "_door"):
		lower := state.Properties["half"] == "lower"
		if direction == "up" && lower || direction == "down" && !lower {
			if !strings.HasSuffix(neighbor.Name, "_door") || neighbor.Properties["half"] == state.Properties["half"] {
				return structureState("air"), nil
			}
			return stateWithProperty(neighbor, "half", state.Properties["half"]), nil
		}
		if direction == "down" && lower {
			rules, err := iglooBlockRules(neighbor)
			if err != nil {
				return state, err
			}
			if !rules.FaceSturdy[1] {
				return structureState("air"), nil
			}
		}
		return state, nil
	default:
		return iglooTemplateUpdateShape(world, p, state, direction, neighbor)
	}
}

func PostProcessShipwreckBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:shipwreck" || random == nil {
		return result, fmt.Errorf("shipwreck requires piece, world and decoration random")
	}
	name, ok := piece.Data.Text("Template")
	if !ok {
		return result, fmt.Errorf("shipwreck lacks Template")
	}
	template, err := LoadShipwreckTemplate(name)
	if err != nil {
		return result, err
	}
	rotationText, _ := piece.Data.Text("Rot")
	rotation := TemplateRotation(255)
	for index, text := range templateRotationNames {
		if rotationText == text {
			rotation = TemplateRotation(index)
		}
	}
	if rotation > TemplateRotationCounterclockwise90 {
		return result, fmt.Errorf("invalid shipwreck rotation %q", rotationText)
	}
	var origin FeaturePosition
	for index, key := range []string{"TPX", "TPY", "TPZ"} {
		value, ok := piece.Data.Number(key)
		if !ok {
			return result, fmt.Errorf("shipwreck lacks %s", key)
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
	beachedValue, ok := piece.Data.Number("isBeached")
	if !ok {
		return result, fmt.Errorf("shipwreck lacks isBeached")
	}
	beached := beachedValue != 0
	if template.Size.X <= 32 && template.Size.Y <= 32 {
		kind := "OCEAN_FLOOR_WG"
		if beached {
			kind = "WORLD_SURFACE_WG"
		}
		minimum, total := int32(320), int32(0)
		count := template.Size.X * template.Size.Z
		for dx := int32(0); dx < template.Size.X; dx++ {
			for dz := int32(0); dz < template.Size.Z; dz++ {
				height, err := structureTemplateHeight(world, kind, origin.X+dx, origin.Z+dz)
				if err != nil {
					return result, err
				}
				minimum = min(minimum, height)
				total += height
			}
		}
		if count == 0 {
			total, err = structureTemplateHeight(world, kind, origin.X, origin.Z)
			if err != nil {
				return result, err
			}
			count = 1
		}
		origin.Y = total / count
		if beached {
			origin.Y = minimum - template.Size.Y/2 - random.NextIntBound(3)
		}
		piece.Data["TPY"] = nbt.Tag{Type: nbt.Int, Value: origin.Y}
	}
	piece.Box, err = template.BoundingBox(origin, shipwreckPivot, TemplateMirrorNone, rotation)
	if err != nil {
		return result, err
	}
	clip := StructureBoundingBox{chunkX * 16, -64, chunkZ * 16, chunkX*16 + 15, 319, chunkZ*16 + 15}
	positionRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	palette := template.Palettes[positionRandom.NextIntBound(int32(len(template.Palettes)))]
	var placed, pendingLiquid []FeaturePosition
	entities := make(map[FeaturePosition]nbt.Object)
	for _, block := range palette {
		if block.State.Name == "minecraft:structure_block" || surfaceAir(block.State) {
			continue
		}
		p, _ := TransformTemplatePosition(block.Position, shipwreckPivot, TemplateMirrorNone, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
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
				pendingLiquid = append(pendingLiquid, p)
			}
		}
		if block.NBT != nil && state.Name == "minecraft:chest" {
			// Original template loads consume a loot seed even before a data
			// marker assigns the loot table; the seed is absent from saved NBT.
			random.NextLong()
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object)
			delete(data, "LootTableSeed")
			if err := world.SetBlockEntity(p, state.Name, data); err != nil {
				return result, err
			}
			entities[p] = data
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
		pendingLiquid = remaining
	}
	if err := finalizeStructureTemplateShape(world, placed, shipwreckTemplateUpdateShape); err != nil {
		return result, err
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
		loot, ok := map[string]string{"map_chest": "minecraft:chests/shipwreck_map", "treasure_chest": "minecraft:chests/shipwreck_treasure", "supply_chest": "minecraft:chests/shipwreck_supply"}[metadata]
		if !ok {
			continue
		}
		p, _ := TransformTemplatePosition(block.Position, shipwreckPivot, TemplateMirrorNone, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		if !structureContains(clip, p) {
			continue
		}
		p.Y--
		state, err := world.BlockAt(p)
		if err != nil {
			return result, err
		}
		if state.Name != "minecraft:chest" {
			continue
		}
		data := entities[p]
		if data == nil {
			data = nbt.Object{"id": {Type: nbt.String, Value: "minecraft:chest"}}
		}
		delete(data, "Items")
		data["LootTable"] = nbt.Tag{Type: nbt.String, Value: loot}
		data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(random.NextLong())}
		if err := world.SetBlockEntity(p, state.Name, data); err != nil {
			return result, err
		}
		entities[p] = data
	}
	return result, nil
}

func structureTemplateHeight(world FeatureBlockWorld, kind string, x, z int32) (int32, error) {
	if source, ok := world.(interface {
		HeightmapAt(string, int32, int32) (int32, error)
	}); ok {
		return source.HeightmapAt(kind, x, z)
	}
	for y := int32(319); y >= -64; y-- {
		state, err := world.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return 0, err
		}
		occupied := !surfaceAir(state)
		if kind == "OCEAN_FLOOR_WG" {
			occupied, err = oceanFloorSolid(state)
			if err != nil {
				return 0, err
			}
		}
		if occupied {
			return y + 1, nil
		}
	}
	return -64, nil
}
