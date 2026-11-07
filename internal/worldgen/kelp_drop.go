package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

// NewKelpBlockDropEntity follows Block.popResource for a non-explosion kelp
// break. Both original block loot tables yield exactly one kelp. Position draws
// occur even when doTileDrops=false; motion/UUID draws require an actual item.
func NewKelpBlockDropEntity(p FeaturePosition, drops bool, worldRandom, entityRandom *LegacyRandom) (nbt.Object, error) {
	if worldRandom == nil {
		return nil, fmt.Errorf("kelp block drop requires world randomness")
	}
	position := [3]float64{
		float64(p.X) + 0.5 + (-0.25 + worldRandom.NextDouble()*0.5),
		float64(p.Y) + 0.5 + (-0.25 + worldRandom.NextDouble()*0.5) - 0.125,
		float64(p.Z) + 0.5 + (-0.25 + worldRandom.NextDouble()*0.5),
	}
	if !drops {
		return nil, nil
	}
	return newDroppedItemEntity(position, structureState("kelp"), worldRandom, entityRandom)
}
