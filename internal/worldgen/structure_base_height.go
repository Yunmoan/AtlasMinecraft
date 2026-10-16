package worldgen

import "fmt"

// BaseOccupiedHeightAt follows NoiseBasedChunkGenerator's single-column
// WORLD_SURFACE_WG query, used during structure-start checks before surface and
// carving. It evaluates one noise cell's column and stops at the highest
// non-air material; it never allocates or generates an entire 16x16 chunk.
func (generator *OverworldNoiseGenerator) BaseOccupiedHeightAt(x, z int32) (int32, error) {
	if x < -30000000 || x >= 30000000 || z < -30000000 || z >= 30000000 {
		return 0, fmt.Errorf("structure base-height column %d,%d is outside vanilla's horizontal bounds", x, z)
	}
	settings := generator.settings
	sampler := NewNoiseChunkDensitySampler(generator.seed, settings, x>>4, z>>4)
	// Vanilla's column query constructs a NoiseChunk with one horizontal
	// noise cell, anchored to that cell rather than the containing chunk.
	sampler.chunk.minX = floorCell(x, settings.CellWidth)
	sampler.chunk.minZ = floorCell(z, settings.CellWidth)
	density := &DensityNode{Kind: "cache_all_in_cell", Children: map[string]*DensityNode{
		"argument": settings.Router["final_density"],
	}}
	density.evaluation = prepareDensityEvaluation(density)
	random := NewXoroshiro(generator.seed).ForkPositional().FromHashOf("minecraft:aquifer").ForkPositional()
	aquifer := newNoiseAquifer(sampler, settings, random)
	for y := settings.MinY + settings.Height - 1; y >= settings.MinY; y-- {
		value, err := sampler.Sample(density, x, y, z)
		if err != nil {
			return 0, fmt.Errorf("structure base density at %d,%d,%d: %w", x, y, z, err)
		}
		material, solid, _ := aquifer.compute(x, y, z, value)
		if aquifer.err != nil {
			return 0, aquifer.err
		}
		if solid || material != NoiseAir {
			return y, nil
		}
	}
	return settings.MinY - 1, nil
}
