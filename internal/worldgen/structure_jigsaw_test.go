package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestJigsawStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_jigsaw_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	groups := make(map[string][]string)
	rows := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 9 || f[0] != "jigsaw_start" {
			t.Fatalf("invalid Jigsaw vector %.200s", scanner.Text())
		}
		groups[f[1]] = append(groups[f[1]], scanner.Text())
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if rows != 1920 || len(groups) != 10 {
		t.Fatalf("incomplete Jigsaw fixture: %d rows, %d families", rows, len(groups))
	}
	var names []string
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	overworld := make(map[int64]*OverworldNoiseGenerator)
	heights := make(map[int64]map[FeaturePosition]int32)
	for _, name := range names {
		t.Run(strings.TrimPrefix(name, "minecraft:"), func(t *testing.T) {
			count, accepted, totalPieces := 0, 0, 0
			for _, line := range groups[name] {
				f := strings.Fields(line)
				if len(f) != 9 || f[0] != "jigsaw_start" {
					t.Fatalf("invalid Jigsaw vector %q", line)
				}
				parse := func(index int) int64 {
					value, err := strconv.ParseInt(f[index], 10, 64)
					if err != nil {
						t.Fatal(err)
					}
					return value
				}
				seed, cx, cz := parse(2), int32(parse(3)), int32(parse(4))
				base := NewLegacyRandom(seed)
				xm, zm := base.NextLong(), base.NextLong()
				random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
				sources := JigsawStartSources{MinY: 0, Height: 256, GenerationDepth: 128, BiomeAtQuart: func(_, _, _ int32) (string, error) { return f[5], nil }}
				if f[1] != "minecraft:bastion_remnant" {
					sources.MinY, sources.Height, sources.GenerationDepth = -64, 384, 384
					g := overworld[seed]
					if g == nil {
						g, err = NewOverworldNoiseGenerator(seed)
						if err != nil {
							t.Fatal(err)
						}
						overworld[seed] = g
						heights[seed] = make(map[FeaturePosition]int32)
					}
					sources.BaseHeightAt = func(x, z int32, kind string) (int32, error) {
						key := FeaturePosition{X: x, Z: z}
						if kind != "WORLD_SURFACE_WG" {
							return 0, fmt.Errorf("unexpected Jigsaw WG heightmap %q", kind)
						}
						height, found := heights[seed][key]
						if !found {
							var err error
							height, err = g.BaseOccupiedHeightAt(x, z)
							if err != nil {
								return 0, err
							}
							heights[seed][key] = height
						}
						return height + 1, nil
					}
				}
				start, err := generateJigsawStructureStart(seed, cx, cz, f[1], sources, random)
				if err != nil {
					t.Fatalf("%s: %v", line, err)
				}
				pieces := 0
				hash := "-"
				if start != nil {
					pieces = len(start.Pieces)
					saved, err := start.NBT()
					if err != nil {
						t.Fatal(err)
					}
					hash = treasureDigest([]string{iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved})})
					if pieces != int(parse(6)) || hash != f[7] {
						if seed == 0 && cx == 0 && cz == 0 {
							os.WriteFile(filepath.Join("..", "..", "artifacts", "worldgen-probes", "jigsaw-go-first-"+strings.TrimPrefix(name, "minecraft:")+".txt"), []byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved})), 0600)
						}
						t.Fatalf("%s: got %d pieces, NBT %s", line, pieces, hash)
					}
					accepted++
					totalPieces += pieces
				}
				if pieces != int(parse(6)) {
					t.Fatalf("%s: acceptance mismatch", line)
				}
				if next := int64(random.NextLong()); next != parse(8) {
					t.Fatalf("%s: random continuation %d", line, next)
				}
				count++
			}
			if count != 192 || accepted != 96 {
				t.Fatalf("incomplete Jigsaw oracle: %d cases, %d accepted", count, accepted)
			}
			t.Logf("matched %d complete original Jigsaw starts, %d accepted, %d pieces, saved NBT and RNG continuation", count, accepted, totalPieces)
		})
	}
}

func TestJigsawStartsParallelAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "worldgen", "structure_jigsaw_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var cases [][]string
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 9 && f[1] == "minecraft:bastion_remnant" && f[6] != "0" {
			cases = append(cases, f)
			if len(cases) == 24 {
				break
			}
		}
	}
	if len(cases) != 24 {
		t.Fatal("incomplete parallel Bastion fixture")
	}
	var wg sync.WaitGroup
	workers := make(chan struct{}, 4)
	for _, f := range cases {
		workers <- struct{}{}
		wg.Add(1)
		go func(f []string) {
			defer wg.Done()
			defer func() { <-workers }()
			parse := func(index int) int64 {
				value, err := strconv.ParseInt(f[index], 10, 64)
				if err != nil {
					t.Error(err)
				}
				return value
			}
			start, err := GenerateJigsawStructureStart(parse(2), int32(parse(3)), int32(parse(4)), f[1], JigsawStartSources{MinY: 0, Height: 256, GenerationDepth: 128, BiomeAtQuart: func(_, _, _ int32) (string, error) { return f[5], nil }})
			if err != nil {
				t.Error(err)
				return
			}
			if start == nil {
				t.Errorf("parallel original start was rejected: %v", f)
				return
			}
			saved, err := start.NBT()
			if err != nil {
				t.Error(err)
				return
			}
			hash := treasureDigest([]string{iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved})})
			if hash != f[7] || len(start.Pieces) != int(parse(6)) {
				t.Errorf("parallel assembly mismatch: %v", f)
			}
		}(f)
	}
	wg.Wait()
}
