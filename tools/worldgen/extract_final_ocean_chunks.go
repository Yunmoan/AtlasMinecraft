//go:build ignore

// Extract complete block-state arrays from a stopped official 1.21.1 save.
package main

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/bits"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		panic("expected stopped official world directory followed by extraction flags")
	}
	flags := flag.NewFlagSet("extract", flag.ExitOnError)
	preset := flags.String("preset", "ocean", "ocean, ocean-ruins or shipwrecks")
	fluids := flags.Bool("fluid-ticks", false, "extract fluid ticks")
	blocks := flags.Bool("block-ticks", false, "extract block ticks")
	entities := flags.Bool("block-entities", false, "extract block entities")
	must(flags.Parse(os.Args[2:]))
	expectedSeed, expectedStructures := int64(4511451270791457643), int64(0)
	positions := [][2]int32{{0, 10}, {-3, 44}}
	if *preset == "ocean-ruins" {
		expectedSeed, expectedStructures = 1234, 1
		positions = [][2]int32{{-30, -56}, {62, -60}}
	} else if *preset == "shipwrecks" {
		expectedSeed, expectedStructures = 1234, 1
		positions = [][2]int32{{4, 12}, {72, -62}}
	} else if *preset != "ocean" {
		panic("unknown fixture preset")
	}
	dir := os.Args[1]
	f, err := os.Open(filepath.Join(dir, "level.dat"))
	must(err)
	r, err := gzip.NewReader(f)
	must(err)
	b, err := io.ReadAll(r)
	must(err)
	must(r.Close())
	must(f.Close())
	_, tag, err := nbt.Parse(b)
	must(err)
	root, _ := tag.Object()
	data, _ := root.Object("Data")
	settings, _ := data.Object("WorldGenSettings")
	seed, _ := settings.Number("seed")
	version, _ := data.Number("DataVersion")
	structures, _ := settings.Number("generate_features")
	if seed != expectedSeed || version != 3955 || structures != expectedStructures {
		panic("world does not match pinned 1.21.1 preset")
	}
	for _, p := range positions {
		_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), p[0], p[1])
		must(err)
		if !exists {
			panic(fmt.Sprintf("missing original chunk %v", p))
		}
		chunk, _ := tag.Object()
		status, _ := chunk.Text("Status")
		cv, _ := chunk.Number("DataVersion")
		if status != "minecraft:full" || cv != 3955 {
			panic("unfinished or wrong-version chunk")
		}
		if *preset == "shipwrecks" {
			name := "minecraft:shipwreck"
			if p[0] == 72 {
				name = "minecraft:shipwreck_beached"
			}
			structures, _ := chunk.Object("structures")
			starts, _ := structures.Object("starts")
			start, _ := starts.Object(name)
			id, _ := start.Text("id")
			if id != name {
				panic(fmt.Sprintf("original chunk %v has no %s start: %q", p, name, id))
			}
		}
		if *entities {
			entries, ok := chunk["block_entities"].Elements()
			if !ok {
				panic("missing original block entities")
			}
			byPosition := map[string]nbt.Object{}
			for _, entry := range entries {
				o, ok := entry.Object()
				if !ok {
					panic("invalid original block entity")
				}
				x, _ := o.Number("x")
				y, _ := o.Number("y")
				z, _ := o.Number("z")
				key := fmt.Sprintf("%d,%d,%d", x, y, z)
				if _, exists := byPosition[key]; exists {
					panic("duplicate original block entity")
				}
				byPosition[key] = o
			}
			// JSON sorts compound keys and retains numeric NBT types. Only the
			// outer position index replaces the unordered block-entity container;
			// nested lists/arrays remain in their original order.
			payload, err := json.Marshal(byPosition)
			must(err)
			fmt.Printf("native_block_entities %d %d %d %s\n", seed, p[0], p[1], base64.StdEncoding.EncodeToString(payload))
			continue
		}
		if *fluids || *blocks {
			kind := "fluid"
			if *blocks {
				kind = "block"
			}
			ticks, ok := chunk[kind+"_ticks"].Elements()
			if !ok {
				panic("missing native " + kind + " tick list")
			}
			fmt.Printf("native_%s_chunk %d %d %d %d\n", kind, seed, p[0], p[1], len(ticks))
			for _, entry := range ticks {
				tick, _ := entry.Object()
				fluid, _ := tick.Text("i")
				fmt.Printf("native_%s_tick %s", kind, fluid)
				for _, key := range []string{"x", "y", "z", "t", "p"} {
					value, ok := tick.Number(key)
					if !ok {
						panic("invalid native fluid tick")
					}
					fmt.Printf(" %d", value)
				}
				fmt.Println()
			}
			continue
		}
		out := make([]byte, 384*256*2)
		sections, _ := chunk["sections"].Elements()
		for _, st := range sections {
			section, _ := st.Object()
			sy, _ := section.Number("Y")
			if sy < -4 || sy > 19 {
				continue
			}
			states, _ := section.Object("block_states")
			palette, _ := states["palette"].Elements()
			if len(palette) == 0 {
				continue
			}
			ids := make([]uint16, len(palette))
			for i, entry := range palette {
				state, _ := entry.Object()
				name, _ := state.Text("Name")
				props, _ := state.Object("Properties")
				properties := make(map[string]string, len(props))
				for k := range props {
					properties[k], _ = props.Text(k)
				}
				id, err := blockstate.BlockStateID(name, properties)
				must(err)
				if id > 65535 {
					panic("block-state ID overflows fixture")
				}
				ids[i] = uint16(id)
			}
			longs, _ := states["data"].Longs()
			width := max(4, bits.Len(uint(len(palette)-1)))
			perLong := 64 / width
			if len(palette) > 1 && len(longs) < (4096+perLong-1)/perLong {
				panic("truncated block-state array")
			}
			for i := 0; i < 4096; i++ {
				pi := 0
				if len(palette) > 1 {
					pi = int(uint64(longs[i/perLong]) >> ((i % perLong) * width) & ((1 << width) - 1))
				}
				if pi >= len(ids) {
					panic("invalid palette index")
				}
				binary.BigEndian.PutUint16(out[(int(sy+4)*4096+i)*2:], ids[pi])
			}
		}
		var compressed bytes.Buffer
		w := gzip.NewWriter(&compressed)
		_, err = w.Write(out)
		must(err)
		must(w.Close())
		fmt.Printf("final_chunk %d %d %d %s\n", seed, p[0], p[1], base64.StdEncoding.EncodeToString(compressed.Bytes()))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
