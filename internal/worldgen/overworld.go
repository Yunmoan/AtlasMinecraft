package worldgen

import (
	"fmt"
	"sort"
	"sync"
)

const overworldCacheLimit = 128

// OverworldGenerationVersion distinguishes newly generated chunks from earlier
// Atlas coverage without rewriting or regenerating any existing full chunks.
const OverworldGenerationVersion int32 = 5

// GeneratedOverworldChunk contains terrain, quart biomes and the decoration
// coverage for one chunk. Unsupported features are explicit, never presented
// as an exact full vanilla generator.
type GeneratedOverworldChunk struct {
	Terrain         *NoiseChunkResult
	Biomes          [24][64]string
	SkippedFeatures []string
}

type overworldBase struct {
	terrain    *NoiseChunkResult
	biomes     [24][64]string
	biomeNames []string
}

type overworldCacheEntry struct {
	ready    chan struct{}
	base     *overworldBase
	err      error
	used     uint64
	complete bool
}

type decorationFeature struct {
	feature   PlacedFeature
	placement *placementTemplate
	place     func(*OverworldFeatureRegion, *WorldgenRandom, FeaturePosition) (bool, error)
}

// OverworldGenerator is safe for concurrent generation. Cached terrain is
// immutable; every decoration replay uses private block overlays. Only the
// requested final chunk is saved, so generating its neighbors never rewrites
// blocks already edited by players.
type OverworldGenerator struct {
	seed  int64
	noise *OverworldNoiseGenerator
	plans map[string]decorationFeature
	mu    sync.Mutex
	clock uint64
	cache map[featureChunkPosition]*overworldCacheEntry
}

func NewOverworldGenerator(seed int64) (*OverworldGenerator, error) {
	noise, err := NewOverworldNoiseGenerator(seed)
	if err != nil {
		return nil, err
	}
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		return nil, err
	}
	generator := &OverworldGenerator{seed: seed, noise: noise, plans: make(map[string]decorationFeature), cache: make(map[featureChunkPosition]*overworldCacheEntry)}
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
		switch configured.Type {
		case "minecraft:tree", "minecraft:random_selector", "minecraft:simple_block", "minecraft:random_patch", "minecraft:flower", "minecraft:no_bonemeal_flower", "minecraft:block_column", "minecraft:huge_brown_mushroom", "minecraft:huge_red_mushroom", "minecraft:random_boolean_selector", "minecraft:spring_feature", "minecraft:freeze_top_layer", "minecraft:ice_spike", "minecraft:simple_random_selector", "minecraft:underwater_magma", "minecraft:dripstone_cluster", "minecraft:large_dripstone", "minecraft:multiface_growth", "minecraft:vegetation_patch", "minecraft:waterlogged_vegetation_patch", "minecraft:root_system", "minecraft:vines", "minecraft:lake", "minecraft:geode":
			place, err := compileFeaturePlan(placed.Feature, 0)
			if err != nil {
				continue
			}
			plan.place = place
		case "minecraft:ore":
			config, err := LoadOreConfiguration(placed.Feature)
			if err != nil {
				return nil, err
			}
			plan.place = func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
				return PlaceOreFeature(region, random, origin, config)
			}
		case "minecraft:forest_rock":
			state, err := LoadForestRockState(placed.Feature)
			if err != nil {
				return nil, err
			}
			plan.place = func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
				return PlaceForestRock(region, random, origin, state)
			}
		case "minecraft:blue_ice":
			plan.place = func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
				return PlaceBlueIce(region, random, origin)
			}
		case "minecraft:disk":
			config, err := LoadDiskConfiguration(placed.Feature)
			if err != nil {
				return nil, err
			}
			plan.place = func(region *OverworldFeatureRegion, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
				return PlaceDiskFeature(region, random, origin, config)
			}
		default:
			continue
		}
		plan.placement, err = compilePlacementTemplate(placed, &feature)
		if err != nil {
			return nil, err
		}
		generator.plans[feature.Name] = plan
	}
	return generator, nil
}

func (generator *OverworldGenerator) generateBase(cx, cz int32) (*overworldBase, error) {
	terrain, err := generator.noise.GenerateCarved(cx, cz)
	if err != nil {
		return nil, err
	}
	// Carving masks are probe/debug data; they are not needed for decoration.
	terrain.CarvingMask = nil
	biomes, err := newOverworldBiomeSampler(generator.seed, generator.noise.settings)
	if err != nil {
		return nil, err
	}
	base := &overworldBase{terrain: terrain}
	present := make(map[string]bool)
	for si := range base.biomes {
		for i := range base.biomes[si] {
			biome, err := biomes.BiomeAtQuart(cx*4+int32(i&3), -16+int32(si*4+i/16), cz*4+int32((i>>2)&3))
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

// Reuse an in-flight base without waiting here. A worker first generates other
// missing members of its neighborhood, then joins the shared results. Waiting
// at each coordinate made adjacent workers idle behind the same terrain job.
func (generator *OverworldGenerator) baseEntryAt(cx, cz int32) *overworldCacheEntry {
	key := featureChunkPosition{cx, cz}
	generator.mu.Lock()
	generator.clock++
	if entry := generator.cache[key]; entry != nil {
		entry.used = generator.clock
		generator.mu.Unlock()
		return entry
	}
	entry := &overworldCacheEntry{ready: make(chan struct{}), used: generator.clock}
	generator.cache[key] = entry
	generator.mu.Unlock()
	base, err := generator.generateBase(cx, cz)
	generator.mu.Lock()
	entry.base, entry.err, entry.complete = base, err, true
	close(entry.ready)
	if err != nil {
		delete(generator.cache, key)
	}
	for len(generator.cache) > overworldCacheLimit {
		var oldest featureChunkPosition
		used := ^uint64(0)
		found := false
		for position, candidate := range generator.cache {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(generator.cache, oldest)
	}
	generator.mu.Unlock()
	return entry
}

func cloneTerrain(source *NoiseChunkResult) *NoiseChunkResult {
	chunk := *source
	// Keep the returned chunk fully independent of the cached template.
	chunk.Blocks = append([]NoiseMaterial(nil), source.Blocks...)
	chunk.FluidUpdates = append([]bool(nil), source.FluidUpdates...)
	chunk.PostProcessing = append([]bool(nil), source.PostProcessing...)
	chunk.SurfaceBlocks = make(map[int]SurfaceBlockState, len(source.SurfaceBlocks))
	for index, state := range source.SurfaceBlocks {
		if state.Properties != nil {
			properties := make(map[string]string, len(state.Properties))
			for key, value := range state.Properties {
				properties[key] = value
			}
			state.Properties = properties
		}
		chunk.SurfaceBlocks[index] = state
	}
	return &chunk
}

func (generator *OverworldGenerator) Generate(cx, cz int32) (*GeneratedOverworldChunk, error) {
	if cx < -1874998 || cx >= 1874998 || cz < -1874998 || cz >= 1874998 {
		return nil, fmt.Errorf("chunk %d,%d has a generation neighborhood outside the world bounds", cx, cz)
	}
	// Neighboring source chunks can place features into the requested chunk.
	// Their own 3x3 neighborhoods require a 5x5 terrain-only region. Replay in
	// a fixed order so worker scheduling and exploration cannot affect saves.
	bases := make(map[featureChunkPosition]*overworldBase, 25)
	chunks := make([]*NoiseChunkResult, 0, 25)
	var entries [5][5]*overworldCacheEntry
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			entries[dx+2][dz+2] = generator.baseEntryAt(cx+dx, cz+dz)
		}
	}
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			entry := entries[dx+2][dz+2]
			<-entry.ready
			if entry.err != nil {
				return nil, entry.err
			}
			base := entry.base
			bases[featureChunkPosition{cx + dx, cz + dz}] = base
			chunks = append(chunks, base.terrain)
		}
	}
	region, err := NewOverworldFeatureRegion(chunks...)
	if err != nil {
		return nil, err
	}
	// Feature writes use sparse private overlays. The returned target alone is
	// materialized and fully detached from all immutable terrain templates.
	region.borrowed = make(map[featureChunkPosition]bool, len(chunks))
	for position := range region.chunks {
		region.borrowed[position] = true
	}
	biomeAtQuart := func(qx, qy, qz int32) (string, error) {
		base := bases[featureChunkPosition{qx >> 2, qz >> 2}]
		if base == nil {
			return "", fmt.Errorf("biome requires missing neighbor %d,%d", qx>>2, qz>>2)
		}
		qy = max(int32(-16), min(int32(79), qy))
		return base.biomes[(qy+16)>>2][((qy&3)<<4)|((qz&3)<<2)|(qx&3)], nil
	}
	region.biomeSeed, region.biomeAtQuart = generator.seed, biomeAtQuart
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
			schedule, err := OverworldFeatureSchedule(biomes)
			if err != nil {
				return nil, err
			}
			decorationSeed := OverworldDecorationSeed(generator.seed, sx, sz)
			random := NewWorldgenRandom(decorationSeed)
			for _, feature := range schedule {
				plan, supported := generator.plans[feature.Name]
				if !supported {
					skipped[feature.Name] = true
					continue
				}
				operations, err := plan.placement.bind(region, biomeAtQuart)
				if err != nil {
					return nil, err
				}
				random.SetSeed(decorationSeed + int64(feature.Index) + int64(10000*feature.Step))
				if err := walkPlacement(operations, random, FeaturePosition{X: sx * 16, Z: sz * 16},
					func(position FeaturePosition, random *WorldgenRandom) error {
						_, err := plan.place(region, random, position)
						return err
					}); err != nil {
					return nil, fmt.Errorf("feature %s in %d,%d: %w", feature.Name, sx, sz, err)
				}
			}
		}
	}
	result := &GeneratedOverworldChunk{Terrain: region.cloneChunk(cx, cz), Biomes: bases[featureChunkPosition{cx, cz}].biomes}
	for name := range skipped {
		result.SkippedFeatures = append(result.SkippedFeatures, name)
	}
	sort.Strings(result.SkippedFeatures)
	return result, nil
}
