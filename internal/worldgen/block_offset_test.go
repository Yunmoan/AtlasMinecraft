package worldgen

import (
	"bufio"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestBlockOffsetAgainstVanillaJAR(t *testing.T) {
	scan := bufio.NewScanner(strings.NewReader(blockOffsetText))
	cases := 0
	for scan.Scan() {
		if !strings.HasPrefix(scan.Text(), "block_offset ") {
			continue
		}
		var label, name string
		var p FeaturePosition
		var want [3]float64
		if _, err := fmt.Sscan(scan.Text(), &label, &name, &p.X, &p.Y, &p.Z, &want[0], &want[1], &want[2]); err != nil {
			t.Fatal(err)
		}
		got := BlockOffset(name, p)
		for i := range got {
			if math.Abs(got[i]-want[i]) > 1e-12 {
				t.Fatalf("%s: got %+v", scan.Text(), got)
			}
		}
		cases++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 132 {
		t.Fatal("incomplete block offset fixture", cases)
	}
}
