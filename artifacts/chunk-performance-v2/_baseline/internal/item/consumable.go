package item

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"errors"
	"fmt"
	"math"
	"strings"
)

func tagList(entries []nbt.Tag) nbt.Tag {
	return nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: entries}
}

func effectDetailsNBT(v map[string]any) nbt.Object {
	o := nbt.Object{"amplifier": integer(v["amplifier"].(int32)), "duration": integer(v["duration"].(int32)),
		"ambient": boolean(v["ambient"].(bool)), "show_particles": boolean(v["showParticles"].(bool)), "show_icon": boolean(v["showIcon"].(bool))}
	if hidden, ok := v["hiddenEffect"].(map[string]any); ok {
		o["hidden_effect"] = compound(effectDetailsNBT(hidden))
	}
	return o
}

func effectNBT(v map[string]any) (nbt.Tag, bool) {
	name, ok := registryName("mob_effect", v["id"].(int32))
	if !ok {
		return nbt.Tag{}, false
	}
	o := effectDetailsNBT(v["details"].(map[string]any))
	o["id"] = text(name)
	return compound(o), true
}

func consumableNBT(name string, v map[string]any) (nbt.Tag, bool) {
	o := nbt.Object{}
	switch name {
	case "potion_contents":
		if id, ok := v["potionId"].(int32); ok {
			name, ok := registryName("potion", id)
			if !ok {
				return nbt.Tag{}, false
			}
			o["potion"] = text(name)
		}
		if color, ok := v["customColor"].(int32); ok {
			o["custom_color"] = integer(color)
		}
		var effects []nbt.Tag
		for _, value := range v["customEffects"].([]any) {
			tag, ok := effectNBT(value.(map[string]any))
			if !ok {
				return nbt.Tag{}, false
			}
			effects = append(effects, tag)
		}
		o["custom_effects"] = tagList(effects)
	case "food":
		o["nutrition"] = integer(v["nutrition"].(int32))
		o["saturation"] = nbt.Tag{Type: nbt.Float, Value: v["saturationModifier"].(float32)}
		o["can_always_eat"] = boolean(v["canAlwaysEat"].(bool))
		o["eat_seconds"] = nbt.Tag{Type: nbt.Float, Value: v["secondsToEat"].(float32)}
		if stack, ok := v["usingConvertsTo"].(Stack); ok {
			tag, err := stack.ToNBT()
			if err != nil {
				return nbt.Tag{}, false
			}
			o["using_converts_to"] = tag
		}
		var effects []nbt.Tag
		for _, value := range v["effects"].([]any) {
			effect := value.(map[string]any)
			tag, ok := effectNBT(effect["effect"].(map[string]any))
			if !ok {
				return nbt.Tag{}, false
			}
			effects = append(effects, compound(nbt.Object{"effect": tag, "probability": {Type: nbt.Float, Value: effect["probability"].(float32)}}))
		}
		o["effects"] = tagList(effects)
	case "suspicious_stew_effects":
		var effects []nbt.Tag
		for _, value := range v["effects"].([]any) {
			effect := value.(map[string]any)
			name, ok := registryName("mob_effect", effect["effect"].(int32))
			if !ok {
				return nbt.Tag{}, false
			}
			effects = append(effects, compound(nbt.Object{"id": text(name), "duration": integer(effect["duration"].(int32))}))
		}
		return tagList(effects), true
	}
	return compound(o), true
}

func savedRegistryID(reg, name string) (int32, error) {
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	id, ok := registryID(reg, name)
	if !ok {
		return 0, fmt.Errorf("unknown %s %q", reg, name)
	}
	return id, nil
}

func savedInt(o nbt.Object, name string, fallback int32) int32 {
	if value, ok := o.Number(name); ok {
		return int32(value)
	}
	return fallback
}

func savedFloat(o nbt.Object, name string, fallback float32) (float32, error) {
	value := fallback
	if tag, ok := o[name]; ok {
		switch v := tag.Value.(type) {
		case float32:
			value = v
		case float64:
			value = float32(v)
		default:
			n, ok := tag.Number()
			if !ok {
				return 0, errors.New("invalid float " + name)
			}
			value = float32(n)
		}
	}
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return 0, errors.New("invalid float " + name)
	}
	return value, nil
}

func savedList(o nbt.Object, key string) ([]nbt.Tag, error) {
	tag, exists := o[key]
	if !exists {
		return nil, nil
	}
	entries, ok := tag.Elements()
	if !ok || len(entries) > 4096 {
		return nil, errors.New("invalid list " + key)
	}
	return entries, nil
}

func writeEffectDetails(w *proto.Writer, o nbt.Object, depth int) error {
	if depth > 32 {
		return errors.New("saved effect nesting limit exceeded")
	}
	particles := savedInt(o, "show_particles", 1) != 0
	icon := particles
	if v, ok := o.Number("show_icon"); ok {
		icon = v != 0
	}
	w.VarInt(savedInt(o, "amplifier", 0)).VarInt(savedInt(o, "duration", 0)).Bool(savedInt(o, "ambient", 0) != 0).Bool(particles).Bool(icon)
	hidden, exists := o["hidden_effect"]
	w.Bool(exists)
	if exists {
		nested, ok := hidden.Object()
		if !ok {
			return errors.New("invalid hidden effect")
		}
		return writeEffectDetails(w, nested, depth+1)
	}
	return nil
}

func writeEffect(w *proto.Writer, tag nbt.Tag) error {
	o, ok := tag.Object()
	if !ok {
		return errors.New("invalid effect")
	}
	name, _ := o.Text("id")
	id, err := savedRegistryID("mob_effect", name)
	if err != nil {
		return err
	}
	w.VarInt(id)
	return writeEffectDetails(w, o, 0)
}

func consumableBytes(name string, tag nbt.Tag) ([]byte, error) {
	w := proto.NewWriter()
	if name == "suspicious_stew_effects" {
		entries, ok := tag.Elements()
		if !ok || len(entries) > 4096 {
			return nil, errors.New("invalid stew effects")
		}
		w.VarInt(int32(len(entries)))
		for _, entry := range entries {
			o, ok := entry.Object()
			if !ok {
				return nil, errors.New("invalid stew effect")
			}
			name, _ := o.Text("id")
			id, err := savedRegistryID("mob_effect", name)
			if err != nil {
				return nil, err
			}
			w.VarInt(id).VarInt(savedInt(o, "duration", 160))
		}
		return w.Payload(), nil
	}
	o, ok := tag.Object()
	if name == "potion_contents" {
		// The persistent codec accepts a registry name as a shorthand.
		if potion, isString := tag.Text(); isString {
			o, ok = nbt.Object{"potion": text(potion)}, true
		}
	}
	if !ok {
		return nil, errors.New("invalid consumable component")
	}
	switch name {
	case "potion_contents":
		potion, exists := o["potion"]
		w.Bool(exists)
		if exists {
			name, ok := potion.Text()
			if !ok {
				return nil, errors.New("invalid potion name")
			}
			id, err := savedRegistryID("potion", name)
			if err != nil {
				return nil, err
			}
			w.VarInt(id)
		}
		color, exists := o["custom_color"]
		w.Bool(exists)
		if exists {
			value, ok := color.Number()
			if !ok {
				return nil, errors.New("invalid potion color")
			}
			w.I32(int32(value))
		}
		effects, err := savedList(o, "custom_effects")
		if err != nil {
			return nil, err
		}
		w.VarInt(int32(len(effects)))
		for _, effect := range effects {
			if err := writeEffect(w, effect); err != nil {
				return nil, err
			}
		}
	case "food":
		saturation, err := savedFloat(o, "saturation", 0)
		if err != nil {
			return nil, err
		}
		seconds, err := savedFloat(o, "eat_seconds", 1.6)
		if err != nil {
			return nil, err
		}
		w.VarInt(savedInt(o, "nutrition", 0)).F32(saturation).Bool(savedInt(o, "can_always_eat", 0) != 0).F32(seconds)
		converted, exists := o["using_converts_to"]
		w.Bool(exists)
		if exists {
			stack, err := FromNBT(converted)
			if err != nil {
				return nil, err
			}
			stack.Write(w)
		}
		effects, err := savedList(o, "effects")
		if err != nil {
			return nil, err
		}
		w.VarInt(int32(len(effects)))
		for _, effect := range effects {
			o, ok := effect.Object()
			if !ok {
				return nil, errors.New("invalid food effect")
			}
			if err := writeEffect(w, o["effect"]); err != nil {
				return nil, err
			}
			probability, err := savedFloat(o, "probability", 1)
			if err != nil {
				return nil, err
			}
			w.F32(probability)
		}
	}
	return w.Payload(), nil
}
