package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed feature_data.json
var featureDefinitionsData []byte

type PlacedFeatureDefinition struct {
	Feature   string            `json:"feature"`
	Placement []json.RawMessage `json:"placement"`
}

type ConfiguredFeatureDefinition struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

type featureDefinitions struct {
	Placed     map[string]PlacedFeatureDefinition     `json:"placed_feature"`
	Configured map[string]ConfiguredFeatureDefinition `json:"configured_feature"`
	OreTags    map[string][]string                     `json:"ore_replaceable_tags"`
}

var featureDefinitionsOnce sync.Once
var builtInFeatureDefinitions featureDefinitions
var featureDefinitionsError error

func loadFeatureDefinitions() (*featureDefinitions, error) {
	featureDefinitionsOnce.Do(func() {
		featureDefinitionsError = json.Unmarshal(featureDefinitionsData, &builtInFeatureDefinitions)
		if featureDefinitionsError != nil {
			return
		}
		if len(builtInFeatureDefinitions.Placed) != 233 || len(builtInFeatureDefinitions.Configured) != 196 {
			featureDefinitionsError = fmt.Errorf("incomplete 1.21.1 feature definitions: %d placed, %d configured",
				len(builtInFeatureDefinitions.Placed), len(builtInFeatureDefinitions.Configured))
			return
		}
		if len(builtInFeatureDefinitions.OreTags) != 4 {
			featureDefinitionsError = fmt.Errorf("expected four 1.21.1 ore replacement tags, got %d", len(builtInFeatureDefinitions.OreTags))
			return
		}
		for name, placed := range builtInFeatureDefinitions.Placed {
			if _, ok := builtInFeatureDefinitions.Configured[placed.Feature]; !ok {
				featureDefinitionsError = fmt.Errorf("placed feature %s references unknown configured feature %s", name, placed.Feature)
				return
			}
			for _, placement := range placed.Placement {
				var modifier struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(placement, &modifier); err != nil || modifier.Type == "" {
					featureDefinitionsError = fmt.Errorf("placed feature %s has invalid placement modifier: %s", name, placement)
					return
				}
			}
		}
		for name, configured := range builtInFeatureDefinitions.Configured {
			if configured.Type == "" || len(configured.Config) == 0 {
				featureDefinitionsError = fmt.Errorf("configured feature %s has missing type or config", name)
				return
			}
		}
	})
	return &builtInFeatureDefinitions, featureDefinitionsError
}

func LoadPlacedFeatureDefinition(name string) (PlacedFeatureDefinition, error) {
	definitions, err := loadFeatureDefinitions()
	if err != nil {
		return PlacedFeatureDefinition{}, err
	}
	definition, ok := definitions.Placed[name]
	if !ok {
		return PlacedFeatureDefinition{}, fmt.Errorf("unknown 1.21.1 placed feature %q", name)
	}
	return definition, nil
}

func LoadConfiguredFeatureDefinition(name string) (ConfiguredFeatureDefinition, error) {
	definitions, err := loadFeatureDefinitions()
	if err != nil {
		return ConfiguredFeatureDefinition{}, err
	}
	definition, ok := definitions.Configured[name]
	if !ok {
		return ConfiguredFeatureDefinition{}, fmt.Errorf("unknown 1.21.1 configured feature %q", name)
	}
	return definition, nil
}
