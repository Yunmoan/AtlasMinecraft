package worldgen

import (
	"bufio"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

//go:embed dimension_feature_order_1.21.1.txt
var dimensionFeatureData string

var dimensionFeatureOnce sync.Once
var dimensionFeatures map[string]*featureOrder
var dimensionFeatureError error

func loadDimensionFeatureOrder(dimension string) (*featureOrder, error) {
	dimensionFeatureOnce.Do(func() {
		dimensionFeatures = make(map[string]*featureOrder)
		indices := make(map[string]map[int]map[string]int)
		for _, name := range []string{"nether", "end"} {
			dimensionFeatures[name] = &featureOrder{biomes: make(map[string]map[int][]int), allowed: make(map[string]bool)}
			indices[name] = make(map[int]map[string]int)
		}
		scanner := bufio.NewScanner(strings.NewReader(dimensionFeatureData))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 0 {
				continue
			}
			if len(fields) < 3 || dimensionFeatures[fields[1]] == nil {
				dimensionFeatureError = fmt.Errorf("invalid dimension feature row: %q", scanner.Text())
				return
			}
			order := dimensionFeatures[fields[1]]
			switch fields[0] {
			case "schedule":
				continue
			case "source":
				if len(fields) != 3 {
					dimensionFeatureError = fmt.Errorf("invalid source row: %q", scanner.Text())
					return
				}
				order.allowed[fields[2]] = true
			case "feature":
				if len(fields) != 5 {
					dimensionFeatureError = fmt.Errorf("invalid feature row: %q", scanner.Text())
					return
				}
				step, err1 := strconv.Atoi(fields[2])
				index, err2 := strconv.Atoi(fields[3])
				if err1 != nil || err2 != nil || step < 0 || index < 0 {
					dimensionFeatureError = fmt.Errorf("invalid feature index: %q", scanner.Text())
					return
				}
				if indices[fields[1]][step] == nil {
					indices[fields[1]][step] = make(map[string]int)
				}
				stepIndices := indices[fields[1]][step]
				if _, exists := stepIndices[fields[4]]; exists || index != len(stepIndices) {
					dimensionFeatureError = fmt.Errorf("unordered feature row: %q", scanner.Text())
					return
				}
				stepIndices[fields[4]] = index
				order.features = append(order.features, PlacedFeature{step, index, fields[4]})
			case "biome":
				if len(fields) != 6 {
					dimensionFeatureError = fmt.Errorf("invalid biome row: %q", scanner.Text())
					return
				}
				step, err1 := strconv.Atoi(fields[3])
				local, err2 := strconv.Atoi(fields[4])
				if err1 != nil || err2 != nil || step < 0 || local < 0 || !order.allowed[fields[2]] {
					dimensionFeatureError = fmt.Errorf("invalid biome index: %q", scanner.Text())
					return
				}
				if order.biomes[fields[2]] == nil {
					order.biomes[fields[2]] = make(map[int][]int)
				}
				global, exists := indices[fields[1]][step][fields[5]]
				if !exists || local != len(order.biomes[fields[2]][step]) {
					dimensionFeatureError = fmt.Errorf("missing or unordered biome feature: %q", scanner.Text())
					return
				}
				order.biomes[fields[2]][step] = append(order.biomes[fields[2]][step], global)
			default:
				dimensionFeatureError = fmt.Errorf("unknown dimension feature row: %q", scanner.Text())
				return
			}
		}
		dimensionFeatureError = scanner.Err()
	})
	if dimensionFeatureError != nil {
		return nil, dimensionFeatureError
	}
	order := dimensionFeatures[dimension]
	if order == nil {
		return nil, fmt.Errorf("unsupported dimension feature order %q", dimension)
	}
	return order, nil
}

func DimensionFeatureSchedule(dimension string, biomes []string) ([]PlacedFeature, error) {
	order, err := loadDimensionFeatureOrder(dimension)
	if err != nil {
		return nil, err
	}
	selected := make(map[int]map[int]bool)
	for _, biome := range biomes {
		if !order.allowed[biome] {
			return nil, fmt.Errorf("biome %q is not in the 1.21.1 %s source", biome, dimension)
		}
		for step, indices := range order.biomes[biome] {
			if selected[step] == nil {
				selected[step] = make(map[int]bool)
			}
			for _, index := range indices {
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
