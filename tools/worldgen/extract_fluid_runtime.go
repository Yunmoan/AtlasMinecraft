//go:build ignore

// Read a flushed official save without invoking Atlas generation or simulation.
package main

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type savedTick struct {
	Type     string `json:"type"`
	X        int32  `json:"x"`
	Y        int32  `json:"y"`
	Z        int32  `json:"z"`
	Delay    int32  `json:"delay"`
	Priority int32  `json:"priority"`
}
type chunkSnapshot struct {
	X               int32       `json:"x"`
	Z               int32       `json:"z"`
	Blocks          []byte      `json:"blocks"`
	Ticks           []savedTick `json:"ticks"`
	BlockTicks      []savedTick `json:"block_ticks"`
	BlockEntities   []byte      `json:"block_entities"`
	FallingEntities []byte      `json:"falling_entities"`
}

func main() {
	if len(os.Args) != 3 {
		panic("expected official world directory and current game time")
	}
	now, err := strconv.ParseInt(os.Args[2], 10, 64)
	must(err)
	var result []chunkSnapshot
	for z := int32(1); z <= 7; z++ {
		for x := int32(1); x <= 4; x++ {
			_, tag, exists, err := anvil.ReadChunk(filepath.Join(os.Args[1], "region"), x, z)
			must(err)
			if !exists {
				panic(fmt.Sprintf("missing chunk %d,%d", x, z))
			}
			chunk, _ := tag.Object()
			status, _ := chunk.Text("Status")
			version, _ := chunk.Number("DataVersion")
			if status != "minecraft:full" || version != 3955 {
				panic("unfinished or wrong-version reference chunk")
			}
			lastUpdate, ok := chunk.Number("LastUpdate")
			if !ok || lastUpdate != now {
				panic(fmt.Sprintf("stale native chunk %d,%d: LastUpdate %d, expected %d", x, z, lastUpdate, now))
			}
			out := chunkSnapshot{X: x, Z: z, Blocks: make([]byte, 16*16*16*2), Ticks: []savedTick{}}
			entities := canonicalNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"block_entities": chunk["block_entities"]}})
			out.BlockEntities = append([]byte{nbt.Compound, 0, 0}, entities[1:]...)
			// Entity UUIDs are independent native randomness. Retain every other
			// falling-entity field and native list order; unrelated entity families
			// (including plant drops) are outside this fixture's comparison scope.
			falling := []nbt.Tag{}
			_, entityTag, entityExists, err := anvil.ReadChunk(filepath.Join(os.Args[1], "entities"), x, z)
			must(err)
			if entityExists {
				entityRoot, _ := entityTag.Object()
				list, ok := entityRoot["Entities"].Elements()
				if !ok {
					panic("missing native entity list")
				}
				for _, entry := range list {
					data, _ := entry.Object()
					id, _ := data.Text("id")
					if id == "minecraft:falling_block" {
						delete(data, "UUID")
						falling = append(falling, entry)
					}
				}
			}
			encoded := canonicalNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"falling_entities": {Type: nbt.List, ListType: nbt.Compound, Value: falling}}})
			out.FallingEntities = append([]byte{nbt.Compound, 0, 0}, encoded[1:]...)
			sections, _ := chunk["sections"].Elements()
			for _, st := range sections {
				section, _ := st.Object()
				sy, _ := section.Number("Y")
				if sy != 3 && sy != 4 {
					continue
				}
				states, _ := section.Object("block_states")
				palette, _ := states["palette"].Elements()
				if len(palette) == 0 {
					panic("missing block palette")
				}
				ids := make([]uint16, len(palette))
				for i, entry := range palette {
					state, _ := entry.Object()
					name, _ := state.Text("Name")
					props, _ := state.Object("Properties")
					properties := map[string]string{}
					for k := range props {
						properties[k], _ = props.Text(k)
					}
					id, err := blockstate.BlockStateID(name, properties)
					must(err)
					if id > 65535 {
						panic("state overflow")
					}
					ids[i] = uint16(id)
				}
				longs, _ := states["data"].Longs()
				width := max(4, bits.Len(uint(len(palette)-1)))
				perLong := 64 / width
				if len(palette) > 1 && len(longs) < (4096+perLong-1)/perLong {
					panic("truncated palette data")
				}
				for i := 0; i < 4096; i++ {
					y := int(sy)*16 + i/256
					if y < 60 || y >= 76 {
						continue
					}
					pi := 0
					if len(palette) > 1 {
						pi = int(uint64(longs[i/perLong]) >> ((i % perLong) * width) & ((1 << width) - 1))
					}
					if pi >= len(ids) {
						panic("invalid palette index")
					}
					binary.BigEndian.PutUint16(out.Blocks[((y-60)*256+i%256)*2:], ids[pi])
				}
			}
			for _, queue := range []struct {
				name string
				out  *[]savedTick
			}{{"fluid_ticks", &out.Ticks}, {"block_ticks", &out.BlockTicks}} {
				ticks, ok := chunk[queue.name].Elements()
				if !ok {
					panic("missing " + queue.name)
				}
				*queue.out = []savedTick{}
				for _, entry := range ticks {
					tick, _ := entry.Object()
					name, _ := tick.Text("i")
					s := savedTick{Type: name}
					for key, dst := range map[string]*int32{"x": &s.X, "y": &s.Y, "z": &s.Z, "t": &s.Delay, "p": &s.Priority} {
						v, ok := tick.Number(key)
						if !ok {
							panic("invalid tick")
						}
						*dst = int32(v)
					}
					*queue.out = append(*queue.out, s)
				}
			}
			result = append(result, out)
		}
	}
	must(json.NewEncoder(os.Stdout).Encode(result))
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Compound field order is not NBT semantics, and the ordinary Go serializer
// iterates maps. Canonicalize only those fields for reproducible oracle bytes;
// lists (including block entities, items and tick queues) retain native order.
func canonicalNBT(tag nbt.Tag) []byte {
	if tag.Type == nbt.Compound {
		out := []byte{nbt.Compound}
		object, _ := tag.Object()
		names := make([]string, 0, len(object))
		for name := range object {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			value := canonicalNBT(object[name])
			nameTag, err := nbt.MarshalNetwork(nbt.Tag{Type: nbt.String, Value: name})
			must(err)
			out = append(out, value[0])
			out = append(out, nameTag[1:]...)
			out = append(out, value[1:]...)
		}
		return append(out, nbt.End)
	}
	if tag.Type == nbt.List {
		items, _ := tag.Elements()
		out := make([]byte, 6)
		out[0] = nbt.List
		out[1] = tag.ListType
		binary.BigEndian.PutUint32(out[2:], uint32(len(items)))
		for _, item := range items {
			if item.Type != tag.ListType {
				panic("mixed native list")
			}
			out = append(out, canonicalNBT(item)[1:]...)
		}
		return out
	}
	out, err := nbt.MarshalNetwork(tag)
	must(err)
	return out
}
