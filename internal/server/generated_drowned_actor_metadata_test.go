package server

import (
	"atlas/internal/nbt"
	"atlas/internal/world"
	"bytes"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGeneratedDrownedActorMetadataAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/structure_drowned_actors_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var entity world.WorldEntity
	count, types := 0, 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		if f[0] == "drowned_actor_type" {
			id, _ := strconv.ParseInt(f[2], 10, 32)
			if got := generatedEntityTypes[f[1]]; got != int32(id) {
				t.Fatal(line, got)
			}
			types++
			continue
		}
		if len(f) != 4 {
			continue
		}
		if f[0] == "drowned_actor_default" || f[0] == "drowned_actor_finalized" {
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
		if f[0] == "drowned_actor_default_metadata" || f[0] == "drowned_actor_finalized_metadata" {
			want, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			got := generatedMobMetadata(77, entity)
			if !bytes.Equal(got, want) {
				t.Fatalf("%s got%x want%x", line, got, want)
			}
			count++
		}
	}
	if count != 4097 || types != 1 {
		t.Fatal(count, types)
	}
	t.Logf("matched %d actual original Drowned metadata packet codecs and official registry entity ID", count)
}
