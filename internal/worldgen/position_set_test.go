package worldgen

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestPositionSetAgainstJavaHashSet(t *testing.T) {
	f, err := os.Open("../../docs/worldgen/position_set_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	mode := ""
	var set javaPositionSet
	cases := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 8 {
			t.Fatal(scanner.Text())
		}
		if fields[1] != mode {
			set = javaPositionSet{}
			mode = fields[1]
		}
		n := func(i int) int32 {
			v, err := strconv.ParseInt(fields[i], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			return int32(v)
		}
		p := FeaturePosition{n(3), n(4), n(5)}
		if fields[2] == "add" {
			set.add(p)
		} else {
			got := set.pop()
			if got != p {
				t.Fatalf("%s: popped %+v", scanner.Text(), got)
			}
		}
		var values []string
		for _, q := range set.positions() {
			values = append(values, fmt.Sprintf("%d %d %d", q.X, q.Y, q.Z))
		}
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(values, "\n"))))
		if set.size != int(n(6)) || hash != fields[7] {
			t.Fatalf("%s: got size=%d hash=%s", scanner.Text(), set.size, hash)
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases < 1000 {
		t.Fatalf("only %d transitions", cases)
	}
}
