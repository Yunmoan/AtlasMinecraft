package item

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/registry"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

func compound(o nbt.Object) nbt.Tag { return nbt.Tag{Type: nbt.Compound, Value: o} }
func integer(v int32) nbt.Tag       { return nbt.Tag{Type: nbt.Int, Value: v} }
func text(v string) nbt.Tag         { return nbt.Tag{Type: nbt.String, Value: v} }
func boolean(v bool) nbt.Tag {
	var b int8
	if v {
		b = 1
	}
	return nbt.Tag{Type: nbt.Byte, Value: b}
}

// An exact network patch is retained alongside standard Java Inventory tags.
// Its fingerprint prevents stale Atlas bytes from overriding external Java edits.
// Common vanilla components are also exported in their standard NBT form.
func (s Stack) ToNBT() (nbt.Tag, error) {
	name, ok := Name(s.ID)
	if !ok || s.Count < 1 || s.Count > 99 {
		return nbt.Tag{}, errors.New("invalid saved item")
	}
	o := nbt.Object{}
	if len(s.Saved) != 0 {
		_, tag, err := nbt.Parse(s.Saved)
		if err != nil {
			return nbt.Tag{}, err
		}
		o, _ = tag.Object()
	}
	o["id"], o["count"] = text(name), integer(s.Count)
	if len(s.Saved) == 0 {
		values, removed, err := s.components()
		if err != nil {
			return nbt.Tag{}, err
		}
		components := nbt.Object{}
		for name, value := range values {
			if tag, ok := componentNBT(name, value); ok {
				components["minecraft:"+name] = tag
			}
		}
		for _, name := range removed {
			components["!minecraft:"+name] = compound(nbt.Object{})
		}
		if len(components) > 0 {
			o["components"] = compound(components)
		}
	}
	delete(o, "Slot")
	delete(o, "AtlasNetworkItem")
	delete(o, "AtlasItemFingerprint")
	fingerprint, err := itemFingerprint(o)
	if err != nil {
		return nbt.Tag{}, err
	}
	o["AtlasNetworkItem"] = nbt.Tag{Type: nbt.ByteArray, Value: s.Bytes()}
	o["AtlasItemFingerprint"] = text(fingerprint)
	return compound(o), nil
}

func itemFingerprint(o nbt.Object) (string, error) {
	// Normalize empty lists and concrete tag widths through NBT first; an
	// empty nil slice must hash identically after disk decoding as an empty list.
	encoded, err := nbt.Marshal("", compound(o))
	if err != nil {
		return "", err
	}
	_, tag, err := nbt.Parse(encoded)
	if err != nil {
		return "", err
	}
	normalized, _ := tag.Object()
	data, err := json.Marshal(normalized)
	return fmt.Sprintf("%x", sha256.Sum256(data)), err
}

func FromNBT(tag nbt.Tag) (Stack, error) {
	o, ok := tag.Object()
	if !ok {
		return Stack{}, errors.New("inventory entry is not a compound")
	}
	copyObject := nbt.Object{}
	for key, value := range o {
		if key != "Slot" && key != "AtlasNetworkItem" && key != "AtlasItemFingerprint" {
			copyObject[key] = value
		}
	}
	fingerprint, err := itemFingerprint(copyObject)
	if err != nil {
		return Stack{}, err
	}
	saved, err := nbt.Marshal("", compound(copyObject))
	if err != nil {
		return Stack{}, err
	}
	if raw, ok := o["AtlasNetworkItem"].Value.([]byte); ok {
		if expected, _ := o.Text("AtlasItemFingerprint"); expected == fingerprint {
			r := proto.NewReader(raw)
			s, err := Read(r)
			if err != nil || r.Remaining() != 0 {
				return Stack{}, errors.New("invalid saved network item")
			}
			s.Saved = saved
			return s, nil
		}
	}
	name, _ := o.Text("id")
	id, ok := ID(name)
	count, countOK := o.Number("count")
	if !ok || !countOK || count < 1 || count > 99 {
		return Stack{}, errors.New("invalid vanilla inventory item")
	}
	components, _ := o.Object("components")
	var added []int32
	var removed []int32
	data := map[int32][]byte{}
	for key, tag := range components {
		remove := strings.HasPrefix(key, "!")
		name := strings.TrimPrefix(strings.TrimPrefix(key, "!"), "minecraft:")
		id, ok := componentID(name)
		if !ok {
			return Stack{}, fmt.Errorf("unknown saved component %s", key)
		}
		if remove {
			removed = append(removed, id)
			continue
		}
		encoded, err := componentBytes(name, tag)
		if err != nil {
			return Stack{}, fmt.Errorf("saved component %s: %w", key, err)
		}
		added = append(added, id)
		data[id] = encoded
	}
	sort.Slice(added, func(i, j int) bool { return added[i] < added[j] })
	sort.Slice(removed, func(i, j int) bool { return removed[i] < removed[j] })
	w := proto.NewWriter().VarInt(int32(len(added))).VarInt(int32(len(removed)))
	for _, id := range added {
		w.VarInt(id).Raw(data[id])
	}
	for _, id := range removed {
		w.VarInt(id)
	}
	s := Stack{ID: id, Count: int32(count), Patch: w.Payload(), Saved: saved}
	// Validate imported component encodings through the same network reader.
	r := proto.NewReader(s.Bytes())
	checked, err := Read(r)
	if err != nil || r.Remaining() != 0 {
		return Stack{}, errors.New("invalid imported item components")
	}
	checked.Saved = saved
	return checked, nil
}

func componentID(name string) (int32, bool) {
	for i := int32(0); i < 57; i++ {
		if componentNames[fmt.Sprint(i)] == name {
			return i, true
		}
	}
	return 0, false
}

var registryOnce sync.Once
var registryNames = map[string][]string{}

//go:embed builtin_registries_767.json
var builtinRegistryData []byte

func loadRegistryNames() {
	registryOnce.Do(func() {
		if err := json.Unmarshal(builtinRegistryData, &registryNames); err != nil {
			panic(err)
		}
		for _, payload := range registry.Packets() {
			r := proto.NewReader(payload)
			name, err := r.String(32767)
			if err != nil {
				panic(err)
			}
			count, err := r.VarInt()
			if err != nil {
				panic(err)
			}
			for range count {
				entry, err := r.String(32767)
				if err != nil {
					panic(err)
				}
				registryNames[name] = append(registryNames[name], entry)
				present, err := r.Bool()
				if err != nil {
					panic(err)
				}
				if present {
					_, n, err := nbt.ParseNetwork(r.Rest())
					if err != nil {
						panic(err)
					}
					r.Bytes(n)
				}
			}
		}
	})
}
func registryName(reg string, id int32) (string, bool) {
	loadRegistryNames()
	names := registryNames["minecraft:"+reg]
	if id < 0 || int(id) >= len(names) {
		return "", false
	}
	return names[id], true
}
func registryID(reg, name string) (int32, bool) {
	loadRegistryNames()
	for i, v := range registryNames["minecraft:"+reg] {
		if v == name {
			return int32(i), true
		}
	}
	return 0, false
}

func componentNBT(name string, value any) (nbt.Tag, bool) {
	if name == "custom_name" || name == "item_name" {
		tag, ok := value.(nbt.Tag)
		if !ok {
			return nbt.Tag{}, false
		}
		data, err := json.Marshal(chatJSON(tag))
		if err != nil {
			return nbt.Tag{}, false
		}
		return text(string(data)), true
	}
	switch v := value.(type) {
	case nbt.Tag:
		return v, true
	case int32:
		if name == "base_color" {
			colors := []string{"white", "orange", "magenta", "light_blue", "yellow", "lime", "pink", "gray", "light_gray", "cyan", "purple", "blue", "brown", "green", "red", "black"}
			if v < 0 || int(v) >= len(colors) {
				return nbt.Tag{}, false
			}
			return text(colors[v]), true
		}
		if name == "map_post_processing" {
			if v == 0 {
				return text("lock"), true
			}
			if v == 1 {
				return text("scale"), true
			}
			return nbt.Tag{}, false
		}
		return integer(v), true
	case string:
		return text(v), true
	case bool:
		if name == "unbreakable" {
			return compound(nbt.Object{"show_in_tooltip": boolean(v)}), true
		}
		return boolean(v), true
	case nil:
		return compound(nbt.Object{}), true
	}
	if name == "lore" {
		var lines []nbt.Tag
		for _, value := range value.([]any) {
			tag := value.(nbt.Tag)
			data, err := json.Marshal(chatJSON(tag))
			if err != nil {
				return nbt.Tag{}, false
			}
			lines = append(lines, text(string(data)))
		}
		return nbt.Tag{Type: nbt.List, ListType: nbt.String, Value: lines}, true
	}
	v, ok := value.(map[string]any)
	if !ok {
		return nbt.Tag{}, false
	}
	switch name {
	case "potion_contents", "food", "suspicious_stew_effects":
		return consumableNBT(name, v)
	case "enchantments", "stored_enchantments":
		levels := nbt.Object{}
		for _, entry := range v["enchantments"].([]any) {
			e := entry.(map[string]any)
			name, ok := registryName("enchantment", e["id"].(int32))
			if !ok {
				return nbt.Tag{}, false
			}
			levels[name] = integer(e["level"].(int32))
		}
		tooltip := v["showTooltip"]
		if tooltip == nil {
			tooltip = v["showInTooltip"]
		}
		return compound(nbt.Object{"levels": compound(levels), "show_in_tooltip": boolean(tooltip.(bool))}), true
	case "dyed_color":
		return compound(nbt.Object{"rgb": integer(v["color"].(int32)), "show_in_tooltip": boolean(v["showTooltip"].(bool))}), true
	case "block_state":
		o := nbt.Object{}
		for _, entry := range v["properties"].([]any) {
			e := entry.(map[string]any)
			o[e["property"].(string)] = text(e["value"].(string))
		}
		return compound(o), true
	case "container", "bundle_contents", "charged_projectiles":
		key := "contents"
		if name == "charged_projectiles" {
			key = "projectiles"
		}
		var entries []nbt.Tag
		for i, entry := range v[key].([]any) {
			s := entry.(Stack)
			if s.Empty() {
				continue
			}
			tag, err := s.ToNBT()
			if err != nil {
				return nbt.Tag{}, false
			}
			if name == "container" {
				tag = compound(nbt.Object{"slot": integer(int32(i)), "item": tag})
			}
			entries = append(entries, tag)
		}
		return nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: entries}, true
	}
	return nbt.Tag{}, false
}

func componentBytes(name string, tag nbt.Tag) ([]byte, error) {
	w := proto.NewWriter()
	switch name {
	case "potion_contents", "food", "suspicious_stew_effects":
		return consumableBytes(name, tag)
	case "custom_name", "item_name":
		component, err := savedChatComponent(tag)
		if err != nil {
			return nil, err
		}
		return nbt.MarshalNetwork(component)
	case "custom_data", "intangible_projectile", "map_decorations", "debug_stick_state", "entity_data", "bucket_entity_data", "block_entity_data", "recipes", "lock", "container_loot":
		return nbt.MarshalNetwork(tag)
	case "max_stack_size", "max_damage", "damage", "custom_model_data", "repair_cost", "map_id", "ominous_bottle_amplifier":
		v, ok := tag.Number()
		if !ok {
			return nil, errors.New("expected integer")
		}
		w.VarInt(int32(v))
	case "base_color":
		value, ok := tag.Text()
		if !ok {
			return nil, errors.New("invalid base color")
		}
		found := false
		for i, color := range []string{"white", "orange", "magenta", "light_blue", "yellow", "lime", "pink", "gray", "light_gray", "cyan", "purple", "blue", "brown", "green", "red", "black"} {
			if value == color {
				w.VarInt(int32(i))
				found = true
			}
		}
		if !found {
			return nil, errors.New("unknown base color")
		}
	case "map_post_processing":
		value, ok := tag.Text()
		if !ok {
			return nil, errors.New("invalid map processing")
		}
		switch value {
		case "lock":
			w.VarInt(0)
		case "scale":
			w.VarInt(1)
		default:
			return nil, errors.New("unknown map processing")
		}
	case "map_color":
		v, ok := tag.Number()
		if !ok {
			return nil, errors.New("expected color")
		}
		w.I32(int32(v))
	case "hide_additional_tooltip", "hide_tooltip", "creative_slot_lock", "fire_resistant":
	case "enchantment_glint_override":
		v, _ := tag.Number()
		w.Bool(v != 0)
	case "unbreakable":
		o, ok := tag.Object()
		if !ok {
			return nil, errors.New("expected unbreakable object")
		}
		v, ok := o.Number("show_in_tooltip")
		w.Bool(!ok || v != 0)
	case "rarity":
		v, _ := tag.Text()
		values := []string{"common", "uncommon", "rare", "epic"}
		found := false
		for i, name := range values {
			if name == v {
				w.VarInt(int32(i))
				found = true
			}
		}
		if !found {
			return nil, errors.New("invalid rarity")
		}
	case "note_block_sound":
		v, ok := tag.Text()
		if !ok {
			return nil, errors.New("invalid sound")
		}
		w.String(v)
	case "lore":
		lines, ok := tag.Elements()
		if !ok {
			return nil, errors.New("invalid lore")
		}
		w.VarInt(int32(len(lines)))
		for _, line := range lines {
			component, err := savedChatComponent(line)
			if err != nil {
				return nil, err
			}
			data, err := nbt.MarshalNetwork(component)
			if err != nil {
				return nil, err
			}
			w.Raw(data)
		}
	case "enchantments", "stored_enchantments":
		o, ok := tag.Object()
		if !ok {
			return nil, errors.New("invalid enchantment object")
		}
		levels, ok := o.Object("levels")
		if !ok {
			return nil, errors.New("missing enchantment levels")
		}
		w.VarInt(int32(len(levels)))
		var keys []string
		for key := range levels {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			id, ok := registryID("enchantment", key)
			if !ok {
				return nil, fmt.Errorf("unknown enchantment %s", key)
			}
			level, ok := levels[key].Number()
			if !ok {
				return nil, errors.New("invalid enchantment level")
			}
			w.VarInt(id).VarInt(int32(level))
		}
		tooltip, ok := o.Number("show_in_tooltip")
		w.Bool(!ok || tooltip != 0)
	case "dyed_color":
		o, ok := tag.Object()
		if !ok {
			return nil, errors.New("invalid dyed color")
		}
		color, _ := o.Number("rgb")
		tooltip, ok := o.Number("show_in_tooltip")
		w.I32(int32(color)).Bool(!ok || tooltip != 0)
	case "block_state":
		o, ok := tag.Object()
		if !ok {
			return nil, errors.New("invalid block state component")
		}
		w.VarInt(int32(len(o)))
		var keys []string
		for k := range o {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v, ok := o.Text(k)
			if !ok {
				return nil, errors.New("invalid block property")
			}
			w.String(k).String(v)
		}
	case "container", "bundle_contents", "charged_projectiles":
		entries, ok := tag.Elements()
		if !ok {
			return nil, errors.New("invalid contained items")
		}
		stacks := map[int]Stack{}
		size := 0
		for i, entry := range entries {
			index := i
			if name == "container" {
				o, ok := entry.Object()
				if !ok {
					return nil, errors.New("invalid container entry")
				}
				n, ok := o.Number("slot")
				if !ok || n < 0 || n >= 256 {
					return nil, errors.New("invalid container slot")
				}
				index = int(n)
				entry = o["item"]
			}
			s, err := FromNBT(entry)
			if err != nil {
				return nil, err
			}
			stacks[index] = s
			size = max(size, index+1)
		}
		w.VarInt(int32(size))
		for i := 0; i < size; i++ {
			stacks[i].Write(w)
		}
	default:
		return nil, fmt.Errorf("import codec not yet available for %s; original player file retained", name)
	}
	return w.Payload(), nil
}
