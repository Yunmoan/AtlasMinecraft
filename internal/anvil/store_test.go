package anvil

import (
	"atlas/internal/nbt"
	"encoding/binary"
	"os"
	"sync"
	"testing"
)

func TestStoreConcurrentRegionReadsAndReplacements(t *testing.T) {
	dir := t.TempDir()
	var store Store
	root := func(x int32) nbt.Tag {
		return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"xPos": {Type: nbt.Int, Value: x}}}
	}
	if err := store.WriteChunk(dir, -1, -1, "", root(-1)); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := int32(0); i < 24; i++ {
		wg.Add(1)
		go func(x int32) {
			defer wg.Done()
			// Several writers share each region; other writers cross its boundary.
			x -= 12
			for pass := 0; pass < 3; pass++ {
				if err := store.WriteChunk(dir, x, -1, "", root(x)); err != nil {
					t.Error(err)
					return
				}
				_, got, exists, err := store.ReadChunk(dir, -1, -1)
				object, _ := got.Object()
				value, ok := object.Number("xPos")
				if err != nil || !exists || !ok || value != -1 {
					t.Errorf("concurrent read: exists=%v value=%d err=%v", exists, value, err)
				}
			}
		}(i)
	}
	wg.Wait()
	for x := int32(-12); x < 12; x++ {
		_, got, exists, err := store.ReadChunk(dir, x, -1)
		object, _ := got.Object()
		value, ok := object.Number("xPos")
		if err != nil || !exists || !ok || value != int64(x) {
			t.Fatalf("write lost at %d,-1: value=%d exists=%v err=%v", x, value, exists, err)
		}
	}
}

func TestReadChunkRejectsInvalidRecordBounds(t *testing.T) {
	for _, tc := range []struct {
		name             string
		location, length uint32
		size             int
	}{
		{"header_sector", 1<<8 | 1, 1, 3 * sector},
		{"zero_sectors", 2 << 8, 1, 3 * sector},
		{"past_file", 4<<8 | 1, 1, 3 * sector},
		{"zero_length", 2<<8 | 1, 0, 3 * sector},
		{"past_allocation", 2<<8 | 1, sector, 3 * sector},
		{"truncated_payload", 2<<8 | 1, 10, 2*sector + 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			data := make([]byte, tc.size)
			binary.BigEndian.PutUint32(data[:4], tc.location)
			if len(data) >= 2*sector+5 {
				binary.BigEndian.PutUint32(data[2*sector:], tc.length)
				data[2*sector+4] = 2
			}
			if err := os.WriteFile(regionPath(dir, 0, 0), data, 0644); err != nil {
				t.Fatal(err)
			}
			if _, _, exists, err := ReadChunk(dir, 0, 0); err == nil || exists {
				t.Fatalf("invalid chunk accepted: exists=%v err=%v", exists, err)
			}
		})
	}
}
