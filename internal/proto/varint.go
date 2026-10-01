package proto

import "errors"

var (
	ErrVarIntTooBig  = errors.New("proto: varint too big")
	ErrTruncated     = errors.New("proto: not enough bytes")
	ErrStringTooLong = errors.New("proto: string exceeds max length")
)

// AppendVarInt appends a Minecraft VarInt (1-5 bytes) to b.
func AppendVarInt(b []byte, v int32) []byte {
	u := uint32(v)
	for {
		if u&^0x7F == 0 {
			return append(b, byte(u))
		}
		b = append(b, byte(u&0x7F)|0x80)
		u >>= 7
	}
}

// AppendVarLong appends a Minecraft VarLong (1-10 bytes) to b.
func AppendVarLong(b []byte, v int64) []byte {
	u := uint64(v)
	for {
		if u&^0x7F == 0 {
			return append(b, byte(u))
		}
		b = append(b, byte(u&0x7F)|0x80)
		u >>= 7
	}
}

// ReadVarInt decodes a VarInt from data starting at pos.
// Returns the value and the number of bytes consumed.
func ReadVarInt(data []byte, pos int) (int32, int, error) {
	var u uint32
	for i := 0; i < 5; i++ {
		if pos+i >= len(data) {
			return 0, 0, ErrTruncated
		}
		c := data[pos+i]
		if i == 4 && c&0x7F > 0x0F {
			// 5th byte carries only 4 significant bits; anything else
			// overflows 32 bits (vanilla rejects this as "VarInt too big")
			return 0, 0, ErrVarIntTooBig
		}
		u |= uint32(c&0x7F) << (7 * i)
		if c&0x80 == 0 {
			return int32(u), i + 1, nil
		}
	}
	return 0, 0, ErrVarIntTooBig
}

// ReadVarLong decodes a VarLong from data starting at pos.
func ReadVarLong(data []byte, pos int) (int64, int, error) {
	var u uint64
	for i := 0; i < 10; i++ {
		if pos+i >= len(data) {
			return 0, 0, ErrTruncated
		}
		c := data[pos+i]
		u |= uint64(c&0x7F) << (7 * i)
		if c&0x80 == 0 {
			return int64(u), i + 1, nil
		}
	}
	return 0, 0, ErrVarIntTooBig
}

// VarIntLen returns the encoded size of v.
func VarIntLen(v int32) int {
	u := uint32(v)
	n := 1
	for u&^0x7F != 0 {
		u >>= 7
		n++
	}
	return n
}
