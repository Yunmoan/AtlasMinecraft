package server

import (
	"atlas/internal/item"
	"atlas/internal/proto"
	"atlas/internal/world"
)

func generatedItemFrameMetadata(id int32, entity world.WorldEntity) []byte {
	w := proto.NewWriter().VarInt(id)
	if stack, err := item.FromNBT(entity.Data["Item"]); err == nil && !stack.Empty() {
		w.U8(8).VarInt(7)
		stack.Write(w)
	}
	if rotation, ok := entity.Data.Number("ItemRotation"); ok && rotation != 0 {
		w.U8(9).VarInt(1).VarInt(int32(rotation) & 7)
	}
	return w.U8(0xff).Payload()
}
