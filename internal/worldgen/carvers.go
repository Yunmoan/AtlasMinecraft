package worldgen

import (
	"fmt"
	"math"
)

var carverReplaceable = map[string]bool{
	"minecraft:stone": true, "minecraft:granite": true, "minecraft:diorite": true,
	"minecraft:andesite": true, "minecraft:tuff": true, "minecraft:deepslate": true,
	"minecraft:dirt": true, "minecraft:grass_block": true, "minecraft:podzol": true,
	"minecraft:coarse_dirt": true, "minecraft:mycelium": true, "minecraft:rooted_dirt": true,
	"minecraft:moss_block": true, "minecraft:mud": true, "minecraft:muddy_mangrove_roots": true,
	"minecraft:sand": true, "minecraft:red_sand": true, "minecraft:suspicious_sand": true,
	"minecraft:terracotta": true, "minecraft:white_terracotta": true,
	"minecraft:orange_terracotta": true, "minecraft:magenta_terracotta": true,
	"minecraft:light_blue_terracotta": true, "minecraft:yellow_terracotta": true,
	"minecraft:lime_terracotta": true, "minecraft:pink_terracotta": true,
	"minecraft:gray_terracotta": true, "minecraft:light_gray_terracotta": true,
	"minecraft:cyan_terracotta": true, "minecraft:purple_terracotta": true,
	"minecraft:blue_terracotta": true, "minecraft:brown_terracotta": true,
	"minecraft:green_terracotta": true, "minecraft:red_terracotta": true,
	"minecraft:black_terracotta": true, "minecraft:iron_ore": true,
	"minecraft:deepslate_iron_ore": true, "minecraft:copper_ore": true,
	"minecraft:deepslate_copper_ore": true, "minecraft:water": true,
	"minecraft:gravel": true, "minecraft:suspicious_gravel": true,
	"minecraft:sandstone": true, "minecraft:red_sandstone": true,
	"minecraft:calcite": true, "minecraft:snow": true, "minecraft:packed_ice": true,
	"minecraft:raw_iron_block": true, "minecraft:raw_copper_block": true,
}

var carverSines = func() [65536]float32 {
	var values [65536]float32
	for index := range values {
		values[index] = float32(math.Sin(float64(index) * math.Pi * 2 / 65536))
	}
	return values
}()

func carverSin(angle float32) float32 {
	return carverSines[uint16(int32(angle*float32(10430.378)))]
}

func carverCos(angle float32) float32 {
	return carverSines[uint16(int32(angle*float32(10430.378)+16384))]
}

func carverUniform(random *LegacyRandom, lower, upper float32) float32 {
	return random.NextFloat()*(upper-lower) + lower
}

type carvingPass struct {
	result      *NoiseChunkResult
	aquifer     *noiseAquifer
	biome       func(int32, int32, int32) (string, error)
	rules       *SurfaceRuleSampler
	seed        int64
	carverIndex int
	err         error
}

func (generator *OverworldNoiseGenerator) ApplyCarvers(result *NoiseChunkResult, biomeAtQuart func(int32, int32, int32) (string, error)) error {
	if result == nil || biomeAtQuart == nil || result.MinY != generator.settings.MinY || result.Height != generator.settings.Height || len(result.Blocks) != int(result.Height)*256 {
		return fmt.Errorf("carving pass requires a complete matching surface chunk and biome source")
	}
	density := NewNoiseChunkDensitySampler(generator.seed, generator.settings, result.X, result.Z)
	aquiferRandom := NewXoroshiro(generator.seed).ForkPositional().FromHashOf("minecraft:aquifer").ForkPositional()
	rules, err := NewSurfaceRuleSampler(generator.seed, generator.settings)
	if err != nil {
		return err
	}
	pass := carvingPass{result: result, aquifer: newNoiseAquifer(density, generator.settings, aquiferRandom),
		biome: biomeAtQuart, rules: rules, seed: generator.seed}
	result.CarvingMask = make([]bool, len(result.Blocks))
	result.CarvingMaskByCarver = [3]int{}
	for dx := int32(-8); dx <= 8; dx++ {
		for dz := int32(-8); dz <= 8; dz++ {
			sourceX, sourceZ := result.X+dx, result.Z+dz
			for index, config := range overworldCarvers {
				random := carverRandom(generator.seed, sourceX, sourceZ, index)
				if random.NextFloat() > config.probability {
					continue
				}
				pass.carverIndex = index
				if index == 2 {
					pass.carveCanyon(random, sourceX, sourceZ)
				} else {
					pass.carveCaves(random, sourceX, sourceZ, index == 1)
				}
				if pass.err != nil {
					return pass.err
				}
			}
		}
	}
	return pass.aquifer.err
}

func (generator *OverworldNoiseGenerator) GenerateCarved(cx, cz int32) (*NoiseChunkResult, error) {
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
	if err := generator.ApplyCarvers(result, biomes.BiomeAtQuart); err != nil {
		return nil, err
	}
	return result, nil
}

func (pass *carvingPass) carveCaves(random *LegacyRandom, sourceX, sourceZ int32, extra bool) {
	tunnelCount := random.NextIntBound(random.NextIntBound(random.NextIntBound(15)+1) + 1)
	for tunnel := int32(0); tunnel < tunnelCount; tunnel++ {
		x := float64(sourceX*16 + random.NextIntBound(16))
		maxY := int32(180)
		if extra {
			maxY = 47
		}
		y := float64(pass.result.MinY + 8 + random.NextIntBound(maxY-(pass.result.MinY+8)+1))
		z := float64(sourceZ*16 + random.NextIntBound(16))
		horizontal := carverUniform(random, 0.7, 1.4)
		vertical := carverUniform(random, 0.8, 1.3)
		floor := carverUniform(random, -1, -0.4)
		branches := int32(1)
		if random.NextIntBound(4) == 0 {
			scale := carverUniform(random, 0.1, 0.9)
			radius := 1 + random.NextFloat()*6
			width := 1.5 + float64(carverSin(float32(1.5707964))*radius)
			pass.carveEllipsoid(x+1, y, z, width, width*float64(scale), floor, nil)
			branches += random.NextIntBound(4)
		}
		for branch := int32(0); branch < branches; branch++ {
			heading := random.NextFloat() * (float32(math.Pi) * 2)
			pitch := (random.NextFloat() - 0.5) / 4
			width := random.NextFloat()*2 + random.NextFloat()
			if random.NextIntBound(10) == 0 {
				width *= random.NextFloat()*random.NextFloat()*3 + 1
			}
			length := int32(112) - random.NextIntBound(28)
			pass.carveCaveTunnel(NewLegacyRandom(int64(random.NextLong())), x, y, z,
				float64(horizontal), float64(vertical), width, heading, pitch, 0, length, floor)
		}
	}
}

func (pass *carvingPass) carveCaveTunnel(random *LegacyRandom, x, y, z, horizontal, vertical float64,
	width, heading, pitch float32, step, length int32, floor float32) {
	split := random.NextIntBound(length/2) + length/4
	gentle := random.NextIntBound(6) == 0
	var headingVelocity, pitchVelocity float32
	for ; step < length; step++ {
		horizontalRadius := 1.5 + float64(carverSin(float32(math.Pi)*float32(step)/float32(length))*width)
		verticalRadius := horizontalRadius
		cosPitch := carverCos(pitch)
		x += float64(carverCos(heading) * cosPitch)
		y += float64(carverSin(pitch))
		z += float64(carverSin(heading) * cosPitch)
		if gentle {
			pitch *= 0.92
		} else {
			pitch *= 0.7
		}
		pitch += pitchVelocity * 0.1
		heading += headingVelocity * 0.1
		pitchVelocity *= 0.9
		headingVelocity *= 0.75
		pitchVelocity += (random.NextFloat() - random.NextFloat()) * random.NextFloat() * 2
		headingVelocity += (random.NextFloat() - random.NextFloat()) * random.NextFloat() * 4
		if step == split && width > 1 {
			firstSeed := random.NextLong()
			branchWidth := random.NextFloat()*0.5 + 0.5
			pass.carveCaveTunnel(NewLegacyRandom(int64(firstSeed)), x, y, z, horizontal, vertical,
				branchWidth, heading-float32(1.5707964), pitch/3, step, length, floor)
			secondSeed := random.NextLong()
			branchWidth = random.NextFloat()*0.5 + 0.5
			pass.carveCaveTunnel(NewLegacyRandom(int64(secondSeed)), x, y, z, horizontal, vertical,
				branchWidth, heading+float32(1.5707964), pitch/3, step, length, floor)
			return
		}
		if random.NextIntBound(4) == 0 {
			continue
		}
		if !pass.pathNearTarget(x, z, step, length, width) {
			return
		}
		pass.carveEllipsoid(x, y, z, horizontalRadius*horizontal, verticalRadius*vertical, floor, nil)
	}
}

func (pass *carvingPass) pathNearTarget(x, z float64, step, length int32, radius float32) bool {
	dx := x - float64(pass.result.X*16+8)
	dz := z - float64(pass.result.Z*16+8)
	remaining := float64(length - step)
	limit := float64(float32(radius+2) + 16)
	return dx*dx+dz*dz-remaining*remaining <= limit*limit
}

func (pass *carvingPass) carveEllipsoid(x, y, z, radius, verticalRadius float64, floor float32,
	skip func(float64, float64, float64, int32) bool) {
	result := pass.result
	centerX, centerZ := float64(result.X*16+8), float64(result.Z*16+8)
	limit := 16 + radius*2
	if math.Abs(x-centerX) > limit || math.Abs(z-centerZ) > limit {
		return
	}
	startX, startZ := result.X*16, result.Z*16
	minX := max(int32(math.Floor(x-radius))-startX-1, 0)
	maxX := min(int32(math.Floor(x+radius))-startX, 15)
	minZ := max(int32(math.Floor(z-radius))-startZ-1, 0)
	maxZ := min(int32(math.Floor(z+radius))-startZ, 15)
	minY := max(int32(math.Floor(y-verticalRadius))-1, result.MinY+1)
	maxY := min(int32(math.Floor(y+verticalRadius))+1, result.MinY+result.Height-8)
	for localX := minX; localX <= maxX; localX++ {
		blockX := startX + localX
		normalX := (float64(blockX) + 0.5 - x) / radius
		for localZ := minZ; localZ <= maxZ; localZ++ {
			blockZ := startZ + localZ
			normalZ := (float64(blockZ) + 0.5 - z) / radius
			if normalX*normalX+normalZ*normalZ >= 1 {
				continue
			}
			wasSurface := false
			for blockY := maxY; blockY > minY; blockY-- {
				normalY := (float64(blockY) - 0.5 - y) / verticalRadius
				if skip != nil && skip(normalX, normalY, normalZ, blockY) || skip == nil && (normalY <= float64(floor) || normalX*normalX+normalY*normalY+normalZ*normalZ >= 1) {
					continue
				}
				index := int((blockY-result.MinY)*256 + localZ*16 + localX)
				if result.CarvingMask[index] {
					continue
				}
				result.CarvingMask[index] = true
				result.CarvingMaskByCarver[pass.carverIndex]++
				state := result.surfaceState(localX, blockY, localZ)
				if state.Name == "minecraft:grass_block" || state.Name == "minecraft:mycelium" {
					wasSurface = true
				}
				if !carverReplaceable[state.Name] {
					continue
				}
				if blockY <= result.MinY+8 {
					result.setSurfaceState(localX, blockY, localZ, SurfaceBlockState{Name: "minecraft:lava"})
					continue
				}
				material, solid, update := pass.aquifer.compute(blockX, blockY, blockZ, 0)
				if pass.aquifer.err != nil {
					pass.err = pass.aquifer.err
					return
				}
				if solid {
					continue
				}
				result.setSurfaceState(localX, blockY, localZ, SurfaceBlockState{Name: material.BlockName()})
				result.FluidUpdates[index] = update && material != NoiseAir
				if wasSurface && blockY > result.MinY && result.surfaceState(localX, blockY-1, localZ).Name == "minecraft:dirt" {
					pass.restoreCarvedSurface(localX, blockY-1, localZ, material != NoiseAir)
					if pass.err != nil {
						return
					}
				}
			}
		}
	}
}

func (pass *carvingPass) restoreCarvedSurface(x, y, z int32, fluid bool) {
	result := pass.result
	blockX, blockZ := result.X*16+x, result.Z*16+z
	depth, err := pass.rules.SurfaceDepth(blockX, blockZ)
	if err != nil {
		pass.err = err
		return
	}
	secondary, err := pass.rules.SecondaryDepth(blockX, blockZ)
	if err != nil {
		pass.err = err
		return
	}
	cornerX, cornerZ := blockX&^15, blockZ&^15
	corners := [4]int32{
		pass.aquifer.preliminarySurface(cornerX, cornerZ),
		pass.aquifer.preliminarySurface(cornerX+16, cornerZ),
		pass.aquifer.preliminarySurface(cornerX, cornerZ+16),
		pass.aquifer.preliminarySurface(cornerX+16, cornerZ+16),
	}
	if pass.aquifer.err != nil {
		pass.err = pass.aquifer.err
		return
	}
	fractionX, fractionZ := float64(blockX&15)/16, float64(blockZ&15)/16
	lerp := func(fraction, start, end float64) float64 { return start + fraction*(end-start) }
	level := int32(math.Floor(lerp(fractionZ,
		lerp(fractionX, float64(corners[0]), float64(corners[1])),
		lerp(fractionX, float64(corners[2]), float64(corners[3]))))) + depth - 8
	lowZ, highZ := max(z-1, 0), min(z+1, 15)
	lowX, highX := max(x-1, 0), min(x+1, 15)
	steep := result.columnHeight(x, highZ) >= result.columnHeight(x, lowZ)+4 ||
		result.columnHeight(lowX, z) >= result.columnHeight(highX, z)+4
	qx, qy, qz := BiomeQuartAtBlock(pass.seed, blockX, y, blockZ)
	biome, err := pass.biome(qx, qy, qz)
	if err != nil {
		pass.err = err
		return
	}
	cold, err := BiomeColdAt(biome, blockX, y, blockZ)
	if err != nil {
		pass.err = err
		return
	}
	waterHeight := int32(math.MinInt32)
	if fluid {
		waterHeight = y + 1
	}
	state, err := pass.rules.Apply(SurfaceContext{
		X: blockX, Y: y, Z: blockZ, Biome: biome,
		SurfaceDepth: depth, SecondaryDepth: secondary,
		StoneDepthAbove: 1, StoneDepthBelow: 1, WaterHeight: waterHeight,
		MinSurfaceLevel: level, Cold: cold, Steep: steep,
	})
	if err != nil {
		pass.err = err
		return
	}
	if state != nil {
		result.setSurfaceState(x, y, z, *state)
	}
}

func (pass *carvingPass) carveCanyon(random *LegacyRandom, sourceX, sourceZ int32) {
	x := float64(sourceX*16 + random.NextIntBound(16))
	y := float64(10 + random.NextIntBound(58))
	z := float64(sourceZ*16 + random.NextIntBound(16))
	heading := random.NextFloat() * (float32(math.Pi) * 2)
	pitch := carverUniform(random, -0.125, 0.125)
	verticalScale := float64(float32(3))
	width := random.NextFloat()*4 + random.NextFloat()*2
	length := int32(float32(112) * carverUniform(random, 0.75, 1))
	pass.carveCanyonTunnel(NewLegacyRandom(int64(random.NextLong())), x, y, z, width, heading, pitch, length, verticalScale)
}

func (pass *carvingPass) carveCanyonTunnel(random *LegacyRandom, x, y, z float64, width, heading, pitch float32,
	length int32, verticalScale float64) {
	widths := make([]float32, pass.result.Height)
	lastWidth := float32(1)
	for index := range widths {
		if index == 0 || random.NextIntBound(3) == 0 {
			lastWidth = 1 + random.NextFloat()*random.NextFloat()
		}
		widths[index] = lastWidth * lastWidth
	}
	var headingVelocity, pitchVelocity float32
	for step := int32(0); step < length; step++ {
		horizontalRadius := 1.5 + float64(carverSin(float32(step)*float32(math.Pi)/float32(length))*width)
		verticalRadius := horizontalRadius * verticalScale
		horizontalRadius *= float64(carverUniform(random, 0.75, 1))
		verticalRadius *= float64(carverUniform(random, 0.75, 1))
		cosPitch := carverCos(pitch)
		x += float64(carverCos(heading) * cosPitch)
		y += float64(carverSin(pitch))
		z += float64(carverSin(heading) * cosPitch)
		pitch *= 0.7
		pitch += pitchVelocity * 0.05
		heading += headingVelocity * 0.05
		pitchVelocity *= 0.8
		headingVelocity *= 0.5
		pitchVelocity += (random.NextFloat() - random.NextFloat()) * random.NextFloat() * 2
		headingVelocity += (random.NextFloat() - random.NextFloat()) * random.NextFloat() * 4
		if random.NextIntBound(4) == 0 {
			continue
		}
		if !pass.pathNearTarget(x, z, step, length, width) {
			return
		}
		pass.carveEllipsoid(x, y, z, horizontalRadius, verticalRadius, 0,
			func(normalX, normalY, normalZ float64, blockY int32) bool {
				factor := float64(widths[blockY-pass.result.MinY-1])
				return (normalX*normalX+normalZ*normalZ)*factor+normalY*normalY/6 >= 1
			})
	}
}
