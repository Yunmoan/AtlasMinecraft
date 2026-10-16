package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestScatteredStructureStartsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_starts_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type target struct {
		seed int64
		name string
		x, z int32
	}
	type column struct{ x, z int32 }
	type startCase struct {
		biome string
		valid bool
	}
	type testCase struct {
		heights map[column]int32
		starts  []startCase
		piece   string
	}
	cases := make(map[target]*testCase)
	scanner := bufio.NewScanner(file)
	vectorCount := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) < 7 {
			t.Fatalf("invalid scattered-start vector %q", scanner.Text())
		}
		parse := func(index int) int64 {
			number, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return number
		}
		key := target{parse(1), f[2], int32(parse(3)), int32(parse(4))}
		item := cases[key]
		if item == nil {
			item = &testCase{heights: make(map[column]int32)}
			cases[key] = item
		}
		switch f[0] {
		case "height":
			if len(f) != 8 {
				t.Fatalf("invalid height vector %q", scanner.Text())
			}
			item.heights[column{int32(parse(5)), int32(parse(6))}] = int32(parse(7))
		case "start":
			if len(f) != 7 {
				t.Fatalf("invalid start vector %q", scanner.Text())
			}
			item.starts = append(item.starts, startCase{f[5], parse(6) == 1})
		case "piece":
			item.piece = scanner.Text()
		default:
			t.Fatalf("unknown scattered-start vector %q", scanner.Text())
		}
		vectorCount++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 144 || vectorCount != 1086 {
		t.Fatalf("got %d cases and %d vectors, want 144 and 1086", len(cases), vectorCount)
	}
	accepted := 0
	for key, item := range cases {
		for _, scenario := range item.starts {
			start, err := GenerateScatteredStructureStart(key.seed, key.name, key.x, key.z, 63,
				func(x, y, z int32) (string, error) {
					centerX, centerZ := key.x*16+8, key.z*16+8
					if x != centerX>>2 || z != centerZ>>2 || y != item.heights[column{centerX, centerZ}]>>2 {
						return "", fmt.Errorf("incorrect start biome coordinates %d,%d,%d", x, y, z)
					}
					return scenario.biome, nil
				}, func(x, z int32) (int32, error) {
					height, found := item.heights[column{x, z}]
					if !found {
						return 0, fmt.Errorf("unexpected base height column %d,%d", x, z)
					}
					return height, nil
				})
			if err != nil || (start != nil) != scenario.valid {
				t.Fatalf("%+v/%s: start=%+v (%v), want accepted=%v", key, scenario.biome, start, err, scenario.valid)
			}
			if start == nil {
				continue
			}
			accepted++
			if len(start.Pieces) != 1 {
				t.Fatalf("expected one original piece, got %d", len(start.Pieces))
			}
			piece, err := start.Pieces[0].NBT()
			if err != nil {
				t.Fatal(err)
			}
			id, _ := piece.Text("id")
			orientation, _ := piece.Number("O")
			depth, _ := piece.Number("GD")
			got := fmt.Sprintf("piece %d %s %d %d %s %d %d", key.seed, key.name, key.x, key.z, id, orientation, depth)
			for _, number := range piece["BB"].Value.([]int32) {
				got += fmt.Sprint(" ", number)
			}
			for _, field := range []string{"Width", "Height", "Depth", "HPos"} {
				value, _ := piece.Number(field)
				got += fmt.Sprint(" ", value)
			}
			fields := []string{"Witch", "Cat"}
			if key.name == "minecraft:desert_pyramid" {
				fields = []string{"hasPlacedChest0", "hasPlacedChest1", "hasPlacedChest2", "hasPlacedChest3"}
			} else if key.name == "minecraft:jungle_pyramid" {
				fields = []string{"placedMainChest", "placedHiddenChest", "placedTrap1", "placedTrap2"}
			}
			for _, field := range fields {
				value, _ := piece.Number(field)
				got += fmt.Sprint(" ", value)
			}
			if got != item.piece {
				t.Fatalf("saved initial piece got %s, want %s", got, item.piece)
			}
			startTag, err := start.NBT()
			if err != nil {
				t.Fatal(err)
			}
			children, _ := startTag["Children"].Elements()
			if len(startTag) != 5 || len(children) != 1 || !reflect.DeepEqual(children[0].Value, piece) {
				t.Fatalf("invalid saved structure start: %+v", startTag)
			}
		}
	}
	if accepted != 78 {
		t.Fatalf("accepted %d scattered starts, want 78", accepted)
	}
	t.Logf("matched %d official start/height/piece vectors with %d accepted starts", vectorCount, accepted)
}

func TestStructureStartReferences(t *testing.T) {
	start := StructureStart{Structure: "minecraft:desert_pyramid", Chunk: StructureChunkPosition{-1, 0}, TerrainAdjustment: "none",
		Pieces: []StructurePiece{{Box: StructureBoundingBox{-16, 64, 0, 4, 78, 20}}}}
	for _, target := range []StructureChunkPosition{{-1, 0}, {0, 0}, {-1, 1}, {0, 1}} {
		got := BuildStructureReferences(target.X, target.Z, []StructureStart{start, start})
		if references := got[start.Structure]; len(references) != 1 || references[0] != 4294967295 {
			t.Fatalf("references for %v: %v", target, got)
		}
	}
	if got := BuildStructureReferences(1, 0, []StructureStart{start}); len(got) != 0 {
		t.Fatalf("unintersecting chunk has references: %v", got)
	}
	start.Chunk.X = -9
	if got := BuildStructureReferences(0, 0, []StructureStart{start}); len(got) != 0 {
		t.Fatalf("outside eight-chunk neighborhood has references: %v", got)
	}
	start.TerrainAdjustment = "beard_box"
	if got, valid := start.BoundingBox(); !valid || got != (StructureBoundingBox{-28, 52, -12, 16, 90, 32}) {
		t.Fatalf("adapted structure bounds: %v/%v", got, valid)
	}
	invalid, err := (StructureStart{}).NBT()
	if err != nil || !reflect.DeepEqual(invalid, nbt.Object{"id": {Type: nbt.String, Value: "INVALID"}}) {
		t.Fatalf("invalid start NBT: %v/%v", invalid, err)
	}
}

func TestStructureStartPrivateReplayAndNBT(t *testing.T) {
	start := StructureStart{Structure: "minecraft:swamp_hut", Chunk: StructureChunkPosition{-1, 2},
		Pieces: []StructurePiece{{ID: "minecraft:tesh", Orientation: "south", Box: StructureBoundingBox{-16, 64, 32, -10, 70, 40},
			Data: nbt.Object{"Witch": {Type: nbt.Byte, Value: int8(0)}, "Array": {Type: nbt.IntArray, Value: []int32{1, 2}}}}}}
	private := start.Clone()
	private.Pieces[0].Data["Witch"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	private.Pieces[0].Data["Array"].Value.([]int32)[0] = 9
	private.Pieces[0].Box.MinY = 65
	flag, _ := start.Pieces[0].Data["Witch"].Number()
	if flag != 0 || start.Pieces[0].Box.MinY != 64 || start.Pieces[0].Data["Array"].Value.([]int32)[0] != 1 {
		t.Fatal("private replay mutated immutable cached start")
	}
	other := start.Clone()
	other.Chunk.X = 0
	compound, err := StructureNBTForChunk(-1, 2, []StructureStart{start, other}, map[string][]int64{
		start.Structure: {12884901887, 12884901887}, "minecraft:stronghold": {},
	})
	if err != nil {
		t.Fatal(err)
	}
	starts, found := compound.Object("starts")
	if !found || len(starts) != 1 {
		t.Fatalf("incorrect owned starts compound %v", compound)
	}
	references, found := compound.Object("References")
	if !found || len(references) != 1 {
		t.Fatalf("incorrect references compound %v", compound)
	}
	packed, found := references[start.Structure].Longs()
	if !found || len(packed) != 1 || packed[0] != 12884901887 {
		t.Fatalf("incorrect packed structure reference %v", references)
	}
}
