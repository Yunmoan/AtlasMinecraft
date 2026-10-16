package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"reflect"
	"testing"
)

func TestStrongholdMetadataMergesOnlyPlacedFlags(t *testing.T) {
	start := worldgen.StructureStart{Structure: "minecraft:stronghold", Chunk: worldgen.StructureChunkPosition{X: -100, Z: -125}, References: 0, Pieces: []worldgen.StructurePiece{
		{ID: "minecraft:shcc", Box: worldgen.StructureBoundingBox{MinX: -10, MinY: -20, MinZ: -30, MaxX: -6, MaxY: -16, MaxZ: -26}, Data: nbt.Object{"Chest": {Type: nbt.Byte, Value: int8(1)}}},
		{ID: "minecraft:shpr", Box: worldgen.StructureBoundingBox{MinX: -40, MinY: -20, MinZ: -30, MaxX: -30, MaxY: -13, MaxZ: -15}, Data: nbt.Object{"Mob": {Type: nbt.Byte, Value: int8(1)}}},
	}}
	saved, err := start.NBT()
	if err != nil {
		t.Fatal(err)
	}
	saved["references"] = nbt.Tag{Type: nbt.Int, Value: int32(7)}
	saved["ForeignStart"] = nbt.Tag{Type: nbt.String, Value: "keep"}
	children, _ := saved["Children"].Elements()
	for i, tag := range children {
		data, _ := tag.Object()
		key := []string{"Chest", "Mob"}[i]
		data[key] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		data["ForeignPiece"] = nbt.Tag{Type: nbt.String, Value: "keep"}
	}
	if !mergeStrongholdPieceFlags(saved, start) {
		t.Fatal("placed clips did not update owner flags")
	}
	if mergeStrongholdPieceFlags(saved, start) {
		t.Fatal("idempotent refresh marked another change")
	}
	if v, _ := saved.Number("references"); v != 7 {
		t.Fatal("locate references overwritten")
	}
	for i, tag := range children {
		data, _ := tag.Object()
		if v, _ := data.Number([]string{"Chest", "Mob"}[i]); v != 1 {
			t.Fatal("flag not merged")
		}
		if v, _ := data.Text("ForeignPiece"); v != "keep" {
			t.Fatal("foreign piece data lost")
		}
	}
	if v, _ := saved.Text("ForeignStart"); v != "keep" {
		t.Fatal("foreign start data lost")
	}
	before, err := cloneEntityData(saved)
	if err != nil {
		t.Fatal(err)
	}
	mismatched := start.Clone()
	mismatched.Pieces[0].Box.MinX++
	mismatched.Pieces[1].ID = "minecraft:unknown"
	if mergeStrongholdPieceFlags(saved, mismatched) || !reflect.DeepEqual(saved, before) {
		t.Fatal("foreign/mismatched pieces changed")
	}
	mismatched = start.Clone()
	mismatched.Chunk.X++
	if mergeStrongholdPieceFlags(saved, mismatched) {
		t.Fatal("foreign owner accepted")
	}
}
