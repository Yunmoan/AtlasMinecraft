package worldgen

import (
	"fmt"
	"math"
)

type NoiseMaterial uint8

const (
	NoiseAir NoiseMaterial = iota
	NoiseStone
	NoiseWater
	NoiseLava
	NoiseGranite
	NoiseTuff
	NoiseCopperOre
	NoiseRawCopper
	NoiseDeepslateIronOre
	NoiseRawIron
	NoiseNetherrack
	NoiseEndStone
)

func (material NoiseMaterial) BlockName() string {
	names := [...]string{
		"minecraft:air", "minecraft:stone", "minecraft:water", "minecraft:lava",
		"minecraft:granite", "minecraft:tuff", "minecraft:copper_ore", "minecraft:raw_copper_block",
		"minecraft:deepslate_iron_ore", "minecraft:raw_iron_block",
		"minecraft:netherrack", "minecraft:end_stone",
	}
	if int(material) >= len(names) {
		return ""
	}
	return names[material]
}

type NoiseChunkResult struct {
	X, Z                int32
	MinY, Height        int32
	Blocks              []NoiseMaterial
	FluidUpdates        []bool
	PostProcessing      []bool
	CarvingMask         []bool
	CarvingMaskByCarver [3]int
	SurfaceBlocks       map[int]SurfaceBlockState
}

func (result *NoiseChunkResult) Index(x, y, z int32) (int, error) {
	if x < 0 || x >= 16 || z < 0 || z >= 16 || y < result.MinY || y >= result.MinY+result.Height {
		return 0, fmt.Errorf("noise chunk position outside bounds: %d,%d,%d", x, y, z)
	}
	return int((y-result.MinY)*256 + z*16 + x), nil
}

type OverworldNoiseGenerator struct {
	seed     int64
	settings *DensitySettings
}

func NewOverworldNoiseGenerator(seed int64) (*OverworldNoiseGenerator, error) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		return nil, err
	}
	return &OverworldNoiseGenerator{seed: seed, settings: settings}, nil
}

func (generator *OverworldNoiseGenerator) GenerateNoise(cx, cz int32) (*NoiseChunkResult, error) {
	return generateNoiseChunk(generator.seed, generator.settings, cx, cz, nil)
}

func generateNoiseChunk(seed int64, settings *DensitySettings, cx, cz int32, beard *StructureBeardifier) (*NoiseChunkResult, error) {
	if cx < -1875000 || cx >= 1875000 || cz < -1875000 || cz >= 1875000 {
		return nil, fmt.Errorf("noise chunk %d,%d is outside vanilla's horizontal bounds", cx, cz)
	}
	sampler := NewNoiseChunkDensitySampler(seed, settings, cx, cz)
	sampler.beardifier = beard
	finalDensity := settings.Router["final_density"]
	if beard != nil {
		finalDensity = &DensityNode{Kind: "add", Children: map[string]*DensityNode{
			"argument1": finalDensity, "argument2": {Kind: "beardifier"},
		}}
		finalDensity.Children["argument2"].evaluation = prepareDensityEvaluation(finalDensity.Children["argument2"])
		finalDensity.evaluation = prepareDensityEvaluation(finalDensity)
	}
	density := &DensityNode{Kind: "cache_all_in_cell", Children: map[string]*DensityNode{
		"argument": finalDensity,
	}}
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
	result := &NoiseChunkResult{
		X: cx, Z: cz, MinY: settings.MinY, Height: settings.Height,
		Blocks:       make([]NoiseMaterial, settings.Height*256),
		FluidUpdates: make([]bool, settings.Height*256),
	}
	width, height := settings.CellWidth, settings.CellHeight
	for cellX := int32(0); cellX < 16/width; cellX++ {
		for cellZ := int32(0); cellZ < 16/width; cellZ++ {
			for cellY := settings.Height/height - 1; cellY >= 0; cellY-- {
				for localY := height - 1; localY >= 0; localY-- {
					y := settings.MinY + cellY*height + localY
					for localX := int32(0); localX < width; localX++ {
						x := cellX*width + localX
						for localZ := int32(0); localZ < width; localZ++ {
							z := cellZ*width + localZ
							blockX, blockZ := cx*16+x, cz*16+z
							value, err := sampler.Sample(density, blockX, y, blockZ)
							if err != nil {
								return nil, fmt.Errorf("final density at %d,%d,%d: %w", blockX, y, blockZ, err)
							}
							material, solid, update := aquifer.compute(blockX, y, blockZ, value)
							if aquifer.err != nil {
								return nil, aquifer.err
							}
							if solid && settings.OreVeins {
								material, err = sampleOreVein(sampler, settings, oreRandom, blockX, y, blockZ)
								if err != nil {
									return nil, err
								}
							} else if solid {
								material = defaultBlock
							}
							index := int((y-settings.MinY)*256 + z*16 + x)
							result.Blocks[index], result.FluidUpdates[index] = material, update
						}
					}
				}
			}
		}
	}
	return result, nil
}

func sampleOreVein(sampler *DensitySampler, settings *DensitySettings, random PositionalRandom, x, y, z int32) (NoiseMaterial, error) {
	vein, err := sampler.Sample(settings.Router["vein_toggle"], x, y, z)
	if err != nil {
		return NoiseStone, err
	}
	minY, maxY := int32(-60), int32(-8)
	ore, raw, filler := NoiseDeepslateIronOre, NoiseRawIron, NoiseTuff
	if vein > 0 {
		minY, maxY = 0, 50
		ore, raw, filler = NoiseCopperOre, NoiseRawCopper, NoiseGranite
	}
	if y < minY || y > maxY {
		return NoiseStone, nil
	}
	distance := math.Min(float64(maxY-y), float64(y-minY))
	roundoff := clampedMap(distance, 0, 20, -0.2, 0)
	strength := math.Abs(vein)
	if strength+roundoff < float64(float32(0.4)) {
		return NoiseStone, nil
	}
	positionRandom := random.At(x, y, z)
	if positionRandom.NextFloat() > float32(0.7) {
		return NoiseStone, nil
	}
	thickness, err := sampler.Sample(settings.Router["vein_ridged"], x, y, z)
	if err != nil || thickness >= 0 {
		return NoiseStone, err
	}
	chance := clampedMap(strength, float64(float32(0.4)), float64(float32(0.6)), float64(float32(0.1)), float64(float32(0.3)))
	if float64(positionRandom.NextFloat()) < chance {
		gap, err := sampler.Sample(settings.Router["vein_gap"], x, y, z)
		if err != nil {
			return NoiseStone, err
		}
		if gap > float64(float32(-0.3)) {
			if positionRandom.NextFloat() < float32(0.02) {
				return raw, nil
			}
			return ore, nil
		}
	}
	return filler, nil
}

func clampedMap(value, fromMin, fromMax, toMin, toMax float64) float64 {
	fraction := (value - fromMin) / (fromMax - fromMin)
	if fraction < 0 {
		return toMin
	}
	if fraction > 1 {
		return toMax
	}
	return toMin + fraction*(toMax-toMin)
}
