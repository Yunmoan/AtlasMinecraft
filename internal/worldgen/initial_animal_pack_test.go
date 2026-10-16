package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

type otherOracleAnimalWorld struct {
	oracleAnimalWorld
	biome, ground string
	tables        *initialAnimalTables
	groundID      uint32
}

func (w otherOracleAnimalWorld) PositionOK(id string, p FeaturePosition) (bool, error) {
	if id == "minecraft:fox" || id == "minecraft:panda" {
		return true, nil
	}
	bit := initialAnimalBit(id)
	if w.mode == 1 && p.X < -24 && w.tables.States[w.groundID][1]&bit == 0 {
		return false, nil
	}
	return w.tables.States[w.groundID][0]&bit != 0, nil
}
func (w otherOracleAnimalWorld) SpawnRules(id string, p FeaturePosition, _ *Xoroshiro) (bool, error) {
	if w.mode == 3 {
		return false, nil
	}
	return animalSpawnGroundAllowed(id, w.biome, w.ground, p.Y, 63)
}
func (w otherOracleAnimalWorld) AnimalBiomeAt(FeaturePosition) (string, error) { return w.biome, nil }

func compareAnimalFields(t *testing.T, description string, got, want nbt.Object) {
	t.Helper()
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
}

func TestAllInitialAnimalPacksAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/other_initial_animals_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	tables, err := loadInitialAnimalTables()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(file), "\n")
	parse := func(s string) int64 {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	cases, entities := 0, 0
	kinds := map[string]int{}
	for i := 0; i < len(lines); i++ {
		f := strings.Fields(lines[i])
		if len(f) == 0 || f[0] != "other_animal_pack" {
			continue
		}
		if len(f) != 10 {
			t.Fatal(lines[i])
		}
		seed := parse(f[3])
		placement, worldRandom := NewLegacyRandom(seed), NewXoroshiro(seed^0x123456789abcdef)
		groundID, err := blockstate.BlockStateID(f[2], nil)
		if err != nil {
			t.Fatal(err)
		}
		w := otherOracleAnimalWorld{oracleAnimalWorld{int(parse(f[6]))}, f[1], f[2], tables, groundID}
		animals, unsupported, err := SpawnInitialAnimals(f[1], int32(parse(f[4])), int32(parse(f[5])), placement, worldRandom, w)
		if err != nil || len(unsupported) > 0 {
			t.Fatal(lines[i], err, unsupported)
		}
		gotNext, gotWorld := int64(placement.NextLong()), int64(worldRandom.NextLong())
		if len(animals) != int(parse(f[7])) || gotNext != parse(f[8]) || gotWorld != parse(f[9]) {
			ids := make([]string, len(animals))
			for i, a := range animals {
				ids[i] = a.ID
			}
			t.Fatalf("%s got %d animals next%d world%d IDs%v", lines[i], len(animals), gotNext, gotWorld, ids)
		}
		for _, animal := range animals {
			i++
			fields := strings.Fields(lines[i])
			if len(fields) != 4 || fields[0] != "other_animal_nbt" {
				t.Fatal("missing animal NBT", lines[i])
			}
			encoded, err := hex.DecodeString(fields[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(encoded)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			id, _ := want.Text("id")
			if animal.ID != id {
				t.Fatal(animal.ID, id)
			}
			compareAnimalFields(t, lines[i], animal.Data, want)
			kinds[id]++
			entities++
		}
		cases++
	}
	if cases != 21760 || len(kinds) != 18 {
		t.Fatal(cases, entities, kinds)
	}
	t.Logf("matched %d actual NaturalSpawner packs, %d animals across all %d overworld creature species, original surfaces and five rejection scenarios", cases, entities, len(kinds))
}

func TestInitialAnimalPlacementLowerAgainstVanillaJAR(t *testing.T) {
	tables, err := loadOtherInitialAnimalTables()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.ReadFile("../../docs/worldgen/other_initial_animals_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) != 3 || f[0] != "animal_path" {
			continue
		}
		id, _ := strconv.Atoi(f[1])
		if tables.PlacementLower[id] != (f[2] == "true") {
			t.Fatal(line)
		}
		count++
	}
	if count != 26684 {
		t.Fatal(count)
	}
	t.Logf("matched all %d original ON_GROUND placement lowering rules", count)
}
