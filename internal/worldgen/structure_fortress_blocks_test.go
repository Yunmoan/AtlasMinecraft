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

type fortressProbeWorld struct {
	hutProbeWorld
	ticks map[FeaturePosition]bool
}

func (w *fortressProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if s, found := w.edits[p]; found {
		return s, nil
	}
	floor := int32(45)
	switch w.scenario {
	case 1:
		mod := (p.X + 2*p.Z) % 5
		if mod < 0 {
			mod += 5
		}
		floor = 52 + mod
	case 2:
		floor = 30
	case 3:
		floor = -1
	}
	if p.Y <= floor {
		return structureState("stone"), nil
	}
	if w.scenario == 2 && p.Y <= 66 {
		return structureState("lava"), nil
	}
	return structureState("air"), nil
}
func (w *fortressProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= 0 && p.Y < 256 }
func (w *fortressProbeWorld) MinBuildHeight() int32          { return 0 }
func (w *fortressProbeWorld) MaxBuildHeight() int32          { return 256 }
func (w *fortressProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	w.ticks[p] = true
	return nil
}

func TestNetherFortressBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_fortress_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	sizes := map[string][3]int32{"nebs": {5, 10, 19}, "nebcr": {19, 10, 19}, "nerc": {7, 9, 7}, "nesr": {7, 11, 7}, "nemt": {7, 8, 9}, "nece": {13, 14, 13}, "nesc": {5, 7, 5}, "nescsc": {5, 7, 5}, "nescrt": {5, 7, 5}, "nesclt": {5, 7, 5}, "neccs": {5, 14, 10}, "nectb": {9, 7, 9}, "necsr": {13, 14, 13}, "nebef": {5, 10, 8}}
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 19 || f[0] != "fortress" {
			t.Fatalf("bad fortress row %q", scanner.Text())
		}
		parse := func(i int) int64 {
			value, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		seed, orientation := parse(2), int(parse(3))
		dims := sizes[f[1]]
		sx, sy, sz := dims[0], dims[1], dims[2]
		if orientation%2 != 0 {
			sx, sz = sz, sx
		}
		ax, az := int32(parse(5)), int32(parse(6))
		piece := StructurePiece{ID: "minecraft:" + f[1], Orientation: []string{"south", "west", "north", "east"}[orientation], Box: StructureBoundingBox{ax, 64, az, ax + sx - 1, 64 + sy - 1, az + sz - 1}, Data: make(nbt.Object)}
		switch f[1] {
		case "nebef":
			piece.Data["Seed"] = nbt.Tag{Type: nbt.Int, Value: int32(seed)}
		case "nescrt", "nesclt":
			flag := int8(0)
			if seed != 0 {
				flag = 1
			}
			piece.Data["Chest"] = nbt.Tag{Type: nbt.Byte, Value: flag}
		case "nemt":
			flag := int8(0)
			if seed == 1234 {
				flag = 1
			}
			piece.Data["Mob"] = nbt.Tag{Type: nbt.Byte, Value: flag}
		}
		world := &fortressProbeWorld{hutProbeWorld: hutProbeWorld{scenario: int(parse(4)), edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}, ticks: make(map[FeaturePosition]bool)}
		random := NewWorldgenRandom(seed)
		_, err := PostProcessNetherFortressBlocks(world, &piece, random, int32(parse(7)), int32(parse(8)))
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		blocks, post, err := structureBlockDigest(&world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		var entities, ticks []string
		for p, entity := range world.entities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
			data["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
			data["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
			data["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
			entities = append(entities, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
		}
		for p := range world.ticks {
			ticks = append(ticks, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
		}
		digest := func(rows []string) string {
			sort.Strings(rows)
			return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
		}
		flags := ""
		for _, name := range []string{"Chest", "Mob"} {
			flag, _ := piece.Data[name].Number()
			if flag != 0 {
				flags += "1"
			} else {
				flags += "0"
			}
		}
		next := int64(random.NextLong())
		entityHash, tickHash := digest(entities), digest(ticks)
		if len(world.edits) != int(parse(9)) || blocks != f[10] || len(world.post) != int(parse(11)) || post != f[12] || len(entities) != int(parse(13)) || entityHash != f[14] || len(ticks) != int(parse(15)) || tickHash != f[16] || flags != f[17] || next != parse(18) {
			t.Fatalf("%s\n got blocks %d/%s post %d/%s entities %d/%s ticks %d/%s flags %s next %d\nentity rows %v", scanner.Text(), len(world.edits), blocks, len(world.post), post, len(entities), entityHash, len(ticks), tickHash, flags, next, entities)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 8064 {
		t.Fatalf("got %d block cases want 8064", count)
	}
}
