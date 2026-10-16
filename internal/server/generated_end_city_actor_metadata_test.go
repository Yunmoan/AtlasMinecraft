package server

import (
	"atlas/internal/nbt"
	"atlas/internal/world"
	"bytes"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGeneratedEndCityActorsAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/structure_end_city_actors_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var entity world.WorldEntity
	metadata, types, spawns := 0, 0, 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		if f[0] == "end_city_actor_type" {
			id, err := strconv.Atoi(f[2])
			if err != nil || generatedEntityTypes[f[1]] != int32(id) {
				t.Fatal("End City registry mismatch", line, err)
			}
			types++
			continue
		}
		if len(f) != 4 {
			continue
		}
		raw, err := hex.DecodeString(f[3])
		if err != nil {
			t.Fatal(err)
		}
		switch f[0] {
		case "end_city_actor_default", "end_city_actor_saved":
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			entity.Data, _ = tag.Object()
			entity.Type = f[1]
			pos, _ := entity.Data["Pos"].Elements()
			for i := range pos {
				entity.Position[i] = pos[i].Value.(float64)
			}
			rotation, _ := entity.Data["Rotation"].Elements()
			for i := range rotation {
				entity.Rotation[i] = rotation[i].Value.(float32)
			}
		case "end_city_actor_default_metadata", "end_city_actor_metadata":
			if got := generatedEntityMetadata(77, entity); !bytes.Equal(got, raw) {
				t.Fatalf("%s: got %x want %x", line, got, raw)
			}
			metadata++
		case "end_city_actor_spawn":
			if got := generatedEntitySpawnPayload(77, entity); !bytes.Equal(got, raw) {
				t.Fatalf("%s: got %x want %x", line, got, raw)
			}
			spawns++
		}
	}
	if metadata != 18 || types != 2 || spawns != 4 {
		t.Fatal("incomplete End City actor codecs", metadata, types, spawns)
	}
	t.Logf("matched %d original End City actor metadata, %d item frame spawn packets and %d registry IDs", metadata, spawns, types)
}

func TestGeneratedShulkerChangedMetadataAgainstVanillaJAR(t *testing.T) {
	// Actual Shulker EntityData codec, independently captured from 1.21.1:
	// WEST attach face, 75 percent peek, lime color, Health 30.
	want, err := hex.DecodeString("4d090341f00000100c0411004b120005ff")
	if err != nil {
		t.Fatal(err)
	}
	entity := world.WorldEntity{Type: "minecraft:shulker", Data: nbt.Object{"Health": {Type: nbt.Float, Value: float32(30)}, "AttachFace": {Type: nbt.Byte, Value: int8(4)}, "Peek": {Type: nbt.Byte, Value: int8(75)}, "Color": {Type: nbt.Byte, Value: int8(5)}}}
	if got := generatedEntityMetadata(77, entity); !bytes.Equal(got, want) {
		t.Fatalf("original changed Shulker metadata: got %x want %x", got, want)
	}
}
