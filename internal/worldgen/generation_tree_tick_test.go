package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

type treeTickRecord struct {
	Kind, Type               string
	X, Y, Z, Delay, Priority int32
}
type treeTickProbeWorld struct {
	*treeProbeWorld
	requests []treeTickRecord
}

func (w *treeTickProbeWorld) ScheduleBlockTick(p FeaturePosition, name string, delay, priority int32) error {
	w.requests = append(w.requests, treeTickRecord{"block", name, p.X, p.Y, p.Z, delay, priority})
	return nil
}
func (w *treeTickProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	s, err := w.BlockAt(p)
	if err != nil {
		return err
	}
	f, err := FluidStateForBlock(s)
	if err != nil {
		return err
	}
	w.requests = append(w.requests, treeTickRecord{"fluid", f.Type.String(), p.X, p.Y, p.Z, FluidTickDelay(f.Type, false), 0})
	return nil
}

func TestTreeGenerationTicksAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/tree_ticks_1.21.1.jsonl.gz")
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
	scan.Buffer(make([]byte, 65536), 4<<20)
	cases, requests, packed := 0, 0, 0
	for scan.Scan() {
		var row struct {
			Name             string
			Seed, Next       int64
			Scenario, Count  int
			Origin           [3]int32
			Placed           bool
			Hash             string
			Requests, Packed []treeTickRecord
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		config, err := LoadTreeConfiguration(row.Name)
		if err != nil {
			t.Fatal(err)
		}
		independent := NewLegacyRandom(row.Seed ^ 0x123456789abcdef0)
		config = config.WithBeehiveShuffle(func(p []FeaturePosition) {
			for i := len(p); i > 1; i-- {
				j := independent.NextIntBound(int32(i))
				p[i-1], p[j] = p[j], p[i-1]
			}
		})
		origin := FeaturePosition{row.Origin[0], row.Origin[1], row.Origin[2]}
		w := &treeTickProbeWorld{treeProbeWorld: &treeProbeWorld{changed: map[FeaturePosition]SurfaceBlockState{}, origin: origin, scenario: row.Scenario}}
		random := NewWorldgenRandom(row.Seed)
		placed, err := PlaceTreeFeature(w, random, origin, config)
		if err != nil {
			t.Fatalf("%s seed=%d scenario=%d: %v", row.Name, row.Seed, row.Scenario, err)
		}
		var output []string
		for p, s := range w.changed {
			id, err := blockstate.BlockStateID(s.Name, s.Properties)
			if err != nil {
				t.Fatal(err)
			}
			output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
		}
		sort.Strings(output)
		hash := sha256.Sum256([]byte(strings.Join(output, "\n")))
		next := int64(random.NextLong())
		if placed != row.Placed || len(w.changed) != row.Count || hex.EncodeToString(hash[:]) != row.Hash || next != row.Next {
			t.Fatalf("%s seed=%d scenario=%d: placement/blocks/RNG differ: %v %d %x %d; want %v %d %s %d", row.Name, row.Seed, row.Scenario, placed, len(w.changed), hash, next, row.Placed, row.Count, row.Hash, row.Next)
		}
		if !slices.Equal(w.requests, row.Requests) {
			i := 0
			for i < min(len(w.requests), len(row.Requests)) && w.requests[i] == row.Requests[i] {
				i++
			}
			t.Fatalf("%s seed=%d scenario=%d: %d requests want %d; first difference %d: got %v want %v", row.Name, row.Seed, row.Scenario, len(w.requests), len(row.Requests), i, w.requests[max(0, i-1):min(len(w.requests), i+2)], row.Requests[max(0, i-1):min(len(row.Requests), i+2)])
		}
		// Queue each captured callback through real generation regions. Compare
		// packed results with native ProtoChunkTicks, not a second Go deduper.
		chunks := map[featureChunkPosition]*NoiseChunkResult{}
		order := []string{}
		seen := map[string]bool{}
		for _, r := range w.requests {
			key := featureChunkPosition{r.X >> 4, r.Z >> 4}
			if chunks[key] == nil {
				chunks[key] = &NoiseChunkResult{X: key.chunkX, Z: key.chunkZ, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
			}
			group := fmt.Sprintf("%s %d %d", r.Kind, key.chunkX, key.chunkZ)
			if !seen[group] {
				seen[group] = true
				order = append(order, group)
			}
		}
		var actual []treeTickRecord
		if len(chunks) > 0 {
			inputs := []*NoiseChunkResult{}
			for _, chunk := range chunks {
				inputs = append(inputs, chunk)
			}
			region, err := NewOverworldFeatureRegion(inputs...)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range w.requests {
				p := FeaturePosition{r.X, r.Y, r.Z}
				if r.Kind == "block" {
					err = region.ScheduleBlockTick(p, r.Type, r.Delay, r.Priority)
				} else {
					err = region.ScheduleFluidTick(PendingFluidTick{Position: p, Fluid: r.Type, Priority: r.Priority})
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, group := range order {
				var kind string
				var x, z int32
				if _, err := fmt.Sscan(group, &kind, &x, &z); err != nil {
					t.Fatal(err)
				}
				if kind == "block" {
					for _, r := range region.blockTicksInChunk(x, z) {
						actual = append(actual, treeTickRecord{kind, r.Block, r.Position.X, r.Position.Y, r.Position.Z, 0, r.Priority})
					}
				} else {
					for _, r := range region.fluidTicksInChunk(x, z) {
						actual = append(actual, treeTickRecord{kind, r.Fluid, r.Position.X, r.Position.Y, r.Position.Z, 0, r.Priority})
					}
				}
			}
		}
		if !slices.Equal(actual, row.Packed) {
			t.Fatalf("%s seed=%d scenario=%d packed ticks differ", row.Name, row.Seed, row.Scenario)
		}
		cases++
		requests += len(w.requests)
		packed += len(actual)
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 2824 {
		t.Fatal("incomplete tree tick reference", cases)
	}
	t.Logf("matched %d trees, %d ordered requests, %d packed ticks", cases, requests, packed)
}
