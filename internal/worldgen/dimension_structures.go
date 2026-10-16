package worldgen

import (
	"fmt"
	"sort"
	"sync"
)

type dimensionStructureBlockKey struct {
	name                         string
	ownerX, ownerZ, clipX, clipZ int32
}
type dimensionStructureCache struct {
	placement *StructurePlacementState
	mu        sync.Mutex
	clock     uint64
	entries   map[featureChunkPosition]*overworldStructureEntry
	blocks    map[dimensionStructureBlockKey]*structureBlockEntry
}

func newDimensionStructureCache(seed int64) (*dimensionStructureCache, error) {
	placement, err := LoadStructurePlacementState(seed)
	if err != nil {
		return nil, err
	}
	return &dimensionStructureCache{placement: placement, entries: make(map[featureChunkPosition]*overworldStructureEntry), blocks: make(map[dimensionStructureBlockKey]*structureBlockEntry)}, nil
}

func (g *DimensionGenerator) StructuresEnabled() bool { return g.structures != nil }

func (g *DimensionGenerator) structureStartsAt(cx, cz int32) ([]StructureStart, error) {
	if g.structures == nil || cx < -1875000 || cx >= 1875000 || cz < -1875000 || cz >= 1875000 {
		return nil, nil
	}
	c := g.structures
	key := featureChunkPosition{cx, cz}
	c.mu.Lock()
	c.clock++
	if entry := c.entries[key]; entry != nil {
		entry.used = c.clock
		c.mu.Unlock()
		<-entry.ready
		return entry.starts, entry.err
	}
	entry := &overworldStructureEntry{ready: make(chan struct{}), used: c.clock}
	c.entries[key] = entry
	c.mu.Unlock()
	starts, err := g.generateStructureStarts(cx, cz)
	c.mu.Lock()
	entry.starts, entry.err, entry.complete = starts, err, true
	close(entry.ready)
	if err != nil {
		delete(c.entries, key)
	}
	for len(c.entries) > structureStartsCacheLimit {
		var oldest featureChunkPosition
		used, found := ^uint64(0), false
		for position, candidate := range c.entries {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(c.entries, oldest)
	}
	c.mu.Unlock()
	return starts, err
}

func (g *DimensionGenerator) generateStructureStarts(cx, cz int32) ([]StructureStart, error) {
	var starts []StructureStart
	var biome *DimensionBiomeSampler
	sets := []string{"minecraft:end_cities"}
	if g.dimension == "nether" {
		sets = []string{"minecraft:nether_complexes", "minecraft:ruined_portals", "minecraft:nether_fossils"}
	}
	for _, name := range sets {
		eligible, err := g.structures.placement.IsStructureChunk(name, cx, cz)
		if err != nil {
			return nil, err
		}
		if !eligible {
			continue
		}
		if biome == nil {
			biome, err = NewDimensionBiomeSampler(g.seed, g.dimension)
			if err != nil {
				return nil, err
			}
		}
		start, err := SelectStructureSetStart(g.seed, cx, cz, g.structures.placement.Sets[name], func(structure string) (*StructureStart, error) {
			switch structure {
			case "minecraft:end_city":
				return GenerateEndCityStructureStart(g.seed, cx, cz, biome.BiomeAtQuart, func(x, z int32) (int32, error) {
					// Starts run inside terrain worker slots. Read the raw noise
					// column directly; recursing into baseEntryAt would deadlock.
					column, err := g.noise.BaseStructureColumnAt(x, z)
					if err != nil {
						return 0, err
					}
					return column.OccupiedHeight(false), nil
				})
			case "minecraft:fortress":
				return GenerateNetherFortressStructureStart(g.seed, cx, cz, biome.BiomeAtQuart)
			case "minecraft:bastion_remnant":
				return GenerateJigsawStructureStart(g.seed, cx, cz, structure, JigsawStartSources{MinY: g.noise.settings.MinY, Height: 256, GenerationDepth: g.noise.settings.Height, BiomeAtQuart: biome.BiomeAtQuart})
			case "minecraft:ruined_portal_nether":
				return GenerateRuinedPortalStructureStart(g.seed, cx, cz, structure, RuinedPortalStartSources{MinY: g.noise.settings.MinY, BiomeAtQuart: biome.BiomeAtQuart, BaseColumnAt: g.noise.BaseStructureColumnAt})
			case "minecraft:nether_fossil":
				return GenerateNetherFossilStructureStart(g.seed, cx, cz, NetherFossilStartSources{MinY: g.noise.settings.MinY, GenerationDepth: g.noise.settings.Height, SeaLevel: g.noise.settings.SeaLevel, BiomeAtQuart: biome.BiomeAtQuart, BaseColumnAt: g.noise.BaseStructureColumnAt})
			default:
				// The six Overworld portal variants cannot pass their original biome
				// tags in any Nether biome. Their separate start streams do not consume
				// the weighted set's random source, so rejecting here retains selection.
				return nil, nil
			}
		})
		if err != nil {
			return nil, fmt.Errorf("%s structure set %s in %d,%d: %w", g.dimension, name, cx, cz, err)
		}
		if start != nil {
			starts = append(starts, *start)
		}
	}
	return starts, nil
}

func (g *DimensionGenerator) structureNeighborhood(cx, cz int32) ([]StructureStart, error) {
	if g.structures == nil {
		return nil, nil
	}
	var starts []StructureStart
	for dx := int32(-8); dx <= 8; dx++ {
		for dz := int32(-8); dz <= 8; dz++ {
			owned, err := g.structureStartsAt(cx+dx, cz+dz)
			if err != nil {
				return nil, err
			}
			starts = append(starts, owned...)
		}
	}
	return starts, nil
}

func (g *DimensionGenerator) decorationStructures(bases map[featureChunkPosition]*dimensionBase, cx, cz int32) []StructureStart {
	if g.structures == nil {
		return nil
	}
	seen := make(map[structureBlockKey]bool)
	var result []StructureStart
	for _, base := range bases {
		for _, start := range base.starts {
			key := structureBlockKey{start.Structure, start.Chunk.X, start.Chunk.Z}
			if seen[key] {
				continue
			}
			seen[key] = true
			box, valid := structureWriteBox(start)
			if valid && box.MaxX >= (cx-1)*16 && box.MinX <= (cx+1)*16+15 && box.MaxZ >= (cz-1)*16 && box.MinZ <= (cz+1)*16+15 {
				result = append(result, start)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		oa, _ := OriginalStructureDecorationOrder(a.Structure)
		ob, _ := OriginalStructureDecorationOrder(b.Structure)
		if oa.Step != ob.Step {
			return oa.Step < ob.Step
		}
		if oa.Index != ob.Index {
			return oa.Index < ob.Index
		}
		if a.Chunk.X != b.Chunk.X {
			return a.Chunk.X < b.Chunk.X
		}
		return a.Chunk.Z < b.Chunk.Z
	})
	return result
}
