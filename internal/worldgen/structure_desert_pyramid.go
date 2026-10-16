package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"sort"
)

type StructureBlockEntityWorld interface {
	FeatureBlockWorld
	SetBlockEntity(FeaturePosition, string, nbt.Object) error
}

// StructurePieceRandom keeps original caller-owned random streams. Vanilla
// uses the decoration WorldgenRandom for the piece and the world's shared
// random source for the archaeology room's collapsed stairs and sand floor.
type StructurePieceRandom interface {
	NextIntBound(int32) int32
	NextLong() uint64
	NextBoolean() bool
	NextFloat() float32
}

func structureState(name string) SurfaceBlockState {
	return SurfaceBlockState{Name: "minecraft:" + name}
}

func scatteredMinimumHeight(world FeatureBlockWorld, piece *StructurePiece, offset int32) error {
	if height, found := piece.Data["HPos"].Number(); found && height >= 0 {
		return nil
	}
	lowest := int32(320)
	box := piece.Box
	for z := box.MinZ; z <= box.MaxZ; z++ {
		for x := box.MinX; x <= box.MaxX; x++ {
			height, err := structureNoLeavesHeight(world, x, z)
			if err != nil {
				return err
			}
			lowest = min(lowest, height)
		}
	}
	piece.Box.MaxY += lowest - piece.Box.MinY + offset
	piece.Box.MinY = lowest + offset
	piece.Data["HPos"] = nbt.Tag{Type: nbt.Int, Value: lowest}
	return nil
}

func createPyramidChest(e *scatteredPieceEditor, writer StructureBlockEntityWorld, random StructurePieceRandom, x, y, z int32) bool {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) || e.err != nil {
		return false
	}
	old, err := e.world.BlockAt(p)
	if err != nil {
		e.err = err
		return false
	}
	if old.Name == "minecraft:chest" {
		return false
	}
	editor := &featureEditor{world: e.world}
	state := monsterRoomChestState(editor, p)
	if editor.err != nil {
		e.err = editor.err
		return false
	}
	e.err = e.world.SetBlock(p, state)
	if e.err == nil {
		e.err = writer.SetBlockEntity(p, "minecraft:chest", nbt.Object{
			"id":            {Type: nbt.String, Value: "minecraft:chest"},
			"LootTable":     {Type: nbt.String, Value: "minecraft:chests/desert_pyramid"},
			"LootTableSeed": {Type: nbt.Long, Value: int64(random.NextLong())},
		})
	}
	return e.err == nil
}

// PostProcessDesertPyramidBlocks ports the original geometry, buried base,
// pressure-plate/TNT trap, four loot chests and 1.21.1 archaeology room plus
// DesertPyramidStructure.afterPlace. Both random streams are explicit, so a
// caller can preserve vanilla's sequential world-random consumption.
func PostProcessDesertPyramidBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random, worldRandom StructurePieceRandom,
	worldSeed int64, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if piece == nil || piece.ID != "minecraft:tedp" || random == nil || worldRandom == nil {
		return result, fmt.Errorf("desert pyramid requires a tedp piece and both original random streams")
	}
	e, err := newScatteredPieceEditor(world, piece, chunkX, chunkZ)
	if err != nil {
		return result, err
	}
	// The argument draw happens even when saved HPos already exists.
	if err := scatteredMinimumHeight(world, piece, -random.NextIntBound(3)); err != nil {
		return result, err
	}
	placeDesertPyramidGeometry(e)
	for _, direction := range []struct {
		index, dx, dz int32
	}{{2, 0, -2}, {3, 2, 0}, {0, 0, 2}, {1, -2, 0}} {
		flag := fmt.Sprintf("hasPlacedChest%d", direction.index)
		placed, _ := piece.Data[flag].Number()
		if placed == 0 && createPyramidChest(e, world, random, 10+direction.dx, -11, 10+direction.dz) {
			piece.Data[flag] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
		}
	}
	candidates, guaranteed := placePyramidArchaeologyRoom(e, worldRandom, worldSeed)
	if e.err != nil {
		return result, e.err
	}
	if err := afterPlacePyramidArchaeology(e, world, worldSeed, candidates, guaranteed); err != nil {
		return result, err
	}
	return result, e.err
}

func placePyramidArchaeologyRoom(e *scatteredPieceEditor, worldRandom StructurePieceRandom, worldSeed int64) ([]FeaturePosition, FeaturePosition) {
	stairs := stateWithProperty(structureState("sandstone_stairs"), "facing", "west")
	e.set(stairs, 13, -1, 17)
	e.set(stairs, 14, -2, 17)
	e.set(stairs, 15, -3, 17)
	sand, sandstone := structureState("sand"), structureState("sandstone")
	collapsed := worldRandom.NextBoolean()
	for x := int32(12); x <= 16; x++ {
		e.set(sand, x, 0, 17)
	}
	e.set(sand, 14, -1, 17)
	left, right := sandstone, sand
	if collapsed {
		left, right = sand, sandstone
	}
	e.set(left, 15, -1, 17)
	e.set(right, 16, -1, 17)
	e.set(sand, 15, -2, 17)
	e.set(sandstone, 16, -2, 17)
	e.set(sand, 16, -3, 17)
	cut, chiseled := structureState("cut_sandstone"), structureState("chiseled_sandstone")
	for _, layer := range []struct {
		y     int32
		state SurfaceBlockState
	}{{-3, cut}, {-2, chiseled}, {-1, cut}} {
		e.fill(13, layer.y, 10, 13, layer.y, 15, layer.state, layer.state, true)
		e.fill(19, layer.y, 10, 19, layer.y, 15, layer.state, layer.state, true)
		e.fill(13, layer.y, 10, 19, layer.y, 11, layer.state, layer.state, true)
		e.fill(13, layer.y, 16, 19, layer.y, 16, layer.state, layer.state, true)
	}
	var candidates []FeaturePosition
	for y := int32(-3); y <= -1; y++ {
		for x := int32(14); x <= 18; x++ {
			for z := int32(11); z <= 15; z++ {
				candidates = append(candidates, e.piece.BlockPosition(x, y, z))
			}
		}
	}
	for x := int32(14); x <= 18; x++ {
		for z := int32(11); z <= 15; z++ {
			state := sand
			if worldRandom.NextFloat() < 0.33 {
				state = sandstone
			}
			e.set(state, x, 0, z)
		}
	}
	positional := NewLegacyRandom(worldSeed).ForkPositional()
	choice := positional.AtPosition(e.piece.BlockPosition(14, 0, 11))
	x, z := 14+choice.NextIntBound(5), 11+choice.NextIntBound(5)
	guaranteed := e.piece.BlockPosition(x, 0, z)
	orange := structureState("orange_terracotta")
	e.set(structureState("blue_terracotta"), 16, -4, 13)
	for _, p := range [][2]int32{{17, 12}, {17, 14}, {15, 12}, {15, 14}, {18, 13}, {14, 13}, {16, 15}, {16, 11}} {
		e.set(orange, p[0], -4, p[1])
	}
	for _, door := range []struct {
		x, z, outX, outZ int32
		lower, upper     int32
	}{{19, 13, 20, 13, -3, -2}, {13, 13, 12, 13, -3, -2}, {16, 16, 0, 0, 0, 0}, {16, 10, 16, 9, -3, -2}} {
		e.set(orange, door.x, -4, door.z)
		candidates = append(candidates, e.piece.BlockPosition(door.x, -3, door.z), e.piece.BlockPosition(door.x, -2, door.z))
		if door.outX != 0 {
			e.set(cut, door.outX, door.lower, door.outZ)
			e.set(chiseled, door.outX, door.upper, door.outZ)
		}
	}
	return candidates, guaranteed
}

// AtPosition exposes the BlockPos overload of the original positional source.
func (p LegacyPositionalRandom) AtPosition(position FeaturePosition) *LegacyRandom {
	return p.At(position.X, position.Y, position.Z)
}

func afterPlacePyramidArchaeology(e *scatteredPieceEditor, world StructureBlockEntityWorld, worldSeed int64, candidates []FeaturePosition, guaranteed FeaturePosition) error {
	placeSuspicious := func(p FeaturePosition) error {
		if !structureContains(e.clip, p) || !world.InWorld(p) {
			return nil
		}
		if err := world.SetBlock(p, structureState("suspicious_sand")); err != nil {
			return err
		}
		return world.SetBlockEntity(p, "minecraft:suspicious_sand", nbt.Object{
			"id":            {Type: nbt.String, Value: "minecraft:brushable_block"},
			"LootTable":     {Type: nbt.String, Value: "minecraft:archaeology/desert_pyramid"},
			"LootTableSeed": {Type: nbt.Long, Value: featurePackedPosition(p)},
		})
	}
	if err := placeSuspicious(guaranteed); err != nil {
		return err
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Y != b.Y {
			return a.Y < b.Y
		}
		if a.Z != b.Z {
			return a.Z < b.Z
		}
		return a.X < b.X
	})
	box := e.piece.Box
	center := FeaturePosition{box.MinX + (box.MaxX-box.MinX+1)/2, box.MinY + (box.MaxY-box.MinY+1)/2, box.MinZ + (box.MaxZ-box.MinZ+1)/2}
	random := NewLegacyRandom(worldSeed).ForkPositional().AtPosition(center)
	for i := len(candidates); i > 1; i-- {
		other := int(random.NextIntBound(int32(i)))
		candidates[i-1], candidates[other] = candidates[other], candidates[i-1]
	}
	count := min(int32(len(candidates)), 5+random.NextIntBound(3))
	for _, p := range candidates {
		if count > 0 {
			count--
			if err := placeSuspicious(p); err != nil {
				return err
			}
		} else if structureContains(e.clip, p) && world.InWorld(p) {
			if err := world.SetBlock(p, structureState("sand")); err != nil {
				return err
			}
		}
	}
	return nil
}
