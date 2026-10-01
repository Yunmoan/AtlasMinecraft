package worldgen

type CarverStart struct {
	SourceX, SourceZ int32
	Index            int
	Name             string
}

var overworldCarvers = [...]struct {
	name        string
	probability float32
}{
	{"minecraft:cave", 0.15},
	{"minecraft:cave_extra_underground", 0.07},
	{"minecraft:canyon", 0.01},
}

func carverRandom(seed int64, sourceX, sourceZ int32, index int) *LegacyRandom {
	carverSeed := seed + int64(index)
	random := NewLegacyRandom(carverSeed)
	xMultiplier, zMultiplier := random.NextLong(), random.NextLong()
	mixed := uint64(int64(sourceX))*xMultiplier ^ uint64(int64(sourceZ))*zMultiplier ^ uint64(carverSeed)
	return NewLegacyRandom(int64(mixed))
}

func OverworldCarverStarts(seed int64, targetX, targetZ int32) []CarverStart {
	var starts []CarverStart
	for dx := int32(-8); dx <= 8; dx++ {
		for dz := int32(-8); dz <= 8; dz++ {
			sourceX, sourceZ := targetX+dx, targetZ+dz
			for index, carver := range overworldCarvers {
				if carverRandom(seed, sourceX, sourceZ, index).NextFloat() <= carver.probability {
					starts = append(starts, CarverStart{sourceX, sourceZ, index, carver.name})
				}
			}
		}
	}
	return starts
}
