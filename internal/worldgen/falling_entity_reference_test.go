package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"
	"testing"
)

type fallingMotionProbeWorld struct{ *fallingBlockProbeWorld }

func (w fallingMotionProbeWorld) MaxBuildHeight() int32 { return 320 }
func (w fallingMotionProbeWorld) FallingGameEvent(p FeaturePosition, event string) error {
	w.events = append(w.events, "game_event "+event)
	return nil
}
func (w fallingMotionProbeWorld) DropFallingBlockItem(p [3]float64, state SurfaceBlockState) error {
	return fmt.Errorf("unexpected item drop in motion fixture")
}
func (w fallingMotionProbeWorld) FallingEntityDropsEnabled() (bool, error) { return true, nil }
func (w fallingMotionProbeWorld) BreakFallingBrushableBlock(p [3]float64, state SurfaceBlockState) error {
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return err
	}
	w.events = append(w.events, fmt.Sprintf("level_event 2001 %d %d %d %d", int32(math.Floor(p[0])), int32(math.Floor(p[1])), int32(math.Floor(p[2])), id), fmt.Sprintf("game_event minecraft:block_destroy %.9f %.9f %.9f", p[0], p[1], p[2]))
	return nil
}

func TestFallingEntityMotionAgainstVanilla(t *testing.T) {
	runFallingMotionReference(t, "../../testdata/worldgen/falling_block_motion_1.21.1.jsonl.gz", 372, 6)
}
func TestBrushableEntityMotionAgainstVanilla(t *testing.T) {
	runFallingMotionReference(t, "../../testdata/worldgen/brushable_motion_1.21.1.jsonl.gz", 992, 16)
}
func runFallingMotionReference(t *testing.T, fixture string, expectedCases, expectedLandings int) {
	var world fallingMotionProbeWorld
	var data nbt.Object
	removed := false
	cases, landings := 0, 0
	scanPostprocessingReference(t, fixture, func(raw []byte) {
		var row struct {
			State, Support   uint32
			Tick             int
			Entities, Events []string
			Removed          []bool
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			t.Fatal(err)
		}
		state := func(id uint32) SurfaceBlockState {
			s, err := blockstate.BlockStateByID(id)
			if err != nil {
				t.Fatal(err)
			}
			return SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		}
		if row.Tick == 0 {
			world = fallingMotionProbeWorld{&fallingBlockProbeWorld{blockShapeProbeWorld: blockShapeProbeWorld{fluidStepProbeWorld: &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{}, floor: -1000}}}}
			for x := int32(-19); x <= -15; x++ {
				for z := int32(45); z <= 49; z++ {
					world.blocks[FeaturePosition{x, 60, z}] = state(row.Support)
				}
			}
			spawn, err := newFallingBlockSpawn(FeaturePosition{-17, 64, 47}, state(row.State))
			if err != nil {
				t.Fatal(err)
			}
			data = spawn.Data
			if brushableBlock(state(row.State).Name) {
				data["CancelDrop"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
			}
			removed = false
		} else if !removed {
			var err error
			data, removed, err = TickFallingEntity(world, data)
			if err != nil {
				t.Fatal(err)
			}
			if removed {
				landings++
			}
		}
		if len(row.Entities) != 1 || len(row.Removed) != 1 {
			t.Fatal("motion fixture must contain one entity")
		}
		encoded, err := hex.DecodeString(row.Entities[0])
		if err != nil {
			t.Fatal(err)
		}
		tag, _, err := nbt.ParseNetwork(encoded)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := tag.Object()
		if !reflect.DeepEqual(data, want) || removed != row.Removed[0] || !slices.Equal(world.events, row.Events) {
			for key, v := range want {
				if !reflect.DeepEqual(data[key], v) {
					t.Errorf("state=%d support=%d tick=%d %s: got=%v want=%v", row.State, row.Support, row.Tick, key, data[key], v)
				}
			}
			t.Fatalf("state=%d support=%d tick=%d removed=%v want=%v events=%v want=%v", row.State, row.Support, row.Tick, removed, row.Removed[0], world.events, row.Events)
		}
		world.events = nil
		cases++
	})
	if cases != expectedCases || landings != expectedLandings {
		t.Fatal("incomplete motion reference", cases, landings)
	}
	t.Logf("matched %d original entity snapshots, including %d landings", cases, landings)
}
