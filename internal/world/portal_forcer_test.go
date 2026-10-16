package world

import (
	"atlas/internal/worldgen"
	"context"
	"math"
	"path/filepath"
	"testing"
)

func TestNetherPortalCreateFindRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	w, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	target := BlockPosition{15, 65, 15}
	first, err := w.FindOrCreateNetherPortal(context.Background(), target, "x")
	if err != nil {
		t.Fatal(err)
	}
	if first.Axis != "x" || first.Width != 2 || first.Height != 3 {
		t.Fatal("invalid created portal", first)
	}
	second, err := w.FindOrCreateNetherPortal(context.Background(), target, "z")
	if err != nil {
		t.Fatal(err)
	}
	if second != first {
		t.Fatal("existing portal was not reused", first, second)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w, err = OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	third, err := w.FindOrCreateNetherPortal(context.Background(), target, "z")
	if err != nil {
		t.Fatal(err)
	}
	if third != first {
		t.Fatal("portal POI did not rebuild from saved block states", first, third)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := w.FindOrCreateNetherPortal(ctx, target, "x"); err == nil {
		t.Fatal("canceled portal search continued")
	}
}

func buildRuntimePortal(t *testing.T, w *World, rect worldgen.PortalRectangle) {
	t.Helper()
	obsidian, err := BlockStateID("minecraft:obsidian", nil)
	if err != nil {
		t.Fatal(err)
	}
	portal, err := BlockStateID("minecraft:nether_portal", map[string]string{"axis": rect.Axis})
	if err != nil {
		t.Fatal(err)
	}
	for x := int32(-1); x <= rect.Width; x++ {
		for y := int32(-1); y <= rect.Height; y++ {
			p := rect.Origin
			p.Y += y
			if rect.Axis == "x" {
				p.X += x
			} else {
				p.Z += x
			}
			id := obsidian
			if x >= 0 && x < rect.Width && y >= 0 && y < rect.Height {
				id = portal
			}
			if err := w.SetBlock(p.X, p.Y, p.Z, id); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestNetherPortalArrivalUsesSourceRectangleAndRotation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "world")
	seed := int64(1234)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	source, err := OpenDimension(noViewers{}, 4, dir, "overworld")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	destination, err := OpenDimension(noViewers{}, 4, dir, "nether")
	if err != nil {
		t.Fatal(err)
	}
	defer destination.Close()
	sourceRect := worldgen.PortalRectangle{Origin: worldgen.FeaturePosition{X: 120, Y: 200, Z: 8}, Axis: "x", Width: 3, Height: 3}
	targetRect := worldgen.PortalRectangle{Origin: worldgen.FeaturePosition{X: 15, Y: 200, Z: 1}, Axis: "z", Width: 2, Height: 3}
	buildRuntimePortal(t, source, sourceRect)
	buildRuntimePortal(t, destination, targetRect)
	position := PlayerPosition{X: 121.75, Y: 201.1, Z: 8.7, Yaw: 23, Pitch: 15, Flying: true, Dimension: "minecraft:overworld"}
	velocity := [3]float64{1, .2, 2}
	got, speed, err := destination.NetherPortalArrival(context.Background(), source, BlockPosition{121, 201, 8}, position, velocity, .6, 1.8)
	if err != nil {
		t.Fatal(err)
	}
	relative := worldgen.PortalRelativePosition(sourceRect, [3]float64{position.X, position.Y, position.Z}, .6, 1.8, true)
	exit, yaw, wantSpeed := worldgen.PortalExitPosition(targetRect, "x", relative, .6, 1.8, position.Yaw, velocity)
	for i, value := range []float64{got.X, got.Y, got.Z} {
		if math.Abs(value-exit[i]) > 1e-9 {
			t.Fatal("arrival did not preserve source relative position", got, exit)
		}
	}
	if got.Dimension != "minecraft:the_nether" || got.Yaw != yaw || got.Pitch != position.Pitch || !got.Flying || speed != wantSpeed {
		t.Fatal("portal transition lost rotation/velocity/player state", got, speed)
	}
}
