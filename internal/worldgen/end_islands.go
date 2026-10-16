package worldgen

import "math"

// EndIslandDensity reproduces EndIslandDensityFunction, including Java's
// integer overflow, truncating negative division and float intermediates.
// It is immutable after construction and safe to sample concurrently.
type EndIslandDensity struct {
	noise *simplexNoise
}

func NewEndIslandDensity(seed int64) *EndIslandDensity {
	random := NewLegacyRandom(seed)
	for i := 0; i < 17292; i++ {
		random.NextInt()
	}
	return &EndIslandDensity{noise: newSimplexNoise(random)}
}

func (n *EndIslandDensity) height(x, z int32) float32 {
	cx, cz := x/2, z/2
	lx, lz := x%2, z%2
	sqrtFloat := func(v float32) float32 { return float32(math.Sqrt(float64(v))) }
	clamp := func(v float32) float32 { return max(float32(-100), min(float32(80), v)) }
	value := clamp(100 - sqrtFloat(float32(x*x+z*z))*8)
	for dx := int32(-12); dx <= 12; dx++ {
		for dz := int32(-12); dz <= 12; dz++ {
			ix, iz := int64(cx+dx), int64(cz+dz)
			if ix*ix+iz*iz <= 4096 || n.noise.sample(float64(ix), float64(iz)) >= float64(float32(-0.9)) {
				continue
			}
			scale := float32(math.Mod(float64(float32(absLong(ix))*3439+float32(absLong(iz))*147), 13)) + 9
			rx, rz := float32(lx-dx*2), float32(lz-dz*2)
			candidate := clamp(100 - sqrtFloat(rx*rx+rz*rz)*scale)
			value = max(value, candidate)
		}
	}
	return value
}

func absLong(v int64) int64 {
	if v < 0 { return -v }
	return v
}

func (n *EndIslandDensity) Sample(x, z int32) float64 {
	v := (float64(n.height(x/8, z/8)) - 8) / 128
	if math.IsNaN(v) {
		// HotSpot's sqrt of a negative overflowed radius produces the negative
		// canonical NaN. Keep its bits for the original-JAR boundary vectors.
		return math.Float64frombits(0xfff8000000000000)
	}
	return v
}
