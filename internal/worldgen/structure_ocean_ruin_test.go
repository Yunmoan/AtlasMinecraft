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

func TestOceanRuinStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_ocean_ruin_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	generators := make(map[int64]*OverworldNoiseGenerator)
	scanner := bufio.NewScanner(file)
	count, accepted, clusters := 0, 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 9 || f[0] != "oceanstart" {
			t.Fatalf("invalid ocean ruin row %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		seed, cx, cz := parse(3), int32(parse(4)), int32(parse(5))
		generator := generators[seed]
		if generator == nil {
			generator, err = NewOverworldNoiseGenerator(seed)
			if err != nil {
				t.Fatal(err)
			}
			generators[seed] = generator
		}
		cold := strings.HasSuffix(f[1], "cold")
		start, err := GenerateOceanRuinStructureStart(seed, cx, cz, cold, func(x, y, z int32) (string, error) {
			if x != (cx*16+8)>>2 || z != (cz*16+8)>>2 || y != int32(parse(6))>>2 {
				return "", fmt.Errorf("incorrect ocean ruin biome coordinates")
			}
			return f[2], nil
		}, func(x, z int32) (int32, error) {
			column, err := generator.BaseStructureColumnAt(x, z)
			if err != nil {
				return 0, err
			}
			height := column.OccupiedHeight(true)
			if height != int32(parse(6)) {
				return 0, fmt.Errorf("base ocean floor height %d want %s", height, f[6])
			}
			return height, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		var saved nbt.Object
		pieces := 0
		if start == nil {
			saved = nbt.Object{"id": {Type: nbt.String, Value: "INVALID"}}
		} else {
			accepted++
			pieces = len(start.Pieces)
			if pieces > 3 {
				clusters++
			}
			saved, err = start.NBT()
			if err != nil {
				t.Fatal(err)
			}
		}
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))))
		if pieces != int(parse(7)) || hash != f[8] {
			t.Fatalf("%s: got %d/%s NBT %v", scanner.Text(), pieces, hash, saved)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 288 || accepted != 192 || clusters == 0 {
		t.Fatalf("incomplete ocean ruin coverage %d/%d/%d", count, accepted, clusters)
	}
	t.Logf("matched %d original cold/warm starts, %d accepted/%d clustered", count, accepted, clusters)
}
