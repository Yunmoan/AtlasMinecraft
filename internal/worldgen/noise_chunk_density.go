package worldgen

import "math/bits"

type densityPosition struct {
	x, y, z int32
}

type densityCacheKey struct {
	node     *DensityNode
	position densityPosition
}

type chunkDensityContext struct {
	minX, minZ                  int32
	fillingCell                 bool
	singlePoint                 bool
	flat                        map[densityCacheKey]float64
	corners                     map[densityCacheKey]float64
	twoD                        map[densityCacheKey]float64
	interpolated                map[*DensityNode]*interpolatedDensityCell
	denseCorners                map[*DensityNode]*densityCornerValues
	minY, height                int32
	cornerWidth, cornerHeight   int
	cornerShiftXZ, cornerShiftY uint
}

type densityCornerValues struct {
	values []float64
	valid  []bool
}

type interpolatedDensityCell struct {
	position densityPosition
	corners  [2][2][2]float64
}

// Reuse only density values, never the selected biome: nearest-biome ties
// must still observe every call in the original traversal order. Bound this
// private cache because spawn searches can visit arbitrarily many columns.
func (s *DensitySampler) sampleHorizontal(node *DensityNode, x, y, z int32) (float64, error) {
	key := densityCacheKey{node: node, position: densityPosition{x: x, z: z}}
	if value, ok := s.horizontal[key]; ok {
		return value, nil
	}
	value, err := s.Sample(node.argument(), x, y, z)
	if err == nil {
		if len(s.horizontal) >= 4096 {
			clear(s.horizontal)
		}
		s.horizontal[key] = value
	}
	return value, err
}

func NewNoiseChunkDensitySampler(seed int64, settings *DensitySettings, cx, cz int32) *DensitySampler {
	sampler := NewCellDensitySampler(seed, settings)
	sampler.chunk = &chunkDensityContext{
		minX: cx * 16, minZ: cz * 16,
		flat:         make(map[densityCacheKey]float64),
		corners:      make(map[densityCacheKey]float64),
		twoD:         make(map[densityCacheKey]float64),
		interpolated: make(map[*DensityNode]*interpolatedDensityCell),
		minY:         settings.MinY, height: settings.Height,
	}
	if settings.CellWidth > 0 && settings.CellHeight > 0 && settings.Height >= 0 &&
		settings.CellWidth&(settings.CellWidth-1) == 0 && settings.CellHeight&(settings.CellHeight-1) == 0 {
		width, height := int(16/settings.CellWidth)+1, int(settings.Height/settings.CellHeight)+1
		if int64(width)*int64(width)*int64(height) <= 65536 {
			sampler.chunk.denseCorners = make(map[*DensityNode]*densityCornerValues)
			sampler.chunk.cornerWidth, sampler.chunk.cornerHeight = width, height
			sampler.chunk.cornerShiftXZ = uint(bits.TrailingZeros32(uint32(settings.CellWidth)))
			sampler.chunk.cornerShiftY = uint(bits.TrailingZeros32(uint32(settings.CellHeight)))
		}
	}
	return sampler
}

func (s *DensitySampler) sampleFlatCache(node *DensityNode, x, y, z int32) (float64, error) {
	quartX, quartZ := floorCell(x, 4), floorCell(z, 4)
	if quartX < s.chunk.minX || quartX > s.chunk.minX+16 || quartZ < s.chunk.minZ || quartZ > s.chunk.minZ+16 {
		return s.Sample(node.argument(), x, y, z)
	}
	key := densityCacheKey{node: node, position: densityPosition{x: quartX, z: quartZ}}
	if value, exists := s.chunk.flat[key]; exists {
		return value, nil
	}
	previous := s.chunk.singlePoint
	s.chunk.singlePoint = true
	value, err := s.Sample(node.argument(), quartX, 0, quartZ)
	s.chunk.singlePoint = previous
	if err == nil {
		s.chunk.flat[key] = value
	}
	return value, err
}

func (s *DensitySampler) sampleCache2D(node *DensityNode, x, y, z int32) (float64, error) {
	key := densityCacheKey{node: node, position: densityPosition{x: x, z: z}}
	if value, exists := s.chunk.twoD[key]; exists {
		return value, nil
	}
	value, err := s.Sample(node.argument(), x, y, z)
	if err == nil {
		s.chunk.twoD[key] = value
	}
	return value, err
}

func (s *DensitySampler) sampleCorner(node *DensityNode, x, y, z int32) (float64, error) {
	if s.chunk == nil {
		return s.Sample(node, x, y, z)
	}
	// Interpolation corners inside this chunk form a small regular lattice.
	// Index it directly instead of hashing a node plus three coordinates and
	// repeatedly growing the much larger coordinate map. Out-of-range probes
	// keep the original map path, including their error/retry behavior.
	dx, dy, dz := x-s.chunk.minX, y-s.chunk.minY, z-s.chunk.minZ
	if s.chunk.denseCorners != nil &&
		dx >= 0 && dx <= 16 && dz >= 0 && dz <= 16 && dy >= 0 && dy <= s.chunk.height &&
		dx&(s.cellWidth-1) == 0 && dz&(s.cellWidth-1) == 0 && dy&(s.cellHeight-1) == 0 {
		width, height := s.chunk.cornerWidth, s.chunk.cornerHeight
		index := (int(dz>>s.chunk.cornerShiftXZ)*width+int(dx>>s.chunk.cornerShiftXZ))*height + int(dy>>s.chunk.cornerShiftY)
		cache := s.chunk.denseCorners[node]
		if cache == nil {
			cache = &densityCornerValues{values: make([]float64, width*width*height), valid: make([]bool, width*width*height)}
			s.chunk.denseCorners[node] = cache
		}
		if cache.valid[index] {
			return cache.values[index], nil
		}
		previous := s.chunk.fillingCell
		s.chunk.fillingCell = false
		value, err := s.Sample(node, x, y, z)
		s.chunk.fillingCell = previous
		if err == nil {
			cache.values[index], cache.valid[index] = value, true
		}
		return value, err
	}
	key := densityCacheKey{node: node, position: densityPosition{x: x, y: y, z: z}}
	if value, exists := s.chunk.corners[key]; exists {
		return value, nil
	}
	previous := s.chunk.fillingCell
	s.chunk.fillingCell = false
	value, err := s.Sample(node, x, y, z)
	s.chunk.fillingCell = previous
	if err == nil {
		s.chunk.corners[key] = value
	}
	return value, err
}

func (s *DensitySampler) sampleSinglePoint(node *DensityNode, x, y, z int32) (float64, error) {
	previous := s.chunk.singlePoint
	s.chunk.singlePoint = true
	value, err := s.Sample(node, x, y, z)
	s.chunk.singlePoint = previous
	return value, err
}
