//go:build ignore

// Extracts bubble coordinates from an actual official-server full Anvil save.
// Usage: go run docs/worldgen/extract_native_bubble_columns.go <world-dir>
package main

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"compress/gzip"
	"fmt"
	"io"
	"math/bits"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	if len(os.Args) != 2 {
		panic("expected official 1.21.1 world directory")
	}
	dir := os.Args[1]
	file, err := os.Open(filepath.Join(dir, "level.dat"))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	levelBytes, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	_, levelTag, err := nbt.Parse(levelBytes)
	if err != nil {
		panic(err)
	}
	level, _ := levelTag.Object()
	data, _ := level.Object("Data")
	version, _ := data.Number("DataVersion")
	settings, _ := data.Object("WorldGenSettings")
	seed, _ := settings.Number("seed")
	if version != 3955 || seed != 4511451270791457643 {
		panic("expected the independent Java 1.21.1 user-seed save")
	}
	for _, p := range [][2]int32{{0, 10}, {-3, 44}} {
		_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), p[0], p[1])
		if err != nil || !exists {
			panic(fmt.Sprintf("missing reference chunk %v: %v", p, err))
		}
		chunk, _ := tag.Object()
		status, _ := chunk.Text("Status")
		if status != "minecraft:full" {
			panic("unfinished official reference chunk")
		}
		sections, _ := chunk["sections"].Elements()
		var rows []string
		for _, sectionTag := range sections {
			section, _ := sectionTag.Object()
			sy, _ := section.Number("Y")
			states, _ := section.Object("block_states")
			palette, _ := states["palette"].Elements()
			if len(palette) == 0 {
				continue
			}
			longs, _ := states["data"].Longs()
			width := max(4, bits.Len(uint(len(palette)-1)))
			perLong := 64 / width
			for i := 0; i < 4096; i++ {
				pi := 0
				if len(palette) > 1 {
					pi = int((uint64(longs[i/perLong]) >> ((i % perLong) * width)) & ((1 << width) - 1))
				}
				state, _ := palette[pi].Object()
				name, _ := state.Text("Name")
				if name != "minecraft:bubble_column" {
					continue
				}
				properties, _ := state.Object("Properties")
				drag, _ := properties.Text("drag")
				rows = append(rows, fmt.Sprintf("native_bubble_column %d %d %d %d %d %d %s", seed, p[0], p[1], p[0]*16+int32(i&15), sy*16+int64(i/256), p[1]*16+int32((i>>4)&15), drag))
			}
		}
		sort.Strings(rows)
		fmt.Printf("native_bubble_chunk %d %d %d %d\n", seed, p[0], p[1], len(rows))
		for _, row := range rows {
			fmt.Println(row)
		}
	}
}
