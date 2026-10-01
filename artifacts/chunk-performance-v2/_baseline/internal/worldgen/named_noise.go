package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed noise_parameters.json
var vanillaNoiseParameters []byte

type NoiseParameters struct {
	FirstOctave int       `json:"firstOctave"`
	Amplitudes  []float64 `json:"amplitudes"`
}

var noiseParametersOnce sync.Once
var noiseParameters map[string]NoiseParameters
var noiseParametersError error

func Parameters(name string) (NoiseParameters, error) {
	noiseParametersOnce.Do(func() {
		noiseParametersError = json.Unmarshal(vanillaNoiseParameters, &noiseParameters)
	})
	if noiseParametersError != nil {
		return NoiseParameters{}, noiseParametersError
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	p, ok := noiseParameters[name]
	if !ok {
		return NoiseParameters{}, fmt.Errorf("unknown 1.21.1 noise %q", name)
	}
	p.Amplitudes = append([]float64(nil), p.Amplitudes...)
	return p, nil
}

// NewNamedNormalNoise follows RandomState / Noises: fork the world's
// Xoroshiro source, hash the namespaced noise key, then build two Perlin
// samplers from the registry's octave parameters.
func NewNamedNormalNoise(worldSeed int64, name string) (*NormalNoise, error) {
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	p, err := Parameters(name)
	if err != nil {
		return nil, err
	}
	random := NewXoroshiro(worldSeed).ForkPositional().FromHashOf(name)
	return NewNormalNoise(random, p.FirstOctave, p.Amplitudes)
}
