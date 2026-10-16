package worldgen

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestStructureBeardifierAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_beardifier_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	box := StructureBoundingBox{-3, 60, 2, 8, 72, 13}
	junctions := []StructureJunction{{5, 65, -2}}
	instances := make(map[string]*StructureBeardifier)
	for _, name := range []string{"none", "bury", "beard_thin", "beard_box", "encapsulate"} {
		for _, delta := range []int32{0, 3, -4} {
			beard, err := NewStructureBeardifier([]StructureRigidPiece{{box, name, delta}}, nil)
			if err != nil {
				t.Fatal(err)
			}
			instances[name+"/"+strconv.Itoa(int(delta))] = beard
		}
	}
	instances["junction/0"], err = NewStructureBeardifier(nil, junctions)
	if err != nil {
		t.Fatal(err)
	}
	instances["mixed/0"], err = NewStructureBeardifier([]StructureRigidPiece{
		{box, "bury", 0}, {box, "beard_thin", 3}, {box, "beard_box", -4}, {box, "encapsulate", 0},
	}, junctions)
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	counts := make(map[string]int)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) < 3 {
			t.Fatalf("invalid Beardifier vector %q", scanner.Text())
		}
		counts[f[0]]++
		switch f[0] {
		case "kernel":
			index, err := strconv.Atoi(f[1])
			if err != nil || index < 0 || index >= len(structureBeardKernel) || len(f) != 3 {
				t.Fatalf("invalid kernel vector %q", scanner.Text())
			}
			want, err := strconv.ParseUint(f[2], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			if got := math.Float32bits(structureBeardKernel[index]); got != uint32(want) {
				t.Fatalf("kernel %d: got %08x, want %08x", index, got, want)
			}
		case "density":
			if len(f) != 7 {
				t.Fatalf("invalid density vector %q", scanner.Text())
			}
			var coordinates [3]int32
			for index := range coordinates {
				value, err := strconv.ParseInt(f[3+index], 10, 32)
				if err != nil {
					t.Fatal(err)
				}
				coordinates[index] = int32(value)
			}
			want, err := strconv.ParseUint(f[6], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			beard := instances[f[1]+"/"+f[2]]
			if beard == nil {
				t.Fatalf("unknown Beardifier scenario %q", scanner.Text())
			}
			if got := math.Float64bits(beard.Sample(coordinates[0], coordinates[1], coordinates[2])); got != want {
				t.Fatalf("%s: got %.17g (%016x), want %.17g (%016x)", scanner.Text(), math.Float64frombits(got), got, math.Float64frombits(want), want)
			}
		default:
			t.Fatalf("unknown Beardifier vector %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if counts["kernel"] != 13824 || counts["density"] != 74800 {
		t.Fatalf("unexpected Beardifier vector coverage %v", counts)
	}
	t.Logf("matched %d official Beardifier vectors bit for bit", counts["kernel"]+counts["density"])
}

func TestStructureBeardifierChunkInputs(t *testing.T) {
	box := StructureBoundingBox{-3, 60, 2, 8, 72, 13}
	terrain := []StructureTerrainPiece{
		{RigidPiece: StructureRigidPiece{box, "beard_box", 3}, Jigsaw: true, Projection: "rigid",
			Junctions: []StructureJunction{{-12, 60, 0}, {-11, 60, 0}, {0, 60, 26}, {0, 60, 27}}},
		{RigidPiece: StructureRigidPiece{box, "bury", 9}},
		{RigidPiece: StructureRigidPiece{box, "beard_thin", 3}, Jigsaw: true, Projection: "terrain_matching",
			Junctions: []StructureJunction{{0, 61, 0}}},
		{RigidPiece: StructureRigidPiece{box, "none", 0}, Jigsaw: true, Projection: "rigid",
			Junctions: []StructureJunction{{0, 62, 0}}},
		{RigidPiece: StructureRigidPiece{StructureBoundingBox{28, 60, 0, 30, 72, 13}, "bury", 0}},
	}
	beard, err := NewStructureBeardifierForChunk(0, 0, terrain)
	if err != nil {
		t.Fatal(err)
	}
	if len(beard.pieces) != 2 || beard.pieces[0].GroundLevelDelta != 3 || beard.pieces[1].GroundLevelDelta != 0 || len(beard.junctions) != 3 {
		t.Fatalf("unexpected chunk Beardifier inputs: %+v", beard)
	}
	before := beard.Sample(0, 60, 0)
	terrain[0].RigidPiece.Box.MinY = 100
	terrain[0].Junctions[1].SourceGroundY = 100
	if got := beard.Sample(0, 60, 0); got != before {
		t.Fatal("Beardifier changed after caller mutated inputs")
	}
	var empty *StructureBeardifier
	if empty.Sample(0, 60, 0) != 0 {
		t.Fatal("empty structure context contributed density")
	}
}
