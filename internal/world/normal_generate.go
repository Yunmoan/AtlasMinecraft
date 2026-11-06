package world

import (
	"atlas/internal/nbt"
	"atlas/internal/registry"
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"fmt"
)

func (v *VanillaSave) generateNormal(cx, cz int32) (*Chunk, error) {
	generated, err := v.normal.Generate(cx, cz)
	if err != nil {
		return nil, err
	}
	chunk, err := chunkFromGeneratedOverworld(generated)
	if err != nil {
		return nil, err
	}
	root, _ := chunk.anvilRoot.Object()
	skipped := make([]nbt.Tag, 0, len(generated.SkippedFeatures))
	for _, name := range generated.SkippedFeatures {
		skipped = append(skipped, nbt.Tag{Type: nbt.String, Value: name})
	}
	stages := make([]nbt.Tag, 0, len(generated.SkippedStages))
	for _, name := range generated.SkippedStages {
		stages = append(stages, nbt.Tag{Type: nbt.String, Value: name})
	}
	root["AtlasWorldgen"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"Version":         {Type: nbt.Int, Value: worldgen.OverworldGenerationVersion},
		"Complete":        {Type: nbt.Byte, Value: int8(0)},
		"SkippedFeatures": {Type: nbt.List, ListType: nbt.String, Value: skipped},
		"SkippedStages":   {Type: nbt.List, ListType: nbt.String, Value: stages},
	}}
	return chunk, nil
}

// chunkFromGeneratedOverworld bridges the full generated chunk — terrain,
// biomes and feature block entities — into the protocol and Anvil palettes.
func chunkFromGeneratedOverworld(generated *worldgen.GeneratedOverworldChunk) (*Chunk, error) {
	chunk, err := chunkFromNoise(generated.Terrain, generated.Biomes)
	if err != nil {
		return nil, err
	}
	if len(generated.BlockEntities) != 0 {
		root, _ := chunk.anvilRoot.Object()
		entities, _ := root["block_entities"].Elements()
		for _, entity := range generated.BlockEntities {
			entities = append(entities, fullGeneratedBlockEntity(entity))
		}
		root["block_entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: entities}
	}
	installGeneratedStructures(chunk, generated)
	return chunk, nil
}

// LevelChunk.getBlockEntityNbtForSaving marks instantiated block entities as
// unpacked. Copy the top-level object so cached generation output stays private.
func fullGeneratedBlockEntity(entity nbt.Object) nbt.Tag {
	data := make(nbt.Object, len(entity)+1)
	for key, value := range entity {
		data[key] = value
	}
	data["keepPacked"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	return nbt.Tag{Type: nbt.Compound, Value: data}
}

// chunkFromNoise bridges the reference-checked generation stages to the
// protocol and Anvil palettes, retaining block properties and 3D quart biomes.
func chunkFromNoise(terrain *worldgen.NoiseChunkResult, biomes [NumSections][64]string) (*Chunk, error) {
	return chunkFromNoiseBounds(terrain, biomes[:], true)
}

func chunkFromNoiseBounds(terrain *worldgen.NoiseChunkResult, biomes [][64]string, hasSky bool) (*Chunk, error) {
	if terrain == nil || terrain.Height <= 0 || terrain.Height%16 != 0 || len(biomes) != int(terrain.Height/16) {
		return nil, fmt.Errorf("terrain and biome bounds do not match")
	}
	chunk := newChunkBounds(terrain.X, terrain.Z, terrain.MinY, terrain.MinY+terrain.Height, hasSky)
	ids := make(map[string]uint32)
	for y := terrain.MinY; y < terrain.MinY+terrain.Height; y++ {
		for z := int32(0); z < 16; z++ {
			for x := int32(0); x < 16; x++ {
				state, err := terrain.BlockStateAt(x, y, z)
				if err != nil {
					return nil, err
				}
				// Most terrain states have no properties; cache that common path.
				id, found := ids[state.Name]
				if !found || len(state.Properties) != 0 {
					id, err = BlockStateID(state.Name, state.Properties)
					if err != nil {
						return nil, err
					}
					if len(state.Properties) == 0 {
						ids[state.Name] = id
					}
				}
				chunk.SetBlock(int(x), int(z), y, id)
			}
		}
	}
	for si := range biomes {
		for i, name := range biomes[si] {
			id, err := registry.BiomeID(name)
			if err != nil {
				return nil, err
			}
			chunk.biomes[si][i] = id
		}
	}
	if err := chunk.initializeAnvilRoot(); err != nil {
		return nil, err
	}
	postProcessing := make([]nbt.Tag, len(chunk.sect))
	for si := range postProcessing {
		postProcessing[si] = nbt.Tag{Type: nbt.List, ListType: nbt.Short, Value: []nbt.Tag{}}
	}
	for _, index := range terrain.PostProcessingOrder {
		if index < 0 || index >= len(terrain.Blocks) {
			return nil, fmt.Errorf("postprocessing index %d outside generated chunk", index)
		}
		si, i := index/SectionVolume, index%SectionVolume
		// Vanilla packs X | Y<<4 | Z<<8, retaining duplicates and positions
		// whose block was subsequently overwritten during generation.
		packed := int16((i & 15) | ((i>>8)&15)<<4 | ((i>>4)&15)<<8)
		postProcessing[si].Value = append(postProcessing[si].Value.([]nbt.Tag), nbt.Tag{Type: nbt.Short, Value: packed})
	}
	root, _ := chunk.anvilRoot.Object()
	root["PostProcessing"] = nbt.Tag{Type: nbt.List, ListType: nbt.List, Value: postProcessing}
	fluidTicks := make([]tickqueue.Saved[string], 0, len(terrain.FluidTicks))
	for _, tick := range terrain.FluidTicks {
		fluidTicks = append(fluidTicks, tickqueue.Saved[string]{
			Type: tick.Fluid, Position: tickqueue.Position{X: tick.Position.X, Y: tick.Position.Y, Z: tick.Position.Z},
			Priority: tickqueue.ClampPriority(tick.Priority),
		})
	}
	root["fluid_ticks"] = tickqueue.EncodeSavedNBT(fluidTicks, func(kind string) string { return kind })
	blockTicks := make([]tickqueue.Saved[string], 0, len(terrain.BlockTicks))
	for _, tick := range terrain.BlockTicks {
		blockTicks = append(blockTicks, tickqueue.Saved[string]{Type: tick.Block, Position: tickqueue.Position{X: tick.Position.X, Y: tick.Position.Y, Z: tick.Position.Z}, Priority: tickqueue.ClampPriority(tick.Priority)})
	}
	root["block_ticks"] = tickqueue.EncodeSavedNBT(blockTicks, func(name string) string { return name })
	return chunk, nil
}

func (v *VanillaSave) generateDimension(cx, cz int32) (*Chunk, error) {
	generated, err := v.dimensionGenerator.Generate(cx, cz)
	if err != nil {
		return nil, err
	}
	chunk, err := chunkFromNoiseBounds(generated.Terrain, generated.Biomes, v.dimension.hasSky)
	if err != nil {
		return nil, err
	}
	root, _ := chunk.anvilRoot.Object()
	entities := make([]nbt.Tag, 0, len(generated.BlockEntities))
	for _, entity := range generated.BlockEntities {
		entities = append(entities, fullGeneratedBlockEntity(entity))
	}
	root["block_entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: entities}
	installGeneratedStructures(chunk, &worldgen.GeneratedOverworldChunk{Structures: generated.Structures, PendingEntities: generated.PendingEntities})
	skipped := make([]nbt.Tag, 0, len(generated.SkippedFeatures))
	for _, name := range generated.SkippedFeatures {
		skipped = append(skipped, nbt.Tag{Type: nbt.String, Value: name})
	}
	stages := make([]nbt.Tag, 0, len(generated.SkippedStages))
	for _, name := range generated.SkippedStages {
		stages = append(stages, nbt.Tag{Type: nbt.String, Value: name})
	}
	root["AtlasWorldgen"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"Version":         {Type: nbt.Int, Value: worldgen.OverworldGenerationVersion},
		"Complete":        {Type: nbt.Byte, Value: int8(0)},
		"Dimension":       {Type: nbt.String, Value: v.dimension.name},
		"SkippedFeatures": {Type: nbt.List, ListType: nbt.String, Value: skipped},
		"SkippedStages":   {Type: nbt.List, ListType: nbt.String, Value: stages},
	}}
	return chunk, nil
}
