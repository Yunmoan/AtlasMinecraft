package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
)

//go:embed village_actor_defaults_1.21.1.txt
var villageActorDefaultsText string

var loadVillageActorDefaults = sync.OnceValues(func() (map[string]nbt.Object, error) {
	result := make(map[string]nbt.Object)
	for _, line := range strings.Split(strings.TrimSpace(villageActorDefaultsText), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid village actor defaults")
		}
		data, err := hex.DecodeString(fields[1])
		if err != nil {
			return nil, err
		}
		tag, used, err := nbt.ParseNetwork(data)
		if err != nil || used != len(data) {
			return nil, fmt.Errorf("invalid village actor NBT: %v", err)
		}
		object, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("village actor defaults require compounds")
		}
		delete(object, "UUID")
		result[fields[0]] = object
	}
	if len(result) != 11 {
		return nil, fmt.Errorf("expected ten village actor types and the jockey's chicken")
	}
	return result, nil
})

type VillageStructureEntityResult struct {
	Data          nbt.Object
	ChickenJockey bool
	ChickenData   nbt.Object
}

// Canonicalize the unordered AttributeInstance map by registry name. Its JAR
// iteration order depends on JVM identity hashes. Authored arrays also carry
// old unqualified registry keys.
func villageActorAttributes(values []nbt.Tag) ([]nbt.Tag, error) {
	byName := make(map[string]nbt.Tag)
	for _, tag := range values {
		object, _ := tag.Object()
		id, _ := object.Text("id")
		id = densityName(id)
		object["id"] = nbt.Tag{Type: nbt.String, Value: id}
		byName[id] = tag
	}
	var result []nbt.Tag
	for _, id := range []string{"minecraft:generic.armor_toughness", "minecraft:generic.jump_strength", "minecraft:zombie.spawn_reinforcements", "minecraft:generic.attack_knockback", "minecraft:generic.knockback_resistance", "minecraft:generic.armor", "minecraft:generic.movement_speed", "minecraft:generic.follow_range", "minecraft:generic.max_health"} {
		if tag, ok := byName[id]; ok {
			result = append(result, tag)
			delete(byName, id)
		}
	}
	if len(byName) != 0 {
		return nil, fmt.Errorf("unsupported custom village actor attributes")
	}
	sort.Slice(result, func(i, j int) bool {
		a, _ := result[i].Object()
		b, _ := result[j].Object()
		x, _ := a.Text("id")
		y, _ := b.Text("id")
		return x < y
	})
	return result, nil
}

func villageActorAttribute(values *[]nbt.Tag, name string, base float64) nbt.Object {
	for _, tag := range *values {
		object, _ := tag.Object()
		if id, _ := object.Text("id"); id == "minecraft:generic."+name {
			return object
		}
	}
	tag := animalAttribute(name, base)
	*values = append(*values, tag)
	return tag.Value.(nbt.Object)
}

func normalizeVillageTemplateEntity(request PendingStructureEntity) (nbt.Object, []nbt.Tag, error) {
	defaults, err := loadVillageActorDefaults()
	if err != nil {
		return nil, nil, err
	}
	base, ok := defaults[request.ID]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported village template actor %s", request.ID)
	}
	result := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: base}).Value.(nbt.Object)
	for key, value := range request.Data {
		_, known := base[key]
		known = known || key == "CustomName" || key == "CustomNameVisible" || key == "Silent" || key == "NoGravity" || key == "NoAI" || key == "Tags"
		known = known || request.ID == "minecraft:zombie_villager" && (key == "Gossips" || key == "Offers")
		if known && key != "UUID" {
			result[key] = cloneFeatureNBT(value)
		}
	}
	result["id"] = nbt.Tag{Type: nbt.String, Value: request.ID}
	// Entity.load reads missing Fire with getShort (zero), rather than keeping
	// its constructor value. The camel template is only {id:"minecraft:camel"}.
	if _, present := request.Data["Fire"]; !present {
		result["Fire"] = nbt.Tag{Type: nbt.Short, Value: int16(0)}
	}
	attributes, _ := result["attributes"].Elements()
	attributes, err = villageActorAttributes(attributes)
	if err != nil {
		return nil, nil, err
	}
	if request.ID == "minecraft:zombie_villager" {
		// Zombie.load always calls setBaby, materializing movement_speed even
		// when the saved authored array is empty and IsBaby is absent.
		villageActorAttribute(&attributes, "movement_speed", float64(float32(.23)))
	}
	return result, attributes, nil
}

func finalizeVillageMob(data nbt.Object, attributes *[]nbt.Tag, random *Xoroshiro, followRange float64) {
	follow := villageActorAttribute(attributes, "follow_range", followRange)
	modifiers, _ := follow["modifiers"].Elements()
	hasBonus := false
	for _, tag := range modifiers {
		modifier, _ := tag.Object()
		id, _ := modifier.Text("id")
		hasBonus = hasBonus || id == "minecraft:random_spawn_bonus"
	}
	if !hasBonus {
		bonus := .11485000000000001 * (random.NextDouble() - random.NextDouble())
		modifiers = append(modifiers, nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
			"id": {Type: nbt.String, Value: "minecraft:random_spawn_bonus"}, "amount": {Type: nbt.Double, Value: bonus}, "operation": {Type: nbt.String, Value: "add_multiplied_base"},
		}})
		follow["modifiers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: modifiers}
	}
	left := int8(0)
	if random.NextFloat() < .05 {
		left = 1
	}
	data["LeftHanded"] = nbt.Tag{Type: nbt.Byte, Value: left}
}

// FinalizeVillageTemplateEntity follows actual Entity.load, STRUCTURE spawn
// finalization and save for every authored village species. The controlled
// region is a fresh NORMAL world without existing chickens. Mob AI, trading,
// and subsequent entity ticks are separate runtime stages.
func FinalizeVillageTemplateEntity(request PendingStructureEntity, random *Xoroshiro, actor *LegacyRandom, biome string, fullMoon, halloween bool) (VillageStructureEntityResult, error) {
	var result VillageStructureEntityResult
	if random == nil || actor == nil {
		return result, fmt.Errorf("village finalization requires both random streams")
	}
	data, attributes, err := normalizeVillageTemplateEntity(request)
	if err != nil {
		return result, err
	}
	switch request.ID {
	case "minecraft:armor_stand":
		// ArmorStand is a LivingEntity, not a Mob: no spawn random draws.
	case "minecraft:horse":
		data["Variant"] = nbt.Tag{Type: nbt.Int, Value: random.NextIntBound(7) | random.NextIntBound(5)<<8}
		health := float64(float32(15) + float32(random.NextIntBound(8)) + float32(random.NextIntBound(9)))
		speed := (float64(float32(.45)) + random.NextDouble()*.3 + random.NextDouble()*.3 + random.NextDouble()*.3) * .25
		jump := float64(float32(.4)) + random.NextDouble()*.2 + random.NextDouble()*.2 + random.NextDouble()*.2
		for _, entry := range []struct {
			name string
			base float64
		}{{"max_health", health}, {"movement_speed", speed}, {"jump_strength", jump}} {
			villageActorAttribute(&attributes, entry.name, entry.base)["base"] = nbt.Tag{Type: nbt.Double, Value: entry.base}
		}
		finalizeVillageMob(data, &attributes, random, 16)
	case "minecraft:sheep":
		color := int8(0)
		switch draw := random.NextIntBound(100); {
		case draw < 5:
			color = 15
		case draw < 10:
			color = 7
		case draw < 15:
			color = 8
		case draw < 18:
			color = 12
		default:
			if random.NextIntBound(500) == 0 {
				color = 6
			}
		}
		data["Color"] = nbt.Tag{Type: nbt.Byte, Value: color}
		finalizeVillageMob(data, &attributes, random, 16)
	case "minecraft:cat":
		finalizeVillageMob(data, &attributes, random, 16)
		variants := []string{"tabby", "black", "red", "siamese", "british_shorthair", "calico", "persian", "ragdoll", "white", "jellie"}
		if fullMoon {
			variants = append(variants, "all_black")
		}
		data["variant"] = nbt.Tag{Type: nbt.String, Value: "minecraft:" + variants[random.NextIntBound(int32(len(variants)))]}
	case "minecraft:villager":
		data["AssignProfessionWhenSpawned"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
		finalizeVillageMob(data, &attributes, random, 48)
	case "minecraft:camel":
		data["LastPoseTick"] = nbt.Tag{Type: nbt.Long, Value: int64(0)}
		finalizeVillageMob(data, &attributes, random, 16)
	case "minecraft:cow", "minecraft:pig", "minecraft:iron_golem":
		finalizeVillageMob(data, &attributes, random, 16)
	case "minecraft:zombie_villager":
		village, _ := data.Object("VillagerData")
		village["type"] = nbt.Tag{Type: nbt.String, Value: villageBiomeType(biome)}
		finalizeVillageMob(data, &attributes, random, 35)
		random.NextFloat() // loot pickup threshold is zero at fresh NORMAL difficulty
		data["CanPickUpLoot"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		baby := random.NextFloat() < .05
		data["IsBaby"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		if baby {
			data["IsBaby"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
			if float64(random.NextFloat()) >= .05 && float64(random.NextFloat()) < .05 {
				result.ChickenJockey = true
				chicken, chickenErr := finalizeInitialAnimal("minecraft:chicken", random, &initialAnimalGroup{})
				defaults, _ := loadVillageActorDefaults()
				result.ChickenData = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: defaults["minecraft:chicken"]}).Value.(nbt.Object)
				for key, value := range chicken {
					result.ChickenData[key] = value
				}
				err = chickenErr
				if err != nil {
					return result, err
				}
				result.ChickenData["Pos"], result.ChickenData["Rotation"] = data["Pos"], data["Rotation"]
				result.ChickenData["IsChickenJockey"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
			}
		}
		random.NextFloat() // supportsBreakDoorGoal is true, but the chance is zero
		data["CanBreakDoors"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		random.NextFloat() // base Mob armor population threshold is zero
		hands, _ := data["HandItems"].Elements()
		if random.NextFloat() < .01 {
			name := "iron_shovel"
			if random.NextIntBound(3) == 0 {
				name = "iron_sword"
			}
			hands[0] = bastionActorItem(name)
		}
		if stack, _ := hands[0].Object(); len(stack) != 0 {
			random.NextFloat()
		}
		if halloween && random.NextFloat() < .25 {
			name := "carved_pumpkin"
			if random.NextFloat() < .1 {
				name = "jack_o_lantern"
			}
			armor, _ := data["ArmorItems"].Elements()
			armor[3] = bastionActorItem(name)
			chances, _ := data["ArmorDropChances"].Elements()
			chances[3] = nbt.Tag{Type: nbt.Float, Value: float32(0)}
		}
		attributes = append(attributes, drownedAttribute("minecraft:zombie.spawn_reinforcements", actor.NextDouble()*float64(float32(.1))))
		knockback := villageActorAttribute(&attributes, "knockback_resistance", 0)
		knockback["modifiers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:random_spawn_bonus"}, "amount": {Type: nbt.Double, Value: actor.NextDouble() * float64(float32(.05))}, "operation": {Type: nbt.String, Value: "add_value"}}}}}
		actor.NextDouble()
		actor.NextFloat()
		data["HandItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: hands}
	}
	attributes, err = villageActorAttributes(attributes)
	if err != nil {
		return result, err
	}
	data["attributes"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: attributes}
	result.Data = data
	return result, nil
}

func villageBiomeType(biome string) string {
	switch biome {
	case "minecraft:desert", "minecraft:badlands", "minecraft:eroded_badlands", "minecraft:wooded_badlands":
		return "minecraft:desert"
	case "minecraft:savanna", "minecraft:savanna_plateau", "minecraft:windswept_savanna":
		return "minecraft:savanna"
	case "minecraft:snowy_plains", "minecraft:ice_spikes", "minecraft:snowy_taiga", "minecraft:frozen_river", "minecraft:snowy_beach", "minecraft:frozen_ocean", "minecraft:deep_frozen_ocean", "minecraft:grove", "minecraft:snowy_slopes", "minecraft:frozen_peaks", "minecraft:jagged_peaks":
		return "minecraft:snow"
	case "minecraft:taiga", "minecraft:old_growth_pine_taiga", "minecraft:old_growth_spruce_taiga", "minecraft:windswept_forest", "minecraft:windswept_hills", "minecraft:windswept_gravelly_hills":
		return "minecraft:taiga"
	case "minecraft:swamp", "minecraft:mangrove_swamp":
		return "minecraft:swamp"
	case "minecraft:jungle", "minecraft:sparse_jungle", "minecraft:bamboo_jungle":
		return "minecraft:jungle"
	default:
		return "minecraft:plains"
	}
}
