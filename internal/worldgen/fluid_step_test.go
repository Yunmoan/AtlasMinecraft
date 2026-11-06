package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"testing"
)

type fluidStepProbeWorld struct {
	blocks      map[FeaturePosition]SurfaceBlockState
	floor       int32
	events      []string
	shapeRandom *LegacyRandom
}

func (w *fluidStepProbeWorld) ShapeRandomInt(bound int32) (int32, error) {
	if w.shapeRandom == nil {
		return 0, fmt.Errorf("probe shape RNG was not configured")
	}
	return w.shapeRandom.NextIntBound(bound), nil
}

type unavailableFluidProbeWorld struct {
	*fluidStepProbeWorld
	available FeaturePosition
	failure   error
}

func (w unavailableFluidProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if p != w.available {
		return SurfaceBlockState{}, w.failure
	}
	return w.fluidStepProbeWorld.BlockAt(p)
}

func TestFluidTickDoesNotIgnoreUnavailableNeighbors(t *testing.T) {
	p := FeaturePosition{0, 64, 0}
	for _, level := range []string{"0", "7"} {
		base := &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: {Name: "minecraft:lava", Properties: map[string]string{"level": level}}}, floor: 63}
		failure := errors.New("neighbor is not available")
		world := unavailableFluidProbeWorld{base, p, failure}
		random := NewLegacyRandom(1234)
		err := TickFluid(world, p, FluidSettings{}, random)
		if !errors.Is(err, failure) || len(base.events) != 0 || random.NextLong() != NewLegacyRandom(1234).NextLong() {
			t.Fatalf("level %s lost read failure or changed world/random: %v %v", level, err, base.events)
		}
	}
}

func (w *fluidStepProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w.blocks[p]; ok {
		return state, nil
	}
	if p.Y <= w.floor {
		return structureState("stone"), nil
	}
	return structureState("air"), nil
}
func (w *fluidStepProbeWorld) SetFluidBlock(p FeaturePosition, state SurfaceBlockState, flags int32) error {
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return err
	}
	old, _ := w.BlockAt(p)
	before, err := blockstate.BlockStateID(old.Name, old.Properties)
	if err != nil {
		return err
	}
	if before == id {
		return nil
	}
	w.blocks[p] = state
	w.events = append(w.events, fmt.Sprintf("set %d %d %d %d %d", p.X, p.Y, p.Z, id, flags))
	return nil
}
func (w *fluidStepProbeWorld) ScheduleFluid(p FeaturePosition, kind FluidType, delay int32) error {
	w.events = append(w.events, fmt.Sprintf("schedule %d %d %d %s %d", p.X, p.Y, p.Z, kind.String(), delay))
	return nil
}
func (w *fluidStepProbeWorld) NotifyFluidNeighbors(p FeaturePosition, name string) error {
	id, err := blockstate.BlockStateID(name, nil)
	if err != nil {
		return err
	}
	w.events = append(w.events, fmt.Sprintf("notify %d %d %d %d", p.X, p.Y, p.Z, id))
	return nil
}
func (w *fluidStepProbeWorld) FluidLevelEvent(p FeaturePosition, event, data int32) error {
	w.events = append(w.events, fmt.Sprintf("event %d %d %d %d %d", p.X, p.Y, p.Z, event, data))
	return nil
}
func (w *fluidStepProbeWorld) BeforeFluidDestroysBlock(p FeaturePosition, _ SurfaceBlockState, kind FluidType) error {
	// The original defined Level is not a ServerLevel: dropResources has no
	// item-entity side effects. Lava still emits the original fizz event.
	if kind.same(FluidLava) {
		return w.FluidLevelEvent(p, 1501, 0)
	}
	return nil
}
func (w *fluidStepProbeWorld) PlaceFluidInContainer(p FeaturePosition, state SurfaceBlockState, fluid FluidState) error {
	// These vectors exercise SimpleWaterloggedBlock containers. Specialized
	// level/block callbacks must be supplied by the eventual runtime adapter.
	waterlogged, ok := state.Properties["waterlogged"]
	if !ok {
		return fmt.Errorf("unsupported probe container %s", state.Name)
	}
	if waterlogged == "true" || fluid.Type != FluidWater {
		return nil
	}
	if err := w.SetFluidBlock(p, stateWithProperty(state, "waterlogged", "true"), 3); err != nil {
		return err
	}
	return w.ScheduleFluid(p, FluidWater, 5)
}

func TestFluidStepsAgainstVanillaJAR(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/fluid_steps_1.21.1.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	scanner := bufio.NewScanner(gz)
	count := 0
	for scanner.Scan() {
		var row struct {
			Warm, Convert bool
			Seed          int64
			Scene, Level  int
			Fluid         string
			Origin        [3]int32
			Floor         int32
			Inputs        [][4]int32
			Events        []string
			Next          string
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		world := &fluidStepProbeWorld{blocks: make(map[FeaturePosition]SurfaceBlockState), floor: row.Floor}
		for _, input := range row.Inputs {
			state, err := blockstate.BlockStateByID(uint32(input[3]))
			if err != nil {
				t.Fatal(err)
			}
			world.blocks[FeaturePosition{input[0], input[1], input[2]}] = SurfaceBlockState{Name: state.Name, Properties: state.Properties}
		}
		random := NewLegacyRandom(row.Seed)
		p := FeaturePosition{row.Origin[0], row.Origin[1], row.Origin[2]}
		err := TickFluid(world, p, FluidSettings{UltraWarm: row.Warm, WaterSourceConversion: row.Convert, LavaSourceConversion: row.Convert}, random)
		next := strconv.FormatUint(random.NextLong(), 10)
		if err != nil || !slices.Equal(world.events, row.Events) || next != row.Next {
			t.Fatalf("case %d warm=%v convert=%v seed=%d %s level=%d scene=%d: %v\ngot %v\nwant %v\nRNG got %s want %s", count, row.Warm, row.Convert, row.Seed, row.Fluid, row.Level, row.Scene, err, world.events, row.Events, next, row.Next)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 10752 {
		t.Fatal("incomplete original fluid-step coverage", count)
	}
}
