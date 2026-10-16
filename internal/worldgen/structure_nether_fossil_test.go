package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type netherFossilProbeWorld struct{ *hutProbeWorld }

func (w *netherFossilProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w.edits[p]; ok {
		return state, nil
	}
	if p.Y <= 59 {
		return structureState("netherrack"), nil
	}
	if p.Y <= 65 && w.scenario > 0 {
		return structureState([]string{"air", "water", "lava"}[w.scenario]), nil
	}
	return structureState("air"), nil
}
func (w *netherFossilProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= 0 && p.Y < 256 }

func TestNetherFossilBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_nether_fossil_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 13 || f[0] != "nether_fossil" {
			t.Fatalf("invalid fossil block vector %q", scanner.Text())
		}
		parse := func(i int) int64 {
			value, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		name := "minecraft:nether_fossils/fossil_" + f[2]
		piece, err := NewNetherFossilTemplatePiece(name, FeaturePosition{int32(parse(5)), 60, int32(parse(6))}, TemplateRotation(parse(3)))
		if err != nil {
			t.Fatal(err)
		}
		w := &netherFossilProbeWorld{&hutProbeWorld{scenario: int(parse(4)), edits: make(map[FeaturePosition]SurfaceBlockState)}}
		random := NewWorldgenRandom(parse(1))
		result, err := PostProcessNetherFossilBlocks(w, piece, random, int32(parse(7)), int32(parse(8)))
		if err != nil {
			t.Fatal(err)
		}
		if len(result.PendingEntities) != 0 || len(w.entities) != 0 {
			t.Fatal("unexpected fossil actors or block entities")
		}
		blocks, _, err := structureBlockDigest(w.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		saved, err := piece.NBT()
		if err != nil {
			t.Fatal(err)
		}
		hash := treasureDigest([]string{iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved})})
		next := int64(random.NextLong())
		if len(w.edits) != int(parse(9)) || blocks != f[10] || hash != f[11] || next != parse(12) {
			t.Fatalf("%s: got blocks %d/%s saved %s random %d", scanner.Text(), len(w.edits), blocks, hash, next)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 3024 {
		t.Fatalf("matched %d cases, want3024", count)
	}
	t.Logf("matched %d original Nether fossil template/rotation/expanded-clip/terrain/NBT/RNG calls", count)
}

func TestNetherFossilStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_nether_fossil_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count, accepted := 0, 0
	generators := make(map[int64]*DimensionNoiseGenerator)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 7 || f[0] != "nether_fossil_start" {
			t.Fatalf("invalid fossil start %q", scanner.Text())
		}
		parse := func(i int) int64 {
			value, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		seed := parse(1)
		g := generators[seed]
		if g == nil {
			g, err = NewDimensionNoiseGenerator(seed, "nether")
			if err != nil {
				t.Fatal(err)
			}
			generators[seed] = g
		}
		start, err := GenerateNetherFossilStructureStart(seed, int32(parse(2)), int32(parse(3)), NetherFossilStartSources{MinY: 0, GenerationDepth: 128, SeaLevel: 32, BaseColumnAt: g.BaseStructureColumnAt, BiomeAtQuart: func(_, _, _ int32) (string, error) { return f[4], nil }})
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		if (start != nil) != (parse(5) != 0) {
			t.Fatalf("%s: acceptance mismatch", scanner.Text())
		}
		if start != nil {
			data, err := start.Pieces[0].NBT()
			if err != nil {
				t.Fatal(err)
			}
			if hash := treasureDigest([]string{iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})}); hash != f[6] {
				t.Fatalf("%s: saved fossil NBT %s", scanner.Text(), hash)
			}
			accepted++
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 192 {
		t.Fatalf("matched %d starts, want192", count)
	}
	if accepted == 0 {
		t.Fatal("no accepted fossil start in official fixture")
	}
	t.Logf("matched %d original Nether fossil starts with %d accepted pieces", count, accepted)
}
