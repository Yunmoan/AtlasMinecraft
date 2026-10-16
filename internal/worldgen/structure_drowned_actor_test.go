package worldgen

import (
	"atlas/internal/nbt"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestDrownedStructureFinalizationAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/structure_drowned_actors_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var random *Xoroshiro
	var actor *LegacyRandom
	var got DrownedStructureResult
	cases, jockeys := 0, 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) != 4 && len(f) != 6 {
			continue
		}
		if f[0] == "drowned_actor_finalized" {
			seed, _ := strconv.ParseInt(f[2], 10, 64)
			random = NewXoroshiro(seed)
			actor = NewLegacyRandom(seed ^ 0x123456789abcdef)
			got, err = FinalizeDrownedStructureEntity(PendingStructureEntity{ID: "minecraft:drowned", Position: FeaturePosition{-16, 50, 16}}, random, actor, false)
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := hex.DecodeString(f[3])
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			delete(want, "UUID")
			compareAnimalFields(t, line, got.Data, want)
			cases++
		}
		if f[0] == "drowned_actor_next" {
			want, _ := strconv.ParseInt(f[3], 10, 64)
			wantActor, _ := strconv.ParseInt(f[4], 10, 64)
			next, nextActor := int64(random.NextLong()), int64(actor.NextLong())
			if next != want || nextActor != wantActor {
				t.Fatalf("%s next%d actor%d", line, next, nextActor)
			}
			mounts, _ := strconv.Atoi(f[5])
			if mounts != 0 {
				jockeys++
			}
			if got.ChickenJockey != (mounts != 0) {
				t.Fatal("chicken branch", line)
			}
		}
		if f[0] == "drowned_actor_extra" {
			raw, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			compareAnimalFields(t, "actual finalized chicken jockey", got.ChickenData, want)
		}
	}
	if cases != 4096 || jockeys != 7 {
		t.Fatal(cases, jockeys)
	}
	t.Logf("matched %d actual fresh NORMAL STRUCTURE Drowned finalizations with two independent next RNG states and %d chicken-jockey branches", cases, jockeys)
}
