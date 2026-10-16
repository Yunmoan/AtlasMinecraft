package worldgen

import (
	"fmt"
	"math"
)

var netherCarverReplaceable = map[string]bool{
	"minecraft:stone": true, "minecraft:granite": true, "minecraft:diorite": true,
	"minecraft:andesite": true, "minecraft:deepslate": true, "minecraft:tuff": true,
	"minecraft:netherrack": true, "minecraft:basalt": true, "minecraft:blackstone": true,
	"minecraft:dirt": true, "minecraft:grass_block": true, "minecraft:podzol": true,
	"minecraft:coarse_dirt": true, "minecraft:mycelium": true, "minecraft:rooted_dirt": true,
	"minecraft:moss_block": true, "minecraft:mud": true, "minecraft:muddy_mangrove_roots": true,
	"minecraft:crimson_nylium": true, "minecraft:warped_nylium": true,
	"minecraft:nether_wart_block": true, "minecraft:warped_wart_block": true,
	"minecraft:soul_sand": true, "minecraft:soul_soil": true,
}

type netherCarvingPass struct{ result *NoiseChunkResult }

// ApplyNetherCarvers ports NetherWorldCarver's cave geometry and replacement
// rules. It uses the 128-block generation range, not the 256-block build range.
func (g *DimensionNoiseGenerator) ApplyNetherCarvers(result *NoiseChunkResult) error {
	if g.settings.Name != "minecraft:nether" || result == nil || result.MinY != g.settings.MinY || result.Height != g.settings.Height || len(result.Blocks) != int(result.Height)*256 {
		return fmt.Errorf("Nether carvers require a complete matching Nether surface chunk")
	}
	result.CarvingMask = make([]bool, len(result.Blocks))
	result.CarvingMaskByCarver = [3]int{}
	pass := netherCarvingPass{result}
	for dx := int32(-8); dx <= 8; dx++ {
		for dz := int32(-8); dz <= 8; dz++ {
			sx, sz := result.X+dx, result.Z+dz
			random := carverRandom(g.seed, sx, sz, 0)
			if random.NextFloat() <= 0.2 {
				pass.carveCaves(random, sx, sz)
			}
		}
	}
	return nil
}

func (p *netherCarvingPass) carveCaves(random *LegacyRandom, sx, sz int32) {
	count := random.NextIntBound(random.NextIntBound(random.NextIntBound(10)+1) + 1)
	for tunnel := int32(0); tunnel < count; tunnel++ {
		x := float64(sx*16 + random.NextIntBound(16))
		y := float64(p.result.MinY + random.NextIntBound(p.result.Height-1))
		z := float64(sz*16 + random.NextIntBound(16))
		branches := int32(1)
		if random.NextIntBound(4) == 0 {
			radius := 1 + random.NextFloat()*6
			width := 1.5 + float64(carverSin(float32(1.5707964))*radius)
			p.carveEllipsoid(x+1, y, z, width, width*0.5)
			branches += random.NextIntBound(4)
		}
		for branch := int32(0); branch < branches; branch++ {
			heading := random.NextFloat() * (float32(math.Pi) * 2)
			pitch := (random.NextFloat() - 0.5) / 4
			width := (random.NextFloat()*2 + random.NextFloat()) * 2
			length := int32(112) - random.NextIntBound(28)
			p.carveTunnel(NewLegacyRandom(int64(random.NextLong())), x, y, z, width, heading, pitch, 0, length, 5)
		}
	}
}

func (p *netherCarvingPass) carveTunnel(random *LegacyRandom, x, y, z float64, width, heading, pitch float32, step, length int32, verticalScale float64) {
	split := random.NextIntBound(length/2) + length/4
	gentle := random.NextIntBound(6) == 0
	var headingVelocity, pitchVelocity float32
	for ; step < length; step++ {
		radius := 1.5 + float64(carverSin(float32(math.Pi)*float32(step)/float32(length))*width)
		verticalRadius := radius * verticalScale
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
			p.carveTunnel(NewLegacyRandom(int64(firstSeed)), x, y, z, branchWidth, heading-float32(1.5707964), pitch/3, step, length, 1)
			secondSeed := random.NextLong()
			branchWidth = random.NextFloat()*0.5 + 0.5
			p.carveTunnel(NewLegacyRandom(int64(secondSeed)), x, y, z, branchWidth, heading+float32(1.5707964), pitch/3, step, length, 1)
			return
		}
		if random.NextIntBound(4) == 0 {
			continue
		}
		dx, dz := x-float64(p.result.X*16+8), z-float64(p.result.Z*16+8)
		remaining, limit := float64(length-step), float64(float32(width+2)+16)
		if dx*dx+dz*dz-remaining*remaining > limit*limit {
			return
		}
		p.carveEllipsoid(x, y, z, radius, verticalRadius)
	}
}

func (p *netherCarvingPass) carveEllipsoid(x, y, z, radius, verticalRadius float64) {
	r := p.result
	limit := 16 + radius*2
	if math.Abs(x-float64(r.X*16+8)) > limit || math.Abs(z-float64(r.Z*16+8)) > limit {
		return
	}
	sx, sz := r.X*16, r.Z*16
	minX, maxX := max(int32(math.Floor(x-radius))-sx-1, 0), min(int32(math.Floor(x+radius))-sx, 15)
	minZ, maxZ := max(int32(math.Floor(z-radius))-sz-1, 0), min(int32(math.Floor(z+radius))-sz, 15)
	minY, maxY := max(int32(math.Floor(y-verticalRadius))-1, r.MinY+1), min(int32(math.Floor(y+verticalRadius))+1, r.MinY+r.Height-8)
	for localX := minX; localX <= maxX; localX++ {
		nx := (float64(sx+localX) + 0.5 - x) / radius
		for localZ := minZ; localZ <= maxZ; localZ++ {
			nz := (float64(sz+localZ) + 0.5 - z) / radius
			if nx*nx+nz*nz >= 1 {
				continue
			}
			for by := maxY; by > minY; by-- {
				ny := (float64(by) - 0.5 - y) / verticalRadius
				if ny <= float64(float32(-0.7)) || nx*nx+ny*ny+nz*nz >= 1 {
					continue
				}
				index := int((by-r.MinY)*256 + localZ*16 + localX)
				if r.CarvingMask[index] {
					continue
				}
				r.CarvingMask[index] = true
				r.CarvingMaskByCarver[0]++
				if !netherCarverReplaceable[r.surfaceState(localX, by, localZ).Name] {
					continue
				}
				state := SurfaceBlockState{Name: "minecraft:cave_air"}
				if by <= r.MinY+31 {
					state.Name = "minecraft:lava"
				}
				r.setSurfaceState(localX, by, localZ, state)
			}
		}
	}
}
