package anvil

import (
	"atlas/internal/nbt"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestWindowsRegionReplacementWaitsForTransientReader(t *testing.T) {
	dir := t.TempDir()
	root := func(x int32) nbt.Tag {
		return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"xPos": {Type: nbt.Int, Value: x}}}
	}
	if err := WriteChunk(dir, 0, 0, "", root(0)); err != nil {
		t.Fatal(err)
	}
	path := regionPath(dir, 0, 0)
	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	probe := filepath.Join(dir, "probe.tmp")
	if err := os.WriteFile(probe, []byte("must not replace region"), 0644); err != nil {
		t.Fatal(err)
	}
	// Reproduce the exact Windows rename Access denied failure first.
	if err := os.Rename(probe, path); !errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
		t.Fatalf("expected Windows Access denied while reader is open, got %v", err)
	} else {
		t.Logf("reproduced startup failure: %v", err)
	}
	closed := make(chan error, 1)
	go func() {
		time.Sleep(80 * time.Millisecond)
		closed <- reader.Close()
	}()
	if err := WriteChunk(dir, 1, 0, "", root(1)); err != nil {
		t.Fatal(err)
	}
	if err := <-closed; err != nil {
		t.Fatal(err)
	}
	for x := int32(0); x <= 1; x++ {
		_, got, exists, err := ReadChunk(dir, x, 0)
		object, _ := got.Object()
		value, ok := object.Number("xPos")
		if err != nil || !exists || !ok || value != int64(x) {
			t.Fatalf("chunk %d lost after replacement: exists=%v value=%d err=%v", x, exists, value, err)
		}
	}
}

func TestWindowsPersistentRegionLockPreservesSave(t *testing.T) {
	dir := t.TempDir()
	root := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}}
	if err := WriteChunk(dir, 0, 0, "", root); err != nil {
		t.Fatal(err)
	}
	path := regionPath(dir, 0, 0)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := WriteChunk(dir, 1, 0, "", root); !errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
		t.Fatalf("persistent lock must return Access denied, got %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("failed replacement changed current region: %v", err)
	}
	leftovers, err := filepath.Glob(filepath.Join(dir, "atlas-region-*.tmp"))
	if err != nil || len(leftovers) != 0 {
		t.Fatalf("temporary files left after failed replacement: %v, %v", leftovers, err)
	}
}

func TestWindowsReplacementRetriesOnlyHandleContention(t *testing.T) {
	for _, code := range []syscall.Errno{5, 32, 33} {
		err := &os.LinkError{Op: "rename", Old: "old", New: "new", Err: code}
		if !transientReplacementError(err) {
			t.Errorf("expected retry for Windows error %d", code)
		}
	}
	for _, code := range []syscall.Errno{2, 3, 17, 112, 123} {
		err := &os.LinkError{Op: "rename", Old: "old", New: "new", Err: code}
		if transientReplacementError(err) {
			t.Errorf("unexpected retry for Windows error %d", code)
		}
	}
}
