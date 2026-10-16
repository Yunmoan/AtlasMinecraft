package server

import (
	"atlas/internal/proto"
	"atlas/internal/world"
)

// Registry order and serializer IDs are captured by GeneratedEntityProbe's
// actual EntityData packet codec. These fields are the persisted generated
// mobs' appearance and base state; this does not implement their AI or ticks.
var generatedCatVariants = map[string]int32{
	"minecraft:tabby": 0, "minecraft:black": 1, "minecraft:red": 2, "minecraft:siamese": 3,
	"minecraft:british_shorthair": 4, "minecraft:calico": 5, "minecraft:persian": 6,
	"minecraft:ragdoll": 7, "minecraft:white": 8, "minecraft:jellie": 9, "minecraft:all_black": 10,
}
var generatedVillagerTypes = map[string]int32{
	"minecraft:desert": 0, "minecraft:jungle": 1, "minecraft:plains": 2, "minecraft:savanna": 3,
	"minecraft:snow": 4, "minecraft:swamp": 5, "minecraft:taiga": 6,
}
var generatedVillagerProfessions = map[string]int32{
	"minecraft:none": 0, "minecraft:armorer": 1, "minecraft:butcher": 2, "minecraft:cartographer": 3,
	"minecraft:cleric": 4, "minecraft:farmer": 5, "minecraft:fisherman": 6, "minecraft:fletcher": 7,
	"minecraft:leatherworker": 8, "minecraft:librarian": 9, "minecraft:mason": 10, "minecraft:nitwit": 11,
	"minecraft:shepherd": 12, "minecraft:toolsmith": 13, "minecraft:weaponsmith": 14,
}

func generatedMobMetadata(id int32, entity world.WorldEntity) []byte {
	switch entity.Type {
	case "minecraft:witch", "minecraft:cat", "minecraft:villager", "minecraft:zombie_villager",
		"minecraft:sheep", "minecraft:pig", "minecraft:chicken", "minecraft:cow", "minecraft:horse", "minecraft:donkey",
		"minecraft:armadillo", "minecraft:fox", "minecraft:frog", "minecraft:goat", "minecraft:llama", "minecraft:mooshroom", "minecraft:panda", "minecraft:parrot", "minecraft:polar_bear", "minecraft:rabbit", "minecraft:turtle", "minecraft:wolf", "minecraft:piglin", "minecraft:piglin_brute", "minecraft:hoglin", "minecraft:drowned", "minecraft:allay", "minecraft:iron_golem", "minecraft:shulker", "minecraft:camel":
	default:
		return nil
	}
	w := proto.NewWriter().VarInt(id)
	if health, ok := generatedSnapshotHealth(entity); ok && health != 1 {
		w.U8(9).VarInt(3).F32(health)
	}
	var flags byte
	if value, _ := entity.Data.Number("NoAI"); value != 0 {
		flags |= 1
	}
	if value, _ := entity.Data.Number("LeftHanded"); value != 0 {
		flags |= 2
	}
	if flags != 0 {
		w.U8(15).VarInt(0).U8(flags)
	}
	age, _ := entity.Data.Number("Age")
	baby, _ := entity.Data.Number("IsBaby")
	zombie := entity.Type == "minecraft:zombie_villager" || entity.Type == "minecraft:drowned" || entity.Type == "minecraft:piglin"
	if entity.Type != "minecraft:witch" && !zombie && age < 0 || zombie && baby != 0 {
		w.U8(16).VarInt(8).Bool(true)
	}
	if entity.Type == "minecraft:sheep" {
		color, _ := entity.Data.Number("Color")
		value := byte(color) & 15
		if sheared, _ := entity.Data.Number("Sheared"); sheared != 0 {
			value |= 16
		}
		if value != 0 {
			w.U8(17).VarInt(0).U8(value)
		}
	}
	if entity.Type == "minecraft:pig" {
		if saddle, _ := entity.Data.Number("Saddle"); saddle != 0 {
			w.U8(17).VarInt(8).Bool(true)
		}
	}
	if entity.Type == "minecraft:horse" || entity.Type == "minecraft:donkey" || entity.Type == "minecraft:llama" {
		var flags byte
		if tame, _ := entity.Data.Number("Tame"); tame != 0 {
			flags |= 2
		}
		if saddle, ok := entity.Data.Object("SaddleItem"); ok {
			if item, _ := saddle.Text("id"); item != "" && item != "minecraft:air" {
				flags |= 4
			}
		}
		if bred, _ := entity.Data.Number("Bred"); bred != 0 {
			flags |= 8
		}
		if grazing, _ := entity.Data.Number("EatingHaystack"); grazing != 0 {
			flags |= 16
		}
		if flags != 0 {
			w.U8(17).VarInt(0).U8(flags)
		}
		if entity.Type == "minecraft:horse" {
			if variant, _ := entity.Data.Number("Variant"); variant != 0 {
				w.U8(18).VarInt(1).VarInt(int32(variant))
			}
		}
		if entity.Type == "minecraft:donkey" || entity.Type == "minecraft:llama" {
			if chest, _ := entity.Data.Number("ChestedHorse"); chest != 0 {
				w.U8(18).VarInt(8).Bool(true)
			}
		}
	}
	appendGeneratedAnimalSubtypeMetadata(w, entity)
	if entity.Type == "minecraft:camel" {
		appendGeneratedCamelMetadata(w, entity)
	}
	if entity.Type == "minecraft:shulker" {
		if face, ok := entity.Data.Number("AttachFace"); ok && face != 0 {
			w.U8(16).VarInt(12).VarInt(int32(face))
		}
		if peek, ok := entity.Data.Number("Peek"); ok && peek != 0 {
			w.U8(17).VarInt(0).I8(int8(peek))
		}
		if color, ok := entity.Data.Number("Color"); ok && color != 16 {
			w.U8(18).VarInt(0).I8(int8(color))
		}
	}
	if entity.Type == "minecraft:allay" {
		if duplicate, ok := entity.Data.Number("CanDuplicate"); ok && duplicate == 0 {
			w.U8(17).VarInt(8).Bool(false)
		}
	}
	if entity.Type == "minecraft:cat" {
		variant, _ := entity.Data.Text("variant")
		if value, ok := generatedCatVariants[variant]; ok && value != 1 {
			w.U8(19).VarInt(22).VarInt(value)
		}
		if collar, ok := entity.Data.Number("CollarColor"); ok && collar != 14 {
			w.U8(22).VarInt(1).VarInt(int32(collar) & 15)
		}
	}
	if entity.Type == "minecraft:villager" || entity.Type == "minecraft:zombie_villager" {
		if data, ok := entity.Data.Object("VillagerData"); ok {
			typeName, _ := data.Text("type")
			profession, _ := data.Text("profession")
			level, _ := data.Number("level")
			typeID, tOK := generatedVillagerTypes[typeName]
			professionID, pOK := generatedVillagerProfessions[profession]
			if tOK && pOK {
				index := byte(18)
				if entity.Type == "minecraft:zombie_villager" {
					index = 20
				}
				w.U8(index).VarInt(19).VarInt(typeID).VarInt(professionID).VarInt(int32(level))
			}
		}
	}
	return w.U8(0xff).Payload()
}
