package worldgen

import "fmt"

func isOverworldJigsawStructure(name string) bool {
	switch name {
	case "minecraft:ancient_city", "minecraft:trail_ruins", "minecraft:trial_chambers", "minecraft:pillager_outpost":
		return true
	default:
		return isOverworldVillageStructure(name)
	}
}

// Jigsaw pieces keep their authored origins and do not need scattered pieces'
// canonical ground-alignment pass. Cache each clipped call independently so
// entering one corner of a city does not generate every chunk in the city.
func (g *OverworldGenerator) jigsawStructureBlocksFor(start StructureStart, cx, cz int32) (*structureBlockResult, error) {
	c := g.structures
	key := dimensionStructureBlockKey{start.Structure, start.Chunk.X, start.Chunk.Z, cx, cz}
	c.mu.Lock()
	c.clock++
	if entry := c.jigsawBlocks[key]; entry != nil {
		entry.used = c.clock
		c.mu.Unlock()
		<-entry.ready
		return entry.result, entry.err
	}
	entry := &structureBlockEntry{ready: make(chan struct{}), used: c.clock}
	c.jigsawBlocks[key] = entry
	c.mu.Unlock()
	result, err := g.generateJigsawStructureBlocks(start, cx, cz)
	c.mu.Lock()
	entry.result, entry.err, entry.complete = result, err, true
	close(entry.ready)
	if err != nil {
		delete(c.jigsawBlocks, key)
	}
	for len(c.jigsawBlocks) > 128 {
		var oldest dimensionStructureBlockKey
		used, found := ^uint64(0), false
		for position, candidate := range c.jigsawBlocks {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(c.jigsawBlocks, oldest)
	}
	c.mu.Unlock()
	return result, err
}

func (g *OverworldGenerator) structureRegionBefore(cx, cz, step int32) (*OverworldFeatureRegion, error) {
	bases, chunks, err := g.decorationBases(cx, cz)
	if err != nil {
		return nil, err
	}
	var preceding []StructureStart
	for _, neighbor := range g.decorationStructures(bases, cx, cz) {
		order, err := OriginalStructureDecorationOrder(neighbor.Structure)
		if err != nil {
			return nil, err
		}
		if order.Step < step {
			preceding = append(preceding, neighbor)
		}
	}
	region, _, err := g.decorateRegion(bases, chunks, cx, cz, int(step), preceding)
	return region, err
}

func (g *OverworldGenerator) generateJigsawStructureBlocks(start StructureStart, cx, cz int32) (*structureBlockResult, error) {
	order, err := OriginalStructureDecorationOrder(start.Structure)
	if err != nil {
		return nil, err
	}
	region, err := g.structureRegionBefore(cx, cz, order.Step)
	if err != nil {
		return nil, err
	}
	result := &structureBlockResult{start: start.Clone(), blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
	world := &structureRecordingWorld{region: region, result: result}
	if len(start.Pieces) == 0 {
		return nil, fmt.Errorf("Jigsaw start has no pieces")
	}
	box := start.Pieces[0].Box
	pivot := FeaturePosition{box.MinX + (box.MaxX-box.MinX+1)/2, box.MinY, box.MinZ + (box.MaxZ-box.MinZ+1)/2}
	seed := OverworldDecorationSeed(g.seed, cx, cz)
	random := NewWorldgenRandom(seed + int64(order.Index) + 10000*int64(order.Step))
	// Template mob finalization uses WorldGenRegion's independent stream,
	// never the structure's block/loot placement stream.
	worldRandom := NewStructureRegionRandom(g.seed, cx, cz)
	for index := range result.start.Pieces {
		piece := &result.start.Pieces[index]
		box := piece.Box
		if box.MaxX < cx*16 || box.MinX > cx*16+15 || box.MaxZ < cz*16 || box.MinZ > cz*16+15 {
			continue
		}
		// Processors inspect the complete template before clipping. Each remote
		// read chunk retains the same preceding-stage replay as a local chunk;
		// injecting raw noise here would lose caves, lakes and earlier structures.
		if err := g.extendJigsawReadRegion(region, box, order.Step); err != nil {
			return nil, err
		}
		placed, err := PostProcessJigsawBlocks(world, piece, random, cx, cz, pivot)
		if err != nil {
			return nil, fmt.Errorf("postprocess%s in%d,%d: %w", start.Structure, cx, cz, err)
		}
		for actorIndex := range placed.PendingEntities {
			request := &placed.PendingEntities[actorIndex]
			// Indices in a clipped request list are local to that clip. Include
			// it in the identity so separate cages crossing chunk boundaries
			// cannot alias, even when actors share a block position.
			request.SavedFlag = fmt.Sprintf("Jigsaw:%s:%d,%d:%d:%d,%d:%d", start.Structure, start.Chunk.X, start.Chunk.Z, index, cx, cz, actorIndex)
			switch start.Structure {
			case "minecraft:pillager_outpost":
				request.Data, err = FinalizeOutpostTemplateEntity(*request, worldRandom)
			default:
				if isOverworldVillageStructure(start.Structure) {
					err = g.finalizeVillageRequest(region, request, worldRandom)
				} else {
					err = fmt.Errorf("unsupported actors in %s", start.Structure)
				}
			}
			if err != nil {
				return nil, err
			}
		}
		result.pending = append(result.pending, placed.PendingEntities...)
	}
	return result, nil
}

func (g *OverworldGenerator) extendJigsawReadRegion(region *OverworldFeatureRegion, box StructureBoundingBox, step int32) error {
	for x := (box.MinX - 1) >> 4; x <= (box.MaxX+1)>>4; x++ {
		for z := (box.MinZ - 1) >> 4; z <= (box.MaxZ+1)>>4; z++ {
			key := featureChunkPosition{x, z}
			if region.chunks[key] != nil {
				continue
			}
			remote, err := g.structureChunkBefore(x, z, step)
			if err != nil {
				return err
			}
			region.installStructureChunk(key, remote)
		}
	}
	return nil
}
