// Package javasave persists Java save files with gzip/NBT and recovery copies.
// Callers serialize writes to the same path. World metadata interpretation and
// generation remain with the world runtime.
package javasave

import (
	"atlas/internal/nbt"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// WriteNBT encodes a complete gzip NBT file and retains its previous contents.
func WriteNBT(path, name string, root nbt.Tag) error {
	encoded, err := nbt.Marshal(name, root)
	if err != nil {
		return err
	}
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	if _, err = zw.Write(encoded); err != nil {
		zw.Close()
		return err
	}
	if err = zw.Close(); err != nil {
		return err
	}
	return WriteFile(path, compressed.Bytes())
}

// WriteFile keeps the previous complete file as the vanilla .dat_old copy.
func WriteFile(path string, data []byte) error {
	if previous, err := os.ReadFile(path); err == nil {
		if err := replace(path+"_old", previous); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return replace(path, data)
}

func replace(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "atlas-vanilla-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// ReadGzip limits decompressed NBT to 32 MiB before it reaches the parser.
func ReadGzip(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	b, err := io.ReadAll(io.LimitReader(zr, (32<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 32<<20 {
		return nil, errors.New("gzip NBT exceeds 32 MiB limit")
	}
	return b, nil
}
