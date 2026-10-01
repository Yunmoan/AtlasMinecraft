package proto

import "encoding/binary"

// NBT is a minimal writer for the NBT shapes this server emits:
// heightmap compounds and text-component strings. Big-endian network NBT
// rooted at an unnamed compound.
type NBT struct{ buf []byte }

// NewNBT starts an anonymous root compound. Since 1.20.2 network NBT roots
// carry no name — only the type byte.
func NewNBT() *NBT {
	return &NBT{buf: []byte{0x0A}}
}

func (n *NBT) name(name string) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(len(name)))
	n.buf = append(n.buf, b...)
	n.buf = append(n.buf, name...)
}

// Finish terminates the root compound and returns the encoded bytes.
func (n *NBT) Finish() []byte { return append(n.buf, 0x00) }

// String writes a TAG_String field. NBT strings use a u16 big-endian
// length prefix — NOT a VarInt.
func (n *NBT) String(name, v string) *NBT {
	n.buf = append(n.buf, 0x08)
	n.name(name)
	b := []byte(v)
	n.buf = binary.BigEndian.AppendUint16(n.buf, uint16(len(b)))
	n.buf = append(n.buf, b...)
	return n
}

func (n *NBT) Int(name string, v int32) *NBT {
	n.buf = append(n.buf, 0x03)
	n.name(name)
	n.buf = binary.BigEndian.AppendUint32(n.buf, uint32(v))
	return n
}

func (n *NBT) Long(name string, v int64) *NBT {
	n.buf = append(n.buf, 0x04)
	n.name(name)
	n.buf = binary.BigEndian.AppendUint64(n.buf, uint64(v))
	return n
}

// LongArray writes a TAG_Long_Array.
func (n *NBT) LongArray(name string, v []int64) *NBT {
	n.buf = append(n.buf, 0x0C)
	n.name(name)
	n.buf = binary.BigEndian.AppendUint32(n.buf, uint32(len(v)))
	for _, x := range v {
		n.buf = binary.BigEndian.AppendUint64(n.buf, uint64(x))
	}
	return n
}

// BeginList writes a TAG_List header; caller writes `count` item payloads.
func (n *NBT) BeginList(name string, elemType byte, count int32) *NBT {
	n.buf = append(n.buf, 0x09)
	n.name(name)
	n.buf = append(n.buf, elemType)
	n.buf = binary.BigEndian.AppendUint32(n.buf, uint32(count))
	return n
}

// BeginCompound writes a named nested compound header.
func (n *NBT) BeginCompound(name string) *NBT {
	n.buf = append(n.buf, 0x0A)
	n.name(name)
	return n
}

// EndCompound terminates a nested compound.
func (n *NBT) EndCompound() *NBT { return n.U8Raw(0x00) }

func (n *NBT) U8Raw(v byte) *NBT {
	n.buf = append(n.buf, v)
	return n
}
