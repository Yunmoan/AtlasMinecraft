package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func prepareFallingItemLanding(t *testing.T, w *World, atLanding bool) {
	t.Helper()
	slab, err := BlockStateID("stone_slab", map[string]string{"type": "bottom", "waterlogged": "false"})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.SetBlock(-17, 60, 47, slab); err != nil {
		t.Fatal(err)
	}
	if atLanding {
		if err := w.fallingEntityOperation(-2, 2, func(actor *Region, ch *Chunk) error {
			data, err := cloneEntityData(ch.entities[0].Data)
			if err != nil {
				return err
			}
			pos, _ := data["Pos"].Elements()
			pos[1].Value = float64(60.5)
			e, err := worldEntityFromNBT(data)
			if err != nil {
				return err
			}
			return persistFallingEntities(actor, ch, []WorldEntity{e})
		}); err != nil {
			t.Fatal(err)
		}
	}
	w.tickRandom = worldgen.NewLegacyRandom(1234)
}

func TestFallingItemConversionAndGameRulePersist(t *testing.T) {
	for _, rule := range []string{"", "true", "false"} {
		t.Run("doEntityDrops="+rule, func(t *testing.T) {
			w, dir, falling := spawnTestFallingSand(t)
			prepareFallingItemLanding(t, w, false)
			v := w.store.vanilla
			v.mu.Lock()
			root, _ := v.root.Object()
			data, _ := root.Object("Data")
			rules, _ := data.Object("GameRules")
			if rules == nil {
				rules = nbt.Object{}
				data["GameRules"] = nbt.Tag{Type: nbt.Compound, Value: rules}
			}
			if rule == "" {
				delete(rules, "doEntityDrops")
			} else {
				rules["doEntityDrops"] = nbt.Tag{Type: nbt.String, Value: rule}
			}
			v.mu.Unlock()
			for i := 0; i < 20; i++ {
				if err := w.TickFallingEntities(func(x, z int32) bool { return x == -2 && z == 2 }); err != nil {
					t.Fatal(err)
				}
			}
			entities, err := w.EntitiesInChunk(-2, 2)
			if err != nil {
				t.Fatal(err)
			}
			expectedCount := 1
			if rule == "false" {
				expectedCount = 0
			}
			if len(entities) != expectedCount {
				t.Fatalf("item count=%d want=%d", len(entities), expectedCount)
			}
			expectedRandom := worldgen.NewLegacyRandom(1234)
			if expectedCount == 1 {
				e := entities[0]
				if e.Type != "minecraft:item" || e.UUID == falling.UUID || e.Position != [3]float64{-16.5, 60.5, 47.5} {
					t.Fatalf("wrong replacement entity: %+v", e)
				}
				stack, _ := e.Data.Object("Item")
				name, _ := stack.Text("id")
				count, _ := stack.Number("count")
				if name != "minecraft:sand" || count != 1 {
					t.Fatalf("wrong item: %v", stack)
				}
				for key, want := range map[string]int64{"Age": 0, "Health": 5, "PickupDelay": 10} {
					got, ok := e.Data.Number(key)
					if !ok || got != want {
						t.Fatalf("%s=%d want=%d", key, got, want)
					}
				}
				expectedRandom.NextDouble()
				expectedRandom.NextDouble()
			}
			if got, want := w.tickRandom.NextLong(), expectedRandom.NextLong(); got != want {
				t.Fatalf("drop consumed wrong level RNG: %d want %d", got, want)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			reopened := fallingTestWorld(t, dir)
			saved, err := reopened.EntitiesInChunk(-2, 2)
			if err != nil {
				t.Fatal(err)
			}
			if len(saved) != len(entities) {
				t.Fatal("item count changed on reload")
			}
			for i := range saved {
				if saved[i].UUID != entities[i].UUID || !reflect.DeepEqual(saved[i].Data, entities[i].Data) {
					t.Fatal("item NBT changed on reload")
				}
			}
		})
	}
}

func TestFallingItemCommitFailureRetainsEntity(t *testing.T) {
	w, dir, falling := spawnTestFallingSand(t)
	prepareFallingItemLanding(t, w, true)
	file := filepath.Join(dir, "entities", "r.-1.0.mca")
	backup := file + ".test-backup"
	if err := os.Rename(file, backup); err != nil {
		t.Fatal(err)
	}
	restored := false
	restore := func() {
		if restored {
			return
		}
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Error(err)
			return
		}
		if err := os.Rename(backup, file); err != nil {
			t.Error(err)
			return
		}
		restored = true
	}
	t.Cleanup(restore)
	if err := os.Mkdir(file, 0755); err != nil {
		t.Fatal(err)
	}
	eligible := func(x, z int32) bool { return x == -2 && z == 2 }
	if err := w.TickFallingEntities(eligible); err == nil {
		t.Fatal("entity save unexpectedly succeeded with obstructed region path")
	}
	entities, err := w.EntitiesInChunk(-2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 1 || entities[0].UUID != falling.UUID || entities[0].Type != "minecraft:falling_block" {
		t.Fatal("failed commit lost falling entity or published item")
	}
	restore()
	if !restored {
		t.Fatal("failed to restore test entity region")
	}
	if err := w.TickFallingEntities(eligible); err != nil {
		t.Fatal(err)
	}
	items, err := w.EntitiesInChunk(-2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Type != "minecraft:item" {
		t.Fatalf("retry entities=%v", items)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := fallingTestWorld(t, dir)
	if err := reopened.TickFallingEntities(eligible); err != nil {
		t.Fatal(err)
	}
	saved, err := reopened.EntitiesInChunk(-2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 1 || saved[0].UUID != items[0].UUID || !reflect.DeepEqual(saved[0].Data, items[0].Data) {
		t.Fatal("retry/reload changed or duplicated item")
	}
}
