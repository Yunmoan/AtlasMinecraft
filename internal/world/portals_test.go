package world

import (
	"atlas/internal/worldgen"
	"path/filepath"
	"testing"
)

func TestNetherPortalIgnitionRestart(t *testing.T) {
	for _, axis := range []string{"x", "z"} {
		t.Run(axis, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "world")
			seed := int64(1234)
			if err := CreateFlatWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			w, err := OpenDimension(noViewers{}, 2, dir, "overworld")
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			obsidian, err := BlockStateID("minecraft:obsidian", nil)
			if err != nil {
				t.Fatal(err)
			}
			base := BlockPosition{15, 65, 15}
			position := func(x, y int32) BlockPosition {
				p := base
				p.Y += y
				if axis == "x" {
					p.X += x
				} else {
					p.Z += x
				}
				return p
			}
			for x := int32(-1); x <= 2; x++ {
				for y := int32(-1); y <= 3; y++ {
					if x >= 0 && x < 2 && y >= 0 && y < 3 {
						continue
					}
					p := position(x, y)
					if err := w.SetBlock(p.X, p.Y, p.Z, obsidian); err != nil {
						t.Fatal(err)
					}
				}
			}
			broken := position(-1, 1)
			if err := w.SetBlock(broken.X, broken.Y, broken.Z, StateAir); err != nil {
				t.Fatal(err)
			}
			if ignited, err := w.IgniteNetherPortal(base.X, base.Y, base.Z); err != nil || ignited {
				t.Fatal("broken frame activated", ignited, err)
			}
			if err := w.SetBlock(broken.X, broken.Y, broken.Z, obsidian); err != nil {
				t.Fatal(err)
			}
			if ignited, err := w.IgniteNetherPortal(base.X, base.Y, base.Z); err != nil || !ignited {
				t.Fatal("valid frame not activated", ignited, err)
			}
			if ignited, err := w.IgniteNetherPortal(base.X, base.Y, base.Z); err != nil || ignited {
				t.Fatal("filled frame treated as empty", ignited, err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			w, err = OpenDimension(noViewers{}, 2, dir, "overworld")
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			for x := int32(0); x < 2; x++ {
				for y := int32(0); y < 3; y++ {
					p := position(x, y)
					id, err := w.BlockAt(p.X, p.Y, p.Z)
					if err != nil {
						t.Fatal(err)
					}
					s, err := BlockStateByID(id)
					if err != nil || s.Name != "minecraft:nether_portal" || s.Properties["axis"] != axis {
						t.Fatal("portal did not persist", s, err)
					}
				}
			}
		})
	}
}

func TestEndArrivalPlatformRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := OpenDimension(noViewers{}, 2, dir, "end")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if _, err := w.IgniteNetherPortal(100, 49, 0); err != nil {
		t.Fatal(err)
	}
	for _, p := range []BlockPosition{{100, 49, 0}, {100, 50, 0}, {100, 51, 0}, {103, 49, 0}} {
		if err := w.SetBlock(p.X, p.Y, p.Z, StateStone); err != nil {
			t.Fatal(err)
		}
	}
	arrival, err := w.PrepareEndArrival()
	if err != nil {
		t.Fatal(err)
	}
	if arrival != (PlayerPosition{X: 100.5, Y: 49, Z: 0.5, Yaw: 90, Dimension: "minecraft:the_end"}) {
		t.Fatal("incorrect End player arrival", arrival)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w, err = OpenDimension(noViewers{}, 2, dir, "end")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	obsidian, err := BlockStateID("minecraft:obsidian", nil)
	if err != nil {
		t.Fatal(err)
	}
	spawn := worldgen.EndSpawnPoint()
	for z := int32(-2); z <= 2; z++ {
		for x := int32(-2); x <= 2; x++ {
			for y := int32(48); y <= 51; y++ {
				want := uint32(StateAir)
				if y == 48 {
					want = obsidian
				}
				got, err := w.BlockAt(spawn.X+x, y, spawn.Z+z)
				if err != nil || got != want {
					t.Fatalf("End platform %d,%d,%d: %d want %d %v", spawn.X+x, y, spawn.Z+z, got, want, err)
				}
			}
		}
	}
	outside, err := w.BlockAt(103, 49, 0)
	if err != nil || outside != StateStone {
		t.Fatal("End platform changed blocks outside its geometry", outside, err)
	}
}
