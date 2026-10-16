package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

type FeatureBlockWorld interface {
	BlockAt(position FeaturePosition) (SurfaceBlockState, error)
	SetBlock(position FeaturePosition, state SurfaceBlockState) error
	InWorld(position FeaturePosition) bool
}

type featureChunkPosition struct {
	chunkX, chunkZ int32
}

type featureColumnPosition struct {
	blockX, blockZ int32
}

type featureChunkOverlay struct {
	blocks         map[int]SurfaceBlockState
	postProcessing map[int]bool
}

// OverworldFeatureRegion exposes already-generated neighboring chunks to block
// features, including edits that cross chunk boundaries.
type OverworldFeatureRegion struct {
	chunks             map[featureChunkPosition]*NoiseChunkResult
	borrowed           map[featureChunkPosition]bool
	overlays           map[featureChunkPosition]*featureChunkOverlay
	heights            map[featureColumnPosition]int32
	motionHeights      map[featureColumnPosition]int32
	surfaceHeights     map[featureColumnPosition]int32
	biomeSeed          int64
	biomeAtQuart       func(int32, int32, int32) (string, error)
	biomeBases         map[featureChunkPosition]*overworldBase
	blockEntities      map[FeaturePosition]featureBlockEntity
	minY, height       int32
	generationDepth    int32
	generationSeaLevel int32
	treeShuffle        *treeShuffleContext
	readBrightness     FeatureBrightnessReader
	featureOrder       *featureOrder
	pendingEntities    []PendingStructureEntity
	structureRecording *structureBlockResult
}

func (region *OverworldFeatureRegion) ScheduleFluidUpdate(p FeaturePosition) error {
	return region.MarkPostProcessing(p)
}

func NewOverworldFeatureRegion(chunks ...*NoiseChunkResult) (*OverworldFeatureRegion, error) {
	for _, chunk := range chunks {
		if chunk == nil || chunk.MinY != -64 || chunk.Height != 384 {
			return nil, fmt.Errorf("feature region requires complete 1.21.1 overworld chunks")
		}
	}
	return NewFeatureRegion(chunks...)
}

// NewFeatureRegion binds chunks from one dimension to a common build range.
// Noise generation bounds (Nether 0..128) can differ from the dimension's
// actual build height; callers must supply chunks expanded to that range.
func NewFeatureRegion(chunks ...*NoiseChunkResult) (*OverworldFeatureRegion, error) {
	if len(chunks) == 0 || chunks[0] == nil {
		return nil, fmt.Errorf("feature region requires at least one complete chunk")
	}
	region := &OverworldFeatureRegion{chunks: make(map[featureChunkPosition]*NoiseChunkResult),
		minY: chunks[0].MinY, height: chunks[0].Height, generationDepth: chunks[0].Height, generationSeaLevel: 63,
		heights: make(map[featureColumnPosition]int32), motionHeights: make(map[featureColumnPosition]int32), surfaceHeights: make(map[featureColumnPosition]int32)}
	for _, chunk := range chunks {
		if chunk == nil || chunk.MinY != region.minY || chunk.Height != region.height || chunk.Height <= 0 || chunk.Height > 4096 || chunk.MinY%16 != 0 || chunk.Height%16 != 0 || len(chunk.Blocks) != int(chunk.Height)*256 {
			return nil, fmt.Errorf("feature region requires complete chunks with matching dimension bounds")
		}
		position := featureChunkPosition{chunk.X, chunk.Z}
		if region.chunks[position] != nil {
			return nil, fmt.Errorf("duplicate feature region chunk %d,%d", chunk.X, chunk.Z)
		}
		region.chunks[position] = chunk
	}
	return region, nil
}

func (region *OverworldFeatureRegion) chunkAt(blockX, blockZ int32) (*NoiseChunkResult, error) {
	chunk := region.chunks[featureChunkPosition{blockX >> 4, blockZ >> 4}]
	if chunk == nil {
		return nil, fmt.Errorf("feature requires missing neighboring chunk %d,%d", blockX>>4, blockZ>>4)
	}
	return chunk, nil
}

func (region *OverworldFeatureRegion) stateAt(chunk *NoiseChunkResult, overlay *featureChunkOverlay, x, y, z int32) SurfaceBlockState {
	if overlay != nil {
		if state, ok := overlay.blocks[int((y-chunk.MinY)*256+z*16+x)]; ok {
			return state
		}
	}
	return chunk.surfaceState(x, y, z)
}

func (region *OverworldFeatureRegion) overlayAt(x, z int32) *featureChunkOverlay {
	return region.overlays[featureChunkPosition{x >> 4, z >> 4}]
}

func (region *OverworldFeatureRegion) OceanFloorHeightAt(blockX, blockZ int32) (int32, error) {
	column := featureColumnPosition{blockX, blockZ}
	if height, exists := region.heights[column]; exists {
		return height, nil
	}
	chunk, err := region.chunkAt(blockX, blockZ)
	if err != nil {
		return 0, err
	}
	overlay := region.overlayAt(blockX, blockZ)
	for y := chunk.MinY + chunk.Height - 1; y >= chunk.MinY; y-- {
		solid, err := oceanFloorSolid(region.stateAt(chunk, overlay, blockX&15, y, blockZ&15))
		if err != nil {
			return 0, err
		}
		if solid {
			region.heights[column] = y + 1
			return y + 1, nil
		}
	}
	region.heights[column] = chunk.MinY
	return chunk.MinY, nil
}

func (region *OverworldFeatureRegion) HeightmapAt(heightmap string, blockX, blockZ int32) (int32, error) {
	if heightmap == "OCEAN_FLOOR_WG" || heightmap == "OCEAN_FLOOR" {
		return region.OceanFloorHeightAt(blockX, blockZ)
	}
	if heightmap == "WORLD_SURFACE" || heightmap == "WORLD_SURFACE_WG" {
		column := featureColumnPosition{blockX, blockZ}
		if height, exists := region.surfaceHeights[column]; exists {
			return height, nil
		}
		chunk, err := region.chunkAt(blockX, blockZ)
		if err != nil {
			return 0, err
		}
		overlay := region.overlayAt(blockX, blockZ)
		for y := chunk.MinY + chunk.Height - 1; y >= chunk.MinY; y-- {
			name := region.stateAt(chunk, overlay, blockX&15, y, blockZ&15).Name
			if name != "minecraft:air" && name != "minecraft:cave_air" && name != "minecraft:void_air" {
				region.surfaceHeights[column] = y + 1
				return y + 1, nil
			}
		}
		region.surfaceHeights[column] = chunk.MinY
		return chunk.MinY, nil
	}
	if heightmap != "MOTION_BLOCKING" {
		return 0, fmt.Errorf("unsupported feature heightmap %q", heightmap)
	}
	column := featureColumnPosition{blockX, blockZ}
	if height, exists := region.motionHeights[column]; exists {
		return height, nil
	}
	chunk, err := region.chunkAt(blockX, blockZ)
	if err != nil {
		return 0, err
	}
	overlay := region.overlayAt(blockX, blockZ)
	for y := chunk.MinY + chunk.Height - 1; y >= chunk.MinY; y-- {
		blocksMotion, err := motionBlocking(region.stateAt(chunk, overlay, blockX&15, y, blockZ&15))
		if err != nil {
			return 0, err
		}
		if blocksMotion {
			region.motionHeights[column] = y + 1
			return y + 1, nil
		}
	}
	region.motionHeights[column] = chunk.MinY
	return chunk.MinY, nil
}

func (region *OverworldFeatureRegion) BlockAt(position FeaturePosition) (SurfaceBlockState, error) {
	if position.Y < region.minY || position.Y >= region.minY+region.height {
		return SurfaceBlockState{Name: "minecraft:air"}, nil
	}
	chunk, err := region.chunkAt(position.X, position.Z)
	if err != nil {
		return SurfaceBlockState{}, err
	}
	return region.stateAt(chunk, region.overlayAt(position.X, position.Z), position.X&15, position.Y, position.Z&15), nil
}

func (region *OverworldFeatureRegion) SetBlock(position FeaturePosition, state SurfaceBlockState) error {
	if !region.InWorld(position) {
		return fmt.Errorf("feature block outside world at %+v", position)
	}
	chunk, err := region.chunkAt(position.X, position.Z)
	if err != nil {
		return err
	}
	if entity, ok := region.blockEntities[position]; ok && entity.block != state.Name {
		delete(region.blockEntities, position)
	}
	key := featureChunkPosition{position.X >> 4, position.Z >> 4}
	if region.borrowed[key] {
		// A sparse write overlay avoids copying the entire surface map and
		// postprocessing array of every neighbor for each decoration replay.
		if region.overlays == nil {
			region.overlays = make(map[featureChunkPosition]*featureChunkOverlay)
		}
		overlay := region.overlays[key]
		if overlay == nil {
			overlay = &featureChunkOverlay{blocks: make(map[int]SurfaceBlockState)}
			region.overlays[key] = overlay
		}
		index, _ := chunk.Index(position.X&15, position.Y, position.Z&15)
		overlay.blocks[index] = state
	} else {
		chunk.setSurfaceState(position.X&15, position.Y, position.Z&15, state)
	}
	if recording := region.structureRecording; recording != nil {
		recording.blocks[position] = cloneSurfaceState(state)
		if entity, ok := recording.blockEntities[position]; ok && entity.block != state.Name {
			delete(recording.blockEntities, position)
		}
	}
	// Underground ore writes cannot change the top of a column. Invalidating
	// every heightmap here made each vein scan from Y=319 all over again.
	column := featureColumnPosition{position.X, position.Z}
	for _, heights := range []map[featureColumnPosition]int32{region.heights, region.motionHeights, region.surfaceHeights} {
		if height, cached := heights[column]; cached && position.Y >= height-1 {
			delete(heights, column)
		}
	}
	return nil
}

func (region *OverworldFeatureRegion) MarkPostProcessing(p FeaturePosition) error {
	if !region.InWorld(p) {
		return nil
	}
	chunk, err := region.chunkAt(p.X, p.Z)
	if err != nil {
		return err
	}
	index, _ := chunk.Index(p.X&15, p.Y, p.Z&15)
	if recording := region.structureRecording; recording != nil {
		recording.postProcessing[p] = true
	}
	key := featureChunkPosition{p.X >> 4, p.Z >> 4}
	if region.borrowed[key] {
		if region.overlays == nil {
			region.overlays = make(map[featureChunkPosition]*featureChunkOverlay)
		}
		overlay := region.overlays[key]
		if overlay == nil {
			overlay = &featureChunkOverlay{blocks: make(map[int]SurfaceBlockState)}
			region.overlays[key] = overlay
		}
		if overlay.postProcessing == nil {
			overlay.postProcessing = make(map[int]bool)
		}
		overlay.postProcessing[index] = true
		return nil
	}
	if chunk.PostProcessing == nil {
		chunk.PostProcessing = make([]bool, len(chunk.Blocks))
	}
	chunk.PostProcessing[index] = true
	return nil
}

func (region *OverworldFeatureRegion) cloneChunk(cx, cz int32) *NoiseChunkResult {
	chunk := cloneTerrain(region.chunks[featureChunkPosition{cx, cz}])
	if overlay := region.overlayAt(cx*16, cz*16); overlay != nil {
		for index, state := range overlay.blocks {
			chunk.SurfaceBlocks[index] = cloneSurfaceState(state)
		}
		if len(overlay.postProcessing) > 0 && chunk.PostProcessing == nil {
			chunk.PostProcessing = make([]bool, len(chunk.Blocks))
		}
		for index := range overlay.postProcessing {
			chunk.PostProcessing[index] = true
		}
	}
	return chunk
}

func (region *OverworldFeatureRegion) InWorld(position FeaturePosition) bool {
	return position.X > -30000000 && position.X < 30000000 &&
		position.Z > -30000000 && position.Z < 30000000 && position.Y >= region.minY && position.Y < region.minY+region.height
}

func (region *OverworldFeatureRegion) MinBuildHeight() int32     { return region.minY }
func (region *OverworldFeatureRegion) MaxBuildHeight() int32     { return region.minY + region.height }
func (region *OverworldFeatureRegion) GenerationDepth() int32    { return region.generationDepth }
func (region *OverworldFeatureRegion) GenerationSeaLevel() int32 { return region.generationSeaLevel }
func (region *OverworldFeatureRegion) WorldSeed() int64          { return region.biomeSeed }

// The generation stage records entity requests until a runtime entity consumer
// has saved and spawned them. Recording a request must not set piece flags.
func (region *OverworldFeatureRegion) AddPendingEntity(entity PendingStructureEntity) error {
	if !region.InWorld(entity.Position) {
		return fmt.Errorf("feature entity outside world at %+v", entity.Position)
	}
	if entity.Data != nil {
		entity.Data = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.Data}).Value.(nbt.Object)
	}
	region.pendingEntities = append(region.pendingEntities, entity)
	if recording := region.structureRecording; recording != nil {
		recording.pending = append(recording.pending, entity)
	}
	return nil
}

func (region *OverworldFeatureRegion) PendingEntitiesInChunk(cx, cz int32) []PendingStructureEntity {
	var result []PendingStructureEntity
	for _, entity := range region.pendingEntities {
		position := pendingEntityBlockPosition(entity)
		if position.X>>4 == cx && position.Z>>4 == cz {
			if entity.Data != nil {
				entity.Data = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.Data}).Value.(nbt.Object)
			}
			result = append(result, entity)
		}
	}
	return result
}

func NewDimensionFeatureRegion(generationDepth int32, chunks ...*NoiseChunkResult) (*OverworldFeatureRegion, error) {
	return NewDimensionFeatureRegionWithSeaLevel(generationDepth, 63, chunks...)
}

// Sea level belongs to the noise generator, independently of build height.
// Vanilla 1.21.1 uses 32 in the Nether and 0 in the End.
func NewDimensionFeatureRegionWithSeaLevel(generationDepth, seaLevel int32, chunks ...*NoiseChunkResult) (*OverworldFeatureRegion, error) {
	region, err := NewFeatureRegion(chunks...)
	if err != nil {
		return nil, err
	}
	if generationDepth <= 0 || generationDepth > region.height || generationDepth%16 != 0 {
		return nil, fmt.Errorf("invalid dimension generation depth %d", generationDepth)
	}
	region.generationDepth = generationDepth
	if seaLevel < region.minY || seaLevel >= region.minY+region.height {
		return nil, fmt.Errorf("invalid dimension sea level %d", seaLevel)
	}
	region.generationSeaLevel = seaLevel
	return region, nil
}

func (region *OverworldFeatureRegion) requireNeighborhood(chunkX, chunkZ int32) error {
	if region == nil {
		return fmt.Errorf("feature requires a chunk region")
	}
	for offsetX := int32(-1); offsetX <= 1; offsetX++ {
		for offsetZ := int32(-1); offsetZ <= 1; offsetZ++ {
			if region.chunks[featureChunkPosition{chunkX + offsetX, chunkZ + offsetZ}] == nil {
				return fmt.Errorf("feature requires the complete 3x3 neighborhood around chunk %d,%d", chunkX, chunkZ)
			}
		}
	}
	return nil
}

// ApplyOverworldOreFeature runs one registered placed ore feature against a
// region. It does not execute preceding structures or decoration stages.
func ApplyOverworldOreFeature(region *OverworldFeatureRegion, worldSeed int64, chunkX, chunkZ int32,
	feature PlacedFeature, biomeAtQuart func(int32, int32, int32) (string, error)) (int, error) {
	if err := region.requireNeighborhood(chunkX, chunkZ); err != nil {
		return 0, err
	}
	placed, err := LoadPlacedFeatureDefinition(feature.Name)
	if err != nil {
		return 0, err
	}
	config, err := LoadOreConfiguration(placed.Feature)
	if err != nil {
		return 0, err
	}
	successes := 0
	err = visitFeaturePositions(region, worldSeed, chunkX, chunkZ, feature, biomeAtQuart, func(position FeaturePosition, random *WorldgenRandom) error {
		changed, err := PlaceOreFeature(region, random, position, config)
		if err != nil {
			return err
		}
		if changed {
			successes++
		}
		return nil
	})
	return successes, err
}
