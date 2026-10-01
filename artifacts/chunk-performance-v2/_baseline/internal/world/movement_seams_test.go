package world

import "testing"

func TestMovementSlidesAroundWallCornerWithoutCuttingThroughIt(t *testing.T) {
	w, _ := mechanicsWorld(t)
	for y := int32(-60); y <= -59; y++ {
		if err := w.SetBlock(1, y, 1, StateStone); err != nil {
			t.Fatal(err)
		}
	}
	allowed, err := w.CanMovePlayer([3]float64{.5, -60, 1.7}, [3]float64{1.5, -60, 2.9}, 1.8, true)
	if err != nil || !allowed {
		t.Fatalf("vanilla axis-resolved corner slide rejected: %v", err)
	}
	allowed, err = w.CanMovePlayer([3]float64{.5, -60, 1.5}, [3]float64{2.5, -60, 1.5}, 1.8, false)
	if err != nil || allowed {
		t.Fatalf("wall crossing accepted: %v", err)
	}
}
