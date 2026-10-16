package worldgen

import "fmt"

// DimensionNoiseGenerator fills the official Nether or End noise stage.
// Surface, biome assignment, carvers, structures and decorations are separate
// stages; this API does not advertise a complete dimension generator.
type DimensionNoiseGenerator struct {
	seed     int64
	settings *DensitySettings
}

func NewDimensionNoiseGenerator(seed int64, dimension string) (*DimensionNoiseGenerator, error) {
	if dimension != "nether" && dimension != "end" {
		return nil, fmt.Errorf("unsupported dimension noise settings %q", dimension)
	}
	settings, err := LoadDensitySettings(dimension)
	if err != nil {
		return nil, err
	}
	return &DimensionNoiseGenerator{seed: seed, settings: settings}, nil
}

func (g *DimensionNoiseGenerator) GenerateNoise(cx, cz int32) (*NoiseChunkResult, error) {
	return generateNoiseChunk(g.seed, g.settings, cx, cz, nil)
}

func (g *DimensionNoiseGenerator) GenerateNoiseWithStructures(cx, cz int32, pieces []StructureTerrainPiece) (*NoiseChunkResult, error) {
	beard, err := NewStructureBeardifierForChunk(cx, cz, pieces)
	if err != nil {
		return nil, err
	}
	return generateNoiseChunk(g.seed, g.settings, cx, cz, beard)
}

func (g *DimensionNoiseGenerator) GenerateSurface(cx, cz int32) (*NoiseChunkResult, error) {
	chunk, err := g.GenerateNoise(cx, cz)
	if err != nil {
		return nil, err
	}
	biomes, err := NewDimensionBiomeSampler(g.seed, g.settings.Name[len("minecraft:"):])
	if err != nil {
		return nil, err
	}
	if err := g.ApplySurface(chunk, biomes.BiomeAtQuart); err != nil {
		return nil, err
	}
	return chunk, nil
}

func (g *DimensionNoiseGenerator) ApplySurface(chunk *NoiseChunkResult, biomeAtQuart func(int32, int32, int32) (string, error)) error {
	pass := &OverworldNoiseGenerator{seed: g.seed, settings: g.settings}
	return pass.ApplySurface(chunk, biomeAtQuart)
}

// GenerateNoiseWithStructures accepts actual neighboring starts/references.
// Callers must produce those before filling density; placement candidates alone
// are insufficient inputs to structure-influenced terrain.
func (generator *OverworldNoiseGenerator) GenerateNoiseWithStructures(cx, cz int32, pieces []StructureTerrainPiece) (*NoiseChunkResult, error) {
	beard, err := NewStructureBeardifierForChunk(cx, cz, pieces)
	if err != nil {
		return nil, err
	}
	return generateNoiseChunk(generator.seed, generator.settings, cx, cz, beard)
}
