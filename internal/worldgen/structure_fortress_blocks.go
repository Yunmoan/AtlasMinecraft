package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"fmt"
)

type fortressPieceEditor struct {
	world       StructureBlockEntityWorld
	piece       *StructurePiece
	clip        StructureBoundingBox
	minY        int32
	transformed map[uint32]SurfaceBlockState
	err         error
}

func fortressProperty(state SurfaceBlockState, key, value string) SurfaceBlockState {
	return stateWithProperty(state, key, value)
}

// PostProcessNetherFortressBlocks places all fourteen original fortress pieces,
// including clipped supports, block entities, lava ticks and saved piece flags.
func PostProcessNetherFortressBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	result := StructurePiecePostprocessResult{}
	if world == nil || piece == nil || piece.Data == nil || random == nil {
		return result, fmt.Errorf("fortress postprocess requires world, saved piece and random")
	}
	minY, maxY := int32(0), int32(256)
	if bounds, ok := world.(interface {
		MinBuildHeight() int32
		MaxBuildHeight() int32
	}); ok {
		minY, maxY = bounds.MinBuildHeight(), bounds.MaxBuildHeight()
	}
	e := &fortressPieceEditor{world: world, piece: piece, minY: minY, clip: StructureBoundingBox{chunkX * 16, minY, chunkZ * 16, chunkX*16 + 15, maxY - 1, chunkZ*16 + 15}, transformed: make(map[uint32]SurfaceBlockState)}
	switch piece.ID {
	case "minecraft:nebs":
		placeFortress_nebs(e, random)
	case "minecraft:nebcr", "minecraft:nestart":
		placeFortress_nebcr(e, random)
	case "minecraft:nerc":
		placeFortress_nerc(e, random)
	case "minecraft:nesr":
		placeFortress_nesr(e, random)
	case "minecraft:nemt":
		placeFortress_nemt(e, random)
	case "minecraft:nece":
		placeFortress_nece(e, random)
	case "minecraft:nesc":
		placeFortress_nesc(e, random)
	case "minecraft:nescsc":
		placeFortress_nescsc(e, random)
	case "minecraft:nescrt":
		placeFortress_nescrt(e, random)
	case "minecraft:nesclt":
		placeFortress_nesclt(e, random)
	case "minecraft:neccs":
		placeFortress_neccs(e, random)
	case "minecraft:nectb":
		placeFortress_nectb(e, random)
	case "minecraft:necsr":
		placeFortress_necsr(e, random)
	case "minecraft:nebef":
		placeFortress_nebef(e, random)
	default:
		return result, fmt.Errorf("unsupported fortress piece %q", piece.ID)
	}
	return result, e.err
}

func (e *fortressPieceEditor) set(state SurfaceBlockState, x, y, z int32) {
	if e.err != nil {
		return
	}
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) || !e.world.InWorld(p) {
		return
	}
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		e.err = err
		return
	}
	transformed, found := e.transformed[id]
	if !found {
		mirror, rotation := TemplateMirrorNone, TemplateRotationNone
		if e.piece.Orientation == "south" || e.piece.Orientation == "west" {
			mirror = TemplateMirrorLeftRight
		}
		if e.piece.Orientation == "east" || e.piece.Orientation == "west" {
			rotation = TemplateRotationClockwise90
		}
		transformed, e.err = TransformTemplateBlockState(state, mirror, rotation)
		if e.err != nil {
			return
		}
		e.transformed[id] = transformed
	}
	e.err = e.world.SetBlock(p, transformed)
	if e.err != nil {
		return
	}
	if featureFluid(transformed) {
		if scheduler, ok := e.world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
			e.err = scheduler.ScheduleFluidUpdate(p)
		}
	}
	if transformed.Name == "minecraft:nether_brick_fence" {
		if marker, ok := e.world.(interface{ MarkPostProcessing(FeaturePosition) error }); ok {
			e.err = marker.MarkPostProcessing(p)
		}
	}
}

func (e *fortressPieceEditor) fill(minX, minY, minZ, maxX, maxY, maxZ int32, boundary, interior SurfaceBlockState, skipAir bool) {
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if skipAir {
					p := e.piece.BlockPosition(x, y, z)
					if !structureContains(e.clip, p) {
						continue
					}
					old, err := e.world.BlockAt(p)
					if err != nil {
						e.err = err
						return
					}
					if surfaceAir(old) {
						continue
					}
				}
				state := interior
				if x == minX || x == maxX || y == minY || y == maxY || z == minZ || z == maxZ {
					state = boundary
				}
				e.set(state, x, y, z)
			}
		}
	}
}

func (e *fortressPieceEditor) fillDown(state SurfaceBlockState, x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) {
		return
	}
	for p.Y > e.minY+1 && e.world.InWorld(p) && e.err == nil {
		old, err := e.world.BlockAt(p)
		if err != nil {
			e.err = err
			return
		}
		if !surfaceAir(old) && !featureFluid(old) && old.Name != "minecraft:glow_lichen" && old.Name != "minecraft:seagrass" && old.Name != "minecraft:tall_seagrass" {
			return
		}
		e.err = e.world.SetBlock(p, state)
		p.Y--
	}
}

func (e *fortressPieceEditor) chest(random StructurePieceRandom, x, y, z int32) {
	flag, _ := e.piece.Data["Chest"].Number()
	p := e.piece.BlockPosition(x, y, z)
	if flag == 0 || !structureContains(e.clip, p) || e.err != nil {
		return
	}
	e.piece.Data["Chest"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	old, err := e.world.BlockAt(p)
	if err != nil {
		e.err = err
		return
	}
	if old.Name == "minecraft:chest" {
		return
	}
	editor := &featureEditor{world: e.world}
	state := monsterRoomChestState(editor, p)
	if editor.err != nil {
		e.err = editor.err
		return
	}
	e.err = e.world.SetBlock(p, state)
	if e.err != nil {
		return
	}
	e.err = e.world.SetBlockEntity(p, "minecraft:chest", nbt.Object{
		"id": {Type: nbt.String, Value: "minecraft:chest"}, "LootTable": {Type: nbt.String, Value: "minecraft:chests/nether_bridge"}, "LootTableSeed": {Type: nbt.Long, Value: int64(random.NextLong())},
	})
}

func (e *fortressPieceEditor) spawner(random StructurePieceRandom, x, y, z int32) {
	flag, _ := e.piece.Data["Mob"].Number()
	p := e.piece.BlockPosition(x, y, z)
	if flag != 0 || !structureContains(e.clip, p) || e.err != nil {
		return
	}
	e.piece.Data["Mob"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	e.err = e.world.SetBlock(p, structureState("spawner"))
	if e.err != nil {
		return
	}
	data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:mob_spawner"}, "SpawnData": {Type: nbt.Compound, Value: nbt.Object{"entity": {Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:blaze"}}}}}, "SpawnPotentials": {Type: nbt.List, ListType: nbt.End, Value: []nbt.Tag{}}}
	for name, value := range map[string]int16{"Delay": 20, "MinSpawnDelay": 200, "MaxSpawnDelay": 800, "SpawnCount": 4, "MaxNearbyEntities": 6, "RequiredPlayerRange": 16, "SpawnRange": 4} {
		data[name] = nbt.Tag{Type: nbt.Short, Value: value}
	}
	e.err = e.world.SetBlockEntity(p, "minecraft:spawner", data)
}

func (e *fortressPieceEditor) lavaTick(x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if structureContains(e.clip, p) && e.err == nil {
		if scheduler, ok := e.world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
			e.err = scheduler.ScheduleFluidUpdate(p)
		}
	}
}
