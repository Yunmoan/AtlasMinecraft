package worldgen

import "math"

// WorldgenRandom is Minecraft's WorldgenRandom wrapping an Xoroshiro source.
// Its LegacyRandomSource-style methods consume high bits of successive
// Xoroshiro longs rather than exposing the source's methods directly.
type WorldgenRandom struct {
	source      *Xoroshiro
	gaussian    float64
	hasGaussian bool
}

// XoroshiroState is a read-only snapshot for independent feature randomness.
func (random *WorldgenRandom) XoroshiroState() (lo,hi uint64) {
	return random.source.lo,random.source.hi
}

// SetSeed matches WorldgenRandom.setSeed: its inherited Gaussian cache survives
// feature reseeding, unlike the cache in the wrapped Xoroshiro source.
func (random *WorldgenRandom) SetSeed(seed int64) { random.source = NewXoroshiro(seed) }

func (random *WorldgenRandom) NextGaussian() float64 {
	if random.hasGaussian {
		random.hasGaussian = false
		return random.gaussian
	}
	for {
		x, y := 2*random.NextDouble()-1, 2*random.NextDouble()-1
		s := x*x + y*y
		if s >= 1 || s == 0 {
			continue
		}
		factor := math.Sqrt(-2 * math.Log(s) / s)
		random.gaussian, random.hasGaussian = y*factor, true
		return x * factor
	}
}

func NewWorldgenRandom(seed int64) *WorldgenRandom {
	return &WorldgenRandom{source: NewXoroshiro(seed)}
}

func (random *WorldgenRandom) NextBits(count uint) int32 {
	if count < 1 || count > 32 {
		panic("worldgen: worldgen random bit count outside 1..32")
	}
	return int32(random.source.NextLong() >> (64 - count))
}

func (random *WorldgenRandom) NextInt() int32 { return random.NextBits(32) }

func (random *WorldgenRandom) NextIntBound(bound int32) int32 {
	if bound <= 0 {
		panic("worldgen: random bound must be positive")
	}
	if bound&(bound-1) == 0 {
		return int32((int64(bound) * int64(random.NextBits(31))) >> 31)
	}
	for {
		bits := random.NextBits(31)
		value := bits % bound
		if bits-value+(bound-1) >= 0 {
			return value
		}
	}
}

func (random *WorldgenRandom) NextLong() uint64 {
	high, low := int64(random.NextBits(32)), int64(random.NextBits(32))
	return uint64((high << 32) + low)
}

func (random *WorldgenRandom) NextBoolean() bool { return random.NextBits(1) != 0 }

func (random *WorldgenRandom) NextFloat() float32 {
	return float32(random.NextBits(24)) * (1.0 / (1 << 24))
}

func (random *WorldgenRandom) NextDouble() float64 {
	high, low := int64(random.NextBits(26)), int64(random.NextBits(27))
	return float64((high<<27)+low) * (1.0 / (1 << 53))
}
