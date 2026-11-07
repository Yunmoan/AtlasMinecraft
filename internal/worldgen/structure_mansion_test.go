package worldgen

import (
	"atlas/internal/nbt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMansionLayoutsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/structure_mansion_layout_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases, total := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		if len(f) != 9 || f[0] != "mansion_layout" {
			t.Fatal("invalid mansion layout vector")
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		rotation := TemplateRotation(255)
		for i, name := range templateRotationNames {
			if name == f[5] {
				rotation = TemplateRotation(i)
			}
		}
		if rotation > 3 {
			t.Fatal("unknown rotation", f[5])
		}
		random := NewLegacyRandom(parse(1))
		pieces, err := assembleMansion(random, FeaturePosition{int32(parse(2)), int32(parse(3)), int32(parse(4))}, rotation)
		if err != nil {
			t.Fatal(err)
		}
		var saved strings.Builder
		for _, piece := range pieces {
			tag, err := piece.NBT()
			if err != nil {
				t.Fatal(err)
			}
			saved.WriteString(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: tag}))
			saved.WriteByte('\n')
		}
		hash, next := stringHexHash(saved.String()), int64(random.NextLong())
		if len(pieces) != int(parse(6)) || hash != f[7] || next != parse(8) {
			t.Fatalf("%s: got %d pieces hash %s next %d\n%s", line, len(pieces), hash, next, saved.String())
		}
		cases++
		total += len(pieces)
	}
	if cases != 84 {
		t.Fatal("incomplete mansion layout corpus", cases)
	}
	t.Logf("matched %d complete mansion layouts, %d ordered template pieces and RNG continuation", cases, total)
}
