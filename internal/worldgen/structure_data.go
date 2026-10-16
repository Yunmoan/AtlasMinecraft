package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed structure_data.json
var builtinStructureData []byte

type StructureSetEntry struct {
	Structure string `json:"structure"`
	Weight    int32  `json:"weight"`
}

type StructureSetDefinition struct {
	Structures []StructureSetEntry `json:"structures"`
	Placement  StructurePlacement  `json:"placement"`
}

type structureRegistry struct {
	Sets       map[string]json.RawMessage `json:"structure_set"`
	Structures map[string]json.RawMessage `json:"structure"`
	BiomeTags  map[string][]string        `json:"biome_tags"`
}

// LoadStructurePlacementState loads the original structure-set registry. The
// returned state contains no generated starts, pieces, or ring positions.
func LoadStructurePlacementState(seed int64) (*StructurePlacementState, error) {
	var data structureRegistry
	if err := json.Unmarshal(builtinStructureData, &data); err != nil {
		return nil, err
	}
	if len(data.Sets) != 20 {
		return nil, fmt.Errorf("expected 20 vanilla structure sets, got %d", len(data.Sets))
	}
	state := &StructurePlacementState{Seed: seed, Sets: make(map[string]StructureSetDefinition), RingPositions: make(map[string][]StructureChunkPosition)}
	for name, raw := range data.Sets {
		definition := StructureSetDefinition{Placement: StructurePlacement{Frequency: 1, FrequencyReductionMethod: "default", SpreadType: "linear"}}
		if err := json.Unmarshal(raw, &definition); err != nil {
			return nil, fmt.Errorf("structure set %s: %w", name, err)
		}
		if err := definition.Placement.validate(); err != nil {
			return nil, fmt.Errorf("structure set %s: %w", name, err)
		}
		if len(definition.Structures) == 0 {
			return nil, fmt.Errorf("structure set %s has no structures", name)
		}
		for _, entry := range definition.Structures {
			if entry.Weight <= 0 || data.Structures[entry.Structure] == nil {
				return nil, fmt.Errorf("invalid structure entry %s in set %s", entry.Structure, name)
			}
		}
		state.Sets[name] = definition
	}
	return state, nil
}

// BuiltinStructureJSON preserves every original structure setting for the
// subsequent structure-specific start and piece generators.
func BuiltinStructureJSON(name string) ([]byte, error) {
	var data structureRegistry
	if err := json.Unmarshal(builtinStructureData, &data); err != nil {
		return nil, err
	}
	raw, found := data.Structures[densityName(name)]
	if !found {
		return nil, fmt.Errorf("unknown vanilla structure %q", name)
	}
	return append([]byte(nil), raw...), nil
}

// StructureBiomeNames resolves the original structure biome holder set,
// including recursive biome tags, without inferring eligibility from terrain.
func StructureBiomeNames(name string) ([]string, error) {
	var data structureRegistry
	if err := json.Unmarshal(builtinStructureData, &data); err != nil {
		return nil, err
	}
	raw, found := data.Structures[densityName(name)]
	if !found {
		return nil, fmt.Errorf("unknown vanilla structure %q", name)
	}
	var settings struct {
		Biomes json.RawMessage `json:"biomes"`
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, err
	}
	var entries []string
	var one string
	if err := json.Unmarshal(settings.Biomes, &one); err == nil {
		entries = []string{one}
	} else if err := json.Unmarshal(settings.Biomes, &entries); err != nil {
		return nil, fmt.Errorf("structure %s biomes: %w", name, err)
	}
	return resolveStructureBiomes(data.BiomeTags, entries)
}

// StructurePreferredBiomeNames resolves the biome tag used to relocate
// concentric ring positions (stronghold_biased_to in the built-in Overworld).
func StructurePreferredBiomeNames(placement StructurePlacement) ([]string, error) {
	var data structureRegistry
	if err := json.Unmarshal(builtinStructureData, &data); err != nil {
		return nil, err
	}
	if placement.PreferredBiomes == "" {
		return nil, nil
	}
	return resolveStructureBiomes(data.BiomeTags, []string{placement.PreferredBiomes})
}

func resolveStructureBiomes(tags map[string][]string, entries []string) ([]string, error) {
	result := make(map[string]bool)
	visiting := make(map[string]bool)
	var resolve func(string) error
	resolve = func(entry string) error {
		if !strings.HasPrefix(entry, "#") {
			result[densityName(entry)] = true
			return nil
		}
		tag := densityName(entry[1:])
		values, found := tags[tag]
		if !found {
			return fmt.Errorf("unknown structure biome tag %s", tag)
		}
		if visiting[tag] {
			return fmt.Errorf("cyclic structure biome tag %s", tag)
		}
		visiting[tag] = true
		defer delete(visiting, tag)
		for _, value := range values {
			if err := resolve(value); err != nil {
				return err
			}
		}
		return nil
	}
	for _, entry := range entries {
		if err := resolve(entry); err != nil {
			return nil, err
		}
	}
	names := make([]string, 0, len(result))
	for name := range result {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
