package world

import (
	"atlas/internal/light"
	"container/list"
)

const chunkPacketCacheBytes = 2 << 20
const chunkPacketCacheEntries = 64

type cachedChunkPacket struct {
	chunk    *Chunk
	revision uint64
	lights   *light.Data
	payload  []byte
}

// Owned by a region actor. Both the byte and entry limits are independent of
// how much terrain has been explored. Queued packets retain immutable buffers
// after eviction or an edit; a later request encodes the new revision.
type chunkPacketCache struct {
	entries map[*Chunk]*list.Element
	recent  list.List
	bytes   int
}

func (c *chunkPacketCache) packet(ch *Chunk) []byte {
	if element := c.entries[ch]; element != nil {
		entry := element.Value.(*cachedChunkPacket)
		if entry.revision == ch.blockRevision && entry.lights == ch.lights {
			c.recent.MoveToFront(element)
			return entry.payload
		}
		c.remove(element)
	}
	payload := buildChunkDataPacket(ch.cx, ch.cz, ch)
	if len(payload) > chunkPacketCacheBytes {
		return payload
	}
	if c.entries == nil {
		c.entries = make(map[*Chunk]*list.Element)
	}
	for c.bytes+len(payload) > chunkPacketCacheBytes || len(c.entries) >= chunkPacketCacheEntries {
		c.remove(c.recent.Back())
	}
	entry := &cachedChunkPacket{chunk: ch, revision: ch.blockRevision, lights: ch.lights, payload: payload}
	c.entries[ch] = c.recent.PushFront(entry)
	c.bytes += len(payload)
	return payload
}

func (c *chunkPacketCache) remove(element *list.Element) {
	entry := element.Value.(*cachedChunkPacket)
	delete(c.entries, entry.chunk)
	c.bytes -= len(entry.payload)
	c.recent.Remove(element)
}
