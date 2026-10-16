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

func TestJungleTempleBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_jungle_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 19 || f[0] != "jungle" {
			t.Fatalf("invalid jungle oracle vector %q", scanner.Text())
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
		xSize, zSize := int32(12), int32(15)
		if orientation%2 != 0 {
			xSize, zSize = zSize, xSize
		}
		piece := StructurePiece{ID: "minecraft:tejp", Orientation: [...]string{"south", "west", "north", "east"}[orientation],
			Box:  StructureBoundingBox{minX, 64, minZ, minX + xSize - 1, 73, minZ + zSize - 1},
			Data: nbt.Object{"HPos": {Type: nbt.Int, Value: int32(-1)}}}
		world := &hutProbeWorld{scenario: int(parse(3)), edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}
		random, worldRandom := NewWorldgenRandom(seed), NewLegacyRandom(seed^0x123456789abcdef)
		result, err := PostProcessJungleTempleBlocks(world, &piece, random, int32(parse(6)), int32(parse(7)))
		if err != nil || len(result.PendingEntities) != 0 {
			t.Fatalf("%s: jungle postprocess result %v (%v)", scanner.Text(), result, err)
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
		for _, field := range []string{"placedMainChest", "placedHiddenChest", "placedTrap1", "placedTrap2"} {
			placed, _ := piece.Data[field].Number()
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
	if count != 960 {
		t.Fatalf("got %d jungle block vectors, want 960", count)
	}
	t.Logf("matched %d original jungle geometry / random moss / traps / puzzle / loot / clipping cases", count)
}
