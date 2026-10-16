package worldgen

import (
	"fmt"
	"math"
)

const aquiferDryLevel int32 = -32512

type aquiferFluid struct {
	level    int32
	material NoiseMaterial
}

func (fluid aquiferFluid) at(y int32) NoiseMaterial {
	if y < fluid.level {
		return fluid.material
	}
	return NoiseAir
}

type noiseAquifer struct {
	sampler    *DensitySampler
	settings   *DensitySettings
	random     PositionalRandom
	centers    map[densityPosition]densityPosition
	fluids     map[densityPosition]aquiferFluid
	surfaces   map[densityPosition]int32
	err        error
	gridMin    densityPosition
	gridHeight int32
	centerGrid []aquiferCenter
}

type aquiferCenter struct {
	position densityPosition
	valid    bool
}

func newNoiseAquifer(sampler *DensitySampler, settings *DensitySettings, random PositionalRandom) *noiseAquifer {
	aquifer := &noiseAquifer{
		sampler: sampler, settings: settings, random: random,
		centers:  make(map[densityPosition]densityPosition),
		fluids:   make(map[densityPosition]aquiferFluid),
		surfaces: make(map[densityPosition]int32),
	}
	if sampler.chunk != nil {
		// Only 3x3 aquifer grid columns can influence blocks in this chunk.
		// A compact array avoids twelve coordinate-map lookups per air/fluid
		// block. Out-of-grid probes still use the general coordinate cache.
		minY := floorCell(settings.MinY+1, 12)/12 - 1
		maxY := floorCell(settings.MinY+settings.Height, 12)/12 + 1
		aquifer.gridMin = densityPosition{sampler.chunk.minX/16 - 1, minY, sampler.chunk.minZ/16 - 1}
		aquifer.gridHeight = maxY - minY + 1
		aquifer.centerGrid = make([]aquiferCenter, 9*aquifer.gridHeight)
	}
	return aquifer
}

func (aquifer *noiseAquifer) centerAt(grid densityPosition) densityPosition {
	x, y, z := grid.x-aquifer.gridMin.x, grid.y-aquifer.gridMin.y, grid.z-aquifer.gridMin.z
	var cached *aquiferCenter
	if x >= 0 && x < 3 && z >= 0 && z < 3 && y >= 0 && y < aquifer.gridHeight {
		cached = &aquifer.centerGrid[(x*aquifer.gridHeight+y)*3+z]
		if cached.valid {
			return cached.position
		}
	} else if center, ok := aquifer.centers[grid]; ok {
		return center
	}
	positionRandom := aquifer.random.At(grid.x, grid.y, grid.z)
	center := densityPosition{
		x: grid.x*16 + positionRandom.NextIntBound(10),
		y: grid.y*12 + positionRandom.NextIntBound(9),
		z: grid.z*16 + positionRandom.NextIntBound(10),
	}
	if cached != nil {
		cached.position, cached.valid = center, true
	} else {
		aquifer.centers[grid] = center
	}
	return center
}

func (aquifer *noiseAquifer) sample(field string, x, y, z int32) float64 {
	if aquifer.err != nil {
		return 0
	}
	value, err := aquifer.sampler.Sample(aquifer.settings.Router[field], x, y, z)
	if err != nil {
		aquifer.err = fmt.Errorf("aquifer %s at %d,%d,%d: %w", field, x, y, z, err)
	}
	return value
}

func (aquifer *noiseAquifer) globalFluid(y int32) aquiferFluid {
	if y < min(int32(-54), aquifer.settings.SeaLevel) {
		return aquiferFluid{level: -54, material: NoiseLava}
	}
	material := NoiseWater
	switch aquifer.settings.DefaultFluid.Name {
	case "minecraft:lava": material = NoiseLava
	case "minecraft:air": material = NoiseAir
	}
	return aquiferFluid{level: aquifer.settings.SeaLevel, material: material}
}

func (aquifer *noiseAquifer) compute(x, y, z int32, density float64) (NoiseMaterial, bool, bool) {
	if density > 0 {
		return NoiseStone, true, false
	}
	global := aquifer.globalFluid(y)
	if !aquifer.settings.Aquifers {
		return global.at(y), false, false
	}
	if global.at(y) == NoiseLava {
		return NoiseLava, false, false
	}
	gridX, gridY, gridZ := floorCell(x-5, 16)/16, floorCell(y+1, 12)/12, floorCell(z-5, 16)/16
	distances := [3]int32{math.MaxInt32, math.MaxInt32, math.MaxInt32}
	var closest [3]densityPosition
	for offsetX := int32(0); offsetX <= 1; offsetX++ {
		for offsetY := int32(-1); offsetY <= 1; offsetY++ {
			for offsetZ := int32(0); offsetZ <= 1; offsetZ++ {
				grid := densityPosition{x: gridX + offsetX, y: gridY + offsetY, z: gridZ + offsetZ}
				center := aquifer.centerAt(grid)
				deltaX, deltaY, deltaZ := center.x-x, center.y-y, center.z-z
				distance := deltaX*deltaX + deltaY*deltaY + deltaZ*deltaZ
				if distances[0] >= distance {
					closest[2], closest[1], closest[0] = closest[1], closest[0], center
					distances[2], distances[1], distances[0] = distances[1], distances[0], distance
				} else if distances[1] >= distance {
					closest[2], closest[1] = closest[1], center
					distances[2], distances[1] = distances[1], distance
				} else if distances[2] >= distance {
					closest[2], distances[2] = center, distance
				}
			}
		}
	}
	first := aquifer.fluidAt(closest[0])
	material := first.at(y)
	similarity := aquiferSimilarity(distances[0], distances[1])
	if similarity <= 0 {
		return material, false, similarity >= aquiferSimilarity(100, 144)
	}
	if material == NoiseWater && aquifer.globalFluid(y-1).at(y-1) == NoiseLava {
		return material, false, true
	}
	barrier := math.NaN()
	second := aquifer.fluidAt(closest[1])
	if density+similarity*aquifer.pressure(x, y, z, &barrier, first, second) > 0 {
		return NoiseStone, true, false
	}
	third := aquifer.fluidAt(closest[2])
	firstThird := aquiferSimilarity(distances[0], distances[2])
	if firstThird > 0 && density+similarity*firstThird*aquifer.pressure(x, y, z, &barrier, first, third) > 0 {
		return NoiseStone, true, false
	}
	secondThird := aquiferSimilarity(distances[1], distances[2])
	if secondThird > 0 && density+similarity*secondThird*aquifer.pressure(x, y, z, &barrier, second, third) > 0 {
		return NoiseStone, true, false
	}
	return material, false, true
}

func aquiferSimilarity(first, second int32) float64 {
	return 1 - math.Abs(float64(second-first))/25
}

func (aquifer *noiseAquifer) pressure(x, y, z int32, barrier *float64, first, second aquiferFluid) float64 {
	materialA, materialB := first.at(y), second.at(y)
	if materialA == NoiseLava && materialB == NoiseWater || materialA == NoiseWater && materialB == NoiseLava {
		return 2
	}
	difference := math.Abs(float64(first.level - second.level))
	if difference == 0 {
		return 0
	}
	middle := 0.5 * float64(first.level+second.level)
	distance := float64(y) + 0.5 - middle
	available := difference/2 - math.Abs(distance)
	var shape float64
	if distance > 0 {
		if available > 0 {
			shape = available / 1.5
		} else {
			shape = available / 2.5
		}
	} else if 3+available > 0 {
		shape = (3 + available) / 3
	} else {
		shape = (3 + available) / 10
	}
	noise := 0.0
	if shape >= -2 && shape <= 2 {
		if math.IsNaN(*barrier) {
			*barrier = aquifer.sample("barrier", x, y, z)
		}
		noise = *barrier
	}
	return 2 * (noise + shape)
}

func (aquifer *noiseAquifer) preliminarySurface(x, z int32) int32 {
	x, z = floorCell(x, 4), floorCell(z, 4)
	key := densityPosition{x: x, z: z}
	if surface, exists := aquifer.surfaces[key]; exists {
		return surface
	}
	surface := int32(math.MaxInt32)
	for y := aquifer.settings.MinY + aquifer.settings.Height; y >= aquifer.settings.MinY; y -= aquifer.settings.CellHeight {
		value, err := aquifer.sampler.sampleSinglePoint(aquifer.settings.Router["initial_density_without_jaggedness"], x, y, z)
		if err != nil {
			aquifer.err = err
			break
		}
		if value > 0.390625 {
			surface = y
			break
		}
	}
	aquifer.surfaces[key] = surface
	return surface
}

func (aquifer *noiseAquifer) fluidAt(position densityPosition) aquiferFluid {
	if fluid, exists := aquifer.fluids[position]; exists {
		return fluid
	}
	fluid := aquifer.calculateFluid(position.x, position.y, position.z)
	aquifer.fluids[position] = fluid
	return fluid
}

func (aquifer *noiseAquifer) calculateFluid(x, y, z int32) aquiferFluid {
	global := aquifer.globalFluid(y)
	minSurface := int32(math.MaxInt32)
	nearSurface := false
	offsets := [...][2]int32{{0, 0}, {-2, -1}, {-1, -1}, {0, -1}, {1, -1}, {-3, 0}, {-2, 0}, {-1, 0}, {1, 0}, {-2, 1}, {-1, 1}, {0, 1}, {1, 1}}
	for _, offset := range offsets {
		surface := aquifer.preliminarySurface(x+offset[0]*16, z+offset[1]*16)
		level := surface + 8
		center := offset[0] == 0 && offset[1] == 0
		if center && y-12 > level {
			return global
		}
		above := y+12 > level
		if above || center {
			fluid := aquifer.globalFluid(level)
			if fluid.at(level) != NoiseAir {
				if center {
					nearSurface = true
				}
				if above {
					return fluid
				}
			}
		}
		minSurface = min(minSurface, surface)
	}
	level := aquifer.fluidLevel(x, y, z, global, minSurface, nearSurface)
	material := global.material
	if level <= -10 && level != aquiferDryLevel && material != NoiseLava {
		lava := aquifer.sample("lava", floorCell(x, 64)/64, floorCell(y, 40)/40, floorCell(z, 64)/64)
		if math.Abs(lava) > 0.3 {
			material = NoiseLava
		}
	}
	return aquiferFluid{level: level, material: material}
}

func (aquifer *noiseAquifer) fluidLevel(x, y, z int32, global aquiferFluid, surface int32, nearSurface bool) int32 {
	erosion := aquifer.sample("erosion", x, y, z)
	if erosion < float64(float32(-0.225)) && aquifer.sample("depth", x, y, z) > float64(float32(0.9)) {
		return aquiferDryLevel
	}
	factor := 0.0
	if nearSurface {
		factor = clampedMap(float64(surface+8-y), 0, 64, 1, 0)
	}
	floodedness := math.Max(-1, math.Min(1, aquifer.sample("fluid_level_floodedness", x, y, z)))
	high := -0.3 + (1-factor)*1.1
	low := -0.8 + (1-factor)*1.2
	if floodedness-high > 0 {
		return global.level
	}
	if floodedness-low <= 0 {
		return aquiferDryLevel
	}
	gridX, gridY, gridZ := floorCell(x, 16)/16, floorCell(y, 40)/40, floorCell(z, 16)/16
	spread := aquifer.sample("fluid_level_spread", gridX, gridY, gridZ) * 10
	quantized := int32(math.Floor(spread/3)) * 3
	return min(surface, gridY*40+20+quantized)
}
