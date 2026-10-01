// Package item owns protocol-767 item stacks, including their component patch.
package item

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

//go:embed items_767.json
var itemData []byte

//go:embed schema_767.json
var schemaData []byte

var entries []struct {
	ID        int32
	Name      string
	StackSize int32
}
var byName = map[string]int32{}
var schema map[string]any
var componentNames map[string]any

func init() {
	if err := json.Unmarshal(itemData, &entries); err != nil {
		panic(err)
	}
	for id, entry := range entries {
		if int(entry.ID) != id {
			panic("unordered item registry")
		}
		byName[entry.Name] = entry.ID
	}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		panic(err)
	}
	componentNames = schema["SlotComponentType"].([]any)[1].(map[string]any)["mappings"].(map[string]any)
}

type Stack struct {
	ID, Count int32
	// Patch starts with the added/removed counts. It is copied at ownership
	// boundaries, preserving every component and nested stack byte-for-byte.
	Patch []byte
	// Saved retains all original vanilla item fields when importing a save.
	Saved []byte
}

func Name(id int32) (string, bool) {
	if id <= 0 || int(id) >= len(entries) {
		return "", false
	}
	return entries[id].Name, true
}
func ID(name string) (int32, bool) {
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	id, ok := byName[name]
	return id, ok && id != 0
}
func New(id, count int32) Stack {
	if _, ok := Name(id); !ok || count <= 0 {
		return Stack{}
	}
	return Stack{ID: id, Count: count, Patch: []byte{0, 0}}
}
func (s Stack) Empty() bool { return s.Count == 0 }
func (s Stack) Clone() Stack {
	s.Patch = bytes.Clone(s.Patch)
	s.Saved = bytes.Clone(s.Saved)
	return s
}
func (s Stack) WithCount(count int32) Stack {
	if count <= 0 {
		return Stack{}
	}
	s.Count = count
	return s.Clone()
}
func (s Stack) SameItem(other Stack) bool {
	if s.Empty() || other.Empty() || s.ID != other.ID {
		return false
	}
	if bytes.Equal(s.Patch, other.Patch) {
		return true
	}
	// Patches are maps on the wire: component order is not item identity.
	// Preserve their original bytes for saves, but compare semantic values
	// when importing Java items or merging differently ordered client patches.
	a, removedA, errA := s.components()
	b, removedB, errB := other.components()
	if errA != nil || errB != nil || len(removedA) != len(removedB) || !componentValueEqual(a, b) {
		return false
	}
	removed := make(map[string]bool, len(removedA))
	for _, name := range removedA {
		removed[name] = true
	}
	for _, name := range removedB {
		if !removed[name] {
			return false
		}
	}
	return true
}

func componentValueEqual(a, b any) bool {
	switch a := a.(type) {
	case Stack:
		other, ok := b.(Stack)
		return ok && a.Count == other.Count && (a.Empty() && other.Empty() || a.SameItem(other))
	case map[string]any:
		other, ok := b.(map[string]any)
		if !ok || len(a) != len(other) {
			return false
		}
		for key, value := range a {
			otherValue, exists := other[key]
			if !exists || !componentValueEqual(value, otherValue) {
				return false
			}
		}
		return true
	case []any:
		other, ok := b.([]any)
		if !ok || len(a) != len(other) {
			return false
		}
		for i, value := range a {
			if !componentValueEqual(value, other[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}
func (s Stack) MaxCount() int32 {
	if s.Empty() {
		return 64
	}
	components, _, err := s.components()
	if err == nil {
		if v, ok := components["max_stack_size"].(int32); ok && v > 0 && v <= 99 {
			return v
		}
	}
	return entries[s.ID].StackSize
}
func (s Stack) Write(w *proto.Writer) {
	w.VarInt(s.Count)
	if !s.Empty() {
		w.VarInt(s.ID).Raw(s.Patch)
	}
}
func (s Stack) Bytes() []byte { w := proto.NewWriter(); s.Write(w); return w.Payload() }

func Read(r *proto.Reader) (Stack, error) { return read(r, 0) }
func read(r *proto.Reader, depth int) (Stack, error) {
	if depth > 32 {
		return Stack{}, errors.New("item nesting limit exceeded")
	}
	count, err := r.VarInt()
	if err != nil || count < 0 || count > 99 {
		return Stack{}, errors.New("invalid item count")
	}
	if count == 0 {
		return Stack{}, nil
	}
	id, err := r.VarInt()
	if _, ok := Name(id); err != nil || !ok {
		return Stack{}, errors.New("invalid item ID")
	}
	rest := r.Rest()
	if _, _, err := readComponents(r, depth+1); err != nil {
		return Stack{}, err
	}
	size := len(rest) - r.Remaining()
	if size > 1<<20 {
		return Stack{}, errors.New("item patch exceeds 1 MiB")
	}
	return Stack{ID: id, Count: count, Patch: bytes.Clone(rest[:size])}, nil
}
func (s Stack) components() (map[string]any, []string, error) {
	return readComponents(proto.NewReader(s.Patch), 0)
}
func readComponents(r *proto.Reader, depth int) (map[string]any, []string, error) {
	added, e1 := r.VarInt()
	removed, e2 := r.VarInt()
	if e1 != nil || e2 != nil || added < 0 || removed < 0 || added+removed > 57 {
		return nil, nil, errors.New("invalid component counts")
	}
	values := make(map[string]any)
	seen := make(map[int32]bool)
	fields := schema["SlotComponent"].([]any)[1].([]any)[1].(map[string]any)["type"].([]any)[1].(map[string]any)["fields"].(map[string]any)
	for range added {
		id, err := r.VarInt()
		name, ok := componentNames[fmt.Sprint(id)].(string)
		if err != nil || !ok || seen[id] {
			return nil, nil, errors.New("invalid or duplicate component ID")
		}
		seen[id] = true
		value, err := readValue(r, fields[name], nil, depth+1)
		if err != nil {
			return nil, nil, fmt.Errorf("component %s: %w", name, err)
		}
		values[name] = value
	}
	var removals []string
	for range removed {
		id, err := r.VarInt()
		name, ok := componentNames[fmt.Sprint(id)].(string)
		if err != nil || !ok || seen[id] {
			return nil, nil, errors.New("invalid or duplicate removed component")
		}
		seen[id] = true
		removals = append(removals, name)
	}
	return values, removals, nil
}

// The schema is restricted to item component framing, not arbitrary packets.
// Collections and recursion are bounded before allocation on untrusted input.
func readValue(r *proto.Reader, typ any, context map[string]any, depth int) (any, error) {
	if depth > 64 {
		return nil, errors.New("component nesting limit exceeded")
	}
	if name, ok := typ.(string); ok {
		switch name {
		case "void":
			return nil, nil
		case "varint":
			return r.VarInt()
		case "i32":
			return r.I32()
		case "i64", "position":
			return r.I64()
		case "f32":
			return r.F32()
		case "f64":
			return r.F64()
		case "bool":
			v, err := r.U8()
			if err != nil || v > 1 {
				return nil, errors.New("invalid boolean")
			}
			return v == 1, nil
		case "UUID":
			return r.UUID()
		case "anonymousNbt", "anonOptionalNbt":
			tag, size, err := nbt.ParseNetwork(r.Rest())
			if err != nil {
				return nil, err
			}
			_, err = r.Bytes(size)
			return tag, err
		case "Slot":
			return read(r, depth+1)
		}
		definition, ok := schema[name]
		if !ok {
			return nil, fmt.Errorf("unknown item schema %s", name)
		}
		return readValue(r, definition, context, depth+1)
	}
	definition, ok := typ.([]any)
	if !ok || len(definition) != 2 {
		return nil, errors.New("invalid embedded item schema")
	}
	kind := definition[0].(string)
	if kind == "container" {
		out := make(map[string]any)
		for _, entry := range definition[1].([]any) {
			field := entry.(map[string]any)
			value, err := readValue(r, field["type"], out, depth+1)
			if err != nil {
				return nil, err
			}
			if anonymous, _ := field["anon"].(bool); anonymous {
				if object, ok := value.(map[string]any); ok {
					for k, v := range object {
						out[k] = v
					}
				}
			} else {
				out[field["name"].(string)] = value
			}
		}
		return out, nil
	}
	if kind == "option" {
		v, err := readValue(r, "bool", context, depth+1)
		if err != nil {
			return nil, err
		}
		if !v.(bool) {
			return nil, nil
		}
		return readValue(r, definition[1], context, depth+1)
	}
	args := definition[1].(map[string]any)
	switch kind {
	case "pstring":
		return r.String(32767)
	case "mapper":
		v, err := readValue(r, args["type"], context, depth+1)
		if err != nil {
			return nil, err
		}
		mapped, ok := args["mappings"].(map[string]any)[fmt.Sprint(v)]
		if !ok {
			return nil, errors.New("invalid component enum")
		}
		return mapped, nil
	case "switch":
		key := fmt.Sprint(context[args["compareTo"].(string)])
		selected, ok := args["fields"].(map[string]any)[key]
		if !ok {
			selected = args["default"]
		}
		return readValue(r, selected, context, depth+1)
	case "array":
		var count int32
		if countType, ok := args["countType"]; ok {
			v, err := readValue(r, countType, context, depth+1)
			if err != nil {
				return nil, err
			}
			count = v.(int32)
		} else {
			count = context[args["count"].(string)].(int32)
		}
		if count < 0 || count > 4096 {
			return nil, errors.New("component array limit exceeded")
		}
		out := make([]any, count)
		for i := range out {
			v, err := readValue(r, args["type"], context, depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case "registryEntryHolder":
		id, err := r.VarInt()
		if err != nil || id < 0 {
			return nil, errors.New("invalid registry holder")
		}
		if id != 0 {
			return map[string]any{"id": id - 1}, nil
		}
		v, err := readValue(r, args["otherwise"].(map[string]any)["type"], context, depth+1)
		return map[string]any{"data": v}, err
	case "registryEntryHolderSet":
		count, err := r.VarInt()
		if err != nil || count < 0 || count > 4097 {
			return nil, errors.New("invalid registry set")
		}
		if count == 0 {
			v, err := readValue(r, args["base"].(map[string]any)["type"], context, depth+1)
			return map[string]any{"tag": v}, err
		}
		out := make([]any, count-1)
		for i := range out {
			v, err := readValue(r, args["otherwise"].(map[string]any)["type"], context, depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return map[string]any{"ids": out}, nil
	}
	return nil, fmt.Errorf("unsupported item schema kind %s", kind)
}
