package world

import (
	"atlas/internal/nbt"
	"atlas/internal/registry"
	"errors"
	"fmt"
	"math/bits"
)

func chunkFromAnvil(name string, root nbt.Tag, cx, cz int32) (*Chunk, error) {
	return chunkFromAnvilBounds(name, root, cx, cz, MinY, MaxY, true)
}

func chunkFromAnvilBounds(name string, root nbt.Tag, cx, cz, minY, maxY int32, hasSky bool) (*Chunk, error) {
	o, ok := root.Object()
	if !ok {
		return nil, errors.New("chunk root is not compound")
	}
	x, ok := o.Number("xPos")
	if !ok || x != int64(cx) {
		return nil, fmt.Errorf("chunk xPos mismatch: %d", x)
	}
	z, ok := o.Number("zPos")
	if !ok || z != int64(cz) {
		return nil, fmt.Errorf("chunk zPos mismatch: %d", z)
	}
	version, ok := o.Number("DataVersion")
	if !ok || version != 3955 {
		return nil, fmt.Errorf("unsupported chunk DataVersion %d (requires 1.21.1/3955)", version)
	}
	sections, ok := o["sections"].Elements()
	if !ok {
		return nil, errors.New("chunk sections missing")
	}
	c := newChunkBounds(cx, cz, minY, maxY, hasSky)
	for _, section := range sections {
		so, ok := section.Object()
		if !ok {
			return nil, errors.New("section is not compound")
		}
		y, ok := so.Number("Y")
		if !ok {
			return nil, errors.New("section Y missing")
		}
		si := int(y) - int(c.minY>>4)
		if si < 0 || si >= len(c.sect) {
			continue
		}
		if blockStates, ok := so.Object("block_states"); ok {
			paletteTag, ok := blockStates["palette"].Elements()
			if !ok || len(paletteTag) == 0 {
				return nil, fmt.Errorf("section %d block palette missing", y)
			}
			palette := make([]uint32, 0, len(paletteTag))
			for _, entry := range paletteTag {
				item, ok := entry.Object()
				if !ok {
					return nil, errors.New("block palette entry not compound")
				}
				name, ok := item.Text("Name")
				if !ok {
					return nil, errors.New("block palette name missing")
				}
				props := make(map[string]string)
				if po, ok := item.Object("Properties"); ok {
					for key, val := range po {
						v, ok := val.Text()
						if !ok {
							return nil, errors.New("block property is not string")
						}
						props[key] = v
					}
				}
				id, err := BlockStateID(name, props)
				if err != nil {
					return nil, err
				}
				palette = append(palette, id)
			}
			var data []int64
			if len(palette) > 1 {
				var ok bool
				data, ok = blockStates["data"].Longs()
				if !ok {
					return nil, fmt.Errorf("section %d block data missing", y)
				}
			}
			if err := decodePalette(4096, palette, data, 4, func(i int, id uint32) { c.SetBlock(i&15, (i>>4)&15, int32(y*16)+int32(i>>8), id) }); err != nil {
				return nil, fmt.Errorf("section %d blocks: %w", y, err)
			}
		}
		if biomes, ok := so.Object("biomes"); ok {
			entries, ok := biomes["palette"].Elements()
			if !ok || len(entries) == 0 {
				return nil, fmt.Errorf("section %d biome palette missing", y)
			}
			palette := make([]uint32, 0, len(entries))
			for _, entry := range entries {
				name, ok := entry.Text()
				if !ok {
					return nil, errors.New("biome palette entry is not string")
				}
				id, err := registry.BiomeID(name)
				if err != nil {
					return nil, err
				}
				palette = append(palette, uint32(id))
			}
			var data []int64
			if len(palette) > 1 {
				var ok bool
				data, ok = biomes["data"].Longs()
				if !ok {
					return nil, fmt.Errorf("section %d biome data missing", y)
				}
			}
			if err := decodePalette(64, palette, data, 1, func(i int, id uint32) { c.biomes[si][i] = int32(id) }); err != nil {
				return nil, fmt.Errorf("section %d biomes: %w", y, err)
			}
		}
	}
	c.anvilName, c.anvilRoot, c.fromAnvil, c.dirty = name, root, true, false
	lighting, err := readAnvilLight(o, sections, c)
	if err != nil {
		return nil, err
	}
	c.lights = lighting
	clear(c.dirtySections)
	return c, nil
}

func decodePalette(count int, palette []uint32, data []int64, minBits int, set func(int, uint32)) error {
	if len(palette) == 1 {
		for i := 0; i < count; i++ {
			set(i, palette[0])
		}
		return nil
	}
	bitsPer := bits.Len(uint(len(palette) - 1))
	if bitsPer < minBits {
		bitsPer = minBits
	}
	perLong := 64 / bitsPer
	needed := (count + perLong - 1) / perLong
	if len(data) != needed {
		return fmt.Errorf("packed array has %d longs, want %d", len(data), needed)
	}
	mask := uint64((1 << bitsPer) - 1)
	for i := 0; i < count; i++ {
		index := int((uint64(data[i/perLong]) >> (uint(i%perLong) * uint(bitsPer))) & mask)
		if index >= len(palette) {
			return fmt.Errorf("palette index %d out of range", index)
		}
		set(i, palette[index])
	}
	return nil
}

// updateAnvilBlocks changes only block_states in each section. Other chunk
// tags (block entities, scheduled ticks, structures, heightmaps, lighting,
// custom data) remain in the NBT tree for later full-fidelity handling.
func (c *Chunk) updateAnvilBlocks() (string, nbt.Tag, error) {
	if !c.fromAnvil {
		return "", nbt.Tag{}, errors.New("chunk has no Anvil source")
	}
	root, _ := c.anvilRoot.Object()
	sections, _ := root["sections"].Elements()
	seen := make([]bool, len(c.sect))
	for _, section := range sections {
		so, _ := section.Object()
		y, _ := so.Number("Y")
		si := int(y) - int(c.minY>>4)
		if si < 0 || si >= len(c.sect) {
			continue
		}
		seen[si] = true
		if !c.dirtySections[si] {
			continue
		}
		states := make([]uint32, 4096)
		for i := range states {
			states[i] = c.GetBlock(i&15, (i>>4)&15, int32(y*16)+int32(i>>8))
		}
		container, err := encodeAnvilBlockStates(states)
		if err != nil {
			return "", nbt.Tag{}, err
		}
		so["block_states"] = nbt.Tag{Type: nbt.Compound, Value: container}
	}
	for si, dirty := range c.dirtySections {
		if !dirty || seen[si] {
			continue
		}
		states := make([]uint32, 4096)
		y := c.minY + int32(si)*16
		for i := range states {
			states[i] = c.GetBlock(i&15, (i>>4)&15, y+int32(i>>8))
		}
		container, err := encodeAnvilBlockStates(states)
		if err != nil {
			return "", nbt.Tag{}, err
		}
		biomes, err := encodeAnvilBiomes(c.biomes[si])
		if err != nil {
			return "", nbt.Tag{}, err
		}
		so := nbt.Object{
			"Y":            {Type: nbt.Byte, Value: int8(y / 16)},
			"block_states": {Type: nbt.Compound, Value: container},
			"biomes":       {Type: nbt.Compound, Value: biomes},
		}
		sections = append(sections, nbt.Tag{Type: nbt.Compound, Value: so})
	}
	root["sections"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: sections}
	motion, surface, ocean := c.anvilHeightmaps()
	heightmaps, ok := root.Object("Heightmaps")
	if !ok {
		heightmaps = nbt.Object{}
		root["Heightmaps"] = nbt.Tag{Type: nbt.Compound, Value: heightmaps}
	}
	heightmaps["MOTION_BLOCKING"] = nbt.Tag{Type: nbt.LongArray, Value: motion}
	heightmaps["WORLD_SURFACE"] = nbt.Tag{Type: nbt.LongArray, Value: surface}
	heightmaps["OCEAN_FLOOR"] = nbt.Tag{Type: nbt.LongArray, Value: ocean}
	c.writeAnvilLight(root)
	return c.anvilName, c.anvilRoot, nil
}

func encodeAnvilBlockStates(states []uint32) (nbt.Object, error) {
	palette := make([]uint32, 0, 16)
	indices := make(map[uint32]uint64)
	for _, state := range states {
		if _, ok := indices[state]; !ok {
			indices[state] = uint64(len(palette))
			palette = append(palette, state)
		}
	}
	entries := make([]nbt.Tag, 0, len(palette))
	for _, id := range palette {
		state, err := BlockStateByID(id)
		if err != nil {
			return nil, err
		}
		o := nbt.Object{"Name": {Type: nbt.String, Value: state.Name}}
		if len(state.Properties) > 0 {
			props := make(nbt.Object, len(state.Properties))
			for k, v := range state.Properties {
				props[k] = nbt.Tag{Type: nbt.String, Value: v}
			}
			o["Properties"] = nbt.Tag{Type: nbt.Compound, Value: props}
		}
		entries = append(entries, nbt.Tag{Type: nbt.Compound, Value: o})
	}
	o := nbt.Object{"palette": {Type: nbt.List, ListType: nbt.Compound, Value: entries}}
	if len(palette) > 1 {
		bitsPer := bits.Len(uint(len(palette) - 1))
		if bitsPer < 4 {
			bitsPer = 4
		}
		perLong := 64 / bitsPer
		longs := make([]int64, (len(states)+perLong-1)/perLong)
		for i, state := range states {
			longs[i/perLong] |= int64(indices[state] << (uint(i%perLong) * uint(bitsPer)))
		}
		o["data"] = nbt.Tag{Type: nbt.LongArray, Value: longs}
	}
	return o, nil
}

func encodeAnvilBiomes(biomes [64]int32) (nbt.Object, error) {
	palette := make([]int32, 0, 8)
	indices := make(map[int32]uint64)
	for _, biome := range biomes {
		if _, exists := indices[biome]; !exists {
			indices[biome] = uint64(len(palette))
			palette = append(palette, biome)
		}
	}
	entries := make([]nbt.Tag, 0, len(palette))
	for _, biome := range palette {
		name, err := registry.BiomeName(biome)
		if err != nil {
			return nil, err
		}
		entries = append(entries, nbt.Tag{Type: nbt.String, Value: name})
	}
	container := nbt.Object{"palette": {Type: nbt.List, ListType: nbt.String, Value: entries}}
	if len(palette) > 1 {
		bitsPer := bits.Len(uint(len(palette) - 1))
		perLong := 64 / bitsPer
		data := make([]int64, (len(biomes)+perLong-1)/perLong)
		for index, biome := range biomes {
			data[index/perLong] |= int64(indices[biome] << (uint(index%perLong) * uint(bitsPer)))
		}
		container["data"] = nbt.Tag{Type: nbt.LongArray, Value: data}
	}
	return container, nil
}

func (c *Chunk) initializeAnvilRoot() error {
	sections := make([]nbt.Tag, 0, len(c.sect))
	for si := range c.sect {
		y := c.minY + int32(si)*16
		states := make([]uint32, 4096)
		for i := range states {
			states[i] = c.GetBlock(i&15, (i>>4)&15, y+int32(i>>8))
		}
		blockStates, err := encodeAnvilBlockStates(states)
		if err != nil {
			return err
		}
		biomes, err := encodeAnvilBiomes(c.biomes[si])
		if err != nil {
			return err
		}
		section := nbt.Object{
			"Y":            {Type: nbt.Byte, Value: int8(y / 16)},
			"block_states": {Type: nbt.Compound, Value: blockStates},
			"biomes":       {Type: nbt.Compound, Value: biomes},
		}
		sections = append(sections, nbt.Tag{Type: nbt.Compound, Value: section})
	}
	motion, surface, ocean := c.anvilHeightmaps()
	root := nbt.Object{
		"DataVersion": {Type: nbt.Int, Value: int32(3955)},
		"xPos":        {Type: nbt.Int, Value: c.cx}, "zPos": {Type: nbt.Int, Value: c.cz},
		"yPos":           {Type: nbt.Int, Value: c.minY / 16},
		"Status":         {Type: nbt.String, Value: "minecraft:full"},
		"LastUpdate":     {Type: nbt.Long, Value: int64(0)},
		"InhabitedTime":  {Type: nbt.Long, Value: int64(0)},
		"isLightOn":      {Type: nbt.Byte, Value: int8(0)},
		"sections":       {Type: nbt.List, ListType: nbt.Compound, Value: sections},
		"block_entities": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{}},
		"block_ticks":    {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{}},
		"fluid_ticks":    {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{}},
		"PostProcessing": {Type: nbt.List, ListType: nbt.List, Value: []nbt.Tag{}},
		"Heightmaps": {Type: nbt.Compound, Value: nbt.Object{
			"MOTION_BLOCKING": {Type: nbt.LongArray, Value: motion},
			"WORLD_SURFACE":   {Type: nbt.LongArray, Value: surface},
			"OCEAN_FLOOR":     {Type: nbt.LongArray, Value: ocean},
		}},
		"structures": {Type: nbt.Compound, Value: nbt.Object{
			"starts":     {Type: nbt.Compound, Value: nbt.Object{}},
			"References": {Type: nbt.Compound, Value: nbt.Object{}},
		}},
	}
	c.anvilName = ""
	c.anvilRoot = nbt.Tag{Type: nbt.Compound, Value: root}
	c.fromAnvil = true
	c.dirty = true
	return nil
}
