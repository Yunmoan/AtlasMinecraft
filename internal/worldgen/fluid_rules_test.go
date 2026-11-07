package worldgen

import (
	"atlas/internal/blockstate"
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"
)

func TestFluidCollisionPassagesAgainstVanillaJAR(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/fluid_walls_1.21.1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	var vectors struct {
		States   []uint32 `json:"wall_states"`
		Passages []byte   `json:"walls"`
	}
	if err := json.NewDecoder(gz).Decode(&vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.States) != 320 || len(vectors.Passages) != 614400 {
		t.Fatal("incomplete native collision coverage")
	}
	states := make([]SurfaceBlockState, len(vectors.States))
	for i, id := range vectors.States {
		s, e := blockstate.BlockStateByID(id)
		if e != nil {
			t.Fatal(e)
		}
		states[i] = SurfaceBlockState{Name: s.Name, Properties: s.Properties}
	}
	offsets := []FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}}
	index := 0
	for i, a := range states {
		for j, b := range states {
			for face, p := range offsets {
				got, e := fluidCanPassThroughWall(a, FeaturePosition{}, b, p, face)
				if e != nil || got != (vectors.Passages[index] != 0) {
					t.Fatalf("states %d %+v / %d %+v face %d: got %v want %d (%v)", vectors.States[i], a, vectors.States[j], b, face, got, vectors.Passages[index], e)
				}
				index++
			}
		}
	}
}

func TestFluidStateRulesMatchIndependentTypeOracle(t *testing.T) {
	for id := uint32(0); id < 26684; id++ {
		s, err := blockstate.BlockStateByID(id)
		if err != nil {
			t.Fatal(err)
		}
		rules, err := fluidRulesForID(id)
		if err != nil {
			t.Fatal(err)
		}
		if got := rules.state.Type.String(); got != generationFluidType(SurfaceBlockState{Name: s.Name, Properties: s.Properties}) {
			t.Fatalf("fluid type mismatch state %d: %s", id, got)
		}
		if rules.state.Amount < 0 || rules.state.Amount > 8 || rules.state.Source() && rules.state.Amount != 8 {
			t.Fatalf("invalid native fluid state %d: %+v", id, rules.state)
		}
	}
}
