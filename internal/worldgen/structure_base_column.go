package worldgen

import "fmt"

// StructureBaseNoiseColumn is the original NoiseBasedChunkGenerator base
// column before surface/carvers, including noise ore veins. Heights are occupied
// block Y values.
type StructureBaseNoiseColumn struct {
	MinY   int32
	Blocks []NoiseMaterial
}

func (column *StructureBaseNoiseColumn) MaterialAt(y int32) NoiseMaterial {
	if column == nil || y < column.MinY || int64(y-column.MinY) >= int64(len(column.Blocks)) {
		return NoiseAir
	}
	return column.Blocks[y-column.MinY]
}

func (column *StructureBaseNoiseColumn) OccupiedHeight(oceanFloor bool) int32 {
	for index := len(column.Blocks) - 1; index >= 0; index-- {
		material := column.Blocks[index]
		if material != NoiseAir && (!oceanFloor || material != NoiseWater && material != NoiseLava) {
			return column.MinY + int32(index)
		}
	}
	return column.MinY - 1
}

type StructureBaseNoiseColumnAt func(x, z int32) (*StructureBaseNoiseColumn, error)

func (generator *OverworldNoiseGenerator) BaseStructureColumnAt(x, z int32) (*StructureBaseNoiseColumn, error) {
	return generateStructureBaseColumn(generator.seed, generator.settings, x, z)
}

func (generator *DimensionNoiseGenerator) BaseStructureColumnAt(x, z int32) (*StructureBaseNoiseColumn, error) {
	return generateStructureBaseColumn(generator.seed, generator.settings, x, z)
}

func generateStructureBaseColumn(seed int64, settings *DensitySettings, x, z int32) (*StructureBaseNoiseColumn, error) {
	if x < -30000000 || x >= 30000000 || z < -30000000 || z >= 30000000 {
		return nil, fmt.Errorf("structure base column outside horizontal bounds: %d,%d", x, z)
	}
	sampler := NewNoiseChunkDensitySampler(seed, settings, x>>4, z>>4)
	sampler.chunk.minX = floorCell(x, settings.CellWidth)
	sampler.chunk.minZ = floorCell(z, settings.CellWidth)
	density := &DensityNode{Kind: "cache_all_in_cell", Children: map[string]*DensityNode{"argument": settings.Router["final_density"]}}
	density.evaluation = prepareDensityEvaluation(density)
	rootRandom := NewXoroshiro(seed).ForkPositional()
	aquifer := newNoiseAquifer(sampler, settings, rootRandom.FromHashOf("minecraft:aquifer").ForkPositional())
	oreRandom := rootRandom.FromHashOf("minecraft:ore").ForkPositional()
	defaultBlock := NoiseStone
	switch settings.DefaultBlock.Name {
	case "minecraft:netherrack":
		defaultBlock = NoiseNetherrack
	case "minecraft:end_stone":
		defaultBlock = NoiseEndStone
	}
	result := &StructureBaseNoiseColumn{MinY: settings.MinY, Blocks: make([]NoiseMaterial, settings.Height)}
	for y := settings.MinY + settings.Height - 1; y >= settings.MinY; y-- {
		value, err := sampler.Sample(density, x, y, z)
		if err != nil {
			return nil, err
		}
		material, solid, _ := aquifer.compute(x, y, z, value)
		if aquifer.err != nil {
			return nil, aquifer.err
		}
		if solid && settings.OreVeins {
			material, err = sampleOreVein(sampler, settings, oreRandom, x, y, z)
			if err != nil {
				return nil, err
			}
		} else if solid {
			material = defaultBlock
		}
		result.Blocks[y-settings.MinY] = material
	}
	return result, nil
}
