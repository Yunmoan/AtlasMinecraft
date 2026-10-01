package worldgen

import (
	"math"
	"sort"
	"sync"
)

// Immutable bounding-volume index over climate intervals. Search order is an
// implementation detail: equal distances retain the original point order, or
// the caller's previous point when it is also a nearest neighbor.
type biomeSearchNode struct {
	ranges         [6][2]int64
	offsetDistance int64
	first          int
	left, right    *biomeSearchNode
	indices        []int
}

var biomeIndexOnce sync.Once
var biomeIndex *biomeSearchNode
var biomeIndexError error

func overworldBiomeIndex() (*biomeSearchNode, error) {
	biomeIndexOnce.Do(func() {
		points, err := OverworldBiomePoints()
		if err != nil {
			biomeIndexError = err
			return
		}
		indices := make([]int, len(points))
		for i := range indices {
			indices[i] = i
		}
		biomeIndex = buildBiomeIndex(points, indices)
	})
	return biomeIndex, biomeIndexError
}

func buildBiomeIndex(points []BiomeParameterPoint, indices []int) *biomeSearchNode {
	node := &biomeSearchNode{first: math.MaxInt, offsetDistance: math.MaxInt64}
	var centers [6][2]int64
	for axis := range node.ranges {
		node.ranges[axis] = [2]int64{math.MaxInt64, math.MinInt64}
		centers[axis] = [2]int64{math.MaxInt64, math.MinInt64}
	}
	for _, i := range indices {
		point := points[i]
		node.first = min(node.first, i)
		node.offsetDistance = min(node.offsetDistance, point.Offset*point.Offset)
		for axis, bounds := range point.Ranges {
			node.ranges[axis][0] = min(node.ranges[axis][0], bounds[0])
			node.ranges[axis][1] = max(node.ranges[axis][1], bounds[1])
			center := bounds[0] + bounds[1]
			centers[axis][0] = min(centers[axis][0], center)
			centers[axis][1] = max(centers[axis][1], center)
		}
	}
	if len(indices) <= 8 {
		node.indices = indices
		return node
	}
	axis := 0
	for i := 1; i < 6; i++ {
		if centers[i][1]-centers[i][0] > centers[axis][1]-centers[axis][0] {
			axis = i
		}
	}
	sort.Slice(indices, func(i, j int) bool {
		a, b := points[indices[i]].Ranges[axis], points[indices[j]].Ranges[axis]
		if a[0]+a[1] == b[0]+b[1] {
			return indices[i] < indices[j]
		}
		return a[0]+a[1] < b[0]+b[1]
	})
	middle := len(indices) / 2
	node.left = buildBiomeIndex(points, indices[:middle])
	node.right = buildBiomeIndex(points, indices[middle:])
	return node
}

func (node *biomeSearchNode) lowerBound(climate [6]int64) int64 {
	distance := node.offsetDistance
	for axis, bounds := range node.ranges {
		delta := int64(0)
		if climate[axis] < bounds[0] {
			delta = bounds[0] - climate[axis]
		}
		if climate[axis] > bounds[1] {
			delta = climate[axis] - bounds[1]
		}
		distance += delta * delta
	}
	return distance
}

func (node *biomeSearchNode) nearest(points []BiomeParameterPoint, climate [6]int64, previous int) int {
	best, index := int64(math.MaxInt64), len(points)
	if previous >= 0 && previous < len(points) {
		best, index = biomeDistance(points[previous], climate), previous
	}
	node.search(points, climate, previous, &best, &index, node.lowerBound(climate))
	return index
}

func (node *biomeSearchNode) search(points []BiomeParameterPoint, climate [6]int64, previous int, best *int64, index *int, bound int64) {
	if bound > *best || (bound == *best && (*index == previous || node.first >= *index)) {
		return
	}
	if node.left == nil {
		for _, i := range node.indices {
			distance := biomeDistance(points[i], climate)
			if distance < *best || (distance == *best && *index != previous && i < *index) {
				*best, *index = distance, i
			}
		}
		return
	}
	left, right := node.left, node.right
	leftBound, rightBound := left.lowerBound(climate), right.lowerBound(climate)
	if rightBound < leftBound || (rightBound == leftBound && right.first < left.first) {
		left, right, leftBound, rightBound = right, left, rightBound, leftBound
	}
	left.search(points, climate, previous, best, index, leftBound)
	right.search(points, climate, previous, best, index, rightBound)
}
