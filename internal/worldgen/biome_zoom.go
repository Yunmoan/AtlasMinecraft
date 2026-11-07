package worldgen

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
)

func BiomeQuartAtBlock(seed int64, x, y, z int32) (int32, int32, int32) {
	zoom := biomeZoom{seed: biomeZoomSeed(seed)}
	return zoom.quartAtBlock(x, y, z)
}

type biomeZoom struct {
	seed                     uint64
	minX, minY, minZ, height int32
	offsets                  [][3]float64
	valid                    []bool
}

func biomeZoomSeed(seed int64) uint64 {
	var seedBytes [8]byte
	binary.LittleEndian.PutUint64(seedBytes[:], uint64(seed))
	hash := sha256.Sum256(seedBytes[:])
	return binary.LittleEndian.Uint64(hash[:8])
}

// The surface pass revisits the same quart corners at almost every block.
// Cache only deterministic offsets, keeping distance arithmetic and strict
// corner tie order identical. The window includes both adjacent quart halos.
func newChunkBiomeZoom(seed int64, cx, cz, minY, height int32) biomeZoom {
	zoom := biomeZoom{seed: biomeZoomSeed(seed), minX: cx*4 - 1, minZ: cz*4 - 1, minY: (minY - 2) >> 2, height: height/4 + 2}
	if height > 0 && height <= 4096 && minY%16 == 0 && height%16 == 0 {
		zoom.offsets = make([][3]float64, 36*int(zoom.height))
		zoom.valid = make([]bool, len(zoom.offsets))
	}
	return zoom
}

func (zoom *biomeZoom) offset(x, y, z int32) [3]float64 {
	if len(zoom.offsets) != 0 {
		dx, dy, dz := x-zoom.minX, y-zoom.minY, z-zoom.minZ
		if dx >= 0 && dx < 6 && dz >= 0 && dz < 6 && dy >= 0 && dy < zoom.height {
			i := int((dz*6+dx)*zoom.height + dy)
			if !zoom.valid[i] {
				zoom.offsets[i] = biomeFiddledOffset(zoom.seed, x, y, z)
				zoom.valid[i] = true
			}
			return zoom.offsets[i]
		}
	}
	return biomeFiddledOffset(zoom.seed, x, y, z)
}

func (zoom *biomeZoom) quartAtBlock(x, y, z int32) (int32, int32, int32) {
	blockX, blockY, blockZ := x-2, y-2, z-2
	quartX, quartY, quartZ := blockX>>2, blockY>>2, blockZ>>2
	fracX, fracY, fracZ := float64(blockX&3)/4, float64(blockY&3)/4, float64(blockZ&3)/4
	bestDistance := math.Inf(1)
	selected := 0
	for corner := 0; corner < 8; corner++ {
		qx, qy, qz := quartX, quartY, quartZ
		dx, dy, dz := fracX, fracY, fracZ
		if corner&4 != 0 {
			qx++
			dx--
		}
		if corner&2 != 0 {
			qy++
			dy--
		}
		if corner&1 != 0 {
			qz++
			dz--
		}
		offset := zoom.offset(qx, qy, qz)
		dx += offset[0]
		dy += offset[1]
		dz += offset[2]
		distance := dx*dx + dy*dy + dz*dz
		if distance < bestDistance {
			bestDistance, selected = distance, corner
		}
	}
	if selected&4 != 0 {
		quartX++
	}
	if selected&2 != 0 {
		quartY++
	}
	if selected&1 != 0 {
		quartZ++
	}
	return quartX, quartY, quartZ
}

func biomeFiddledDistance(seed uint64, x, y, z int32, dx, dy, dz float64) float64 {
	offset := biomeFiddledOffset(seed, x, y, z)
	dx += offset[0]
	dy += offset[1]
	dz += offset[2]
	return dx*dx + dy*dy + dz*dz
}

func biomeFiddledOffset(seed uint64, x, y, z int32) [3]float64 {
	mix := func(value uint64, coordinate int64) uint64 {
		return value*(value*6364136223846793005+1442695040888963407) + uint64(coordinate)
	}
	hash := seed
	for range 2 {
		hash = mix(hash, int64(x))
		hash = mix(hash, int64(y))
		hash = mix(hash, int64(z))
	}
	offset := func(value uint64) float64 {
		remainder := (int64(value) >> 24) % 1024
		if remainder < 0 {
			remainder += 1024
		}
		return (float64(remainder)/1024 - 0.5) * 0.9
	}
	offsetX := offset(hash)
	hash = mix(hash, int64(seed))
	offsetY := offset(hash)
	hash = mix(hash, int64(seed))
	offsetZ := offset(hash)
	return [3]float64{offsetX, offsetY, offsetZ}
}
