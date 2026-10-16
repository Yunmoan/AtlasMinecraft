package worldgen

import "strconv"

// finalizeGeneratedLeafDistances settles the one-tick leaf updates requested
// by TreeFeature's boundary shape updates. Its per-tree distance search cannot
// see logs or leaves beyond that tree's bounding box, whereas LeavesBlock's
// scheduled update can. This pass changes only distance, leaving decay and
// later gameplay ticks to the runtime.
func finalizeGeneratedLeafDistances(region *OverworldFeatureRegion, cx, cz int32) error {
	leaves := make(map[FeaturePosition]SurfaceBlockState)
	var frontier []FeaturePosition
	collect := func(chunk *NoiseChunkResult, index int, state SurfaceBlockState) {
		p := FeaturePosition{chunk.X*16 + int32(index%16), chunk.MinY + int32(index/256), chunk.Z*16 + int32(index/16%16)}
		if treeTags["logs"][state.Name] {
			frontier = append(frontier, p)
		} else if treeTags["leaves"][state.Name] {
			leaves[p] = state
		}
	}
	// Logs and leaves cannot originate in the noise material array. Inspect
	// sparse surface states and private decoration writes, including neighbors.
	for key, chunk := range region.chunks {
		overlay := region.overlays[key]
		for index, state := range chunk.SurfaceBlocks {
			if overlay != nil {
				if _, replaced := overlay.blocks[index]; replaced {
					continue
				}
			}
			collect(chunk, index, state)
		}
		if overlay != nil {
			for index, state := range overlay.blocks {
				collect(chunk, index, state)
			}
		}
	}
	if len(leaves) == 0 {
		return nil
	}
	distances := make(map[FeaturePosition]int, len(leaves))
	directions := [...]FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}}
	for distance := 1; distance < 7 && len(frontier) > 0; distance++ {
		var next []FeaturePosition
		for _, p := range frontier {
			for _, delta := range directions {
				q := treeOffset(p, delta.X, delta.Y, delta.Z)
				if _, found := leaves[q]; !found {
					continue
				}
				if _, visited := distances[q]; visited {
					continue
				}
				distances[q] = distance
				next = append(next, q)
			}
		}
		frontier = next
	}
	for p, state := range leaves {
		if p.X>>4 != cx || p.Z>>4 != cz {
			continue
		}
		distance, found := distances[p]
		if !found {
			distance = 7
		}
		value := strconv.Itoa(distance)
		if state.Properties["distance"] != value {
			if err := region.SetBlock(p, stateWithProperty(state, "distance", value)); err != nil {
				return err
			}
		}
	}
	return nil
}
