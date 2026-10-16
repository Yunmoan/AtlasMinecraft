package worldgen

import (
	"atlas/internal/nbt"
	"crypto/sha256"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestEndCityStartsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/structure_end_city_starts_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	noise := make(map[int64]*DimensionNoiseGenerator)
	heights := make(map[int64]map[FeaturePosition]int32)
	layouts, starts, accepted, totalPieces := 0, 0, 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(fields[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		seed := parse(1)
		switch fields[0] {
		case "end_city_layout":
			random := NewLegacyRandom(seed)
			pieces, err := GenerateEndCityPieces(FeaturePosition{int32(parse(3)), int32(parse(4)), int32(parse(5))}, TemplateRotation(parse(2)), random)
			if err != nil {
				t.Fatal(err)
			}
			var saved strings.Builder
			for _, piece := range pieces {
				object, err := piece.NBT()
				if err != nil {
					t.Fatal(err)
				}
				saved.WriteString(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: object}))
				saved.WriteByte('\n')
			}
			hash := stringHexHash(saved.String())
			if len(pieces) != int(parse(6)) || hash != fields[7] || int64(random.NextLong()) != parse(8) {
				t.Fatalf("%s: got%d pieces hash%s; canonical%s", line, len(pieces), hash, saved.String())
			}
			layouts++
			totalPieces += len(pieces)
		case "end_city_start":
			g := noise[seed]
			if g == nil {
				g, err = NewDimensionNoiseGenerator(seed, "end")
				if err != nil {
					t.Fatal(err)
				}
				noise[seed] = g
				heights[seed] = make(map[FeaturePosition]int32)
			}
			cx, cz := int32(parse(2)), int32(parse(3))
			base := NewLegacyRandom(seed)
			xm, zm := base.NextLong(), base.NextLong()
			random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
			start, err := generateEndCityStructureStart(cx, cz, func(_, _, _ int32) (string, error) { return fields[4], nil }, func(x, z int32) (int32, error) {
				key := FeaturePosition{X: x, Z: z}
				if height, found := heights[seed][key]; found {
					return height, nil
				}
				column, err := g.BaseStructureColumnAt(x, z)
				if err != nil {
					return 0, err
				}
				height := column.OccupiedHeight(false)
				heights[seed][key] = height
				return height, nil
			}, random)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			hash := "-"
			if start != nil {
				count = len(start.Pieces)
				object, err := start.NBT()
				if err != nil {
					t.Fatal(err)
				}
				hash = stringHexHash(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: object}))
				accepted++
			}
			if count != int(parse(5)) || hash != fields[6] || int64(random.NextLong()) != parse(7) {
				t.Fatalf("%s: got%d pieces hash%s", line, count, hash)
			}
			starts++
		default:
			t.Fatal("unknown End City fixture", line)
		}
	}
	if layouts != 112 || starts != 224 || accepted == 0 {
		t.Fatalf("got%d layouts%d starts%d accepted", layouts, starts, accepted)
	}
	t.Logf("matched%d original layouts/%d pieces and%d actual full starts (%d accepted), full NBT and RNG", layouts, totalPieces, starts, accepted)
}

func stringHexHash(text string) string {
	const digits = "0123456789abcdef"
	hash := sha256.Sum256([]byte(text))
	result := make([]byte, 64)
	for index, value := range hash {
		result[index*2] = digits[value>>4]
		result[index*2+1] = digits[value&15]
	}
	return string(result)
}
