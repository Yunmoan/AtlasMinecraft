package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
)

// Only the owning actor changes saved metadata. The generator returns a
// detached snapshot under its short cache lock and does no world callbacks.
// Merge only original postprocess flags, retaining references and foreign NBT.
func (v *VanillaSave) refreshStrongholdFlags(ch *Chunk) {
	if v.normal == nil {
		return
	}
	root, ok := ch.anvilRoot.Object()
	if !ok {
		return
	}
	structures, _ := root.Object("structures")
	starts, _ := structures.Object("starts")
	saved, exists := starts.Object("minecraft:stronghold")
	if !exists {
		return
	}
	snapshot, exists := v.normal.StrongholdStartSnapshot(ch.cx, ch.cz)
	if !exists {
		return
	}
	if mergeStrongholdPieceFlags(saved, snapshot) {
		ch.dirty = true
	}
}

func mergeStrongholdPieceFlags(saved nbt.Object, snapshot worldgen.StructureStart) bool {
	cx, xOK := saved.Number("ChunkX")
	cz, zOK := saved.Number("ChunkZ")
	if !xOK || !zOK || cx != int64(snapshot.Chunk.X) || cz != int64(snapshot.Chunk.Z) {
		return false
	}
	children, ok := saved["Children"].Elements()
	if !ok {
		return false
	}
	changed := false
	for i, piece := range snapshot.Pieces {
		if i >= len(children) {
			break
		}
		key := ""
		switch piece.ID {
		case "minecraft:shcc":
			key = "Chest"
		case "minecraft:shpr":
			key = "Mob"
		default:
			continue
		}
		placed, _ := piece.Data.Number(key)
		if placed == 0 {
			continue
		}
		data, ok := children[i].Object()
		if !ok {
			continue
		}
		id, _ := data.Text("id")
		box, ok := data["BB"].Value.([]int32)
		want := []int32{piece.Box.MinX, piece.Box.MinY, piece.Box.MinZ, piece.Box.MaxX, piece.Box.MaxY, piece.Box.MaxZ}
		if id != piece.ID || !ok || len(box) != 6 {
			continue
		}
		match := true
		for j, v := range box {
			match = match && v == want[j]
		}
		if !match {
			continue
		}
		old, _ := data.Number(key)
		if old != 0 {
			continue
		}
		data[key] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
		changed = true
	}
	return changed
}
