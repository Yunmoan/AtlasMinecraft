package worldgen

import (
	"fmt"
	"strings"
	"sync"
)

// Starts are computed before noise, from original placement, biome and base
// height sources. The cache contains immutable STRUCTURE_STARTS results;
// decoration mutates private clones, never another worker's start.
type overworldStructureCache struct {
	placement        *StructurePlacementState
	mu               sync.Mutex
	clock            uint64
	entries          map[featureChunkPosition]*overworldStructureEntry
	blocks           map[structureBlockKey]*structureBlockEntry
	jigsawBlocks     map[dimensionStructureBlockKey]*structureBlockEntry
	strongholdBlocks map[dimensionStructureBlockKey]*structureBlockEntry
	strongholdSaved  map[featureChunkPosition]StructureStart
	ringsOnce        sync.Once
	ringsErr         error
}

type overworldStructureEntry struct {
	ready    chan struct{}
	starts   []StructureStart
	err      error
	used     uint64
	complete bool
}

const structureStartsCacheLimit = 2048

var supportedOverworldStructures = []struct{ set, name string }{
	{"minecraft:strongholds", "minecraft:stronghold"},
	{"minecraft:desert_pyramids", "minecraft:desert_pyramid"},
	{"minecraft:swamp_huts", "minecraft:swamp_hut"},
	{"minecraft:jungle_temples", "minecraft:jungle_pyramid"},
	{"minecraft:igloos", "minecraft:igloo"},
	{"minecraft:buried_treasures", "minecraft:buried_treasure"},
	{"minecraft:shipwrecks", "minecraft:shipwreck"},
	{"minecraft:shipwrecks", "minecraft:shipwreck_beached"},
	{"minecraft:ocean_ruins", "minecraft:ocean_ruin_cold"},
	{"minecraft:ocean_ruins", "minecraft:ocean_ruin_warm"},
	{"minecraft:ancient_cities", "minecraft:ancient_city"},
	{"minecraft:trail_ruins", "minecraft:trail_ruins"},
	{"minecraft:trial_chambers", "minecraft:trial_chambers"},
	{"minecraft:pillager_outposts", "minecraft:pillager_outpost"},
	{"minecraft:villages", "minecraft:village_desert"},
	{"minecraft:villages", "minecraft:village_plains"},
	{"minecraft:villages", "minecraft:village_savanna"},
	{"minecraft:villages", "minecraft:village_snowy"},
	{"minecraft:villages", "minecraft:village_taiga"},
	{"minecraft:mineshafts", "minecraft:mineshaft"},
	{"minecraft:mineshafts", "minecraft:mineshaft_mesa"},
	{"minecraft:ruined_portals", "minecraft:ruined_portal"},
	{"minecraft:ruined_portals", "minecraft:ruined_portal_desert"},
	{"minecraft:ruined_portals", "minecraft:ruined_portal_jungle"},
	{"minecraft:ruined_portals", "minecraft:ruined_portal_mountain"},
	{"minecraft:ruined_portals", "minecraft:ruined_portal_ocean"},
	{"minecraft:ruined_portals", "minecraft:ruined_portal_swamp"},
}

func (g *OverworldGenerator) StructuresEnabled() bool { return g.structures != nil }

func newOverworldStructureCache(seed int64) (*overworldStructureCache, error) {
	placement, err := LoadStructurePlacementState(seed)
	if err != nil {
		return nil, err
	}
	return &overworldStructureCache{placement: placement, entries: make(map[featureChunkPosition]*overworldStructureEntry), blocks: make(map[structureBlockKey]*structureBlockEntry), jigsawBlocks: make(map[dimensionStructureBlockKey]*structureBlockEntry), strongholdBlocks: make(map[dimensionStructureBlockKey]*structureBlockEntry), strongholdSaved: make(map[featureChunkPosition]StructureStart)}, nil
}

func (g *OverworldGenerator) structureStartsAt(cx, cz int32) ([]StructureStart, error) {
	if g.structures == nil {
		return nil, nil
	}
	if cx < -1875000 || cx >= 1875000 || cz < -1875000 || cz >= 1875000 {
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

func (g *OverworldGenerator) generateStructureStarts(cx, cz int32) ([]StructureStart, error) {
	var starts []StructureStart
	var biome *OverworldBiomeSampler
	seenSets := make(map[string]bool)
	for _, candidate := range supportedOverworldStructures {
		if seenSets[candidate.set] {
			continue
		}
		seenSets[candidate.set] = true
		var eligible bool
		var err error
		if candidate.name == "minecraft:stronghold" {
			eligible, err = g.strongholdPlacementAt(cx, cz)
		} else {
			eligible, err = g.structures.placement.IsStructureChunk(candidate.set, cx, cz)
		}
		if err != nil {
			return nil, err
		}
		if !eligible {
			continue
		}
		if biome == nil {
			biome, err = newOverworldBiomeSampler(g.seed, g.noise.settings)
			if err != nil {
				return nil, err
			}
		}
		start, err := SelectStructureSetStart(g.seed, cx, cz, g.structures.placement.Sets[candidate.set], func(name string) (*StructureStart, error) {
			switch {
			case name == "minecraft:stronghold":
				return GenerateStrongholdStructureStart(g.seed, cx, cz, StrongholdStartSources{MinY: g.noise.settings.MinY, SeaLevel: g.noise.settings.SeaLevel, BiomeAtQuart: biome.BiomeAtQuart})
			case isOverworldJigsawStructure(name):
				return GenerateJigsawStructureStart(g.seed, cx, cz, name, JigsawStartSources{
					MinY: g.noise.settings.MinY, Height: g.noise.settings.Height, GenerationDepth: g.noise.settings.Height,
					BiomeAtQuart: biome.BiomeAtQuart,
					BaseHeightAt: func(x, z int32, kind string) (int32, error) {
						if kind != "WORLD_SURFACE_WG" {
							return 0, fmt.Errorf("unsupported Jigsaw base heightmap %s", kind)
						}
						height, err := g.noise.BaseOccupiedHeightAt(x, z)
						return height + 1, err
					},
				})
			case isOverworldMineshaftStructure(name):
				return GenerateMineshaftStructureStart(g.seed, cx, cz, name == "minecraft:mineshaft_mesa", MineshaftStartSources{
					MinY: g.noise.settings.MinY, SeaLevel: g.noise.settings.SeaLevel, BiomeAtQuart: biome.BiomeAtQuart,
					WorldSurfaceHeightAt: func(x, z int32) (int32, error) {
						height, err := g.noise.BaseOccupiedHeightAt(x, z)
						return height + 1, err
					},
				})
			case name == "minecraft:igloo":
				return GenerateIglooStructureStart(g.seed, cx, cz, biome.BiomeAtQuart, g.noise.BaseOccupiedHeightAt)
			case name == "minecraft:buried_treasure":
				return GenerateBuriedTreasureStructureStart(cx, cz, biome.BiomeAtQuart, func(x, z int32) (int32, error) {
					column, err := g.noise.BaseStructureColumnAt(x, z)
					if err != nil {
						return 0, err
					}
					return column.OccupiedHeight(true), nil
				})
			case name == "minecraft:shipwreck" || name == "minecraft:shipwreck_beached":
				beached := name == "minecraft:shipwreck_beached"
				return GenerateShipwreckStructureStart(g.seed, cx, cz, beached, biome.BiomeAtQuart, func(x, z int32) (int32, error) {
					if beached {
						return g.noise.BaseOccupiedHeightAt(x, z)
					}
					column, err := g.noise.BaseStructureColumnAt(x, z)
					if err != nil {
						return 0, err
					}
					return column.OccupiedHeight(true), nil
				})
			case name == "minecraft:ocean_ruin_cold" || name == "minecraft:ocean_ruin_warm":
				return GenerateOceanRuinStructureStart(g.seed, cx, cz, name == "minecraft:ocean_ruin_cold", biome.BiomeAtQuart, func(x, z int32) (int32, error) {
					column, err := g.noise.BaseStructureColumnAt(x, z)
					if err != nil {
						return 0, err
					}
					return column.OccupiedHeight(true), nil
				})
			case strings.HasPrefix(name, "minecraft:ruined_portal"):
				return GenerateRuinedPortalStructureStart(g.seed, cx, cz, name, RuinedPortalStartSources{MinY: g.noise.settings.MinY, BiomeAtQuart: biome.BiomeAtQuart, BaseColumnAt: g.noise.BaseStructureColumnAt})
			default:
				return GenerateScatteredStructureStart(g.seed, name, cx, cz, g.noise.settings.SeaLevel, biome.BiomeAtQuart, g.noise.BaseOccupiedHeightAt)
			}
		})
		if err != nil {
			return nil, fmt.Errorf("structure set %s in %d,%d: %w", candidate.set, cx, cz, err)
		}
		if start != nil {
			starts = append(starts, *start)
		}
	}
	return starts, nil
}

func (g *OverworldGenerator) structureNeighborhood(cx, cz int32) ([]StructureStart, error) {
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

func structureTerrainPieces(starts []StructureStart) []StructureTerrainPiece {
	var pieces []StructureTerrainPiece
	for _, start := range starts {
		for _, piece := range start.Pieces {
			terrain := StructureTerrainPiece{RigidPiece: StructureRigidPiece{Box: piece.Box, TerrainAdjustment: start.TerrainAdjustment}}
			if piece.ID == "minecraft:jigsaw" {
				terrain.Jigsaw = true
				element, _ := piece.Data.Object("pool_element")
				terrain.Projection, _ = element.Text("projection")
				delta, _ := piece.Data.Number("ground_level_delta")
				terrain.RigidPiece.GroundLevelDelta = int32(delta)
				if junctions, ok := piece.Data["junctions"].Elements(); ok {
					for _, tag := range junctions {
						junction, ok := tag.Object()
						if !ok {
							continue
						}
						x, _ := junction.Number("source_x")
						y, _ := junction.Number("source_ground_y")
						z, _ := junction.Number("source_z")
						terrain.Junctions = append(terrain.Junctions, StructureJunction{SourceX: int32(x), SourceGroundY: int32(y), SourceZ: int32(z)})
					}
				}
			}
			pieces = append(pieces, terrain)
		}
	}
	return pieces
}
