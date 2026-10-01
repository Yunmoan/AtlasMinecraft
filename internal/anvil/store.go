package anvil

import (
	"atlas/internal/nbt"
	"sync"
)

// Store serializes atomic region replacements with reads of their raw
// records. Different regions can progress independently; readers of the same
// region share a lock. Compression and NBT parsing run outside these locks.
// The zero value is ready to use. Use one Store for a save directory.
type Store struct {
	locks [64]sync.RWMutex
}

func (s *Store) regionLock(cx, cz int32) *sync.RWMutex {
	x, z := uint32(cx>>5), uint32(cz>>5)
	return &s.locks[(x*0x9e3779b9^z*0x85ebca6b)%uint32(len(s.locks))]
}

func (s *Store) ReadChunk(dir string, cx, cz int32) (string, nbt.Tag, bool, error) {
	lock := s.regionLock(cx, cz)
	lock.RLock()
	kind, compressed, exists, err := readChunkRecord(dir, cx, cz)
	lock.RUnlock()
	if err != nil || !exists {
		return "", nbt.Tag{}, false, err
	}
	return decodeChunkRecord(kind, compressed, cx, cz)
}

func (s *Store) WriteChunk(dir string, cx, cz int32, name string, root nbt.Tag) error {
	payload, err := encodeChunkRecord(name, root)
	if err != nil {
		return err
	}
	lock := s.regionLock(cx, cz)
	lock.Lock()
	defer lock.Unlock()
	return writeChunkRecord(dir, cx, cz, payload)
}
