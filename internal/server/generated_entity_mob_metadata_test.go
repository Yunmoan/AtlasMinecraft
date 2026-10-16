package server

import (
	"atlas/internal/nbt"
	"atlas/internal/world"
	"bufio"
	"bytes"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGeneratedMobMetadataAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/generated_entities_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var entity world.WorldEntity
	cases, registries := 0, 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 4 {
			continue
		}
		if fields[0] == "entity_registry" {
			id, err := strconv.ParseInt(fields[3], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			registry := map[string]map[string]int32{"minecraft:cat_variant": generatedCatVariants, "minecraft:villager_type": generatedVillagerTypes, "minecraft:villager_profession": generatedVillagerProfessions}[fields[1]]
			if got, ok := registry[fields[2]]; !ok || got != int32(id) {
				t.Fatalf("registry %s %s=%d, want%d", fields[1], fields[2], got, id)
			}
			registries++
			continue
		}
		if fields[0] == "entity_default" || fields[0] == "entity_finalized" || fields[0] == "entity_finalized_fullmoon" || fields[0] == "entity_changed" {
			encoded, err := hex.DecodeString(fields[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(encoded)
			if err != nil {
				t.Fatal(err)
			}
			entity.Data, _ = tag.Object()
			entity.Type = fields[1]
			continue
		}
		if !strings.HasSuffix(fields[0], "_metadata") || entity.Type == "minecraft:end_crystal" {
			continue
		}
		want, err := hex.DecodeString(fields[3])
		if err != nil {
			t.Fatal(err)
		}
		got := generatedMobMetadata(77, entity)
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s seed%s got%x want%x", fields[0], fields[1], fields[2], got, want)
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 74 || registries != 33 {
		t.Fatalf("checked metadata=%d registries=%d", cases, registries)
	}
	t.Logf("matched %d actual JAR mob metadata packets and %d registry entries", cases, registries)
}
