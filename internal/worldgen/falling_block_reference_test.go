package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"testing"
)

type fallingBlockProbeWorld struct {
	blockShapeProbeWorld
	entities []nbt.Object
}

func (w *fallingBlockProbeWorld) MinBuildHeight() int32 { return -64 }
func (w *fallingBlockProbeWorld) SpawnFallingBlock(spawn FallingBlockSpawn) error {
	if err := w.SetFluidBlock(spawn.Source, spawn.Replacement, 3); err != nil {
		return err
	}
	w.events = append(w.events, fmt.Sprintf("spawn minecraft:falling_block %.1f %.1f %.1f", float64(spawn.Source.X)+0.5, float64(spawn.Source.Y), float64(spawn.Source.Z)+0.5))
	w.entities = append(w.entities, spawn.Data)
	return nil
}

func TestFallingBlockCallbacksAgainstVanilla(t *testing.T) {
	cases, spawned := 0, 0
	scanPostprocessingReference(t, "../../testdata/worldgen/falling_block_callbacks_1.21.1.jsonl.gz", func(data []byte) {
		var row struct {
			State, Below     uint32
			Y                int32
			Action, Next     string
			Events, Entities []string
		}
		if err := json.Unmarshal(data, &row); err != nil {
			t.Fatal(err)
		}
		state := func(id uint32) SurfaceBlockState {
			s, err := blockstate.BlockStateByID(id)
			if err != nil {
				t.Fatal(err)
			}
			return SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		}
		p := FeaturePosition{-17, row.Y, 47}
		s := state(row.State)
		w := &fallingBlockProbeWorld{blockShapeProbeWorld: blockShapeProbeWorld{fluidStepProbeWorld: &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: s, treeOffset(p, 0, -1, 0): state(row.Below)}, floor: -1000}}}
		var err error
		switch row.Action {
		case "tick":
			err = TickScheduledBlock(w, p, s.Name)
		case "place":
			err = NotifyFallingBlockPlacement(w, p, s)
		case "shape":
			var after SurfaceBlockState
			after, err = UpdateFluidNeighborShape(w, p, s, 0, state(row.Below))
			if err == nil {
				id, e := blockstate.BlockStateID(after.Name, after.Properties)
				err = e
				w.events = append(w.events, fmt.Sprintf("result %d", id))
			}
		default:
			t.Fatal("unknown action", row.Action)
		}
		if err != nil || !slices.Equal(w.events, row.Events) {
			t.Fatalf("state=%d below=%d y=%d action=%s: events=%v want=%v err=%v", row.State, row.Below, row.Y, row.Action, w.events, row.Events, err)
		}
		if len(w.entities) != len(row.Entities) {
			t.Fatal("entity count", row.State, row.Y, row.Action)
		}
		for i, value := range row.Entities {
			bytes, err := hex.DecodeString(value)
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(bytes)
			if err != nil {
				t.Fatal(err)
			}
			want, ok := tag.Object()
			if !ok || !reflect.DeepEqual(w.entities[i], want) {
				t.Fatalf("state=%d y=%d entity NBT: got=%v want=%v", row.State, row.Y, w.entities[i], want)
			}
			spawned++
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil || NewLegacyRandom(1234).NextLong() != next {
			t.Fatal("original RNG consumption changed", row.State, row.Y, row.Action)
		}
		cases++
	})
	if cases != 3360 || spawned != 512 {
		t.Fatal("incomplete falling reference", cases, spawned)
	}
	t.Logf("matched %d original callbacks and %d full entity NBT snapshots", cases, spawned)
}
