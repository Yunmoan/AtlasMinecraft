package worldgen

import (
	"fmt"
	"math"
)

// PerlinNoise combines ImprovedNoise octaves using Minecraft's modern
// positional Xoroshiro initialization path.
type PerlinNoise struct {
	levels      []*ImprovedNoise
	amplitudes  []float64
	inputFactor float64
	valueFactor float64
}

func NewPerlinNoise(random *Xoroshiro, firstOctave int, amplitudes []float64) (*PerlinNoise, error) {
	positional := random.ForkPositional()
	return newPositionalPerlinNoise(firstOctave, amplitudes, func(name string) octaveRandom { return positional.FromHashOf(name) })
}

func newPositionalPerlinNoise(firstOctave int, amplitudes []float64, fromHash func(string) octaveRandom) (*PerlinNoise, error) {
	if len(amplitudes) == 0 || len(amplitudes) > 64 {
		return nil, fmt.Errorf("invalid octave count %d", len(amplitudes))
	}
	p := &PerlinNoise{
		levels:      make([]*ImprovedNoise, len(amplitudes)),
		amplitudes:  append([]float64(nil), amplitudes...),
		inputFactor: math.Pow(2, float64(firstOctave)),
		valueFactor: math.Pow(2, float64(len(amplitudes)-1)) / (math.Pow(2, float64(len(amplitudes))) - 1),
	}
	for i, amplitude := range amplitudes {
		if amplitude != 0 {
			name := fmt.Sprintf("octave_%d", firstOctave+i)
			p.levels[i] = newImprovedNoise(fromHash(name))
		}
	}
	return p, nil
}

func (p *PerlinNoise) Noise(x, y, z float64) float64 {
	value := 0.0
	frequency := p.inputFactor
	weight := p.valueFactor
	for i, level := range p.levels {
		if level != nil {
			sample := level.NoiseShifted(wrapNoise(x*frequency), wrapNoise(y*frequency), wrapNoise(z*frequency), 0, 0)
			value += p.amplitudes[i] * sample * weight
		}
		frequency *= 2
		weight /= 2
	}
	return value
}

func wrapNoise(x float64) float64 {
	const period = 33554432.0
	return x - float64(fastFloorLong(x/period+0.5))*period
}

func fastFloorLong(x float64) int64 {
	i := int64(x)
	if x < float64(i) {
		return i - 1
	}
	return i
}
