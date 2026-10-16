package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIglooOriginalTemplatesLoad(t *testing.T) {
	for _, name := range []string{"top", "middle", "bottom"} {
		template, err := LoadIglooTemplate(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("igloo %s: size %v, palettes %d, blocks %d, entities %d", name, template.Size, len(template.Palettes), len(template.Palettes[0]), len(template.Entities))
	}
}

func TestIglooStartsAndTemplatesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_igloo_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type key struct {
		seed int64
		x, z int32
	}
	starts := make(map[key]*StructureStart)
	scanner := bufio.NewScanner(file)
	count, pieceCount := 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		switch f[0] {
		case "template":
			template, err := LoadIglooTemplate(f[1])
			if err != nil {
				t.Fatal(err)
			}
			if template.Size != (FeaturePosition{int32(parse(2)), int32(parse(3)), int32(parse(4))}) || len(template.Palettes[0]) != int(parse(5)) || len(template.Entities) != int(parse(6)) {
				t.Fatalf("%s: got %+v", scanner.Text(), template)
			}
		case "template_block":
			template, err := LoadIglooTemplate(f[1])
			if err != nil {
				t.Fatal(err)
			}
			block := template.Palettes[0][parse(2)]
			id, err := blockstate.BlockStateID(block.State.Name, block.State.Properties)
			if err != nil {
				t.Fatal(err)
			}
			name := "none"
			if block.NBT != nil {
				name, _ = block.NBT.Text("id")
			}
			if block.Position != (FeaturePosition{int32(parse(3)), int32(parse(4)), int32(parse(5))}) || id != uint32(parse(6)) || name != f[7] {
				t.Fatalf("%s: got %v state %d nbt %s", scanner.Text(), block.Position, id, name)
			}
		case "height":
			k := key{parse(1), int32(parse(2)), int32(parse(3))}
			start, err := GenerateIglooStructureStart(k.seed, k.x, k.z, func(x, y, z int32) (string, error) {
				if x != (k.x*16+8)>>2 || z != (k.z*16+8)>>2 || y != int32(parse(4))>>2 {
					return "", fmt.Errorf("incorrect igloo biome coordinates")
				}
				return "minecraft:snowy_plains", nil
			}, func(x, z int32) (int32, error) {
				if x != k.x*16+8 || z != k.z*16+8 {
					return 0, fmt.Errorf("incorrect igloo base height coordinates")
				}
				return int32(parse(4)), nil
			})
			if err != nil || start == nil {
				t.Fatalf("%s: got start %v (%v)", scanner.Text(), start, err)
			}
			starts[k] = start
		case "start":
			k := key{parse(1), int32(parse(2)), int32(parse(3))}
			if len(starts[k].Pieces) != int(parse(4)) {
				t.Fatalf("%s: got %d pieces", scanner.Text(), len(starts[k].Pieces))
			}
		case "piece":
			k := key{parse(1), int32(parse(2)), int32(parse(3))}
			piece := starts[k].Pieces[parse(4)]
			tag, err := piece.NBT()
			if err != nil {
				t.Fatal(err)
			}
			got := fmt.Sprintf("piece %d %d %d %d", k.seed, k.x, k.z, parse(4))
			for _, field := range []string{"id", "Template", "Rot"} {
				value, _ := tag.Text(field)
				got += " " + value
			}
			for _, field := range []string{"O", "GD", "TPX", "TPY", "TPZ"} {
				value, _ := tag.Number(field)
				got += fmt.Sprint(" ", value)
			}
			for _, value := range tag["BB"].Value.([]int32) {
				got += fmt.Sprint(" ", value)
			}
			if got != scanner.Text() {
				t.Fatalf("got %s, want %s", got, scanner.Text())
			}
			pieceCount++
		default:
			t.Fatalf("unknown igloo oracle vector %q", scanner.Text())
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 713 || len(starts) != 48 || pieceCount != 203 {
		t.Fatalf("got %d vectors/%d starts/%d pieces", count, len(starts), pieceCount)
	}
	t.Logf("matched %d original template order/start/NBT vectors with %d pieces", count, pieceCount)
}
