package nbt

import (
	"errors"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Java DataInput/DataOutput use modified UTF-8 for NBT strings: NUL is C0 80,
// and supplementary characters are two three-byte UTF-16 surrogate units.
func modifiedUTF8String(s string) (string, error) {
	if !utf8.ValidString(s) {
		return "", errors.New("invalid UTF-8 in NBT string")
	}
	needsEncoding := false
	for _, r := range s {
		if r == 0 || r > 0xffff {
			needsEncoding = true
			break
		}
	}
	if !needsEncoding {
		return s, nil
	}
	var b strings.Builder
	unit := func(u uint16) {
		switch {
		case u >= 1 && u <= 0x7f:
			b.WriteByte(byte(u))
		case u <= 0x7ff:
			b.WriteByte(0xc0 | byte(u>>6))
			b.WriteByte(0x80 | byte(u&0x3f))
		default:
			b.WriteByte(0xe0 | byte(u>>12))
			b.WriteByte(0x80 | byte(u>>6&0x3f))
			b.WriteByte(0x80 | byte(u&0x3f))
		}
	}
	for _, r := range s {
		if r <= 0xffff {
			unit(uint16(r))
		} else {
			high, low := utf16.EncodeRune(r)
			unit(uint16(high))
			unit(uint16(low))
		}
	}
	return b.String(), nil
}

func parseModifiedUTF8(b []byte) (string, error) {
	// Also accept ordinary UTF-8 written by earlier Atlas versions; all new
	// writes use Java's encoding. ASCII/BMP strings need no conversion.
	if utf8.Valid(b) {
		return string(b), nil
	}
	units := make([]uint16, 0, len(b))
	for i := 0; i < len(b); {
		first := b[i]
		switch {
		case first < 0x80:
			units = append(units, uint16(first))
			i++
		case first&0xe0 == 0xc0:
			if i+1 >= len(b) || b[i+1]&0xc0 != 0x80 {
				return "", errors.New("invalid modified UTF-8 in NBT")
			}
			units = append(units, uint16(first&0x1f)<<6|uint16(b[i+1]&0x3f))
			i += 2
		case first&0xf0 == 0xe0:
			if i+2 >= len(b) || b[i+1]&0xc0 != 0x80 || b[i+2]&0xc0 != 0x80 {
				return "", errors.New("invalid modified UTF-8 in NBT")
			}
			units = append(units, uint16(first&0xf)<<12|uint16(b[i+1]&0x3f)<<6|uint16(b[i+2]&0x3f))
			i += 3
		default:
			return "", errors.New("invalid modified UTF-8 in NBT")
		}
	}
	for i := 0; i < len(units); i++ {
		if units[i] >= 0xd800 && units[i] <= 0xdbff {
			if i+1 >= len(units) || units[i+1] < 0xdc00 || units[i+1] > 0xdfff {
				return "", errors.New("unpaired surrogate in NBT string")
			}
			i++
		} else if units[i] >= 0xdc00 && units[i] <= 0xdfff {
			return "", errors.New("unpaired surrogate in NBT string")
		}
	}
	return string(utf16.Decode(units)), nil
}
