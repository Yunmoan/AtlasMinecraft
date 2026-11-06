package worldgen

import "fmt"

func (g *OverworldGenerator) generateMansionStructureBlocks(start StructureStart, cx, cz int32) (*structureBlockResult, error) {
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
	random := NewWorldgenRandom(OverworldDecorationSeed(g.seed, cx, cz) + int64(order.Index) + 10000*int64(order.Step))
	regionRandom := NewStructureRegionRandom(g.seed, cx, cz)
	for i := range result.start.Pieces {
		piece := &result.start.Pieces[i]
		box := piece.Box
		if box.MaxX < cx*16 || box.MinX > cx*16+15 || box.MaxZ < cz*16 || box.MinZ > cz*16+15 {
			continue
		}
		placed, err := PostProcessMansionBlocks(world, piece, random, regionRandom, cx, cz)
		if err != nil {
			return nil, fmt.Errorf("mansion piece %d in %d,%d: %w", i, cx, cz, err)
		}
		for actorIndex := range placed.PendingEntities {
			request := &placed.PendingEntities[actorIndex]
			request.SavedFlag = fmt.Sprintf("Mansion:%d,%d:%d:%s", start.Chunk.X, start.Chunk.Z, i, request.SavedFlag)
		}
		result.pending = append(result.pending, placed.PendingEntities...)
	}
	clip := StructureBoundingBox{cx * 16, region.minY, cz * 16, cx*16 + 15, region.MaxBuildHeight() - 1, cz*16 + 15}
	if err := PostProcessMansionFoundation(world, &result.start, clip); err != nil {
		return nil, err
	}
	return result, nil
}
