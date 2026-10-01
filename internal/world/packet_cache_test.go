package world

import (
	"atlas/internal/light"
	"bytes"
	"testing"
)

func TestCachedChunkPacketTracksBlocksHeightmapsAndLighting(t *testing.T) {
	ch := NewChunk(-1, 0)
	ch.lights = &light.Data{Sky: make([][]byte, NumSections+2), Block: make([][]byte, NumSections+2)}
	var cache chunkPacketCache
	first := cache.packet(ch)
	saved := append([]byte(nil), first...)
	if again := cache.packet(ch); &first[0] != &again[0] {
		t.Fatal("unchanged chunk was encoded again")
	}
	for _, state := range []uint32{StateStone, StateAir} {
		ch.SetBlock(15, 8, 75, state)
		if got := cache.packet(ch); !bytes.Equal(got, buildChunkDataPacket(-1, 0, ch)) {
			t.Fatal("cached blocks or heightmap are stale")
		}
	}
	ch.lights = &light.Data{Sky: make([][]byte, NumSections+2), Block: make([][]byte, NumSections+2), Revision: 2}
	ch.lights.Block[3] = bytes.Repeat([]byte{0x77}, 2048)
	if got := cache.packet(ch); !bytes.Equal(got, buildChunkDataPacket(-1, 0, ch)) || bytes.Equal(got, first) {
		t.Fatal("cached light data is stale")
	}
	if !bytes.Equal(first, saved) {
		t.Fatal("a queued packet changed after edits")
	}
	for i := int32(0); i < 100; i++ {
		other := NewChunk(i, 1)
		other.lights = ch.lights
		cache.packet(other)
		if cache.bytes > chunkPacketCacheBytes || len(cache.entries) > chunkPacketCacheEntries {
			t.Fatal("packet cache exceeded its bound")
		}
	}
	if !bytes.Equal(cache.packet(ch), buildChunkDataPacket(-1, 0, ch)) {
		t.Fatal("eviction lost current chunk data")
	}
}

func TestLoadChunkReturnsOwnedPacketBytes(t *testing.T) {
	w, _ := mechanicsWorld(t)
	first, err := w.LoadChunk(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	wanted := append([]byte(nil), first...)
	clear(first)
	again, err := w.LoadChunk(0, 0)
	if err != nil || !bytes.Equal(again, wanted) {
		t.Fatalf("caller poisoned cached packet: %v", err)
	}
}
