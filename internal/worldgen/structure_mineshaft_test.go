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

func TestMineshaftStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_mineshaft_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	parse := func(text string) int64 {
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	generators := make(map[int64]*OverworldNoiseGenerator)
	var seed int64
	var cx, cz, sx, sz, surfaceHeight, expectedShift int32
	var mesa bool
	layouts, starts, accepted, totalPieces := 0, 0, 0, 0
	heightAt := func(x, z int32) (int32, error) {
		if x != sx || z != sz {
			return 0, fmt.Errorf("mesa surface coordinates %d,%d want %d,%d", x, z, sx, sz)
		}
		g := generators[seed]
		if g == nil {
			g, err = NewOverworldNoiseGenerator(seed)
			if err != nil {
				return 0, err
			}
			generators[seed] = g
		}
		height, err := g.BaseOccupiedHeightAt(x, z)
		if err != nil {
			return 0, err
		}
		height++
		if height != surfaceHeight {
			return 0, fmt.Errorf("mesa base surface %d want %d", height, surfaceHeight)
		}
		return height, nil
	}
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) == 12 && f[0] == "mineslayout" {
			seed, cx, cz, mesa = parse(f[1]), int32(parse(f[2])), int32(parse(f[3])), parse(f[4]) != 0
			sx, sz, surfaceHeight, expectedShift = int32(parse(f[5])), int32(parse(f[6])), int32(parse(f[7])), int32(parse(f[8]))
			base := NewLegacyRandom(seed)
			xm, zm := base.NextLong(), base.NextLong()
			random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
			pieces, shift, err := assembleMineshaft(random, cx*16+2, cz*16+2, mesa, -64, 63, heightAt)
			if err != nil {
				t.Fatal(err)
			}
			var saved strings.Builder
			for _, piece := range pieces {
				value, err := piece.NBT()
				if err != nil {
					t.Fatal(err)
				}
				saved.WriteString(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: value}))
				saved.WriteByte('\n')
			}
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(saved.String())))
			next := int64(random.NextLong())
			if len(pieces) != int(parse(f[9])) || shift != expectedShift || hash != f[10] || next != parse(f[11]) {
				t.Fatalf("%s: got %d pieces shift%d hash%s next%d", scanner.Text(), len(pieces), shift, hash, next)
			}
			layouts++
			totalPieces += len(pieces)
		} else if len(f) == 8 && f[0] == "minesstart" {
			if seed != parse(f[3]) || cx != int32(parse(f[4])) || cz != int32(parse(f[5])) {
				t.Fatal("start fixture lost its layout")
			}
			start, err := GenerateMineshaftStructureStart(seed, cx, cz, mesa, MineshaftStartSources{MinY: -64, SeaLevel: 63, WorldSurfaceHeightAt: heightAt, BiomeAtQuart: func(x, y, z int32) (string, error) {
				if x != cx*4 || z != cz*4 || y != (50+expectedShift)>>2 {
					return "", fmt.Errorf("mineshaft biome coordinates %d,%d,%d", x, y, z)
				}
				return f[2], nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			saved := nbt.Object{"id": {Type: nbt.String, Value: "INVALID"}}
			count := 0
			if start != nil {
				accepted++
				count = len(start.Pieces)
				saved, err = start.NBT()
				if err != nil {
					t.Fatal(err)
				}
			}
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))))
			if count != int(parse(f[6])) || hash != f[7] {
				t.Fatalf("%s: got %d/%s", scanner.Text(), count, hash)
			}
			starts++
		} else {
			t.Fatalf("invalid mineshaft row %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if layouts != 112 || starts != 224 || accepted != 112 {
		t.Fatalf("incomplete mineshaft cases %d/%d/%d", layouts, starts, accepted)
	}
	t.Logf("matched %d original mineshaft layouts with %d complete pieces and RNG, plus %d starts/%d accepted", layouts, totalPieces, starts, accepted)
}
