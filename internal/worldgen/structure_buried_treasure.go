package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

// GenerateBuriedTreasureStructureStart checks the biome at the original
// ocean-floor chunk center. The unprocessed piece starts at local (9,90,9);
// its final depth is selected later from the surface/carved terrain.
func GenerateBuriedTreasureStructureStart(chunkX, chunkZ int32, biomeAt StructureBiomeAtQuart, oceanFloorAt StructureBaseOccupiedHeightAt) (*StructureStart, error) {
	if biomeAt == nil || oceanFloorAt == nil {
		return nil, fmt.Errorf("buried treasure requires base ocean-floor heights and biomes")
	}
	x, z := chunkX*16+8, chunkZ*16+8
	y, err := oceanFloorAt(x, z)
	if err != nil {
		return nil, err
	}
	biome, err := biomeAt(x>>2, y>>2, z>>2)
	if err != nil {
		return nil, err
	}
	allowed, err := StructureBiomeNames("minecraft:buried_treasure")
	if err != nil {
		return nil, err
	}
	for _, name := range allowed {
		if biome == name {
			return &StructureStart{Structure: "minecraft:buried_treasure", Chunk: StructureChunkPosition{chunkX, chunkZ}, TerrainAdjustment: "none", Pieces: []StructurePiece{{ID: "minecraft:btp", Box: StructureBoundingBox{chunkX*16 + 9, 90, chunkZ*16 + 9, chunkX*16 + 9, 90, chunkZ*16 + 9}}}}, nil
		}
	}
	return nil, nil
}

// Vanilla deliberately recognizes only default/source water and lava here.
// Flowing liquid and waterlogged blocks are preserved as solid fill material.
func buriedTreasureLiquid(state SurfaceBlockState) bool {
	if state.Name != "minecraft:water" && state.Name != "minecraft:lava" {
		return false
	}
	return state.Properties["level"] == "" || state.Properties["level"] == "0"
}

func buriedTreasureBase(state SurfaceBlockState) bool {
	switch state.Name {
	case "minecraft:sandstone", "minecraft:stone", "minecraft:andesite", "minecraft:granite", "minecraft:diorite":
		return true
	}
	return false
}

// PostProcessBuriedTreasureBlocks ports the original downward search, six
// neighbor repairs, chest reorientation and loot RNG. Neighbor fill writes
// are intentionally unclipped; only the chest itself uses the chunk clip.
func PostProcessBuriedTreasureBlocks(world StructureBlockEntityWorld, piece *StructurePiece, random StructurePieceRandom, chunkX, chunkZ int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || piece.ID != "minecraft:btp" || random == nil {
		return result, fmt.Errorf("buried treasure requires a btp piece, world and decoration random")
	}
	x, z := piece.Box.MinX, piece.Box.MinZ
	var y int32
	if heights, ok := world.(interface {
		HeightmapAt(string, int32, int32) (int32, error)
	}); ok {
		var err error
		y, err = heights.HeightmapAt("OCEAN_FLOOR_WG", x, z)
		if err != nil {
			return result, err
		}
	} else {
		for y = 319; y >= -64; y-- {
			state, err := world.BlockAt(FeaturePosition{x, y, z})
			if err != nil {
				return result, err
			}
			solid, err := oceanFloorSolid(state)
			if err != nil {
				return result, err
			}
			if solid {
				y++
				break
			}
		}
	}
	minY := int32(-64)
	if bounds, ok := world.(interface{ MinBuildHeight() int32 }); ok {
		minY = bounds.MinBuildHeight()
	}
	for ; y > minY; y-- {
		position := FeaturePosition{x, y, z}
		state, err := world.BlockAt(position)
		if err != nil {
			return result, err
		}
		base, err := world.BlockAt(treeOffset(position, 0, -1, 0))
		if err != nil {
			return result, err
		}
		if !buriedTreasureBase(base) {
			continue
		}
		fill := state
		if surfaceAir(state) || buriedTreasureLiquid(state) {
			fill = structureState("sand")
		}
		for _, offset := range [][3]int32{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}} {
			target := treeOffset(position, offset[0], offset[1], offset[2])
			neighbor, err := world.BlockAt(target)
			if err != nil {
				return result, err
			}
			if !surfaceAir(neighbor) && !buriedTreasureLiquid(neighbor) {
				continue
			}
			below, err := world.BlockAt(treeOffset(target, 0, -1, 0))
			if err != nil {
				return result, err
			}
			replacement := fill
			if (surfaceAir(below) || buriedTreasureLiquid(below)) && offset[1] != 1 {
				replacement = base
			}
			if err := world.SetBlock(target, replacement); err != nil {
				return result, err
			}
		}
		piece.Box = StructureBoundingBox{x, y, z, x, y, z}
		if x>>4 != chunkX || z>>4 != chunkZ || !world.InWorld(position) {
			return result, nil
		}
		old, err := world.BlockAt(position)
		if err != nil || old.Name == "minecraft:chest" {
			return result, err
		}
		editor := &featureEditor{world: world}
		chest := monsterRoomChestState(editor, position)
		if editor.err != nil {
			return result, editor.err
		}
		if err := world.SetBlock(position, chest); err != nil {
			return result, err
		}
		return result, world.SetBlockEntity(position, "minecraft:chest", nbt.Object{
			"id":            {Type: nbt.String, Value: "minecraft:chest"},
			"LootTable":     {Type: nbt.String, Value: "minecraft:chests/buried_treasure"},
			"LootTableSeed": {Type: nbt.Long, Value: int64(random.NextLong())},
		})
	}
	return result, nil
}
