package worldgen

import (
	"bufio"
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

// The built-in 1.21.1 OverworldBiomeBuilder emits these points. They were
// extracted by docs/worldgen/BiomePresetProbe.java from the supplied JAR.
//
//go:embed biome_preset_1.21.1.txt
var overworldBiomeData string

type BiomeParameterPoint struct {
	Biome  string
	Ranges [6][2]int64 // temperature, humidity, continentalness, erosion, depth, weirdness
	Offset int64
}

var biomePointsOnce sync.Once
var biomePoints []BiomeParameterPoint
var biomePointsError error

func OverworldBiomePoints() ([]BiomeParameterPoint, error) {
	biomePointsOnce.Do(func() {
		scanner := bufio.NewScanner(strings.NewReader(overworldBiomeData))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 0 || fields[0] == "query" || fields[0] == "sequence" {
				continue
			}
			if len(fields) != 15 || fields[0] != "point" {
				biomePointsError = fmt.Errorf("invalid 1.21.1 biome parameter row: %q", scanner.Text())
				return
			}
			var point BiomeParameterPoint
			point.Biome = fields[1]
			for i := 0; i < 6; i++ {
				for j := 0; j < 2; j++ {
					value, err := strconv.ParseInt(fields[2+i*2+j], 10, 64)
					if err != nil {
						biomePointsError = err
						return
					}
					point.Ranges[i][j] = value
				}
			}
			point.Offset, biomePointsError = strconv.ParseInt(fields[14], 10, 64)
			if biomePointsError != nil {
				return
			}
			biomePoints = append(biomePoints, point)
		}
		if biomePointsError == nil {
			biomePointsError = scanner.Err()
		}
		if biomePointsError == nil && len(biomePoints) != 7593 {
			biomePointsError = fmt.Errorf("expected 7593 overworld biome points, got %d", len(biomePoints))
		}
	})
	return biomePoints, biomePointsError
}

// QuantizeClimate follows Climate.quantizeCoord(float), including float32
// rounding and truncation toward zero.
func QuantizeClimate(value float64) int64 {
	return int64(float32(value) * 10000)
}

// SelectOverworldBiome chooses the closest built-in climate parameter point.
// Ties retain the first leaf in vanilla's climate search-tree order.
func SelectOverworldBiome(climate [6]int64) (string, error) {
	points, err := OverworldBiomePoints()
	if err != nil {
		return "", err
	}
	index, err := overworldBiomeIndex()
	if err != nil {
		return "", err
	}
	return points[index.nearest(points, climate, -1)].Biome, nil
}

func biomeDistance(point BiomeParameterPoint, climate [6]int64) int64 {
	distance := point.Offset * point.Offset
	for i, bounds := range point.Ranges {
		delta := int64(0)
		if climate[i] < bounds[0] {
			delta = bounds[0] - climate[i]
		} else if climate[i] > bounds[1] {
			delta = climate[i] - bounds[1]
		}
		distance += delta * delta
	}
	return distance
}

func nearestBiomePoint(points []BiomeParameterPoint, climate [6]int64, previous int) int {
	best := int64(math.MaxInt64)
	index := 0
	if previous >= 0 && previous < len(points) {
		best, index = biomeDistance(points[previous], climate), previous
	}
	for i, point := range points {
		if distance := biomeDistance(point, climate); distance < best {
			best, index = distance, i
		}
	}
	return index
}

// OverworldBiomeSampler combines the vanilla noise router with the biome
// parameter list at quart coordinates (one biome sample per four blocks).
type OverworldBiomeSampler struct {
	settings *DensitySettings
	density  *DensitySampler
	index    *biomeSearchNode
	previous int // vanilla's per-thread previous biome tie preference
}

func NewOverworldBiomeSampler(seed int64) (*OverworldBiomeSampler, error) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		return nil, err
	}
	return newOverworldBiomeSampler(seed, settings)
}

// Density settings and the index are immutable; sampling history/noise caches
// remain private to each chunk so generation workers never share mutable state.
func newOverworldBiomeSampler(seed int64, settings *DensitySettings) (*OverworldBiomeSampler, error) {
	index, err := overworldBiomeIndex()
	if err != nil {
		return nil, err
	}
	return &OverworldBiomeSampler{settings: settings, density: NewDensitySampler(seed), index: index, previous: -1}, nil
}

func (s *OverworldBiomeSampler) BiomeAtQuart(qx, qy, qz int32) (string, error) {
	climate, err := s.ClimateAtBlock(qx*4, qy*4, qz*4)
	if err != nil {
		return "", err
	}
	points, err := OverworldBiomePoints()
	if err != nil {
		return "", err
	}
	s.previous = s.index.nearest(points, climate, s.previous)
	return points[s.previous].Biome, nil
}

// ClimateAtBlock samples the six vanilla climate density functions at block
// coordinates and quantizes them exactly like Climate.Sampler.sample.
func (s *OverworldBiomeSampler) ClimateAtBlock(x, y, z int32) ([6]int64, error) {
	var climate [6]int64
	for i, field := range climateFields {
		value, err := s.density.Sample(s.settings.Router[field], x, y, z)
		if err != nil {
			return climate, fmt.Errorf("%s biome climate: %w", field, err)
		}
		climate[i] = QuantizeClimate(value)
	}
	return climate, nil
}
