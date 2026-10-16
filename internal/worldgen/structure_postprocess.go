package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"strings"
)

// PendingStructureEntity is an explicit request for a structure's runtime
// entity stage. Block generation does not imply this request was fulfilled.
type PendingStructureEntity struct {
	ID         string
	Position   FeaturePosition
	Persistent bool
	SavedFlag  string
	Data       nbt.Object
}

type StructurePiecePostprocessResult struct {
	PendingEntities []PendingStructureEntity
}

type scatteredPieceEditor struct {
	world FeatureBlockWorld
	piece *StructurePiece
	clip  StructureBoundingBox
	err   error
}

func (piece StructurePiece) BlockPosition(x, y, z int32) FeaturePosition {
	box := piece.Box
	switch piece.Orientation {
	case "north":
		return FeaturePosition{box.MinX + x, box.MinY + y, box.MaxZ - z}
	case "south":
		return FeaturePosition{box.MinX + x, box.MinY + y, box.MinZ + z}
	case "west":
		return FeaturePosition{box.MaxX - z, box.MinY + y, box.MinZ + x}
	case "east":
		return FeaturePosition{box.MinX + z, box.MinY + y, box.MinZ + x}
	default:
		return FeaturePosition{x, y, z}
	}
}

func structureContains(box StructureBoundingBox, p FeaturePosition) bool {
	return p.X >= box.MinX && p.X <= box.MaxX && p.Y >= box.MinY && p.Y <= box.MaxY && p.Z >= box.MinZ && p.Z <= box.MaxZ
}

// Mirror precedes rotation. Cardinal connections use the same transformation
// as facing; stairs swap left/right shapes under LEFT_RIGHT
// only when their initial facing axis is Z, exactly as StairBlock.mirror.
func scatteredBlockState(state SurfaceBlockState, orientation string) SurfaceBlockState {
	if state.Name == "minecraft:redstone_wire" || state.Name == "minecraft:tripwire" || state.Name == "minecraft:vine" {
		original := state
		for _, direction := range []string{"north", "east", "south", "west"} {
			value := original.Properties[direction]
			if value == "" {
				value = "false"
				if state.Name == "minecraft:redstone_wire" {
					value = "none"
				}
			}
			target := direction
			if orientation == "south" || orientation == "west" {
				if target == "north" {
					target = "south"
				} else if target == "south" {
					target = "north"
				}
			}
			if orientation == "east" || orientation == "west" {
				target = map[string]string{"north": "east", "east": "south", "south": "west", "west": "north"}[target]
			}
			state = stateWithProperty(state, target, value)
		}
	}
	facing := state.Properties["facing"]
	if facing == "" {
		return state
	}
	if orientation == "south" || orientation == "west" {
		if facing == "north" || facing == "south" {
			facing = map[string]string{"north": "south", "south": "north"}[facing]
			if strings.HasSuffix(state.Name, "_stairs") {
				shape := state.Properties["shape"]
				if swapped, found := map[string]string{"inner_left": "inner_right", "inner_right": "inner_left", "outer_left": "outer_right", "outer_right": "outer_left"}[shape]; found {
					state = stateWithProperty(state, "shape", swapped)
				}
			}
		}
	}
	if (orientation == "east" || orientation == "west") && facing != "up" && facing != "down" {
		facing = map[string]string{"north": "east", "east": "south", "south": "west", "west": "north"}[facing]
	}
	return stateWithProperty(state, "facing", facing)
}

func (e *scatteredPieceEditor) set(state SurfaceBlockState, x, y, z int32) {
	if e.err != nil {
		return
	}
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) || !e.world.InWorld(p) {
		return
	}
	state = scatteredBlockState(state, e.piece.Orientation)
	e.err = e.world.SetBlock(p, state)
	if e.err != nil {
		return
	}
	if featureFluid(state) {
		if scheduler, ok := e.world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
			e.err = scheduler.ScheduleFluidUpdate(p)
		}
	}
	if state.Name == "minecraft:oak_fence" {
		if marker, ok := e.world.(interface{ MarkPostProcessing(FeaturePosition) error }); ok {
			e.err = marker.MarkPostProcessing(p)
		}
	}
}

func (e *scatteredPieceEditor) fill(minX, minY, minZ, maxX, maxY, maxZ int32, boundary, interior SurfaceBlockState, skipAir bool) {
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if skipAir {
					p := e.piece.BlockPosition(x, y, z)
					if !structureContains(e.clip, p) {
						continue
					}
					state, err := e.world.BlockAt(p)
					if err != nil {
						e.err = err
						return
					}
					if surfaceAir(state) {
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

func (e *scatteredPieceEditor) fillDown(state SurfaceBlockState, x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) {
		return
	}
	for p.Y > -63 && e.world.InWorld(p) && e.err == nil {
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

func structureNoLeavesHeight(world FeatureBlockWorld, x, z int32) (int32, error) {
	for y := int32(319); y >= -64; y-- {
		state, err := world.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return 0, err
		}
		if strings.HasSuffix(state.Name, "_leaves") {
			continue
		}
		motion, err := motionBlocking(state)
		if err != nil {
			return 0, err
		}
		if motion {
			return y + 1, nil
		}
	}
	return -64, nil
}

func scatteredAverageHeight(world FeatureBlockWorld, piece *StructurePiece, clip StructureBoundingBox) (bool, error) {
	if height, found := piece.Data["HPos"].Number(); found && height >= 0 {
		return true, nil
	}
	total, count := int32(0), int32(0)
	box := piece.Box
	for z := box.MinZ; z <= box.MaxZ; z++ {
		for x := box.MinX; x <= box.MaxX; x++ {
			if !structureContains(clip, FeaturePosition{x, 64, z}) {
				continue
			}
			height, err := structureNoLeavesHeight(world, x, z)
			if err != nil {
				return false, err
			}
			total += height
			count++
		}
	}
	if count == 0 {
		return false, nil
	}
	height := total / count
	piece.Box.MaxY += height - piece.Box.MinY
	piece.Box.MinY = height
	piece.Data["HPos"] = nbt.Tag{Type: nbt.Int, Value: height}
	return true, nil
}

func newScatteredPieceEditor(world FeatureBlockWorld, piece *StructurePiece, chunkX, chunkZ int32) (*scatteredPieceEditor, error) {
	if world == nil || piece == nil || piece.Data == nil {
		return nil, fmt.Errorf("scattered piece postprocessing needs a world and saved piece")
	}
	return &scatteredPieceEditor{world: world, piece: piece, clip: StructureBoundingBox{chunkX * 16, -64, chunkZ * 16, chunkX*16 + 15, 319, chunkZ*16 + 15}}, nil
}
