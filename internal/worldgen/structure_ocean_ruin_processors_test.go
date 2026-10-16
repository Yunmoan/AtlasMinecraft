package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestOceanRuinProcessorsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_ocean_ruin_processors_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 11 || f[0] != "oceanprocess" {
			t.Fatalf("invalid ocean processor row %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		integrity, err := strconv.ParseFloat(f[3], 32)
		if err != nil {
			t.Fatal(err)
		}
		template, err := LoadOceanRuinTemplate(f[1])
		if err != nil {
			t.Fatal(err)
		}
		blocks, err := ProcessOceanRuinTemplateBlocks(template, FeaturePosition{int32(parse(6)), int32(parse(7)), int32(parse(8))}, TemplateRotation(parse(2)), float32(integrity), parse(4) != 0, parse(5))
		if err != nil {
			t.Fatal(err)
		}
		var text strings.Builder
		for _, block := range blocks {
			id, err := blockstate.BlockStateID(block.State.Name, block.State.Properties)
			if err != nil {
				t.Fatal(err)
			}
			tag := "none"
			if block.NBT != nil {
				tag = iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: block.NBT})
			}
			fmt.Fprintf(&text, "%d %d %d %d %s\n", block.Position.X, block.Position.Y, block.Position.Z, id, tag)
		}
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(text.String())))
		if len(blocks) != int(parse(9)) || hash != f[10] {
			t.Fatalf("%s: got %d/%s", scanner.Text(), len(blocks), hash)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 13824 {
		t.Fatalf("incomplete ocean processor cases %d", count)
	}
	t.Logf("matched %d original BlockRot/ignore/capped archaeology/NBT cases", count)
}
