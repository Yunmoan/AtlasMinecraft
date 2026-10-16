package worldgen

import (
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Compare the complete saved actor after adding the Go finalization to an
// official constructor snapshot. This catches subtype side effects such as a
// baby goat's attack attribute, in addition to fields returned by finalization.
func TestCompleteInitialAnimalSavedStateAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/other_initial_animals_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	parseNBT := func(s string) nbt.Object {
		raw, err := hex.DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		tag, _, err := nbt.ParseNetwork(raw)
		if err != nil {
			t.Fatal(err)
		}
		o, _ := tag.Object()
		return o
	}
	defaults := map[string]string{}
	var group initialAnimalGroup
	var random *Xoroshiro
	var data nbt.Object
	var description string
	count := 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "animal_default":
			defaults[f[1]] = f[3]
		case "animal_finalized_case":
			seed, _ := strconv.ParseInt(f[3], 10, 64)
			if f[4] == "0" {
				group = initialAnimalGroup{}
				random = NewXoroshiro(seed)
			}
			data, err = finalizeInitialAnimalAt(f[1], f[2], FeaturePosition{-16, 63, 16}, random, &group)
			if err != nil {
				t.Fatal(err)
			}
			description = line
		case "animal_finalized":
			got, want := parseNBT(defaults[f[1]]), parseNBT(f[3])
			for key, value := range data {
				if key != "attributes" {
					got[key] = value
					continue
				}
				base, _ := got[key].Elements()
				added, _ := value.Elements()
				for _, a := range added {
					o, _ := a.Object()
					id, _ := o.Text("id")
					replaced := false
					for i, b := range base {
						bo, _ := b.Object()
						bid, _ := bo.Text("id")
						if bid == id {
							base[i] = a
							replaced = true
							break
						}
					}
					if !replaced {
						base = append(base, a)
					}
				}
				got[key] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: base}
			}
			for _, object := range []nbt.Object{got, want} {
				delete(object, "scute_time") // Independent constructor RNG, not either worldgen stream.
				values, _ := object["attributes"].Elements()
				sort.Slice(values, func(i, j int) bool {
					a, _ := values[i].Object()
					b, _ := values[j].Object()
					an, _ := a.Text("id")
					bn, _ := b.Text("id")
					return an < bn
				})
			}
			g := iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: got})
			w := iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: want})
			if g != w {
				for key, value := range want {
					if iglooCanonicalNBT(t, value) != iglooCanonicalNBT(t, got[key]) {
						t.Errorf("%s field%s got%s want%s", description, key, iglooCanonicalNBT(t, got[key]), iglooCanonicalNBT(t, value))
					}
				}
				t.FailNow()
			}
			count++
		}
	}
	if count != 2016 {
		t.Fatal(count)
	}
	t.Logf("matched complete saved state of %d original subtype finalizations, including inherited attribute side effects", count)
}
