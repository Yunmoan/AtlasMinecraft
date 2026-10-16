package worldgen

import (
	"encoding/json"
	"fmt"
)

type jigsawAliasBinding struct {
	Type    string `json:"type"`
	Alias   string `json:"alias"`
	Target  string `json:"target"`
	Targets []struct {
		Data   string `json:"data"`
		Weight int32  `json:"weight"`
	} `json:"targets"`
	Groups []struct {
		Data   []jigsawAliasBinding `json:"data"`
		Weight int32                `json:"weight"`
	} `json:"groups"`
}

func buildJigsawAliases(seed int64, position FeaturePosition, raw []json.RawMessage) (map[string]string, error) {
	aliases := make(map[string]string)
	if len(raw) == 0 {
		return aliases, nil
	}
	// PoolAliasLookup owns an independent Legacy positional stream derived
	// from the world seed and the original start height, before projection.
	random := NewLegacyRandom(seed).ForkPositional().At(position.X, position.Y, position.Z)
	var bind func(jigsawAliasBinding) error
	store := func(alias, target string) error {
		alias = densityName(alias)
		if _, ok := aliases[alias]; ok {
			return fmt.Errorf("duplicate Jigsaw pool alias %s", alias)
		}
		aliases[alias] = densityName(target)
		return nil
	}
	bind = func(binding jigsawAliasBinding) error {
		switch binding.Type {
		case "minecraft:direct":
			return store(binding.Alias, binding.Target)
		case "minecraft:random":
			var total int32
			for _, entry := range binding.Targets {
				if entry.Weight <= 0 {
					return fmt.Errorf("invalid Jigsaw alias weight")
				}
				total += entry.Weight
			}
			if total <= 0 {
				return nil
			}
			draw := random.NextIntBound(total)
			for _, entry := range binding.Targets {
				draw -= entry.Weight
				if draw < 0 {
					return store(binding.Alias, entry.Data)
				}
			}
		case "minecraft:random_group":
			var total int32
			for _, entry := range binding.Groups {
				if entry.Weight <= 0 {
					return fmt.Errorf("invalid Jigsaw alias group weight")
				}
				total += entry.Weight
			}
			if total <= 0 {
				return nil
			}
			draw := random.NextIntBound(total)
			for _, entry := range binding.Groups {
				draw -= entry.Weight
				if draw < 0 {
					for _, child := range entry.Data {
						if err := bind(child); err != nil {
							return err
						}
					}
					return nil
				}
			}
		default:
			return fmt.Errorf("unsupported Jigsaw alias type %q", binding.Type)
		}
		return nil
	}
	for _, data := range raw {
		var binding jigsawAliasBinding
		if err := json.Unmarshal(data, &binding); err != nil {
			return nil, err
		}
		if err := bind(binding); err != nil {
			return nil, err
		}
	}
	return aliases, nil
}
