package worldgen

import (
	"fmt"
	"math"
)

// StructureChunkPosition is a potential start chunk, not a generated start.
// Vanilla still checks structure-specific terrain and biome suitability after
// applying placement, frequency reduction and exclusion zones.
type StructureChunkPosition struct{ X, Z int32 }

type StructureExclusionZone struct {
	OtherSet   string `json:"other_set"`
	ChunkCount int32  `json:"chunk_count"`
}

// StructurePlacement contains the original vanilla structure-set placement.
// Concentric rings need their biome-relocated positions supplied separately;
// they must not be silently treated as random-spread placements.
type StructurePlacement struct {
	Type                     string                  `json:"type"`
	Salt                     int32                   `json:"salt"`
	Spacing                  int32                   `json:"spacing"`
	Separation               int32                   `json:"separation"`
	SpreadType               string                  `json:"spread_type"`
	Frequency                float32                 `json:"frequency"`
	FrequencyReductionMethod string                  `json:"frequency_reduction_method"`
	LocateOffset             [3]int32                `json:"locate_offset"`
	ExclusionZone            *StructureExclusionZone `json:"exclusion_zone"`
	Distance                 int32                   `json:"distance"`
	Spread                   int32                   `json:"spread"`
	Count                    int32                   `json:"count"`
	PreferredBiomes          string                  `json:"preferred_biomes"`
}

func (p StructurePlacement) validate() error {
	if p.Frequency < 0 || p.Frequency > 1 || math.IsNaN(float64(p.Frequency)) {
		return fmt.Errorf("structure frequency must be in [0,1]")
	}
	switch p.FrequencyReductionMethod {
	case "", "default", "legacy_type_1", "legacy_type_2", "legacy_type_3":
	default:
		return fmt.Errorf("unsupported structure frequency reduction %q", p.FrequencyReductionMethod)
	}
	if p.ExclusionZone != nil && (p.ExclusionZone.ChunkCount < 1 || p.ExclusionZone.ChunkCount > 16) {
		return fmt.Errorf("structure exclusion radius must be in [1,16]")
	}
	switch p.Type {
	case "minecraft:random_spread":
		if p.Spacing <= p.Separation || p.Separation < 0 || p.Spacing > 4096 {
			return fmt.Errorf("structure spacing must exceed separation and be in [1,4096]")
		}
		if p.SpreadType != "" && p.SpreadType != "linear" && p.SpreadType != "triangular" {
			return fmt.Errorf("unsupported structure spread type %q", p.SpreadType)
		}
	case "minecraft:concentric_rings":
		if p.Distance < 0 || p.Distance > 1023 || p.Spread < 0 || p.Spread > 1023 || p.Count < 0 || p.Count > 4095 {
			return fmt.Errorf("invalid concentric-rings structure placement")
		}
	default:
		return fmt.Errorf("unsupported structure placement type %q", p.Type)
	}
	return nil
}

// PotentialChunk follows RandomSpreadStructurePlacement.getPotentialStructureChunk.
// Division must round toward negative infinity; Java's int arithmetic also
// deliberately wraps at the outer world-coordinate boundaries.
func (p StructurePlacement) PotentialChunk(seed int64, chunkX, chunkZ int32) (StructureChunkPosition, error) {
	if err := p.validate(); err != nil {
		return StructureChunkPosition{}, err
	}
	if p.Type != "minecraft:random_spread" {
		return StructureChunkPosition{}, fmt.Errorf("potential chunk requires random_spread, got %s", p.Type)
	}
	regionX, regionZ := structureFloorDiv(chunkX, p.Spacing), structureFloorDiv(chunkZ, p.Spacing)
	random := structureSaltRandom(seed, regionX, regionZ, p.Salt)
	bound := p.Spacing - p.Separation
	nextOffset := func() int32 {
		if p.SpreadType == "triangular" {
			return (random.NextIntBound(bound) + random.NextIntBound(bound)) / 2
		}
		return random.NextIntBound(bound)
	}
	return StructureChunkPosition{regionX*p.Spacing + nextOffset(), regionZ*p.Spacing + nextOffset()}, nil
}

func structureFloorDiv(value, divisor int32) int32 {
	quotient := value / divisor
	if value%divisor < 0 {
		quotient--
	}
	return quotient
}

func structureSaltRandom(seed int64, x, z, salt int32) *LegacyRandom {
	return NewLegacyRandom(int64(x)*341873128712 + int64(z)*132897987541 + seed + int64(salt))
}

// AllowsFrequency follows StructurePlacement.applyAdditionalChunkRestrictions.
// In particular, mineshafts use NextDouble, buried treasure uses a fixed salt,
// and outposts use the old discarded-int regional draw rather than the salt.
func (p StructurePlacement) AllowsFrequency(seed int64, chunkX, chunkZ int32) (bool, error) {
	if err := p.validate(); err != nil {
		return false, err
	}
	if p.Frequency >= 1 {
		return true, nil
	}
	switch p.FrequencyReductionMethod {
	case "", "default":
		// The vanilla default reducer passes (salt, chunkX, chunkZ) to
		// setLargeFeatureWithSalt. This unusual ordering is intentional.
		return structureSaltRandom(seed, p.Salt, chunkX, chunkZ).NextFloat() < p.Frequency, nil
	case "legacy_type_1":
		x, z := chunkX>>4, chunkZ>>4
		random := NewLegacyRandom(int64(x^(z<<4)) ^ seed)
		random.NextInt()
		// Java saturates +Inf when narrowing float to int.
		bound := int32(math.MaxInt32)
		if inverse := float32(1) / p.Frequency; inverse < math.MaxInt32 {
			bound = int32(inverse)
		}
		return random.NextIntBound(bound) == 0, nil
	case "legacy_type_2":
		return structureSaltRandom(seed, chunkX, chunkZ, 10387320).NextFloat() < p.Frequency, nil
	case "legacy_type_3":
		random := NewLegacyRandom(seed)
		xMultiplier, zMultiplier := random.NextLong(), random.NextLong()
		mixed := uint64(int64(chunkX))*xMultiplier ^ uint64(int64(chunkZ))*zMultiplier ^ uint64(seed)
		return NewLegacyRandom(int64(mixed)).NextDouble() < float64(p.Frequency), nil
	default:
		panic("validated structure frequency reducer")
	}
}

// LocatePosition applies the vanilla locate offset to the chunk's minimum
// block corner. It does not identify the actual piece entrance or height.
func (p StructurePlacement) LocatePosition(chunk StructureChunkPosition) FeaturePosition {
	return FeaturePosition{chunk.X*16 + p.LocateOffset[0], p.LocateOffset[1], chunk.Z*16 + p.LocateOffset[2]}
}

// StructurePlacementState checks structure-set placements and exclusion zones.
// It has no mutable sampling state and is safe to share between workers after
// construction. It never reports a potential placement as a generated start.
type StructurePlacementState struct {
	Seed int64
	Sets map[string]StructureSetDefinition
	// RingPositions must contain biome-relocated positions from the ring
	// generation stage before a concentric_rings set can be queried.
	RingPositions map[string][]StructureChunkPosition
}

func (s *StructurePlacementState) IsStructureChunk(setName string, chunkX, chunkZ int32) (bool, error) {
	return s.isStructureChunk(densityName(setName), chunkX, chunkZ, make(map[string]bool))
}

func (s *StructurePlacementState) isStructureChunk(setName string, chunkX, chunkZ int32, visiting map[string]bool) (bool, error) {
	set, found := s.Sets[setName]
	if !found {
		return false, fmt.Errorf("unknown structure set %q", setName)
	}
	if visiting[setName] {
		return false, fmt.Errorf("cyclic structure exclusion involving %s", setName)
	}
	p := set.Placement
	if err := p.validate(); err != nil {
		return false, fmt.Errorf("structure set %s: %w", setName, err)
	}
	placement := false
	if p.Type == "minecraft:random_spread" {
		potential, err := p.PotentialChunk(s.Seed, chunkX, chunkZ)
		if err != nil {
			return false, err
		}
		placement = potential.X == chunkX && potential.Z == chunkZ
	} else {
		positions, found := s.RingPositions[setName]
		if !found {
			return false, fmt.Errorf("structure set %s needs biome-relocated concentric-ring positions", setName)
		}
		for _, position := range positions {
			if position.X == chunkX && position.Z == chunkZ {
				placement = true
				break
			}
		}
	}
	if !placement {
		return false, nil
	}
	allowed, err := p.AllowsFrequency(s.Seed, chunkX, chunkZ)
	if err != nil || !allowed || p.ExclusionZone == nil {
		return allowed, err
	}
	visiting[setName] = true
	defer delete(visiting, setName)
	zone := p.ExclusionZone
	// Vanilla uses an inclusive square in chunk coordinates, and checks
	// placements even where the excluded structure's biome would reject it.
	for dx := -zone.ChunkCount; dx <= zone.ChunkCount; dx++ {
		for dz := -zone.ChunkCount; dz <= zone.ChunkCount; dz++ {
			forbidden, err := s.isStructureChunk(densityName(zone.OtherSet), chunkX+dx, chunkZ+dz, visiting)
			if err != nil {
				return false, err
			}
			if forbidden {
				return false, nil
			}
		}
	}
	return true, nil
}
