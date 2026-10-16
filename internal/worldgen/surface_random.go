package worldgen

type surfaceRandom interface {
	NextDouble() float64
	NextFloat() float32
	NextIntBound(int32) int32
	NextBoolean() bool
}

func (s *SurfaceRuleSampler) randomAt(x,y,z int32) surfaceRandom {
	if s.legacyRandom!=nil { return s.legacyRandom.At(x,y,z) }
	return s.random.At(x,y,z)
}

func (s *SurfaceRuleSampler) hashRandom(name string) surfaceRandom {
	if s.legacyRandom!=nil { return s.legacyRandom.FromHashOf(name) }
	return s.random.FromHashOf(name)
}
