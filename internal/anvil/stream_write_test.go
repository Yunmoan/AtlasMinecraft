package anvil

import (
	"atlas/internal/nbt"
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

func TestStreamRewritePreservesRawRecordsTimestampsAndPadding(t *testing.T) {
	dir := t.TempDir()
	old := make([]byte, 12*sector)
	for i, kind := range []byte{1, 2, 3, 4, 0x82} {
		offset := 2 + 2*i
		binary.BigEndian.PutUint32(old[i*4:], uint32(offset<<8|2))
		binary.BigEndian.PutUint32(old[sector+i*4:], uint32(12345+i))
		raw := old[offset*sector : (offset+2)*sector]
		for j := range raw {
			raw[j] = byte(j*31 + i)
		}
		binary.BigEndian.PutUint32(raw, 50)
		raw[4] = kind
	}
	path := regionPath(dir, 0, 0)
	if err := os.WriteFile(path, old, 0644); err != nil {
		t.Fatal(err)
	}
	root := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"xPos": {Type: nbt.Int, Value: int32(2)}}}
	if err := WriteChunk(dir, 2, 0, "", root); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if i == 2 {
			continue
		}
		loc := binary.BigEndian.Uint32(got[i*4:])
		offset, count := int(loc>>8), int(loc&255)
		if count != 2 || !bytes.Equal(old[(2+2*i)*sector:(4+2*i)*sector], got[offset*sector:(offset+count)*sector]) {
			t.Fatalf("raw record/padding changed: %d", i)
		}
		if !bytes.Equal(old[sector+i*4:sector+i*4+4], got[sector+i*4:sector+i*4+4]) {
			t.Fatalf("neighbor timestamp changed: %d", i)
		}
	}
	if _, _, exists, err := ReadChunk(dir, 2, 0); err != nil || !exists {
		t.Fatalf("replacement not readable: %v", err)
	}
}

func TestStreamRewriteInvalidNeighborLeavesOriginalUntouched(t *testing.T) {
	dir := t.TempDir()
	old := make([]byte, 3*sector)
	binary.BigEndian.PutUint32(old[4:], uint32(100<<8|1))
	path := regionPath(dir, 0, 0)
	if err := os.WriteFile(path, old, 0644); err != nil {
		t.Fatal(err)
	}
	root := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
	if err := WriteChunk(dir, 0, 0, "", root); err == nil {
		t.Fatal("invalid neighbor was silently dropped")
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(old, got) {
		t.Fatal("failed replacement changed original")
	}
}
