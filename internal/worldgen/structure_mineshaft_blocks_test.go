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

type mineshaftProbeWorld struct {
	hutProbeWorld
	mesa  bool
	ticks map[FeaturePosition]bool
}

func (w *mineshaftProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if s, found := w.edits[p]; found {
		return s, nil
	}
	mode := w.scenario
	if mode == 4 {
		return structureState("air"), nil
	}
	if mode == 2 && p.Y == 65 {
		return structureState("water"), nil
	}
	mod := func(n, m int32) int32 {
		v := n % m
		if v < 0 {
			v += m
		}
		return v
	}
	if mode == 7 && mod(p.X+p.Y+p.Z, 7) == 0 {
		wood := "oak"
		if w.mesa {
			wood = "dark_oak"
		}
		return structureState([]string{wood + "_planks", wood + "_log", wood + "_fence", "chain"}[mod(p.X+p.Z, 4)]), nil
	}
	if mode == 0 || mode == 2 || mode == 3 || mode == 7 {
		if p.Y <= 100 {
			return structureState("stone"), nil
		}
		return structureState("air"), nil
	}
	floor, roof := int32(40), int32(71)
	switch mode {
	case 1:
		floor, roof = 50, 72
	case 5:
		roof = 90
	case 6:
		roof = 100
	case 8:
		roof = 73
	}
	if p.Y <= floor {
		return structureState("stone"), nil
	}
	if mode == 5 && p.Y <= 50 {
		return structureState("water"), nil
	}
	if mode == 6 && p.Y <= 50 {
		return structureState("lava"), nil
	}
	if p.Y >= roof {
		if p.Y == roof && mode == 8 {
			return structureState("sand"), nil
		}
		if p.Y == roof && mode == 9 {
			return structureState("oak_leaves"), nil
		}
		return structureState("stone"), nil
	}
	return structureState("air"), nil
}
func (w *mineshaftProbeWorld) MineshaftBiomeAt(FeaturePosition) (string, error) {
	if w.scenario == 3 {
		return "minecraft:deep_dark", nil
	}
	return "minecraft:plains", nil
}
func (w *mineshaftProbeWorld) MinBuildHeight() int32 { return -64 }
func (w *mineshaftProbeWorld) MaxBuildHeight() int32 { return 320 }
func (w *mineshaftProbeWorld) HeightmapAt(kind string, x, z int32) (int32, error) {
	if kind != "OCEAN_FLOOR_WG" {
		return 0, fmt.Errorf("unexpected mine heightmap %s", kind)
	}
	for y := int32(319); y >= -64; y-- {
		s, err := w.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return 0, err
		}
		solid, err := oceanFloorSolid(s)
		if err != nil {
			return 0, err
		}
		if solid {
			return y + 1, nil
		}
	}
	return -64, nil
}
func (w *mineshaftProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	w.ticks[p] = true
	return nil
}

func mineshaftProbePiece(spec string, mesa bool, orientation int, ax, az int32) StructurePiece {
	width, height, depth := int32(3), int32(3), int32(20)
	id := "mscorridor"
	switch spec {
	case "rail":
		depth = 15
	case "spider", "spawned":
		depth = 10
	case "crossing":
		width, height, depth, id = 5, 3, 5, "mscrossing"
	case "tallcrossing":
		width, height, depth, id = 5, 7, 5, "mscrossing"
	case "stairs":
		height, depth, id = 8, 9, "msstairs"
	case "room":
		width, height, depth, id = 11, 9, 13, "msroom"
	}
	sx, sz := width, depth
	if orientation%2 != 0 && spec != "room" {
		sx, sz = sz, sx
	}
	direction := []string{"south", "west", "north", "east"}[orientation]
	if id == "mscrossing" || id == "msroom" {
		direction = ""
	}
	typeID := int32(0)
	if mesa {
		typeID = 1
	}
	piece := StructurePiece{ID: "minecraft:" + id, Orientation: direction, Box: StructureBoundingBox{ax, 64, az, ax + sx - 1, 64 + height - 1, az + sz - 1}, Data: nbt.Object{"MST": {Type: nbt.Int, Value: typeID}}}
	if id == "mscorridor" {
		piece.Data["Num"] = nbt.Tag{Type: nbt.Int, Value: depth / 5}
		for name, enabled := range map[string]bool{"hr": spec == "rail", "sc": spec == "spider" || spec == "spawned", "hps": spec == "spawned"} {
			value := int8(0)
			if enabled {
				value = 1
			}
			piece.Data[name] = nbt.Tag{Type: nbt.Byte, Value: value}
		}
	}
	if id == "mscrossing" {
		value := int8(0)
		if spec == "tallcrossing" {
			value = 1
		}
		piece.Data["tf"] = nbt.Tag{Type: nbt.Byte, Value: value}
		piece.Data["D"] = nbt.Tag{Type: nbt.Int, Value: int32(orientation)}
	}
	if id == "msroom" {
		piece.Data["Entrances"] = nbt.Tag{Type: nbt.List, ListType: nbt.IntArray, Value: []nbt.Tag{{Type: nbt.IntArray, Value: []int32{ax + 2, 65, az, ax + 4, 67, az + 1}}, {Type: nbt.IntArray, Value: []int32{ax + sx - 2, 66, az + 5, ax + sx - 1, 68, az + 7}}}}
	}
	return piece
}

func TestMineshaftBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_mineshaft_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count, minecarts, spawners, protected := 0, 0, 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 21 || f[0] != "mineblocks" {
			t.Fatalf("invalid mineshaft row %q", scanner.Text())
		}
		parse := func(i int) int64 {
			n, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
		seed, mesa, orientation, mode, ax, az, clipMode := parse(2), parse(3) != 0, int(parse(4)), int(parse(5)), int32(parse(6)), int32(parse(7)), int(parse(8))
		piece := mineshaftProbePiece(f[1], mesa, orientation, ax, az)
		world := &mineshaftProbeWorld{hutProbeWorld: hutProbeWorld{scenario: mode, edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}, mesa: mesa, ticks: make(map[FeaturePosition]bool)}
		random := NewWorldgenRandom(seed)
		var pending []PendingStructureEntity
		x0, x1, z0, z1 := piece.Box.MinX>>4, piece.Box.MaxX>>4, piece.Box.MinZ>>4, piece.Box.MaxZ>>4
		if clipMode < 2 {
			x0, x1, z0, z1 = ax>>4, ax>>4, az>>4, az>>4
			if clipMode == 1 {
				x0 += 3
				x1 += 3
			}
		}
		for x := x0; x <= x1; x++ {
			for z := z0; z <= z1; z++ {
				result, err := PostProcessMineshaftBlocks(world, &piece, random, x, z)
				if err != nil {
					t.Fatalf("%s: %v", scanner.Text(), err)
				}
				pending = append(pending, result.PendingEntities...)
			}
		}
		blocksHash, postHash, err := structureBlockDigest(&world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		digest := func(rows []string) string {
			sort.Strings(rows)
			return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
		}
		var entities, ticks, actors []string
		for p, entry := range world.entities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entry.data}).Value.(nbt.Object)
			data["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
			data["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
			data["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
			entities = append(entities, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
		}
		for p := range world.ticks {
			ticks = append(ticks, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
		}
		for _, request := range pending {
			actors = append(actors, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: request.Data}))
			if request.ID != "minecraft:chest_minecart" || request.Persistent {
				t.Fatal("incorrect loot minecart request")
			}
		}
		saved, err := piece.NBT()
		if err != nil {
			t.Fatal(err)
		}
		pieceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))))
		next := int64(random.NextLong())
		if os.Getenv("ATLAS_MINESHAFT_DEBUG") != "" {
			var rows []string
			for p, s := range world.edits {
				id, err := blockstate.BlockStateID(s.Name, s.Properties)
				if err != nil {
					t.Fatal(err)
				}
				rows = append(rows, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
			}
			sort.Strings(rows)
			if err := os.WriteFile(filepath.Join("..", "..", "artifacts", "worldgen-probes", "mineshaft-got.txt"), []byte(strings.Join(rows, "\n")), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if len(world.edits) != int(parse(9)) || blocksHash != f[10] || len(world.post) != int(parse(11)) || postHash != f[12] || len(entities) != int(parse(13)) || digest(entities) != f[14] || len(ticks) != int(parse(15)) || digest(ticks) != f[16] || len(actors) != int(parse(17)) || digest(actors) != f[18] || pieceHash != f[19] || next != parse(20) {
			t.Fatalf("%s: got blocks%d/%s post%d/%s entities%d/%s ticks%d/%s actors%d/%s piece%s next%d", scanner.Text(), len(world.edits), blocksHash, len(world.post), postHash, len(entities), digest(entities), len(ticks), digest(ticks), len(actors), digest(actors), pieceHash, next)
		}
		count++
		minecarts += len(actors)
		spawners += len(entities)
		if mode == 7 {
			protected++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 11520 || minecarts == 0 || spawners == 0 || protected == 0 {
		t.Fatalf("incomplete mineshaft cases %d/%d/%d/%d", count, minecarts, spawners, protected)
	}
	t.Logf("matched %d original mine piece calls, blocks/NBT/supports/clips/RNG with %d loot minecart requests and %d spider spawners", count, minecarts, spawners)
}
