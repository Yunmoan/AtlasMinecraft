package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

// JungleTemplePiece's MossStoneSelector draws even for clipped-out positions.
func (e *scatteredPieceEditor) jungleFill(random StructurePieceRandom, minX, minY, minZ, maxX, maxY, maxZ int32) {
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				state := structureState("mossy_cobblestone")
				if random.NextFloat() < 0.4 {
					state = structureState("cobblestone")
				}
				e.set(state, x, y, z)
			}
		}
	}
}

func placeJungleLoot(e *scatteredPieceEditor, writer StructureBlockEntityWorld, random StructurePieceRandom, flag string,
	x, y, z int32, loot, facing string) {
	placed, _ := e.piece.Data[flag].Number()
	if placed != 0 || e.err != nil {
		return
	}
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) {
		return
	}
	old, err := e.world.BlockAt(p)
	if err != nil {
		e.err = err
		return
	}
	block := "minecraft:chest"
	if facing != "" {
		block = "minecraft:dispenser"
	}
	if old.Name == block {
		return
	}
	state := SurfaceBlockState{Name: block}
	if facing == "" {
		editor := &featureEditor{world: e.world}
		state = monsterRoomChestState(editor, p)
		if editor.err != nil {
			e.err = editor.err
			return
		}
		e.err = e.world.SetBlock(p, state)
	} else {
		state = stateWithProperty(state, "facing", facing)
		e.set(state, x, y, z)
	}
	if e.err != nil {
		return
	}
	e.err = writer.SetBlockEntity(p, block, nbt.Object{
		"id":            {Type: nbt.String, Value: block},
		"LootTable":     {Type: nbt.String, Value: loot},
		"LootTableSeed": {Type: nbt.Long, Value: int64(random.NextLong())},
	})
	if e.err == nil {
		e.piece.Data[flag] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	}
}

// PostProcessJungleTempleBlocks places the original temple geometry, two
// tripwire/dispenser traps, redstone/lever/piston puzzle and both loot chests.
// Loot table seeds remain persisted for the normal container loot stage.
func PostProcessJungleTempleBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom,
	chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if piece == nil || piece.ID != "minecraft:tejp" || random == nil {
		return result, fmt.Errorf("jungle temple requires a tejp piece and original decoration random")
	}
	e, err := newScatteredPieceEditor(world, piece, chunkX, chunkZ)
	if err != nil {
		return result, err
	}
	ready, err := scatteredAverageHeight(world, piece, e.clip)
	if err != nil || !ready {
		return result, err
	}
	placeJungleTempleGeometry(e, world, random)
	return result, e.err
}
