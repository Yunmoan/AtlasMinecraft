package worldgen

// BlendedNoise is Java 1.21.1's old_blended_noise density function. Each of its
// three octave stacks consumes the terrain Xoroshiro source sequentially.
type BlendedNoise struct {
	minimum                                 [16]*ImprovedNoise
	maximum                                 [16]*ImprovedNoise
	main                                    [8]*ImprovedNoise
	xzMultiplier, yMultiplier               float64
	xzFactor, yFactor, smearScaleMultiplier float64
}

func NewBlendedNoise(worldSeed int64, xzScale, yScale, xzFactor, yFactor, smearScaleMultiplier float64) *BlendedNoise {
	random := NewXoroshiro(worldSeed).ForkPositional().FromHashOf("minecraft:terrain")
	b := &BlendedNoise{
		xzMultiplier:         684.412 * xzScale,
		yMultiplier:          684.412 * yScale,
		xzFactor:             xzFactor,
		yFactor:              yFactor,
		smearScaleMultiplier: smearScaleMultiplier,
	}
	fillBlendedOctaves(random, b.minimum[:])
	fillBlendedOctaves(random, b.maximum[:])
	fillBlendedOctaves(random, b.main[:])
	return b
}

func fillBlendedOctaves(random *Xoroshiro, octaves []*ImprovedNoise) {
	// PerlinNoise.create(random, IntStream.rangeClosed(-N+1, 0)) builds
	// octave zero first and then walks toward the lowest octave.
	for i := len(octaves) - 1; i >= 0; i-- {
		octaves[i] = NewImprovedNoise(random)
	}
}

func (b *BlendedNoise) Noise(x, y, z int32) float64 {
	dx := float64(x) * b.xzMultiplier
	dy := float64(y) * b.yMultiplier
	dz := float64(z) * b.xzMultiplier
	mainX, mainY, mainZ := dx/b.xzFactor, dy/b.yFactor, dz/b.xzFactor
	smear := b.yMultiplier * b.smearScaleMultiplier
	mainSmear := smear / b.yFactor
	var mainValue, minValue, maxValue float64
	frequency := 1.0
	for i := range b.main {
		octave := b.main[len(b.main)-1-i]
		mainValue += octave.NoiseShifted(
			wrapNoise(mainX*frequency), wrapNoise(mainY*frequency), wrapNoise(mainZ*frequency),
			mainSmear*frequency, mainY*frequency) / frequency
		frequency /= 2
	}
	interpolation := (mainValue/10 + 1) / 2
	skipMinimum := interpolation >= 1
	skipMaximum := interpolation <= 0
	frequency = 1
	for i := range b.minimum {
		x, y, z := wrapNoise(dx*frequency), wrapNoise(dy*frequency), wrapNoise(dz*frequency)
		yScale := smear * frequency
		yMax := dy * frequency
		if !skipMinimum {
			minValue += b.minimum[len(b.minimum)-1-i].NoiseShifted(x, y, z, yScale, yMax) / frequency
		}
		if !skipMaximum {
			maxValue += b.maximum[len(b.maximum)-1-i].NoiseShifted(x, y, z, yScale, yMax) / frequency
		}
		frequency /= 2
	}
	if interpolation < 0 {
		interpolation = 0
	} else if interpolation > 1 {
		interpolation = 1
	}
	return (minValue/512 + interpolation*(maxValue/512-minValue/512)) / 128
}
