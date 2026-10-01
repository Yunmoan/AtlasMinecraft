// Package anvil handles Minecraft Java region files. Region coordinates use
// floor division, so negative chunk positions map to the correct .mca file.
package anvil

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"atlas/internal/nbt"
)

const sector = 4096
const maxRegionBytes = 256 << 20
const maxChunkBytes = 32 << 20
const maxExternalBytes = 40 << 20

func regionPath(dir string, cx, cz int32) string {
	return filepath.Join(dir, fmt.Sprintf("r.%d.%d.mca", cx>>5, cz>>5))
}
func chunkIndex(cx, cz int32) int { return int(cz&31)*32 + int(cx&31) }

// ReadChunk reads a named-root chunk NBT compound. Missing region files and
// empty location entries are reported as exists=false.
func ReadChunk(dir string, cx, cz int32) (string, nbt.Tag, bool, error) {
	kind, compressed, exists, err := readChunkRecord(dir, cx, cz)
	if err != nil || !exists {
		return "", nbt.Tag{}, false, err
	}
	return decodeChunkRecord(kind, compressed, cx, cz)
}

// Read only the location, record header and compressed payload. Close all
// handles before decompression so region replacement is never held up by CPU
// work on Windows, and reading one chunk costs no whole-region allocation.
func readChunkRecord(dir string, cx, cz int32) (byte, []byte, bool, error) {
	path := regionPath(dir, cx, cz)
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil, false, nil
	}
	if err != nil {
		return 0, nil, false, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return 0, nil, false, err
	}
	if info.Size() < 2*sector || info.Size() > maxRegionBytes {
		return 0, nil, false, fmt.Errorf("invalid region size: %s", path)
	}
	i := chunkIndex(cx, cz)
	var location [4]byte
	if _, err := f.ReadAt(location[:], int64(i*4)); err != nil {
		return 0, nil, false, err
	}
	loc := binary.BigEndian.Uint32(location[:])
	if loc == 0 {
		return 0, nil, false, nil
	}
	offset, count := int(loc>>8), int(loc&255)
	if offset < 2 || count < 1 || int64(offset*sector+5) > info.Size() {
		return 0, nil, false, fmt.Errorf("invalid chunk location %d,%d", cx, cz)
	}
	at := offset * sector
	var header [5]byte
	if _, err := f.ReadAt(header[:], int64(at)); err != nil {
		return 0, nil, false, err
	}
	n := int(binary.BigEndian.Uint32(header[:4]))
	if n < 1 || n+4 > count*sector || int64(at+4+n) > info.Size() {
		return 0, nil, false, fmt.Errorf("invalid chunk length %d,%d", cx, cz)
	}
	kind := header[4]
	var compressed []byte
	if kind&0x80 != 0 {
		compressed, err = readExternalChunk(filepath.Join(dir, fmt.Sprintf("c.%d.%d.mcc", cx, cz)))
		if err != nil {
			return 0, nil, false, err
		}
		kind &^= 0x80
	} else {
		compressed = make([]byte, n-1)
		if _, err := f.ReadAt(compressed, int64(at+5)); err != nil {
			return 0, nil, false, err
		}
	}
	return kind, compressed, true, nil
}

func decodeChunkRecord(kind byte, compressed []byte, cx, cz int32) (string, nbt.Tag, bool, error) {
	data, err := decompress(kind, compressed)
	if err != nil {
		return "", nbt.Tag{}, false, fmt.Errorf("decompress chunk %d,%d: %w", cx, cz, err)
	}
	name, root, err := nbt.Parse(data)
	if err != nil {
		return "", nbt.Tag{}, false, fmt.Errorf("parse chunk %d,%d: %w", cx, cz, err)
	}
	return name, root, true, nil
}

func readExternalChunk(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxExternalBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxExternalBytes {
		return nil, errors.New("external Anvil chunk exceeds 40 MiB limit")
	}
	return data, nil
}

func decompress(kind byte, b []byte) ([]byte, error) {
	var r io.ReadCloser
	switch kind {
	case 1:
		var err error
		r, err = gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case 2:
		var err error
		r, err = zlib.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case 3:
		r = io.NopCloser(bytes.NewReader(b))
	case 4:
		return decompressLZ4BlockStream(b)
	default:
		return nil, fmt.Errorf("unsupported Anvil compression type %d", kind)
	}
	defer r.Close()
	data, err := io.ReadAll(io.LimitReader(r, maxChunkBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxChunkBytes {
		return nil, errors.New("chunk NBT exceeds 32 MiB limit")
	}
	return data, nil
}

// WriteChunk rewrites a region atomically while preserving every other raw
// chunk record and timestamp. Unknown NBT fields inside the edited chunk are
// preserved by the nbt tree. The caller must not concurrently edit the same
// region from another process.
func WriteChunk(dir string, cx, cz int32, name string, root nbt.Tag) error {
	payload, err := encodeChunkRecord(name, root)
	if err != nil {
		return err
	}
	return writeChunkRecord(dir, cx, cz, payload)
}

var chunkCompressors = sync.Pool{New: func() any { return zlib.NewWriter(io.Discard) }}

func encodeChunkRecord(name string, root nbt.Tag) ([]byte, error) {
	data, err := nbt.Marshal(name, root)
	if err != nil {
		return nil, err
	}
	var compressed bytes.Buffer
	zw := chunkCompressors.Get().(*zlib.Writer)
	zw.Reset(&compressed)
	defer func() {
		zw.Reset(io.Discard) // do not retain a previous chunk's buffer
		chunkCompressors.Put(zw)
	}()
	if _, err = zw.Write(data); err != nil {
		zw.Close()
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}
	payload := compressed.Bytes()
	if len(payload)+5 > 255*sector {
		return nil, errors.New("compressed chunk requires external .mcc storage")
	}
	return payload, nil
}

func writeChunkRecord(dir string, cx, cz int32, payload []byte) error {
	path := regionPath(dir, cx, cz)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	old, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		old = make([]byte, 2*sector)
	} else if err != nil {
		return err
	}
	if len(old) < 2*sector || len(old) > maxRegionBytes {
		return errors.New("invalid existing region size")
	}
	idx := chunkIndex(cx, cz)
	newRegion := make([]byte, 2*sector)
	copy(newRegion[sector:2*sector], old[sector:2*sector])
	for i := 0; i < 1024; i++ {
		if i == idx {
			continue
		}
		loc := binary.BigEndian.Uint32(old[i*4 : i*4+4])
		if loc == 0 {
			continue
		}
		off, count := int(loc>>8), int(loc&255)
		if off < 2 || count < 1 || off*sector+5 > len(old) {
			return fmt.Errorf("invalid existing chunk entry %d", i)
		}
		newOff := len(newRegion) / sector
		if newOff > 0xffffff {
			return errors.New("region offset overflow")
		}
		raw := make([]byte, count*sector)
		copy(raw, old[off*sector:min((off+count)*sector, len(old))])
		newRegion = append(newRegion, raw...)
		binary.BigEndian.PutUint32(newRegion[i*4:i*4+4], uint32(newOff<<8|count))
	}
	newOff := len(newRegion) / sector
	count := (len(payload) + 5 + sector - 1) / sector
	if newOff > 0xffffff || len(newRegion)+count*sector > maxRegionBytes {
		return errors.New("region exceeds size limit")
	}
	record := make([]byte, count*sector)
	binary.BigEndian.PutUint32(record[:4], uint32(len(payload)+1))
	record[4] = 2
	copy(record[5:], payload)
	newRegion = append(newRegion, record...)
	binary.BigEndian.PutUint32(newRegion[idx*4:idx*4+4], uint32(newOff<<8|count))
	binary.BigEndian.PutUint32(newRegion[sector+idx*4:sector+idx*4+4], uint32(time.Now().Unix()))
	tmp, err := os.CreateTemp(dir, "atlas-region-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(newRegion); err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
