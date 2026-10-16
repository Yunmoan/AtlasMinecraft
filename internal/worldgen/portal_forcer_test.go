package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

type portalForcerProbeWorld struct {
	minY, maxY, logical int32
	mode                int
	edits               portalTestWorld
	writes              strings.Builder
	count               int
}

func (w *portalForcerProbeWorld) InWorld(p FeaturePosition) bool {
	return p.Y >= w.minY && p.Y < w.maxY
}
func (w *portalForcerProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w.edits[p]; ok {
		return state, nil
	}
	name := "minecraft:air"
	if w.InWorld(p) {
		switch {
		case w.mode == 2:
			name = "minecraft:stone"
		case w.mode == 3:
			name = "minecraft:lava"
		case w.mode == 4:
		case w.mode == 6:
			if p.Y <= 40 || p.Y >= 43 {
				name = "minecraft:stone"
			}
		case w.mode == 8 && p.Y >= 120:
			name = "minecraft:stone"
		case p.Y <= 40:
			if w.mode == 1 {
				name = "minecraft:water"
			} else if w.mode != 5 || p.Z == 0 {
				name = "minecraft:stone"
			}
		}
	}
	return SurfaceBlockState{Name: name}, nil
}
func (w *portalForcerProbeWorld) SetBlock(p FeaturePosition, s SurfaceBlockState) error {
	id, err := blockstate.BlockStateID(s.Name, s.Properties)
	if err != nil {
		return err
	}
	w.edits[p] = s
	fmt.Fprintf(&w.writes, "%d %d %d %d\n", p.X, p.Y, p.Z, id)
	w.count++
	return nil
}
func (w *portalForcerProbeWorld) HeightmapAt(_ string, x, z int32) (int32, error) {
	for y := w.maxY - 1; y >= w.minY; y-- {
		s, err := w.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return 0, err
		}
		blocking, err := motionBlocking(s)
		if err != nil {
			return 0, err
		}
		if blocking {
			return y + 1, nil
		}
	}
	return w.minY, nil
}

func TestPortalForcerAgainstVanillaJAR(t *testing.T) {
	f, err := os.Open("../../docs/worldgen/portal_forcer_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	bounds := [][3]int32{{0, 256, 128}, {-64, 320, 384}, {0, 80, 80}}
	scan := bufio.NewScanner(f)
	cases := 0
	for scan.Scan() {
		var label, axis, hash string
		var b, mode, count int
		var found bool
		var want FeaturePosition
		if _, err := fmt.Sscan(scan.Text(), &label, &axis, &b, &mode, &found, &want.X, &want.Y, &want.Z, &count, &hash); err != nil || label != "portal_create" {
			t.Fatal(scan.Text(), err)
		}
		world := &portalForcerProbeWorld{minY: bounds[b][0], maxY: bounds[b][1], logical: bounds[b][2], mode: mode, edits: make(portalTestWorld)}
		size := 1000.0
		if mode == 7 {
			size = 2
		}
		border, err := NewPortalBorder(0, 0, size)
		if err != nil {
			t.Fatal(err)
		}
		target := FeaturePosition{0, 65, 0}
		rect, ok, err := CreateNetherPortal(world, target, axis, world.minY, world.maxY, world.logical, border)
		if err != nil || ok != found || found && rect.Origin != want {
			t.Fatalf("%s: got %+v found=%v %v", scan.Text(), rect, ok, err)
		}
		gotHash := fmt.Sprintf("%x", sha256.Sum256([]byte(world.writes.String())))
		if count != world.count || hash != gotHash {
			t.Fatalf("%s: writes count=%d hash=%s", scan.Text(), world.count, gotHash)
		}
		cases++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 54 {
		t.Fatal("incomplete PortalForcer fixture", cases)
	}
}
