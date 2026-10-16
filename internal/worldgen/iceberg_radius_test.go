package worldgen

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestIcebergRadiiAgainstVanillaJAR(t *testing.T) {
	f, err := os.Open("../../docs/worldgen/iceberg_radius_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	cases := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 8 {
			t.Fatal(scanner.Text())
		}
		n := func(i int) int64 {
			v, err := strconv.ParseInt(fields[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		r := NewWorldgenRandom(n(2))
		y, h, rad := int32(n(3)), int32(n(4)), int32(n(5))
		v := int32(0)
		if fields[1] == "a" {
			v = icebergRoundRadius(r, y, h, rad)
		} else {
			v = icebergSteepRadius(r, y, h, rad)
		}
		next := int64(r.NextLong())
		if v != int32(n(6)) || next != n(7) {
			t.Errorf("%s: got %d %d", scanner.Text(), v, next)
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 2820 {
		t.Fatalf("got %d radii", cases)
	}
}
