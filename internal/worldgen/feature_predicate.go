package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"encoding/json"
	"fmt"
	"strings"
)

type featurePredicate func(FeatureBlockWorld, FeaturePosition) (bool, error)

func hasSourceWater(state SurfaceBlockState) bool {
	return state.Properties["waterlogged"] == "true" || (state.Name == "minecraft:water" && (state.Properties["level"] == "" || state.Properties["level"] == "0"))
}

func compileFeaturePredicate(raw json.RawMessage) (featurePredicate, error) {
	return decodeFeaturePredicate(raw, 0)
}

func decodeFeaturePredicate(raw json.RawMessage, depth int) (featurePredicate, error) {
	if depth > 32 {
		return nil, fmt.Errorf("block predicate nesting is too deep")
	}
	var config struct {
		Type       string            `json:"type"`
		Offset     []int32           `json:"offset"`
		Blocks     json.RawMessage   `json:"blocks"`
		Fluids     json.RawMessage   `json:"fluids"`
		State      SurfaceBlockState `json:"state"`
		Predicate  json.RawMessage   `json:"predicate"`
		Predicates []json.RawMessage `json:"predicates"`
		Direction  string            `json:"direction"`
		Tag        string            `json:"tag"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	offset := FeaturePosition{}
	if len(config.Offset) != 0 {
		if len(config.Offset) != 3 {
			return nil, fmt.Errorf("invalid block predicate offset")
		}
		offset = FeaturePosition{config.Offset[0], config.Offset[1], config.Offset[2]}
	}
	switch config.Type {
	case "minecraft:replaceable":
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			s, err := world.BlockAt(treeOffset(p, offset.X, offset.Y, offset.Z))
			if err != nil {
				return false, err
			}
			id, err := blockstate.BlockStateID(s.Name, s.Properties)
			if err != nil {
				return false, err
			}
			rule, err := blockrules.Lookup(id)
			return rule.Replaceable, err
		}, nil
	case "minecraft:inside_world_bounds":
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			q := treeOffset(p, offset.X, offset.Y, offset.Z)
			if bounds,ok:=world.(interface{MinBuildHeight() int32;MaxBuildHeight() int32});ok {
				return q.Y>=bounds.MinBuildHeight() && q.Y<bounds.MaxBuildHeight(),nil
			}
			return q.Y >= -64 && q.Y < 320, nil
		}, nil
	case "minecraft:matching_block_tag":
		tag := strings.TrimPrefix(config.Tag, "minecraft:")
		blocks := vegetationTags[tag]
		if blocks == nil {
			blocks = treeTags[tag]
		}
		if blocks == nil {
			return nil, fmt.Errorf("unsupported block predicate tag %s", tag)
		}
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			s, err := world.BlockAt(treeOffset(p, offset.X, offset.Y, offset.Z))
			return blocks[s.Name], err
		}, nil
	case "minecraft:has_sturdy_face":
		direction := -1
		for i, name := range directionNames {
			if name == config.Direction {
				direction = i
			}
		}
		if direction < 0 {
			return nil, fmt.Errorf("invalid sturdy face direction")
		}
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			e := featureEditor{world: world}
			ok := featureSturdy(&e, treeOffset(p, offset.X, offset.Y, offset.Z), direction)
			return ok, e.err
		}, nil
	case "minecraft:true":
		return func(FeatureBlockWorld, FeaturePosition) (bool, error) { return true, nil }, nil
	case "minecraft:not":
		predicate, err := decodeFeaturePredicate(config.Predicate, depth+1)
		if err != nil {
			return nil, err
		}
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			ok, err := predicate(world, p)
			return !ok, err
		}, nil
	case "minecraft:all_of", "minecraft:any_of":
		var predicates []featurePredicate
		for _, entry := range config.Predicates {
			predicate, err := decodeFeaturePredicate(entry, depth+1)
			if err != nil {
				return nil, err
			}
			predicates = append(predicates, predicate)
		}
		all := config.Type == "minecraft:all_of"
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			for _, predicate := range predicates {
				ok, err := predicate(world, p)
				if err != nil {
					return false, err
				}
				if ok != all {
					return ok, nil
				}
			}
			return all, nil
		}, nil
	case "minecraft:matching_blocks", "minecraft:matching_fluids":
		values := config.Blocks
		if config.Type == "minecraft:matching_fluids" {
			values = config.Fluids
		}
		var single string
		var names []string
		if json.Unmarshal(values, &single) == nil {
			names = []string{single}
		} else if err := json.Unmarshal(values, &names); err != nil {
			return nil, err
		}
		matching := make(map[string]bool)
		for _, name := range names {
			matching[name] = true
		}
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			state, err := world.BlockAt(treeOffset(p, offset.X, offset.Y, offset.Z))
			if err != nil {
				return false, err
			}
			if config.Type == "minecraft:matching_fluids" {
				name := "minecraft:empty"
				if hasSourceWater(state) {
					name = "minecraft:water"
				} else if state.Name == "minecraft:water" {
					name = "minecraft:flowing_water"
				}
				if state.Name == "minecraft:lava" {
					name = "minecraft:lava"
					if state.Properties["level"] != "" && state.Properties["level"] != "0" {
						name = "minecraft:flowing_lava"
					}
				}
				return matching[name], nil
			}
			return matching[state.Name], nil
		}, nil
	case "minecraft:would_survive":
		if !isFeatureSapling(config.State.Name) && config.State.Name != "minecraft:mangrove_propagule" && config.State.Name != "minecraft:cactus" && config.State.Name != "minecraft:sugar_cane" {
			return nil, fmt.Errorf("unsupported survival predicate for %s", config.State.Name)
		}
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			return featurePlantSurvives(world, treeOffset(p, offset.X, offset.Y, offset.Z), config.State)
		}, nil
	case "minecraft:solid":
		return func(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
			state, err := world.BlockAt(treeOffset(p, offset.X, offset.Y, offset.Z))
			if err != nil {
				return false, err
			}
			return featureSolid(state)
		}, nil
	default:
		return nil, fmt.Errorf("unsupported block predicate %s", config.Type)
	}
}
