package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type endCityProbeWorld struct {
	*hutProbeWorld
	water bool
}

func (w *endCityProbeWorld) MinBuildHeight() int32          { return 0 }
func (w *endCityProbeWorld) MaxBuildHeight() int32          { return 256 }
func (w *endCityProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= 0 && p.Y < 256 }
func (w *endCityProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w.edits[p]; ok {
		return state, nil
	}
	if p.Y <= 63 {
		return structureState("end_stone"), nil
	}
	if w.water && p.Y <= 110 {
		return structureState("water"), nil
	}
	return structureState("air"), nil
}
func TestEndCityBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/structure_end_city_blocks_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count, actors := 0, 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 18 || f[0] != "end_city_blocks" {
			t.Fatal("invalid End City fixture", scanner.Text())
		}
		parse := func(i int) int64 {
			n, e := strconv.ParseInt(f[i], 10, 64)
			if e != nil {
				t.Fatal(e)
			}
			return n
		}
		piece, err := NewEndCityTemplatePiece(f[5], FeaturePosition{int32(parse(6)), int32(parse(7)), int32(parse(8))}, TemplateRotation(parse(2)), parse(3) != 0)
		if err != nil {
			t.Fatal(err)
		}
		world := &endCityProbeWorld{hutProbeWorld: &hutProbeWorld{edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}, water: parse(4) != 0}
		random := NewWorldgenRandom(parse(1))
		placed, err := PostProcessEndCityBlocks(world, piece, random, int32(parse(9)), int32(parse(10)))
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
			data["UUID"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{0, 0, 0, 0}}
			actorRows = append(actorRows, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: data}))
		}
		sort.Strings(entities)
		sort.Strings(actorRows)
		next := int64(random.NextLong())
		if len(world.edits) != int(parse(11)) || blockHash != f[12] || len(entities) != int(parse(13)) || stringHexHash(strings.Join(entities, "\n")) != f[14] || len(actorRows) != int(parse(15)) || stringHexHash(strings.Join(actorRows, "\n")) != f[16] || next != parse(17) {
			t.Fatalf("%s:\n got blocks %d/%s entities %d/%s actors %d/%s RNG %d; entities %v actors %v", scanner.Text(), len(world.edits), blockHash, len(entities), stringHexHash(strings.Join(entities, "\n")), len(actorRows), stringHexHash(strings.Join(actorRows, "\n")), next, entities, actorRows)
		}
		count++
		actors += len(actorRows)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 3088 || actors == 0 {
		t.Fatalf("incomplete End City fixtures: %d calls, %d actors", count, actors)
	}
	t.Logf("matched %d original End City template calls and %d marker actors, blocks/full NBT/rotations/air overwrite/liquids/clips/RNG", count, actors)
}
