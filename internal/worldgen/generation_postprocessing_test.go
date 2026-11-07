package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

func scanPostprocessingReference(t *testing.T, path string, visit func([]byte)) {
	t.Helper()
	f, err := os.Open(path)
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
	scan.Buffer(make([]byte, 65536), 8<<20)
	for scan.Scan() {
		visit(scan.Bytes())
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
}

func packedPostprocessing(chunk *NoiseChunkResult) [][]int {
	sections := make([][]int, chunk.Height/16)
	for _, index := range chunk.PostProcessingOrder {
		sections[index/4096] = append(sections[index/4096], index&15|(index>>8&15)<<4|(index>>4&15)<<8)
	}
	return sections
}

func TestTerrainPostprocessingAgainstVanilla(t *testing.T) {
	cases, entries, repeats := 0, 0, 0
	scanPostprocessingReference(t, "../../testdata/worldgen/terrain_postprocessing_1.21.1.jsonl.gz", func(data []byte) {
		var row struct {
			Seed     int64
			X, Z     int32
			Stage    string
			Sections [][]int
		}
		if err := json.Unmarshal(data, &row); err != nil {
			t.Fatal(err)
		}
		g, err := NewOverworldNoiseGenerator(row.Seed)
		if err != nil {
			t.Fatal(err)
		}
		c, err := g.GenerateNoise(row.X, row.Z)
		if err != nil {
			t.Fatal(err)
		}
		if row.Stage == "carved" {
			biomes, err := NewOverworldBiomeSampler(row.Seed)
			if err != nil {
				t.Fatal(err)
			}
			if err := g.ApplySurface(c, biomes.BiomeAtQuart); err != nil {
				t.Fatal(err)
			}
			if err := g.ApplyCarvers(c, biomes.BiomeAtQuart); err != nil {
				t.Fatal(err)
			}
		} else if row.Stage != "noise" {
			t.Fatal("unknown stage", row.Stage)
		}
		got := packedPostprocessing(c)
		if len(got) != len(row.Sections) {
			t.Fatal("section bounds", len(got), len(row.Sections))
		}
		for section, want := range row.Sections {
			if !slices.Equal(got[section], want) {
				t.Fatalf("seed=%d chunk=%d,%d stage=%s section=%d: got %v want %v", row.Seed, row.X, row.Z, row.Stage, section, got[section], want)
			}
			seen := map[int]bool{}
			for _, p := range want {
				if seen[p] {
					repeats++
				}
				seen[p] = true
			}
			entries += len(want)
		}
		cases++
	})
	if cases != 24 || entries == 0 {
		t.Fatal("incomplete native terrain lists", cases, entries)
	}
	t.Logf("matched %d original stage lists: %d entries, %d repeats", cases, entries, repeats)
}

func TestFeaturePostprocessingAgainstVanilla(t *testing.T) {
	templates := map[[4]int32]*OverworldFeatureRegion{}
	plans := map[string]configuredFeaturePlacer{}
	cases, entries, repeats := 0, 0, 0
	scanPostprocessingReference(t, "../../testdata/worldgen/features_postprocessing_1.21.1.jsonl.gz", func(data []byte) {
		var row struct {
			Feature, Hash string
			Seed, Next    int64
			Scene, Count  int
			Origin        [3]int32
			Placed        bool
			Sections      map[string][]int
		}
		if err := json.Unmarshal(data, &row); err != nil {
			t.Fatal(err)
		}
		place := plans[row.Feature]
		if place == nil {
			var err error
			place, err = compileFeaturePlan(row.Feature, 0)
			if err != nil {
				t.Fatal(err)
			}
			plans[row.Feature] = place
		}
		p := FeaturePosition{row.Origin[0], row.Origin[1], row.Origin[2]}
		key := [4]int32{p.X, p.Y, p.Z, int32(row.Scene)}
		template := templates[key]
		if template == nil {
			template = frozenSpringProbeRegion(t, p, row.Scene)
			templates[key] = template
		}
		var chunks []*NoiseChunkResult
		for _, c := range template.chunks {
			chunks = append(chunks, c)
		}
		r, err := NewOverworldFeatureRegion(chunks...)
		if err != nil {
			t.Fatal(err)
		}
		r.borrowed, r.biomeAtQuart, r.biomeSeed = template.borrowed, template.biomeAtQuart, row.Seed
		random := NewWorldgenRandom(row.Seed)
		placed, err := place(r, random, p)
		if err != nil {
			t.Fatal(err)
		}
		var output []string
		actual := map[string][]int{}
		for key, overlay := range r.overlays {
			for index, s := range overlay.blocks {
				id, err := blockstate.BlockStateID(s.Name, s.Properties)
				if err != nil {
					t.Fatal(err)
				}
				output = append(output, fmt.Sprintf("%d %d %d %d", key.chunkX*16+int32(index&15), -64+int32(index/256), key.chunkZ*16+int32((index>>4)&15), id))
			}
			for section, list := range packedPostprocessing(r.cloneChunk(key.chunkX, key.chunkZ)) {
				if len(list) > 0 {
					actual[fmt.Sprintf("%d %d %d", key.chunkX, key.chunkZ, section)] = list
				}
			}
		}
		sort.Strings(output)
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(output, "\n"))))
		next := int64(random.NextLong())
		if placed != row.Placed || len(output) != row.Count || hash != row.Hash || next != row.Next {
			t.Fatalf("%s seed=%d scene=%d: feature output/RNG changed", row.Feature, row.Seed, row.Scene)
		}
		if !reflect.DeepEqual(actual, row.Sections) {
			for key, want := range row.Sections {
				if !slices.Equal(actual[key], want) {
					t.Fatalf("%s seed=%d scene=%d section=%s: got %v want %v", row.Feature, row.Seed, row.Scene, key, actual[key], want)
				}
			}
			t.Fatalf("%s seed=%d scene=%d: unexpected marked sections", row.Feature, row.Seed, row.Scene)
		}
		for _, list := range row.Sections {
			seen := map[int]bool{}
			for _, p := range list {
				if seen[p] {
					repeats++
				}
				seen[p] = true
			}
			entries += len(list)
		}
		cases++
	})
	if cases != 10752 || entries == 0 || repeats == 0 {
		t.Fatal("incomplete native feature lists", cases, entries, repeats)
	}
	t.Logf("matched %d original placements: %d ordered entries, %d repeats", cases, entries, repeats)
}
