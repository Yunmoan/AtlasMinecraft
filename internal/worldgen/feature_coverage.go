package worldgen

import (
	"encoding/json"
	"sort"
)

// OverworldFeatureCoverage records actual compiled runtime plans. Enabling a
// plan is separate from validating complete decorated worlds against Java.
type OverworldFeatureCoverage struct {
	Step       int    `json:"step"`
	Index      int    `json:"index"`
	Placed     string `json:"placed"`
	Configured string `json:"configured"`
	Type       string `json:"type"`
	Enabled    bool   `json:"enabled"`
	Reason     string `json:"reason,omitempty"`
}

func (generator *OverworldGenerator) FeatureCoverage() ([]OverworldFeatureCoverage, error) {
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		return nil, err
	}
	return featureCoverage(order, generator.plans)
}

func (generator *DimensionGenerator) FeatureCoverage() ([]OverworldFeatureCoverage, error) {
	return featureCoverage(generator.order, generator.plans)
}

func featureCoverage(order *featureOrder, plans map[string]decorationFeature) ([]OverworldFeatureCoverage, error) {
	result := make([]OverworldFeatureCoverage, 0, len(order.features))
	for _, feature := range order.features {
		placed, err := LoadPlacedFeatureDefinition(feature.Name)
		if err != nil {
			return nil, err
		}
		configured, err := LoadConfiguredFeatureDefinition(placed.Feature)
		if err != nil {
			return nil, err
		}
		_, enabled := plans[feature.Name]
		entry := OverworldFeatureCoverage{Step: feature.Step, Index: feature.Index, Placed: feature.Name, Configured: placed.Feature, Type: configured.Type, Enabled: enabled}
		if !enabled {
			if _, err := compileFeaturePlan(placed.Feature, 0); err != nil {
				entry.Reason = err.Error()
			} else {
				entry.Reason = "unsupported top-level placement"
			}
		}
		result = append(result, entry)
	}
	return result, nil
}

type StructureCoverage struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// StructureCoverage reports Overworld-native integrated starts and block
// placers, not placement or template APIs outside the chunk pipeline. The five
// Nether/End definitions in the shared registry have their own coverage below.
func (generator *OverworldGenerator) StructureCoverage() ([]StructureCoverage, error) {
	var data structureRegistry
	if err := json.Unmarshal(builtinStructureData, &data); err != nil {
		return nil, err
	}
	enabled := make(map[string]bool)
	if generator.structures != nil {
		for _, entry := range supportedOverworldStructures {
			enabled[entry.name] = true
		}
	}
	var result []StructureCoverage
	for name, raw := range data.Structures {
		switch name {
		case "minecraft:bastion_remnant", "minecraft:fortress", "minecraft:nether_fossil", "minecraft:ruined_portal_nether", "minecraft:end_city":
			continue
		}
		var definition struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &definition); err != nil {
			return nil, err
		}
		result = append(result, StructureCoverage{Name: name, Type: definition.Type, Enabled: enabled[name]})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// Dimension structure coverage includes only native structure families and
// requires an integrated block placer before marking a family enabled.
func (generator *DimensionGenerator) StructureCoverage() ([]StructureCoverage, error) {
	names := []string{"minecraft:end_city"}
	if generator.dimension == "nether" {
		names = []string{"minecraft:bastion_remnant", "minecraft:fortress", "minecraft:nether_fossil", "minecraft:ruined_portal_nether"}
	}
	var result []StructureCoverage
	for _, name := range names {
		data, err := BuiltinStructureJSON(name)
		if err != nil {
			return nil, err
		}
		var definition struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(data, &definition); err != nil {
			return nil, err
		}
		enabled := generator.structures != nil
		result = append(result, StructureCoverage{Name: name, Type: definition.Type, Enabled: enabled})
	}
	return result, nil
}
