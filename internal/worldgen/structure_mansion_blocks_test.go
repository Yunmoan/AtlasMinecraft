package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type mansionProbeWorld struct {
	*hutProbeWorld
	water bool
	ticks map[FeaturePosition]bool
}

func (w *mansionProbeWorld) FeatureRawBrightness(p FeaturePosition, skyDarkening int32) (int32, error) {
	return max(0, 15-skyDarkening), nil
}
func (w *mansionProbeWorld) MinBuildHeight() int32          { return -64 }
func (w *mansionProbeWorld) MaxBuildHeight() int32          { return 320 }
func (w *mansionProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }
func (w *mansionProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w.edits[p]; ok {
		return state, nil
	}
	if p.Y <= 63 {
		return structureState("stone"), nil
	}
	if w.water && p.Y <= 110 {
		return structureState("water"), nil
	}
	return structureState("air"), nil
}
func (w *mansionProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	w.ticks[p] = true
	return nil
}
func TestMansionBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../testdata/worldgen/structure_mansion_blocks_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count, actors := 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 17 || f[0] != "mansion_blocks" {
			t.Fatal("invalid mansion fixture", scanner.Text())
		}
		parse := func(i int) int64 {
			n, e := strconv.ParseInt(f[i], 10, 64)
			if e != nil {
				t.Fatal(e)
			}
			return n
		}
		piece, err := NewMansionTemplatePiece(f[5], FeaturePosition{7, 70, -9}, TemplateRotation(parse(2)), TemplateMirror(parse(3)))
		if err != nil {
			t.Fatal(err)
		}
		world := &mansionProbeWorld{hutProbeWorld: &hutProbeWorld{edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}, water: parse(4) != 0, ticks: make(map[FeaturePosition]bool)}
		random := NewWorldgenRandom(parse(1))
		regionRandom := NewXoroshiro(parse(1) ^ 0x7654321)
		clip := StructureBoundingBox{-128, -64, -128, 128, 319, 128}
		if parse(6) == 1 {
			clip = StructureBoundingBox{0, -64, -16, 15, 319, -1}
		}
		placed, err := placeMansionTemplateBlocks(world, piece, random, regionRandom, clip)
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		blockHash, _, err := structureBlockDigest(world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		var entities, actorRows []string
		for p, entity := range world.entities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: entity.data}).Value.(nbt.Object)
			data["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
			data["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
			data["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
			entities = append(entities, fmt.Sprintf("%d %d %d %s", p.X, p.Y, p.Z, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data})))
		}
		for _, actor := range placed.PendingEntities {
			data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: actor.Data}).Value.(nbt.Object)
			data["attributes"] = canonicalOutpostActorAttributes(data["attributes"])
			actorRows = append(actorRows, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data}))
		}
		sort.Strings(entities)
		sort.Strings(actorRows)
		var ticks []string
		for p := range world.ticks {
			ticks = append(ticks, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
		}
		sort.Strings(ticks)
		next, regionNext := int64(random.NextLong()), int64(regionRandom.NextLong())
		entityHash, actorHash, tickHash := stringHexHash(strings.Join(entities, "\n")), stringHexHash(strings.Join(actorRows, "\n")), stringHexHash(strings.Join(ticks, "\n"))
		if len(world.edits) != int(parse(7)) || blockHash != f[8] || len(entities) != int(parse(9)) || entityHash != f[10] || len(ticks) != int(parse(11)) || tickHash != f[12] || len(actorRows) != int(parse(13)) || actorHash != f[14] || next != parse(15) || regionNext != parse(16) {
			for p, s := range world.edits {
				id, _ := blockstate.BlockStateID(s.Name, s.Properties)
				t.Logf("block %d %d %d %d", p.X, p.Y, p.Z, id)
			}
			t.Fatalf("%s:\n got blocks %d/%s entities %d/%s ticks %d/%s actors %d/%s RNG %d/%d; entities %v actors %v", scanner.Text(), len(world.edits), blockHash, len(entities), entityHash, len(ticks), tickHash, len(actorRows), actorHash, next, regionNext, entities, actorRows)
		}
		count++
		actors += len(actorRows)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 7008 || actors == 0 {
		t.Fatalf("incomplete mansion fixtures: %d calls, %d actors", count, actors)
	}
	t.Logf("matched %d original mansion template calls and %d marker actors, blocks/full NBT/rotations/air overwrite/liquids/clips/RNG", count, actors)
}
