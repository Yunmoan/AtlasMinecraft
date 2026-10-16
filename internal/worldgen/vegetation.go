package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed vegetation_tags.json
var vegetationTagData []byte

var vegetationTags = func() map[string]map[string]bool {
	var data map[string][]string
	if err := json.Unmarshal(vegetationTagData, &data); err != nil {
		panic(err)
	}
	tags := make(map[string]map[string]bool)
	for name, blocks := range data {
		tags[name] = make(map[string]bool)
		for _, block := range blocks {
			tags[name][block] = true
		}
	}
	return tags
}()

func setFeatureBlock(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState) error {
	if !world.InWorld(p) {
		return nil
	}
	return world.SetBlock(p, state)
}

var groundPlants = map[string]bool{
	"minecraft:short_grass": true, "minecraft:fern": true,
	"minecraft:tall_grass": true, "minecraft:large_fern": true,
	"minecraft:sunflower": true, "minecraft:lilac": true, "minecraft:rose_bush": true, "minecraft:peony": true,
	"minecraft:dandelion": true, "minecraft:poppy": true, "minecraft:blue_orchid": true,
	"minecraft:allium": true, "minecraft:azure_bluet": true,
	"minecraft:red_tulip": true, "minecraft:orange_tulip": true, "minecraft:white_tulip": true, "minecraft:pink_tulip": true,
	"minecraft:oxeye_daisy": true, "minecraft:cornflower": true, "minecraft:lily_of_the_valley": true,
	"minecraft:pink_petals": true, "minecraft:sweet_berry_bush": true,
}

var doublePlants = map[string]bool{
	"minecraft:tall_grass": true, "minecraft:large_fern": true, "minecraft:sunflower": true,
	"minecraft:lilac": true, "minecraft:rose_bush": true, "minecraft:peony": true,
	"minecraft:small_dripleaf": true,
}

func validateVegetationState(state SurfaceBlockState) error {
	switch state.Name {
	case "minecraft:fire", "minecraft:soul_fire", "minecraft:crimson_roots", "minecraft:warped_roots", "minecraft:crimson_fungus", "minecraft:warped_fungus", "minecraft:nether_sprouts":
		return nil
	}
	if state.Name == "minecraft:melon" {
		return nil
	}
	if state.Name == "minecraft:seagrass" {
		return nil
	}
	if state.Name == "minecraft:brown_mushroom" || state.Name == "minecraft:red_mushroom" {
		return nil
	}
	if state.Name == "minecraft:azalea" || state.Name == "minecraft:flowering_azalea" || state.Name == "minecraft:moss_carpet" || state.Name == "minecraft:small_dripleaf" || state.Name == "minecraft:hanging_roots" {
		return nil
	}
	if !groundPlants[state.Name] && state.Name != "minecraft:dead_bush" && state.Name != "minecraft:pumpkin" && state.Name != "minecraft:cactus" && state.Name != "minecraft:sugar_cane" && state.Name != "minecraft:lily_pad" && state.Name != "minecraft:spore_blossom" {
		return fmt.Errorf("unsupported simple-block survival for %s", state.Name)
	}
	if doublePlants[state.Name] && state.Properties["half"] != "lower" {
		return fmt.Errorf("unsupported double-plant half for %s", state.Name)
	}
	return nil
}

// SupportedPlant/DoublePlant expose the same ground rules to block updates.
func SupportedPlant(name string) bool { return groundPlants[name] || name == "minecraft:dead_bush" }
func DoublePlant(name string) bool    { return doublePlants[name] }
func PlantGroundSurvives(name, ground string) bool {
	if name == "minecraft:dead_bush" {
		return vegetationTags["dead_bush_may_place_on"][ground]
	}
	return vegetationTags["dirt"][ground] || ground == "minecraft:farmland"
}

func placeSimpleVegetation(world FeatureBlockWorld, random *WorldgenRandom, origin FeaturePosition, provider featureStateProvider) (bool, error) {
	state := provider(random, origin)
	survives, err := featurePlantSurvives(world, origin, state)
	if err != nil {
		return false, err
	}
	if !survives {
		return false, nil
	}
	if doublePlants[state.Name] {
		above := treeOffset(origin, 0, 1, 0)
		top, err := world.BlockAt(above)
		if err != nil {
			return false, err
		}
		if !surfaceAir(top) {
			return false, nil
		}
		lower, upper := stateWithProperty(state, "half", "lower"), stateWithProperty(state, "half", "upper")
		if state.Name == "minecraft:small_dripleaf" {
			current, err := world.BlockAt(origin)
			if err != nil {
				return false, err
			}
			lower = stateWithProperty(lower, "waterlogged", fmt.Sprint(hasSourceWater(current)))
			upper = stateWithProperty(upper, "waterlogged", fmt.Sprint(hasSourceWater(top)))
		}
		if err := setFeatureBlock(world, origin, lower); err != nil {
			return false, err
		}
		if err := setFeatureBlock(world, above, upper); err != nil {
			return false, err
		}
	} else if err := setFeatureBlock(world, origin, state); err != nil {
		return false, err
	}
	return true, nil
}

func compileVegetationPlan(definition ConfiguredFeatureDefinition, depth int) (configuredFeaturePlacer, error) {
	switch definition.Type {
	case "minecraft:simple_block":
		var config struct {
			Provider json.RawMessage `json:"to_place"`
		}
		if err := json.Unmarshal(definition.Config, &config); err != nil {
			return nil, err
		}
		provider, err := decodeFeatureStateProvider(config.Provider, validateVegetationState)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			return placeSimpleVegetation(region, random, p, provider)
		}, nil
	case "minecraft:random_patch", "minecraft:flower", "minecraft:no_bonemeal_flower":
		var config struct {
			Tries    int32           `json:"tries"`
			XZSpread int32           `json:"xz_spread"`
			YSpread  int32           `json:"y_spread"`
			Feature  json.RawMessage `json:"feature"`
		}
		if err := json.Unmarshal(definition.Config, &config); err != nil {
			return nil, err
		}
		if config.Tries < 0 || config.Tries > 4096 || config.XZSpread < 0 || config.XZSpread > 16 || config.YSpread < 0 || config.YSpread > 64 {
			return nil, fmt.Errorf("unsupported random patch dimensions")
		}
		place, err := compileNestedFeaturePlacement(config.Feature, depth+1)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, random *WorldgenRandom, p FeaturePosition) (bool, error) {
			success := false
			for range config.Tries {
				// Even a zero spread consumes two bounded draws, as in the JAR.
				dx := random.NextIntBound(config.XZSpread+1) - random.NextIntBound(config.XZSpread+1)
				dy := random.NextIntBound(config.YSpread+1) - random.NextIntBound(config.YSpread+1)
				dz := random.NextIntBound(config.XZSpread+1) - random.NextIntBound(config.XZSpread+1)
				placed, err := place(region, random, treeOffset(p, dx, dy, dz))
				if err != nil {
					return false, err
				}
				success = success || placed
			}
			return success, nil
		}, nil
	default:
		return nil, fmt.Errorf("unsupported vegetation feature %s", definition.Type)
	}
}
