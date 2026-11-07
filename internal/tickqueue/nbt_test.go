package tickqueue

import (
	"atlas/internal/nbt"
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSavedTickNBTAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/saved_tick_nbt_1.21.1.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	decode := func(encoded string) nbt.Tag {
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatal(err)
		}
		_, tag, err := nbt.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		object, _ := tag.Object()
		return object["fluid_ticks"]
	}
	resolve := func(name string) (string, bool) {
		if !strings.Contains(name, ":") {
			name = "minecraft:" + name
		} else if strings.HasPrefix(name, ":") {
			name = "minecraft" + name
		}
		switch name {
		case "minecraft:empty", "minecraft:water", "minecraft:flowing_water", "minecraft:lava", "minecraft:flowing_lava":
			return name, true
		}
		return "", false
	}
	scan := bufio.NewScanner(z)
	cases := 0
	for scan.Scan() {
		var row struct {
			Case                 int
			X, Z                 int32
			Input, Saved, Active string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		cases++
		t.Run(fmt.Sprintf("case%d/chunk%d,%d", row.Case, row.X, row.Z), func(t *testing.T) {
			c := NewChunk(DecodeSavedNBT(decode(row.Input), row.X, row.Z, resolve))
			name := func(value string) string { return value }
			got := EncodeSavedNBT(c.Save(0), name)
			if !reflect.DeepEqual(got, decode(row.Saved)) {
				t.Fatalf("saved NBT differs: got %+v want %+v", got, decode(row.Saved))
			}
			c.Unpack(100)
			got = EncodeSavedNBT(c.Save(104), name)
			if !reflect.DeepEqual(got, decode(row.Active)) {
				t.Fatalf("active NBT differs: got %+v want %+v", got, decode(row.Active))
			}
		})
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 126 {
		t.Fatalf("got %d native NBT scenarios", cases)
	}
}
