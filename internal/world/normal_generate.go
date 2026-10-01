package world

import (
	"atlas/internal/nbt"
	"atlas/internal/registry"
	"atlas/internal/worldgen"
)

func (v *VanillaSave) generateNormal(cx, cz int32) (*Chunk, error) {
	generated, err := v.normal.Generate(cx, cz)
	if err != nil {
		return nil, err
	}
	chunk, err := chunkFromNoise(generated.Terrain, generated.Biomes)
	if err != nil {
		return nil, err
	}
	root, _ := chunk.anvilRoot.Object()
	skipped := make([]nbt.Tag, 0, len(generated.SkippedFeatures))
	for _, name := range generated.SkippedFeatures {
		skipped = append(skipped, nbt.Tag{Type: nbt.String, Value: name})
	}
	root["AtlasWorldgen"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"Version":         {Type: nbt.Int, Value: worldgen.OverworldGenerationVersion},
		"Complete":        {Type: nbt.Byte, Value: int8(0)},
		"SkippedFeatures": {Type: nbt.List, ListType: nbt.String, Value: skipped},
	}}
	return chunk, nil
}

// chunkFromNoise bridges the reference-checked generation stages to the
// protocol and Anvil palettes, retaining block properties and 3D quart biomes.
func chunkFromNoise(terrain *worldgen.NoiseChunkResult, biomes [NumSections][64]string) (*Chunk, error) {
	chunk := NewChunk(terrain.X, terrain.Z)
	ids := make(map[string]uint32)
	for y := int32(MinY); y < MaxY; y++ {
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
	postProcessing := make([]nbt.Tag, NumSections)
	for si := range postProcessing {
		positions := []nbt.Tag{}
		for i := 0; i < SectionVolume; i++ {
			index := si*SectionVolume + i
			featureMarked := index < len(terrain.PostProcessing) && terrain.PostProcessing[index]
			if !featureMarked && (index >= len(terrain.FluidUpdates) || !terrain.FluidUpdates[index]) {
				continue
			}
			state, err := terrain.BlockStateAt(int32(i&15), MinY+int32(si*16+i/256), int32((i>>4)&15))
			if err != nil {
				return nil, err
			}
			if featureMarked || state.Name == "minecraft:water" || state.Name == "minecraft:lava" {
				// Vanilla packs post-processing as X | Y<<4 | Z<<8.
				packed := int16((i & 15) | ((i>>8)&15)<<4 | ((i>>4)&15)<<8)
				positions = append(positions, nbt.Tag{Type: nbt.Short, Value: packed})
			}
		}
		postProcessing[si] = nbt.Tag{Type: nbt.List, ListType: nbt.Short, Value: positions}
	}
	root, _ := chunk.anvilRoot.Object()
	root["PostProcessing"] = nbt.Tag{Type: nbt.List, ListType: nbt.List, Value: postProcessing}
	return chunk, nil
}
