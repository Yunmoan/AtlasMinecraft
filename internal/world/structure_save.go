package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
)

func installGeneratedStructures(chunk *Chunk, generated *worldgen.GeneratedOverworldChunk) {
	root, _ := chunk.anvilRoot.Object()
	if generated.Structures != nil {
		root["structures"] = nbt.Tag{Type: nbt.Compound, Value: generated.Structures}
	}
	if len(generated.PendingEntities) > 0 {
		pending := make([]nbt.Tag, 0, len(generated.PendingEntities))
		for _, entity := range generated.PendingEntities {
			persistent := int8(0)
			if entity.Persistent {
				persistent = 1
			}
			data := nbt.Object{
				"id":         {Type: nbt.String, Value: entity.ID},
				"x":          {Type: nbt.Int, Value: entity.Position.X},
				"y":          {Type: nbt.Int, Value: entity.Position.Y},
				"z":          {Type: nbt.Int, Value: entity.Position.Z},
				"Persistent": {Type: nbt.Byte, Value: persistent},
				"PieceFlag":  {Type: nbt.String, Value: entity.SavedFlag},
			}
			if entity.Data != nil {
				data["Data"] = nbt.Tag{Type: nbt.Compound, Value: entity.Data}
			}
			pending = append(pending, nbt.Tag{Type: nbt.Compound, Value: data})
		}
		root["AtlasPendingStructureEntities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: pending}
	}
}
