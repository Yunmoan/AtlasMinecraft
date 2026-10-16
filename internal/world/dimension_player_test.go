package world

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"bytes"
	"os"
	"testing"
)

func TestPlayerDimensionRestartAndRespawnPersistence(t *testing.T) {
	dir := t.TempDir()
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	s, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	uuid := [16]byte{7, 6, 7}
	p := PlayerPosition{X: 8.5, Y: 90, Z: -2.5, Yaw: 90, Flying: true, Dimension: "minecraft:the_nether"}
	inv := item.Inventory{Selected: 3}
	if err := s.SavePlayerState(uuid, p, inv); err != nil {
		t.Fatal(err)
	}
	path := s.vanilla.playerPath(uuid)
	data, err := readGzip(path)
	if err != nil {
		t.Fatal(err)
	}
	name, root, err := nbt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := root.Object()
	o["FixtureCustomTag"] = nbt.Tag{Type: nbt.String, Value: "preserved across dimensions"}
	if err := writeGzipNBTAtomic(path, name, root); err != nil {
		t.Fatal(err)
	}
	nether, err := OpenStoreDimension(dir, "nether")
	if err != nil {
		t.Fatal(err)
	}
	if err := nether.SavePlayerSpawn(uuid, 8, 90, -2, 180); err != nil {
		t.Fatal(err)
	}
	data, err = readGzip(path)
	if err != nil {
		t.Fatal(err)
	}
	_, root, err = nbt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	o, _ = root.Object()
	if got, _ := o.Text("SpawnDimension"); got != "minecraft:the_nether" {
		t.Fatalf("respawn dimension %q", got)
	}
	if got, _ := o.Text("Dimension"); got != p.Dimension {
		t.Fatal("spawnpoint changed current player dimension")
	}
	if got, _ := o.Text("FixtureCustomTag"); got != "preserved across dimensions" {
		t.Fatal("custom NBT tag lost")
	}
	reopened, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, found, err := reopened.LoadPlayer(uuid)
	if err != nil || !found || got != p {
		t.Fatalf("Nether player restart: %+v %v %v", got, found, err)
	}
	p.Dimension = "minecraft:the_end"
	p.Y = 80
	if err := reopened.SavePlayer(uuid, p); err != nil {
		t.Fatal(err)
	}
	got, found, err = s.LoadPlayer(uuid)
	if err != nil || !found || got != p {
		t.Fatalf("End player reload: %+v %v %v", got, found, err)
	}
	gotInventory, err := s.LoadInventory(uuid)
	if err != nil || gotInventory.Selected != 3 {
		t.Fatal("dimension change discarded inventory", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p.Dimension = "minecraft:unsupported"
	if err := s.SavePlayer(uuid, p); err == nil {
		t.Fatal("unknown dimension accepted")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("invalid dimension changed player file", err)
	}
	p.Dimension = ""
	if err := s.SavePlayer(uuid, p); err != nil {
		t.Fatal(err)
	}
	got, found, err = s.LoadPlayer(uuid)
	if err != nil || !found || got.Dimension != "minecraft:overworld" {
		t.Fatal("empty legacy dimension did not default to Overworld", err)
	}
}
