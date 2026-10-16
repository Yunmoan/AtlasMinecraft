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

func TestDesertPyramidBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_pyramid_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 19 || f[0] != "pyramid" {
			t.Fatalf("invalid pyramid oracle vector %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		seed, orientation := parse(1), parse(2)
		minX, minZ := int32(parse(4)), int32(parse(5))
		piece := StructurePiece{ID: "minecraft:tedp", Orientation: [...]string{"south", "west", "north", "east"}[orientation],
			Box:  StructureBoundingBox{minX, 64, minZ, minX + 20, 78, minZ + 20},
			Data: nbt.Object{"HPos": {Type: nbt.Int, Value: int32(-1)}}}
		world := &hutProbeWorld{scenario: int(parse(3)), edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}
		random, worldRandom := NewWorldgenRandom(seed), NewLegacyRandom(seed^0x123456789abcdef)
		result, err := PostProcessDesertPyramidBlocks(world, &piece, random, worldRandom, seed, int32(parse(6)), int32(parse(7)))
		if err != nil || len(result.PendingEntities) != 0 {
			t.Fatalf("%s: pyramid postprocess result %v (%v)", scanner.Text(), result, err)
		}
		height, _ := piece.Data["HPos"].Number()
		blockHash, postHash, err := structureBlockDigest(world)
		if err != nil {
			t.Fatal(err)
		}
		var entities []string
		for p, entity := range world.entities {
			id, _ := entity.data.Text("id")
			loot, _ := entity.data.Text("LootTable")
			lootSeed, _ := entity.data.Number("LootTableSeed")
			entities = append(entities, fmt.Sprintf("%d %d %d %s %s %d", p.X, p.Y, p.Z, id, loot, lootSeed))
		}
		sort.Strings(entities)
		entityHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(entities, "\n"))))
		flags := ""
		for index := 0; index < 4; index++ {
			placed, _ := piece.Data[fmt.Sprintf("hasPlacedChest%d", index)].Number()
			if placed == 0 {
				flags += "0"
			} else {
				flags += "1"
			}
		}
		next, worldNext := int64(random.NextLong()), int64(worldRandom.NextLong())
		if height != parse(8) || len(world.edits) != int(parse(9)) || blockHash != f[10] || len(world.post) != int(parse(11)) || postHash != f[12] || len(entities) != int(parse(13)) || entityHash != f[14] || flags != f[15] || int64(piece.Box.MinY) != parse(16) || next != parse(17) || worldNext != parse(18) {
			t.Fatalf("%s:\n got HPos=%d blocks=%d/%s post=%d/%s entities=%d/%s flags=%s minY=%d random=%d worldRandom=%d", scanner.Text(), height, len(world.edits), blockHash, len(world.post), postHash, len(entities), entityHash, flags, piece.Box.MinY, next, worldNext)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 768 {
		t.Fatalf("got %d pyramid block vectors, want 768", count)
	}
	t.Logf("matched %d original pyramid geometry / traps / loot / archaeology / clipping cases", count)
}
