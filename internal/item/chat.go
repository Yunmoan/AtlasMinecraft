package item

import (
	"atlas/internal/nbt"
	"encoding/json"
	"errors"
	"math"
)

// Save codecs use JSON strings for chat components; the protocol uses NBT.
// Keep this conversion separate from custom_data, which is ordinary NBT.
func chatJSON(tag nbt.Tag) any {
	if o, ok := tag.Object(); ok {
		out := map[string]any{}
		for key, value := range o {
			out[key] = chatJSON(value)
			switch key {
			case "bold", "italic", "underlined", "strikethrough", "obfuscated", "interpret":
				if n, ok := value.Number(); ok {
					out[key] = n != 0
				}
			}
		}
		return out
	}
	if values, ok := tag.Elements(); ok {
		out := make([]any, len(values))
		for i, value := range values {
			out[i] = chatJSON(value)
		}
		return out
	}
	return tag.Value
}
func savedChatComponent(tag nbt.Tag) (nbt.Tag, error) {
	value, ok := tag.Text()
	if !ok {
		return nbt.Tag{}, errors.New("saved chat component is not a JSON string")
	}
	var component any
	if err := json.Unmarshal([]byte(value), &component); err != nil {
		return nbt.Tag{}, err
	}
	return chatNBT(component)
}
func chatNBT(value any) (nbt.Tag, error) {
	switch v := value.(type) {
	case string:
		return text(v), nil
	case bool:
		return boolean(v), nil
	case float64:
		if math.Trunc(v) == v && v >= math.MinInt32 && v <= math.MaxInt32 {
			return integer(int32(v)), nil
		}
		return nbt.Tag{Type: nbt.Double, Value: v}, nil
	case map[string]any:
		o := nbt.Object{}
		for key, value := range v {
			tag, err := chatNBT(value)
			if err != nil {
				return nbt.Tag{}, err
			}
			o[key] = tag
		}
		return compound(o), nil
	case []any:
		entries := []nbt.Tag{}
		kind := byte(nbt.End)
		for _, value := range v {
			tag, err := chatNBT(value)
			if err != nil {
				return nbt.Tag{}, err
			}
			if kind == nbt.End {
				kind = tag.Type
			}
			if kind != tag.Type {
				kind = nbt.Compound
			}
			entries = append(entries, tag)
		}
		if kind == nbt.Compound {
			for i, tag := range entries {
				if tag.Type == nbt.String {
					entries[i] = compound(nbt.Object{"text": tag})
				} else if tag.Type != nbt.Compound {
					return nbt.Tag{}, errors.New("mixed non-component chat list")
				}
			}
		}
		return nbt.Tag{Type: nbt.List, ListType: kind, Value: entries}, nil
	}
	return nbt.Tag{}, errors.New("invalid JSON chat component")
}
