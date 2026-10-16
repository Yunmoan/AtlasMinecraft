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

func TestRuinedPortalBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_ruined_portal_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 19 || f[0] != "portal" {
			t.Fatalf("invalid portal block vector %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		variant := int(parse(5))
		placement := RuinedPortalOnLandSurface
		if variant == 2 {
			placement = RuinedPortalOnOceanFloor
		} else if variant == 4 {
			placement = RuinedPortalInNether
		}
		piece, err := NewRuinedPortalTemplatePiece(f[1], FeaturePosition{int32(parse(7)), int32(parse(8)), int32(parse(9))}, placement, ruinedPortalProbeProperties(variant), TemplateMirror(parse(4)), TemplateRotation(parse(3)))
		if err != nil {
			t.Fatal(err)
		}
		world := &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(6)), edits: make(map[FeaturePosition]SurfaceBlockState)}}
		random := NewWorldgenRandom(parse(2))
		_, err = PostProcessRuinedPortalBlocks(world, piece, random, int32(parse(10)), int32(parse(11)))
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
		if len(world.edits) != int(parse(13)) || hash != f[14] || len(world.entities) != int(parse(15)) || entityHash != f[16] || next != parse(17) || savedHash != f[18] {
			if os.Getenv("ATLAS_RUINED_PORTAL_DEBUG") != "" {
				var edits []string
				for p, state := range world.edits {
					id, err := blockstate.BlockStateID(state.Name, state.Properties)
					if err != nil {
						t.Fatal(err)
					}
					edits = append(edits, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
				}
				sort.Strings(edits)
				if err := os.WriteFile(filepath.Join("..", "..", "artifacts", "worldgen-probes", "ruined-portal-got.txt"), []byte(strings.Join(edits, "\n")), 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Fatalf("%s:\n got %d/%s entities %d/%s random %d saved %s", scanner.Text(), len(world.edits), hash, len(world.entities), entityHash, next, savedHash)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 22464 {
		t.Fatalf("got %d portal block vectors", count)
	}
	t.Logf("matched %d full original ruined portal block/shape/liquid/loot/spread/vine/clip/NBT/RNG cases", count)
}
