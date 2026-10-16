package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

//go:embed drowned_actor_defaults_1.21.1.txt
var drownedActorDefaultsText string

var loadDrownedActorDefault = sync.OnceValues(func() (nbt.Object, error) {
	f := strings.Fields(drownedActorDefaultsText)
	if len(f) != 2 {
		return nil, fmt.Errorf("invalid drowned default")
	}
	raw, err := hex.DecodeString(f[1])
	if err != nil {
		return nil, err
	}
	tag, used, err := nbt.ParseNetwork(raw)
	if err != nil || used != len(raw) {
		return nil, fmt.Errorf("invalid drowned default NBT: %v", err)
	}
	o, _ := tag.Object()
	delete(o, "UUID")
	return o, nil
})

type DrownedStructureResult struct {
	Data          nbt.Object
	ChickenJockey bool
	ChickenData   nbt.Object
}

func drownedAttribute(id string, base float64) nbt.Tag {
	return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: id}, "base": {Type: nbt.Double, Value: base}}}
}

func drownedModifier(attribute nbt.Tag, id string, amount float64, operation string) nbt.Tag {
	o, _ := attribute.Object()
	o["modifiers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: id}, "amount": {Type: nbt.Double, Value: amount}, "operation": {Type: nbt.String, Value: operation}}}}}
	return attribute
}

// FinalizeDrownedStructureEntity transcribes fresh NORMAL ocean-ruin spawning.
// Region and entity random sources are independent. The controlled generation
// neighborhood contains no pre-existing chickens. Older/HARD-world enchanted
// equipment requires the separate local-difficulty/enchantment implementation.
func FinalizeDrownedStructureEntity(request PendingStructureEntity, random *Xoroshiro, actor *LegacyRandom, halloween bool) (DrownedStructureResult, error) {
	var result DrownedStructureResult
	if request.ID != "minecraft:drowned" || random == nil || actor == nil {
		return result, fmt.Errorf("drowned structure finalization requires both random streams")
	}
	base, err := loadDrownedActorDefault()
	if err != nil {
		return result, err
	}
	data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: base}).Value.(nbt.Object)
	position := [3]float64{float64(request.Position.X) + .5, float64(request.Position.Y), float64(request.Position.Z) + .5}
	data["Pos"] = animalPositionNBT(position)
	data["Rotation"] = nbt.Tag{Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{{Type: nbt.Float, Value: float32(0)}, {Type: nbt.Float, Value: float32(0)}}}
	data["PersistenceRequired"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	attributes := []nbt.Tag{drownedModifier(animalAttribute("follow_range", 35), "minecraft:random_spawn_bonus", .11485000000000001*(random.NextDouble()-random.NextDouble()), "add_multiplied_base")}
	left := int8(0)
	if random.NextFloat() < .05 {
		left = 1
	}
	data["LeftHanded"] = nbt.Tag{Type: nbt.Byte, Value: left}
	random.NextFloat() // CanPickUpLoot: zero special multiplier, still draws.
	data["CanPickUpLoot"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	baby := random.NextFloat() < .05
	data["IsBaby"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	if baby {
		data["IsBaby"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
		attributes = append(attributes, animalAttribute("movement_speed", float64(float32(.23))))
		if float64(random.NextFloat()) >= .05 && float64(random.NextFloat()) < .05 {
			result.ChickenJockey = true
			result.ChickenData, err = finalizeInitialAnimal("minecraft:chicken", random, &initialAnimalGroup{})
			if err != nil {
				return result, err
			}
			result.ChickenData["Pos"] = animalPositionNBT(position)
			result.ChickenData["Rotation"] = data["Rotation"]
			result.ChickenData["IsChickenJockey"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
		}
	}
	data["CanBreakDoors"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)} // Drowned.supportsBreakDoorGoal is false.
	hands, _ := data["HandItems"].Elements()
	if float64(random.NextFloat()) > .9 {
		name := "fishing_rod"
		if random.NextIntBound(16) < 10 {
			name = "trident"
		}
		hands[0] = bastionActorItem(name)
	}
	if object, _ := hands[0].Object(); len(object) > 0 {
		random.NextFloat()
	} // enchantment threshold is zero.
	if halloween && random.NextFloat() < .25 {
		name := "carved_pumpkin"
		if random.NextFloat() < .1 {
			name = "jack_o_lantern"
		}
		armor, _ := data["ArmorItems"].Elements()
		armor[3] = bastionActorItem(name)
		data["ArmorItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: armor}
		chances, _ := data["ArmorDropChances"].Elements()
		chances[3] = nbt.Tag{Type: nbt.Float, Value: float32(0)}
		data["ArmorDropChances"] = nbt.Tag{Type: nbt.List, ListType: nbt.Float, Value: chances}
	}
	attributes = append(attributes, drownedAttribute("minecraft:zombie.spawn_reinforcements", actor.NextDouble()*float64(float32(.1))))
	attributes = append(attributes, drownedModifier(animalAttribute("knockback_resistance", 0), "minecraft:random_spawn_bonus", actor.NextDouble()*float64(float32(.05)), "add_value"))
	actor.NextDouble() // Additional follow-range multiplier is zero.
	actor.NextFloat()  // Leader threshold is zero.
	if random.NextFloat() < .03 {
		hands[1] = bastionActorItem("nautilus_shell")
		chances, _ := data["HandDropChances"].Elements()
		chances[1] = nbt.Tag{Type: nbt.Float, Value: float32(2)}
		data["HandDropChances"] = nbt.Tag{Type: nbt.List, ListType: nbt.Float, Value: chances}
	}
	data["HandItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: hands}
	data["attributes"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: attributes}
	result.Data = data
	return result, nil
}
