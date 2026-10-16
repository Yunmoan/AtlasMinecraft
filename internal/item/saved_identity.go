package item

import (
	"atlas/internal/nbt"
	"bytes"
	"reflect"
)

// SameSavedFields compares opaque saved item fields semantically. ID, count,
// components and the inventory Slot are controlled/compared separately;
// compound field order is not identity.
func (s Stack) SameSavedFields(other Stack) bool {
	if bytes.Equal(s.Saved, other.Saved) {
		return true
	}
	fields := func(raw []byte) (nbt.Object, bool) {
		if len(raw) == 0 {
			return nbt.Object{}, true
		}
		_, tag, err := nbt.Parse(raw)
		if err != nil {
			return nil, false
		}
		data, ok := tag.Object()
		if !ok {
			return nil, false
		}
		for _, name := range []string{"id", "count", "components", "Slot"} {
			delete(data, name)
		}
		return data, true
	}
	a, aOK := fields(s.Saved)
	b, bOK := fields(other.Saved)
	return aOK && bOK && reflect.DeepEqual(a, b)
}
