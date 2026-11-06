package worldgen

import "atlas/internal/worldgen/randomsource"

// The generator keeps its existing public entry points while the algorithms
// and their reference tests live in a standalone, registry-free package.
type Xoroshiro = randomsource.Xoroshiro
type PositionalRandom = randomsource.PositionalRandom
type LegacyRandom = randomsource.LegacyRandom
type LegacyPositionalRandom = randomsource.LegacyPositionalRandom
type WorldgenRandom = randomsource.WorldgenRandom

func NewXoroshiro(seed int64) *Xoroshiro           { return randomsource.NewXoroshiro(seed) }
func NewXoroshiroPair(lo, hi uint64) *Xoroshiro    { return randomsource.NewXoroshiroPair(lo, hi) }
func NewLegacyRandom(seed int64) *LegacyRandom     { return randomsource.NewLegacyRandom(seed) }
func NewWorldgenRandom(seed int64) *WorldgenRandom { return randomsource.NewWorldgenRandom(seed) }
func positionSeed(x, y, z int32) int64             { return randomsource.PositionSeed(x, y, z) }
