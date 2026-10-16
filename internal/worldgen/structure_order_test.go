package worldgen

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestStructureOrderAndRegionRandomAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_order_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	orders, randoms := 0, 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(fields[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		switch fields[0] {
		case "order":
			if len(fields) != 4 {
				t.Fatalf("invalid order vector %q", scanner.Text())
			}
			order, err := OriginalStructureDecorationOrder(fields[1])
			if err != nil || order.Step != int32(parse(2)) || order.Index != int32(parse(3)) {
				t.Fatalf("%s: got %+v (%v)", scanner.Text(), order, err)
			}
			orders++
		case "region_random":
			if len(fields) != 12 {
				t.Fatalf("invalid random vector %q", scanner.Text())
			}
			random := NewStructureRegionRandom(parse(1), int32(parse(2)), int32(parse(3)))
			for index := 4; index < 12; index++ {
				if got := int64(random.NextLong()); got != parse(index) {
					t.Fatalf("%s: random draw %d got %d", scanner.Text(), index-4, got)
				}
			}
			randoms++
		default:
			t.Fatalf("unknown structure order vector %q", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if orders != 34 || randoms != 48 {
		t.Fatalf("got %d orders and %d region randoms, want 34 and 48", orders, randoms)
	}
}
