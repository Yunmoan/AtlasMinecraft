package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

//go:embed mansion_actor_defaults_1.21.1.txt
var mansionActorDefaultsText string

var loadMansionActorDefaults = sync.OnceValues(func() (map[string]nbt.Object, error) {
	result := make(map[string]nbt.Object)
	for _, line := range strings.Split(strings.TrimSpace(mansionActorDefaultsText), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid Mansion actor defaults")
		}
		data, err := hex.DecodeString(fields[1])
		if err != nil {
			return nil, err
		}
		tag, used, err := nbt.ParseNetwork(data)
		if err != nil || used != len(data) {
			return nil, fmt.Errorf("invalid Mansion actor NBT: %v", err)
		}
		object, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("Mansion actor defaults require compounds")
		}
		delete(object, "UUID")
		result[fields[0]] = object
	}
	if len(result) != 3 {
		return nil, fmt.Errorf("expected three original Mansion actor types")
	}
	return result, nil
})

// FinalizeMansionEntity matches marker-spawned actors in a fresh NORMAL world.
// Older/HARD-world equipment enchantment needs the shared local-difficulty
// implementation; the zero-threshold equipment draw still occurs here.
func FinalizeMansionEntity(request PendingStructureEntity, random *Xoroshiro) (nbt.Object, error) {
	defaults, err := loadMansionActorDefaults()
	if err != nil {
		return nil, err
	}
	base, ok := defaults[request.ID]
	if !ok || random == nil {
		return nil, fmt.Errorf("unsupported mansion actor %s", request.ID)
	}
	result := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: base}).Value.(nbt.Object)
	p := request.Position
	result["Pos"] = animalPositionNBT([3]float64{float64(p.X) + .5, float64(p.Y), float64(p.Z) + .5})
	follow := float64(12)
	if request.ID == "minecraft:allay" {
		follow = 48
	}
	attribute := drownedModifier(animalAttribute("follow_range", follow), "minecraft:random_spawn_bonus", 0.0+.11485000000000001*(random.NextDouble()-random.NextDouble()), "add_multiplied_base")
	attributes, _ := result["attributes"].Elements()
	replaced := false
	for i, tag := range attributes {
		object, _ := tag.Object()
		id, _ := object.Text("id")
		if id == "minecraft:generic.follow_range" {
			attributes[i] = attribute
			replaced = true
		}
	}
	if !replaced {
		attributes = append(attributes, attribute)
	}
	result["attributes"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: attributes}
	left := int8(0)
	if random.NextFloat() < .05 {
		left = 1
	}
	result["LeftHanded"] = nbt.Tag{Type: nbt.Byte, Value: left}
	if request.ID != "minecraft:allay" {
		result["CanJoinRaid"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	}
	if request.ID == "minecraft:vindicator" {
		hands, _ := result["HandItems"].Elements()
		hands[0] = bastionActorItem("iron_axe")
		result["HandItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: hands}
		random.NextFloat() // Nonempty main-hand enchantment check at multiplier zero.
	}
	return result, nil
}
