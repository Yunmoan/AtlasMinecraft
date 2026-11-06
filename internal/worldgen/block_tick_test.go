package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strconv"
	"testing"
)

func TestLeafDistanceLookupAgainstVanilla(t *testing.T) {
	want, err := os.ReadFile("../../testdata/worldgen/leaf_distances_1.21.1.bin")
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != 26684 {
		t.Fatal("incomplete distance vector")
	}
	for id, distance := range want {
		s, err := blockstate.BlockStateByID(uint32(id))
		if err != nil {
			t.Fatal(err)
		}
		if got := scheduledLeafDistance(SurfaceBlockState{Name: s.Name, Properties: s.Properties}); got != int(distance) {
			t.Fatalf("state %d %s: distance %d want %d", id, s.Name, got, distance)
		}
	}
}

func TestLeafScheduledTicksAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/leaf_ticks_1.21.1.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	scan := bufio.NewScanner(gz)
	count := 0
	seen := map[[3]int64]bool{}
	for scan.Scan() {
		var row struct {
			State  uint32
			Scene  int
			Seed   int64
			Inputs [][4]int32
			Events []string
			Next   string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		key := [3]int64{int64(row.State), int64(row.Scene), row.Seed}
		if seen[key] {
			t.Fatal("duplicate leaf tick", key)
		}
		seen[key] = true
		w := &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{}, floor: -64}
		for _, input := range row.Inputs {
			s, err := blockstate.BlockStateByID(uint32(input[3]))
			if err != nil {
				t.Fatal(err)
			}
			w.blocks[FeaturePosition{input[0], input[1], input[2]}] = SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		}
		p := FeaturePosition{8, 64, 8}
		s, _ := w.BlockAt(p)
		if err := TickScheduledBlock(w, p, s.Name); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(w.events, row.Events) {
			t.Fatalf("state=%d scene=%d events=%v want=%v", row.State, row.Scene, w.events, row.Events)
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if uint64(NewLegacyRandom(row.Seed).NextLong()) != next {
			t.Fatal("native leaf tick consumed random", row.State, row.Scene)
		}
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 14280 {
		t.Fatal("incomplete leaf tick coverage", count)
	}
}

func TestScheduledBlockUnsupportedStaleAndNoop(t *testing.T) {
	p := FeaturePosition{8, 64, 8}
	w := &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: structureState("farmland")}, floor: 63}
	if err := TickScheduledBlock(w, p, "minecraft:oak_leaves"); err != nil {
		t.Fatal("stale block tick", err)
	}
	if err := TickScheduledBlock(w, p, "minecraft:farmland"); !errors.Is(err, ErrUnsupportedScheduledBlock) {
		t.Fatal("unsupported tick silently consumed", err)
	}
	w.blocks[p] = structureState("stone")
	if err := TickScheduledBlock(w, p, "minecraft:stone"); err != nil {
		t.Fatal("inherited empty handler", err)
	}
	if len(w.events) != 0 {
		t.Fatal("stale/noop tick edited blocks", w.events)
	}
}
