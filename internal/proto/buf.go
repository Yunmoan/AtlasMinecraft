package proto

import (
	"encoding/binary"
	"math"
)

// Reader walks a packet payload (after the packet ID has been stripped).
// All methods are zero-copy where possible; nothing allocates except String.
type Reader struct {
	data []byte
	pos  int
}

func NewReader(data []byte) *Reader { return &Reader{data: data} }

func (r *Reader) Remaining() int { return len(r.data) - r.pos }

func (r *Reader) take(n int) ([]byte, error) {
	if r.Remaining() < n {
		return nil, ErrTruncated
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *Reader) VarInt() (int32, error) {
	v, n, err := ReadVarInt(r.data, r.pos)
	r.pos += n
	return v, err
}

func (r *Reader) VarLong() (int64, error) {
	v, n, err := ReadVarLong(r.data, r.pos)
	r.pos += n
	return v, err
}

func (r *Reader) U8() (byte, error) {
	b, err := r.take(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *Reader) Bool() (bool, error) {
	v, err := r.U8()
	return v != 0, err
}

func (r *Reader) I8() (int8, error) {
	v, err := r.U8()
	return int8(v), err
}

func (r *Reader) U16() (uint16, error) {
	b, err := r.take(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func (r *Reader) I16() (int16, error) {
	v, err := r.U16()
	return int16(v), err
}

func (r *Reader) I32() (int32, error) {
	b, err := r.take(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

func (r *Reader) I64() (int64, error) {
	b, err := r.take(8)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

func (r *Reader) F32() (float32, error) {
	v, err := r.I32()
	return math.Float32frombits(uint32(v)), err
}

func (r *Reader) F64() (float64, error) {
	v, err := r.I64()
	return math.Float64frombits(uint64(v)), err
}

func (r *Reader) Bytes(n int) ([]byte, error) { return r.take(n) }

func (r *Reader) Rest() []byte { return r.data[r.pos:] }

// String reads a VarInt-length-prefixed UTF-8 string (maxLen in bytes, 0 = unlimited).
func (r *Reader) String(maxLen int) (string, error) {
	n, err := r.VarInt()
	if err != nil {
		return "", err
	}
	if n < 0 || (maxLen > 0 && int(n) > maxLen*4) { // vanilla strings may expand 4x in UTF-8
		return "", ErrStringTooLong
	}
	b, err := r.take(int(n))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UUID reads a 16-byte UUID.
func (r *Reader) UUID() ([16]byte, error) {
	var u [16]byte
	b, err := r.take(16)
	if err != nil {
		return u, err
	}
	copy(u[:], b)
	return u, nil
}

// Position reads a packed block position (x:26, z:26, y:12).
func (r *Reader) Position() (x, y, z int32, err error) {
	v, err := r.I64()
	if err != nil {
		return 0, 0, 0, err
	}
	x = int64ToInt26(v >> 38)
	z = int64ToInt26(v << 26 >> 38)
	y = int64ToInt12(v << 52 >> 52)
	return
}

func int64ToInt26(v int64) int32 {
	v = v << 38 >> 38 // sign-extend low 26 bits
	return int32(v)
}

func int64ToInt12(v int64) int32 {
	v = v << 52 >> 52
	return int32(v)
}

// Writer builds a packet payload.
type Writer struct{ buf []byte }

func NewWriter() *Writer                 { return &Writer{} }
func NewWriterCap(cap int) *Writer       { return &Writer{buf: make([]byte, 0, cap)} }
func (w *Writer) Payload() []byte        { return w.buf }
func (w *Writer) Len() int               { return len(w.buf) }
func (w *Writer) Raw(b []byte) *Writer   { w.buf = append(w.buf, b...); return w }
func (w *Writer) U8(v byte) *Writer      { w.buf = append(w.buf, v); return w }
func (w *Writer) Bool(v bool) *Writer    { return w.U8(map[bool]byte{false: 0, true: 1}[v]) }
func (w *Writer) I8(v int8) *Writer      { return w.U8(byte(v)) }
func (w *Writer) VarInt(v int32) *Writer { w.buf = AppendVarInt(w.buf, v); return w }

func (w *Writer) VarLong(v int64) *Writer { w.buf = AppendVarLong(w.buf, v); return w }

func (w *Writer) U16(v uint16) *Writer {
	w.buf = binary.BigEndian.AppendUint16(w.buf, v)
	return w
}

func (w *Writer) I16(v int16) *Writer { return w.U16(uint16(v)) }

func (w *Writer) I32(v int32) *Writer {
	w.buf = binary.BigEndian.AppendUint32(w.buf, uint32(v))
	return w
}

func (w *Writer) I64(v int64) *Writer {
	w.buf = binary.BigEndian.AppendUint64(w.buf, uint64(v))
	return w
}

func (w *Writer) F32(v float32) *Writer { return w.I32(int32(math.Float32bits(v))) }

func (w *Writer) F64(v float64) *Writer { return w.I64(int64(math.Float64bits(v))) }

func (w *Writer) String(s string) *Writer {
	w.buf = AppendVarInt(w.buf, int32(len(s)))
	w.buf = append(w.buf, s...)
	return w
}

func (w *Writer) UUID(u [16]byte) *Writer { return w.Raw(u[:]) }

// PackPosition packs (x,y,z) into a single long (x:26, z:26, y:12).
func PackPosition(x, y, z int32) int64 {
	return (int64(x&0x3FFFFFF) << 38) | (int64(z&0x3FFFFFF) << 12) | int64(y&0xFFF)
}
