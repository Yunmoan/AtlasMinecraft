package worldgen

import (
	"fmt"
)

var (
	monumentPrismarine = structureState("prismarine")
	monumentBricks     = structureState("prismarine_bricks")
	monumentDark       = structureState("dark_prismarine")
	monumentLantern    = structureState("sea_lantern")
	monumentWater      = structureState("water")
)

type monumentPieceEditor struct {
	fortressPieceEditor
	child    *monumentChild
	seaLevel int32
	result   *StructurePiecePostprocessResult
}

// PostProcessOceanMonumentBlocks reconstructs the original hidden room graph
// from the world seed and saved outer piece, then clips placement to one chunk.
// Entity requests describe the three elder guardians; actor finalization is a
// separate stage, as for the other structure block placers.
func PostProcessOceanMonumentBlocks(world StructureBlockEntityWorld, start *StructureStart, seed int64, random StructurePieceRandom, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	result := StructurePiecePostprocessResult{}
	if world == nil || start == nil || random == nil || start.Structure != "minecraft:monument" || len(start.Pieces) != 1 || start.Pieces[0].ID != "minecraft:omb" {
		return result, fmt.Errorf("monument postprocess requires a world, original saved building and random source")
	}
	rng := NewLegacyRandom(seed)
	xm, zm := rng.NextLong(), rng.NextLong()
	rng = NewLegacyRandom(int64(uint64(int64(start.Chunk.X))*xm ^ uint64(int64(start.Chunk.Z))*zm ^ uint64(seed)))
	fallback := [...]string{"north", "east", "south", "west"}[rng.NextIntBound(4)]
	piece := start.Pieces[0]
	direction := piece.Orientation
	if direction == "" {
		direction = fallback
	}
	layout, err := assembleMonument(rng, piece.Box.MinX, piece.Box.MinZ, direction)
	if err != nil {
		return result, err
	}
	minY, maxY, seaLevel := int32(-64), int32(320), int32(63)
	if bounds, ok := world.(interface {
		MinBuildHeight() int32
		MaxBuildHeight() int32
	}); ok {
		minY, maxY = bounds.MinBuildHeight(), bounds.MaxBuildHeight()
	}
	if sea, ok := world.(interface{ GenerationSeaLevel() int32 }); ok {
		seaLevel = sea.GenerationSeaLevel()
	}
	clip := StructureBoundingBox{chunkX * 16, minY, chunkZ * 16, chunkX*16 + 15, maxY - 1, chunkZ*16 + 15}
	return placeMonumentLayout(world, layout, random, clip, seaLevel)
}

func placeMonumentLayout(world StructureBlockEntityWorld, layout *monumentLayout, random StructurePieceRandom, clip StructureBoundingBox, seaLevel int32) (StructurePiecePostprocessResult, error) {
	result := StructurePiecePostprocessResult{}
	e := newMonumentEditor(world, &layout.building, clip, seaLevel, &result)
	e.waterBox(0, 0, 0, 58, max(seaLevel, 64)-layout.building.Box.MinY, 58)
	e.placeWingShell(false, 0)
	e.placeWingShell(true, 33)
	e.placeEntranceArches()
	e.placeEntranceWall()
	e.placeRoof()
	e.placeLowerWall()
	e.placeMiddleWall()
	e.placeUpperWall()
	for x := int32(0); x < 7; x++ {
		for z := int32(0); z < 7; {
			if z == 0 && x == 3 {
				z = 6
			}
			for dx := int32(0); dx < 4; dx++ {
				for dz := int32(0); dz < 4; dz++ {
					e.set(monumentBricks, x*9+dx, 0, z*9+dz)
					e.fillDown(monumentBricks, x*9+dx, -1, z*9+dz)
				}
			}
			if x == 0 || x == 6 {
				z++
			} else {
				z += 6
			}
		}
	}
	for i := int32(0); i < 5; i++ {
		e.waterBox(-1-i, i*2, -1-i, -1-i, 23, 58+i)
		e.waterBox(58+i, i*2, -1-i, 58+i, 23, 58+i)
		e.waterBox(-i, i*2, -1-i, 57+i, 23, -1-i)
		e.waterBox(-i, i*2, 58+i, 57+i, 23, 58+i)
	}
	if e.err != nil {
		return result, e.err
	}
	for i := range layout.children {
		child := &layout.children[i]
		if !strongholdBoxesIntersect(child.piece.Box, clip) {
			continue
		}
		editor := newMonumentEditor(world, &child.piece, clip, seaLevel, &result)
		editor.child = child
		if err := editor.placeRoom(random); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (room *monumentRoom) openingCount() int {
	count := 0
	for _, open := range room.open {
		if open {
			count++
		}
	}
	return count
}

func newMonumentEditor(world StructureBlockEntityWorld, piece *StructurePiece, clip StructureBoundingBox, seaLevel int32, result *StructurePiecePostprocessResult) *monumentPieceEditor {
	minY := int32(-64)
	if bounds, ok := world.(interface{ MinBuildHeight() int32 }); ok {
		minY = bounds.MinBuildHeight()
	}
	return &monumentPieceEditor{fortressPieceEditor: fortressPieceEditor{world: world, piece: piece, clip: clip, minY: minY, transformed: make(map[uint32]SurfaceBlockState)}, seaLevel: seaLevel, result: result}
}

func (e *monumentPieceEditor) local(x, y, z int32) SurfaceBlockState {
	p := e.piece.BlockPosition(x, y, z)
	if e.err != nil || !structureContains(e.clip, p) || !e.world.InWorld(p) {
		return structureState("air")
	}
	state, err := e.world.BlockAt(p)
	if err != nil {
		e.err = err
		return structureState("air")
	}
	return state
}

// generateWaterBox preserves water (including flowing states) and all three
// ice blocks. Above sea level it carves air; below it fills source water.
func (e *monumentPieceEditor) waterBox(x1, y1, z1, x2, y2, z2 int32) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			for z := z1; z <= z2; z++ {
				state := e.local(x, y, z)
				switch state.Name {
				case "minecraft:water", "minecraft:ice", "minecraft:packed_ice", "minecraft:blue_ice":
					continue
				}
				replacement := monumentWater
				if e.piece.Box.MinY+y >= e.seaLevel {
					replacement = structureState("air")
				}
				e.set(replacement, x, y, z)
			}
		}
	}
}

func (e *monumentPieceEditor) waterOnlyFill(x1, y1, z1, x2, y2, z2 int32, state SurfaceBlockState) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			for z := z1; z <= z2; z++ {
				old := e.local(x, y, z)
				if old.Name == "minecraft:water" && (old.Properties["level"] == "0" || old.Properties["level"] == "") {
					e.set(state, x, y, z)
				}
			}
		}
	}
}

func (e *monumentPieceEditor) floor(x, z int32, opening bool) {
	if !opening {
		e.fill(x, 0, z, x+7, 0, z+7, monumentPrismarine, monumentPrismarine, false)
		return
	}
	for _, box := range [][4]int32{{x, z, x + 2, z + 7}, {x + 5, z, x + 7, z + 7}, {x + 3, z, x + 4, z + 2}, {x + 3, z + 5, x + 4, z + 7}} {
		e.fill(box[0], 0, box[1], box[2], 0, box[3], monumentPrismarine, monumentPrismarine, false)
	}
	for _, box := range [][4]int32{{x + 3, z + 2, x + 4, z + 2}, {x + 3, z + 5, x + 4, z + 5}, {x + 2, z + 3, x + 2, z + 4}, {x + 5, z + 3, x + 5, z + 4}} {
		e.fill(box[0], 0, box[1], box[2], 0, box[3], monumentBricks, monumentBricks, false)
	}
}

func (e *monumentPieceEditor) intersects(x1, z1, x2, z2 int32) bool {
	a, b := e.piece.BlockPosition(x1, 0, z1), e.piece.BlockPosition(x2, 0, z2)
	return max(a.X, b.X) >= e.clip.MinX && min(a.X, b.X) <= e.clip.MaxX && max(a.Z, b.Z) >= e.clip.MinZ && min(a.Z, b.Z) <= e.clip.MaxZ
}

func (e *monumentPieceEditor) spawnGuardian(x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if e.err == nil && structureContains(e.clip, p) {
		e.result.PendingEntities = append(e.result.PendingEntities, PendingStructureEntity{ID: "minecraft:elder_guardian", Position: p})
	}
}

func (e *monumentPieceEditor) placeRoom(random StructurePieceRandom) error {
	switch e.piece.ID {
	case "minecraft:omentry":
		e.placeEntry(random)
	case "minecraft:omcr":
		e.placeCore(random)
	case "minecraft:omdxr":
		e.placeDoubleX(random)
	case "minecraft:omdxyr":
		e.placeDoubleXY(random)
	case "minecraft:omdyr":
		e.placeDoubleY(random)
	case "minecraft:omdyzr":
		e.placeDoubleYZ(random)
	case "minecraft:omdzr":
		e.placeDoubleZ(random)
	case "minecraft:omsimple":
		e.placeSimple(random)
	case "minecraft:omsimplet":
		e.placeSimpleTop(random)
	case "minecraft:omwr":
		e.placeWing(random)
	case "minecraft:ompenthouse":
		e.placePenthouse(random)
	default:
		return fmt.Errorf("unsupported monument child %q", e.piece.ID)
	}
	return e.err
}
