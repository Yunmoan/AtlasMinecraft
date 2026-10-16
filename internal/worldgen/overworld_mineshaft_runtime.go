package worldgen

import (
	"fmt"
	"sort"
)

func isOverworldMineshaftStructure(name string) bool {
	return name == "minecraft:mineshaft" || name == "minecraft:mineshaft_mesa"
}

func (w *structureRecordingWorld) MaxBuildHeight() int32 {
	return w.region.minY + w.region.height
}

func (w *structureRecordingWorld) MineshaftBiomeAt(p FeaturePosition) (string, error) {
	if w.region.biomeAtQuart == nil {
		return "", fmt.Errorf("mineshaft region has no biome source")
	}
	qx, qy, qz := BiomeQuartAtBlock(w.region.biomeSeed, p.X, p.Y, p.Z)
	return w.region.biomeAtQuart(qx, qy, qz)
}

// A single private start is replayed in canonical chunk order. Corridors retain
// their hps flag across clips, so a spider corridor creates at most one spawner
// and later clips consume the original continuation of their random streams.
func (g *OverworldGenerator) generateMineshaftStructureBlocks(start StructureStart) (*structureBlockResult, error) {
	order, err := OriginalStructureDecorationOrder(start.Structure)
	if err != nil {
		return nil, err
	}
	_, valid := start.BoundingBox()
	if !valid {
		return nil, fmt.Errorf("mineshaft start has no pieces")
	}
	region, err := g.structureRegionBefore(start.Chunk.X, start.Chunk.Z, order.Step)
	if err != nil {
		return nil, err
	}
	var biomes *OverworldBiomeSampler
	region.biomeAtQuart = func(qx, qy, qz int32) (string, error) {
		qy = max(int32(-16), min(int32(79), qy))
		if base := region.biomeBases[featureChunkPosition{qx >> 2, qz >> 2}]; base != nil {
			return base.biomes[(qy+16)>>2][((qy&3)<<4)|((qz&3)<<2)|(qx&3)], nil
		}
		// BiomeManager's fuzzy quart selection can choose a neighbor across a
		// clipped piece edge. That biome needs no terrain snapshot or decoration.
		if biomes == nil {
			var err error
			biomes, err = newOverworldBiomeSampler(g.seed, g.noise.settings)
			if err != nil {
				return "", err
			}
		}
		return biomes.BiomeAtQuart(qx, qy, qz)
	}
	return g.replayMineshaftStructureBlocks(start, region, order)
}

func (g *OverworldGenerator) replayMineshaftStructureBlocks(start StructureStart, region *OverworldFeatureRegion, order StructureDecorationOrder) (*structureBlockResult, error) {
	result := &structureBlockResult{start: start.Clone(), blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
	world := &structureRecordingWorld{region: region, result: result}
	// Mine readers stay inside the current X/Z clip: the expanded liquid test
	// is clipped, supports are vertical and cobweb neighbors have clip guards.
	// Only clips touched by pieces need terrain. Empty gaps and a surrounding
	// chunk ring cannot influence writes, flags or the per-clip random stream.
	clips := make(map[featureChunkPosition][]int)
	for index, piece := range result.start.Pieces {
		for x := piece.Box.MinX >> 4; x <= piece.Box.MaxX>>4; x++ {
			for z := piece.Box.MinZ >> 4; z <= piece.Box.MaxZ>>4; z++ {
				key := featureChunkPosition{x, z}
				clips[key] = append(clips[key], index)
			}
		}
	}
	positions := make([]featureChunkPosition, 0, len(clips))
	for key := range clips {
		positions = append(positions, key)
	}
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].chunkX < positions[j].chunkX || positions[i].chunkX == positions[j].chunkX && positions[i].chunkZ < positions[j].chunkZ
	})
	for _, key := range positions {
		if region.chunks[key] == nil {
			// Retain the same preceding-stage snapshot as the original full-box
			// replay; bare noise would lose caves, lakes and prior structures.
			remote, err := g.structureChunkBefore(key.chunkX, key.chunkZ, order.Step)
			if err != nil {
				return nil, err
			}
			region.installStructureChunk(key, remote)
		}
		seed := OverworldDecorationSeed(g.seed, key.chunkX, key.chunkZ)
		random := NewWorldgenRandom(seed + int64(order.Index) + 10000*int64(order.Step))
		for _, index := range clips[key] {
			piece := &result.start.Pieces[index]
			placed, err := PostProcessMineshaftBlocks(world, piece, random, key.chunkX, key.chunkZ)
			if err != nil {
				return nil, fmt.Errorf("postprocess %s in %d,%d: %w", start.Structure, key.chunkX, key.chunkZ, err)
			}
			result.pending = append(result.pending, placed.PendingEntities...)
		}
	}
	return result, nil
}
