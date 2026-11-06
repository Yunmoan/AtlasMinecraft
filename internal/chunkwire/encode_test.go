package chunkwire

import (
	"atlas/internal/light"
	"atlas/internal/proto"
	"bytes"
	"testing"
)

func TestUniformSectionIncludesBothEmptyDataArrays(t *testing.T) {
	section := Section{NonAir: 4096, Palette: []uint32{1}}
	got := Sections([]Section{section})
	want := []byte{0x10, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("%x != %x", got, want)
	}
}

func TestColumnPreservesPreparedPayloads(t *testing.T) {
	column := Column{X: -1, Z: 2, Sections: []byte{1, 2, 3}, Heightmaps: []byte{10, 0}, BlockEntities: []byte{0}, Light: &light.Data{Sky: make([][]byte, 2), Block: make([][]byte, 2)}}
	payload := Encode(column)
	wantHeader := proto.NewWriter().I32(-1).I32(2).Raw(column.Heightmaps).VarInt(3).Raw(column.Sections).Raw(column.BlockEntities).Payload()
	if !bytes.HasPrefix(payload, wantHeader) || len(payload) == len(wantHeader) {
		t.Fatalf("invalid assembled column %x", payload)
	}
}
