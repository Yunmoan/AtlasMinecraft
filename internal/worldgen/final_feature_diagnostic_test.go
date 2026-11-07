package worldgen

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func diagnosticNativeChunk(t *testing.T, dir string, cx, cz int32) ([]uint32, [24][64]string) {
	t.Helper()
	_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), cx, cz)
	if err != nil || !exists {
		t.Fatalf("native chunk %d,%d: exists=%t err=%v", cx, cz, exists, err)
	}
	chunk, _ := tag.Object()
	states := make([]uint32, 384*256)
	var biomes [24][64]string
	sections, _ := chunk["sections"].Elements()
	for _, section := range sections {
		s, _ := section.Object()
		sy, _ := s.Number("Y")
		if sy < -4 || sy > 19 {
			continue
		}
		decode := func(container nbt.Object, floor, size int, visit func(int, nbt.Tag)) {
			palette, _ := container["palette"].Elements()
			packed, _ := container["data"].Longs()
			width := max(floor, bits.Len(uint(len(palette)-1)))
			for i := 0; i < size; i++ {
				index := 0
				if len(palette) > 1 {
					per := 64 / width
					index = int(uint64(packed[i/per]) >> ((i % per) * width) & ((1 << width) - 1))
				}
				visit(i, palette[index])
			}
		}
		if container, found := s.Object("block_states"); found {
			decode(container, 4, 4096, func(i int, tag nbt.Tag) {
				state, _ := tag.Object()
				name, _ := state.Text("Name")
				properties := make(map[string]string)
				props, _ := state.Object("Properties")
				for key, value := range props {
					properties[key], _ = value.Text()
				}
				id, err := blockstate.BlockStateID(name, properties)
				if err != nil {
					t.Fatal(err)
				}
				states[(int(sy)*16+64)*256+i] = id
			})
		}
		if container, found := s.Object("biomes"); found {
			decode(container, 1, 64, func(i int, tag nbt.Tag) { biomes[sy+4][i], _ = tag.Text() })
		}
	}
	return states, biomes
}

type diagnosticFeatureWriter struct {
	feature string
	x, z    int32
}

// This intentionally duplicates only the structure-disabled diagnostic replay,
// so alternate source orders can be inspected without modifying production.
func diagnosticReplayFeatures(t *testing.T, g *OverworldGenerator, bases map[featureChunkPosition]*overworldBase, chunks []*NoiseChunkResult, cx, cz int32, order string) (*OverworldFeatureRegion, map[FeaturePosition]diagnosticFeatureWriter) {
	t.Helper()
	region, err := NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	region.borrowed = make(map[featureChunkPosition]bool)
	region.BeginFeatureStage()
	for p := range region.chunks {
		region.borrowed[p] = true
	}
	biomeAt := func(qx, qy, qz int32) (string, error) {
		base := bases[featureChunkPosition{qx >> 2, qz >> 2}]
		if base == nil {
			return "", fmt.Errorf("missing biome %d,%d", qx>>2, qz>>2)
		}
		qy = max(int32(-16), min(int32(79), qy))
		return base.biomes[(qy+16)>>2][((qy&3)<<4)|((qz&3)<<2)|(qx&3)], nil
	}
	region.biomeSeed, region.biomeAtQuart, region.biomeBases = g.seed, biomeAt, bases
	region.treeShuffle = &treeShuffleContext{seed: g.seed, salt: g.beehiveShuffleSalt}
	var positions []featureChunkPosition
	for x := cx - 1; x <= cx+1; x++ {
		for z := cz - 1; z <= cz+1; z++ {
			positions = append(positions, featureChunkPosition{x, z})
		}
	}
	switch order {
	case "trace":
		positions = diagnosticTracePositions(t, os.Getenv("ATLAS_FINAL_DIAG_TRACE"), cx, cz)
	case "zx":
		sort.Slice(positions, func(i, j int) bool {
			a, b := positions[i], positions[j]
			return a.chunkZ < b.chunkZ || a.chunkZ == b.chunkZ && a.chunkX < b.chunkX
		})
	case "reverse":
		for i, j := 0, len(positions)-1; i < j; i, j = i+1, j-1 {
			positions[i], positions[j] = positions[j], positions[i]
		}
	case "center-first":
		sort.SliceStable(positions, func(i, j int) bool {
			a, b := positions[i], positions[j]
			return max(abs32(a.chunkX-cx), abs32(a.chunkZ-cz)) < max(abs32(b.chunkX-cx), abs32(b.chunkZ-cz))
		})
	}
	last := make(map[FeaturePosition]diagnosticFeatureWriter)
	for _, p := range positions {
		sx, sz := p.chunkX, p.chunkZ
		region.treeShuffle.sourceX, region.treeShuffle.sourceZ = sx, sz
		present := make(map[string]bool)
		for x := sx - 1; x <= sx+1; x++ {
			for z := sz - 1; z <= sz+1; z++ {
				for _, name := range bases[featureChunkPosition{x, z}].biomeNames {
					present[name] = true
				}
			}
		}
		var biomes []string
		for name := range present {
			biomes = append(biomes, name)
		}
		schedule, err := OverworldFeatureSchedule(biomes)
		if err != nil {
			t.Fatal(err)
		}
		seed := OverworldDecorationSeed(g.seed, sx, sz)
		random := NewWorldgenRandom(seed)
		for _, feature := range schedule {
			plan, exists := g.plans[feature.Name]
			if !exists {
				continue
			}
			region.treeShuffle.step, region.treeShuffle.index = feature.Step, feature.Index
			operations, err := plan.placement.bind(region, biomeAt)
			if err != nil {
				t.Fatal(err)
			}
			writes := &structureBlockResult{blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
			region.structureRecording = writes
			random.SetSeed(seed + int64(feature.Index) + int64(10000*feature.Step))
			traceRandom := os.Getenv("ATLAS_FINAL_DIAG_FEATURE_TRACE") != "" && abs32(sx-cx) <= 1 && abs32(sz-cz) <= 1
			if traceRandom {
				lo, hi := random.XoroshiroState()
				t.Logf("feature_random begin %d %d %d %d %d %d", sx, sz, feature.Step, feature.Index, lo, hi)
			}
			if feature.Step == 6 && feature.Index <= 10 && abs32(sx-cx) <= 1 && abs32(sz-cz) <= 1 {
				diagnosticFeatureSnapshot(t, region, cx, cz, sx, sz, feature.Index, "begin")
			}
			if err := walkPlacement(operations, random, FeaturePosition{X: sx * 16, Z: sz * 16}, func(origin FeaturePosition, r *WorldgenRandom) error {
				if os.Getenv("ATLAS_FINAL_DIAG_FEATURE_TRACE") != "" && sx == -3 && sz == 44 && feature.Step == 6 && feature.Index == 0 {
					lo, hi := r.XoroshiroState()
					t.Logf("configured_origin %d %d %d %d %d %d %d %d %d", sx, sz, feature.Step, feature.Index, origin.X, origin.Y, origin.Z, lo, hi)
				}
				trace := os.Getenv("ATLAS_FINAL_DIAG_SOURCE_TRACE") != "" && sx == cx && sz == cz && cx == 0 && cz == 43 && feature.Name == "minecraft:ore_copper"
				var beforeLo, beforeHi uint64
				var floor int32
				if trace {
					beforeLo, beforeHi = r.XoroshiroState()
					floor, _ = region.OceanFloorHeightAt(origin.X, origin.Z)
				}
				placed, err := plan.place(region, r, origin)
				if trace {
					afterLo, afterHi := r.XoroshiroState()
					t.Logf("%s copper origin=%v floor=%d placed=%t random-before=%d,%d random-after=%d,%d", order, origin, floor, placed, beforeLo, beforeHi, afterLo, afterHi)
				}
				return err
			}); err != nil {
				t.Fatal(err)
			}
			if traceRandom {
				lo, hi := random.XoroshiroState()
				t.Logf("feature_random end %d %d %d %d %d %d", sx, sz, feature.Step, feature.Index, lo, hi)
			}
			if feature.Step == 6 && feature.Index <= 10 && abs32(sx-cx) <= 1 && abs32(sz-cz) <= 1 {
				diagnosticFeatureSnapshot(t, region, cx, cz, sx, sz, feature.Index, "end")
			}
			for p := range writes.blocks {
				if p.X>>4 == cx && p.Z>>4 == cz {
					last[p] = diagnosticFeatureWriter{feature.Name, sx, sz}
				}
			}
		}
	}
	region.structureRecording = nil
	return region, last
}

func diagnosticFeatureSnapshot(t *testing.T, region *OverworldFeatureRegion, cx, cz, sx, sz int32, index int, phase string) {
	t.Helper()
	dir := os.Getenv("ATLAS_FINAL_DIAG_SNAPSHOTS")
	if dir == "" {
		return
	}
	name := fmt.Sprintf("%s_%d_%d_6_%d.gz", phase, sx, sz, index)
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	r, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	r.Close()
	f.Close()
	if err != nil || len(b) != 98304*2 {
		t.Fatalf("invalid snapshot %s: %v", name, err)
	}
	count := 0
	first := ""
	pairs := make(map[string]int)
	for i := 0; i < 98304; i++ {
		p := FeaturePosition{cx*16 + int32(i%16), int32(i/256) - 64, cz*16 + int32(i/16%16)}
		s, err := region.BlockAt(p)
		if err != nil {
			t.Fatal(err)
		}
		id, err := blockstate.BlockStateID(s.Name, s.Properties)
		if err != nil {
			t.Fatal(err)
		}
		want := uint32(binary.BigEndian.Uint16(b[i*2:]))
		if id == want {
			continue
		}
		expected, err := blockstate.BlockStateByID(want)
		if err != nil {
			t.Fatal(err)
		}
		pairs[s.Name+" -> "+expected.Name]++
		if count == 0 {
			first = fmt.Sprintf("%v got=%v want=%v", p, s, expected)
		}
		count++
	}
	t.Logf("feature_snapshot %s differences=%d pairs=%v first=%s", name, count, pairs, first)
}

// Read only original calls in the requested 7x7 neighborhood. Reject overlapping
// calls here: a serial replay cannot explain concurrent native writes.
func diagnosticTracePositions(t *testing.T, path string, cx, cz int32) []featureChunkPosition {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var positions []featureChunkPosition
	active := make(map[featureChunkPosition]bool)
	seen := make(map[featureChunkPosition]bool)
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) == 0 || fields[0] != "decoration" {
			continue
		}
		if len(fields) != 6 {
			t.Fatal("invalid decoration trace", s.Text())
		}
		x, err := strconv.ParseInt(fields[3], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		z, err := strconv.ParseInt(fields[4], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		p := featureChunkPosition{int32(x), int32(z)}
		if abs32(p.chunkX-cx) > 3 || abs32(p.chunkZ-cz) > 3 {
			continue
		}
		if fields[2] == "begin" {
			if len(active) != 0 || seen[p] {
				t.Fatal("nonserial or repeated native decoration", s.Text())
			}
			active[p], seen[p] = true, true
			positions = append(positions, p)
		} else if fields[2] == "end" {
			if !active[p] {
				t.Fatal("unmatched native decoration exit", s.Text())
			}
			delete(active, p)
		} else {
			t.Fatal("unknown trace phase", s.Text())
		}
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 || len(positions) == 0 {
		t.Fatal("incomplete native trace")
	}
	return positions
}

func TestFinalFeatureDiagnostic(t *testing.T) {
	dir := os.Getenv("ATLAS_FINAL_DIAG_WORLD")
	if dir == "" {
		t.Skip("ATLAS_FINAL_DIAG_WORLD is not set")
	}
	g, err := NewOverworldGeneratorWithBeehiveSalt(4511451270791457643, 5393695769601225306)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range [][2]int32{{0, 43}, {-3, 44}} {
		t.Run(fmt.Sprint(p), func(t *testing.T) {
			cx, cz := p[0], p[1]
			want, _ := diagnosticNativeChunk(t, dir, cx, cz)
			if output := os.Getenv("ATLAS_FINAL_DIAG_OUTPUT"); output != "" {
				var text string
				for i, id := range want {
					state, _ := blockstate.BlockStateByID(id)
					if treeTags["leaves"][state.Name] {
						text += fmt.Sprintf("leaf %d %d %d %d\n", cx*16+int32(i%16), int32(i/256)-64, cz*16+int32(i/16%16), id)
					}
				}
				if err := os.WriteFile(filepath.Join(output, fmt.Sprintf("native-%d-%d-leaves.txt", cx, cz)), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			bases, chunks, err := g.decorationBases(cx, cz)
			if err != nil {
				t.Fatal(err)
			}
			if os.Getenv("ATLAS_FINAL_DIAG_TRACE") != "" {
				for x := cx - 4; x <= cx+4; x++ {
					for z := cz - 4; z <= cz+4; z++ {
						p := featureChunkPosition{x, z}
						if bases[p] != nil {
							continue
						}
						entry := g.baseEntryAt(x, z)
						<-entry.ready
						if entry.err != nil {
							t.Fatal(entry.err)
						}
						bases[p] = entry.base
						chunks = append(chunks, entry.base.terrain)
					}
				}
			}
			for x := cx - 1; x <= cx+1; x++ {
				for z := cz - 1; z <= cz+1; z++ {
					_, biomes := diagnosticNativeChunk(t, dir, x, z)
					different := 0
					for section, entries := range biomes {
						for i, name := range entries {
							if name != bases[featureChunkPosition{x, z}].biomes[section][i] {
								different++
							}
						}
					}
					t.Logf("source %d,%d biome differences=%d/1536", x, z, different)
				}
			}
			orders := []string{"xz", "zx", "reverse", "center-first"}
			if os.Getenv("ATLAS_FINAL_DIAG_TRACE") != "" {
				orders = []string{"trace"}
			}
			if os.Getenv("ATLAS_FINAL_DIAG_SOURCE_TRACE") != "" {
				orders = []string{"xz", "center-first"}
			}
			for _, order := range orders {
				region, last := diagnosticReplayFeatures(t, g, bases, chunks, cx, cz, order)
				chunk := region.cloneChunk(cx, cz)
				pairs, writers := make(map[string]int), make(map[diagnosticFeatureWriter]int)
				different := 0
				for i := range chunk.Blocks {
					p := FeaturePosition{cx*16 + int32(i%16), int32(i/256) - 64, cz*16 + int32(i/16%16)}
					s, _ := chunk.BlockStateAt(p.X&15, p.Y, p.Z&15)
					id, err := blockstate.BlockStateID(s.Name, s.Properties)
					if err != nil {
						t.Fatal(err)
					}
					if id != want[i] {
						different++
						native, _ := blockstate.BlockStateByID(want[i])
						pairs[fmt.Sprint(native)+" -> "+fmt.Sprint(s)]++
						writers[last[p]]++
					}
				}
				t.Logf("order %s differences=%d pairs=%v last-writers=%v", order, different, pairs, writers)
			}
		})
	}
}
