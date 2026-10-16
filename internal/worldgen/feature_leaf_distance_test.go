package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestGeneratedLeafDistancesAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/leaves_distance_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		p         FeaturePosition
		initialID uint32
		wantID    uint32
	}
	cases := make(map[int][]record)
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "leaf_state ") {
			continue
		}
		var label string
		var scenario int
		var r record
		if _, err := fmt.Sscan(line, &label, &scenario, &r.p.X, &r.p.Y, &r.p.Z, &r.initialID, &r.wantID); err != nil {
			t.Fatal(err)
		}
		cases[scenario] = append(cases[scenario], r)
	}
	if len(cases) != 4 {
		t.Fatalf("native leaf scenarios=%d, want 4", len(cases))
	}
	for scenario, records := range cases {
		t.Run(strconv.Itoa(scenario), func(t *testing.T) {
			chunks := make(map[featureChunkPosition]*NoiseChunkResult)
			for _, r := range records {
				key := featureChunkPosition{r.p.X >> 4, r.p.Z >> 4}
				if chunks[key] == nil {
					chunks[key] = &NoiseChunkResult{X: key.chunkX, Z: key.chunkZ, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
				}
			}
			var inputs []*NoiseChunkResult
			for _, chunk := range chunks {
				inputs = append(inputs, chunk)
			}
			region, err := NewOverworldFeatureRegion(inputs...)
			if err != nil {
				t.Fatal(err)
			}
			region.borrowed = make(map[featureChunkPosition]bool)
			for key := range chunks {
				region.borrowed[key] = true
			}
			// Mix authored template states in the immutable base with late tree
			// writes in private overlays. Both must participate in native updates.
			for i, r := range records {
				s, err := blockstate.BlockStateByID(r.initialID)
				if err != nil {
					t.Fatal(err)
				}
				state := SurfaceBlockState{Name: s.Name, Properties: s.Properties}
				if i%2 == 0 {
					chunk := chunks[featureChunkPosition{r.p.X >> 4, r.p.Z >> 4}]
					chunk.setSurfaceState(r.p.X&15, r.p.Y, r.p.Z&15, state)
				} else if err := region.SetBlock(r.p, state); err != nil {
					t.Fatal(err)
				}
			}
			for key := range chunks {
				if err := finalizeGeneratedLeafDistances(region, key.chunkX, key.chunkZ); err != nil {
					t.Fatal(err)
				}
			}
			for _, r := range records {
				state, err := region.BlockAt(r.p)
				if err != nil {
					t.Fatal(err)
				}
				id, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil || id != r.wantID {
					t.Fatalf("native leaf update at %v: state %d (%v), want %d (%v)", r.p, id, state, r.wantID, err)
				}
			}
		})
	}
}

func TestGeneratedAdjacentTreesLeafDistances(t *testing.T) {
	config, err := LoadTreeConfiguration("minecraft:oak")
	if err != nil {
		t.Fatal(err)
	}
	for _, xs := range [][2]int32{{14, 17}, {-2, 1}} {
		t.Run(fmt.Sprint(xs), func(t *testing.T) {
			var chunks []*NoiseChunkResult
			for cx := xs[0]>>4 - 1; cx <= xs[1]>>4+1; cx++ {
				chunk := &NoiseChunkResult{X: cx, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
				for x := int32(0); x < 16; x++ {
					for z := int32(0); z < 16; z++ {
						chunk.setSurfaceState(x, 63, z, SurfaceBlockState{Name: "minecraft:grass_block"})
					}
				}
				chunks = append(chunks, chunk)
			}
			region, err := NewOverworldFeatureRegion(chunks...)
			if err != nil {
				t.Fatal(err)
			}
			for i, x := range xs {
				placed, err := PlaceTreeFeature(region, NewWorldgenRandom(int64(i)), FeaturePosition{x, 64, 8}, config)
				if err != nil || !placed {
					t.Fatalf("adjacent tree at %d: placed=%t err=%v", x, placed, err)
				}
			}
			for _, chunk := range chunks {
				if err := finalizeGeneratedLeafDistances(region, chunk.X, 0); err != nil {
					t.Fatal(err)
				}
			}
			leaves, boundaryLeaves := 0, 0
			for _, chunk := range chunks {
				for i, state := range chunk.SurfaceBlocks {
					if !treeTags["leaves"][state.Name] {
						continue
					}
					leaves++
					p := FeaturePosition{chunk.X*16 + int32(i%16), int32(i/256) - 64, int32(i / 16 % 16)}
					if p.X&15 == 0 || p.X&15 == 15 {
						boundaryLeaves++
					}
					// The actual LeavesBlock tick rule independently defines
					// the local fixed point, rather than mirroring our BFS.
					want := 7
					for _, d := range []FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}} {
						n, err := region.BlockAt(treeOffset(p, d.X, d.Y, d.Z))
						if err != nil {
							t.Fatal(err)
						}
						if treeTags["logs"][n.Name] {
							want = 1
						} else if treeTags["leaves"][n.Name] {
							distance, _ := strconv.Atoi(n.Properties["distance"])
							want = min(want, distance+1)
						}
					}
					if state.Properties["distance"] != strconv.Itoa(want) {
						t.Fatalf("leaf at chunk boundary %v violates native tick rule: %v want %d", p, state, want)
					}
				}
			}
			if leaves == 0 || boundaryLeaves == 0 {
				t.Fatalf("adjacent tree test did not cross a chunk boundary: leaves=%d boundary=%d", leaves, boundaryLeaves)
			}
		})
	}
}

func TestGeneratedLeavesAgainstVanillaFinalWorld(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/river_edge_final_leaves_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	want := make(map[FeaturePosition]uint32)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var label string
		var p FeaturePosition
		var id uint32
		if _, err := fmt.Sscan(scanner.Text(), &label, &p.X, &p.Y, &p.Z, &id); err != nil {
			t.Fatal(err)
		}
		want[p] = id
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(want) != 47 {
		t.Fatalf("independent final-world leaf count=%d, want 47", len(want))
	}
	g, err := NewOverworldGeneratorWithBeehiveSalt(4511451270791457643, 5393695769601225306)
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := g.Generate(-3, 44)
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[FeaturePosition]uint32)
	for i, state := range chunk.Terrain.SurfaceBlocks {
		if treeTags["leaves"][state.Name] {
			p := FeaturePosition{-48 + int32(i%16), int32(i/256) - 64, 704 + int32(i/16%16)}
			got[p], err = blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("final leaf records differ from independently generated official world:\n got %v\nwant %v", got, want)
	}
}
