package item

import (
	"atlas/internal/nbt"
	"testing"
)

func TestOpaqueSavedItemIdentityIgnoresFieldOrderAndStorageSlot(t *testing.T) {
	stone, _ := ID("stone")
	original := New(stone, 1)
	saved, err := FromNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:stone"}, "count": {Type: nbt.Int, Value: int32(1)}, "Slot": {Type: nbt.Byte, Value: int8(3)}}})
	if err != nil {
		t.Fatal(err)
	}
	if !original.SameSavedFields(saved) {
		t.Fatal("storage Slot made ordinary item opaque state different")
	}
	for i := 0; i < 32; i++ {
		tag, err := saved.ToNBT()
		if err != nil {
			t.Fatal(err)
		}
		data, _ := tag.Object()
		data["Slot"] = nbt.Tag{Type: nbt.Byte, Value: int8(i)}
		again, err := FromNBT(tag)
		if err != nil || !saved.SameSavedFields(again) {
			t.Fatal("compound field order or storage slot changed identity", err)
		}
	}
	tag, err := saved.ToNBT()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := tag.Object()
	data["ForeignItemField"] = nbt.Tag{Type: nbt.String, Value: "retain this field"}
	opaque, err := FromNBT(tag)
	if err != nil {
		t.Fatal(err)
	}
	if saved.SameSavedFields(opaque) {
		t.Fatal("opaque saved item field was ignored")
	}
	again, err := FromNBT(tag)
	if err != nil || !opaque.SameSavedFields(again) {
		t.Fatal("equal opaque field changed identity", err)
	}
}
