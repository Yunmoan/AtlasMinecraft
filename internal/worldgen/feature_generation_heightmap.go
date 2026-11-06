package worldgen

// BeginFeatureStage preserves the WG heightmaps at the CARVERS boundary.
// Original ProtoChunk.setBlockState subsequently maintains only the four
// non-WG maps. Reading WG heights during decoration must not see trees or
// structure writes added in this stage.
func (region *OverworldFeatureRegion) BeginFeatureStage() {
	if region.worldgenHeights == nil {
		region.worldgenHeights = make(map[featureColumnPosition][2]int32)
	}
}

func (region *OverworldFeatureRegion) frozenWorldgenHeights(x, z int32) ([2]int32, error) {
	key := featureColumnPosition{x, z}
	if heights, found := region.worldgenHeights[key]; found {
		return heights, nil
	}
	chunk, err := region.chunkAt(x, z)
	if err != nil {
		return [2]int32{}, err
	}
	heights := [2]int32{region.minY, region.minY}
	floor, surface := false, false
	for y := region.minY + region.height - 1; y >= region.minY && !floor; y-- {
		// Borrowed chunks remain immutable: overlays contain only FEATURES
		// writes. Mutable regions capture a column before its first such write.
		state := chunk.surfaceState(x&15, y, z&15)
		if !surface && !surfaceAir(state) {
			heights[1], surface = y+1, true
		}
		solid, err := oceanFloorSolid(state)
		if err != nil {
			return [2]int32{}, err
		}
		if solid {
			heights[0], floor = y+1, true
		}
	}
	region.worldgenHeights[key] = heights
	return heights, nil
}
