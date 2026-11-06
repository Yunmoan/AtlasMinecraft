package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type kelpTickProbeWorld struct {
	*fluidStepProbeWorld
	destroyed, drops     bool
	random, entityRandom *LegacyRandom
	item                 nbt.Object
}

func (w *kelpTickProbeWorld) DestroyKelpBlock(p FeaturePosition, _ SurfaceBlockState) error {
	w.destroyed = true
	w.events = append(w.events, fmt.Sprintf("destroy %d %d %d true 512", p.X, p.Y, p.Z))
	var err error
	w.item, err = NewKelpBlockDropEntity(p, w.drops, w.random, w.entityRandom)
	return err
}

func TestKelpScheduledTicksAndDropsAgainstVanilla(t *testing.T) {
	cases, items := 0, 0
	scanPostprocessingReference(t, "../../testdata/worldgen/kelp_callbacks_1.21.1.jsonl.gz", func(raw []byte) {
		var row struct {
			State, Below     uint32
			Seed, EntitySeed int64
			Drops, Destroyed bool
			Events, Entities []string
			Next, ItemNext   string
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
		p := FeaturePosition{-17, 64, 47}
		s := state(row.State)
		w := &kelpTickProbeWorld{fluidStepProbeWorld: &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: s, treeOffset(p, 0, -1, 0): state(row.Below)}, floor: -1000}, drops: row.Drops, random: NewLegacyRandom(row.Seed), entityRandom: NewLegacyRandom(row.EntitySeed)}
		if err := TickScheduledBlock(w, p, s.Name); err != nil {
			t.Fatal(err)
		}
		if w.destroyed != row.Destroyed {
			t.Fatalf("state=%d below=%d survival differs", row.State, row.Below)
		}
		wantCount := 0
		if w.item != nil {
			wantCount = 1
		}
		if len(row.Entities) != wantCount || len(row.Events) != len(w.events)+wantCount {
			t.Fatalf("case %d events/items differ: %v/%v", cases, w.events, row.Events)
		}
		for i, event := range w.events {
			if event != row.Events[i] {
				t.Fatalf("destroy callback: %v/%v", w.events, row.Events)
			}
		}
		if w.item != nil {
			bytes, err := hex.DecodeString(row.Entities[0])
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(bytes)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			if !reflect.DeepEqual(w.item, want) {
				t.Fatalf("state=%d below=%d seed=%d item NBT: got=%v want=%v", row.State, row.Below, row.Seed, w.item, want)
			}
			next, err := strconv.ParseUint(row.ItemNext, 10, 64)
			if err != nil || w.entityRandom.NextLong() != next {
				t.Fatal("item RNG differs", cases)
			}
			items++
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil || w.random.NextLong() != next {
			t.Fatal("world RNG differs", cases)
		}
		cases++
	})
	if cases != 3240 || items != 810 {
		t.Fatal("incomplete kelp coverage", cases, items)
	}
	t.Logf("matched %d native scheduled survival/drop cases and %d full item snapshots", cases, items)
}
