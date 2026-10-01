package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type ruleDiskWorld struct {
	dirtDiskWorld
	scenario int
	marked   []FeaturePosition
}

func (w *ruleDiskWorld) MarkPostProcessing(p FeaturePosition) error {
	w.marked = append(w.marked, p)
	return nil
}

func (w *ruleDiskWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if s, ok := w.changed[p]; ok {
		return s, nil
	}
	if w.scenario == 1 && p.Y == 39 {
		return SurfaceBlockState{Name: "minecraft:stone"}, nil
	}
	if w.scenario == 2 && p.Y == 41 {
		return SurfaceBlockState{Name: "minecraft:water"}, nil
	}
	return w.dirtDiskWorld.BlockAt(p)
}

func TestRuleDisksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/rule_disk_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cases := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 13 || f[0] != "disk_rule_case" {
			t.Fatal(scanner.Text())
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		config, err := LoadDiskConfiguration(f[1])
		if err != nil {
			t.Fatal(err)
		}
		w := &ruleDiskWorld{dirtDiskWorld: dirtDiskWorld{changed: map[FeaturePosition]SurfaceBlockState{}}, scenario: int(parse(3))}
		random := NewWorldgenRandom(parse(2))
		placed, err := PlaceDiskFeature(w, random, FeaturePosition{int32(parse(4)), int32(parse(5)), int32(parse(6))}, config)
		if err != nil {
			t.Fatal(err)
		}
		var records []string
		for p, state := range w.changed {
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
			records = append(records, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
		}
		sort.Strings(records)
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(records, "\n"))))
		next := int64(random.NextLong())
		if placed != (parse(7) != 0) || len(records) != int(parse(8)) || next != parse(9) || hash != f[10] {
			t.Fatalf("%s: placed=%t count=%d next=%d hash=%s", scanner.Text(), placed, len(records), next, hash)
		}
		var marked []string
		for _, p := range w.marked {
			packed := (p.X & 15) | ((p.Y & 15) << 4) | ((p.Z & 15) << 8)
			marked = append(marked, fmt.Sprintf("%d %d", (p.Y+64)>>4, packed))
		}
		sort.Strings(marked)
		postHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(marked, "\n"))))
		if len(marked) != int(parse(11)) || postHash != f[12] {
			t.Fatalf("%s: postprocessing count=%d hash=%s", scanner.Text(), len(marked), postHash)
		}
		cases++
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 96 {
		t.Fatal(cases)
	}
}
