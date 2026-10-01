// Package nbt reads and writes the named-root NBT format used by Java Edition
// save files. The tree retains tag widths, list element types, and unknown
// fields so an edited chunk can be written without discarding other data.
package nbt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

const (
	End byte = iota
	Byte
	Short
	Int
	Long
	Float
	Double
	ByteArray
	String
	List
	Compound
	IntArray
	LongArray
)

const maxDepth = 64
const maxData = 32 << 20

type Tag struct {
	Type     byte
	Value    any
	ListType byte // retained even for an empty list
}

type Object map[string]Tag

func (t Tag) Object() (Object, bool) { v, ok := t.Value.(Object); return v, ok && t.Type == Compound }
func (t Tag) Text() (string, bool)   { v, ok := t.Value.(string); return v, ok && t.Type == String }
func (t Tag) Number() (int64, bool) {
	switch v := t.Value.(type) {
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	default:
		return 0, false
	}
}
func (t Tag) Elements() ([]Tag, bool) { v, ok := t.Value.([]Tag); return v, ok && t.Type == List }
func (t Tag) Longs() ([]int64, bool)  { v, ok := t.Value.([]int64); return v, ok && t.Type == LongArray }

func (o Object) Object(name string) (Object, bool) { return o[name].Object() }
func (o Object) Text(name string) (string, bool)   { return o[name].Text() }
func (o Object) Number(name string) (int64, bool)  { return o[name].Number() }

type decoder struct {
	data []byte
	pos  int
}

func Parse(data []byte) (string, Tag, error) {
	if len(data) > maxData {
		return "", Tag{}, errors.New("NBT exceeds 32 MiB limit")
	}
	d := &decoder{data: data}
	t, err := d.u8()
	if err != nil {
		return "", Tag{}, err
	}
	if t != Compound {
		return "", Tag{}, errors.New("NBT root must be a compound")
	}
	name, err := d.text()
	if err != nil {
		return "", Tag{}, err
	}
	value, err := d.value(t, 0)
	if err != nil {
		return "", Tag{}, err
	}
	if d.pos != len(data) {
		return "", Tag{}, fmt.Errorf("%d trailing NBT bytes", len(data)-d.pos)
	}
	return name, value, nil
}

// ParseNetwork reads one anonymous network tag and reports its consumed length.
// Since 1.20.2 the root name is omitted; text components may have any tag type.
func ParseNetwork(data []byte) (Tag, int, error) {
	if len(data) > maxData {
		return Tag{}, 0, errors.New("NBT exceeds 32 MiB limit")
	}
	d := &decoder{data: data}
	t, err := d.u8()
	if err != nil {
		return Tag{}, 0, err
	}
	if t == End {
		return Tag{Type: End}, 1, nil
	}
	value, err := d.value(t, 0)
	return value, d.pos, err
}

func MarshalNetwork(tag Tag) ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte(tag.Type)
	if tag.Type != End {
		if err := writeValue(&b, tag, 0); err != nil {
			return nil, err
		}
	}
	if b.Len() > maxData {
		return nil, errors.New("NBT exceeds 32 MiB limit")
	}
	return b.Bytes(), nil
}

func (d *decoder) take(n int) ([]byte, error) {
	if n < 0 || n > len(d.data)-d.pos {
		return nil, errors.New("truncated or invalid NBT length")
	}
	b := d.data[d.pos : d.pos+n]
	d.pos += n
	return b, nil
}
func (d *decoder) u8() (byte, error) {
	b, err := d.take(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}
func (d *decoder) u16() (uint16, error) {
	b, err := d.take(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}
func (d *decoder) u32() (uint32, error) {
	b, err := d.take(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}
func (d *decoder) u64() (uint64, error) {
	b, err := d.take(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}
func (d *decoder) text() (string, error) {
	n, err := d.u16()
	if err != nil {
		return "", err
	}
	b, err := d.take(int(n))
	return string(b), err
}
func (d *decoder) count(width int) (int, error) {
	v, err := d.u32()
	if err != nil {
		return 0, err
	}
	if v > uint32((len(d.data)-d.pos)/width) {
		return 0, errors.New("invalid NBT collection length")
	}
	return int(v), nil
}

func (d *decoder) value(t byte, depth int) (Tag, error) {
	if depth > maxDepth {
		return Tag{}, errors.New("NBT nesting limit exceeded")
	}
	var v any
	var listType byte
	switch t {
	case Byte:
		x, err := d.u8()
		if err != nil {
			return Tag{}, err
		}
		v = int8(x)
	case Short:
		x, err := d.u16()
		if err != nil {
			return Tag{}, err
		}
		v = int16(x)
	case Int:
		x, err := d.u32()
		if err != nil {
			return Tag{}, err
		}
		v = int32(x)
	case Long:
		x, err := d.u64()
		if err != nil {
			return Tag{}, err
		}
		v = int64(x)
	case Float:
		x, err := d.u32()
		if err != nil {
			return Tag{}, err
		}
		v = math.Float32frombits(x)
	case Double:
		x, err := d.u64()
		if err != nil {
			return Tag{}, err
		}
		v = math.Float64frombits(x)
	case String:
		x, err := d.text()
		if err != nil {
			return Tag{}, err
		}
		v = x
	case ByteArray:
		n, err := d.count(1)
		if err != nil {
			return Tag{}, err
		}
		b, err := d.take(n)
		if err != nil {
			return Tag{}, err
		}
		v = append([]byte(nil), b...)
	case IntArray:
		n, err := d.count(4)
		if err != nil {
			return Tag{}, err
		}
		a := make([]int32, n)
		for i := range a {
			x, _ := d.u32()
			a[i] = int32(x)
		}
		v = a
	case LongArray:
		n, err := d.count(8)
		if err != nil {
			return Tag{}, err
		}
		a := make([]int64, n)
		for i := range a {
			x, _ := d.u64()
			a[i] = int64(x)
		}
		v = a
	case List:
		var err error
		listType, err = d.u8()
		if err != nil {
			return Tag{}, err
		}
		if listType > LongArray {
			return Tag{}, errors.New("invalid NBT list element type")
		}
		n, err := d.count(1)
		if err != nil {
			return Tag{}, err
		}
		if listType == End && n != 0 {
			return Tag{}, errors.New("nonempty NBT end list")
		}
		a := make([]Tag, 0, n)
		for i := 0; i < n; i++ {
			item, e := d.value(listType, depth+1)
			if e != nil {
				return Tag{}, e
			}
			a = append(a, item)
		}
		v = a
	case Compound:
		o := make(Object)
		for {
			kind, err := d.u8()
			if err != nil {
				return Tag{}, err
			}
			if kind == End {
				break
			}
			if kind > LongArray {
				return Tag{}, errors.New("invalid NBT tag type")
			}
			name, err := d.text()
			if err != nil {
				return Tag{}, err
			}
			item, err := d.value(kind, depth+1)
			if err != nil {
				return Tag{}, err
			}
			o[name] = item
		}
		v = o
	default:
		return Tag{}, fmt.Errorf("invalid NBT tag type %d", t)
	}
	return Tag{Type: t, Value: v, ListType: listType}, nil
}

func Marshal(name string, root Tag) ([]byte, error) {
	if root.Type != Compound {
		return nil, errors.New("NBT root must be compound")
	}
	var b bytes.Buffer
	if err := b.WriteByte(Compound); err != nil {
		return nil, err
	}
	if err := writeText(&b, name); err != nil {
		return nil, err
	}
	if err := writeValue(&b, root, 0); err != nil {
		return nil, err
	}
	if b.Len() > maxData {
		return nil, errors.New("NBT exceeds 32 MiB limit")
	}
	return b.Bytes(), nil
}

func writeText(b *bytes.Buffer, s string) error {
	if len(s) > 65535 {
		return errors.New("NBT string too long")
	}
	var head [2]byte
	binary.BigEndian.PutUint16(head[:], uint16(len(s)))
	b.Write(head[:])
	b.WriteString(s)
	return nil
}
func writeValue(b *bytes.Buffer, t Tag, depth int) error {
	if depth > maxDepth {
		return errors.New("NBT nesting limit exceeded")
	}
	switch t.Type {
	case Byte:
		b.WriteByte(byte(t.Value.(int8)))
	case Short:
		var a [2]byte
		binary.BigEndian.PutUint16(a[:], uint16(t.Value.(int16)))
		b.Write(a[:])
	case Int:
		var a [4]byte
		binary.BigEndian.PutUint32(a[:], uint32(t.Value.(int32)))
		b.Write(a[:])
	case Long:
		var a [8]byte
		binary.BigEndian.PutUint64(a[:], uint64(t.Value.(int64)))
		b.Write(a[:])
	case Float:
		var a [4]byte
		binary.BigEndian.PutUint32(a[:], math.Float32bits(t.Value.(float32)))
		b.Write(a[:])
	case Double:
		var a [8]byte
		binary.BigEndian.PutUint64(a[:], math.Float64bits(t.Value.(float64)))
		b.Write(a[:])
	case String:
		return writeText(b, t.Value.(string))
	case ByteArray:
		a := t.Value.([]byte)
		if len(a) > maxData {
			return errors.New("NBT array too large")
		}
		writeCount(b, len(a))
		b.Write(a)
	case IntArray:
		a := t.Value.([]int32)
		writeCount(b, len(a))
		for _, x := range a {
			writeU32(b, uint32(x))
		}
	case LongArray:
		a := t.Value.([]int64)
		writeCount(b, len(a))
		for _, x := range a {
			writeU64(b, uint64(x))
		}
	case List:
		a := t.Value.([]Tag)
		b.WriteByte(t.ListType)
		writeCount(b, len(a))
		for _, item := range a {
			if item.Type != t.ListType {
				return errors.New("mixed NBT list types")
			}
			if err := writeValue(b, item, depth+1); err != nil {
				return err
			}
		}
	case Compound:
		o := t.Value.(Object)
		for name, item := range o {
			if item.Type == End || item.Type > LongArray {
				return errors.New("invalid NBT compound item")
			}
			b.WriteByte(item.Type)
			if err := writeText(b, name); err != nil {
				return err
			}
			if err := writeValue(b, item, depth+1); err != nil {
				return err
			}
		}
		b.WriteByte(End)
	default:
		return fmt.Errorf("invalid NBT tag type %d", t.Type)
	}
	if b.Len() > maxData {
		return errors.New("NBT exceeds 32 MiB limit")
	}
	return nil
}
func writeCount(b *bytes.Buffer, n int) { writeU32(b, uint32(n)) }
func writeU32(b *bytes.Buffer, v uint32) {
	var a [4]byte
	binary.BigEndian.PutUint32(a[:], v)
	b.Write(a[:])
}
func writeU64(b *bytes.Buffer, v uint64) {
	var a [8]byte
	binary.BigEndian.PutUint64(a[:], v)
	b.Write(a[:])
}
