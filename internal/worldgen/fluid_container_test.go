package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"testing"
)

type fluidContainerProbeWorld struct {
	*fluidStepProbeWorld
	random *LegacyRandom
}

func (w fluidContainerProbeWorld) FluidSound(_ FeaturePosition, sound string) error {
	w.events = append(w.events, fmt.Sprintf("sound %s 4 %d", sound, int64(w.random.NextLong())))
	return nil
}
func (w fluidContainerProbeWorld) FluidCampfireDowse(FeaturePosition) error { return nil } // native probe has no block entity
func (w fluidContainerProbeWorld) FluidBlockChangeEvent(FeaturePosition) error {
	w.events = append(w.events, "game_event minecraft:block_change")
	return nil
}

func TestFluidContainersAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/fluid_containers_1.21.1.jsonl.gz")
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
	seen := make(map[uint32]uint8)
	for scan.Scan() {
		var row struct {
			State    uint32
			Kind     FluidType
			Seed     int64
			Accepted bool
			Events   []string
			Next     string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		if row.Kind > FluidLava || seen[row.State]&(1<<row.Kind) != 0 {
			t.Fatal("invalid or duplicate native container case", row.State, row.Kind)
		}
		seen[row.State] |= 1 << row.Kind
		s, err := blockstate.BlockStateByID(row.State)
		if err != nil {
			t.Fatal(err)
		}
		state := SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		p := FeaturePosition{8, 64, 8}
		world := fluidContainerProbeWorld{&fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: state}, floor: 63}, NewLegacyRandom(row.Seed)}
		accepted, err := PlaceFluidInBlock(world, p, state, FluidState{Type: row.Kind})
		if err != nil || accepted != row.Accepted || !slices.Equal(world.events, row.Events) {
			t.Fatalf("state %d kind %s: accepted %v err %v events %v; want %v %v", row.State, row.Kind, accepted, err, world.events, row.Accepted, row.Events)
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if got := uint64(world.random.NextLong()); got != next {
			t.Fatalf("state %d kind %s RNG %d want %d", row.State, row.Kind, got, next)
		}
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 91970 {
		t.Fatalf("incomplete container reference: %d", count)
	}
	for id := uint32(0); id < 26684; id++ {
		rules, err := fluidRulesForID(id)
		if err != nil {
			t.Fatal(err)
		}
		if rules.container() && seen[id] != 31 || !rules.container() && seen[id] != 0 {
			t.Fatal("container coverage differs from independent native registry vector", id, seen[id])
		}
	}
}
