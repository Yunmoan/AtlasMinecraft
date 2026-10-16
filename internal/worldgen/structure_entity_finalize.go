package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

// FinalizeSwampHutEntity captures the actual STRUCTURE spawn stage in the
// caller's WorldGenRegion random stream. Cat's random variant selection still
// consumes a draw before the swamp-hut predicate forces the black variant.
func FinalizeSwampHutEntity(id string, random *Xoroshiro, fullMoon bool) (nbt.Object, error) {
	if id != "minecraft:witch" && id != "minecraft:cat" {
		return nil, fmt.Errorf("unsupported swamp hut entity %s", id)
	}
	bonus := 0.0 + 0.11485000000000001*(random.NextDouble()-random.NextDouble())
	left := int8(0)
	if random.NextFloat() < 0.05 {
		left = 1
	}
	modifier := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:random_spawn_bonus"}, "amount": {Type: nbt.Double, Value: bonus}, "operation": {Type: nbt.String, Value: "add_multiplied_base"}}
	attribute := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:generic.follow_range"}, "base": {Type: nbt.Double, Value: float64(16)}, "modifiers": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: modifier}}}}
	result := nbt.Object{"attributes": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: attribute}}}, "LeftHanded": {Type: nbt.Byte, Value: left}}
	if id == "minecraft:witch" {
		result["CanJoinRaid"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	} else {
		bound := int32(10)
		if fullMoon {
			bound = 11
		}
		random.NextIntBound(bound)
		result["variant"] = nbt.Tag{Type: nbt.String, Value: "minecraft:all_black"}
	}
	return result, nil
}
