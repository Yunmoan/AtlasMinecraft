package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"math"
)

// NewFallingBlockDropEntity reproduces Entity.spawnAtLocation for supported
// terrain blocks. World motion randomness and the fresh entity's independent
// RNG are separate. The latter consumes UUID draws before bob offset and yaw.
// The returned NBT includes its UUID and id, ready for an entity-region commit.
func NewFallingBlockDropEntity(position [3]float64, state SurfaceBlockState, worldRandom, entityRandom *LegacyRandom) (nbt.Object, error) {
	if state.Name != "minecraft:sand" && state.Name != "minecraft:red_sand" && state.Name != "minecraft:gravel" && !brushableBlock(state.Name) {
		return nil, fmt.Errorf("falling item constructor is not implemented: %s", state.Name)
	}
	return newDroppedItemEntity(position, state, worldRandom, entityRandom)
}

func newDroppedItemEntity(position [3]float64, state SurfaceBlockState, worldRandom, entityRandom *LegacyRandom) (nbt.Object, error) {
	for _, coordinate := range position {
		if math.IsNaN(coordinate) || math.IsInf(coordinate, 0) {
			return nil, fmt.Errorf("invalid falling item position")
		}
	}
	if worldRandom == nil || entityRandom == nil {
		return nil, fmt.Errorf("falling item requires independent random sources")
	}
	velocity := [3]float64{worldRandom.NextDouble()*0.2 - 0.1, 0.2, worldRandom.NextDouble()*0.2 - 0.1}
	hi := entityRandom.NextLong()&0xffffffffffff0fff | 0x4000
	lo := entityRandom.NextLong()&0x3fffffffffffffff | 0x8000000000000000
	entityRandom.NextFloat() // Unsaved bob offset.
	yaw := entityRandom.NextFloat() * 360
	vector := func(values [3]float64) nbt.Tag {
		tags := make([]nbt.Tag, 3)
		for i, v := range values {
			tags[i] = nbt.Tag{Type: nbt.Double, Value: v}
		}
		return nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: tags}
	}
	return nbt.Object{
		"id":   {Type: nbt.String, Value: "minecraft:item"},
		"UUID": {Type: nbt.IntArray, Value: []int32{int32(hi >> 32), int32(hi), int32(lo >> 32), int32(lo)}},
		"Pos":  vector(position), "Motion": vector(velocity),
		"Rotation": {Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{{Type: nbt.Float, Value: yaw}, {Type: nbt.Float, Value: float32(0)}}},
		"Air":      {Type: nbt.Short, Value: int16(300)}, "Fire": {Type: nbt.Short, Value: int16(-1)},
		"FallDistance": {Type: nbt.Float, Value: float32(0)}, "Invulnerable": {Type: nbt.Byte, Value: int8(0)},
		"OnGround": {Type: nbt.Byte, Value: int8(0)}, "PortalCooldown": {Type: nbt.Int, Value: int32(0)},
		"Age": {Type: nbt.Short, Value: int16(0)}, "Health": {Type: nbt.Short, Value: int16(5)}, "PickupDelay": {Type: nbt.Short, Value: int16(10)},
		"Item": {Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: state.Name}, "count": {Type: nbt.Int, Value: int32(1)}}},
	}, nil
}
