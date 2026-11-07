package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestLiquidBlockCallbacksAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/liquid_callbacks_1.21.1.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	scan := bufio.NewScanner(z)
	cases := 0
	for scan.Scan() {
		var row struct {
			Warm                     bool
			Fluid                    string
			Level, Scene, Hook, Face int
			Floor                    int32
			Origin                   [3]int32
			Inputs                   [][4]int32
			Events                   []string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		w := &fluidStepProbeWorld{blocks: make(map[FeaturePosition]SurfaceBlockState), floor: row.Floor}
		for _, input := range row.Inputs {
			s, err := blockstate.BlockStateByID(uint32(input[3]))
			if err != nil {
				t.Fatal(err)
			}
			w.blocks[FeaturePosition{input[0], input[1], input[2]}] = SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		}
		p := FeaturePosition{row.Origin[0], row.Origin[1], row.Origin[2]}
		state, _ := w.BlockAt(p)
		settings := FluidSettings{UltraWarm: row.Warm}
		if row.Hook == 2 {
			other, _ := w.BlockAt(fluidNeighbor(p, row.Face))
			err = UpdateLiquidShape(w, p, state, other, settings)
		} else {
			err = NotifyLiquidBlock(w, p, state, settings)
		}
		if err != nil || !slices.Equal(w.events, row.Events) {
			t.Fatalf("case %d warm=%v %s level=%d scene=%d hook=%d: %v\ngot %v\nwant %v", cases, row.Warm, row.Fluid, row.Level, row.Scene, row.Hook, err, w.events, row.Events)
		}
		cases++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 4992 {
		t.Fatalf("got %d native callbacks", cases)
	}
}
