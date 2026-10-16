package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRuinedPortalStartsAndBaseColumnsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_ruined_portal_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type generatorKey struct {
		dimension string
		seed      int64
	}
	type columnKey struct {
		generatorKey
		x, z int32
	}
	generators := make(map[generatorKey]StructureBaseNoiseColumnAt)
	columns := make(map[columnKey]*StructureBaseNoiseColumn)
	getColumn := func(key columnKey) (*StructureBaseNoiseColumn, error) {
		if column := columns[key]; column != nil {
			return column, nil
		}
		generate := generators[key.generatorKey]
		if generate == nil {
			if key.dimension == "overworld" {
				generator, err := NewOverworldNoiseGenerator(key.seed)
				if err != nil {
					return nil, err
				}
				generate = generator.BaseStructureColumnAt
			} else {
				generator, err := NewDimensionNoiseGenerator(key.seed, key.dimension)
				if err != nil {
					return nil, err
				}
				generate = generator.BaseStructureColumnAt
			}
			generators[key.generatorKey] = generate
		}
		column, err := generate(key.x, key.z)
		if err == nil {
			columns[key] = column
		}
		return column, err
	}
	scanner := bufio.NewScanner(file)
	starts, columnCount, accepted := 0, 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 8 {
			t.Fatalf("invalid portal start/column row %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		switch f[0] {
		case "column":
			key := columnKey{generatorKey{f[1], parse(2)}, int32(parse(3)), int32(parse(4))}
			column, err := getColumn(key)
			if err != nil {
				t.Fatal(err)
			}
			want, err := base64.StdEncoding.DecodeString(f[7])
			if err != nil {
				t.Fatal(err)
			}
			if column.MinY != int32(parse(5)) || len(column.Blocks) != len(want) || len(want) != int(parse(6)) {
				t.Fatalf("invalid column shape %s", scanner.Text())
			}
			for index, material := range column.Blocks {
				if byte(material) != want[index] {
					t.Fatalf("column %s %d %d,%d at Y=%d: got %s want %s", key.dimension, key.seed, key.x, key.z, column.MinY+int32(index), material.BlockName(), NoiseMaterial(want[index]).BlockName())
				}
			}
			columnCount++
		case "start":
			dimension, minY := "overworld", int32(-64)
			if strings.HasSuffix(f[1], "nether") {
				dimension, minY = "nether", 0
			}
			key := generatorKey{dimension, parse(3)}
			sources := RuinedPortalStartSources{MinY: minY, BiomeAtQuart: func(x, y, z int32) (string, error) { return f[2], nil }, BaseColumnAt: func(x, z int32) (*StructureBaseNoiseColumn, error) { return getColumn(columnKey{key, x, z}) }}
			start, err := GenerateRuinedPortalStructureStart(key.seed, int32(parse(4)), int32(parse(5)), f[1], sources)
			if err != nil {
				t.Fatal(err)
			}
			var saved nbt.Object
			pieces := 0
			if start == nil {
				saved = nbt.Object{"id": {Type: nbt.String, Value: "INVALID"}}
			} else {
				saved, err = start.NBT()
				if err != nil {
					t.Fatal(err)
				}
				pieces = len(start.Pieces)
				accepted++
			}
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))))
			if pieces != int(parse(6)) || hash != f[7] {
				t.Fatalf("%s: got pieces %d hash %s; NBT %v", scanner.Text(), pieces, hash, saved)
			}
			starts++
		default:
			t.Fatalf("unknown portal row %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if starts != 720 || accepted != 384 || columnCount != 820 || len(columns) != 820 {
		t.Fatalf("got %d starts/%d accepted/%d columns/%d cached", starts, accepted, columnCount, len(columns))
	}
	t.Logf("matched %d original full starts (%d accepted) and %d full material columns", starts, accepted, columnCount)
}
