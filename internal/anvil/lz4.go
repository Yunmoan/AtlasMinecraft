package anvil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
)

// Minecraft's compression type 4 uses the LZ4Block stream format from
// net.jpountz.lz4, not the standard LZ4 frame format.
func decompressLZ4BlockStream(data []byte) ([]byte, error) {
	const headerSize = 21
	out := make([]byte, 0)
	for {
		if len(data) < headerSize || !bytes.Equal(data[:8], []byte("LZ4Block")) {
			return nil, errors.New("invalid LZ4Block header")
		}
		method := data[8] & 0xf0
		compressed := int(binary.LittleEndian.Uint32(data[9:13]))
		original := int(binary.LittleEndian.Uint32(data[13:17]))
		checksum := binary.LittleEndian.Uint32(data[17:21])
		data = data[headerSize:]
		if compressed == 0 && original == 0 {
			if checksum != 0 || len(data) != 0 {
				return nil, errors.New("invalid LZ4Block stream terminator")
			}
			return out, nil
		}
		if compressed < 1 || original < 1 || compressed > len(data) || original > maxChunkBytes-len(out) {
			return nil, errors.New("invalid LZ4Block lengths")
		}
		block := data[:compressed]
		data = data[compressed:]
		var decoded []byte
		switch method {
		case 0x10:
			if compressed != original {
				return nil, errors.New("invalid raw LZ4Block length")
			}
			decoded = block
		case 0x20:
			var err error
			decoded, err = decodeLZ4Block(block, original)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported LZ4Block method %x", method)
		}
		// lz4-java's Checksum adapter exposes only the low 28 bits.
		if lz4Checksum(decoded)&0x0fffffff != checksum {
			return nil, errors.New("LZ4Block checksum mismatch")
		}
		out = append(out, decoded...)
	}
}

func decodeLZ4Block(src []byte, expected int) ([]byte, error) {
	out := make([]byte, 0, expected)
	for len(src) > 0 {
		token := src[0]
		src = src[1:]
		literal := int(token >> 4)
		if literal == 15 {
			for {
				if len(src) == 0 {
					return nil, errors.New("truncated LZ4 literal length")
				}
				n := int(src[0])
				src = src[1:]
				literal += n
				if n != 255 {
					break
				}
				if literal > expected {
					return nil, errors.New("LZ4 literal length exceeds output")
				}
			}
		}
		if literal > len(src) || literal > expected-len(out) {
			return nil, errors.New("invalid LZ4 literal")
		}
		out = append(out, src[:literal]...)
		src = src[literal:]
		if len(src) == 0 {
			break
		}
		if len(src) < 2 {
			return nil, errors.New("truncated LZ4 match offset")
		}
		offset := int(binary.LittleEndian.Uint16(src[:2]))
		src = src[2:]
		if offset < 1 || offset > len(out) {
			return nil, errors.New("invalid LZ4 match offset")
		}
		match := int(token&0x0f) + 4
		if token&0x0f == 15 {
			for {
				if len(src) == 0 {
					return nil, errors.New("truncated LZ4 match length")
				}
				n := int(src[0])
				src = src[1:]
				match += n
				if n != 255 {
					break
				}
				if match > expected-len(out) {
					return nil, errors.New("LZ4 match exceeds output")
				}
			}
		}
		if match > expected-len(out) {
			return nil, errors.New("LZ4 match exceeds output")
		}
		for i := 0; i < match; i++ {
			out = append(out, out[len(out)-offset])
		}
	}
	if len(out) != expected {
		return nil, errors.New("LZ4Block output length mismatch")
	}
	return out, nil
}

func lz4Checksum(data []byte) uint32 {
	const seed uint32 = 0x9747b28c
	const p1 uint32 = 2654435761
	const p2 uint32 = 2246822519
	const p3 uint32 = 3266489917
	const p4 uint32 = 668265263
	const p5 uint32 = 374761393
	var h uint32
	remaining := data
	if len(remaining) >= 16 {
		v1, v2, v3, v4 := seed, seed, seed, seed
		v1 += p1
		v1 += p2
		v2 += p2
		v4 -= p1
		for len(remaining) >= 16 {
			v1 = bits.RotateLeft32(v1+binary.LittleEndian.Uint32(remaining[0:4])*p2, 13) * p1
			v2 = bits.RotateLeft32(v2+binary.LittleEndian.Uint32(remaining[4:8])*p2, 13) * p1
			v3 = bits.RotateLeft32(v3+binary.LittleEndian.Uint32(remaining[8:12])*p2, 13) * p1
			v4 = bits.RotateLeft32(v4+binary.LittleEndian.Uint32(remaining[12:16])*p2, 13) * p1
			remaining = remaining[16:]
		}
		h = bits.RotateLeft32(v1, 1) + bits.RotateLeft32(v2, 7) + bits.RotateLeft32(v3, 12) + bits.RotateLeft32(v4, 18)
	} else {
		h = seed + p5
	}
	h += uint32(len(data))
	for len(remaining) >= 4 {
		h = bits.RotateLeft32(h+binary.LittleEndian.Uint32(remaining[:4])*p3, 17) * p4
		remaining = remaining[4:]
	}
	for _, b := range remaining {
		h = bits.RotateLeft32(h+uint32(b)*p5, 11) * p1
	}
	h ^= h >> 15
	h *= p2
	h ^= h >> 13
	h *= p3
	h ^= h >> 16
	return h
}
