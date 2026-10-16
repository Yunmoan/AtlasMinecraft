package world

import (
	"atlas/internal/light"
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/registry"
	"time"
)

const (
	MinY        = -64
	MaxY        = 320 // exclusive
	WorldHeight = MaxY - MinY
	NumSections = WorldHeight / 16
	MaxSectionY = NumSections - 1

	SectionVolume = 4096
	// docs/protocol/blocks.json covers every 1.21.1 state from 0..26683.
	MaxBlockState = 26683

	// Block state ids for 1.21.1 (PrismarineJS blocks.json defaultState).
	StateAir     = 0
	StateStone   = 1
	StateGrass   = 9 // grass_block
	StateDirt    = 10
	StateBedrock = 79
	StateVoidAir = 12958
	StateCaveAir = 12959
)

func isAir(state uint32) bool {
	return state == StateAir || state == StateVoidAir || state == StateCaveAir
}

// Section is one 16x16x16 chunk section. bits==0 means uniform: the whole
// section is palette[0] and no packed data exists.
type Section struct {
	nonAir  int16
	bits    uint8
	palette []uint32
	data    []int64
}

func newUniformSection(v uint32) *Section {
	return &Section{palette: []uint32{v}}
}

func (s *Section) entriesPerLong() int { return 64 / int(s.bits) }

func (s *Section) get(i int) uint32 {
	if s.bits == 0 {
		return s.palette[0]
	}
	if s.bits == 15 {
		// direct encoding: the raw global state id, 15 bits per entry
		return uint32((s.data[i/4] >> (uint(i%4) * 15)) & 0x7FFF)
	}
	idx := (s.data[i/s.entriesPerLong()] >> (uint(i%s.entriesPerLong()) * uint(s.bits))) &
		((int64(1) << s.bits) - 1)
	return s.palette[idx]
}

func (s *Section) write(i int, paletteIdx int64) {
	epL := s.entriesPerLong()
	sh := uint(i%epL) * uint(s.bits)
	mask := ((int64(1) << s.bits) - 1) << sh
	p := &s.data[i/epL]
	*p = (*p &^ mask) | (paletteIdx << sh & mask)
}

func (s *Section) writeDirect(i int, v uint32) {
	sh := uint(i%4) * 15
	mask := int64(0x7FFF) << sh
	p := &s.data[i/4]
	*p = (*p &^ mask) | (int64(v) << sh & mask)
}

// toDirect converts the section to 15-bit direct encoding (no palette) —
// the escape hatch when the palette outgrows 8 bits.
func (s *Section) toDirect() {
	old := make([]uint32, SectionVolume)
	for i := 0; i < SectionVolume; i++ {
		old[i] = s.get(i)
	}
	s.bits = 15
	s.palette = nil
	s.data = make([]int64, (SectionVolume+3)/4)
	for i, v := range old {
		if v != 0 {
			s.writeDirect(i, v)
		}
	}
}

func (s *Section) set(i int, v uint32) {
	old := s.get(i)
	if old == v {
		return
	}
	if !isAir(old) && isAir(v) {
		s.nonAir--
	} else if isAir(old) && !isAir(v) {
		s.nonAir++
	}
	if s.bits == 0 {
		s.expand(i, v)
		return
	}
	if s.bits == 15 {
		s.writeDirect(i, v)
		return
	}
	idx := s.indexOf(v)
	if idx < 0 {
		if len(s.palette) >= 1<<s.bits {
			if s.bits >= 8 {
				s.toDirect()
				s.writeDirect(i, v)
				return
			}
			s.grow()
		}
		s.palette = append(s.palette, v)
		idx = int64(len(s.palette) - 1)
	}
	s.write(i, idx)
}

func (s *Section) indexOf(v uint32) int64 {
	for j, p := range s.palette {
		if p == v {
			return int64(j)
		}
	}
	return -1
}

// expand converts a uniform section into a 4-bit paletted one with two
// entries (the old uniform value at index 0, v at index 1) and writes i.
func (s *Section) expand(i int, v uint32) {
	old := s.palette[0]
	s.bits = 4
	s.palette = []uint32{old, v}
	s.data = make([]int64, (SectionVolume+s.entriesPerLong()-1)/s.entriesPerLong())
	s.write(i, 1)
}

// grow bumps bits per entry and repacks, making room for one more palette
// entry. Our generator uses few block types per section; a full transition
// to direct encoding is out of scope for this build.
func (s *Section) grow() {
	if s.bits >= 8 {
		panic("world: palette overflow; direct encoding not implemented")
	}
	old := make([]uint32, len(s.palette))
	copy(old, s.palette)
	oldBits := s.bits
	oldData := s.data

	s.bits++
	s.data = make([]int64, (SectionVolume+s.entriesPerLong()-1)/s.entriesPerLong())
	for i := 0; i < SectionVolume; i++ {
		var v uint32
		if oldBits == 0 {
			v = old[0]
		} else {
			epL := 64 / int(oldBits)
			idx := (oldData[i/epL] >> (uint(i%epL) * uint(oldBits))) & ((int64(1) << oldBits) - 1)
			v = old[idx]
		}
		// write via palette index in the (still old-sized) palette
		pi := int64(-1)
		for j, p := range s.palette {
			if p == v {
				pi = int64(j)
				break
			}
		}
		s.write(i, pi)
	}
}

// Chunk is a 16x384x16 column.
type Chunk struct {
	cx, cz        int32
	heights       [256]int16 // motion-blocking top, absolute Y of first air above ground
	sect          []*Section
	biomes        [][64]int32
	anvilName     string
	anvilRoot     nbt.Tag
	entities      []WorldEntity
	fromAnvil     bool
	dirty         bool
	dirtySections []bool
	minY, maxY    int32
	hasSky        bool
	lights        *light.Data
	lightStale    bool
	blockRevision uint64
	lastAccess    time.Time // region actor only; lighting changes do not keep it alive
}

func NewChunk(cx, cz int32) *Chunk {
	return newChunkBounds(cx, cz, MinY, MaxY, true)
}

func newChunkBounds(cx, cz, minY, maxY int32, hasSky bool) *Chunk {
	count := int((maxY - minY) / 16)
	c := &Chunk{cx: cx, cz: cz, minY: minY, maxY: maxY, hasSky: hasSky,
		sect: make([]*Section, count), biomes: make([][64]int32, count), dirtySections: make([]bool, count)}
	for i := range c.heights {
		c.heights[i] = int16(minY)
	}
	for si := range c.biomes {
		for i := range c.biomes[si] {
			c.biomes[si][i] = registry.BiomePlains
		}
	}
	return c
}

func sectionOf(y int32) int { return int((y - MinY) >> 4) }

func (c *Chunk) sectionOf(y int32) int { return int((y - c.minY) >> 4) }
func (c *Chunk) MinBuildHeight() int32 { return c.minY }
func (c *Chunk) MaxBuildHeight() int32 { return c.maxY }

func localIndexOf(lx, ly, lz int) int {
	return (ly&0xF)<<8 | lz<<4 | lx
}

// SetBlock writes a state at chunk-local x/z (0..15) and absolute y.
func (c *Chunk) SetBlock(lx, lz int, y int32, state uint32) {
	si := c.sectionOf(y)
	if si < 0 || si >= len(c.sect) {
		return
	}
	old := c.GetBlock(lx, lz, y)
	if old == state {
		return
	}
	sec := c.sect[si]
	if sec == nil {
		if state == StateAir {
			return
		}
		sec = newUniformSection(StateAir)
		c.sect[si] = sec
	}
	sec.set(localIndexOf(lx, int(y), lz), state)
	c.removeReplacedBlockEntity(lx, lz, y, old, state)
	c.blockRevision++
	c.dirty = true
	c.dirtySections[si] = true

	hi := &c.heights[lz<<4|lx]
	if !isAir(state) {
		if y >= int32(*hi) {
			*hi = int16(y + 1)
		}
	} else if int32(*hi) == y+1 {
		// scan down for the new top
		ny := y
		for ny > c.minY && isAir(c.GetBlock(lx, lz, ny-1)) {
			ny--
		}
		*hi = int16(ny)
	}
}

// GetBlock reads a state at chunk-local x/z and absolute y.
func (c *Chunk) GetBlock(lx, lz int, y int32) uint32 {
	si := c.sectionOf(y)
	if si < 0 || si >= len(c.sect) {
		return StateAir
	}
	sec := c.sect[si]
	if sec == nil {
		return StateAir
	}
	return sec.get(localIndexOf(lx, int(y), lz))
}

func (c *Chunk) Height(lx, lz int) int32 { return int32(c.heights[lz<<4|lx]) }

// EncodeSections serializes the 24 block-state sections plus biome
// containers (uniform plains) into the wire-format chunk data blob.
//
// Palette Container wire layout (verified byte-exact against a vanilla
// 1.21.1 capture): Bits Per Entry, Palette (absent when bits==0, where the
// single value follows), then the Data Array — which is ALWAYS
// length-prefixed, even when bits==0 and the array is empty. Omitting that
// varint shifts every subsequent section and corrupts the client's world.
func (c *Chunk) EncodeSections() []byte {
	w := proto.NewWriterCap(8192)
	for si := range c.sect {
		sec := c.sect[si]
		if sec == nil || sec.nonAir == 0 {
			// uniform air section
			w.I16(0).U8(0).VarInt(StateAir).VarInt(0)
			c.encodeBiomes(w, si)
			continue
		}
		w.I16(sec.nonAir)
		if sec.bits == 0 {
			w.U8(0).VarInt(int32(sec.palette[0])).VarInt(0)
		} else if sec.bits == 15 {
			// direct encoding: no palette, raw global state ids
			w.U8(15)
			w.VarInt(int32(len(sec.data)))
			for _, l := range sec.data {
				w.I64(l)
			}
		} else {
			w.U8(sec.bits).VarInt(int32(len(sec.palette)))
			for _, p := range sec.palette {
				w.VarInt(int32(p))
			}
			w.VarInt(int32(len(sec.data)))
			for _, l := range sec.data {
				w.I64(l)
			}
		}
		c.encodeBiomes(w, si)
	}
	return w.Payload()
}

func (c *Chunk) encodeBiomes(w *proto.Writer, si int) {
	palette := make([]int32, 0, 8)
	indices := make(map[int32]uint64)
	for _, id := range c.biomes[si] {
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
	for i, id := range c.biomes[si] {
		v := uint64(id)
		if bits <= 3 {
			v = indices[id]
		}
		longs[i/perLong] |= int64(v << (uint(i%perLong) * uint(bits)))
	}
	w.VarInt(int32(len(longs)))
	for _, v := range longs {
		w.I64(v)
	}
}

// HeightmapNBT builds the heightmap compound (MOTION_BLOCKING +
// WORLD_SURFACE, both present in vanilla chunk packets).
func (c *Chunk) HeightmapNBT() []byte {
	motion, surface, _ := c.anvilHeightmaps()
	return proto.NewNBT().
		LongArray("MOTION_BLOCKING", motion).
		LongArray("WORLD_SURFACE", surface).
		Finish()
}
