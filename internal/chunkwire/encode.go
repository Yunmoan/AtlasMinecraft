// Package chunkwire encodes Java protocol 767 chunks and light updates.
// Input slices are borrowed only during synchronous encoding; world ownership,
// generation, persistence, and lighting computation remain with callers.
package chunkwire

import (
	"atlas/internal/light"
	"atlas/internal/proto"
	"math/bits"
)

type Section struct {
	NonAir  int16
	Bits    uint8
	Palette []uint32
	Data    []int64
	Biomes  [64]int32
}

type Column struct {
	X, Z                                int32
	Sections, Heightmaps, BlockEntities []byte
	Light                               *light.Data
}

// Encode assembles a full chunk-data payload from already prepared contents.
func Encode(column Column) []byte {
	w := proto.NewWriterCap(len(column.Sections) + 6*2048 + 512)
	w.I32(column.X).I32(column.Z).Raw(column.Heightmaps)
	w.VarInt(int32(len(column.Sections))).Raw(column.Sections).Raw(column.BlockEntities)
	WriteLight(w, column.Light)
	return w.Payload()
}

// Sections always includes a length-prefixed data array, even for a uniform
// palette. Omitting the zero length would shift all subsequent section data.
func Sections(sections []Section) []byte {
	w := proto.NewWriterCap(8192)
	for _, sec := range sections {
		if sec.NonAir == 0 {
			w.I16(0).U8(0).VarInt(0).VarInt(0)
			encodeBiomes(w, sec.Biomes)
			continue
		}
		w.I16(sec.NonAir)
		if sec.Bits == 0 {
			w.U8(0).VarInt(int32(sec.Palette[0])).VarInt(0)
		} else if sec.Bits == 15 {
			w.U8(15).VarInt(int32(len(sec.Data)))
			for _, value := range sec.Data {
				w.I64(value)
			}
		} else {
			w.U8(sec.Bits).VarInt(int32(len(sec.Palette)))
			for _, value := range sec.Palette {
				w.VarInt(int32(value))
			}
			w.VarInt(int32(len(sec.Data)))
			for _, value := range sec.Data {
				w.I64(value)
			}
		}
		encodeBiomes(w, sec.Biomes)
	}
	return w.Payload()
}

func encodeBiomes(w *proto.Writer, biomes [64]int32) {
	palette := make([]int32, 0, 8)
	indices := make(map[int32]uint64)
	for _, id := range biomes {
		if _, ok := indices[id]; !ok {
			indices[id] = uint64(len(palette))
			palette = append(palette, id)
		}
	}
	if len(palette) == 1 {
		w.U8(0).VarInt(palette[0]).VarInt(0)
		return
	}
	bits := uint8(1)
	for 1<<bits < len(palette) {
		bits++
	}
	if bits > 3 {
		bits = 6
	}
	w.U8(bits)
	if bits <= 3 {
		w.VarInt(int32(len(palette)))
		for _, id := range palette {
			w.VarInt(id)
		}
	}
	perLong := 64 / int(bits)
	longs := make([]int64, (64+perLong-1)/perLong)
	for i, id := range biomes {
		value := uint64(id)
		if bits <= 3 {
			value = indices[id]
		}
		longs[i/perLong] |= int64(value << (uint(i%perLong) * uint(bits)))
	}
	w.VarInt(int32(len(longs)))
	for _, value := range longs {
		w.I64(value)
	}
}

func Heightmaps(motion, surface []int64) []byte {
	return proto.NewNBT().LongArray("MOTION_BLOCKING", motion).LongArray("WORLD_SURFACE", surface).Finish()
}

// WriteLight is shared by full chunk delivery and incremental light packets.
func WriteLight(w *proto.Writer, data *light.Data) {
	var skyMask, blockMask int64
	for i := range data.Sky {
		if data.Sky[i] != nil {
			skyMask |= 1 << i
		}
		if data.Block[i] != nil {
			blockMask |= 1 << i
		}
	}
	all := int64(1)<<len(data.Sky) - 1
	w.VarInt(1).I64(skyMask).VarInt(1).I64(blockMask).VarInt(1).I64(all &^ skyMask).VarInt(1).I64(all &^ blockMask)
	for _, item := range []struct {
		mask   int64
		layers [][]byte
	}{{skyMask, data.Sky}, {blockMask, data.Block}} {
		w.VarInt(int32(bits.OnesCount64(uint64(item.mask))))
		for i, layer := range item.layers {
			if item.mask&(1<<i) != 0 {
				w.VarInt(int32(len(layer))).Raw(layer)
			}
		}
	}
}
