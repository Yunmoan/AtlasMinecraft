package world

import (
	"atlas/internal/light"
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/bits"
)

type lightChanged struct {
	cx, cz int32
	data   *light.Data
}

func denseLightStates(ch *Chunk) []uint16 {
	states := make([]uint16, int(ch.maxY-ch.minY)*256)
	for y := ch.minY; y < ch.maxY; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				states[int(y-ch.minY)*256+z*16+x] = uint16(ch.GetBlock(x, z, y))
			}
		}
	}
	return states
}

func lightBlockHash(ch *Chunk) []byte {
	dense := denseLightStates(ch)
	encoded := make([]byte, len(dense)*2)
	for i, state := range dense {
		binary.LittleEndian.PutUint16(encoded[i*2:], state)
	}
	sum := sha256.Sum256(encoded)
	return sum[:]
}

func readAnvilLight(root nbt.Object, sections []nbt.Tag, ch *Chunk) (*light.Data, error) {
	on, _ := root.Number("isLightOn")
	marker, _ := root.Object("AtlasLight")
	version, _ := marker.Number("Version")
	stale, _ := marker.Number("Stale")
	data := &light.Data{Sky: make([][]byte, len(ch.sect)+2), Block: make([][]byte, len(ch.sect)+2), Complete: on != 0}
	hash, _ := marker["BlocksSHA256"].Value.([]byte)
	data.Preserved = version == 2 && stale == 0 && bytes.Equal(hash, lightBlockHash(ch))
	for _, section := range sections {
		o, _ := section.Object()
		y, _ := o.Number("Y")
		index := int(y) - (int(ch.minY)/16 - 1)
		if index < 0 || index >= len(data.Sky) {
			continue
		}
		for _, pair := range []struct {
			name   string
			layers [][]byte
		}{{"SkyLight", data.Sky}, {"BlockLight", data.Block}} {
			if tag, exists := o[pair.name]; exists {
				bytes, ok := tag.Value.([]byte)
				if tag.Type != nbt.ByteArray || !ok || len(bytes) != 2048 {
					return nil, fmt.Errorf("section %d invalid %s array", y, pair.name)
				}
				pair.layers[index] = append([]byte(nil), bytes...)
			}
		}
	}
	if on == 0 && !data.Preserved {
		return nil, nil
	}
	return data, nil
}

func (c *Chunk) writeAnvilLight(root nbt.Object) {
	if c.lights == nil {
		root["isLightOn"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		return
	}
	sections, _ := root["sections"].Elements()
	byY := make(map[int]nbt.Object)
	for _, section := range sections {
		o, _ := section.Object()
		y, _ := o.Number("Y")
		byY[int(y)] = o
	}
	for index := range c.lights.Sky {
		y := int(c.minY)/16 - 1 + index
		o := byY[y]
		if o == nil {
			if c.lights.Sky[index] == nil && c.lights.Block[index] == nil {
				continue
			}
			o = nbt.Object{"Y": {Type: nbt.Byte, Value: int8(y)}}
			sections = append(sections, nbt.Tag{Type: nbt.Compound, Value: o})
		}
		for _, pair := range []struct {
			name   string
			layers [][]byte
		}{{"SkyLight", c.lights.Sky}, {"BlockLight", c.lights.Block}} {
			if pair.layers[index] == nil {
				delete(o, pair.name)
			} else {
				o[pair.name] = nbt.Tag{Type: nbt.ByteArray, Value: pair.layers[index]}
			}
		}
	}
	root["sections"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: sections}
	complete := int8(0)
	if c.lights.Complete && !c.lightStale {
		complete = 1
	}
	root["isLightOn"] = nbt.Tag{Type: nbt.Byte, Value: complete}
	stale := int8(0)
	if c.lightStale {
		stale = 1
	}
	root["AtlasLight"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"Version": {Type: nbt.Int, Value: int32(2)}, "Stale": {Type: nbt.Byte, Value: stale}, "BlocksSHA256": {Type: nbt.ByteArray, Value: lightBlockHash(c)}}}
}

func writeLightData(w *proto.Writer, data *light.Data) {
	var skyMask, blockMask int64
	for i := range data.Sky {
		if data.Sky[i] != nil {
			skyMask |= 1 << i
		}
		if data.Block[i] != nil {
			blockMask |= 1 << i
		}
	}
	all := int64(1)<<len(data.Sky) - 1
	w.VarInt(1).I64(skyMask).VarInt(1).I64(blockMask).VarInt(1).I64(all &^ skyMask).VarInt(1).I64(all &^ blockMask)
	for _, item := range []struct {
		mask   int64
		layers [][]byte
	}{{skyMask, data.Sky}, {blockMask, data.Block}} {
		w.VarInt(int32(bits.OnesCount64(uint64(item.mask))))
		for i, layer := range item.layers {
			if item.mask&(1<<i) != 0 {
				w.VarInt(int32(len(layer))).Raw(layer)
			}
		}
	}
}

func detachedLight(ch *Chunk) *light.Data {
	engine, err := light.New(int(ch.minY), int(ch.maxY-ch.minY), ch.hasSky)
	if err == nil {
		if _, err := engine.Install(light.Position{X: ch.cx, Z: ch.cz}, denseLightStates(ch), ch.lights); err == nil {
			return engine.Snapshot(light.Position{X: ch.cx, Z: ch.cz})
		}
	}
	// Embedded rules are validated before a real world load. A malformed
	// stand-alone test chunk has no valid lighting rather than synthetic light.
	return &light.Data{Sky: make([][]byte, len(ch.sect)+2), Block: make([][]byte, len(ch.sect)+2)}
}

func (w *World) notifyLight(updates map[light.Position]*light.Data, except light.Position) {
	for position, data := range updates {
		if position == except {
			continue
		}
		_ = w.post(context.Background(), position.X, position.Z, lightChanged{position.X, position.Z, data})
	}
}

func (r *Region) applyLight(ch *Chunk, data *light.Data, broadcast bool) {
	if data == nil || ch.lights != nil && data.Revision <= ch.lights.Revision {
		return
	}
	ch.lights = data
	ch.lightStale = false
	ch.dirty = true
	if broadcast && r.world.hub != nil {
		writer := proto.NewWriter().VarInt(ch.cx).VarInt(ch.cz)
		writeLightData(writer, data)
		for _, sink := range r.world.hub.ViewersOf(ch.cx*16, ch.cz*16) {
			if chunkDelivered(sink, ch.cx, ch.cz) {
				sink.SendPacket(proto.SUpdateLight, writer.Payload())
			}
		}
	}
}
