package server

import (
	"atlas/internal/nbt"
	"atlas/internal/world"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGeneratedVillageActorsAgainstVanillaJAR(t *testing.T) {
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
	var entity world.WorldEntity
	metadata, loaded, equipment, types := 0, 0, 0, 0
	for _, line := range strings.Split(string(raw), "\n") {
		f := strings.Fields(line)
		if len(f) == 3 && f[0] == "village_actor_type" {
			id, err := strconv.Atoi(f[2])
			if err != nil || generatedEntityTypes[f[1]] != int32(id) {
				t.Fatal("village entity registry mismatch", line, err)
			}
			types++
			continue
		}
		if len(f) != 4 {
			continue
		}
		switch f[0] {
		case "village_actor_default", "village_actor_support_default", "village_actor_finalized", "village_actor_camel_time":
			data, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, used, err := nbt.ParseNetwork(data)
			if err != nil || used != len(data) {
				t.Fatal(err)
			}
			entity.Data, _ = tag.Object()
			entity.Type = f[1]
		case "village_actor_default_metadata", "village_actor_metadata", "village_actor_camel_time_metadata", "village_actor_loaded_metadata":
			want, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			entity.LoadedFromDisk = f[0] == "village_actor_loaded_metadata"
			if got := generatedEntityMetadata(77, entity); !bytes.Equal(got, want) {
				t.Fatalf("%s got%x want%x", line, got, want)
			}
			if entity.LoadedFromDisk {
				loaded++
			} else {
				metadata++
			}
		case "village_actor_equipment":
			want, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			got, err := generatedEntityEquipment(77, entity)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("%s equipment got%x want%x err%v", line, got, want, err)
			}
			equipment++
		}
	}
	if metadata != 7910 || loaded != 7910 || types != 11 || equipment < 48 {
		t.Fatal("incomplete village protocol coverage", metadata, loaded, equipment, types)
	}
	t.Logf("matched %d fresh and %d saved/reloaded original village actor metadata, %d complete equipment packets and %d registry IDs", metadata, loaded, equipment, types)
}
