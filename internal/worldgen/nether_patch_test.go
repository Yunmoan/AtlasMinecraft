package worldgen

import (
	"encoding/json"
	"testing"
)

func TestNetherPatchesAgainstVanillaJAR(t *testing.T) {
	testNetherDecorationVectors(t, "nether_patch_1.21.1.txt", "nether_patch", func(name string) (configuredFeaturePlacer, error) { return compileFeaturePlan(name, 0) }, 720, 152)
}

func TestNetherPatchPlacementsAgainstVanillaJAR(t *testing.T) {
	var counts [2]int
	compiler := func(name string) (configuredFeaturePlacer, error) {
		definition, err := LoadPlacedFeatureDefinition(name)
		if err != nil {
			return nil, err
		}
		var modifiers []json.RawMessage
		for _, raw := range definition.Placement {
			var kind struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(raw, &kind); err != nil {
				return nil, err
			}
			// The independent JAR fixture accepts the biome to isolate placement,
			// configured feature execution and their shared lazy random stream.
			if kind.Type != "minecraft:biome" {
				modifiers = append(modifiers, raw)
			}
		}
		definition.Placement = modifiers
		template, err := compilePlacementTemplate(definition, nil)
		if err != nil {
			return nil, err
		}
		place, err := compileFeaturePlan(definition.Feature, 0)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
			counts = [2]int{}
			operations, err := template.bind(region, nil)
			if err != nil {
				return false, err
			}
			err = walkPlacement(operations, r, origin, func(p FeaturePosition, r *WorldgenRandom) error {
				counts[0]++
				placed, err := place(region, r, p)
				if placed {
					counts[1]++
				}
				return err
			})
			return counts[1] > 0, err
		}, nil
	}
	testNetherDecorationVectors(t, "nether_patch_placed_1.21.1.txt", "nether_patch_placed", compiler, 5760, 171, func() [2]int { return counts })
}
