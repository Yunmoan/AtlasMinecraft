package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestIglooBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_igloo_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 16 || f[0] != "igloo" {
			t.Fatalf("invalid igloo vector %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		name := "minecraft:igloo/" + f[4]
		offset := iglooOffsets[name]
		depth := int32(0)
		if f[4] == "bottom" {
			depth = 15
		}
		origin := FeaturePosition{int32(parse(5)) + offset.X, 90 + offset.Y - depth, int32(parse(6)) + offset.Z}
		template, err := LoadIglooTemplate(name)
		if err != nil {
			t.Fatal(err)
		}
		rotation := TemplateRotation(parse(2))
		box, err := template.BoundingBox(origin, iglooPivots[name], TemplateMirrorNone, rotation)
		if err != nil {
			t.Fatal(err)
		}
		piece := StructurePiece{ID: "minecraft:iglu", Orientation: "north", Box: box, Data: nbt.Object{"Template": {Type: nbt.String, Value: name}, "Rot": {Type: nbt.String, Value: templateRotationNames[rotation]}, "TPX": {Type: nbt.Int, Value: origin.X}, "TPY": {Type: nbt.Int, Value: origin.Y}, "TPZ": {Type: nbt.Int, Value: origin.Z}}}
		world := &hutProbeWorld{scenario: int(parse(3)), edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}
		random := NewWorldgenRandom(parse(1))
		_, err = PostProcessIglooBlocks(world, &piece, random, int32(parse(7)), int32(parse(8)))
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		hash, _, err := structureBlockDigest(world)
		if err != nil {
			t.Fatal(err)
		}
		var entityRows, fullEntities []string
		for p, entity := range world.entities {
			id, _ := entity.data.Text("id")
			loot, _ := entity.data.Text("LootTable")
			seed, _ := entity.data.Number("LootTableSeed")
			entityRows = append(entityRows, fmt.Sprintf("%d %d %d %s %s %d", p.X, p.Y, p.Z, id, loot, seed))
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
			data["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
			data["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
			data["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
			fullEntities = append(fullEntities, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
		}
		sort.Strings(entityRows)
		entityHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(entityRows, "\n"))))
		sort.Strings(fullEntities)
		fullHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(fullEntities, "\n"))))
		next := int64(random.NextLong())
		if len(world.edits) != int(parse(9)) || hash != f[10] || len(world.entities) != int(parse(11)) || entityHash != f[12] || int64(piece.Box.MinY) != parse(13) || next != parse(14) {
			t.Fatalf("%s:\n got edits %d/%s entities %d/%s minY %d random %d", scanner.Text(), len(world.edits), hash, len(world.entities), entityHash, piece.Box.MinY, next)
		}
		if fullHash != f[15] {
			t.Fatalf("%s: full NBT hash got %s; values: %v", scanner.Text(), fullHash, world.entities)
		}
		if y, _ := piece.Data.Number("TPY"); y != int64(origin.Y) {
			t.Fatalf("temporary igloo alignment changed saved TPY: %d, want %d", y, origin.Y)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 2304 {
		t.Fatalf("matched %d igloo vectors, want 2304", count)
	}
	t.Logf("matched %d original igloo blocks/shape updates/loot/clips/rotation/RNG cases", count)
}

func iglooCanonicalNBT(t *testing.T, tag nbt.Tag) string {
	if object, ok := tag.Object(); ok {
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var text strings.Builder
		text.WriteString("C{")
		for _, key := range keys {
			text.WriteString(base64.StdEncoding.EncodeToString([]byte(key)))
			text.WriteByte(':')
			text.WriteString(iglooCanonicalNBT(t, object[key]))
			text.WriteByte(';')
		}
		text.WriteByte('}')
		return text.String()
	}
	if list, ok := tag.Elements(); ok {
		var text strings.Builder
		fmt.Fprintf(&text, "L%d[", tag.ListType)
		for _, value := range list {
			text.WriteString(iglooCanonicalNBT(t, value))
			text.WriteByte(';')
		}
		text.WriteByte(']')
		return text.String()
	}
	data, err := nbt.MarshalNetwork(tag)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(data)
}
