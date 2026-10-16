package worldgen

import (
	"atlas/internal/blockrules"
	"math"
	"sort"
)

// PortalCollisionSearchBounds is the AABB searched by PortalShape's
// findCollisionFreePosition and CollisionGetter.findFreePosition.
func PortalCollisionSearchBounds(position [3]float64, width, height float32) blockrules.Box {
	w, h := float64(width), float64(height)
	centerY := position[1] + h/2
	return blockrules.Box{position[0] - w/2 - 1e-6 - w, centerY - 1e-6 - h, position[2] - w/2 - 1e-6 - w, position[0] + w/2 + 1e-6 + w, centerY + 1 + 1e-6 + h, position[2] + w/2 + 1e-6 + w}
}

// FindCollisionFreePortalPosition subtracts the expanded block collision
// union from PortalShape's search shape and chooses its closest remaining
// point. Equal-distance regions retain VoxelShape's Y/X/Z traversal order.
func FindCollisionFreePortalPosition(position [3]float64, width, height float32, boxes []blockrules.Box) [3]float64 {
	if width > 4 || height > 4 {
		return position
	}
	w, h := float64(width), float64(height)
	center := [3]float64{position[0], position[1] + h/2, position[2]}
	search := blockrules.Box{center[0] - w/2 - 1e-6, center[1] - 1e-6, center[2] - w/2 - 1e-6, center[0] + w/2 + 1e-6, center[1] + 1 + 1e-6, center[2] + w/2 + 1e-6}
	coordinates := [3][]float64{}
	for axis := 0; axis < 3; axis++ {
		coordinates[axis] = []float64{search[axis], search[axis+3]}
	}
	obstacles := make([]blockrules.Box, 0, len(boxes))
	for _, box := range boxes {
		for axis, size := range []float64{w, h, w} {
			box[axis] -= size / 2
			box[axis+3] += size / 2
		}
		if box[0] >= search[3] || box[3] <= search[0] || box[1] >= search[4] || box[4] <= search[1] || box[2] >= search[5] || box[5] <= search[2] {
			continue
		}
		obstacles = append(obstacles, box)
		for axis := 0; axis < 3; axis++ {
			for _, v := range []float64{box[axis], box[axis+3]} {
				if v > search[axis] && v < search[axis+3] {
					coordinates[axis] = append(coordinates[axis], v)
				}
			}
		}
	}
	for axis, values := range coordinates {
		sort.Float64s(values)
		unique := values[:0]
		for _, v := range values {
			if len(unique) == 0 || v-unique[len(unique)-1] > 1e-7 {
				unique = append(unique, v)
			}
		}
		coordinates[axis] = unique
	}
	best := position
	bestDistance := math.Inf(1)
	for y := 0; y < len(coordinates[1])-1; y++ {
		for x := 0; x < len(coordinates[0])-1; x++ {
			for z := 0; z < len(coordinates[2])-1; z++ {
				indices := [3]int{x, y, z}
				var cell blockrules.Box
				var middle [3]float64
				for axis, i := range indices {
					cell[axis], cell[axis+3] = coordinates[axis][i], coordinates[axis][i+1]
					middle[axis] = (cell[axis] + cell[axis+3]) / 2
				}
				occupied := false
				for _, box := range obstacles {
					if middle[0] > box[0] && middle[0] < box[3] && middle[1] > box[1] && middle[1] < box[4] && middle[2] > box[2] && middle[2] < box[5] {
						occupied = true
						break
					}
				}
				if occupied {
					continue
				}
				var candidate [3]float64
				distance := 0.0
				for axis, v := range center {
					candidate[axis] = javaClamp(v, cell[axis], cell[axis+3])
					d := candidate[axis] - v
					distance += d * d
				}
				if distance < bestDistance {
					bestDistance = distance
					best = candidate
					best[1] -= h / 2
				}
			}
		}
	}
	return best
}
