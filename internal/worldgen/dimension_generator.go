package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"runtime"
	"sort"
	"sync"
)

// GeneratedDimensionChunk uses the dimension's full build height (0..256),
// while its noise generation depth can be smaller. Skipped stages remain
// explicit until their original algorithms and structure pipeline are ported.
type GeneratedDimensionChunk struct {
	Terrain         *NoiseChunkResult
	Biomes          [][64]string
	BlockEntities   []nbt.Object
	SkippedFeatures []string
	SkippedStages   []string
	PendingEntities []PendingStructureEntity
	Structures      nbt.Object
}

type dimensionBase struct {
	terrain    *NoiseChunkResult
	biomes     [][64]string
	biomeNames []string
	starts     []StructureStart
}

type dimensionCacheEntry struct {
	ready    chan struct{}
	base     *dimensionBase
	err      error
	used     uint64
	complete bool
}

// DimensionGenerator owns immutable cached terrain. Each request privately
// replays source-chunk decorations so worker scheduling and exploration order
// cannot rewrite saved neighbors or change the requested chunk.
type DimensionGenerator struct {
	seed         int64
	dimension    string
	noise        *DimensionNoiseGenerator
	order        *featureOrder
	plans        map[string]decorationFeature
	mu           sync.Mutex
	clock        uint64
	cache        map[featureChunkPosition]*dimensionCacheEntry
	structures   *dimensionStructureCache
	terrainSlots chan struct{}
}

// Configure before starting generation. All final-chunk jobs share the same
// terrain budget, including the neighborhoods prepared by structure placers.
func (g *DimensionGenerator) SetTerrainWorkers(workers int) {
	g.terrainSlots = make(chan struct{}, max(1, min(256, workers)))
}

func NewDimensionGenerator(seed int64, dimension string) (*DimensionGenerator, error) {
	return NewDimensionGeneratorWithOptions(seed, dimension, DimensionGeneratorOptions{})
}

type DimensionGeneratorOptions struct{ GenerateStructures bool }

func NewDimensionGeneratorWithOptions(seed int64, dimension string, options DimensionGeneratorOptions) (*DimensionGenerator, error) {
	noise, err := NewDimensionNoiseGenerator(seed, dimension)
	if err != nil {
		return nil, err
	}
	order, err := loadDimensionFeatureOrder(dimension)
	if err != nil {
		return nil, err
	}
	g := &DimensionGenerator{seed: seed, dimension: dimension, noise: noise, order: order,
		plans: make(map[string]decorationFeature), cache: make(map[featureChunkPosition]*dimensionCacheEntry)}
	g.SetTerrainWorkers(min(4, runtime.GOMAXPROCS(0)))
	if options.GenerateStructures {
		g.structures, err = newDimensionStructureCache(seed)
		if err != nil {
			return nil, err
		}
	}
	for _, feature := range order.features {
		placed, err := LoadPlacedFeatureDefinition(feature.Name)
		if err != nil {
			return nil, err
		}
		configured, err := LoadConfiguredFeatureDefinition(placed.Feature)
		if err != nil {
			return nil, err
		}
		plan := decorationFeature{feature: feature}
		if configured.Type == "minecraft:ore" {
			config, err := LoadOreConfiguration(placed.Feature)
			if err != nil {
				return nil, err
			}
			plan.place = func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
				return PlaceOreFeature(region, random, origin, config)
			}
		} else {
			place, err := compileFeaturePlan(placed.Feature, 0)
			if err != nil {
				continue
			}
			plan.place = place
		}
		plan.placement, err = compilePlacementTemplate(placed, &feature)
		if err != nil {
			return nil, err
		}
		g.plans[feature.Name] = plan
	}
	return g, nil
}

func (g *DimensionGenerator) generateBase(cx, cz int32) (*dimensionBase, error) {
	starts, err := g.structureNeighborhood(cx, cz)
	if err != nil {
		return nil, err
	}
	biomes, err := NewDimensionBiomeSampler(g.seed, g.dimension)
	if err != nil {
		return nil, err
	}
	terrain, err := g.noise.GenerateNoiseWithStructures(cx, cz, structureTerrainPieces(starts))
	if err != nil {
		return nil, err
	}
	if err := g.noise.ApplySurface(terrain, biomes.BiomeAtQuart); err != nil {
		return nil, err
	}
	if g.dimension == "nether" {
		if err := g.noise.ApplyNetherCarvers(terrain); err != nil {
			return nil, err
		}
	}
	terrain.CarvingMask = nil
	// Java dimension types allow blocks above the 128-block noise ceiling.
	// Keep this space in the chunk and biome palettes for decorations/players.
	expandDimensionBuildHeight(terrain, 256)
	base := &dimensionBase{terrain: terrain, biomes: make([][64]string, terrain.Height/16), starts: starts}
	present := make(map[string]bool)
	for si := range base.biomes {
		for i := range base.biomes[si] {
			biome, err := biomes.BiomeAtQuart(cx*4+int32(i&3), terrain.MinY/4+int32(si*4+i/16), cz*4+int32((i>>2)&3))
			if err != nil {
				return nil, err
			}
			base.biomes[si][i] = biome
			present[biome] = true
		}
	}
	for biome := range present {
		base.biomeNames = append(base.biomeNames, biome)
	}
	sort.Strings(base.biomeNames)
	return base, nil
}

func expandDimensionBuildHeight(chunk *NoiseChunkResult, height int32) {
	if chunk.Height >= height {
		return
	}
	n := int(height) * 256
	blocks := make([]NoiseMaterial, n)
	copy(blocks, chunk.Blocks)
	chunk.Blocks = blocks
	updates := make([]bool, n)
	copy(updates, chunk.FluidUpdates)
	chunk.FluidUpdates = updates
	post := make([]bool, n)
	copy(post, chunk.PostProcessing)
	chunk.PostProcessing = post
	chunk.Height = height
}

func (g *DimensionGenerator) baseEntryAt(cx, cz int32) *dimensionCacheEntry {
	key := featureChunkPosition{cx, cz}
	g.mu.Lock()
	g.clock++
	if entry := g.cache[key]; entry != nil {
		entry.used = g.clock
		g.mu.Unlock()
		return entry
	}
	entry := &dimensionCacheEntry{ready: make(chan struct{}), used: g.clock}
	g.cache[key] = entry
	g.mu.Unlock()
	g.terrainSlots <- struct{}{}
	base, err := g.generateBase(cx, cz)
	<-g.terrainSlots
	g.mu.Lock()
	entry.base, entry.err, entry.complete = base, err, true
	close(entry.ready)
	if err != nil {
		delete(g.cache, key)
	}
	for len(g.cache) > overworldCacheLimit {
		var oldest featureChunkPosition
		used, found := ^uint64(0), false
		for position, candidate := range g.cache {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(g.cache, oldest)
	}
	g.mu.Unlock()
	return entry
}

func (g *DimensionGenerator) decorationBases(cx, cz int32) (map[featureChunkPosition]*dimensionBase, []*NoiseChunkResult, error) {
	if cx < -1874998 || cx >= 1874998 || cz < -1874998 || cz >= 1874998 {
		return nil, nil, fmt.Errorf("chunk %d,%d has a generation neighborhood outside the world bounds", cx, cz)
	}
	bases := make(map[featureChunkPosition]*dimensionBase, 25)
	chunks := make([]*NoiseChunkResult, 0, 25)
	var entries [5][5]*dimensionCacheEntry
	jobs := make(chan featureChunkPosition, 25)
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			jobs <- featureChunkPosition{dx, dz}
		}
	}
	close(jobs)
	var ready sync.WaitGroup
	for worker := 0; worker < min(25, cap(g.terrainSlots)); worker++ {
		ready.Add(1)
		go func() {
			defer ready.Done()
			for offset := range jobs {
				entries[offset.chunkX+2][offset.chunkZ+2] = g.baseEntryAt(cx+offset.chunkX, cz+offset.chunkZ)
			}
		}()
	}
	ready.Wait()
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			entry := entries[dx+2][dz+2]
			<-entry.ready
			if entry.err != nil {
				return nil, nil, entry.err
			}
			bases[featureChunkPosition{cx + dx, cz + dz}] = entry.base
			chunks = append(chunks, entry.base.terrain)
		}
	}
	return bases, chunks, nil
}

func (g *DimensionGenerator) decorateRegion(bases map[featureChunkPosition]*dimensionBase, chunks []*NoiseChunkResult,
	cx, cz, stopBeforeStep int32, starts []StructureStart) (*OverworldFeatureRegion, []string, error) {
	region, err := NewDimensionFeatureRegionWithSeaLevel(g.noise.settings.Height, g.noise.settings.SeaLevel, chunks...)
	if err != nil {
		return nil, nil, err
	}
	region.featureOrder = g.order
	region.borrowed = make(map[featureChunkPosition]bool, len(chunks))
	for position := range region.chunks {
		region.borrowed[position] = true
	}
	biomeAtQuart := func(qx, qy, qz int32) (string, error) {
		base := bases[featureChunkPosition{qx >> 2, qz >> 2}]
		if base == nil {
			return "", fmt.Errorf("biome requires missing neighbor %d,%d", qx>>2, qz>>2)
		}
		minQuart, maxQuart := base.terrain.MinY/4, (base.terrain.MinY+base.terrain.Height)/4-1
		qy = max(minQuart, min(maxQuart, qy))
		return base.biomes[(qy-minQuart)>>2][((qy&3)<<4)|((qz&3)<<2)|(qx&3)], nil
	}
	region.biomeSeed, region.biomeAtQuart = g.seed, biomeAtQuart
	skipped := make(map[string]bool)
	for dx := int32(-1); dx <= 1; dx++ {
		for dz := int32(-1); dz <= 1; dz++ {
			sx, sz := cx+dx, cz+dz
			present := make(map[string]bool)
			for nx := int32(-1); nx <= 1; nx++ {
				for nz := int32(-1); nz <= 1; nz++ {
					for _, biome := range bases[featureChunkPosition{sx + nx, sz + nz}].biomeNames {
						present[biome] = true
					}
				}
			}
			var biomes []string
			for biome := range present {
				biomes = append(biomes, biome)
			}
			schedule, err := DimensionFeatureSchedule(g.dimension, biomes)
			if err != nil {
				return nil, nil, err
			}
			decorationSeed := OverworldDecorationSeed(g.seed, sx, sz)
			random := NewWorldgenRandom(decorationSeed)
			for step := int32(0); step < stopBeforeStep; step++ {
				if err := g.applyStructureStep(region, starts, sx, sz, step); err != nil {
					return nil, nil, err
				}
				for _, feature := range schedule {
					if int32(feature.Step) != step {
						continue
					}
					plan, supported := g.plans[feature.Name]
					if !supported {
						skipped[feature.Name] = true
						continue
					}
					operations, err := plan.placement.bind(region, biomeAtQuart)
					if err != nil {
						return nil, nil, err
					}
					random.SetSeed(decorationSeed + int64(feature.Index) + int64(10000*feature.Step))
					if err := walkPlacement(operations, random, FeaturePosition{X: sx * 16, Z: sz * 16},
						func(position FeaturePosition, random *WorldgenRandom) error {
							_, err := plan.place(region, random, position)
							return err
						}); err != nil {
						return nil, nil, fmt.Errorf("%s feature %s in %d,%d: %w", g.dimension, feature.Name, sx, sz, err)
					}
				}
			}
		}
	}
	var skippedFeatures []string
	for name := range skipped {
		skippedFeatures = append(skippedFeatures, name)
	}
	sort.Strings(skippedFeatures)
	return region, skippedFeatures, nil
}

func (g *DimensionGenerator) Generate(cx, cz int32) (*GeneratedDimensionChunk, error) {
	bases, chunks, err := g.decorationBases(cx, cz)
	if err != nil {
		return nil, err
	}
	starts := g.decorationStructures(bases, cx, cz)
	region, skipped, err := g.decorateRegion(bases, chunks, cx, cz, 11, starts)
	if err != nil {
		return nil, err
	}
	if err := finalizeGeneratedLeafDistances(region, cx, cz); err != nil {
		return nil, err
	}
	base := bases[featureChunkPosition{cx, cz}]
	result := &GeneratedDimensionChunk{Terrain: region.cloneChunk(cx, cz), Biomes: append([][64]string(nil), base.biomes...),
		BlockEntities: region.BlockEntitiesInChunk(cx, cz), PendingEntities: region.PendingEntitiesInChunk(cx, cz), SkippedFeatures: skipped}
	if _, err := PostProcessGeneratedBubbleColumns(result.Terrain); err != nil {
		return nil, err
	}
	if g.structures != nil {
		result.Structures, err = g.generatedStructureNBT(cx, cz, base.starts)
		if err != nil {
			return nil, err
		}
	} else {
		result.SkippedStages = append(result.SkippedStages, "structures")
	}
	result.SkippedStages = append(result.SkippedStages, "neighbor_decoration_order", "generation_lighting_and_ticks", "initial_mob_spawning")
	if len(result.PendingEntities) > 0 {
		result.SkippedStages = append(result.SkippedStages, "feature_entities")
	}
	return result, nil
}
