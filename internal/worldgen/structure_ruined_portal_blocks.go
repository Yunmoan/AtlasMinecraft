package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"fmt"
	"strings"
)

type RuinedPortalPieceRandom interface {
	StructurePieceRandom
	NextDouble() float64
}

var templateMirrorNames = [...]string{"NONE", "LEFT_RIGHT", "FRONT_BACK"}

func ruinedPortalPropertiesNBT(properties RuinedPortalProperties) nbt.Object {
	result := nbt.Object{"mossiness": {Type: nbt.Float, Value: properties.Mossiness}}
	for name, value := range map[string]bool{"cold": properties.Cold, "air_pocket": properties.AirPocket, "overgrown": properties.Overgrown, "vines": properties.Vines, "replace_with_blackstone": properties.ReplaceWithBlackstone} {
		bit := int8(0)
		if value {
			bit = 1
		}
		result[name] = nbt.Tag{Type: nbt.Byte, Value: bit}
	}
	return result
}

// NewRuinedPortalTemplatePiece constructs the original template piece and
// saved fields. Terrain/biome start selection is a separate earlier stage.
func NewRuinedPortalTemplatePiece(name string, origin FeaturePosition, placement RuinedPortalVerticalPlacement, properties RuinedPortalProperties, mirror TemplateMirror, rotation TemplateRotation) (*StructurePiece, error) {
	if mirror > TemplateMirrorFrontBack || rotation > TemplateRotationCounterclockwise90 {
		return nil, fmt.Errorf("invalid ruined portal mirror/rotation")
	}
	switch placement {
	case RuinedPortalOnLandSurface, RuinedPortalPartlyBuried, RuinedPortalOnOceanFloor, RuinedPortalInMountain, RuinedPortalUnderground, RuinedPortalInNether:
	default:
		return nil, fmt.Errorf("invalid ruined portal vertical placement %q", placement)
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:ruined_portal/" + strings.TrimPrefix(name, "ruined_portal/")
	}
	template, err := LoadRuinedPortalTemplate(name)
	if err != nil {
		return nil, err
	}
	box, err := template.BoundingBox(origin, FeaturePosition{template.Size.X / 2, 0, template.Size.Z / 2}, mirror, rotation)
	if err != nil {
		return nil, err
	}
	return &StructurePiece{ID: "minecraft:rupo", Orientation: "north", Box: box, Data: nbt.Object{
		"Template": {Type: nbt.String, Value: name}, "TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z},
		"Rotation": {Type: nbt.String, Value: templateRotationNames[rotation]}, "Mirror": {Type: nbt.String, Value: templateMirrorNames[mirror]},
		"VerticalPlacement": {Type: nbt.String, Value: string(placement)}, "Properties": {Type: nbt.Compound, Value: ruinedPortalPropertiesNBT(properties)},
	}}, nil
}

func ruinedPortalSavedSettings(piece *StructurePiece) (string, FeaturePosition, RuinedPortalVerticalPlacement, RuinedPortalProperties, TemplateMirror, TemplateRotation, error) {
	name, _ := piece.Data.Text("Template")
	placement, _ := piece.Data.Text("VerticalPlacement")
	props, _ := piece.Data.Object("Properties")
	var origin FeaturePosition
	for index, key := range []string{"TPX", "TPY", "TPZ"} {
		value, ok := piece.Data.Number(key)
		if !ok {
			return "", origin, "", RuinedPortalProperties{}, 0, 0, fmt.Errorf("ruined portal piece lacks %s", key)
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
	properties := RuinedPortalProperties{}
	if props == nil {
		return "", origin, "", properties, 0, 0, fmt.Errorf("ruined portal piece lacks Properties")
	}
	mossiness, ok := props["mossiness"].Value.(float32)
	if !ok {
		return "", origin, "", properties, 0, 0, fmt.Errorf("invalid ruined portal mossiness")
	}
	properties.Mossiness = mossiness
	for key, target := range map[string]*bool{"cold": &properties.Cold, "air_pocket": &properties.AirPocket, "overgrown": &properties.Overgrown, "vines": &properties.Vines, "replace_with_blackstone": &properties.ReplaceWithBlackstone} {
		value, ok := props.Number(key)
		if !ok {
			return "", origin, "", properties, 0, 0, fmt.Errorf("ruined portal properties lack %s", key)
		}
		*target = value != 0
	}
	mirrorText, _ := piece.Data.Text("Mirror")
	rotationText, _ := piece.Data.Text("Rotation")
	mirror, rotation := TemplateMirror(255), TemplateRotation(255)
	for index, text := range templateMirrorNames {
		if mirrorText == text {
			mirror = TemplateMirror(index)
		}
	}
	for index, text := range templateRotationNames {
		if rotationText == text {
			rotation = TemplateRotation(index)
		}
	}
	if mirror > TemplateMirrorFrontBack || rotation > TemplateRotationCounterclockwise90 {
		return "", origin, "", properties, 0, 0, fmt.Errorf("invalid saved ruined portal transform")
	}
	return name, origin, RuinedPortalVerticalPlacement(placement), properties, mirror, rotation, nil
}

func ruinedPortalHeight(world FeatureBlockWorld, placement RuinedPortalVerticalPlacement, x, z int32) (int32, error) {
	kind := "WORLD_SURFACE_WG"
	if placement == RuinedPortalOnOceanFloor {
		kind = "OCEAN_FLOOR_WG"
	}
	if source, ok := world.(interface {
		HeightmapAt(string, int32, int32) (int32, error)
	}); ok {
		height, err := source.HeightmapAt(kind, x, z)
		return height - 1, err
	}
	for y := int32(319); y >= -64; y-- {
		state, err := world.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return 0, err
		}
		occupied := !surfaceAir(state)
		if placement == RuinedPortalOnOceanFloor {
			occupied, err = oceanFloorSolid(state)
			if err != nil {
				return 0, err
			}
		}
		if occupied {
			return y, nil
		}
	}
	return -65, nil
}

// PostProcessRuinedPortalBlocks uses the original center-chunk gate, then
// expands the template clip to the whole portal. Spread/drips/vines may also
// write outside that box, as in the original piece method.
func PostProcessRuinedPortalBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random RuinedPortalPieceRandom, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:rupo" || random == nil {
		return result, fmt.Errorf("ruined portal requires piece, world and decoration random")
	}
	name, origin, placement, properties, mirror, rotation, err := ruinedPortalSavedSettings(piece)
	if err != nil {
		return result, err
	}
	template, err := LoadRuinedPortalTemplate(name)
	if err != nil {
		return result, err
	}
	pivot := FeaturePosition{template.Size.X / 2, 0, template.Size.Z / 2}
	box, err := template.BoundingBox(origin, pivot, mirror, rotation)
	if err != nil {
		return result, err
	}
	center := FeaturePosition{box.MinX + (box.MaxX-box.MinX+1)/2, box.MinY + (box.MaxY-box.MinY+1)/2, box.MinZ + (box.MaxZ-box.MinZ+1)/2}
	clip := StructureBoundingBox{chunkX * 16, -64, chunkZ * 16, chunkX*16 + 15, 319, chunkZ*16 + 15}
	if !structureContains(clip, center) {
		return result, nil
	}
	piece.Box = box
	blocks, err := ProcessRuinedPortalTemplateBlocks(world, template, origin, placement, properties, mirror, rotation)
	if err != nil {
		return result, err
	}
	placed := make([]FeaturePosition, 0, len(blocks))
	var pendingLiquid []FeaturePosition
	templateLiquidSources := make(map[FeaturePosition]bool)
	for _, block := range blocks {
		p := block.Position
		old, err := world.BlockAt(p)
		if err != nil {
			return result, err
		}
		state, err := TransformTemplateBlockState(block.State, mirror, rotation)
		if err != nil {
			return result, err
		}
		if _, source := ruinedPortalFluid(state); source {
			templateLiquidSources[p] = true
		}
		if old.Name == "minecraft:water" || old.Properties["waterlogged"] == "true" {
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				return result, err
			}
			full, err := blockstate.BlockStateByID(id)
			if err != nil {
				return result, err
			}
			if _, ok := full.Properties["waterlogged"]; ok && full.Properties["type"] != "double" {
				state = stateWithProperty(state, "waterlogged", "true")
			}
		}
		if err := world.SetBlock(p, state); err != nil {
			return result, err
		}
		placed = append(placed, p)
		if _, container := state.Properties["waterlogged"]; container && state.Properties["type"] != "double" {
			if _, source := ruinedPortalFluid(state); !source {
				pendingLiquid = append(pendingLiquid, p)
			}
		}
		if block.NBT != nil && state.Name == "minecraft:chest" {
			data := block.NBT
			data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(random.NextLong())}
			if _, ok := data["LootTable"]; ok {
				delete(data, "Items")
			}
			if err := world.SetBlockEntity(p, state.Name, data); err != nil {
				return result, err
			}
		}
	}
	// StructureTemplate fills empty liquid containers from surrounding source
	// fluids, repeating until no further container can be reached. A source
	// directly placed by the template is excluded from this propagation.
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
				fluidKind, fluidSource := ruinedPortalFluid(neighbor)
				if fluidSource && !templateLiquidSources[q] {
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
	if err := finalizeStructureTemplateShape(world, placed, ruinedPortalTemplateUpdateShape); err != nil {
		return result, err
	}
	// TemplateStructurePiece replaces jigsaws from the source palette even
	// when a processor rejected the block that originally occupied that point.
	for _, block := range template.Palettes[0] {
		if block.State.Name != "minecraft:jigsaw" || block.NBT == nil {
			continue
		}
		p, _ := TransformTemplatePosition(block.Position, pivot, mirror, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		text, _ := block.NBT.Text("final_state")
		state, err := ruinedPortalFinalState(text)
		if err != nil {
			return result, err
		}
		if err := world.SetBlock(p, state); err != nil {
			return result, err
		}
	}
	placeNetherrack := func(p FeaturePosition) error {
		state := structureState("netherrack")
		if !properties.Cold && random.NextFloat() < 0.07 {
			state = structureState("magma_block")
		}
		return world.SetBlock(p, state)
	}
	addLeaves := func(p FeaturePosition) error {
		if random.NextFloat() >= 0.5 {
			return nil
		}
		state, err := world.BlockAt(p)
		if err != nil {
			return err
		}
		if state.Name != "minecraft:netherrack" {
			return nil
		}
		above := treeOffset(p, 0, 1, 0)
		state, err = world.BlockAt(above)
		if err != nil {
			return err
		}
		if surfaceAir(state) {
			return world.SetBlock(above, stateWithProperty(structureState("jungle_leaves"), "persistent", "true"))
		}
		return nil
	}
	drip := func(p FeaturePosition) error {
		if err := placeNetherrack(p); err != nil {
			return err
		}
		for remaining := 8; remaining > 0 && random.NextFloat() < 0.5; remaining-- {
			p.Y--
			if err := placeNetherrack(p); err != nil {
				return err
			}
		}
		return nil
	}
	probabilities := [...]float32{1, 1, 1, 1, 1, 1, 1, 0.9, 0.9, 0.8, 0.7, 0.6, 0.4, 0.2}
	span := (box.MaxX - box.MinX + 1 + box.MaxZ - box.MinZ + 1) / 2
	distanceOffset := random.NextIntBound(max(1, 8-span/2))
	for x := center.X - 14; x <= center.X+14; x++ {
		for z := center.Z - 14; z <= center.Z+14; z++ {
			distance := max(0, abs32(x-center.X)+abs32(z-center.Z)+distanceOffset)
			if distance >= 14 || random.NextDouble() >= float64(probabilities[distance]) {
				continue
			}
			y, err := ruinedPortalHeight(world, placement, x, z)
			if err != nil {
				return result, err
			}
			if placement != RuinedPortalOnLandSurface && placement != RuinedPortalOnOceanFloor {
				y = min(box.MinY, y)
			}
			p := FeaturePosition{x, y, z}
			if abs32(y-box.MinY) > 3 {
				continue
			}
			state, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			flags, err := ruinedPortalStateFlags(state)
			if err != nil {
				return result, err
			}
			if state.Name == "minecraft:air" || state.Name == "minecraft:obsidian" || flags&1 != 0 || (placement != RuinedPortalInNether && state.Name == "minecraft:lava") {
				continue
			}
			if err := placeNetherrack(p); err != nil {
				return result, err
			}
			if properties.Overgrown {
				if err := addLeaves(p); err != nil {
					return result, err
				}
			}
			if err := drip(treeOffset(p, 0, -1, 0)); err != nil {
				return result, err
			}
		}
	}
	for x := box.MinX + 1; x < box.MaxX; x++ {
		for z := box.MinZ + 1; z < box.MaxZ; z++ {
			p := FeaturePosition{x, box.MinY, z}
			state, err := world.BlockAt(p)
			if err != nil {
				return result, err
			}
			if state.Name == "minecraft:netherrack" {
				if err := drip(treeOffset(p, 0, -1, 0)); err != nil {
					return result, err
				}
			}
		}
	}
	if properties.Vines || properties.Overgrown {
		// BlockPos.betweenClosed iterates X fastest, then Y, then Z.
		for z := box.MinZ; z <= box.MaxZ; z++ {
			for y := box.MinY; y <= box.MaxY; y++ {
				for x := box.MinX; x <= box.MaxX; x++ {
					p := FeaturePosition{x, y, z}
					if properties.Vines {
						state, err := world.BlockAt(p)
						if err != nil {
							return result, err
						}
						if !surfaceAir(state) && state.Name != "minecraft:vine" {
							direction := []string{"north", "east", "south", "west"}[random.NextIntBound(4)]
							q := iglooDirectionOffset(p, direction)
							neighbor, err := world.BlockAt(q)
							if err != nil {
								return result, err
							}
							if surfaceAir(neighbor) {
								flags, err := ruinedPortalStateFlags(state)
								if err != nil {
									return result, err
								}
								face := map[string]int{"north": 2, "south": 3, "west": 4, "east": 5}[direction]
								if flags&(1<<(6+face)) != 0 {
									if err := world.SetBlock(q, stateWithProperty(structureState("vine"), iglooOpposite(direction), "true")); err != nil {
										return result, err
									}
								}
							}
						}
					}
					if properties.Overgrown {
						if err := addLeaves(p); err != nil {
							return result, err
						}
					}
				}
			}
		}
	}
	return result, nil
}

func ruinedPortalFluid(state SurfaceBlockState) (string, bool) {
	if state.Name == "minecraft:water" || state.Name == "minecraft:lava" {
		return strings.TrimPrefix(state.Name, "minecraft:"), state.Properties["level"] == "" || state.Properties["level"] == "0"
	}
	if state.Properties["waterlogged"] == "true" {
		return "water", true
	}
	return "", false
}

func ruinedPortalFinalState(text string) (SurfaceBlockState, error) {
	state := SurfaceBlockState{Name: text}
	if index := strings.IndexByte(text, '['); index >= 0 {
		if !strings.HasSuffix(text, "]") {
			return state, fmt.Errorf("invalid jigsaw final state %q", text)
		}
		state.Name = text[:index]
		state.Properties = make(map[string]string)
		for _, property := range strings.Split(text[index+1:len(text)-1], ",") {
			key, value, ok := strings.Cut(property, "=")
			if !ok {
				return state, fmt.Errorf("invalid jigsaw property %q", property)
			}
			state.Properties[key] = value
		}
	}
	if _, err := blockstate.BlockStateID(state.Name, state.Properties); err != nil {
		return state, err
	}
	return state, nil
}
