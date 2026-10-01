package main

import (
	"atlas/internal/world"
	"os"
	"path/filepath"
	"testing"
)

func TestNewWorldTypeGate(t *testing.T) {
	for _, kind := range []string{"minecraft:normal", "normal", "minecraft:flat", "flat"} {
		t.Run(kind, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "new-world")
			seed := int64(1234)
			if err := createWorldIfMissing(dir, kind, &seed); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(filepath.Join(dir, "level.dat"))
			if err != nil {
				t.Fatal(err)
			}
			w, err := world.Open(nil, 1, dir)
			if err != nil {
				t.Fatal(err)
			}
			if w.Seed() != seed {
				t.Fatalf("saved seed=%d", w.Seed())
			}
			if w.IsFlat() != (kind == "flat" || kind == "minecraft:flat") {
				t.Fatal("wrong saved generator")
			}
			if err := createWorldIfMissing(dir, "minecraft:normal", nil); err != nil {
				t.Fatal(err)
			}
			after, _ := os.ReadFile(filepath.Join(dir, "level.dat"))
			if string(before) != string(after) {
				t.Fatal("existing world was overwritten")
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
	dir := filepath.Join(t.TempDir(), "unsupported")
	if err := createWorldIfMissing(dir, "minecraft:unknown", nil); err == nil {
		t.Fatal("unknown generator accepted")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("invalid generator created a directory")
	}
}
