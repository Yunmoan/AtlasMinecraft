package worldgen

import (
	"atlas/internal/nbt"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type oracleAnimalWorld struct{ mode int }

func (w oracleAnimalWorld) SpawnPosition(_ string, x, z int32) (FeaturePosition, error) {
	return FeaturePosition{x, 64, z}, nil
}
func (w oracleAnimalWorld) PositionOK(_ string, p FeaturePosition) (bool, error) {
	return w.mode != 1 || p.X >= -24, nil
}
func (w oracleAnimalWorld) NoCollision(_ string, _ [3]float64) (bool, error) { return w.mode != 2, nil }
func (w oracleAnimalWorld) SpawnRules(_ string, _ FeaturePosition, _ *Xoroshiro) (bool, error) {
	return w.mode != 3, nil
}
func (w oracleAnimalWorld) MobSpawnOK(_ string, _ [3]float64) (bool, error) { return w.mode != 4, nil }

func TestInitialAnimalSpawnAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile(filepath.Join("..", "..", "docs", "worldgen", "initial_animals_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(file)), "\n")
	cases, entities := 0, 0
	parse := func(s string) int64 {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	for i := 0; i < len(lines); i++ {
		f := strings.Fields(lines[i])
		if len(f) == 0 || f[0] != "animal_pack" {
			continue
		}
		if len(f) != 9 {
			t.Fatal(lines[i])
		}
		seed := parse(f[2])
		random, worldRandom := NewLegacyRandom(seed), NewXoroshiro(seed^0x123456789abcdef)
		animals, unsupported, err := SpawnInitialAnimals(f[1], int32(parse(f[3])), int32(parse(f[4])), random, worldRandom, oracleAnimalWorld{int(parse(f[5]))})
		if err != nil || len(unsupported) > 0 {
			t.Fatalf("%s: %v %v", lines[i], err, unsupported)
		}
		gotNext, gotWorld := int64(random.NextLong()), int64(worldRandom.NextLong())
		if len(animals) != int(parse(f[6])) || gotNext != parse(f[7]) || gotWorld != parse(f[8]) {
			t.Fatalf("%s: got count=%d next=%d worldNext=%d", lines[i], len(animals), gotNext, gotWorld)
		}
		for _, animal := range animals {
			i++
			fixture := strings.Fields(lines[i])
			if len(fixture) != 4 || fixture[0] != "animal_nbt" {
				t.Fatal("missing animal NBT")
			}
			data, err := hex.DecodeString(fixture[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(data)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			id, _ := want.Text("id")
			if animal.ID != id {
				t.Fatalf("species %s != %s", animal.ID, id)
			}
			for key, value := range animal.Data {
				if key != "attributes" {
					if got, expected := iglooCanonicalNBT(t, value), iglooCanonicalNBT(t, want[key]); got != expected {
						t.Fatalf("seed%d %s %s: %s != %s", seed, id, key, got, expected)
					}
					continue
				}
				expected, _ := want[key].Elements()
				actual, _ := value.Elements()
				for _, a := range actual {
					o, _ := a.Object()
					attributeID, _ := o.Text("id")
					found := false
					for _, e := range expected {
						eo, _ := e.Object()
						eid, _ := eo.Text("id")
						if eid == attributeID {
							found = true
							if iglooCanonicalNBT(t, a) != iglooCanonicalNBT(t, e) {
								t.Fatalf("seed%d %s %s: %s != %s", seed, id, attributeID, iglooCanonicalNBT(t, a), iglooCanonicalNBT(t, e))
							}
						}
					}
					if !found {
						t.Fatal("missing official attribute", attributeID)
					}
				}
			}
			entities++
		}
		cases++
	}
	if cases != 1920 || entities == 0 {
		t.Fatal(cases, entities)
	}
	t.Logf("matched %d actual NaturalSpawner packs and %d actual saved animals, including separate placement/world RNG and collision/rule/obstruction branches", cases, entities)
}

func TestInitialAnimalStateRulesAgainstVanillaJAR(t *testing.T) {
	tables, err := loadInitialAnimalTables()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "worldgen", "initial_animals_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		var id, ground, space int
		if n, _ := fmt.Sscanf(line, "animal_state %d %d %d", &id, &ground, &space); n == 3 {
			if tables.States[id][0]&63 != uint32(ground) || tables.States[id][1]&63 != uint32(space) {
				t.Fatal(line)
			}
			count++
		}
	}
	if count != 26684 {
		t.Fatal(count)
	}
	t.Logf("matched %d original ground/space rules for six entity types", count)
}

func TestOtherInitialAnimalFinalizationAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/other_initial_animals_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var group initialAnimalGroup
	var random *Xoroshiro
	var got nbt.Object
	var description string
	cases, nexts := 0, 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "animal_finalized_case":
			seed, err := strconv.ParseInt(f[3], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			if f[4] == "0" {
				group = initialAnimalGroup{}
				random = NewXoroshiro(seed)
			}
			got, err = finalizeInitialAnimalAt(f[1], f[2], FeaturePosition{-16, 63, 16}, random, &group)
			if err != nil {
				t.Fatal(err)
			}
			description = line
		case "animal_finalized":
			data, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(data)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			for key, value := range got {
				if key != "attributes" {
					if iglooCanonicalNBT(t, value) != iglooCanonicalNBT(t, want[key]) {
						t.Fatalf("%s %s: got %s want %s", description, key, iglooCanonicalNBT(t, value), iglooCanonicalNBT(t, want[key]))
					}
					continue
				}
				actual, _ := value.Elements()
				expected, _ := want[key].Elements()
				for _, a := range actual {
					o, _ := a.Object()
					id, _ := o.Text("id")
					found := false
					for _, e := range expected {
						eo, _ := e.Object()
						eid, _ := eo.Text("id")
						if eid == id {
							found = true
							if iglooCanonicalNBT(t, a) != iglooCanonicalNBT(t, e) {
								t.Fatalf("%s %s: got %s want %s", description, id, iglooCanonicalNBT(t, a), iglooCanonicalNBT(t, e))
							}
						}
					}
					if !found {
						t.Fatal("missing official attribute", description, id)
					}
				}
			}
			cases++
		case "animal_finalized_next":
			want, err := strconv.ParseInt(f[4], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			if got := int64(random.NextLong()); got != want {
				t.Fatalf("%s: got next %d", line, got)
			}
			nexts++
		}
	}
	if cases != 2016 || nexts != 504 {
		t.Fatal(cases, nexts)
	}
	t.Logf("matched %d actual subtype finalizations and %d independent next world RNG states", cases, nexts)
}

func TestOtherInitialAnimalStateRulesAgainstVanillaJAR(t *testing.T) {
	tables, err := loadInitialAnimalTables()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.ReadFile("../../docs/worldgen/other_initial_animals_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(file), "\n") {
		var id, ground, space uint32
		if n, _ := fmt.Sscanf(line, "animal_state %d %d %d", &id, &ground, &space); n == 3 {
			if tables.States[id][0]>>6 != ground || tables.States[id][1]>>6 != space {
				t.Fatal(line)
			}
			count++
		}
	}
	if count != 26684 {
		t.Fatal(count)
	}
	t.Logf("matched %d original ground/space rules for twelve further entity types", count)
}
