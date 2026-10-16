package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
)

//go:embed bastion_actor_defaults_1.21.1.txt
var bastionActorDefaultsText string

var loadBastionActorDefaults = sync.OnceValues(func() (map[string]nbt.Object, error) {
	result := make(map[string]nbt.Object)
	for _, line := range strings.Split(strings.TrimSpace(bastionActorDefaultsText), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid Bastion actor default")
		}
		data, err := hex.DecodeString(fields[1])
		if err != nil {
			return nil, err
		}
		tag, consumed, err := nbt.ParseNetwork(data)
		if err != nil || consumed != len(data) {
			return nil, fmt.Errorf("invalid Bastion actor NBT: %v", err)
		}
		object, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("Bastion actor default must be a compound")
		}
		delete(object, "UUID")
		result[fields[0]] = object
	}
	if len(result) != 3 {
		return nil, fmt.Errorf("expected three Bastion actor defaults")
	}
	return result, nil
})

// FinalizeBastionTemplateEntity applies the actual STRUCTURE mob spawn stage
// for original Bastion actors. The caller supplies WorldGenRegion's stream,
// independently of the piece placement random. This currently covers fresh
// NORMAL worlds (local difficulty special multiplier zero); enchanted equipment
// in older or HARD worlds requires the separate enchantment implementation.
func FinalizeBastionTemplateEntity(request PendingStructureEntity, random *Xoroshiro) (nbt.Object, error) {
	defaults, err := loadBastionActorDefaults()
	if err != nil {
		return nil, err
	}
	base, supported := defaults[request.ID]
	if !supported || random == nil {
		return nil, fmt.Errorf("unsupported Bastion actor %s", request.ID)
	}
	result := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: base}).Value.(nbt.Object)
	// Entity.load/save drops obsolete and unknown template fields, including the
	// pre-resource-key Dimension integer. Species flags are saved conditionally.
	for key, value := range request.Data {
		_, known := base[key]
		known = known || key == "CustomName" || key == "CustomNameVisible" || key == "Silent" || key == "NoGravity" || key == "NoAI" || key == "Tags"
		switch request.ID {
		case "minecraft:piglin":
			known = known || key == "CannotHunt" || key == "IsBaby" || key == "IsImmuneToZombification"
		case "minecraft:piglin_brute":
			known = known || key == "IsImmuneToZombification"
		case "minecraft:hoglin":
			known = known || key == "CannotBeHunted" || key == "IsImmuneToZombification"
		}
		if known && key != "UUID" {
			result[key] = cloneFeatureNBT(value)
		}
	}
	result["id"] = nbt.Tag{Type: nbt.String, Value: request.ID}
	attributes, _ := result["attributes"].Elements()
	// Piglin's constructor marks its movement attribute dirty even when loading
	// the authored empty attributes list. The default save alone has no entry.
	if request.ID == "minecraft:piglin" {
		found := false
		for _, attribute := range attributes {
			object, _ := attribute.Object()
			id, _ := object.Text("id")
			found = found || id == "minecraft:generic.movement_speed"
		}
		if !found {
			attributes = append(attributes, animalAttribute("movement_speed", float64(float32(.35))))
		}
	}
	brain, ok := result.Object("Brain")
	if !ok {
		brain = make(nbt.Object)
		result["Brain"] = nbt.Tag{Type: nbt.Compound, Value: brain}
	}
	memories, ok := brain.Object("memories")
	if !ok {
		memories = make(nbt.Object)
		brain["memories"] = nbt.Tag{Type: nbt.Compound, Value: memories}
	}
	var hoglinAttack *nbt.Tag
	switch request.ID {
	case "minecraft:piglin":
		memories["minecraft:hunted_recently"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
			"value": {Type: nbt.Byte, Value: int8(1)},
			"ttl":   {Type: nbt.Long, Value: int64(600 + random.NextIntBound(1801))},
		}}
		baby, _ := result.Number("IsBaby")
		armor, _ := result["ArmorItems"].Elements()
		if len(armor) != 4 {
			return nil, fmt.Errorf("invalid Bastion piglin armor list")
		}
		if baby == 0 {
			for index := 3; index >= 0; index-- {
				if random.NextFloat() < .1 {
					armor[index] = bastionActorItem([]string{"golden_boots", "golden_leggings", "golden_chestplate", "golden_helmet"}[index])
				}
			}
		}
		result["ArmorItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: armor}
		hand, _ := result["HandItems"].Elements()
		if len(hand) != 2 {
			return nil, fmt.Errorf("invalid Bastion piglin hand list")
		}
		// Mob.populateDefaultEquipmentEnchantments still draws at multiplier0.
		if object, _ := hand[0].Object(); len(object) != 0 {
			random.NextFloat()
		}
		for _, item := range armor {
			if object, _ := item.Object(); len(object) != 0 {
				random.NextFloat()
			}
		}
	case "minecraft:piglin_brute":
		position, ok := result["Pos"].Elements()
		if !ok || len(position) != 3 {
			return nil, fmt.Errorf("invalid Bastion brute position")
		}
		coordinates := make([]int32, 3)
		for axis, tag := range position {
			value, ok := tag.Value.(float64)
			if !ok || math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("invalid Bastion brute coordinate")
			}
			coordinates[axis] = int32(math.Floor(value))
		}
		memories["minecraft:home"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"value": {Type: nbt.Compound, Value: nbt.Object{
			"pos":       {Type: nbt.IntArray, Value: coordinates},
			"dimension": {Type: nbt.String, Value: "minecraft:the_nether"},
		}}}}
		hand, _ := result["HandItems"].Elements()
		if len(hand) != 2 {
			return nil, fmt.Errorf("invalid Bastion brute hand list")
		}
		hand[0] = bastionActorItem("golden_axe")
		result["HandItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: hand}
	case "minecraft:hoglin":
		if random.NextFloat() < .2 {
			result["Age"] = nbt.Tag{Type: nbt.Int, Value: int32(-24000)}
			tag := animalAttribute("attack_damage", .5)
			hoglinAttack = &tag
		}
	}
	var follow nbt.Object
	for _, tag := range attributes {
		object, _ := tag.Object()
		if id, _ := object.Text("id"); id == "minecraft:generic.follow_range" {
			follow = object
			break
		}
	}
	if follow == nil {
		tag := animalAttribute("follow_range", 16)
		follow = tag.Value.(nbt.Object)
		attributes = append(attributes, tag)
	}
	modifiers, _ := follow["modifiers"].Elements()
	hasSpawnBonus := false
	for _, modifier := range modifiers {
		object, _ := modifier.Object()
		id, _ := object.Text("id")
		hasSpawnBonus = hasSpawnBonus || id == "minecraft:random_spawn_bonus"
	}
	if !hasSpawnBonus {
		bonus := 0.0 + .11485000000000001*(random.NextDouble()-random.NextDouble())
		modifiers = append(modifiers, nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
			"id":        {Type: nbt.String, Value: "minecraft:random_spawn_bonus"},
			"amount":    {Type: nbt.Double, Value: bonus},
			"operation": {Type: nbt.String, Value: "add_multiplied_base"},
		}})
		follow["modifiers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: modifiers}
	}
	if hoglinAttack != nil {
		attributes = append(attributes, *hoglinAttack)
	}
	result["attributes"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: attributes}
	left := int8(0)
	if random.NextFloat() < .05 {
		left = 1
	}
	result["LeftHanded"] = nbt.Tag{Type: nbt.Byte, Value: left}
	return result, nil
}

func bastionActorItem(id string) nbt.Tag {
	return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"id":    {Type: nbt.String, Value: "minecraft:" + id},
		"count": {Type: nbt.Int, Value: int32(1)},
	}}
}
