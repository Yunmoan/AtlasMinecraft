package worldgen

import "unicode/utf16"

// LegacyRandom is Minecraft 1.21.1's 48-bit LegacyRandomSource. It is used by
// generation paths that intentionally retain Java's original LCG behavior.
type LegacyRandom struct{ seed uint64 }

const legacyMask uint64 = (1 << 48) - 1

func NewLegacyRandom(seed int64) *LegacyRandom {
	return &LegacyRandom{seed: (uint64(seed) ^ 0x5deece66d) & legacyMask}
}

func (r *LegacyRandom) NextBits(count uint) int32 {
	if count < 1 || count > 32 {
		panic("worldgen: legacy bit count outside 1..32")
	}
	r.seed = (r.seed*0x5deece66d + 11) & legacyMask
	return int32(r.seed >> (48 - count))
}

func (r *LegacyRandom) NextInt() int32 { return r.NextBits(32) }

func (r *LegacyRandom) NextIntBound(bound int32) int32 {
	if bound <= 0 {
		panic("worldgen: random bound must be positive")
	}
	if bound&(bound-1) == 0 {
		return int32((int64(bound) * int64(r.NextBits(31))) >> 31)
	}
	for {
		bits := r.NextBits(31)
		value := bits % bound
		if bits-value+(bound-1) >= 0 {
			return value
		}
	}
}

func (r *LegacyRandom) NextLong() uint64 {
	high, low := int64(r.NextBits(32)), int64(r.NextBits(32))
	return uint64((high << 32) + low)
}

func (r *LegacyRandom) NextBoolean() bool { return r.NextBits(1) != 0 }
func (r *LegacyRandom) NextFloat() float32 {
	return float32(r.NextBits(24)) * (1.0 / (1 << 24))
}
func (r *LegacyRandom) NextDouble() float64 {
	hi, lo := int64(r.NextBits(26)), int64(r.NextBits(27))
	return float64((hi<<27)+lo) * (1.0 / (1 << 53))
}

func (r *LegacyRandom) Fork() *LegacyRandom { return NewLegacyRandom(int64(r.NextLong())) }

type LegacyPositionalRandom struct{ seed uint64 }

func (r *LegacyRandom) ForkPositional() LegacyPositionalRandom {
	return LegacyPositionalRandom{seed: r.NextLong()}
}

func (p LegacyPositionalRandom) At(x, y, z int32) *LegacyRandom {
	return NewLegacyRandom(int64(p.seed ^ uint64(positionSeed(x, y, z))))
}

func (p LegacyPositionalRandom) FromHashOf(name string) *LegacyRandom {
	var hash int32
	for _, unit := range utf16.Encode([]rune(name)) {
		hash = hash*31 + int32(unit)
	}
	return NewLegacyRandom(int64(p.seed ^ uint64(int64(hash))))
}
