package worldgen

import (
	"atlas/internal/nbt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestStrongholdStartsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/structure_stronghold_starts_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	layouts, starts, accepted, totalPieces, retries := 0, 0, 0, 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		seed, cx, cz := parse(1), int32(parse(2)), int32(parse(3))
		switch f[0] {
		case "stronghold_layout":
			pieces, random, attempts := assembleStronghold(seed, cx, cz, int32(parse(4)), int32(parse(5)))
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
			if attempts != int32(parse(6)) || len(pieces) != int(parse(7)) || hash != f[8] || next != parse(9) {
				t.Fatalf("%s: got %d attempts %d pieces hash %s next %d; saved %s", line, attempts, len(pieces), hash, next, saved.String())
			}
			layouts++
			totalPieces += len(pieces)
			if attempts > 1 {
				retries++
			}
		case "stronghold_start":
			start, err := GenerateStrongholdStructureStart(seed, cx, cz, StrongholdStartSources{-64, 63, func(x, y, z int32) (string, error) {
				if x != cx*4 || y != 0 || z != cz*4 {
					t.Fatalf("wrong generation-stub biome %d,%d,%d", x, y, z)
				}
				return f[4], nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			random := strongholdLargeFeatureRandom(seed, cx, cz)
			count, hash := 0, "-"
			if start != nil {
				_, random, _ = assembleStronghold(seed, cx, cz, -64, 63)
				count = len(start.Pieces)
				tag, err := start.NBT()
				if err != nil {
					t.Fatal(err)
				}
				hash = stringHexHash(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: tag}))
				accepted++
			}
			next := int64(random.NextLong())
			if count != int(parse(5)) || hash != f[6] || next != parse(7) {
				t.Fatalf("%s: got %d pieces hash %s next %d", line, count, hash, next)
			}
			starts++
		default:
			t.Fatal("unknown stronghold fixture", line)
		}
	}
	if layouts != 126 || starts != 126 || accepted != 63 || retries == 0 {
		t.Fatalf("got %d layouts %d starts %d accepted %d retries", layouts, starts, accepted, retries)
	}
	t.Logf("matched %d original layouts / %d pieces (%d retried), %d actual full starts (%d accepted), complete NBT and RNG", layouts, totalPieces, retries, starts, accepted)
}
