package worldgen

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
)

func BiomeQuartAtBlock(seed int64, x, y, z int32) (int32, int32, int32) {
	var seedBytes [8]byte
	binary.LittleEndian.PutUint64(seedBytes[:], uint64(seed))
	hash := sha256.Sum256(seedBytes[:])
	zoomSeed := binary.LittleEndian.Uint64(hash[:8])
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
		distance := biomeFiddledDistance(zoomSeed, qx, qy, qz, dx, dy, dz)
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
	dx += offsetX
	dy += offsetY
	dz += offsetZ
	return dx*dx + dy*dy + dz*dz
}
