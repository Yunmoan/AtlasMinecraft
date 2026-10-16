package worldgen

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestBastionGenerationStubAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_bastion_stub_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count, accepted := 0, 0
	templates, rotations := make(map[string]bool), make(map[TemplateRotation]bool)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 21 || f[0] != "bastion_stub" {
			t.Fatalf("invalid Bastion stub vector %q", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(f[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		seed, cx, cz := parse(1), int32(parse(2)), int32(parse(3))
		stub, err := findBastionGenerationStub(seed, cx, cz)
		if err != nil {
			t.Fatal(err)
		}
		want := BastionGenerationStub{
			Template: f[6], Rotation: TemplateRotation(parse(7)),
			Position: FeaturePosition{int32(parse(8)), int32(parse(9)), int32(parse(10))},
			Origin:   FeaturePosition{int32(parse(11)), int32(parse(12)), int32(parse(13))},
			Box:      StructureBoundingBox{MinX: int32(parse(14)), MinY: int32(parse(15)), MinZ: int32(parse(16)), MaxX: int32(parse(17)), MaxY: int32(parse(18)), MaxZ: int32(parse(19))},
		}
		if *stub != want {
			t.Fatalf("%s: got %+v, want %+v", scanner.Text(), *stub, want)
		}
		// The pre-assembly stream has consumed rotation and start-pool selection.
		base := NewLegacyRandom(seed)
		xm, zm := base.NextLong(), base.NextLong()
		random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
		random.NextIntBound(4)
		random.NextIntBound(4)
		if next := int64(random.NextLong()); next != parse(20) {
			t.Fatalf("%s: RNG continuation %d", scanner.Text(), next)
		}
		calls := 0
		checked, err := GenerateBastionGenerationStub(seed, cx, cz, func(x, y, z int32) (string, error) {
			calls++
			if got := (FeaturePosition{x, y, z}); got != (FeaturePosition{want.Position.X >> 2, want.Position.Y >> 2, want.Position.Z >> 2}) {
				t.Fatalf("original Bastion biome lookup at wrong quart %+v", got)
			}
			return f[4], nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if calls != 1 || (checked != nil) != (parse(5) != 0) {
			t.Fatalf("%s: incorrect biome acceptance", scanner.Text())
		}
		if checked != nil {
			accepted++
		}
		templates[stub.Template], rotations[stub.Rotation] = true, true
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 480 || accepted != 384 || len(templates) != 4 || len(rotations) != 4 {
		t.Fatalf("incomplete Bastion oracle: %d calls, %d accepted, %d templates, %d rotations", count, accepted, len(templates), len(rotations))
	}
	t.Logf("matched %d original Bastion pre-assembly stubs, %d accepted, all templates and rotations", count, accepted)
}
