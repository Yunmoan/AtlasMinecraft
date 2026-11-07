package javasave

import (
	"atlas/internal/nbt"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteNBTRoundTripAndRecoveryCopy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "level.dat")
	root := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"value": {Type: nbt.Int, Value: int32(1)}}}
	if err := WriteNBT(path, "Level", root); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	root.Value.(nbt.Object)["value"] = nbt.Tag{Type: nbt.Int, Value: int32(2)}
	if err := WriteNBT(path, "Level", root); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(path + "_old")
	if err != nil || !bytes.Equal(first, old) {
		t.Fatal("previous complete save was not retained", err)
	}
	data, err := ReadGzip(path)
	if err != nil {
		t.Fatal(err)
	}
	name, decoded, err := nbt.Parse(data)
	if err != nil || name != "Level" {
		t.Fatal(name, err)
	}
	object, _ := decoded.Object()
	if value, _ := object.Number("value"); value != 2 {
		t.Fatal(value)
	}
}

func TestReadGzipRejectsOversizedInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.dat")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zip := gzip.NewWriter(file)
	if _, err := zip.Write(make([]byte, (32<<20)+1)); err != nil {
		t.Fatal(err)
	}
	if err := zip.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGzip(path); err == nil || !strings.Contains(err.Error(), "32 MiB") {
		t.Fatal(err)
	}
}

func TestWriteFailureDoesNotReplaceExistingSave(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "level.dat")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+"_old", 0700); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(path, []byte("replacement")); err == nil {
		t.Fatal("expected recovery-copy failure")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "original" {
		t.Fatal("original changed", err)
	}
}
