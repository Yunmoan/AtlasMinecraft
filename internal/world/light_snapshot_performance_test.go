package world

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"
)

func TestLightSectionExpansionAndHashMatchBlockReads(t *testing.T) {
	for _, bounds := range [][2]int32{{-64, 320}, {0, 256}} {
		chunk := newChunkBounds(-17, 23, bounds[0], bounds[1], true)
		chunk.sect[0] = newUniformSection(StateStone)
		chunk.sect[1] = newUniformSection(StateAir)
		chunk.sect[2] = newUniformSection(StateCaveAir)
		chunk.sect[3] = newUniformSection(StateVoidAir)
		for si := 4; si < min(11, len(chunk.sect)); si++ {
			section := newUniformSection(StateAir)
			// Force 4..8-bit palettes, then direct encoding; non-divisible
			// packed tails and state IDs above 16383 are both exercised.
			for i := 0; i < SectionVolume; i++ {
				section.set(i, uint32(i%(1<<uint(si-1)))+26000)
			}
			chunk.sect[si] = section
		}
		got := denseLightStates(chunk)
		encoded := make([]byte, len(got)*2)
		for i, state := range got {
			want := chunk.GetBlock(i&15, i>>4&15, chunk.minY+int32(i>>8))
			if uint32(state) != want {
				t.Fatalf("bounds %v block %d: %d/%d", bounds, i, state, want)
			}
			binary.LittleEndian.PutUint16(encoded[i*2:], uint16(want))
		}
		want := sha256.Sum256(encoded)
		if !bytes.Equal(lightBlockHash(chunk), want[:]) {
			t.Fatalf("bounds %v hash differs from original dense little-endian stream", bounds)
		}
	}
}
