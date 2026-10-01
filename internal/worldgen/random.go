// Package worldgen contains exact Java 1.21.1 generation primitives.
// A complete noise, feature, and structure generator is not yet implemented.
package worldgen

import (
	"crypto/md5"
	"encoding/binary"
	"math/bits"
)

// Xoroshiro implements Minecraft 1.21.1 XoroshiroRandomSource.
type Xoroshiro struct {
	lo, hi uint64
}

func NewXoroshiro(seed int64) *Xoroshiro {
	lo := uint64(seed) ^ 0x6a09e667f3bcc909
	hi := lo + 0x9e3779b97f4a7c15
	return NewXoroshiroPair(mixStafford13(lo), mixStafford13(hi))
}

func NewXoroshiroPair(lo, hi uint64) *Xoroshiro {
	if lo|hi == 0 {
		lo, hi = 0x9e3779b97f4a7c15, 0x6a09e667f3bcc909
	}
	return &Xoroshiro{lo: lo, hi: hi}
}

func mixStafford13(x uint64) uint64 {
	x = (x ^ x>>30) * 0xbf58476d1ce4e5b9
	x = (x ^ x>>27) * 0x94d049bb133111eb
	return x ^ x>>31
}

func (r *Xoroshiro) NextLong() uint64 {
	lo, hi := r.lo, r.hi
	result := bits.RotateLeft64(lo+hi, 17) + lo
	hi ^= lo
	r.lo = bits.RotateLeft64(lo, 49) ^ hi ^ hi<<21
	r.hi = bits.RotateLeft64(hi, 28)
	return result
}

func (r *Xoroshiro) NextInt() int32 { return int32(r.NextLong()) }

// NextIntBound uses the unsigned multiply/rejection path in 1.21.1.
func (r *Xoroshiro) NextIntBound(bound int32) int32 {
	if bound <= 0 {
		panic("worldgen: random bound must be positive")
	}
	b := uint32(bound)
	threshold := (uint32(0) - b) % b
	for {
		product := uint64(uint32(r.NextInt())) * uint64(b)
		if uint32(product) >= threshold {
			return int32(product >> 32)
		}
	}
}

func (r *Xoroshiro) NextBoolean() bool { return r.NextLong()&1 != 0 }
func (r *Xoroshiro) NextFloat() float32 {
	return float32(r.NextLong()>>40) * (1.0 / (1 << 24))
}
func (r *Xoroshiro) NextDouble() float64 {
	return float64(r.NextLong()>>11) * (1.0 / (1 << 53))
}

func (r *Xoroshiro) Fork() *Xoroshiro {
	return NewXoroshiroPair(r.NextLong(), r.NextLong())
}

type PositionalRandom struct{ lo, hi uint64 }

func (r *Xoroshiro) ForkPositional() PositionalRandom {
	return PositionalRandom{lo: r.NextLong(), hi: r.NextLong()}
}

func (p PositionalRandom) At(x, y, z int32) *Xoroshiro {
	return NewXoroshiroPair(p.lo^uint64(positionSeed(x, y, z)), p.hi)
}

func (p PositionalRandom) FromHashOf(name string) *Xoroshiro {
	hash := md5.Sum([]byte(name))
	return NewXoroshiroPair(p.lo^binary.BigEndian.Uint64(hash[:8]), p.hi^binary.BigEndian.Uint64(hash[8:]))
}

func positionSeed(x, y, z int32) int64 {
	// Java multiplies x as an int before widening it to long.
	n := int64(x*3129871) ^ int64(z)*116129781 ^ int64(y)
	n = n*n*42317861 + n*11
	return n >> 16
}
