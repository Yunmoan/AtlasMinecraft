package server

import (
	"atlas/internal/proto"
	"atlas/internal/world"
)

var generatedWolfVariants = map[string]int32{"minecraft:ashen": 0, "minecraft:black": 1, "minecraft:chestnut": 2, "minecraft:pale": 3, "minecraft:rusty": 4, "minecraft:snowy": 5, "minecraft:spotted": 6, "minecraft:striped": 7, "minecraft:woods": 8}
var generatedFrogVariants = map[string]int32{"minecraft:temperate": 0, "minecraft:warm": 1, "minecraft:cold": 2}
var generatedPandaGenes = map[string]byte{"normal": 0, "lazy": 1, "worried": 2, "playful": 3, "brown": 4, "weak": 5, "aggressive": 6}

func appendGeneratedAnimalSubtypeMetadata(w *proto.Writer, entity world.WorldEntity) {
	integer := func(index byte, key string, defaultValue int64) {
		if value, ok := entity.Data.Number(key); ok && value != defaultValue {
			w.U8(index).VarInt(1).VarInt(int32(value))
		}
	}
	boolean := func(index byte, key string, defaultValue bool) {
		if value, ok := entity.Data.Number(key); ok && (value != 0) != defaultValue {
			w.U8(index).VarInt(8).Bool(value != 0)
		}
	}
	switch entity.Type {
	case "minecraft:fox":
		if variant, _ := entity.Data.Text("Type"); variant == "snow" {
			w.U8(17).VarInt(1).VarInt(1)
		}
	case "minecraft:frog":
		variant, _ := entity.Data.Text("variant")
		if id, ok := generatedFrogVariants[variant]; ok && id != 0 {
			w.U8(17).VarInt(24).VarInt(id)
		}
	case "minecraft:goat":
		boolean(17, "IsScreamingGoat", false)
		boolean(18, "HasLeftHorn", true)
		boolean(19, "HasRightHorn", true)
	case "minecraft:llama":
		integer(19, "Strength", 0)
		integer(20, "Variant", 0)
	case "minecraft:mooshroom":
		if variant, _ := entity.Data.Text("Type"); variant == "brown" {
			w.U8(17).VarInt(4).String(variant)
		}
	case "minecraft:panda":
		for _, entry := range []struct {
			index byte
			name  string
		}{{20, "MainGene"}, {21, "HiddenGene"}} {
			gene, _ := entity.Data.Text(entry.name)
			if id := generatedPandaGenes[gene]; id != 0 {
				w.U8(entry.index).VarInt(0).U8(id)
			}
		}
	case "minecraft:parrot":
		integer(19, "Variant", 0)
	case "minecraft:polar_bear":
		boolean(17, "Standing", false)
	case "minecraft:rabbit":
		integer(17, "RabbitType", 0)
	case "minecraft:turtle":
		x, _ := entity.Data.Number("HomePosX")
		y, _ := entity.Data.Number("HomePosY")
		z, _ := entity.Data.Number("HomePosZ")
		if x != 0 || y != 0 || z != 0 {
			w.U8(17).VarInt(10).I64(proto.PackPosition(int32(x), int32(y), int32(z)))
		}
		boolean(18, "HasEgg", false)
		x, _ = entity.Data.Number("TravelPosX")
		y, _ = entity.Data.Number("TravelPosY")
		z, _ = entity.Data.Number("TravelPosZ")
		if x != 0 || y != 0 || z != 0 {
			w.U8(21).VarInt(10).I64(proto.PackPosition(int32(x), int32(y), int32(z)))
		}
	case "minecraft:wolf":
		integer(20, "CollarColor", 14)
		variant, _ := entity.Data.Text("variant")
		if id, ok := generatedWolfVariants[variant]; ok && id != 3 {
			w.U8(22).VarInt(23).VarInt(id + 1)
		}
	}
}
