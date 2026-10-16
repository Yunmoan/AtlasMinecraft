package light

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

// The fixture contains original block states, actual checkBlock edits, and
// every original SKY/BLOCK sample in all nine columns after each update.
func TestDynamicLightUpdatesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/light_updates_1.21.1.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	integer := func(value string) int {
		t.Helper()
		n, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	decode := func(value string) []byte {
		t.Helper()
		data, err := hex.DecodeString(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	var engine *Engine
	var name string
	var area map[Position][]uint16
	var cases, edits, arrays int
	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 65536), 128*1024)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		switch f[0] {
		case "light_update_case":
			name = f[1]
			engine, err = New(0, 32, f[2] == "true")
			if err != nil {
				t.Fatal(err)
			}
			area = make(map[Position][]uint16)
			cases++
		case "light_update_states":
			data := decode(f[3])
			if len(data) != 32*256*2 {
				t.Fatal("invalid original block array")
			}
			values := make([]uint16, len(data)/2)
			for i := range values {
				values[i] = binary.BigEndian.Uint16(data[2*i:])
			}
			area[Position{int32(integer(f[1])), int32(integer(f[2]))}] = values
			if len(area) == 9 {
				if _, err := engine.InitializeArea(Position{}, area); err != nil {
					t.Fatal(err)
				}
			}
		case "light_update_edit":
			x, y, z, id := integer(f[2]), integer(f[3]), integer(f[4]), integer(f[5])
			if _, err := engine.Edit(Position{int32(x >> 4), int32(z >> 4)}, []Change{{x & 15, y, z & 15, uint16(id)}}); err != nil {
				t.Fatal(name, err)
			}
			edits++
		case "light_update":
			data := engine.Snapshot(Position{int32(integer(f[2])), int32(integer(f[3]))})
			want := decode(f[5])
			if len(want) != 32*256 || data == nil {
				t.Fatal("missing full original light array")
			}
			for i, expected := range want {
				x, y, z := i&15, i>>8, i>>4&15
				if got := level(data, f[4] == "SKY", x, y, z, 0); got != expected {
					t.Fatalf("%s step %s chunk %s,%s %s at %d,%d,%d: got %d want original %d", name, f[1], f[2], f[3], f[4], x, y, z, got, expected)
				}
			}
			arrays++
		default:
			t.Fatal("unknown original fixture row", f[0])
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 9 || edits != 21 || arrays != 540 {
		t.Fatal("incomplete dynamic original coverage", cases, edits, arrays)
	}
}
