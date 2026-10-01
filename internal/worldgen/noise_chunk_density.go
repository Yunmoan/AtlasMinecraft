package worldgen

type densityPosition struct {
	x, y, z int32
}

type densityCacheKey struct {
	node     *DensityNode
	position densityPosition
}

type chunkDensityContext struct {
	minX, minZ   int32
	fillingCell  bool
	singlePoint  bool
	flat         map[densityCacheKey]float64
	corners      map[densityCacheKey]float64
	twoD         map[densityCacheKey]float64
	interpolated map[*DensityNode]*interpolatedDensityCell
}

type interpolatedDensityCell struct {
	position densityPosition
	corners  [2][2][2]float64
}

func NewNoiseChunkDensitySampler(seed int64, settings *DensitySettings, cx, cz int32) *DensitySampler {
	sampler := NewCellDensitySampler(seed, settings)
	sampler.chunk = &chunkDensityContext{
		minX: cx * 16, minZ: cz * 16,
		flat:         make(map[densityCacheKey]float64),
		corners:      make(map[densityCacheKey]float64),
		twoD:         make(map[densityCacheKey]float64),
		interpolated: make(map[*DensityNode]*interpolatedDensityCell),
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
