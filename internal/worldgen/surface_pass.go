package worldgen

import (
	"fmt"
	"math"
)

func (result *NoiseChunkResult) BlockStateAt(x, y, z int32) (SurfaceBlockState, error) {
	index, err := result.Index(x, y, z)
	if err != nil {
		return SurfaceBlockState{}, err
	}
	if state, exists := result.SurfaceBlocks[index]; exists {
		return state, nil
	}
	return SurfaceBlockState{Name: result.Blocks[index].BlockName()}, nil
}

func (result *NoiseChunkResult) surfaceState(x, y, z int32) SurfaceBlockState {
	if y < result.MinY || y >= result.MinY+result.Height {
		return SurfaceBlockState{Name: "minecraft:air"}
	}
	index := int((y-result.MinY)*256 + z*16 + x)
	if state, exists := result.SurfaceBlocks[index]; exists {
		return state
	}
	return SurfaceBlockState{Name: result.Blocks[index].BlockName()}
}

func (result *NoiseChunkResult) setSurfaceState(x, y, z int32, state SurfaceBlockState) {
	index := int((y-result.MinY)*256 + z*16 + x)
	if result.SurfaceBlocks == nil {
		result.SurfaceBlocks = make(map[int]SurfaceBlockState)
	}
	result.SurfaceBlocks[index] = state
}

func surfaceAir(state SurfaceBlockState) bool {
	return state.Name == "minecraft:air" || state.Name == "minecraft:cave_air" || state.Name == "minecraft:void_air"
}

func surfaceFluid(state SurfaceBlockState) bool {
	return state.Name == "minecraft:water" || state.Name == "minecraft:lava"
}

func surfaceSolid(state SurfaceBlockState) bool {
	return !surfaceAir(state) && !surfaceFluid(state)
}

func (result *NoiseChunkResult) columnHeight(x, z int32) int32 {
	for y := result.MinY + result.Height - 1; y >= result.MinY; y-- {
		if !surfaceAir(result.surfaceState(x, y, z)) {
			return y + 1
		}
	}
	return result.MinY
}

func (generator *OverworldNoiseGenerator) ApplySurface(result *NoiseChunkResult, biomeAtQuart func(int32, int32, int32) (string, error)) error {
	if result == nil || biomeAtQuart == nil || result.MinY != generator.settings.MinY || result.Height != generator.settings.Height || len(result.Blocks) != int(result.Height)*256 {
		return fmt.Errorf("surface pass requires a complete matching noise chunk and biome source")
	}
	rules, err := NewSurfaceRuleSampler(generator.seed, generator.settings)
	if err != nil {
		return err
	}
	density := NewNoiseChunkDensitySampler(generator.seed, generator.settings, result.X, result.Z)
	aquifer := newNoiseAquifer(density, generator.settings, PositionalRandom{})
	var heights [16][16]int32
	for x := int32(0); x < 16; x++ {
		for z := int32(0); z < 16; z++ {
			heights[x][z] = result.columnHeight(x, z)
		}
	}
	type quartPosition struct{ x, y, z int32 }
	biomes := make(map[quartPosition]string)
	for x := int32(0); x < 16; x++ {
		for z := int32(0); z < 16; z++ {
			blockX, blockZ := result.X*16+x, result.Z*16+z
			depth, err := rules.SurfaceDepth(blockX, blockZ)
			if err != nil {
				return err
			}
			secondary, err := rules.SecondaryDepth(blockX, blockZ)
			if err != nil {
				return err
			}
			getBiome := func(y int32) (string, error) {
				qx, qy, qz := BiomeQuartAtBlock(generator.seed, blockX, y, blockZ)
				position := quartPosition{x: qx, y: qy, z: qz}
				if biome, exists := biomes[position]; exists {
					return biome, nil
				}
				biome, err := biomeAtQuart(qx, qy, qz)
				if err == nil {
					biomes[position] = biome
				}
				return biome, err
			}
			initialHeight := heights[x][z]
			initialBiome, err := getBiome(initialHeight)
			if err != nil {
				return err
			}
			if initialBiome == "minecraft:eroded_badlands" {
				if err := generator.applyErodedBadlandsColumn(result, rules, x, z, initialHeight); err != nil {
					return err
				}
			}
			heights[x][z] = result.columnHeight(x, z)
			preliminary := func(sampleX, sampleZ int32) int32 { return aquifer.preliminarySurface(sampleX, sampleZ) }
			cornerX, cornerZ := blockX&^15, blockZ&^15
			corners := [4]int32{
				preliminary(cornerX, cornerZ), preliminary(cornerX+16, cornerZ),
				preliminary(cornerX, cornerZ+16), preliminary(cornerX+16, cornerZ+16),
			}
			if aquifer.err != nil {
				return aquifer.err
			}
			fractionX, fractionZ := float64(blockX&15)/16, float64(blockZ&15)/16
			lerp := func(fraction, start, end float64) float64 { return start + fraction*(end-start) }
			level := int32(math.Floor(lerp(fractionZ,
				lerp(fractionX, float64(corners[0]), float64(corners[1])),
				lerp(fractionX, float64(corners[2]), float64(corners[3]))))) + depth - 8
			steep := func() bool {
				lowZ, highZ := max(z-1, 0), min(z+1, 15)
				if heights[x][highZ] >= heights[x][lowZ]+4 {
					return true
				}
				lowX, highX := max(x-1, 0), min(x+1, 15)
				return heights[lowX][z] >= heights[highX][z]+4
			}
			stoneAbove, waterHeight := int32(0), int32(math.MinInt32)
			stoneCeiling := int32(math.MaxInt32)
			for y := heights[x][z]; y >= result.MinY; y-- {
				state := result.surfaceState(x, y, z)
				if surfaceAir(state) {
					stoneAbove, waterHeight = 0, math.MinInt32
					continue
				}
				if surfaceFluid(state) {
					if waterHeight == math.MinInt32 {
						waterHeight = y + 1
					}
					continue
				}
				if stoneCeiling >= y {
					stoneCeiling = aquiferDryLevel
					for below := y - 1; below >= result.MinY-1; below-- {
						if surfaceSolid(result.surfaceState(x, below, z)) {
							continue
						}
						stoneCeiling = below + 1
						break
					}
				}
				biome, err := getBiome(y)
				if err != nil {
					return err
				}
				cold, err := BiomeColdAt(biome, blockX, y, blockZ)
				if err != nil {
					return err
				}
				context := SurfaceContext{
					X: blockX, Y: y, Z: blockZ, Biome: biome,
					SurfaceDepth: depth, SecondaryDepth: secondary,
					StoneDepthAbove: stoneAbove + 1, StoneDepthBelow: y - stoneCeiling + 1,
					WaterHeight: waterHeight, MinSurfaceLevel: level,
					Cold: cold, Steep: steep(),
				}
				stoneAbove++
				if state.Name != "minecraft:stone" {
					continue
				}
				replacement, err := rules.applyRule(rules.rule, context)
				if err != nil {
					return err
				}
				if replacement != nil {
					result.setSurfaceState(x, y, z, *replacement)
				}
			}
			if initialBiome == "minecraft:frozen_ocean" || initialBiome == "minecraft:deep_frozen_ocean" {
				if err := generator.applyFrozenOceanColumn(result, rules, x, z, initialHeight, level, initialBiome); err != nil {
					return err
				}
				heights[x][z] = result.columnHeight(x, z)
			}
		}
	}
	return nil
}

func (generator *OverworldNoiseGenerator) applyErodedBadlandsColumn(result *NoiseChunkResult, rules *SurfaceRuleSampler, x, z, initialHeight int32) error {
	blockX, blockZ := result.X*16+x, result.Z*16+z
	pillar, err := rules.noise("minecraft:badlands_pillar")
	if err != nil {
		return err
	}
	roof, err := rules.noise("minecraft:badlands_pillar_roof")
	if err != nil {
		return err
	}
	surface, err := rules.noise("minecraft:badlands_surface")
	if err != nil {
		return err
	}
	strength := math.Min(math.Abs(surface.Noise(float64(blockX), 0, float64(blockZ))*8.25),
		pillar.Noise(float64(blockX)*0.2, 0, float64(blockZ)*0.2)*15)
	if strength <= 0 {
		return nil
	}
	variation := math.Abs(roof.Noise(float64(blockX)*0.75, 0, float64(blockZ)*0.75) * 1.5)
	height := int32(math.Floor(64 + math.Min(strength*strength*2.5, math.Ceil(variation*50)+24)))
	if initialHeight > height {
		return nil
	}
	for y := height; y >= result.MinY; y-- {
		state := result.surfaceState(x, y, z)
		if state.Name == "minecraft:stone" {
			break
		}
		if state.Name == "minecraft:water" {
			return nil
		}
	}
	for y := height; y >= result.MinY && surfaceAir(result.surfaceState(x, y, z)); y-- {
		result.setSurfaceState(x, y, z, SurfaceBlockState{Name: "minecraft:stone"})
	}
	return nil
}

func (generator *OverworldNoiseGenerator) applyFrozenOceanColumn(result *NoiseChunkResult, rules *SurfaceRuleSampler, x, z, initialHeight, minSurfaceLevel int32, biome string) error {
	blockX, blockZ := result.X*16+x, result.Z*16+z
	pillar, err := rules.noise("minecraft:iceberg_pillar")
	if err != nil {
		return err
	}
	roof, err := rules.noise("minecraft:iceberg_pillar_roof")
	if err != nil {
		return err
	}
	surface, err := rules.noise("minecraft:iceberg_surface")
	if err != nil {
		return err
	}
	strength := math.Min(math.Abs(surface.Noise(float64(blockX), 0, float64(blockZ))*8.25),
		pillar.Noise(float64(blockX)*1.28, 0, float64(blockZ)*1.28)*15)
	if strength <= 1.8 {
		return nil
	}
	variation := math.Abs(roof.Noise(float64(blockX)*1.17, 0, float64(blockZ)*1.17) * 1.5)
	iceHeight := math.Min(strength*strength*1.2, math.Ceil(variation*40)+14)
	temperature, err := BiomeTemperatureAt(biome, blockX, generator.settings.SeaLevel, blockZ)
	if err != nil {
		return err
	}
	if temperature > float32(0.1) {
		iceHeight -= 2
	}
	top, bottom := 0.0, 0.0
	if iceHeight > 2 {
		bottom = float64(generator.settings.SeaLevel) - iceHeight - 7
		top = iceHeight + float64(generator.settings.SeaLevel)
	}
	random := rules.random.At(blockX, 0, blockZ)
	snowLimit := 2 + random.NextIntBound(4)
	snowHeight := generator.settings.SeaLevel + 18 + random.NextIntBound(10)
	snowCount := int32(0)
	for y := max(initialHeight, int32(top)+1); y >= minSurfaceLevel; y-- {
		state := result.surfaceState(x, y, z)
		place := surfaceAir(state) && y < int32(top) && random.NextDouble() > 0.01
		if !place && state.Name == "minecraft:water" && y > int32(bottom) && y < generator.settings.SeaLevel && bottom != 0 {
			place = random.NextDouble() > 0.15
		}
		if !place {
			continue
		}
		if snowCount <= snowLimit && y > snowHeight {
			result.setSurfaceState(x, y, z, SurfaceBlockState{Name: "minecraft:snow_block"})
			snowCount++
		} else {
			result.setSurfaceState(x, y, z, SurfaceBlockState{Name: "minecraft:packed_ice"})
		}
	}
	return nil
}

func (generator *OverworldNoiseGenerator) GenerateSurface(cx, cz int32) (*NoiseChunkResult, error) {
	result, err := generator.GenerateNoise(cx, cz)
	if err != nil {
		return nil, err
	}
	biomes, err := newOverworldBiomeSampler(generator.seed, generator.settings)
	if err != nil {
		return nil, err
	}
	if err := generator.ApplySurface(result, biomes.BiomeAtQuart); err != nil {
		return nil, err
	}
	return result, nil
}
