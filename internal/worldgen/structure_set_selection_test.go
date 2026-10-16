package worldgen

import (
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

func TestStructureSetSelectionAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_set_selection_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	placements, err := LoadStructurePlacementState(0)
	if err != nil {
		t.Fatal(err)
	}
	original := placements.Sets["minecraft:ruined_portals"]
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
	count, accepted, retried, allRejected := 0, 0, 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 9 || f[0] != "set" {
			t.Fatalf("invalid set selection row %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		set := original
		set.Structures = append([]StructureSetEntry(nil), original.Structures...)
		switch parse(1) {
		case 0:
		case 1:
			for index := range set.Structures {
				set.Structures[index].Weight = int32(index + 1)
			}
		case 2:
			set.Structures = set.Structures[:1]
		default:
			t.Fatalf("unknown selection variant %s", f[1])
		}
		key := generatorKey{f[2], parse(4)}
		cx, cz := int32(parse(5)), int32(parse(6))
		minY := int32(-64)
		if key.dimension == "nether" {
			minY = 0
		}
		sources := RuinedPortalStartSources{MinY: minY, BiomeAtQuart: func(x, y, z int32) (string, error) { return f[3], nil }, BaseColumnAt: func(x, z int32) (*StructureBaseNoiseColumn, error) { return getColumn(columnKey{key, x, z}) }}
		attempts := 0
		start, err := SelectStructureSetStart(key.seed, cx, cz, set, func(name string) (*StructureStart, error) {
			attempts++
			return GenerateRuinedPortalStructureStart(key.seed, cx, cz, name, sources)
		})
		if err != nil {
			t.Fatal(err)
		}
		name, hash := "INVALID", "none"
		if start != nil {
			accepted++
			name = start.Structure
			saved, err := start.NBT()
			if err != nil {
				t.Fatal(err)
			}
			hash = fmt.Sprintf("%x", sha256.Sum256([]byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))))
		} else {
			allRejected++
		}
		if attempts > 1 {
			retried++
		}
		if name != f[7] || hash != f[8] {
			t.Fatalf("%s: got %s/%s after %d attempts", scanner.Text(), name, hash, attempts)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 1728 || accepted == 0 || retried == 0 || allRejected == 0 {
		t.Fatalf("incomplete selection coverage %d/%d/%d/%d", count, accepted, retried, allRejected)
	}
	t.Logf("matched %d original createStructures selections (%d accepted, %d retries, %d all-rejected), %d cached original noise columns", count, accepted, retried, allRejected, len(columns))
}
