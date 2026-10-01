package worldgen

import (
	"bufio"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

//go:embed feature_order_1.21.1.txt
var overworldFeatureData string

type PlacedFeature struct {
	Step  int
	Index int
	Name  string
}

type featureOrder struct {
	features []PlacedFeature
	biomes   map[string]map[int][]int
	allowed  map[string]bool
}

var featureOrderOnce sync.Once
var overworldFeatures featureOrder
var featureOrderError error

func loadOverworldFeatureOrder() (*featureOrder, error) {
	featureOrderOnce.Do(func() {
		order := featureOrder{biomes: make(map[string]map[int][]int), allowed: make(map[string]bool)}
		indices := make(map[int]map[string]int)
		scanner := bufio.NewScanner(strings.NewReader(overworldFeatureData))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 0 || fields[0] == "schedule" {
				continue
			}
			if fields[0] == "biome" {
				if len(fields) != 5 {
					featureOrderError = fmt.Errorf("invalid biome feature row: %q", scanner.Text())
					return
				}
				step, stepErr := strconv.Atoi(fields[2])
				localIndex, indexErr := strconv.Atoi(fields[3])
				if stepErr != nil || indexErr != nil || step < 0 || localIndex < 0 {
					featureOrderError = fmt.Errorf("invalid biome feature index: %q", scanner.Text())
					return
				}
				if order.biomes[fields[1]] == nil {
					order.biomes[fields[1]] = make(map[int][]int)
				}
				if localIndex != len(order.biomes[fields[1]][step]) {
					featureOrderError = fmt.Errorf("unordered biome feature: %q", scanner.Text())
					return
				}
				index, exists := indices[step][fields[4]]
				if !exists {
					index = -1
				}
				order.biomes[fields[1]][step] = append(order.biomes[fields[1]][step], index)
				continue
			}
			if fields[0] != "feature" {
				featureOrderError = fmt.Errorf("invalid feature order row: %q", scanner.Text())
				return
			}
			if len(fields) != 4 {
				featureOrderError = fmt.Errorf("invalid placed feature row: %q", scanner.Text())
				return
			}
			step, stepErr := strconv.Atoi(fields[1])
			index, indexErr := strconv.Atoi(fields[2])
			if stepErr != nil || indexErr != nil || step < 0 || index < 0 {
				featureOrderError = fmt.Errorf("invalid placed feature index: %q", scanner.Text())
				return
			}
			if indices[step] == nil {
				indices[step] = make(map[string]int)
			}
			if previous, exists := indices[step][fields[3]]; exists || (len(order.features) > 0 && (step < order.features[len(order.features)-1].Step || step == order.features[len(order.features)-1].Step && index <= order.features[len(order.features)-1].Index)) {
				featureOrderError = fmt.Errorf("duplicate or unordered placed feature: %q (previous index %d)", scanner.Text(), previous)
				return
			}
			indices[step][fields[3]] = index
			order.features = append(order.features, PlacedFeature{step, index, fields[3]})
		}
		if featureOrderError == nil {
			featureOrderError = scanner.Err()
		}
		if featureOrderError != nil {
			return
		}
		if len(order.features) != 151 {
			featureOrderError = fmt.Errorf("expected 151 overworld placed features, got %d", len(order.features))
			return
		}
		points, err := OverworldBiomePoints()
		if err != nil {
			featureOrderError = err
			return
		}
		for _, point := range points {
			order.allowed[point.Biome] = true
		}
		overworldFeatures = order
	})
	return &overworldFeatures, featureOrderError
}

// OverworldFeatureSchedule returns the vanilla global feature order for the
// distinct biomes present in the 3x3 chunk neighborhood of a target chunk.
// This is only the decoration schedule; it does not place any features.
func OverworldFeatureSchedule(biomes []string) ([]PlacedFeature, error) {
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		return nil, err
	}
	selected := make(map[int]map[int]bool)
	for _, biome := range biomes {
		if !order.allowed[biome] {
			return nil, fmt.Errorf("biome %q is not in the 1.21.1 overworld source", biome)
		}
		for step, indices := range order.biomes[biome] {
			if selected[step] == nil {
				selected[step] = make(map[int]bool)
			}
			for _, index := range indices {
				if index < 0 {
					return nil, fmt.Errorf("biome %q has a feature absent from the overworld global order", biome)
				}
				selected[step][index] = true
			}
		}
	}
	var schedule []PlacedFeature
	for _, feature := range order.features {
		if selected[feature.Step][feature.Index] {
			schedule = append(schedule, feature)
		}
	}
	return schedule, nil
}

// OverworldDecorationSeed follows WorldgenRandom.setDecorationSeed with the
// 1.21.1 Xoroshiro source and the target chunk's block coordinates.
func OverworldDecorationSeed(worldSeed int64, chunkX, chunkZ int32) int64 {
	random := NewWorldgenRandom(worldSeed)
	xMultiplier := random.NextLong() | 1
	zMultiplier := random.NextLong() | 1
	return int64((uint64(int64(chunkX*16))*xMultiplier + uint64(int64(chunkZ*16))*zMultiplier) ^ uint64(worldSeed))
}

// OverworldPlacedFeatureRandom follows WorldgenRandom.setFeatureSeed.
func OverworldPlacedFeatureRandom(decorationSeed int64, feature PlacedFeature) *WorldgenRandom {
	return NewWorldgenRandom(decorationSeed + int64(feature.Index) + int64(10000*feature.Step))
}
