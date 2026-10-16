package worldgen

import (
	"atlas/internal/nbt"
	"compress/gzip"
	"encoding/hex"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestVillageTemplateActorsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/structure_village_actors_1.21.1.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	raw, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	parse := func(line string) nbt.Object {
		t.Helper()
		f := strings.Fields(line)
		data, err := hex.DecodeString(f[3])
		if err != nil {
			t.Fatal(err)
		}
		tag, used, err := nbt.ParseNetwork(data)
		if err != nil || used != len(data) {
			t.Fatal("invalid native village NBT", err)
		}
		object, _ := tag.Object()
		delete(object, "UUID")
		return object
	}
	cases, species, templates := 0, map[string]bool{}, map[string]bool{}
	biomes, babies, weapons, jockeys, camelTimes := 0, 0, 0, 0, 0
	for index, line := range lines {
		f := strings.Fields(line)
		if len(f) == 3 && f[0] == "village_actor_biome" {
			if got := villageBiomeType(f[1]); got != f[2] {
				t.Fatalf("native biome %s type %s != %s", f[1], got, f[2])
			}
			biomes++
			continue
		}
		if len(f) == 4 && f[0] == "village_actor_camel_time" {
			time, _ := strconv.ParseInt(f[2], 10, 64)
			data := parse(line)
			want, _ := data.Number("LastPoseTick")
			if got := max(int64(0), time-53); got != want {
				t.Fatalf("camel original gametime%d pose tick%d != %d", time, got, want)
			}
			camelTimes++
			continue
		}
		if len(f) == 0 || f[0] != "village_actor_case" {
			continue
		}
		actorIndex, _ := strconv.Atoi(f[2])
		seed, _ := strconv.ParseInt(f[4], 10, 64)
		actorSeed, _ := strconv.ParseInt(f[5], 10, 64)
		input, want := parse(lines[index+1]), parse(lines[index+2])
		id, _ := input.Text("id")
		if id == "minecraft:zombie_villager" {
			baby, _ := want.Number("IsBaby")
			babies += int(baby)
			hands, _ := want["HandItems"].Elements()
			if item, _ := hands[0].Object(); len(item) > 0 {
				weapons++
			}
		}
		template, err := LoadJigsawTemplate(f[1])
		if err != nil || actorIndex >= len(template.Entities) {
			t.Fatal("village actor template missing", err)
		}
		for _, mode := range []string{"normalized", "authored"} {
			actorInput := input
			if mode == "authored" {
				actorInput = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: template.Entities[actorIndex].NBT}).Value.(nbt.Object)
				actorInput["Pos"], actorInput["Rotation"] = input["Pos"], input["Rotation"]
			}
			before := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: actorInput})
			random, actor := NewXoroshiro(seed), NewLegacyRandom(actorSeed)
			got, err := FinalizeVillageTemplateEntity(PendingStructureEntity{ID: id, Data: actorInput}, random, actor, f[3], f[7] == "1", false)
			if err != nil {
				t.Fatalf("%s %s: %v", line, mode, err)
			}
			if !reflect.DeepEqual(before.Value, actorInput) {
				t.Fatal("village finalization mutated authored NBT")
			}
			if len(got.Data) != len(want) {
				t.Fatalf("%s %s fields %d != %d", line, mode, len(got.Data), len(want))
			}
			for key, expected := range want {
				if a, b := iglooCanonicalNBT(t, got.Data[key]), iglooCanonicalNBT(t, expected); a != b {
					t.Fatalf("%s %s %s: got %s want %s", line, mode, key, a, b)
				}
			}
			next := strings.Fields(lines[index+4])
			regionNext, _ := strconv.ParseInt(next[2], 10, 64)
			actorNext, _ := strconv.ParseInt(next[3], 10, 64)
			if actual := int64(random.NextLong()); actual != regionNext {
				t.Fatalf("%s %s region RNG %d != %d", line, mode, actual, regionNext)
			}
			if actual := int64(actor.NextLong()); actual != actorNext {
				t.Fatalf("%s %s entity RNG %d != %d", line, mode, actual, actorNext)
			}
			extra, _ := strconv.Atoi(next[4])
			if got.ChickenJockey != (extra == 1) {
				t.Fatal("village chicken jockey mismatch")
			}
			if got.ChickenJockey {
				wantChicken := parse(lines[index+5])
				delete(wantChicken, "Passengers")
				gotAttrs, _ := got.ChickenData["attributes"].Elements()
				gotAttrs, _ = villageActorAttributes(gotAttrs)
				got.ChickenData["attributes"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: gotAttrs}
				if a, b := iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: got.ChickenData}), iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: wantChicken}); a != b {
					t.Fatalf("%s chicken full NBT got%s != %s", line, a, b)
				}
			}
		}
		extra, _ := strconv.Atoi(strings.Fields(lines[index+4])[4])
		jockeys += extra
		cases++
		species[id], templates[f[1]] = true, true
	}
	if cases != 7888 || len(species) != 10 || len(templates) != 47 || biomes != 64 || camelTimes != 11 || babies == 0 || weapons == 0 || jockeys == 0 {
		t.Fatal("incomplete village actor coverage", cases, len(species), len(templates), biomes, camelTimes, babies, weapons, jockeys)
	}
	t.Logf("matched %d original STRUCTURE finalizations, full NBT and both RNG streams from normalized/authored inputs; %d species in %d actor templates; %d builtin biome mappings, %d camel game times; %d baby zombies, %d weapons and %d jockeys", cases, len(species), len(templates), biomes, camelTimes, babies, weapons, jockeys)
}
