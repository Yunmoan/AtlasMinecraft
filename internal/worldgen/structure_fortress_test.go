package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNetherFortressStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_fortress_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 5 || f[0] != "start" {
			t.Fatalf("bad start row %q", scanner.Text())
		}
		seed, _ := strconv.ParseInt(f[1], 10, 64)
		cx64, _ := strconv.ParseInt(f[2], 10, 32)
		cz64, _ := strconv.ParseInt(f[3], 10, 32)
		size, _ := strconv.Atoi(f[4])
		cx, cz := int32(cx64), int32(cz64)
		random := NewLegacyRandom(seed)
		mx, mz := random.NextLong(), random.NextLong()
		random = NewLegacyRandom(int64(uint64(int64(cx))*mx ^ uint64(int64(cz))*mz ^ uint64(seed)))
		pieces := assembleNetherFortress(random, cx*16+2, cz*16+2)
		if len(pieces) != size {
			t.Fatalf("seed %d chunk %d,%d pieces=%d want=%d", seed, cx, cz, len(pieces), size)
		}
		for index, piece := range pieces {
			if !scanner.Scan() {
				t.Fatal("missing piece")
			}
			tag, err := piece.NBT()
			if err != nil {
				t.Fatal(err)
			}
			orientation, _ := tag["O"].Number()
			b := piece.Box
			actual := fmt.Sprintf("piece %s %d %d %d %d %d %d %d %d", piece.ID, int32(orientation), piece.GenerationDepth, b.MinX, b.MinY, b.MinZ, b.MaxX, b.MaxY, b.MaxZ)
			for _, name := range []string{"Seed", "Chest", "Mob"} {
				value, ok := piece.Data[name].Number()
				if ok {
					actual += fmt.Sprintf(" %d", int32(value))
				} else {
					actual += " -"
				}
			}
			if actual != scanner.Text() {
				t.Fatalf("seed %d chunk %d,%d piece %d\n got %s\nwant %s", seed, cx, cz, index, actual, scanner.Text())
			}
		}
		if !scanner.Scan() {
			t.Fatal("missing random continuation")
		}
		actual := fmt.Sprintf("next %d", int64(random.NextLong()))
		if actual != scanner.Text() {
			t.Fatalf("seed %d chunk %d,%d random got %s want %s", seed, cx, cz, actual, scanner.Text())
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 56 {
		t.Fatalf("got %d starts want 56", count)
	}
}

func TestNetherFortressBiomeAcceptanceAndSavedClone(t *testing.T) {
	start, err := GenerateNetherFortressStructureStart(1, 2, -3, func(x, y, z int32) (string, error) {
		if x != 8 || y != 16 || z != -12 {
			t.Fatalf("biome query %d %d %d", x, y, z)
		}
		return "minecraft:nether_wastes", nil
	})
	if err != nil || start == nil {
		t.Fatalf("accepted start %v %v", start, err)
	}
	rejected, err := GenerateNetherFortressStructureStart(1, 2, -3, func(x, y, z int32) (string, error) { return "minecraft:plains", nil })
	if err != nil || rejected != nil {
		t.Fatalf("rejected start %v %v", rejected, err)
	}
	copy := start.Clone()
	copy.Pieces[0].Data["probe"] = nbt.Tag{Type: nbt.Int, Value: int32(1)}
	if _, found := start.Pieces[0].Data["probe"]; found {
		t.Fatal("start clone aliases saved piece data")
	}
}
