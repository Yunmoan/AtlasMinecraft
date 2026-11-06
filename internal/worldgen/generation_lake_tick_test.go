package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestLakeGenerationTicksAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/lake_ticks_1.21.1.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	scan := bufio.NewScanner(gz)
	scan.Buffer(make([]byte, 65536), 1<<20)
	templates := map[[4]int32]*OverworldFeatureRegion{}
	place, err := compileFeaturePlan("minecraft:lake_lava", 0)
	if err != nil {
		t.Fatal(err)
	}
	cases, requests := 0, 0
	for scan.Scan() {
		var row struct {
			Seed, Next   int64
			Scene, Count int
			Origin       [3]int32
			Placed       bool
			Hash         string
			Requests     []string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		p := FeaturePosition{row.Origin[0], row.Origin[1], row.Origin[2]}
		key := [4]int32{p.X, p.Y, p.Z, int32(row.Scene)}
		template := templates[key]
		if template == nil {
			template = frozenSpringProbeRegion(t, p, row.Scene)
			templates[key] = template
		}
		chunks := []*NoiseChunkResult{}
		for _, c := range template.chunks {
			chunks = append(chunks, c)
		}
		r, err := NewOverworldFeatureRegion(chunks...)
		if err != nil {
			t.Fatal(err)
		}
		r.borrowed, r.biomeAtQuart = template.borrowed, template.biomeAtQuart
		r.biomeSeed = row.Seed
		// Structure recording observes the real ordered generation queue writes.
		r.structureRecording = &structureBlockResult{blocks: map[FeaturePosition]SurfaceBlockState{}, postProcessing: map[FeaturePosition]bool{}}
		random := NewWorldgenRandom(row.Seed)
		placed, err := place(r, random, p)
		if err != nil {
			t.Fatal(err)
		}
		var output, actual []string
		for key, overlay := range r.overlays {
			for index, s := range overlay.blocks {
				id, err := blockstate.BlockStateID(s.Name, s.Properties)
				if err != nil {
					t.Fatal(err)
				}
				output = append(output, fmt.Sprintf("%d %d %d %d", key.chunkX*16+int32(index&15), -64+int32(index/256), key.chunkZ*16+int32((index>>4)&15), id))
			}
		}
		for _, tick := range r.structureRecording.blockTicks.ticks {
			actual = append(actual, fmt.Sprintf("%s %d %d %d 0 %d", tick.Block, tick.Position.X, tick.Position.Y, tick.Position.Z, tick.Priority))
		}
		sort.Strings(output)
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(output, "\n"))))
		next := int64(random.NextLong())
		if placed != row.Placed || len(output) != row.Count || hash != row.Hash || next != row.Next || !slices.Equal(actual, row.Requests) {
			t.Fatalf("seed=%d scene=%d: placed=%v count=%d hash=%s next=%d ticks=%d; want %v %d %s %d %d", row.Seed, row.Scene, placed, len(output), hash, next, len(actual), row.Placed, row.Count, row.Hash, row.Next, len(row.Requests))
		}
		cases++
		requests += len(actual)
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 512 || requests != 44550 {
		t.Fatal("incomplete lake reference", cases, requests)
	}
	t.Logf("matched %d original lake placements and %d ordered block requests", cases, requests)
}
