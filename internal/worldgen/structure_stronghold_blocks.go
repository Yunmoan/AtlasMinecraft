package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

type strongholdPieceEditor struct{ fortressPieceEditor }

// PostProcessStrongholdBlocks places every original stronghold piece using the
// chunk's structure-feature RNG, preserving authored doors, loot seeds, portal
// eyes, silverfish spawner NBT and original clipped saved flags.
func PostProcessStrongholdBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	result := StructurePiecePostprocessResult{}
	if world == nil || piece == nil || piece.Data == nil || random == nil {
		return result, fmt.Errorf("stronghold postprocess requires world, saved piece and random")
	}
	minY, maxY := int32(-64), int32(320)
	if bounds, ok := world.(interface {
		MinBuildHeight() int32
		MaxBuildHeight() int32
	}); ok {
		minY, maxY = bounds.MinBuildHeight(), bounds.MaxBuildHeight()
	}
	e := &strongholdPieceEditor{fortressPieceEditor{world: world, piece: piece, minY: minY, clip: StructureBoundingBox{chunkX * 16, minY, chunkZ * 16, chunkX*16 + 15, maxY - 1, chunkZ*16 + 15}, transformed: make(map[uint32]SurfaceBlockState)}}
	switch piece.ID {
	case "minecraft:shstart", "minecraft:shsd":
		placeStronghold_shsd(e, random)
	case "minecraft:shs":
		placeStronghold_shs(e, random)
	case "minecraft:shph":
		placeStronghold_shph(e, random)
	case "minecraft:shlt":
		placeStronghold_shlt(e, random)
	case "minecraft:shrt":
		placeStronghold_shrt(e, random)
	case "minecraft:shrc":
		placeStronghold_shrc(e, random)
	case "minecraft:shssd":
		placeStronghold_shssd(e, random)
	case "minecraft:sh5c":
		placeStronghold_sh5c(e, random)
	case "minecraft:shcc":
		placeStronghold_shcc(e, random)
	case "minecraft:shli":
		placeStronghold_shli(e, random)
	case "minecraft:shpr":
		placeStronghold_shpr(e, random)
	case "minecraft:shfc":
		placeStronghold_shfc(e, random)
	default:
		return result, fmt.Errorf("unsupported stronghold piece %q", piece.ID)
	}
	return result, e.err
}

func (e *strongholdPieceEditor) flag(name string) bool { return strongholdFlag(*e.piece, name) }
func (e *strongholdPieceEditor) number(name string) int32 {
	value, _ := e.piece.Data.Number(name)
	return int32(value)
}
func (e *strongholdPieceEditor) entryDoor() string {
	value, _ := e.piece.Data.Text("EntryDoor")
	return value
}

func (e *strongholdPieceEditor) set(state SurfaceBlockState, x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) || !e.world.InWorld(p) || e.err != nil {
		return
	}
	e.fortressPieceEditor.set(state, x, y, z)
	if e.err != nil {
		return
	}
	if state.Name == "minecraft:end_portal" {
		e.err = e.world.SetBlockEntity(p, state.Name, nbt.Object{"id": {Type: nbt.String, Value: "minecraft:end_portal"}})
		if e.err != nil {
			return
		}
	}
	switch state.Name {
	case "minecraft:torch", "minecraft:wall_torch", "minecraft:oak_fence", "minecraft:ladder", "minecraft:iron_bars":
		if marker, ok := e.world.(interface{ MarkPostProcessing(FeaturePosition) error }); ok {
			e.err = marker.MarkPostProcessing(p)
		}
	}
}

func (e *strongholdPieceEditor) skipAir(x, y, z int32) bool {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) {
		return true
	}
	state, err := e.world.BlockAt(p)
	if err != nil {
		e.err = err
		return true
	}
	return surfaceAir(state)
}

func (e *strongholdPieceEditor) fill(minX, minY, minZ, maxX, maxY, maxZ int32, boundary, interior SurfaceBlockState, skipAir bool) {
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if skipAir && e.skipAir(x, y, z) {
					continue
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

func (e *strongholdPieceEditor) randomBox(minX, minY, minZ, maxX, maxY, maxZ int32, skipAir bool, random StructurePieceRandom) {
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if skipAir && e.skipAir(x, y, z) {
					continue
				}
				state := structureState("cave_air")
				if x == minX || x == maxX || y == minY || y == maxY || z == minZ || z == maxZ {
					state = structureState("stone_bricks")
					roll := random.NextFloat()
					switch {
					case roll < float32(0.2):
						state = structureState("cracked_stone_bricks")
					case roll < float32(0.5):
						state = structureState("mossy_stone_bricks")
					case roll < float32(0.55):
						state = structureState("infested_stone_bricks")
					}
				}
				e.set(state, x, y, z)
			}
		}
	}
}

func (e *strongholdPieceEditor) maybe(random StructurePieceRandom, chance float32, state SurfaceBlockState, x, y, z int32) {
	if random.NextFloat() < chance {
		e.set(state, x, y, z)
	}
}

func (e *strongholdPieceEditor) randomFill(random StructurePieceRandom, chance float32, minX, minY, minZ, maxX, maxY, maxZ int32, boundary, interior SurfaceBlockState) {
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if random.NextFloat() > chance {
					continue
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

func (e *strongholdPieceEditor) door(kind string, x, y, z int32) {
	if kind == "OPENING" {
		e.fill(x, y, z, x+2, y+2, z, structureState("cave_air"), structureState("cave_air"), false)
		return
	}
	if kind == "GRATES" {
		e.set(structureState("cave_air"), x+1, y, z)
		e.set(structureState("cave_air"), x+1, y+1, z)
		west := stateWithProperty(structureState("iron_bars"), "west", "true")
		east := stateWithProperty(structureState("iron_bars"), "east", "true")
		both := stateWithProperty(west, "east", "true")
		e.set(west, x, y, z)
		e.set(west, x, y+1, z)
		e.set(both, x, y+2, z)
		e.set(both, x+1, y+2, z)
		e.set(both, x+2, y+2, z)
		e.set(east, x+2, y+1, z)
		e.set(east, x+2, y, z)
		return
	}
	for _, at := range [][2]int32{{0, 0}, {0, 1}, {0, 2}, {1, 2}, {2, 2}, {2, 1}, {2, 0}} {
		e.set(structureState("stone_bricks"), x+at[0], y+at[1], z)
	}
	name := "oak_door"
	if kind == "IRON_DOOR" {
		name = "iron_door"
	}
	e.set(structureState(name), x+1, y, z)
	e.set(stateWithProperty(structureState(name), "half", "upper"), x+1, y+1, z)
	if kind == "IRON_DOOR" {
		e.set(stateWithProperty(structureState("stone_button"), "facing", "north"), x+2, y+1, z+1)
		e.set(stateWithProperty(structureState("stone_button"), "facing", "south"), x+2, y+1, z-1)
	}
}

func (e *strongholdPieceEditor) flaggedChest(random StructurePieceRandom, x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if e.flag("Chest") || !structureContains(e.clip, p) || e.err != nil {
		return
	}
	strongholdSetBool(e.piece.Data, "Chest", true)
	e.chest(random, "stronghold_corridor", x, y, z)
}

func (e *strongholdPieceEditor) chest(random StructurePieceRandom, loot string, x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) || e.err != nil {
		return
	}
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
	if e.err == nil {
		e.err = e.world.SetBlockEntity(p, "minecraft:chest", nbt.Object{"id": {Type: nbt.String, Value: "minecraft:chest"}, "LootTable": {Type: nbt.String, Value: "minecraft:chests/" + loot}, "LootTableSeed": {Type: nbt.Long, Value: int64(random.NextLong())}})
	}
}

func (e *strongholdPieceEditor) silverfishSpawner(x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if e.flag("Mob") || !structureContains(e.clip, p) || e.err != nil {
		return
	}
	strongholdSetBool(e.piece.Data, "Mob", true)
	e.err = e.world.SetBlock(p, structureState("spawner"))
	if e.err != nil {
		return
	}
	data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:mob_spawner"}, "SpawnData": {Type: nbt.Compound, Value: nbt.Object{"entity": {Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:silverfish"}}}}}, "SpawnPotentials": {Type: nbt.List, ListType: nbt.End, Value: []nbt.Tag{}}}
	for name, value := range map[string]int16{"Delay": 20, "MinSpawnDelay": 200, "MaxSpawnDelay": 800, "SpawnCount": 4, "MaxNearbyEntities": 6, "RequiredPlayerRange": 16, "SpawnRange": 4} {
		data[name] = nbt.Tag{Type: nbt.Short, Value: value}
	}
	e.err = e.world.SetBlockEntity(p, "minecraft:spawner", data)
}
