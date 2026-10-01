package worldgen

import "fmt"

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

// OverworldFeatureRegion exposes already-generated neighboring chunks to block
// features, including edits that cross chunk boundaries.
type OverworldFeatureRegion struct {
	chunks         map[featureChunkPosition]*NoiseChunkResult
	borrowed       map[featureChunkPosition]bool
	heights        map[featureColumnPosition]int32
	motionHeights  map[featureColumnPosition]int32
	surfaceHeights map[featureColumnPosition]int32
}

func NewOverworldFeatureRegion(chunks ...*NoiseChunkResult) (*OverworldFeatureRegion, error) {
	region := &OverworldFeatureRegion{chunks: make(map[featureChunkPosition]*NoiseChunkResult),
		heights: make(map[featureColumnPosition]int32), motionHeights: make(map[featureColumnPosition]int32), surfaceHeights: make(map[featureColumnPosition]int32)}
	for _, chunk := range chunks {
		if chunk == nil || chunk.MinY != -64 || chunk.Height != 384 || len(chunk.Blocks) != 384*256 {
			return nil, fmt.Errorf("feature region requires complete 1.21.1 overworld chunks")
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

func (region *OverworldFeatureRegion) OceanFloorHeightAt(blockX, blockZ int32) (int32, error) {
	column := featureColumnPosition{blockX, blockZ}
	if height, exists := region.heights[column]; exists {
		return height, nil
	}
	chunk, err := region.chunkAt(blockX, blockZ)
	if err != nil {
		return 0, err
	}
	for y := chunk.MinY + chunk.Height - 1; y >= chunk.MinY; y-- {
		solid, err := oceanFloorSolid(chunk.surfaceState(blockX&15, y, blockZ&15))
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
		for y := chunk.MinY + chunk.Height - 1; y >= chunk.MinY; y-- {
			name := chunk.surfaceState(blockX&15, y, blockZ&15).Name
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
	for y := chunk.MinY + chunk.Height - 1; y >= chunk.MinY; y-- {
		blocksMotion, err := motionBlocking(chunk.surfaceState(blockX&15, y, blockZ&15))
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
	if position.Y < -64 || position.Y >= 320 {
		return SurfaceBlockState{Name: "minecraft:air"}, nil
	}
	chunk, err := region.chunkAt(position.X, position.Z)
	if err != nil {
		return SurfaceBlockState{}, err
	}
	return chunk.surfaceState(position.X&15, position.Y, position.Z&15), nil
}

func (region *OverworldFeatureRegion) SetBlock(position FeaturePosition, state SurfaceBlockState) error {
	if !region.InWorld(position) {
		return fmt.Errorf("feature block outside world at %+v", position)
	}
	chunk, err := region.chunkAt(position.X, position.Z)
	if err != nil {
		return err
	}
	key := featureChunkPosition{position.X >> 4, position.Z >> 4}
	if region.borrowed[key] {
		private := *chunk
		private.PostProcessing = append([]bool(nil), chunk.PostProcessing...)
		private.SurfaceBlocks = make(map[int]SurfaceBlockState, len(chunk.SurfaceBlocks))
		for index, state := range chunk.SurfaceBlocks {
			private.SurfaceBlocks[index] = state
		}
		chunk = &private
		region.chunks[key] = chunk
		delete(region.borrowed, key)
	}
	chunk.setSurfaceState(position.X&15, position.Y, position.Z&15, state)
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
	// SetBlock obtains a private overlay without changing the block or random.
	state, err := region.BlockAt(p)
	if err != nil {
		return err
	}
	if err = region.SetBlock(p, state); err != nil {
		return err
	}
	chunk, err := region.chunkAt(p.X, p.Z)
	if err != nil {
		return err
	}
	if chunk.PostProcessing == nil {
		chunk.PostProcessing = make([]bool, len(chunk.Blocks))
	}
	index, _ := chunk.Index(p.X&15, p.Y, p.Z&15)
	chunk.PostProcessing[index] = true
	return nil
}

func (region *OverworldFeatureRegion) InWorld(position FeaturePosition) bool {
	return position.X > -30000000 && position.X < 30000000 &&
		position.Z > -30000000 && position.Z < 30000000 && position.Y >= -64 && position.Y < 320
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
