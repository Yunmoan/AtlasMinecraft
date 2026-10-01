package world

import (
	"testing"
)

func unpackHeight(data []int64, column int) int32 {
	return MinY + int32((uint64(data[column/7])>>(uint(column%7)*9))&511)
}

func TestGeneratedHeightmapsUseVanillaBlockClassifications(t *testing.T) {
	water, err := BlockStateID("minecraft:water", nil)
	if err != nil {
		t.Fatal(err)
	}
	grass, err := BlockStateID("minecraft:short_grass", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !stateFlag(motionFlags, water) || stateFlag(motionFlags, grass) || stateFlag(oceanFloorFlags, water) {
		t.Fatal("unexpected 1.21.1 water/grass heightmap classification")
	}
	chunk := NewChunk(0, 0)
	chunk.SetBlock(0, 0, 0, StateStone)
	chunk.SetBlock(0, 0, 1, water)
	chunk.SetBlock(0, 0, 2, grass)
	motion, surface, ocean := chunk.anvilHeightmaps()
	if got := unpackHeight(motion, 0); got != 2 {
		t.Fatalf("motion height=%d, want 2", got)
	}
	if got := unpackHeight(surface, 0); got != 3 {
		t.Fatalf("surface height=%d, want 3", got)
	}
	if got := unpackHeight(ocean, 0); got != 1 {
		t.Fatalf("ocean floor height=%d, want 1", got)
	}
	if got := unpackHeight(surface, 1); got != MinY {
		t.Fatalf("empty column height=%d", got)
	}
	if err := chunk.initializeAnvilRoot(); err != nil {
		t.Fatal(err)
	}
	chunk.SetBlock(0, 0, 2, StateAir)
	_, root, err := chunk.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	object, _ := root.Object()
	heightmaps, _ := object.Object("Heightmaps")
	storedSurface, ok := heightmaps["WORLD_SURFACE"].Longs()
	if !ok || len(storedSurface) != 37 {
		t.Fatal("edited Anvil surface heightmap missing")
	}
	if unpackHeight(storedSurface, 0) != 2 {
		t.Fatalf("edited Anvil surface height=%d", unpackHeight(storedSurface, 0))
	}
}
