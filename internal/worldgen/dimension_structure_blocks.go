package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

// Fortress pieces have fixed Y and clip-local chest/spawner flags. Cache each
// clipped call rather than constructing every chunk touched by a large fortress
// when a player first enters one end of it. Template pieces that expand their
// clip are recorded once in their original canonical center/owner chunk.
func dimensionStructureClip(start StructureStart, cx, cz int32) (int32, int32) {
	switch start.Structure {
	case "minecraft:ruined_portal_nether":
		box, _ := start.BoundingBox()
		return (box.MinX + (box.MaxX-box.MinX+1)/2) >> 4, (box.MinZ + (box.MaxZ-box.MinZ+1)/2) >> 4
	case "minecraft:nether_fossil":
		return start.Chunk.X, start.Chunk.Z
	default:
		return cx, cz
	}
}

func (g *DimensionGenerator) structureBlocksFor(start StructureStart, cx, cz int32) (*structureBlockResult, error) {
	cx, cz = dimensionStructureClip(start, cx, cz)
	c := g.structures
	key := dimensionStructureBlockKey{start.Structure, start.Chunk.X, start.Chunk.Z, cx, cz}
	c.mu.Lock()
	c.clock++
	if entry := c.blocks[key]; entry != nil {
		entry.used = c.clock
		c.mu.Unlock()
		<-entry.ready
		return entry.result, entry.err
	}
	entry := &structureBlockEntry{ready: make(chan struct{}), used: c.clock}
	c.blocks[key] = entry
	c.mu.Unlock()
	result, err := g.generateStructureBlocks(start, cx, cz)
	c.mu.Lock()
	entry.result, entry.err, entry.complete = result, err, true
	close(entry.ready)
	if err != nil {
		delete(c.blocks, key)
	}
	for len(c.blocks) > 128 {
		var oldest dimensionStructureBlockKey
		used, found := ^uint64(0), false
		for position, candidate := range c.blocks {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(c.blocks, oldest)
	}
	c.mu.Unlock()
	return result, err
}

type dimensionStructureRecordingWorld struct{ *structureRecordingWorld }

func (w *dimensionStructureRecordingWorld) MaxBuildHeight() int32 { return w.region.MaxBuildHeight() }

func (g *DimensionGenerator) generateStructureBlocks(start StructureStart, cx, cz int32) (*structureBlockResult, error) {
	order, err := OriginalStructureDecorationOrder(start.Structure)
	if err != nil {
		return nil, err
	}
	bases, chunks, err := g.decorationBases(cx, cz)
	if err != nil {
		return nil, err
	}
	var preceding []StructureStart
	for _, neighbor := range g.decorationStructures(bases, cx, cz) {
		previous, err := OriginalStructureDecorationOrder(neighbor.Structure)
		if err != nil {
			return nil, err
		}
		if previous.Step < order.Step {
			preceding = append(preceding, neighbor)
		}
	}
	region, _, err := g.decorateRegion(bases, chunks, cx, cz, order.Step, preceding)
	if err != nil {
		return nil, err
	}
	result := &structureBlockResult{start: start.Clone(), blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
	world := &dimensionStructureRecordingWorld{&structureRecordingWorld{region: region, result: result}}
	decorationSeed := OverworldDecorationSeed(g.seed, cx, cz)
	random := NewWorldgenRandom(decorationSeed + int64(order.Index) + 10000*int64(order.Step))
	worldRandom := NewStructureRegionRandom(g.seed, cx, cz)
	var jigsawPivot FeaturePosition
	if len(start.Pieces) > 0 {
		box := start.Pieces[0].Box
		jigsawPivot = FeaturePosition{box.MinX + (box.MaxX-box.MinX+1)/2, box.MinY, box.MinZ + (box.MaxZ-box.MinZ+1)/2}
	}
	for i := range result.start.Pieces {
		piece := &result.start.Pieces[i]
		b := piece.Box
		if b.MaxX < cx*16 || b.MinX > cx*16+15 || b.MaxZ < cz*16 || b.MinZ > cz*16+15 {
			continue
		}
		var placed StructurePiecePostprocessResult
		switch start.Structure {
		case "minecraft:end_city":
			placed, err = PostProcessEndCityBlocks(world, piece, random, cx, cz)
		case "minecraft:fortress":
			placed, err = PostProcessNetherFortressBlocks(world, piece, random, cx, cz)
		case "minecraft:ruined_portal_nether":
			placed, err = PostProcessRuinedPortalBlocks(world, piece, random, cx, cz)
		case "minecraft:nether_fossil":
			placed, err = PostProcessNetherFossilBlocks(world, piece, random, cx, cz)
		case "minecraft:bastion_remnant":
			placed, err = PostProcessJigsawBlocks(world, piece, random, cx, cz, jigsawPivot)
		default:
			return nil, fmt.Errorf("unsupported %s structure block placer %s", g.dimension, start.Structure)
		}
		if err != nil {
			return nil, fmt.Errorf("postprocess %s in %d,%d: %w", start.Structure, cx, cz, err)
		}
		if start.Structure == "minecraft:bastion_remnant" {
			for i := range placed.PendingEntities {
				request := &placed.PendingEntities[i]
				request.Data, err = FinalizeBastionTemplateEntity(*request, worldRandom)
				if err != nil {
					return nil, err
				}
			}
		}
		result.pending = append(result.pending, placed.PendingEntities...)
	}
	return result, nil
}

func (g *DimensionGenerator) applyStructureStep(region *OverworldFeatureRegion, starts []StructureStart, cx, cz, step int32) error {
	for _, start := range starts {
		order, err := OriginalStructureDecorationOrder(start.Structure)
		if err != nil {
			return err
		}
		if order.Step != step {
			continue
		}
		box, valid := structureWriteBox(start)
		if !valid || box.MaxX < cx*16 || box.MinX > cx*16+15 || box.MaxZ < cz*16 || box.MinZ > cz*16+15 {
			continue
		}
		cached, err := g.structureBlocksFor(start, cx, cz)
		if err != nil {
			return err
		}
		for p, state := range cached.blocks {
			if p.X>>4 == cx && p.Z>>4 == cz {
				if err := region.SetBlock(p, state); err != nil {
					return err
				}
			}
		}
		for p := range cached.postProcessing {
			if p.X>>4 == cx && p.Z>>4 == cz {
				if err := region.MarkPostProcessing(p); err != nil {
					return err
				}
			}
		}
		for p, entity := range cached.blockEntities {
			if p.X>>4 == cx && p.Z>>4 == cz {
				if err := region.SetBlockEntity(p, entity.block, entity.data); err != nil {
					return err
				}
			}
		}
		for _, entity := range cached.pending {
			if entity.Position.X>>4 == cx && entity.Position.Z>>4 == cz {
				if err := region.AddPendingEntity(entity); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (g *DimensionGenerator) generatedStructureNBT(cx, cz int32, starts []StructureStart) (nbt.Object, error) {
	var owned []StructureStart
	for _, start := range starts {
		if start.Chunk.X == cx && start.Chunk.Z == cz {
			cached, err := g.structureBlocksFor(start, cx, cz)
			if err != nil {
				return nil, err
			}
			owned = append(owned, cached.start.Clone())
		}
	}
	return StructureNBTForChunk(cx, cz, owned, BuildStructureReferences(cx, cz, starts))
}
