package worldgen

// Vanilla 1.21.1 initial spawn search (MinecraftServer.setInitialSpawn plus
// Climate$SpawnFinder), transcribed from the reference JAR and cross-checked
// against docs/worldgen/spawn_search_1.21.1.txt. The search ring-samples the
// six overworld climate density functions and picks the position whose
// climate best matches the noise settings' spawn_target parameter points;
// an origin-distance penalty keeps the result near (0,0). Deterministic per
// seed and free of chunk generation, so world creation can run it directly.

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
)

// climateFields maps the vanilla TargetPoint order to the density-router
// field names: temperature, humidity, continentalness, erosion, depth,
// weirdness.
var climateFields = [...]string{"temperature", "vegetation", "continents", "erosion", "depth", "ridges"}

type spawnTargetPoint struct {
	ranges [6][2]int64 // quantized inclusive min/max per climate field
	offset int64
}

// fitness follows Climate$ParameterPoint.fitness: the summed squared distances
// of each climate value to its parameter range plus the squared offset. The
// sampled depth is forced to zero before this runs, matching SpawnFinder.
func (t spawnTargetPoint) fitness(climate [6]int64) int64 {
	var sum int64
	for i, bounds := range t.ranges {
		distance := climateDistance(bounds, climate[i])
		sum += distance * distance
	}
	return sum + t.offset*t.offset
}

func climateDistance(bounds [2]int64, value int64) int64 {
	if value > bounds[1] {
		return value - bounds[1]
	}
	if bounds[0] > value {
		return bounds[0] - value
	}
	return 0
}

var (
	spawnTargetOnce sync.Once
	spawnTargets    []spawnTargetPoint
	spawnTargetErr  error
)

func getSpawnTargets() ([]spawnTargetPoint, error) {
	spawnTargetOnce.Do(func() {
		registry, err := getDensityRegistry()
		if err != nil {
			spawnTargetErr = err
			return
		}
		raw, ok := registry.Settings["minecraft:overworld"]
		if !ok {
			spawnTargetErr = fmt.Errorf("missing overworld noise settings for spawn_target")
			return
		}
		var settings struct {
			SpawnTarget []struct {
				Temperature     json.RawMessage `json:"temperature"`
				Humidity        json.RawMessage `json:"humidity"`
				Continentalness json.RawMessage `json:"continentalness"`
				Erosion         json.RawMessage `json:"erosion"`
				Depth           json.RawMessage `json:"depth"`
				Weirdness       json.RawMessage `json:"weirdness"`
				Offset          float64         `json:"offset"`
			} `json:"spawn_target"`
		}
		if err := json.Unmarshal(raw, &settings); err != nil {
			spawnTargetErr = fmt.Errorf("parse overworld spawn_target: %w", err)
			return
		}
		for _, entry := range settings.SpawnTarget {
			var point spawnTargetPoint
			fields := [...]json.RawMessage{entry.Temperature, entry.Humidity, entry.Continentalness, entry.Erosion, entry.Depth, entry.Weirdness}
			for i, raw := range fields {
				var bounds [2]float64
				if err := json.Unmarshal(raw, &bounds); err == nil {
					point.ranges[i] = [2]int64{QuantizeClimate(bounds[0]), QuantizeClimate(bounds[1])}
					continue
				}
				var scalar float64
				if err := json.Unmarshal(raw, &scalar); err != nil {
					spawnTargetErr = fmt.Errorf("parse overworld spawn_target parameter %d: %w", i, err)
					return
				}
				quantized := QuantizeClimate(scalar)
				point.ranges[i] = [2]int64{quantized, quantized}
			}
			point.offset = QuantizeClimate(entry.Offset)
			spawnTargets = append(spawnTargets, point)
		}
		if len(spawnTargets) == 0 {
			spawnTargetErr = fmt.Errorf("overworld noise settings define no spawn_target")
		}
	})
	return spawnTargets, spawnTargetErr
}

type spawnSearchResult struct {
	x, z    int32
	fitness int64
}

// FindSpawnPosition returns the block position vanilla selects for this seed.
// The Y coordinate is always 0; callers derive the spawn chunk from X/Z.
func FindSpawnPosition(seed int64) (int32, int32, error) {
	targets, err := getSpawnTargets()
	if err != nil {
		return 0, 0, err
	}
	sampler, err := NewOverworldBiomeSampler(seed)
	if err != nil {
		return 0, 0, err
	}
	result := spawnFitness(targets, sampler, 0, 0)
	result = spawnRingSearch(targets, sampler, result, 2048, 512)
	result = spawnRingSearch(targets, sampler, result, 512, 32)
	return result.x, result.z, nil
}

// spawnRingSearch mirrors Climate$SpawnFinder.search: spiral ring sampling
// with float32 angle accumulation around the entry's best position.
func spawnRingSearch(targets []spawnTargetPoint, sampler *OverworldBiomeSampler, best spawnSearchResult, maxRadius, step float32) spawnSearchResult {
	center := best
	var angle float32
	radius := step
	for radius <= maxRadius {
		x := center.x + int32(math.Sin(float64(angle))*float64(radius))
		z := center.z + int32(math.Cos(float64(angle))*float64(radius))
		if candidate := spawnFitness(targets, sampler, x, z); candidate.fitness < best.fitness {
			best = candidate
		}
		angle += step / radius
		if float64(angle) > math.Pi*2 {
			angle = 0
			radius += step
		}
	}
	return best
}

// spawnFitness mirrors Climate$SpawnFinder.getFitness: the sampled climate at
// the four-block grid under (x,z), depth forced to zero, scored against the
// spawn targets plus an origin-distance penalty.
func spawnFitness(targets []spawnTargetPoint, sampler *OverworldBiomeSampler, x, z int32) spawnSearchResult {
	climate, err := sampler.ClimateAtBlock(x>>2<<2, 0, z>>2<<2)
	if err != nil {
		// Climate sampling is total for built-in overworld settings; a failure
		// keeps the candidate ineligible rather than aborting the search.
		return spawnSearchResult{x: x, z: z, fitness: math.MaxInt64}
	}
	climate[4] = 0
	best := int64(math.MaxInt64)
	for _, target := range targets {
		if fitness := target.fitness(climate); fitness < best {
			best = fitness
		}
	}
	// Mth.square(2500.0) as the divisor and Mth.square(10000.0f) as the factor
	// applied to the squared block distance, exactly as in the JAR.
	squared := int64(x)*int64(x) + int64(z)*int64(z)
	penalty := int64(1e8 * math.Pow(float64(squared)/6250000.0, 2.0))
	return spawnSearchResult{x: x, z: z, fitness: penalty + best}
}
