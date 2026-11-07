package world

import (
	"atlas/internal/nbt"
	"atlas/internal/tickqueue"
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestSavedBlockTickNBTAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/saved_block_tick_nbt_1.21.1.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	decode := func(encoded string) nbt.Tag {
		t.Helper()
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatal(err)
		}
		_, tag, err := nbt.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		root, _ := tag.Object()
		return root["block_ticks"]
	}
	scan := bufio.NewScanner(gz)
	count := 0
	for scan.Scan() {
		var row struct {
			Case                 int
			X, Z                 int32
			Input, Saved, Active string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		q := tickqueue.NewChunk(tickqueue.DecodeSavedNBT(decode(row.Input), row.X, row.Z, parseScheduledBlock))
		name := func(s string) string { return s }
		if got := tickqueue.EncodeSavedNBT(q.Save(0), name); !reflect.DeepEqual(got, decode(row.Saved)) {
			t.Fatalf("packed block ticks case %d chunk %d,%d differ", row.Case, row.X, row.Z)
		}
		q.Unpack(100)
		if got := tickqueue.EncodeSavedNBT(q.Save(104), name); !reflect.DeepEqual(got, decode(row.Active)) {
			t.Fatalf("unpacked block ticks case %d chunk %d,%d differ", row.Case, row.X, row.Z)
		}
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 126 {
		t.Fatal("incomplete native block queue NBT", count)
	}
}
