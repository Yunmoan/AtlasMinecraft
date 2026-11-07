package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

//go:embed monument_actor_defaults_1.21.1.txt
var monumentActorDefaultsText string

var loadMonumentActorDefault = sync.OnceValues(func() (nbt.Object, error) {
	fields := strings.Fields(monumentActorDefaultsText)
	if len(fields) != 2 || fields[0] != "minecraft:elder_guardian" {
		return nil, fmt.Errorf("invalid monument actor defaults")
	}
	encoded, err := hex.DecodeString(fields[1])
	if err != nil {
		return nil, err
	}
	tag, used, err := nbt.ParseNetwork(encoded)
	if err != nil || used != len(encoded) {
		return nil, fmt.Errorf("invalid monument actor NBT: %v", err)
	}
	object, ok := tag.Object()
	if !ok {
		return nil, fmt.Errorf("monument actor default must be a compound")
	}
	return object, nil
})

// FinalizeMonumentEntity reproduces Mob's STRUCTURE finalization for an elder
// guardian. Its constructor already sets persistence and max health. The
// follow-range bonus and handedness consume the region's independent stream.
func FinalizeMonumentEntity(request PendingStructureEntity, random *Xoroshiro) (nbt.Object, error) {
	if request.ID != "minecraft:elder_guardian" || random == nil {
		return nil, fmt.Errorf("invalid monument actor or random source")
	}
	base, err := loadMonumentActorDefault()
	if err != nil {
		return nil, err
	}
	result := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: base}).Value.(nbt.Object)
	p := request.Position
	result["Pos"] = animalPositionNBT([3]float64{float64(p.X) + 0.5, float64(p.Y), float64(p.Z) + 0.5})
	bonus := 0.0 + 0.11485000000000001*(random.NextDouble()-random.NextDouble())
	modifier := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:random_spawn_bonus"}, "amount": {Type: nbt.Double, Value: bonus}, "operation": {Type: nbt.String, Value: "add_multiplied_base"}}
	attribute := animalAttribute("follow_range", 16)
	attribute.Value.(nbt.Object)["modifiers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: modifier}}}
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
	if random.NextFloat() < 0.05 {
		left = 1
	}
	result["LeftHanded"] = nbt.Tag{Type: nbt.Byte, Value: left}
	return result, nil
}
