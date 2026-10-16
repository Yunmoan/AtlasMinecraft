package worldgen

import (
	"atlas/internal/blockstate"
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

func TestShipwreckStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_shipwreck_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	generators := make(map[int64]*OverworldNoiseGenerator)
	scanner := bufio.NewScanner(file)
	count, accepted := 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 9 || f[0] != "shipstart" {
			t.Fatalf("invalid shipwreck start row %q", scanner.Text())
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
		beached := strings.HasSuffix(f[1], "beached")
		start, err := GenerateShipwreckStructureStart(seed, cx, cz, beached, func(x, y, z int32) (string, error) {
			if x != (cx*16+8)>>2 || z != (cz*16+8)>>2 || y != int32(parse(6))>>2 {
				return "", fmt.Errorf("incorrect shipwreck biome coordinates %d,%d,%d", x, y, z)
			}
			return f[2], nil
		}, func(x, z int32) (int32, error) {
			column, err := generator.BaseStructureColumnAt(x, z)
			if err != nil {
				return 0, err
			}
			height := column.OccupiedHeight(!beached)
			if height != int32(parse(6)) {
				return 0, fmt.Errorf("base occupied height got %d want %s", height, f[6])
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
	if count != 288 || accepted != 192 {
		t.Fatalf("incomplete shipwreck starts %d/%d", count, accepted)
	}
	t.Logf("matched %d original ocean/beached starts, %d accepted", count, accepted)
}

type shipwreckProbeWorld struct {
	*ruinedPortalProbeWorld
	ticks map[FeaturePosition]bool
}

func (world *shipwreckProbeWorld) ScheduleFluidUpdate(position FeaturePosition) error {
	if world.ticks == nil {
		world.ticks = make(map[FeaturePosition]bool)
	}
	world.ticks[position] = true
	return nil
}

func TestShipwreckBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_shipwreck_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 18 || f[0] != "ship" && f[0] != "shipsequence" {
			t.Fatalf("invalid shipwreck block row %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		piece, err := NewShipwreckTemplatePiece(f[1], FeaturePosition{int32(parse(6)), 90, int32(parse(7))}, TemplateRotation(parse(3)), parse(4) != 0)
		if err != nil {
			t.Fatal(err)
		}
		world := &shipwreckProbeWorld{ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(5)), edits: make(map[FeaturePosition]SurfaceBlockState)}}}
		random := NewWorldgenRandom(parse(2))
		if f[0] == "ship" {
			_, err = PostProcessShipwreckBlocks(world, piece, random, int32(parse(8)), int32(parse(9)))
		} else {
			box := piece.Box
			for x := box.MinX >> 4; x <= box.MaxX>>4; x++ {
				for z := box.MinZ >> 4; z <= box.MaxZ>>4; z++ {
					if _, err = PostProcessShipwreckBlocks(world, piece, random, x, z); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		hash, _, err := structureBlockDigest(world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		var rows []string
		for p, entity := range world.entities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
			data["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
			data["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
			data["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
			rows = append(rows, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
		}
		sort.Strings(rows)
		entityHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
		saved, err := piece.NBT()
		if err != nil {
			t.Fatal(err)
		}
		savedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))))
		next := int64(random.NextLong())
		var ticks []string
		for position := range world.ticks {
			ticks = append(ticks, fmt.Sprintf("%d %d %d", position.X, position.Y, position.Z))
		}
		sort.Strings(ticks)
		tickHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(ticks, "\n"))))
		if len(world.edits) != int(parse(10)) || hash != f[11] || len(world.entities) != int(parse(12)) || entityHash != f[13] || next != parse(14) || savedHash != f[15] || len(world.ticks) != int(parse(16)) || tickHash != f[17] {
			if os.Getenv("ATLAS_SHIPWRECK_DEBUG") != "" {
				var edits []string
				for p, state := range world.edits {
					id, err := blockstate.BlockStateID(state.Name, state.Properties)
					if err != nil {
						t.Fatal(err)
					}
					edits = append(edits, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
				}
				sort.Strings(edits)
				if err := os.WriteFile(filepath.Join("..", "..", "artifacts", "worldgen-probes", "shipwreck-got.txt"), []byte(strings.Join(edits, "\n")), 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Fatalf("%s:\n got %d/%s entities %d/%s random %d saved %s ticks %d/%s; NBT %v", scanner.Text(), len(world.edits), hash, len(world.entities), entityHash, next, savedHash, len(world.ticks), tickHash, saved)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 8640 {
		t.Fatalf("got %d shipwreck block cases", count)
	}
	t.Logf("matched %d original shipwreck placement/liquid/shape/loot/NBT/RNG cases", count)
}
