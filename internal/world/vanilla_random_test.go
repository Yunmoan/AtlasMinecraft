package world

import (
	"atlas/internal/nbt"
	"os"
	"path/filepath"
	"testing"
)

func TestBeehiveShuffleSaltPersistsAndMigrates(t *testing.T) {
	dir := t.TempDir()
	seed := int64(1234)
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	readSalt := func() int64 {
		t.Helper()
		save, err := openVanilla(dir)
		if err != nil {
			t.Fatal(err)
		}
		root, _ := save.root.Object()
		level, _ := root.Object("Data")
		marker, _ := level.Object("AtlasWorldgen")
		salt, ok := marker.Number("BeehiveShuffleSalt")
		if !ok {
			t.Fatal("world salt missing")
		}
		return salt
	}
	first := readSalt()
	if second := readSalt(); second != first {
		t.Fatal("world salt changed on reopen")
	}
	save, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := save.root.Object()
	level, _ := root.Object("Data")
	marker, _ := level.Object("AtlasWorldgen")
	delete(marker, "BeehiveShuffleSalt")
	delete(marker, "SpawnReady") // Existing worlds without this tag keep spawn.
	level["SpawnX"] = nbt.Tag{Type: nbt.Int, Value: int32(177)}
	level["SpawnY"] = nbt.Tag{Type: nbt.Int, Value: int32(88)}
	level["SpawnZ"] = nbt.Tag{Type: nbt.Int, Value: int32(-999)}
	if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), save.name, save.root); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "region"), 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := []byte("existing region contents remain untouched")
	if err := os.WriteFile(filepath.Join(dir, "region", "sentinel.mca"), sentinel, 0644); err != nil {
		t.Fatal(err)
	}
	migrated := readSalt()
	if reopened := readSalt(); reopened != migrated {
		t.Fatal("migrated salt changed on restart")
	}
	save, err = openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	if x, y, z, _ := save.spawnPoint(); x != 177 || y != 88 || z != -999 || save.spawnUninitialized() {
		t.Fatal("salt migration moved or reinitialized existing spawn")
	}
	data, err := os.ReadFile(filepath.Join(dir, "region", "sentinel.mca"))
	if err != nil || string(data) != string(sentinel) {
		t.Fatal("salt migration touched region data")
	}
}
