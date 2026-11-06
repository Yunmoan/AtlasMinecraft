package worldgen

import (
	"atlas/internal/nbt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMansionStartsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/structure_mansion_starts_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	noise := make(map[int64]*OverworldNoiseGenerator)
	heights := make(map[int64]map[FeaturePosition]int32)
	cases, accepted := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		if len(f) != 8 || f[0] != "mansion_start" {
			t.Fatal("invalid mansion start vector")
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		seed, cx, cz := parse(1), int32(parse(2)), int32(parse(3))
		g := noise[seed]
		if g == nil {
			g, err = NewOverworldNoiseGenerator(seed)
			if err != nil {
				t.Fatal(err)
			}
			noise[seed] = g
			heights[seed] = make(map[FeaturePosition]int32)
		}
		base := NewLegacyRandom(seed)
		xm, zm := base.NextLong(), base.NextLong()
		random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
		start, err := generateMansionStructureStart(cx, cz, func(_, _, _ int32) (string, error) { return f[4], nil }, func(x, z int32) (int32, error) {
			key := FeaturePosition{X: x, Z: z}
			if h, ok := heights[seed][key]; ok {
				return h, nil
			}
			column, err := g.BaseStructureColumnAt(x, z)
			if err != nil {
				return 0, err
			}
			h := column.OccupiedHeight(false)
			heights[seed][key] = h
			return h, nil
		}, random)
		if err != nil {
			t.Fatal(err)
		}
		count, hash := 0, "-"
		if start != nil {
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
		cases++
	}
	if cases != 224 || accepted == 0 {
		t.Fatal("incomplete start coverage", cases, accepted)
	}
	t.Logf("matched %d original mansion starts (%d accepted), full NBT and RNG continuation", cases, accepted)
}
