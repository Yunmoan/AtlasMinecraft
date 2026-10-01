package worldgen

import "errors"

// NormalNoise is the two-Perlin sampler used by vanilla density functions.
type NormalNoise struct {
	first, second *PerlinNoise
	valueFactor   float64
}

func NewNormalNoise(random *Xoroshiro, firstOctave int, amplitudes []float64) (*NormalNoise, error) {
	return newNormalNoise(firstOctave, amplitudes, func() (*PerlinNoise, error) { return NewPerlinNoise(random, firstOctave, amplitudes) })
}

// State providers use modern positional octaves seeded by LegacyRandomSource,
// rather than the Xoroshiro source used by the terrain density router.
func newProviderNormalNoise(seed int64, firstOctave int, amplitudes []float64) (*NormalNoise, error) {
	random := NewLegacyRandom(seed)
	return newNormalNoise(firstOctave, amplitudes, func() (*PerlinNoise, error) {
		positional := random.ForkPositional()
		return newPositionalPerlinNoise(firstOctave, amplitudes, func(name string) octaveRandom { return positional.FromHashOf(name) })
	})
}

func newNormalNoise(firstOctave int, amplitudes []float64, perlin func() (*PerlinNoise, error)) (*NormalNoise, error) {
	minimum, maximum := len(amplitudes), -1
	for i, amplitude := range amplitudes {
		if amplitude != 0 {
			if i < minimum {
				minimum = i
			}
			maximum = i
		}
	}
	if maximum < minimum {
		return nil, errors.New("normal noise needs a nonzero amplitude")
	}
	first, err := perlin()
	if err != nil {
		return nil, err
	}
	second, err := perlin()
	if err != nil {
		return nil, err
	}
	deviation := 0.1 * (1 + 1/float64(maximum-minimum+1))
	return &NormalNoise{first: first, second: second, valueFactor: 0.16666666666666666 / deviation}, nil
}

func (n *NormalNoise) Noise(x, y, z float64) float64 {
	const inputFactor = 1.0181268882175227
	shiftX, shiftY, shiftZ := x*inputFactor, y*inputFactor, z*inputFactor
	return (n.first.Noise(x, y, z) + n.second.Noise(shiftX, shiftY, shiftZ)) * n.valueFactor
}
