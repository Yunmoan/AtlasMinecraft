package worldgen

import (
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestBastionTemplateActorsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "worldgen", "structure_bastion_actors_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	parseObject := func(line string) nbt.Object {
		t.Helper()
		fields := strings.Fields(line)
		data, err := hex.DecodeString(fields[3])
		if err != nil {
			t.Fatal(err)
		}
		tag, _, err := nbt.ParseNetwork(data)
		if err != nil {
			t.Fatal(err)
		}
		object, _ := tag.Object()
		delete(object, "UUID")
		return object
	}
	cases := 0
	for index, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "bastion_actor_case" {
			continue
		}
		seed, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		input, want := parseObject(lines[index+1]), parseObject(lines[index+2])
		id, _ := input.Text("id")
		// Both normalized Entity.load/save inputs and the original authored
		// templates must give the same final saved state.
		for _, mode := range []string{"normalized", "authored"} {
			actorInput := input
			if mode == "authored" {
				template, err := LoadJigsawTemplate("minecraft:bastion/mobs/" + fields[1])
				if err != nil || len(template.Entities) != 1 {
					t.Fatalf("%s template: %v", line, err)
				}
				actorInput = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: template.Entities[0].NBT}).Value.(nbt.Object)
				actorInput["Pos"], actorInput["Rotation"] = input["Pos"], input["Rotation"]
			}
			before := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: actorInput})
			random := NewXoroshiro(seed)
			got, err := FinalizeBastionTemplateEntity(PendingStructureEntity{ID: id, Position: FeaturePosition{99, 99, 99}, Data: actorInput}, random)
			if err != nil {
				t.Fatalf("%s %s: %v", line, mode, err)
			}
			if !reflect.DeepEqual(before.Value, actorInput) {
				t.Fatalf("%s %s mutated template", line, mode)
			}
			if len(got) != len(want) {
				t.Fatalf("%s %s field count %d != %d; got %v", line, mode, len(got), len(want), got)
			}
			for key, expected := range want {
				if a, b := iglooCanonicalNBT(t, got[key]), iglooCanonicalNBT(t, expected); a != b {
					t.Fatalf("%s %s %s: got %s want %s", line, mode, key, a, b)
				}
			}
			nextFields := strings.Fields(lines[index+4])
			next, _ := strconv.ParseInt(nextFields[4], 10, 64)
			if actual := int64(random.NextLong()); actual != next {
				t.Fatalf("%s %s RNG: %d != %d", line, mode, actual, next)
			}
		}
		cases++
	}
	if cases != 120 {
		t.Fatalf("expected120 cases, got%d", cases)
	}
	t.Logf("matched%d actual STRUCTURE finalizations with full NBT and RNG, both normalized and original authored inputs", cases)
}
