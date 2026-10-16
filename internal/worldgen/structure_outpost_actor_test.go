package worldgen

import (
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestOutpostTemplateActorsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "worldgen", "structure_outpost_actors_1.21.1.txt"))
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
		tag, used, err := nbt.ParseNetwork(data)
		if err != nil || used != len(data) {
			t.Fatal("invalid native Outpost NBT", err)
		}
		object, _ := tag.Object()
		delete(object, "UUID")
		return object
	}
	cases := 0
	for index, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "outpost_actor_case" {
			continue
		}
		seed, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		actorIndex, _ := strconv.Atoi(fields[2])
		input, want := parseObject(lines[index+1]), parseObject(lines[index+2])
		id, _ := input.Text("id")
		for _, mode := range []string{"normalized", "authored"} {
			actorInput := input
			if mode == "authored" {
				template, err := LoadJigsawTemplate(fields[1])
				if err != nil || actorIndex >= len(template.Entities) {
					t.Fatal("Outpost actor template missing", err)
				}
				actorInput = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: template.Entities[actorIndex].NBT}).Value.(nbt.Object)
				actorInput["Pos"], actorInput["Rotation"] = input["Pos"], input["Rotation"]
			}
			before := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: actorInput})
			random := NewXoroshiro(seed)
			got, err := FinalizeOutpostTemplateEntity(PendingStructureEntity{ID: id, Data: actorInput}, random)
			if err != nil {
				t.Fatalf("%s %s: %v", line, mode, err)
			}
			if !reflect.DeepEqual(before.Value, actorInput) {
				t.Fatal("Outpost finalization mutated authored NBT")
			}
			if len(got) != len(want) {
				t.Fatalf("%s %s fields %d != %d", line, mode, len(got), len(want))
			}
			for key, expected := range want {
				actual := got[key]
				if key == "attributes" {
					actual, expected = canonicalOutpostActorAttributes(actual), canonicalOutpostActorAttributes(expected)
				}
				if a, b := iglooCanonicalNBT(t, actual), iglooCanonicalNBT(t, expected); a != b {
					t.Fatalf("%s %s %s: got %s want %s", line, mode, key, a, b)
				}
			}
			nextFields := strings.Fields(lines[index+4])
			next, _ := strconv.ParseInt(nextFields[5], 10, 64)
			if actual := int64(random.NextLong()); actual != next {
				t.Fatalf("%s %s RNG %d != %d", line, mode, actual, next)
			}
		}
		cases++
	}
	if cases != 72 {
		t.Fatalf("matched%d original cases, want72", cases)
	}
	t.Logf("matched72 actual STRUCTURE finalizations with full NBT and RNG using authored and normalized inputs")
}

// The original AttributeInstanceMap uses object identity hashes. Compare all
// attribute values and modifiers while ignoring only its unordered list order.
func canonicalOutpostActorAttributes(tag nbt.Tag) nbt.Tag {
	tag = cloneFeatureNBT(tag)
	values, _ := tag.Elements()
	sort.Slice(values, func(i, j int) bool {
		a, _ := values[i].Object()
		b, _ := values[j].Object()
		x, _ := a.Text("id")
		y, _ := b.Text("id")
		return x < y
	})
	return tag
}
