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

func TestGeneratedAnimalMetadataAgainstVanillaJAR(t *testing.T) {
	fixture, err := os.ReadFile("../../docs/worldgen/initial_animals_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var entity world.WorldEntity
	cases, types := 0, 0
	for _, line := range strings.Split(string(fixture), "\n") {
		f := strings.Fields(line)
		if len(f) < 4 {
			continue
		}
		if f[0] == "animal_type" {
			id, err := strconv.ParseInt(f[2], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			if generatedEntityTypes[f[1]] != int32(id) {
				t.Fatal(f[1], generatedEntityTypes[f[1]], id)
			}
			types++
			continue
		}
		if f[0] == "animal_default" || f[0] == "animal_changed" {
			data, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(data)
			if err != nil {
				t.Fatal(err)
			}
			entity.Data, _ = tag.Object()
			entity.Type = f[1]
			continue
		}
		if f[0] != "animal_default_metadata" && f[0] != "animal_changed_metadata" {
			continue
		}
		want, err := hex.DecodeString(f[3])
		if err != nil {
			t.Fatal(err)
		}
		got := generatedMobMetadata(77, entity)
		if !bytes.Equal(got, want) {
			t.Fatalf("%s: got%x want%x", line, got, want)
		}
		cases++
	}
	if cases != 102 || types != 6 {
		t.Fatal(cases, types)
	}
	t.Logf("matched %d actual default/loaded animal metadata packets and %d official entity type IDs", cases, types)
}

func TestGeneratedOtherAnimalMetadataAgainstVanillaJAR(t *testing.T) {
	fixture, err := os.ReadFile("../../docs/worldgen/other_initial_animals_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var entity world.WorldEntity
	cases, types, registries := 0, 0, 0
	for _, line := range strings.Split(string(fixture), "\n") {
		f := strings.Fields(line)
		if len(f) < 4 {
			continue
		}
		switch f[0] {
		case "animal_registry":
			id, _ := strconv.ParseInt(f[3], 10, 32)
			registry := map[string]map[string]int32{"minecraft:wolf_variant": generatedWolfVariants, "minecraft:frog_variant": generatedFrogVariants}[f[1]]
			if got, ok := registry[f[2]]; !ok || got != int32(id) {
				t.Fatal(line, got)
			}
			registries++
		case "animal_type":
			id, _ := strconv.ParseInt(f[2], 10, 32)
			if got := generatedEntityTypes[f[1]]; got != int32(id) {
				t.Fatal(line, got)
			}
			types++
		case "animal_default", "animal_finalized":
			encoded, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(encoded)
			if err != nil {
				t.Fatal(err)
			}
			entity.Data, _ = tag.Object()
			entity.Type = f[1]
		case "animal_default_metadata", "animal_finalized_metadata":
			want, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			got := generatedMobMetadata(77, entity)
			if !bytes.Equal(got, want) {
				t.Fatalf("%s got%x want%x", line, got, want)
			}
			cases++
		}
	}
	if cases != 2028 || types != 12 || registries != 12 {
		t.Fatal(cases, types, registries)
	}
	t.Logf("matched %d actual JAR subtype metadata packets, %d entity type IDs and %d configured variant registry entries", cases, types, registries)
}
