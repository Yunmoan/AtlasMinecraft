package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type fallingDropProbeWorld struct {
	fallingMotionProbeWorld
	drops                     bool
	levelRandom, entityRandom *LegacyRandom
	items                     []nbt.Object
}

func (w *fallingDropProbeWorld) FallingEntityDropsEnabled() (bool, error) { return w.drops, nil }

func (w *fallingDropProbeWorld) DropFallingBlockItem(pos [3]float64, state SurfaceBlockState) error {
	if !w.drops {
		return nil
	}
	data, err := NewFallingBlockDropEntity(pos, state, w.levelRandom, w.entityRandom)
	if err != nil {
		return err
	}
	w.items = append(w.items, data)
	return nil
}

func TestFallingItemDropsAgainstVanilla(t *testing.T) {
	runFallingDropReference(t, "../../testdata/worldgen/falling_block_drops_1.21.1.jsonl.gz", 1200, 135)
}
func TestBrushableItemDropsAgainstVanilla(t *testing.T) {
	runFallingDropReference(t, "../../testdata/worldgen/brushable_drops_1.21.1.jsonl.gz", 800, 90)
}
func runFallingDropReference(t *testing.T, fixture string, expectedCases, expectedItems int) {
	cases, items := 0, 0
	scanPostprocessingReference(t, fixture, func(raw []byte) {
		var row struct {
			State, Support          uint32
			Seed, EntitySeed        int64
			Age                     int32
			Y                       float64
			Drops, DropItem, Cancel bool
			Entities, Events        []string
			Removed                 []bool
			Next, ItemNext          string
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
		w := &fallingDropProbeWorld{
			fallingMotionProbeWorld: fallingMotionProbeWorld{&fallingBlockProbeWorld{blockShapeProbeWorld: blockShapeProbeWorld{fluidStepProbeWorld: &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{}, floor: -1000}}}},
			drops:                   row.Drops, levelRandom: NewLegacyRandom(row.Seed), entityRandom: NewLegacyRandom(row.EntitySeed),
		}
		for x := int32(-19); x <= -15; x++ {
			for z := int32(45); z <= 49; z++ {
				w.blocks[FeaturePosition{x, 60, z}] = state(row.Support)
			}
		}
		spawn, err := newFallingBlockSpawn(FeaturePosition{-17, 64, 47}, state(row.State))
		if err != nil {
			t.Fatal(err)
		}
		data := spawn.Data
		position, _ := data["Pos"].Elements()
		position[1].Value = row.Y
		data["Time"] = nbt.Tag{Type: nbt.Int, Value: row.Age}
		for key, value := range map[string]bool{"DropItem": row.DropItem, "CancelDrop": row.Cancel} {
			v := int8(0)
			if value {
				v = 1
			}
			data[key] = nbt.Tag{Type: nbt.Byte, Value: v}
		}
		data["id"] = nbt.Tag{Type: nbt.String, Value: "minecraft:falling_block"}
		got, removed, err := TickFallingEntity(w, data)
		if err != nil {
			t.Fatal(err)
		}
		actual := append([]nbt.Object{got}, w.items...)
		if len(actual) != len(row.Entities) || removed != row.Removed[0] {
			t.Fatalf("case %d state=%d age=%d y=%g drops=%v DropItem=%v CancelDrop=%v: entity counts %d/%d removal %v/%v", cases, row.State, row.Age, row.Y, row.Drops, row.DropItem, row.Cancel, len(actual), len(row.Entities), removed, row.Removed)
		}
		for i, encoded := range row.Entities {
			raw, err := hex.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			tag, _, err := nbt.ParseNetwork(raw)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := tag.Object()
			if i == 0 {
				delete(want, "UUID")
			}
			if !reflect.DeepEqual(actual[i], want) {
				for key, value := range want {
					if !reflect.DeepEqual(actual[i][key], value) {
						t.Errorf("case %d entity %d field %s got=%v want=%v", cases, i, key, actual[i][key], value)
					}
				}
				t.Fatalf("case %d entity %d NBT differs", cases, i)
			}
		}
		var otherEvents []string
		spawnEvents := 0
		for _, event := range row.Events {
			if strings.HasPrefix(event, "spawn minecraft:item ") {
				spawnEvents++
			} else {
				otherEvents = append(otherEvents, event)
			}
		}
		if spawnEvents != len(w.items) || !slices.Equal(otherEvents, w.events) {
			t.Fatalf("unexpected native callbacks: %v Go %v", row.Events, w.events)
		}
		if row.ItemNext != "" {
			next, err := strconv.ParseUint(row.ItemNext, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			if w.entityRandom.NextLong() != next {
				t.Fatalf("case %d item RNG diverged", cases)
			}
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if w.levelRandom.NextLong() != next {
			t.Fatalf("case %d world RNG diverged", cases)
		}
		cases++
		items += len(w.items)
	})
	if cases != expectedCases || items != expectedItems {
		t.Fatalf("incomplete drop coverage: %d cases, %d items", cases, items)
	}
	t.Logf("matched %d drop decisions, %d full item NBT records including UUID and random-stream continuation", cases, items)
}
