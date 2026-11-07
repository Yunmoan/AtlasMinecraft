package worldgen

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

type mansionFoundationWorld struct {
	hutProbeWorld
	floorY int32
}

func (w *mansionFoundationWorld) MinBuildHeight() int32 { return -64 }
func (w *mansionFoundationWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w.edits[p]; ok {
		return state, nil
	}
	mod := func(n, d int32) int32 {
		v := n % d
		if v < 0 {
			v += d
		}
		return v
	}
	if p.Y == w.floorY {
		if mod(p.X+p.Z, 3) == 0 {
			return structureState("air"), nil
		}
		return structureState("dark_oak_planks"), nil
	}
	floor := int32(60)
	switch w.scenario {
	case 1:
		floor = 30
	case 2:
		floor = 47 + mod(p.X*3+p.Z, 7)
	case 3:
		floor = -65
	}
	if p.Y <= floor {
		return structureState("stone"), nil
	}
	if (w.scenario == 1 || w.scenario == 2) && p.Y < w.floorY {
		return structureState("water"), nil
	}
	return structureState("air"), nil
}

func TestMansionFoundationsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/structure_mansion_foundation_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		if len(f) != 7 || f[0] != "mansion_foundation" {
			t.Fatal("invalid foundation vector")
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		pieces, err := assembleMansion(NewLegacyRandom(parse(1)), FeaturePosition{7, 70, -9}, TemplateRotation(parse(2)))
		if err != nil {
			t.Fatal(err)
		}
		start := StructureStart{Structure: "minecraft:mansion", Pieces: pieces}
		box, _ := start.BoundingBox()
		clip := StructureBoundingBox{box.MinX - 5, -64, box.MinZ - 5, box.MaxX + 5, 319, box.MaxZ + 5}
		switch parse(4) {
		case 1:
			clip = StructureBoundingBox{0, -64, -16, 15, 319, -1}
		case 2:
			x, z := (box.MinX>>4)*16, (box.MinZ>>4)*16
			clip = StructureBoundingBox{x, -64, z, x + 15, 319, z + 15}
		}
		world := &mansionFoundationWorld{hutProbeWorld: hutProbeWorld{scenario: int(parse(3)), edits: make(map[FeaturePosition]SurfaceBlockState)}, floorY: box.MinY}
		if err := PostProcessMansionFoundation(world, &start, clip); err != nil {
			t.Fatal(err)
		}
		hash, _, err := structureBlockDigest(&world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		if len(world.edits) != int(parse(5)) || hash != f[6] {
			t.Fatalf("%s: got %d/%s", line, len(world.edits), hash)
		}
		cases++
	}
	if cases != 144 {
		t.Fatal("incomplete foundation corpus", cases)
	}
	t.Logf("matched %d original full/clipped mansion foundations across air, water, uneven ground and void", cases)
}
