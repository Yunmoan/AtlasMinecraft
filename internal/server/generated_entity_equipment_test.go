package server

import (
	"atlas/internal/nbt"
	"atlas/internal/world"
	"bytes"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

func TestGeneratedEntityEquipmentAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/generated_entity_equipment_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var entity world.WorldEntity
	count := 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) != 4 {
			continue
		}
		if f[0] == "actor_equipment_nbt" {
			raw, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			entity.Data, _ = tag.Object()
			entity.Type = f[1]
		}
		if f[0] == "actor_equipment" {
			want, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			got, err := generatedEntityEquipment(77, entity)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s got%x want%x", line, got, want)
			}
			count++
		}
	}
	if count == 0 {
		t.Fatal("empty equipment fixture")
	}
	t.Logf("matched %d actual original equipment packet codecs with saved held items", count)
}
