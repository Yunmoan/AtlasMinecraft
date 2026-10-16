package anvil

import (
	"atlas/internal/nbt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Store serializes atomic region replacements with reads of their raw
// records. Different regions can progress independently; readers of the same
// region share a lock. Compression and NBT parsing run outside these locks.
// The zero value is ready to use. Locks are shared with other Store instances
// and the package functions, so all access to one save directory coordinates.
type Store struct{}

var defaultStore Store
var regionLocks [256]sync.RWMutex

func regionLock(dir string, cx, cz int32) *sync.RWMutex {
	path, err := filepath.Abs(dir)
	if err != nil {
		path = filepath.Clean(dir)
	}
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	// A bounded table avoids retaining a lock for every explored region.
	hash := uint32(2166136261)
	for i := 0; i < len(path); i++ {
		hash = (hash ^ uint32(path[i])) * 16777619
	}
	x, z := uint32(cx>>5), uint32(cz>>5)
	return &regionLocks[(hash^x*0x9e3779b9^z*0x85ebca6b)%uint32(len(regionLocks))]
}

func (s *Store) ReadChunk(dir string, cx, cz int32) (string, nbt.Tag, bool, error) {
	lock := regionLock(dir, cx, cz)
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
	lock := regionLock(dir, cx, cz)
	lock.Lock()
	defer lock.Unlock()
	return writeChunkRecord(dir, cx, cz, payload)
}
