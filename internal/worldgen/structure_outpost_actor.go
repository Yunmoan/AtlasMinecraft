package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

//go:embed outpost_actor_defaults_1.21.1.txt
var outpostActorDefaultsText string

var loadOutpostActorDefaults = sync.OnceValues(func() (map[string]nbt.Object, error) {
	result := make(map[string]nbt.Object)
	for _, line := range strings.Split(strings.TrimSpace(outpostActorDefaultsText), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid Outpost actor defaults")
		}
		data, err := hex.DecodeString(fields[1])
		if err != nil {
			return nil, err
		}
		tag, used, err := nbt.ParseNetwork(data)
		if err != nil || used != len(data) {
			return nil, fmt.Errorf("invalid Outpost actor NBT: %v", err)
		}
		object, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("Outpost actor defaults require compounds")
		}
		delete(object, "UUID")
		result[fields[0]] = object
	}
	if len(result) != 2 {
		return nil, fmt.Errorf("expected two original Outpost actor types")
	}
	return result, nil
})

// FinalizeOutpostTemplateEntity handles the original authored cage golem and
// two allays. Pillagers are not authored template actors. The region random is
// independent of the piece placement stream. Neither species has a specialized
// STRUCTURE finalize stage; Mob preserves existing spawn-bonus modifiers and
// then draws handedness even for species that do not wield equipment.
func FinalizeOutpostTemplateEntity(request PendingStructureEntity, random *Xoroshiro) (nbt.Object, error) {
	defaults, err := loadOutpostActorDefaults()
	if err != nil {
		return nil, err
	}
	base, ok := defaults[request.ID]
	if !ok || random == nil {
		return nil, fmt.Errorf("unsupported Outpost template actor %s", request.ID)
	}
	result := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: base}).Value.(nbt.Object)
	for key, value := range request.Data {
		_, known := base[key]
		known = known || key == "CustomName" || key == "CustomNameVisible" || key == "Silent" || key == "NoGravity" || key == "NoAI" || key == "Tags"
		if known && key != "UUID" {
			result[key] = cloneFeatureNBT(value)
		}
	}
	result["id"] = nbt.Tag{Type: nbt.String, Value: request.ID}
	if request.ID == "minecraft:allay" {
		// Allay.load reads missing CanDuplicate as false, even though its
		// constructor initializes the tracked flag to true. These authored
		// templates predate the saved duplication fields.
		if _, present := request.Data["CanDuplicate"]; !present {
			result["CanDuplicate"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		}
	}
	attributes, _ := result["attributes"].Elements()
	byName := make(map[string]nbt.Tag)
	for _, tag := range attributes {
		object, _ := tag.Object()
		id, _ := object.Text("id")
		id = densityName(id)
		object["id"] = nbt.Tag{Type: nbt.String, Value: id}
		byName[id] = tag
	}
	// Entity.load resolves old unqualified attribute names through the original
	// registry. Original saving iterates an unordered attribute-instance map;
	// keep a deterministic order without changing its values or modifiers.
	order := []string{"movement_speed", "follow_range"}
	if request.ID == "minecraft:iron_golem" {
		order = append(order, "attack_knockback", "knockback_resistance", "armor_toughness", "armor", "max_health")
	}
	attributes = nil
	for _, name := range order {
		if tag, exists := byName["minecraft:generic."+name]; exists {
			attributes = append(attributes, tag)
			delete(byName, "minecraft:generic."+name)
		}
	}
	if len(byName) != 0 {
		return nil, fmt.Errorf("unsupported custom Outpost actor attributes")
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
		value := float64(16)
		if request.ID == "minecraft:allay" {
			value = 48
		}
		tag := animalAttribute("follow_range", value)
		follow = tag.Value.(nbt.Object)
		attributes = append(attributes, tag)
	}
	modifiers, _ := follow["modifiers"].Elements()
	hasSpawnBonus := false
	for _, tag := range modifiers {
		modifier, _ := tag.Object()
		id, _ := modifier.Text("id")
		hasSpawnBonus = hasSpawnBonus || id == "minecraft:random_spawn_bonus"
	}
	if !hasSpawnBonus {
		bonus := 0.0 + .11485000000000001*(random.NextDouble()-random.NextDouble())
		modifiers = append(modifiers, nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
			"id": {Type: nbt.String, Value: "minecraft:random_spawn_bonus"}, "amount": {Type: nbt.Double, Value: bonus}, "operation": {Type: nbt.String, Value: "add_multiplied_base"},
		}})
		follow["modifiers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: modifiers}
	}
	result["attributes"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: attributes}
	left := int8(0)
	if random.NextFloat() < .05 {
		left = 1
	}
	result["LeftHanded"] = nbt.Tag{Type: nbt.Byte, Value: left}
	return result, nil
}
