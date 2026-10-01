package worldgen

import (
	"atlas/internal/blockstate"
	"encoding/json"
	"fmt"
)

type DiskConfiguration struct {
	HalfHeight int32
	MinRadius  int32
	MaxRadius  int32
	State      SurfaceBlockState
	Targets    map[string]bool
	provider   func(FeatureBlockWorld, *WorldgenRandom, FeaturePosition) (SurfaceBlockState, error)
}

func LoadDiskConfiguration(name string) (DiskConfiguration, error) {
	definition, err := LoadConfiguredFeatureDefinition(name)
	if err != nil {
		return DiskConfiguration{}, err
	}
	if definition.Type != "minecraft:disk" {
		return DiskConfiguration{}, fmt.Errorf("configured feature %s is %s, not minecraft:disk", name, definition.Type)
	}
	var raw struct {
		HalfHeight int32 `json:"half_height"`
		Radius     struct {
			Type         string `json:"type"`
			MinInclusive int32  `json:"min_inclusive"`
			MaxInclusive int32  `json:"max_inclusive"`
		} `json:"radius"`
		Target struct {
			Type   string   `json:"type"`
			Blocks []string `json:"blocks"`
		} `json:"target"`
		StateProvider struct {
			Fallback struct {
				Type  string            `json:"type"`
				State SurfaceBlockState `json:"state"`
			} `json:"fallback"`
			Rules []struct {
				IfTrue json.RawMessage `json:"if_true"`
				Then   json.RawMessage `json:"then"`
			} `json:"rules"`
		} `json:"state_provider"`
	}
	if err := json.Unmarshal(definition.Config, &raw); err != nil {
		return DiskConfiguration{}, err
	}
	if raw.HalfHeight < 0 || raw.HalfHeight > 32 || raw.Radius.Type != "minecraft:uniform" ||
		raw.Radius.MinInclusive < 0 || raw.Radius.MaxInclusive > 32 ||
		raw.Radius.MinInclusive > raw.Radius.MaxInclusive || raw.Target.Type != "minecraft:matching_blocks" ||
		len(raw.Target.Blocks) == 0 || raw.StateProvider.Fallback.Type != "minecraft:simple_state_provider" ||
		raw.StateProvider.Fallback.State.Name == "" {
		return DiskConfiguration{}, fmt.Errorf("unsupported disk configuration %s", name)
	}
	configuration := DiskConfiguration{
		HalfHeight: raw.HalfHeight, MinRadius: raw.Radius.MinInclusive,
		MaxRadius: raw.Radius.MaxInclusive, State: raw.StateProvider.Fallback.State,
		Targets: make(map[string]bool, len(raw.Target.Blocks)),
	}
	for _, block := range raw.Target.Blocks {
		configuration.Targets[block] = true
	}
	var predicates []featurePredicate
	var providers []featureStateProvider
	for _, rule := range raw.StateProvider.Rules {
		predicate, err := compileFeaturePredicate(rule.IfTrue)
		if err != nil {
			return DiskConfiguration{}, err
		}
		provider, err := decodeFeatureStateProvider(rule.Then, func(state SurfaceBlockState) error {
			_, err := blockstate.BlockStateID(state.Name, state.Properties)
			return err
		})
		if err != nil {
			return DiskConfiguration{}, err
		}
		predicates = append(predicates, predicate)
		providers = append(providers, provider)
	}
	configuration.provider = func(world FeatureBlockWorld, random *WorldgenRandom, p FeaturePosition) (SurfaceBlockState, error) {
		for i, predicate := range predicates {
			ok, err := predicate(world, p)
			if err != nil {
				return SurfaceBlockState{}, err
			}
			if ok {
				return providers[i](random, p), nil
			}
		}
		return configuration.State, nil
	}
	return configuration, nil
}

func PlaceDiskFeature(world FeatureBlockWorld, random *WorldgenRandom, origin FeaturePosition,
	configuration DiskConfiguration) (bool, error) {
	if world == nil || random == nil || configuration.HalfHeight < 0 ||
		configuration.MinRadius < 0 || configuration.MaxRadius < configuration.MinRadius ||
		configuration.MaxRadius > 32 || configuration.State.Name == "" || len(configuration.Targets) == 0 {
		return false, fmt.Errorf("disk placement requires a world, random source, and valid configuration")
	}
	radius := randomBetweenInclusive(random, configuration.MinRadius, configuration.MaxRadius)
	placed := false
	for blockZ := origin.Z - radius; blockZ <= origin.Z+radius; blockZ++ {
		for blockX := origin.X - radius; blockX <= origin.X+radius; blockX++ {
			deltaX, deltaZ := blockX-origin.X, blockZ-origin.Z
			if deltaX*deltaX+deltaZ*deltaZ > radius*radius {
				continue
			}
			for blockY := origin.Y + configuration.HalfHeight; blockY >= origin.Y-configuration.HalfHeight; blockY-- {
				position := FeaturePosition{X: blockX, Y: blockY, Z: blockZ}
				if !world.InWorld(position) {
					continue
				}
				state, err := world.BlockAt(position)
				if err != nil {
					return placed, err
				}
				if !configuration.Targets[state.Name] {
					continue
				}
				state = configuration.State
				if configuration.provider != nil {
					state, err = configuration.provider(world, random, position)
					if err != nil {
						return placed, err
					}
				}
				if err := world.SetBlock(position, state); err != nil {
					return placed, err
				}
				if marker, ok := world.(interface{ MarkPostProcessing(FeaturePosition) error }); ok {
					for dy := int32(1); dy <= 2; dy++ {
						above := treeOffset(position, 0, dy, 0)
						state, err := world.BlockAt(above)
						if err != nil {
							return placed, err
						}
						if surfaceAir(state) {
							break
						}
						if err = marker.MarkPostProcessing(above); err != nil {
							return placed, err
						}
					}
				}
				placed = true
			}
		}
	}
	return placed, nil
}
