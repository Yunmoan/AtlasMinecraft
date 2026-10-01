package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// These registry entries come from the bundled 1.21.1 server JAR. The JAR is
// used only by the extraction and verification scripts, never at runtime.
//
//go:embed density_data.json
var builtinDensityData []byte

type densityRegistry struct {
	Functions map[string]json.RawMessage `json:"density_function"`
	Settings  map[string]json.RawMessage `json:"noise_settings"`
}

var densityOnce sync.Once
var densities densityRegistry
var densityError error

func getDensityRegistry() (*densityRegistry, error) {
	densityOnce.Do(func() {
		densityError = json.Unmarshal(builtinDensityData, &densities)
		if densityError == nil && (len(densities.Functions) != 35 || len(densities.Settings) != 7) {
			densityError = fmt.Errorf("unexpected built-in 1.21.1 density registry size: %d functions, %d settings", len(densities.Functions), len(densities.Settings))
		}
	})
	return &densities, densityError
}

func densityName(name string) string {
	if !strings.Contains(name, ":") {
		return "minecraft:" + name
	}
	return name
}

func BuiltinDensityNames() ([]string, error) {
	r, err := getDensityRegistry()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(r.Functions))
	for name := range r.Functions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func BuiltinNoiseSettingsNames() ([]string, error) {
	r, err := getDensityRegistry()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(r.Settings))
	for name := range r.Settings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// BuiltinNoiseSettingsJSON returns an independent copy of the complete vanilla
// settings entry, including surface rules, for later generation stages.
func BuiltinNoiseSettingsJSON(name string) ([]byte, error) {
	r, err := getDensityRegistry()
	if err != nil {
		return nil, err
	}
	data, ok := r.Settings[densityName(name)]
	if !ok {
		return nil, fmt.Errorf("unknown 1.21.1 noise settings %q", name)
	}
	return append([]byte(nil), data...), nil
}
