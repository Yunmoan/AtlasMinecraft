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

func TestTemplateEntityTransformsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_entity_transforms_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	positions, angles := 0, 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		bits := func(index int) uint64 {
			value, err := strconv.ParseUint(fields[index], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		integer := func(index int) int32 {
			value, err := strconv.ParseInt(fields[index], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			return int32(value)
		}
		switch {
		case len(fields) == 12 && fields[0] == "position":
			p := [3]float64{math.Float64frombits(bits(1)), math.Float64frombits(bits(2)), math.Float64frombits(bits(3))}
			pivot := FeaturePosition{integer(4), integer(5), integer(6)}
			out := transformTemplateEntityPosition(p, pivot, TemplateMirror(integer(7)), TemplateRotation(integer(8)))
			for axis, value := range out {
				if math.Float64bits(value) != bits(9+axis) {
					t.Fatalf("%s: axis %d got %016x", scanner.Text(), axis, math.Float64bits(value))
				}
			}
			positions++
		case len(fields) == 5 && fields[0] == "yaw":
			out := transformTemplateEntityYaw(math.Float32frombits(uint32(bits(1))), TemplateMirror(integer(2)), TemplateRotation(integer(3)))
			if math.Float32bits(out) != uint32(bits(4)) {
				t.Fatalf("%s: yaw got %08x", scanner.Text(), math.Float32bits(out))
			}
			angles++
		default:
			t.Fatalf("invalid entity transform vector %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if positions != 480 || angles != 144 {
		t.Fatalf("got %d position/%d yaw vectors", positions, angles)
	}
	t.Logf("matched %d double-position and %d float-yaw vectors, exact bits", positions, angles)
}
