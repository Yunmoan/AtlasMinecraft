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

type strongholdProbeWorld struct {
	hutProbeWorld
	ticks map[FeaturePosition]bool
}

func (w *strongholdProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if s, found := w.edits[p]; found {
		return s, nil
	}
	floor := int32(80)
	switch w.scenario {
	case 1:
		mod := (p.X + 2*p.Z) % 5
		if mod < 0 {
			mod += 5
		}
		floor = 42 + mod
	case 2:
		floor = 25
	case 3:
		floor = -65
	}
	if p.Y <= floor {
		return structureState("stone"), nil
	}
	if w.scenario == 2 && p.Y <= 60 {
		return structureState("water"), nil
	}
	return structureState("air"), nil
}
func (w *strongholdProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }
func (w *strongholdProbeWorld) MinBuildHeight() int32          { return -64 }
func (w *strongholdProbeWorld) MaxBuildHeight() int32          { return 320 }
func (w *strongholdProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	w.ticks[p] = true
	return nil
}

func TestStrongholdBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_stronghold_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	sizes := map[string][3]int32{"shs": {5, 5, 7}, "shph": {9, 5, 11}, "shlt": {5, 5, 5}, "shrt": {5, 5, 5}, "shrc": {11, 7, 11}, "shssd": {5, 11, 8}, "shsd": {5, 11, 5}, "sh5c": {10, 9, 11}, "shcc": {5, 5, 7}, "shli": {14, 11, 15}, "shli_small": {14, 6, 15}, "shpr": {11, 8, 16}, "shfc": {5, 5, 3}, "shstart": {5, 11, 5}}
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 19 || f[0] != "stronghold_block" {
			t.Fatalf("bad stronghold row %q", scanner.Text())
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
		piece := StructurePiece{ID: "minecraft:" + strings.TrimSuffix(f[1], "_small"), Orientation: []string{"south", "west", "north", "east"}[orientation], Box: StructureBoundingBox{ax, 40, az, ax + sx - 1, 40 + sy - 1, az + sz - 1}, Data: make(nbt.Object)}
		scenario := int(parse(4))
		piece.Data["EntryDoor"] = nbt.Tag{Type: nbt.String, Value: []string{"OPENING", "WOOD_DOOR", "GRATES", "IRON_DOOR"}[(scenario+orientation)%4]}
		switch f[1] {
		case "shs":
			strongholdSetBool(piece.Data, "Left", scenario&1 == 0)
			strongholdSetBool(piece.Data, "Right", scenario&2 == 0)
		case "shrc":
			kind := scenario
			if seed == 1234 {
				kind++
			}
			piece.Data["Type"] = nbt.Tag{Type: nbt.Int, Value: int32(kind % 5)}
		case "shsd", "shstart":
			strongholdSetBool(piece.Data, "Source", f[1] == "shstart")
		case "sh5c":
			strongholdSetBool(piece.Data, "leftLow", scenario&1 == 0)
			strongholdSetBool(piece.Data, "leftHigh", scenario&2 == 0)
			strongholdSetBool(piece.Data, "rightLow", scenario&1 != 0)
			strongholdSetBool(piece.Data, "rightHigh", scenario&2 != 0)
		case "shcc":
			strongholdSetBool(piece.Data, "Chest", seed == 1234)
		case "shli", "shli_small":
			strongholdSetBool(piece.Data, "Tall", sy > 6)
		case "shpr":
			strongholdSetBool(piece.Data, "Mob", seed == 1234)
		case "shfc":
			piece.Data["Steps"] = nbt.Tag{Type: nbt.Int, Value: int32(3)}
		}
		world := &strongholdProbeWorld{hutProbeWorld: hutProbeWorld{scenario: int(parse(4)), edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}, ticks: make(map[FeaturePosition]bool)}
		var random StructurePieceRandom = NewWorldgenRandom(seed)
		if f[1] == "shpr" && seed == 99 {
			random = strongholdAllEyesRandom{random}
		}
		_, err := PostProcessStrongholdBlocks(world, &piece, random, int32(parse(7)), int32(parse(8)))
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
		saved, err := piece.NBT()
		if err != nil {
			t.Fatal(err)
		}
		flags := stringHexHash(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: saved}))
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

// This oracle input deliberately exercises the original extremely rare
// twelve-eye portal branch while consuming every original float draw.
type strongholdAllEyesRandom struct{ StructurePieceRandom }

func (random strongholdAllEyesRandom) NextFloat() float32 {
	random.StructurePieceRandom.NextFloat()
	return float32(0.95)
}
