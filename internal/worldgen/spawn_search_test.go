package worldgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixture is produced by docs/worldgen/SpawnSearchProbe against the
// reference JAR; this test pins the Go port of Climate$SpawnFinder to it.
func TestFindSpawnPositionMatchesJarSearch(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "worldgen", "spawn_search_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var seed int64
		var wantX, wantZ int32
		if _, err := fmt.Sscanf(line, "spawn_search %d %d %d", &seed, &wantX, &wantZ); err != nil {
			t.Fatalf("fixture line %q: %v", line, err)
		}
		gotX, gotZ, err := FindSpawnPosition(seed)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if gotX != wantX || gotZ != wantZ {
			t.Fatalf("seed %d: spawn search (%d, %d), JAR reports (%d, %d)", seed, gotX, gotZ, wantX, wantZ)
		}
	}
}
