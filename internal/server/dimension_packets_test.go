package server

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/registry"
	"bufio"
	"bytes"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestDimensionPacketsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/protocol/dimension_packets_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	expected := map[string][]byte{}
	ids := map[string]int32{}
	cases := 0
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		f := strings.Fields(scan.Text())
		if f[0] == "dimension_registry" {
			v, err := strconv.ParseInt(f[2], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			ids[f[1]] = int32(v)
			continue
		}
		if len(f) != 7 || f[0] != "dimension_respawn" {
			t.Fatal(scan.Text())
		}
		cases++
		if f[2] != "1234" {
			continue
		}
		data, err := hex.DecodeString(f[6])
		if err != nil {
			t.Fatal(err)
		}
		expected[strings.Join([]string{f[1], f[3], f[4]}, "/")] = data
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 48 || len(ids) != 4 {
		t.Fatal("incomplete actual-JAR respawn/config registry fixture")
	}
	var registryNames []string
	for _, payload := range registry.Packets() {
		r := proto.NewReader(payload)
		name, err := r.String(128)
		if err != nil {
			t.Fatal(err)
		}
		if name != "minecraft:dimension_type" {
			continue
		}
		count, err := r.VarInt()
		if err != nil {
			t.Fatal(err)
		}
		for i := int32(0); i < count; i++ {
			name, err := r.String(128)
			if err != nil {
				t.Fatal(err)
			}
			registryNames = append(registryNames, name)
			if ids[name] != i {
				t.Fatalf("actual configured registry %s index%d differs from JAR-decoded registry %d", name, i, ids[name])
			}
			present, err := r.Bool()
			if err != nil || !present {
				t.Fatal("missing dimension type data")
			}
			remaining, _ := r.Bytes(r.Remaining())
			_, used, err := nbt.ParseNetwork(remaining)
			if err != nil {
				t.Fatal(err)
			}
			r = proto.NewReader(remaining[used:])
		}
		if r.Remaining() != 0 {
			t.Fatal("dimension registry trailing bytes")
		}
	}
	if len(registryNames) != 4 {
		t.Fatal("dimension type registry missing")
	}
	s, p, _ := dimensionTestServer(t)
	for _, w := range s.Worlds() {
		name := w.DimensionName()
		if dimensionTypeID(name) != ids[name] {
			t.Fatal("runtime dimension type ID differs from actual configuration")
		}
		p.SetInitialDimension(name)
		flat := strconv.FormatBool(w.IsFlat())
		want := expected[name+"/"+flat+"/1"]
		if got := s.respawnPayload(p, w); !bytes.Equal(got, want) {
			t.Fatalf("%s respawn differs from JAR: got%x want%x", name, got, want)
		}
		join := s.joinGamePayload(p)
		r := proto.NewReader(join)
		r.I32()
		r.Bool()
		count, _ := r.VarInt()
		for i := int32(0); i < count; i++ {
			r.String(128)
		}
		r.VarInt()
		r.VarInt()
		r.VarInt()
		r.Bool()
		r.Bool()
		r.Bool()
		spawn, _ := r.Bytes(r.Remaining() - 1)
		spawn = append(append([]byte(nil), spawn...), 3)
		want = expected[name+"/"+flat+"/-1"]
		if !bytes.Equal(spawn, want) {
			t.Fatalf("%s login CommonPlayerSpawnInfo differs from JAR: got%x want%x", name, spawn, want)
		}
	}
	t.Log("actual JAR decoded all four Atlas dimension type entries and validated three Login/Respawn CommonPlayerSpawnInfo payloads")
}
