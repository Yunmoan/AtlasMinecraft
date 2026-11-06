package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"testing"
)

func TestBubbleScheduledTicksAndPlacementAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/bubble_ticks_1.21.1.jsonl.gz")
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
	for scan.Scan() {
		var row struct {
			State  uint32
			Scene  int
			Seed   int64
			Action string
			Inputs [][4]int32
			Events []string
			Next   string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		w := blockShapeProbeWorld{fluidStepProbeWorld: &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{}, floor: -64}}
		for _, input := range row.Inputs {
			s, err := blockstate.BlockStateByID(uint32(input[3]))
			if err != nil {
				t.Fatal(err)
			}
			w.blocks[FeaturePosition{input[0], input[1], input[2]}] = SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		}
		p := FeaturePosition{8, 64, 8}
		s, _ := w.BlockAt(p)
		if row.Action == "tick" {
			err = TickScheduledBlock(w, p, s.Name)
		} else if row.Action == "place" {
			err = NotifyBubbleBlockPlacement(w, p, s)
		} else {
			t.Fatal("unknown action", row.Action)
		}
		if err != nil || !slices.Equal(w.events, row.Events) {
			t.Fatalf("state=%d scene=%d action=%s events=%v want=%v err=%v", row.State, row.Scene, row.Action, w.events, row.Events, err)
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil || uint64(NewLegacyRandom(row.Seed).NextLong()) != next {
			t.Fatal("unexpected original random consumption", row.State, row.Scene, err)
		}
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 312 {
		t.Fatal("incomplete bubble tick reference", count)
	}
	t.Logf("matched %d original scheduled-tick/onPlace callbacks, including write order and flags", count)
}

func TestGenerationPostprocessingPredicatesAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/generation_postprocess_1.21.1.jsonl.gz")
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
	for scan.Scan() {
		var row struct {
			Neighbor int
			States   []uint32
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		var got []uint32
		for id := uint32(0); id < 26684; id++ {
			s, err := blockstate.BlockStateByID(id)
			if err != nil {
				t.Fatal(err)
			}
			if generationNeedsPostProcessing(SurfaceBlockState{Name: s.Name, Properties: s.Properties}) {
				got = append(got, id)
			}
		}
		if !slices.Equal(got, row.States) {
			t.Fatalf("surrounding state %d: postprocess states %v want %v", row.Neighbor, got, row.States)
		}
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 7 {
		t.Fatal("incomplete generation predicate fixture", count)
	}
	t.Log("matched 26684 states in seven original block-access contexts")
}

func TestGenerationAutomaticPostprocessingRetainsOverwrittenPositions(t *testing.T) {
	base := &NoiseChunkResult{X: 1, Z: -1, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
	r, err := NewFeatureRegion(base)
	if err != nil {
		t.Fatal(err)
	}
	r.borrowed = map[featureChunkPosition]bool{{1, -1}: true}
	p := FeaturePosition{16, 40, -1}
	record := &structureBlockResult{blocks: make(map[FeaturePosition]SurfaceBlockState)}
	w := &structureRecordingWorld{region: r, result: record}
	if err := w.SetBlock(p, structureState("magma_block")); err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(p, structureState("stone")); err != nil {
		t.Fatal(err)
	}
	chunk := r.cloneChunk(1, -1)
	index, _ := chunk.Index(0, 40, 15)
	if !chunk.PostProcessing[index] || !record.postProcessing[p] || len(base.PostProcessing) != 0 || len(chunk.BlockTicks) != 0 {
		t.Fatal("automatic marker lost, aliased or converted to a generation tick")
	}
	chunk.PostProcessing[index] = false
	if !r.cloneChunk(1, -1).PostProcessing[index] {
		t.Fatal("detached marker aliases region")
	}
}
