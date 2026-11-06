package worldgen

import (
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMansionActorsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/structure_mansion_actors_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var random *Xoroshiro
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		if f[0] == "mansion_actor_default" {
			continue
		}
		if len(f) != 6 || f[0] != "mansion_actor" {
			t.Fatalf("invalid actor vector %q", line)
		}
		seed, err := strconv.ParseInt(f[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if f[3] == "0" {
			random = NewXoroshiro(seed)
		}
		encoded, err := hex.DecodeString(f[4])
		if err != nil {
			t.Fatal(err)
		}
		tag, used, err := nbt.ParseNetwork(encoded)
		if err != nil || used != len(encoded) {
			t.Fatal("invalid actor NBT", err)
		}
		want, _ := tag.Object()
		got, err := FinalizeMansionEntity(PendingStructureEntity{ID: f[1], Position: FeaturePosition{0, 42, 0}}, random)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) {
			t.Fatalf("actor fields: got %d want %d", len(got), len(want))
		}
		for key, expected := range want {
			actual := got[key]
			if key == "attributes" {
				actual, expected = canonicalOutpostActorAttributes(actual), canonicalOutpostActorAttributes(expected)
			}
			if iglooCanonicalNBT(t, actual) != iglooCanonicalNBT(t, expected) {
				t.Fatalf("seed %d actor %s field %s: got %v want %v", seed, f[3], key, actual, expected)
			}
		}
		next, err := strconv.ParseInt(f[5], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		copy := *random
		if int64(copy.NextLong()) != next {
			t.Fatal("actor region RNG differs", line)
		}
		count++
	}
	if count != 54 {
		t.Fatalf("got %d actor cases, want 54", count)
	}
	t.Log("matched 54 original mansion actor saves and shared region RNG continuations")
}
