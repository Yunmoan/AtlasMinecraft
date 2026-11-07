package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"testing"
)

type brushableTickProbeWorld struct {
	*fallingBlockProbeWorld
	progress *BrushableProgress
	now      int64
}

func (w *brushableTickProbeWorld) ResetBrushableBlockProgress(p FeaturePosition, state SurfaceBlockState) error {
	if w.progress == nil {
		return nil
	}
	after, err := ResetBrushableProgress(w, p, state, *w.progress, w.now)
	if err == nil {
		*w.progress = after
	}
	return err
}

func TestBrushableScheduledTicksAgainstVanilla(t *testing.T) {
	cases, spawns := 0, 0
	scanPostprocessingReference(t, "../../testdata/worldgen/brushable_callbacks_1.21.1.jsonl.gz", func(raw []byte) {
		var row struct {
			State, Below     uint32
			Y, Count         int32
			Now              int64
			Next             string
			Events, Entities []string
			Progress         []int64
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
		p := FeaturePosition{-17, row.Y, 47}
		s := state(row.State)
		w := &brushableTickProbeWorld{fallingBlockProbeWorld: &fallingBlockProbeWorld{blockShapeProbeWorld: blockShapeProbeWorld{fluidStepProbeWorld: &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: s, treeOffset(p, 0, -1, 0): state(row.Below)}, floor: -1000}}}, now: row.Now}
		if row.Count >= 0 {
			w.progress = &BrushableProgress{Count: row.Count, ResetAt: 100, CooldownEnd: 150, HitDirection: 2}
		}
		if err := TickScheduledBlock(w, p, s.Name); err != nil {
			t.Fatal(err)
		}
		var progress []int64
		if w.progress != nil {
			progress = []int64{int64(w.progress.Count), w.progress.ResetAt, w.progress.CooldownEnd, int64(w.progress.HitDirection)}
		}
		if !slices.Equal(w.events, row.Events) || !slices.Equal(progress, row.Progress) || len(w.entities) != len(row.Entities) {
			t.Fatalf("state=%d below=%d y=%d count=%d now=%d: events=%v want=%v progress=%v want=%v entities=%d/%d", row.State, row.Below, row.Y, row.Count, row.Now, w.events, row.Events, progress, row.Progress, len(w.entities), len(row.Entities))
		}
		for i, encoded := range row.Entities {
			data, err := hex.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(data)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			if !reflect.DeepEqual(w.entities[i], want) {
				t.Fatalf("case %d spawned NBT: got=%v want=%v", cases, w.entities[i], want)
			}
			spawns++
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil || NewLegacyRandom(1234).NextLong() != next {
			t.Fatal("brushable tick consumed randomness", cases)
		}
		cases++
	})
	if cases != 9240 || spawns != 4224 {
		t.Fatal("incomplete brushable tick coverage", cases, spawns)
	}
	t.Logf("matched %d native brushing resets/ticks and %d complete falling entity snapshots", cases, spawns)
}
