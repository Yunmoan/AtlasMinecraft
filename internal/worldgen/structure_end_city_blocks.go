package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

//go:embed end_city_actor_defaults_1.21.1.txt
var endCityActorDefaults string

var loadEndCityActorDefaults = sync.OnceValues(func() (map[string]nbt.Object, error) {
	result := make(map[string]nbt.Object)
	for _, line := range strings.Split(endCityActorDefaults, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		encoded, err := hex.DecodeString(fields[1])
		if err != nil {
			return nil, err
		}
		tag, size, err := nbt.ParseNetwork(encoded)
		if err != nil || size != len(encoded) {
			return nil, fmt.Errorf("invalid End City actor default %s: %v", fields[0], err)
		}
		data, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("invalid End City actor compound")
		}
		result[fields[0]] = data
	}
	if len(result) != 2 {
		return nil, fmt.Errorf("incomplete original End City actor defaults")
	}
	return result, nil
})

func endCityMarkerActor(marker string, p FeaturePosition, rotation TemplateRotation) (PendingStructureEntity, error) {
	id := "minecraft:shulker"
	if strings.HasPrefix(marker, "Elytra") {
		id = "minecraft:item_frame"
	}
	defaults, err := loadEndCityActorDefaults()
	if err != nil {
		return PendingStructureEntity{}, err
	}
	data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: defaults[id]}).Value.(nbt.Object)
	delete(data, "UUID")
	position := [3]float64{float64(p.X) + .5, float64(p.Y), float64(p.Z) + .5}
	if id == "minecraft:item_frame" {
		facing := [...]int8{3, 4, 2, 5}[rotation]
		position[1] += .5
		switch facing {
		case 3:
			position[2] -= .46875
		case 4:
			position[0] += .46875
		case 2:
			position[2] += .46875
		case 5:
			position[0] -= .46875
		}
		data["Facing"] = nbt.Tag{Type: nbt.Byte, Value: facing}
		data["TileX"] = nbt.Tag{Type: nbt.Int, Value: p.X}
		data["TileY"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
		data["TileZ"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
		data["Rotation"] = nbt.Tag{Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{{Type: nbt.Float, Value: float32(rotation) * 90}, {Type: nbt.Float, Value: float32(0)}}}
		data["Item"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:elytra"}, "count": {Type: nbt.Int, Value: int32(1)}}}
		data["ItemRotation"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		data["ItemDropChance"] = nbt.Tag{Type: nbt.Float, Value: float32(1)}
	}
	data["Pos"] = nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: []nbt.Tag{{Type: nbt.Double, Value: position[0]}, {Type: nbt.Double, Value: position[1]}, {Type: nbt.Double, Value: position[2]}}}
	return PendingStructureEntity{ID: id, Position: p, Data: data}, nil
}

// PostProcessEndCityBlocks retains original palette order, air overwrite,
// template load/loot random draws and the Chest/Sentry/Elytra data markers.
// EndCityPiece ignores template entities; marker actors are created directly
// without Mob.finalizeSpawn or persistence flags in the original implementation.
func PostProcessEndCityBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, cx, cz int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:ecp" || random == nil {
		return result, fmt.Errorf("End City requires an ecp piece, world and decoration random")
	}
	name, ok := piece.Data.Text("Template")
	if !ok {
		return result, fmt.Errorf("End City piece lacks Template")
	}
	template, err := LoadEndCityTemplate(name)
	if err != nil {
		return result, err
	}
	text, _ := piece.Data.Text("Rot")
	rotation := TemplateRotation(255)
	for i, name := range templateRotationNames {
		if text == name {
			rotation = TemplateRotation(i)
		}
	}
	if rotation > TemplateRotationCounterclockwise90 {
		return result, fmt.Errorf("invalid End City rotation %q", text)
	}
	for _, key := range []string{"TPX", "TPY", "TPZ", "OW"} {
		if _, ok := piece.Data.Number(key); !ok {
			return result, fmt.Errorf("End City piece lacks %s", key)
		}
	}
	origin := endCityOrigin(piece)
	overwrite, _ := piece.Data.Number("OW")
	piece.Box, err = template.BoundingBox(origin, FeaturePosition{}, TemplateMirrorNone, rotation)
	if err != nil {
		return result, err
	}
	minY, maxY := int32(0), int32(256)
	if bounds, ok := world.(interface {
		MinBuildHeight() int32
		MaxBuildHeight() int32
	}); ok {
		minY, maxY = bounds.MinBuildHeight(), bounds.MaxBuildHeight()
	}
	clip := StructureBoundingBox{cx * 16, minY, cz * 16, cx*16 + 15, maxY - 1, cz*16 + 15}
	positionRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	palette := template.Palettes[positionRandom.NextIntBound(int32(len(template.Palettes)))]
	var placed, pendingLiquid []FeaturePosition
	entities := make(map[FeaturePosition]nbt.Object)
	for _, block := range palette {
		if block.State.Name == "minecraft:structure_block" || overwrite == 0 && surfaceAir(block.State) {
			continue
		}
		p, _ := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		if !structureContains(clip, p) || !world.InWorld(p) {
			continue
		}
		state, err := TransformTemplateBlockState(block.State, TemplateMirrorNone, rotation)
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
			data, err := endCityBlockEntity(state, block.NBT, random)
			if err != nil {
				return result, err
			}
			if data != nil {
				if err := world.SetBlockEntity(p, state.Name, data); err != nil {
					return result, err
				}
				entities[p] = data
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
	if err := finalizeStructureTemplateShape(world, placed, endCityTemplateUpdateShape); err != nil {
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
		p, _ := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		if strings.HasPrefix(marker, "Chest") {
			p.Y--
			if !structureContains(clip, p) {
				continue
			}
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
			data["LootTable"] = nbt.Tag{Type: nbt.String, Value: "minecraft:chests/end_city_treasure"}
			seed := int64(random.NextLong())
			if seed != 0 {
				data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: seed}
			} else {
				delete(data, "LootTableSeed")
			}
			if err := world.SetBlockEntity(p, state.Name, data); err != nil {
				return result, err
			}
			entities[p] = data
		} else if structureContains(clip, p) && p.Y >= -20000000 && p.Y < 20000000 && (strings.HasPrefix(marker, "Sentry") || strings.HasPrefix(marker, "Elytra")) {
			actor, err := endCityMarkerActor(marker, p, rotation)
			if err != nil {
				return result, err
			}
			actor.SavedFlag = fmt.Sprintf("%s:%d,%d,%d:%s", name, origin.X, origin.Y, origin.Z, marker)
			result.PendingEntities = append(result.PendingEntities, actor)
		}
	}
	return result, nil
}

func endCityBlockEntity(state SurfaceBlockState, input nbt.Object, random StructurePieceRandom) (nbt.Object, error) {
	if state.Name == "minecraft:magenta_wall_banner" {
		return cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: input}).Value.(nbt.Object), nil
	}
	if state.Name == "minecraft:ender_chest" {
		return nbt.Object{"id": {Type: nbt.String, Value: "minecraft:ender_chest"}}, nil
	}
	if state.Name == "minecraft:dragon_wall_head" {
		if _, ok := input["profile"]; ok {
			return nil, fmt.Errorf("unexpected End City skull profile")
		}
		return nbt.Object{"id": {Type: nbt.String, Value: "minecraft:skull"}}, nil
	}
	if state.Name == "minecraft:brewing_stand" {
		data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: input}).Value.(nbt.Object)
		for _, key := range []string{"x", "y", "z"} {
			delete(data, key)
		}
		data["id"] = nbt.Tag{Type: nbt.String, Value: "minecraft:brewing_stand"}
		time, _ := data.Number("BrewTime")
		fuel, _ := data.Number("Fuel")
		data["BrewTime"] = nbt.Tag{Type: nbt.Short, Value: int16(time)}
		data["Fuel"] = nbt.Tag{Type: nbt.Byte, Value: int8(fuel)}
		if _, ok := data["Items"]; !ok {
			data["Items"] = nbt.Tag{Type: nbt.List, ListType: nbt.End, Value: []nbt.Tag{}}
		}
		return data, nil
	}
	return jigsawPlacedBlockEntity(state, input, random)
}

func endCityTemplateUpdateShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState, direction string, neighbor SurfaceBlockState) (SurfaceBlockState, error) {
	if strings.HasSuffix(state.Name, "_wall_banner") && direction == iglooOpposite(state.Properties["facing"]) {
		rules, err := iglooBlockRules(neighbor)
		if err != nil {
			return state, err
		}
		if !rules.Solid {
			return structureState("air"), nil
		}
	}
	return shipwreckTemplateUpdateShape(world, p, state, direction, neighbor)
}
