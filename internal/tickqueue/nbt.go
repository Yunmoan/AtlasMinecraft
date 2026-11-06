package tickqueue

import (
	"atlas/internal/nbt"
	"math"
)

// DecodeSavedNBT is SavedTick.loadTickList. It filters unknown registry types
// and positions owned by another chunk, but preserves order and duplicates.
// Missing/wrong-type numeric fields have the original CompoundTag defaults.
func DecodeSavedNBT[T comparable](list nbt.Tag, x, z int32, resolve func(string) (T, bool)) []Saved[T] {
	result := make([]Saved[T], 0)
	entries, ok := list.Elements()
	if !ok || list.ListType != nbt.Compound {
		return result
	}
	owner := ChunkKey(x, z)
	for _, entry := range entries {
		object, _ := entry.Object()
		name, _ := object.Text("i")
		kind, known := resolve(name)
		if !known {
			continue
		}
		p := Position{savedNBTInt(object["x"]), savedNBTInt(object["y"]), savedNBTInt(object["z"])}
		if p.chunk() != owner {
			continue
		}
		result = append(result, Saved[T]{kind, p, savedNBTInt(object["t"]), ClampPriority(savedNBTInt(object["p"]))})
	}
	return result
}

func EncodeSavedNBT[T comparable](ticks []Saved[T], name func(T) string) nbt.Tag {
	entries := make([]nbt.Tag, 0, len(ticks))
	for _, t := range ticks {
		entries = append(entries, nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
			"i": {Type: nbt.String, Value: name(t.Type)}, "x": {Type: nbt.Int, Value: t.Position.X},
			"y": {Type: nbt.Int, Value: t.Position.Y}, "z": {Type: nbt.Int, Value: t.Position.Z},
			"t": {Type: nbt.Int, Value: t.Delay}, "p": {Type: nbt.Int, Value: int32(t.Priority)},
		}})
	}
	elementType := byte(nbt.Compound)
	if len(entries) == 0 {
		elementType = nbt.End
	}
	return nbt.Tag{Type: nbt.List, ListType: elementType, Value: entries}
}

func savedNBTInt(tag nbt.Tag) int32 {
	if n, ok := tag.Number(); ok {
		return int32(n)
	}
	switch v := tag.Value.(type) {
	case float32:
		n := javaNarrowInt(float64(v))
		if v < float32(n) {
			n--
		}
		return n
	case float64:
		n := javaNarrowInt(v)
		if v < float64(n) {
			n--
		}
		return n
	}
	return 0
}

// Java narrowing saturates before Mth.floor's comparison/subtraction. This
// includes the overflow of MIN_VALUE-1 for a value below the integer range.
func javaNarrowInt(v float64) int32 {
	if math.IsNaN(v) {
		return 0
	}
	if v >= math.MaxInt32 {
		return math.MaxInt32
	}
	if v <= math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}
