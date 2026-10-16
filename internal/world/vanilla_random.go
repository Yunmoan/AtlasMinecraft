package world

import (
	"atlas/internal/nbt"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// Read or create the independent world salt before any decoration can run.
// Missing salt migration does not touch existing chunks or move their spawn.
func savedBeehiveShuffleSalt(level nbt.Object) (salt int64, changed bool, err error) {
	marker, ok := level.Object("AtlasWorldgen")
	if !ok {
		marker = make(nbt.Object)
		level["AtlasWorldgen"] = nbt.Tag{Type: nbt.Compound, Value: marker}
	}
	if tag, exists := marker["BeehiveShuffleSalt"]; exists {
		if tag.Type != nbt.Long {
			return 0, false, fmt.Errorf("AtlasWorldgen.BeehiveShuffleSalt is not a long")
		}
		value, ok := tag.Number()
		if !ok {
			return 0, false, fmt.Errorf("invalid AtlasWorldgen.BeehiveShuffleSalt")
		}
		return value, false, nil
	}
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return 0, false, err
	}
	salt = int64(binary.LittleEndian.Uint64(raw[:]))
	marker["BeehiveShuffleSalt"] = nbt.Tag{Type: nbt.Long, Value: salt}
	return salt, true, nil
}
