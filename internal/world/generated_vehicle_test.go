package world

import (
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedVehicleSavedStateAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/generated_vehicles_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) != 4 || f[0] != "vehicle_default" && f[0] != "vehicle_saved" {
			continue
		}
		raw, err := hex.DecodeString(f[3])
		if err != nil {
			t.Fatal(err)
		}
		tag, _, err := nbt.ParseNetwork(raw)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := tag.Object()
		got, ok, err := defaultGeneratedEntity(f[1])
		if err != nil || !ok {
			t.Fatal(ok, err)
		}
		if f[0] == "vehicle_saved" {
			pending, _ := pendingEntityTag(f[1], -16, 70, 16, "Minecart:0").Object()
			pending["Data"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"LootTable": want["LootTable"], "LootTableSeed": want["LootTableSeed"]}}
			entity, ok, err := realizeGeneratedEntity(1234, "minecraft:overworld", pending)
			if err != nil || !ok {
				t.Fatal(ok, err)
			}
			got = entity.Data
			want["UUID"] = got["UUID"]
			want["Pos"] = got["Pos"]
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s got%v want%v", line, got, want)
		}
		cases++
	}
	if cases != 4 {
		t.Fatal(cases)
	}
}
