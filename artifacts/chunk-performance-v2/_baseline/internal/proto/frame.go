package proto

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"io"
	"net"
	"sync"
)

// MaxPacketSize is the largest framed packet we will read (8 MiB, generous).
const MaxPacketSize = 1 << 23

var zlibWriterPool = sync.Pool{New: func() any { return zlib.NewWriter(nil) }}

// FramedConn wraps a net.Conn with Minecraft packet framing and optional
// zlib compression. It is not safe for concurrent use; the server gives each
// connection a single reader goroutine and serializes writes itself.
type FramedConn struct {
	Conn      net.Conn
	rd        *bufio.Reader
	wr        io.Writer
	Threshold int32 // < 0 means compression disabled
}

func NewFramedConn(c net.Conn) *FramedConn {
	return &FramedConn{
		Conn:      c,
		rd:        bufio.NewReaderSize(c, 1<<16),
		wr:        c,
		Threshold: -1,
	}
}

// EnableCompression turns on zlib framing from this point on.
func (f *FramedConn) EnableCompression(threshold int32) {
	f.Threshold = threshold
}

// ReadPacket reads one framed packet and returns its VarInt packet ID
// and payload (ID already stripped).
func (f *FramedConn) ReadPacket() (int32, []byte, error) {
	length, err := readVarIntFrom(f.rd, 5)
	if err != nil {
		return 0, nil, err
	}
	if length < 0 || length > MaxPacketSize {
		return 0, nil, ErrVarIntTooBig
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(f.rd, body); err != nil {
		return 0, nil, err
	}

	if f.Threshold < 0 {
		return splitID(body)
	}

	dataLen, n, err := ReadVarInt(body, 0)
	if err != nil {
		return 0, nil, err
	}
	raw := body[n:]
	if dataLen == 0 {
		// stored uncompressed
		return splitID(raw)
	}
	if dataLen < 0 || dataLen > MaxPacketSize {
		return 0, nil, ErrVarIntTooBig
	}
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	defer zr.Close()
	out := make([]byte, dataLen)
	if _, err := io.ReadFull(zr, out); err != nil {
		return 0, nil, err
	}
	return splitID(out)
}

func splitID(body []byte) (int32, []byte, error) {
	id, n, err := ReadVarInt(body, 0)
	if err != nil {
		return 0, nil, err
	}
	return id, body[n:], nil
}

// WritePacket frames and writes one packet.
func (f *FramedConn) WritePacket(id int32, payload []byte) error {
	head := AppendVarInt(nil, id)
	body := append(head, payload...)

	var frame []byte
	if f.Threshold < 0 {
		frame = AppendVarInt(nil, int32(len(body)))
		frame = append(frame, body...)
	} else if int32(len(body)) < f.Threshold {
		frame = AppendVarInt(nil, int32(VarIntLen(0)+len(body)))
		frame = AppendVarInt(frame, 0)
		frame = append(frame, body...)
	} else {
		zw := zlibWriterPool.Get().(*zlib.Writer)
		defer zlibWriterPool.Put(zw)
		var comp bytes.Buffer
		zw.Reset(&comp)
		if _, err := zw.Write(body); err != nil {
			return err
		}
		if err := zw.Close(); err != nil {
			return err
		}
		frame = AppendVarInt(nil, int32(VarIntLen(int32(len(body)))+comp.Len()))
		frame = AppendVarInt(frame, int32(len(body)))
		frame = append(frame, comp.Bytes()...)
	}
	n, err := f.wr.Write(frame)
	if err == nil && n != len(frame) {
		return io.ErrShortWrite
	}
	return err
}

func readVarIntFrom(r io.ByteReader, maxBytes int) (int32, error) {
	var u uint32
	for i := 0; i < maxBytes; i++ {
		c, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if i == 4 && c&0x7F > 0x0F {
			return 0, ErrVarIntTooBig
		}
		u |= uint32(c&0x7F) << (7 * i)
		if c&0x80 == 0 {
			return int32(u), nil
		}
	}
	return 0, ErrVarIntTooBig
}
