package world

import (
	"testing"
)

func TestDebugBlockEntityRemoval(t *testing.T) {
	chunk := testGeneratedFeatureChunk(t, "room")
	root, _ := chunk.anvilRoot.Object()
	entities, _ := root["block_entities"].Elements()
	t.Logf("entities: %d", len(entities))
	for _, tag := range entities {
		e, _ := tag.Object()
		x, _ := e.Number("x")
		y, _ := e.Number("y")
		z, _ := e.Number("z")
		id, _ := e.Text("id")
		t.Logf("entity %s at %d %d %d", id, x, y, z)
	}
	x, _, _ := 0, 0, 0
	_ = x
	e0, _ := entities[0].Object()
	ex, _ := e0.Number("x")
	ey, _ := e0.Number("y")
	ez, _ := e0.Number("z")
	t.Logf("block at %d %d %d = %d", ex&15, ez&15, ey, chunk.GetBlock(int(ex&15), int(ez&15), int32(ey)))
	chunk.SetBlock(int(ex&15), int(ez&15), int32(ey), StateStone)
	root2, _ := chunk.anvilRoot.Object()
	left, _ := root2["block_entities"].Elements()
	t.Logf("after replace: %d entities remain", len(left))
}
