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

func TestOceanRuinBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_ocean_ruin_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count, actors, suspicious := 0, 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 19 || f[0] != "ocean" && f[0] != "oceansequence" {
			t.Fatalf("invalid ocean ruin block row %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		cold, large := !strings.Contains(f[1], "warm"), strings.Contains(f[1], "big_")
		integrity := float32(0.8)
		if strings.Contains(f[1], "cracked") {
			integrity = 0.7
		} else if strings.Contains(f[1], "mossy") {
			integrity = 0.5
		} else if large {
			integrity = 0.9
		}
		piece, err := NewOceanRuinTemplatePiece(f[1], FeaturePosition{int32(parse(5)), 90, int32(parse(6))}, TemplateRotation(parse(3)), integrity, cold, large)
		if err != nil {
			t.Fatal(err)
		}
		world := &shipwreckProbeWorld{ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(4)), edits: make(map[FeaturePosition]SurfaceBlockState)}}}
		random := NewWorldgenRandom(parse(2))
		var pending []PendingStructureEntity
		if f[0] == "ocean" {
			result, err := PostProcessOceanRuinBlocks(world, piece, random, parse(2), int32(parse(7)), int32(parse(8)))
			if err != nil {
				t.Fatalf("%s: %v", scanner.Text(), err)
			}
			pending = append(pending, result.PendingEntities...)
		} else {
			box := piece.Box
			for x := box.MinX >> 4; x <= box.MaxX>>4; x++ {
				for z := box.MinZ >> 4; z <= box.MaxZ>>4; z++ {
					result, err := PostProcessOceanRuinBlocks(world, piece, random, parse(2), x, z)
					if err != nil {
						t.Fatalf("%s: %v", scanner.Text(), err)
					}
					pending = append(pending, result.PendingEntities...)
				}
			}
		}
		hash, _, err := structureBlockDigest(world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		digest := func(rows []string) string {
			sort.Strings(rows)
			return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
		}
		var entityRows []string
		for p, entity := range world.entities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
			data["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
			data["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
			data["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
			entityRows = append(entityRows, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
			if entity.block == "minecraft:suspicious_sand" || entity.block == "minecraft:suspicious_gravel" {
				suspicious++
			}
		}
		var ticks, actorRows []string
		for p := range world.ticks {
			ticks = append(ticks, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
		}
		for _, actor := range pending {
			actorRows = append(actorRows, fmt.Sprintf("%s %d %d %d %t", actor.ID, actor.Position.X, actor.Position.Y, actor.Position.Z, actor.Persistent))
		}
		actors += len(pending)
		saved, err := piece.NBT()
		if err != nil {
			t.Fatal(err)
		}
		savedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))))
		entityHash, tickHash, actorHash := digest(entityRows), digest(ticks), digest(actorRows)
		next := int64(random.NextLong())
		if len(world.edits) != int(parse(9)) || hash != f[10] || len(world.entities) != int(parse(11)) || entityHash != f[12] || next != parse(13) || savedHash != f[14] || len(world.ticks) != int(parse(15)) || tickHash != f[16] || len(pending) != int(parse(17)) || actorHash != f[18] {
			if os.Getenv("ATLAS_OCEAN_RUIN_DEBUG") != "" {
				var edits []string
				for p, state := range world.edits {
					id, err := blockstate.BlockStateID(state.Name, state.Properties)
					if err != nil {
						t.Fatal(err)
					}
					edits = append(edits, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
				}
				sort.Strings(edits)
				if err := os.WriteFile(filepath.Join("..", "..", "artifacts", "worldgen-probes", "ocean-ruin-got.txt"), []byte(strings.Join(edits, "\n")), 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Fatalf("%s:\n got %d/%s entities %d/%s random %d saved %s ticks %d/%s actors %d/%s; NBT %v", scanner.Text(), len(world.edits), hash, len(world.entities), entityHash, next, savedHash, len(world.ticks), tickHash, len(pending), actorHash, saved)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 10368 || suspicious == 0 || actors == 0 {
		t.Fatalf("incomplete ocean block coverage %d/%d/%d", count, suspicious, actors)
	}
	t.Logf("matched %d original ocean ruin blocks/archaeology/NBT/liquid/ticks/markers/RNG cases (%d suspicious records, %d persistent drowned requests)", count, suspicious, actors)
}
