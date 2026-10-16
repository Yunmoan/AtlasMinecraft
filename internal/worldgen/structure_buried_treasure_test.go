package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type treasureProbeWorld struct {
	*hutProbeWorld
	anchor FeaturePosition
}

func (w *treasureProbeWorld) MinBuildHeight() int32 {
	if w.scenario == 23 {
		return 0
	}
	return -64
}
func (w *treasureProbeWorld) HeightmapAt(_ string, _, _ int32) (int32, error) {
	if w.scenario == 22 {
		return w.MinBuildHeight(), nil
	}
	if w.scenario == 23 {
		return 1, nil
	}
	if w.scenario >= 18 && w.scenario <= 21 {
		return 62, nil
	}
	return 61, nil
}
func (w *treasureProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w.edits[p]; ok {
		return state, nil
	}
	mode := w.scenario
	if mode == 22 {
		return structureState("air"), nil
	}
	floor := int32(60)
	if mode == 23 {
		floor = 0
	}
	if p.Y < floor {
		return structureState("bedrock"), nil
	}
	if p.Y == floor {
		return structureState([]string{"stone", "sandstone", "andesite", "granite", "diorite", "deepslate"}[mode%6]), nil
	}
	if mode == 21 && p.X == w.anchor.X && p.Z == w.anchor.Z && p.Y == floor+1 {
		return structureState("chest"), nil
	}
	if mode >= 12 && mode < 18 && p.Y == floor+1 && p.X == w.anchor.X+1 {
		return structureState("dirt"), nil
	}
	if mode >= 18 && mode < 21 && p.Y == floor+1 && p.X == w.anchor.X && p.Z == w.anchor.Z {
		return structureState("sand"), nil
	}
	if p.Y <= floor+3 {
		names := []string{"air", "water", "lava", "water"}
		state := structureState(names[(mode/6)%4])
		if mode/6%4 == 3 {
			state.Properties = map[string]string{"level": "1"}
		}
		return state, nil
	}
	return structureState("air"), nil
}

func treasureDigest(rows []string) string {
	sort.Strings(rows)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
}

func TestBuriedTreasureBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_buried_treasure_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 13 || f[0] != "treasure" {
			t.Fatalf("invalid vector %q", scanner.Text())
		}
		parse := func(i int) int64 {
			n, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		anchor := FeaturePosition{int32(parse(3)), 90, int32(parse(4))}
		piece := StructurePiece{ID: "minecraft:btp", Box: StructureBoundingBox{anchor.X, 90, anchor.Z, anchor.X, 90, anchor.Z}}
		w := &treasureProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(2)), edits: make(map[FeaturePosition]SurfaceBlockState)}, anchor: anchor}
		random := NewWorldgenRandom(parse(1))
		_, err := PostProcessBuriedTreasureBlocks(w, &piece, random, int32(parse(5)), int32(parse(6)))
		if err != nil {
			t.Fatal(err)
		}
		blocks, _, err := structureBlockDigest(w.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		var entities []string
		for p, entity := range w.entities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
			data["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
			data["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
			data["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
			entities = append(entities, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
		}
		saved, err := piece.NBT()
		if err != nil {
			t.Fatal(err)
		}
		nbtHash := treasureDigest([]string{iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved})})
		next := int64(random.NextLong())
		if len(w.edits) != int(parse(7)) || blocks != f[8] || len(w.entities) != int(parse(9)) || treasureDigest(entities) != f[10] || nbtHash != f[11] || next != parse(12) {
			t.Fatalf("%s: got blocks %d/%s entities %d/%s saved %s next %d", scanner.Text(), len(w.edits), blocks, len(w.entities), treasureDigest(entities), nbtHash, next)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 576 {
		t.Fatalf("matched %d cases, want576", count)
	}
	t.Logf("matched %d original buried-treasure block/clip/liquid/NBT/RNG cases", count)
}

func TestBuriedTreasureStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_buried_treasure_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	count := 0
	generators := make(map[int64]*OverworldNoiseGenerator)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 8 || f[0] != "treasure_start" {
			t.Fatalf("invalid start %q", scanner.Text())
		}
		parse := func(i int) int64 {
			n, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		seed, cx, cz := parse(1), int32(parse(2)), int32(parse(3))
		g := generators[seed]
		if g == nil {
			g, err = NewOverworldNoiseGenerator(seed)
			if err != nil {
				t.Fatal(err)
			}
			generators[seed] = g
		}
		height := func(x, z int32) (int32, error) {
			column, err := g.BaseStructureColumnAt(x, z)
			if err != nil {
				return 0, err
			}
			y := column.OccupiedHeight(true)
			if int64(y) != parse(5) {
				t.Fatalf("%s: actual base height %d", scanner.Text(), y)
			}
			return y, nil
		}
		biome := func(x, y, z int32) (string, error) {
			if x != (cx*16+8)>>2 || z != (cz*16+8)>>2 || y != int32(parse(5))>>2 {
				t.Fatalf("unexpected start biome lookup %d,%d,%d", x, y, z)
			}
			return f[4], nil
		}
		start, err := GenerateBuriedTreasureStructureStart(cx, cz, biome, height)
		if err != nil {
			t.Fatal(err)
		}
		if (start != nil) != (parse(6) != 0) {
			t.Fatalf("%s: acceptance mismatch", scanner.Text())
		}
		if start != nil {
			data, err := start.Pieces[0].NBT()
			if err != nil {
				t.Fatal(err)
			}
			if got := treasureDigest([]string{iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})}); got != f[7] {
				t.Fatalf("%s: start NBT %s", scanner.Text(), got)
			}
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 96 {
		t.Fatalf("matched %d starts, want96", count)
	}
	t.Logf("matched %d actual buried-treasure starts/base ocean-floor heights", count)
}
