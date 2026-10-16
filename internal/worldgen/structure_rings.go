package worldgen

import (
	"fmt"
	"math"
)

// StructureBiomeAtQuart supplies the dimension's climate biome source, not the
// hashed Voronoi biome zoom used by block decoration.
type StructureBiomeAtQuart func(x, y, z int32) (string, error)

// GenerateConcentricRingPositions follows the original ring angles, distances,
// per-position LegacyRandom forks and preferred-biome reservoir selection.
// Biome sampling is at Y=0, within 112 blocks of each candidate chunk center.
// Every returned position is still only a placement; it is not a structure.
func GenerateConcentricRingPositions(seed int64, placement StructurePlacement, preferredBiomes []string, biomeAtQuart StructureBiomeAtQuart) ([]StructureChunkPosition, error) {
	if err := placement.validate(); err != nil {
		return nil, err
	}
	if placement.Type != "minecraft:concentric_rings" {
		return nil, fmt.Errorf("ring positions require concentric_rings, got %s", placement.Type)
	}
	if placement.Count == 0 {
		return []StructureChunkPosition{}, nil
	}
	if placement.Spread == 0 {
		return nil, fmt.Errorf("nonempty concentric rings require positive spread")
	}
	if len(preferredBiomes) != 0 && biomeAtQuart == nil {
		return nil, fmt.Errorf("concentric-ring biome relocation needs a biome source")
	}
	preferred := make(map[string]bool, len(preferredBiomes))
	for _, name := range preferredBiomes {
		preferred[densityName(name)] = true
	}
	random := NewLegacyRandom(seed)
	angle := random.NextDouble() * math.Pi * 2
	spread := placement.Spread
	withinRing, ring := int32(0), int32(0)
	positions := make([]StructureChunkPosition, 0, placement.Count)
	for index := int32(0); index < placement.Count; index++ {
		distance := float64(4*placement.Distance+placement.Distance*ring*6) + (random.NextDouble()-0.5)*(float64(placement.Distance)*2.5)
		// Java Math.round is floor(value+0.5), including negative values.
		position := StructureChunkPosition{int32(math.Floor(math.Cos(angle)*distance + 0.5)), int32(math.Floor(math.Sin(angle)*distance + 0.5))}
		fork := random.Fork()
		if len(preferred) != 0 {
			relocated, found, err := relocateRingBiome(position, preferred, fork, biomeAtQuart)
			if err != nil {
				return nil, err
			}
			if found {
				position = relocated
			}
		}
		positions = append(positions, position)
		angle += math.Pi * 2 / float64(spread)
		withinRing++
		if withinRing == spread {
			withinRing = 0
			ring++
			spread += 2 * spread / (ring + 1)
			spread = min(spread, placement.Count-index)
			angle += random.NextDouble() * math.Pi * 2
		}
	}
	return positions, nil
}

func relocateRingBiome(position StructureChunkPosition, preferred map[string]bool, random *LegacyRandom, biomeAtQuart StructureBiomeAtQuart) (StructureChunkPosition, bool, error) {
	quartX, quartZ := (position.X*16+8)>>2, (position.Z*16+8)>>2
	var result StructureChunkPosition
	matches := int32(0)
	for dz := int32(-28); dz <= 28; dz++ {
		for dx := int32(-28); dx <= 28; dx++ {
			qx, qz := quartX+dx, quartZ+dz
			biome, err := biomeAtQuart(qx, 0, qz)
			if err != nil {
				return result, false, err
			}
			if !preferred[biome] {
				continue
			}
			if matches == 0 || random.NextIntBound(matches+1) == 0 {
				result = StructureChunkPosition{(qx * 4) >> 4, (qz * 4) >> 4}
			}
			matches++
		}
	}
	return result, matches != 0, nil
}

// PopulateOverworldStrongholdRings prepares the placement prerequisite using
// Atlas's reference-checked climate biome source. Invoke before publishing the
// placement state to concurrent workers; this method mutates its ring map.
func (state *StructurePlacementState) PopulateOverworldStrongholdRings() error {
	set, found := state.Sets["minecraft:strongholds"]
	if !found {
		return fmt.Errorf("missing stronghold structure set")
	}
	preferred, err := StructurePreferredBiomeNames(set.Placement)
	if err != nil {
		return err
	}
	sampler, err := NewOverworldBiomeSampler(state.Seed)
	if err != nil {
		return err
	}
	positions, err := GenerateConcentricRingPositions(state.Seed, set.Placement, preferred, sampler.BiomeAtQuart)
	if err != nil {
		return err
	}
	if state.RingPositions == nil {
		state.RingPositions = make(map[string][]StructureChunkPosition)
	}
	state.RingPositions["minecraft:strongholds"] = positions
	return nil
}
