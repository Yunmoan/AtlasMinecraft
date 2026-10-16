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

func TestGeneratedVehiclePacketsAgainstVanillaJAR(t *testing.T) {
	file, err := os.ReadFile("../../docs/worldgen/generated_vehicles_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var entity world.WorldEntity
	types, metadata, passengerPackets := 0, 0, 0
	for _, line := range strings.Split(string(file), "\n") {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		switch f[0] {
		case "vehicle_type":
			want, _ := strconv.ParseInt(f[2], 10, 32)
			got, ok := generatedEntityTypes[f[1]]
			if !ok || got != int32(want) {
				t.Fatal(line, got)
			}
			types++
		case "vehicle_default", "vehicle_saved":
			raw, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			entity.Type = f[1]
			entity.Data, _ = tag.Object()
		case "vehicle_metadata":
			want, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			if got := generatedEntityMetadata(77, entity); !bytes.Equal(got, want) {
				t.Fatalf("%s got%x", line, got)
			}
			metadata++
		case "vehicle_passengers":
			vehicle, _ := strconv.ParseInt(f[1], 10, 32)
			var ids []int32
			if f[2] != "-" {
				for _, value := range strings.Split(f[2], ",") {
					id, _ := strconv.ParseInt(value, 10, 32)
					ids = append(ids, int32(id))
				}
			}
			want, err := hex.DecodeString(f[3])
			if err != nil {
				t.Fatal(err)
			}
			if got := generatedPassengersPayload(int32(vehicle), ids); !bytes.Equal(got, want) {
				t.Fatalf("%s got%x", line, got)
			}
			passengerPackets++
		}
	}
	if types != 1 || metadata != 4 || passengerPackets != 12 {
		t.Fatal(types, metadata, passengerPackets)
	}
	t.Logf("matched chest minecart registry ID, %d actual saved metadata codecs and %d passenger packet codecs", metadata, passengerPackets)
}
