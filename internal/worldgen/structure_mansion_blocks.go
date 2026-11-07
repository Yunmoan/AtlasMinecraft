package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"strings"
)

// PostProcessMansionBlocks places one original template piece. Marker actors
// consume the region stream during placement, so successive allay group sizes
// see preceding Mob.finalizeSpawn draws in the original order.
func PostProcessMansionBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, regionRandom *Xoroshiro, cx, cz int32) (StructurePiecePostprocessResult, error) {
	minY, maxY := int32(-64), int32(320)
	if bounds, ok := world.(interface {
		MinBuildHeight() int32
		MaxBuildHeight() int32
	}); ok {
		minY, maxY = bounds.MinBuildHeight(), bounds.MaxBuildHeight()
	}
	clip := StructureBoundingBox{cx * 16, minY, cz * 16, cx*16 + 15, maxY - 1, cz*16 + 15}
	return placeMansionTemplateBlocks(world, piece, random, regionRandom, clip)
}

func placeMansionTemplateBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, regionRandom *Xoroshiro, clip StructureBoundingBox) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:wmp" || random == nil || regionRandom == nil {
		return result, fmt.Errorf("mansion placement requires a saved wmp piece, world and both random streams")
	}
	name, ok := piece.Data.Text("Template")
	if !ok {
		return result, fmt.Errorf("mansion piece lacks Template")
	}
	template, err := LoadMansionTemplate(name)
	if err != nil {
		return result, err
	}
	rotation, mirror := TemplateRotation(255), TemplateMirror(255)
	text, _ := piece.Data.Text("Rot")
	for i, n := range templateRotationNames {
		if n == text {
			rotation = TemplateRotation(i)
		}
	}
	text, _ = piece.Data.Text("Mi")
	for i, n := range []string{"NONE", "LEFT_RIGHT", "FRONT_BACK"} {
		if n == text {
			mirror = TemplateMirror(i)
		}
	}
	if rotation > 3 || mirror > 2 {
		return result, fmt.Errorf("invalid mansion rotation or mirror")
	}
	for _, key := range []string{"TPX", "TPY", "TPZ"} {
		if _, ok := piece.Data.Number(key); !ok {
			return result, fmt.Errorf("mansion piece lacks %s", key)
		}
	}
	origin := endCityOrigin(piece)
	piece.Box, err = template.BoundingBox(origin, FeaturePosition{}, mirror, rotation)
	if err != nil {
		return result, err
	}
	positionRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	palette := template.Palettes[positionRandom.NextIntBound(int32(len(template.Palettes)))]
	var placed, pendingLiquid []FeaturePosition
	for _, block := range palette {
		if block.State.Name == "minecraft:structure_block" {
			continue
		}
		p, _ := TransformTemplatePosition(block.Position, FeaturePosition{}, mirror, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		if !structureContains(clip, p) || !world.InWorld(p) {
			continue
		}
		state, err := TransformTemplateBlockState(block.State, mirror, rotation)
		if err != nil {
			return result, err
		}
		old, err := world.BlockAt(p)
		if err != nil {
			return result, err
		}
		_, container := state.Properties["waterlogged"]
		container = container && state.Properties["type"] != "double"
		if container && (old.Name == "minecraft:water" || old.Properties["waterlogged"] == "true") {
			state = jigsawWaterloggedState(state)
		}
		if block.NBT != nil {
			if err := world.SetBlock(p, structureState("barrier")); err != nil {
				return result, err
			}
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
		if block.NBT != nil {
			data, err := mansionBlockEntity(state, block.NBT, random)
			if err != nil {
				return result, err
			}
			if data != nil {
				if err := world.SetBlockEntity(p, state.Name, data); err != nil {
					return result, err
				}
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
				neighbor, err := world.BlockAt(iglooDirectionOffset(p, direction))
				if err != nil {
					return result, err
				}
				nk, ns := ruinedPortalFluid(neighbor)
				if ns {
					kind, source = nk, ns
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
	if err := finalizeStructureTemplateShape(world, placed, mansionTemplateUpdateShape); err != nil {
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
		marker, _ := block.NBT.Text("metadata")
		p, _ := TransformTemplatePosition(block.Position, FeaturePosition{}, mirror, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		if !structureContains(clip, p) {
			continue
		}
		if strings.HasPrefix(marker, "Chest") {
			old, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			if old.Name == "minecraft:chest" {
				continue
			}
			state := structureState("chest")
			facing := "north"
			switch marker {
			case "ChestWest":
				facing = "west"
			case "ChestEast":
				facing = "east"
			case "ChestSouth":
				facing = "south"
			case "ChestNorth":
				facing = "north"
			}
			state = stateWithProperty(state, "facing", facing)
			state, err = TransformTemplateBlockState(state, TemplateMirrorNone, rotation)
			if err != nil {
				return result, err
			}
			if err := world.SetBlock(p, state); err != nil {
				return result, err
			}
			data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:chest"}, "LootTable": {Type: nbt.String, Value: "minecraft:chests/woodland_mansion"}}
			seed := int64(random.NextLong())
			if seed != 0 {
				data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: seed}
			}
			if err := world.SetBlockEntity(p, state.Name, data); err != nil {
				return result, err
			}
			continue
		}
		id, count := "", int32(1)
		switch marker {
		case "Mage":
			id = "minecraft:evoker"
		case "Warrior":
			id = "minecraft:vindicator"
		case "Group of Allays":
			id = "minecraft:allay"
			count = regionRandom.NextIntBound(3) + 1
		default:
			continue
		}
		for i := int32(0); i < count; i++ {
			request := PendingStructureEntity{ID: id, Position: p, Persistent: true, SavedFlag: fmt.Sprintf("Mansion:%s:%d,%d,%d:%s:%d,%d,%d:%d", name, origin.X, origin.Y, origin.Z, marker, p.X, p.Y, p.Z, i)}
			request.Data, err = FinalizeMansionEntity(request, regionRandom)
			if err != nil {
				return result, err
			}
			result.PendingEntities = append(result.PendingEntities, request)
			if err := world.SetBlock(p, structureState("air")); err != nil {
				return result, err
			}
		}
	}
	return result, nil
}

// PostProcessMansionFoundation reproduces WoodlandMansionStructure.afterPlace:
// only non-air columns inside a piece at the building's lowest Y get supports.
func PostProcessMansionFoundation(world FeatureBlockWorld, start *StructureStart, clip StructureBoundingBox) error {
	if world == nil || start == nil || start.Structure != "minecraft:mansion" {
		return fmt.Errorf("mansion foundation requires a world and mansion start")
	}
	box, ok := start.BoundingBox()
	if !ok {
		return nil
	}
	minY := int32(-64)
	if bounds, ok := world.(interface{ MinBuildHeight() int32 }); ok {
		minY = bounds.MinBuildHeight()
	}
	for x := clip.MinX; x <= clip.MaxX; x++ {
		for z := clip.MinZ; z <= clip.MaxZ; z++ {
			p := FeaturePosition{x, box.MinY, z}
			if !structureContains(box, p) {
				continue
			}
			state, err := world.BlockAt(p)
			if err != nil {
				return err
			}
			if surfaceAir(state) {
				continue
			}
			inside := false
			for _, piece := range start.Pieces {
				if structureContains(piece.Box, p) {
					inside = true
					break
				}
			}
			if !inside {
				continue
			}
			for p.Y--; p.Y > minY; p.Y-- {
				state, err := world.BlockAt(p)
				if err != nil {
					return err
				}
				if !surfaceAir(state) && state.Name != "minecraft:water" && state.Name != "minecraft:lava" {
					break
				}
				if err := world.SetBlock(p, structureState("cobblestone")); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func mansionTemplateUpdateShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	if state.Name == "minecraft:chest" || state.Name == "minecraft:trapped_chest" {
		updated, err := shipwreckTemplateUpdateShape(world, p, state, direction, neighbor)
		if err != nil {
			return state, err
		}
		connected := func(s SurfaceBlockState) string {
			if s.Properties["type"] == "left" {
				return iglooClockwise(s.Properties["facing"])
			}
			return iglooCounterclockwise(s.Properties["facing"])
		}
		if neighbor.Name == state.Name && direction != "up" && direction != "down" {
			if state.Properties["type"] == "single" && neighbor.Properties["type"] != "single" && state.Properties["facing"] == neighbor.Properties["facing"] && connected(neighbor) == iglooOpposite(direction) {
				kind := "left"
				if neighbor.Properties["type"] == "left" {
					kind = "right"
				}
				return stateWithProperty(updated, "type", kind), nil
			}
		} else if connected(state) == direction {
			return stateWithProperty(updated, "type", "single"), nil
		}
		return updated, nil
	}
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
	if state.Name == "minecraft:brown_mushroom" || state.Name == "minecraft:red_mushroom" {
		below, err := world.BlockAt(treeOffset(p, 0, -1, 0))
		if err != nil {
			return state, err
		}
		survives, err := featureMushroomSurvives(world, p, below)
		if err != nil {
			return state, err
		}
		if !survives {
			return structureState("air"), nil
		}
	}
	return endCityTemplateUpdateShape(world, p, state, direction, neighbor)
}
