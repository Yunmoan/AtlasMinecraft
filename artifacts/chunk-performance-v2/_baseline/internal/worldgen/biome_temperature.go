package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
)

//go:embed biome_climate_data.json
var biomeClimateData []byte

type biomeClimate struct {
	Temperature float32 `json:"temperature"`
	Modifier    string  `json:"modifier"`
}

var biomeClimateOnce sync.Once
var biomeClimates map[string]biomeClimate
var biomeClimateError error

func loadBiomeClimates() (map[string]biomeClimate, error) {
	biomeClimateOnce.Do(func() {
		biomeClimateError = json.Unmarshal(biomeClimateData, &biomeClimates)
		if biomeClimateError == nil && len(biomeClimates) != 64 {
			biomeClimateError = fmt.Errorf("expected 64 built-in biome climates, got %d", len(biomeClimates))
		}
	})
	return biomeClimates, biomeClimateError
}

type simplexNoise struct {
	permutation [256]uint8
}

func newSimplexNoise(random *LegacyRandom) *simplexNoise {
	for range 3 {
		random.NextDouble()
	}
	noise := &simplexNoise{}
	for index := range noise.permutation {
		noise.permutation[index] = uint8(index)
	}
	for index := range noise.permutation {
		swap := index + int(random.NextIntBound(int32(256-index)))
		noise.permutation[index], noise.permutation[swap] = noise.permutation[swap], noise.permutation[index]
	}
	return noise
}

func (noise *simplexNoise) perm(index int32) int32 {
	return int32(noise.permutation[uint8(index)])
}

func simplexContribution(gradient int32, x, z float64) float64 {
	attenuation := 0.5 - x*x - z*z
	if attenuation < 0 {
		return 0
	}
	attenuation *= attenuation
	return attenuation * attenuation * gradientDot(gradient, x, z, 0)
}

func (noise *simplexNoise) sample(x, z float64) float64 {
	skewFactor := (math.Sqrt(3) - 1) / 2
	unskewFactor := (3 - math.Sqrt(3)) / 6
	skew := (x + z) * skewFactor
	cellX, cellZ := fastFloor(x+skew), fastFloor(z+skew)
	unskew := float64(cellX+cellZ) * unskewFactor
	localX, localZ := x-(float64(cellX)-unskew), z-(float64(cellZ)-unskew)
	secondX, secondZ := int32(0), int32(1)
	if localX > localZ {
		secondX, secondZ = 1, 0
	}
	firstGradient := noise.perm(cellX+noise.perm(cellZ)) % 12
	secondGradient := noise.perm(cellX+secondX+noise.perm(cellZ+secondZ)) % 12
	thirdGradient := noise.perm(cellX+1+noise.perm(cellZ+1)) % 12
	first := simplexContribution(firstGradient, localX, localZ)
	second := simplexContribution(secondGradient, localX-float64(secondX)+unskewFactor, localZ-float64(secondZ)+unskewFactor)
	third := simplexContribution(thirdGradient, localX-1+2*unskewFactor, localZ-1+2*unskewFactor)
	return 70 * (first + second + third)
}

type biomeTemperatureNoise struct {
	height       *simplexNoise
	frozenDetail *simplexNoise
	frozen       [3]*simplexNoise
}

var temperatureNoiseOnce sync.Once
var temperatureNoise biomeTemperatureNoise

func temperatureNoises() biomeTemperatureNoise {
	temperatureNoiseOnce.Do(func() {
		temperatureNoise.height = newSimplexNoise(NewLegacyRandom(1234))
		temperatureNoise.frozenDetail = newSimplexNoise(NewLegacyRandom(2345))
		frozenRandom := NewLegacyRandom(3456)
		for index := range temperatureNoise.frozen {
			temperatureNoise.frozen[index] = newSimplexNoise(frozenRandom)
		}
	})
	return temperatureNoise
}

func (noise biomeTemperatureNoise) frozenValue(x, z float64) float64 {
	return noise.frozen[0].sample(x, z)/7 + noise.frozen[1].sample(x/2, z/2)*2/7 + noise.frozen[2].sample(x/4, z/4)*4/7
}

func BiomeTemperatureAt(biomeName string, x, y, z int32) (float32, error) {
	climates, err := loadBiomeClimates()
	if err != nil {
		return 0, err
	}
	climate, exists := climates[strings.TrimPrefix(biomeName, "minecraft:")]
	if !exists {
		return 0, fmt.Errorf("unknown 1.21.1 biome %q", biomeName)
	}
	noise := temperatureNoises()
	value := climate.Temperature
	if climate.Modifier == "frozen" {
		coarse := noise.frozenValue(float64(x)*0.05, float64(z)*0.05) * 7
		fine := noise.frozenDetail.sample(float64(x)*0.2, float64(z)*0.2)
		if coarse+fine < 0.3 && noise.frozenDetail.sample(float64(x)*0.09, float64(z)*0.09) < 0.8 {
			value = 0.2
		}
	} else if climate.Modifier != "none" {
		return 0, fmt.Errorf("unknown biome temperature modifier %q", climate.Modifier)
	}
	if y > 80 {
		offset := float32(noise.height.sample(float64(float32(x)/8), float64(float32(z)/8)) * 8)
		value -= (offset + float32(y) - 80) * float32(0.05) / float32(40)
	}
	return value, nil
}

func BiomeColdAt(biomeName string, x, y, z int32) (bool, error) {
	value, err := BiomeTemperatureAt(biomeName, x, y, z)
	return value < float32(0.15), err
}
