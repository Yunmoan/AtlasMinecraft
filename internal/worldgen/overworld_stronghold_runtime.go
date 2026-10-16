package worldgen

import "fmt"

func (g *OverworldGenerator) strongholdPlacementAt(cx, cz int32) (bool, error) {
	c := g.structures
	p := c.placement.Sets["minecraft:strongholds"].Placement
	// The first raw ring radius is >=2.75*distance. Chunk rounding and the
	// original 112-block relocation move it by <11 chunks. For distance>=16,
	// the radius-2*distance disk therefore contains no possible ring position.
	// Normal spawn chunks need not populate all 128 distant rings at startup.
	distance := int64(p.Distance)
	if distance >= 16 && int64(cx)*int64(cx)+int64(cz)*int64(cz) < 4*distance*distance {
		return false, nil
	}
	c.ringsOnce.Do(func() { c.ringsErr = c.placement.PopulateOverworldStrongholdRings() })
	if c.ringsErr != nil {
		return false, c.ringsErr
	}
	return c.placement.IsStructureChunk("minecraft:strongholds", cx, cz)
}

// Stronghold pieces read only their clip and one-block chest neighbors. A
// clipped cache prevents a portal-room visit from generating the entire large
// stronghold. Each clip uses its own original decoration stream.
func (g *OverworldGenerator) strongholdStructureBlocksFor(start StructureStart, cx, cz int32) (*structureBlockResult, error) {
	c := g.structures
	key := dimensionStructureBlockKey{start.Structure, start.Chunk.X, start.Chunk.Z, cx, cz}
	c.mu.Lock()
	c.clock++
	if entry := c.strongholdBlocks[key]; entry != nil {
		entry.used = c.clock
		c.mu.Unlock()
		<-entry.ready
		return entry.result, entry.err
	}
	entry := &structureBlockEntry{ready: make(chan struct{}), used: c.clock}
	c.strongholdBlocks[key] = entry
	c.mu.Unlock()
	result, err := g.generateStrongholdStructureBlocks(start, cx, cz)
	c.mu.Lock()
	entry.result, entry.err, entry.complete = result, err, true
	close(entry.ready)
	if err != nil {
		delete(c.strongholdBlocks, key)
	}
	for len(c.strongholdBlocks) > 128 {
		var oldest dimensionStructureBlockKey
		used, found := ^uint64(0), false
		for position, candidate := range c.strongholdBlocks {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(c.strongholdBlocks, oldest)
	}
	c.mu.Unlock()
	return result, err
}

// Neighbor decoration computes clips that are never returned or saved. Only
// the successfully generated target commits its placement flags; otherwise a
// saved owner could claim a spawner exists in a still-uncreated neighbor.
func (g *OverworldGenerator) commitStrongholdChunk(starts []StructureStart, cx, cz int32) error {
	for _, start := range starts {
		if start.Structure != "minecraft:stronghold" {
			continue
		}
		box, ok := start.BoundingBox()
		if !ok || box.MaxX < cx*16 || box.MinX > cx*16+15 || box.MaxZ < cz*16 || box.MinZ > cz*16+15 {
			continue
		}
		result, err := g.strongholdStructureBlocksFor(start, cx, cz)
		if err != nil {
			return err
		}
		c := g.structures
		owner := featureChunkPosition{start.Chunk.X, start.Chunk.Z}
		c.mu.Lock()
		saved, exists := c.strongholdSaved[owner]
		if !exists {
			saved = start.Clone()
		}
		for index, piece := range result.start.Pieces {
			for _, flag := range []string{"Chest", "Mob"} {
				if strongholdFlag(piece, flag) {
					strongholdSetBool(saved.Pieces[index].Data, flag, true)
				}
			}
		}
		c.strongholdSaved[owner] = saved
		c.mu.Unlock()
	}
	return nil
}

func (g *OverworldGenerator) generateStrongholdStructureBlocks(start StructureStart, cx, cz int32) (*structureBlockResult, error) {
	order, err := OriginalStructureDecorationOrder(start.Structure)
	if err != nil {
		return nil, err
	}
	if len(start.Pieces) == 0 {
		return nil, fmt.Errorf("stronghold start has no pieces")
	}
	region, err := g.structureRegionBefore(cx, cz, order.Step)
	if err != nil {
		return nil, err
	}
	result := &structureBlockResult{start: start.Clone(), blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
	world := &structureRecordingWorld{region: region, result: result}
	seed := OverworldDecorationSeed(g.seed, cx, cz)
	random := NewWorldgenRandom(seed + int64(order.Index) + 10000*int64(order.Step))
	for index := range result.start.Pieces {
		piece := &result.start.Pieces[index]
		box := piece.Box
		if box.MaxX < cx*16 || box.MinX > cx*16+15 || box.MaxZ < cz*16 || box.MinZ > cz*16+15 {
			continue
		}
		if _, err := PostProcessStrongholdBlocks(world, piece, random, cx, cz); err != nil {
			return nil, fmt.Errorf("postprocess stronghold in %d,%d: %w", cx, cz, err)
		}
	}
	return result, nil
}

func (g *OverworldGenerator) strongholdSavedStart(start StructureStart) StructureStart {
	if g.structures == nil {
		return start.Clone()
	}
	c := g.structures
	c.mu.Lock()
	defer c.mu.Unlock()
	if saved, exists := c.strongholdSaved[featureChunkPosition{start.Chunk.X, start.Chunk.Z}]; exists {
		return saved.Clone()
	}
	return start.Clone()
}

// StrongholdStartSnapshots returns detached metadata after every generated
// clip's Chest/Mob flags have been merged. World saving can refresh an already
// generated owner chunk even when its portal clip was generated later.
// Geometry replay keeps the immutable original starts and is order-independent.
func (g *OverworldGenerator) StrongholdStartSnapshots() []StructureStart {
	if g.structures == nil {
		return nil
	}
	c := g.structures
	c.mu.Lock()
	defer c.mu.Unlock()
	starts := make([]StructureStart, 0, len(c.strongholdSaved))
	for _, start := range c.strongholdSaved {
		starts = append(starts, start.Clone())
	}
	return starts
}

// StrongholdStartSnapshot avoids cloning unrelated starts when persisting one
// existing owner chunk. It never generates terrain or acquires a world lock.
func (g *OverworldGenerator) StrongholdStartSnapshot(cx, cz int32) (StructureStart, bool) {
	if g.structures == nil {
		return StructureStart{}, false
	}
	c := g.structures
	c.mu.Lock()
	defer c.mu.Unlock()
	start, exists := c.strongholdSaved[featureChunkPosition{cx, cz}]
	if !exists {
		return StructureStart{}, false
	}
	return start.Clone(), true
}
