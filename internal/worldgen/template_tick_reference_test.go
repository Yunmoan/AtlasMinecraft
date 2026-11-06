package worldgen

import (
	"atlas/internal/blockstate"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"testing"
)

type templateTickProbeWorld struct {
	blocks   map[FeaturePosition]SurfaceBlockState
	fill     SurfaceBlockState
	requests []treeTickRecord
	region   *OverworldFeatureRegion
	groups   []templateTickGroup
	seen     map[templateTickGroup]bool
	random   *LegacyRandom
}

func (w *templateTickProbeWorld) ShapeRandomInt(bound int32) (int32, error) {
	return w.random.NextIntBound(bound), nil
}

type templateTickGroup struct {
	kind string
	x, z int32
}

func (w *templateTickProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if s, ok := w.blocks[p]; ok {
		return s, nil
	}
	return w.fill, nil
}
func (w *templateTickProbeWorld) SetBlock(p FeaturePosition, s SurfaceBlockState) error {
	w.blocks[p] = s
	return nil
}
func (w *templateTickProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }
func (w *templateTickProbeWorld) record(r treeTickRecord) {
	w.requests = append(w.requests, r)
	g := templateTickGroup{r.Kind, r.X >> 4, r.Z >> 4}
	if !w.seen[g] {
		w.seen[g] = true
		w.groups = append(w.groups, g)
	}
}
func (w *templateTickProbeWorld) ScheduleBlockTick(p FeaturePosition, name string, delay, priority int32) error {
	w.record(treeTickRecord{"block", name, p.X, p.Y, p.Z, delay, priority})
	return w.region.ScheduleBlockTick(p, name, delay, priority)
}
func (w *templateTickProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	s, err := w.BlockAt(p)
	if err != nil {
		return err
	}
	f, err := FluidStateForBlock(s)
	if err != nil {
		return err
	}
	w.record(treeTickRecord{"fluid", f.Type.String(), p.X, p.Y, p.Z, FluidTickDelay(f.Type, false), 0})
	return w.region.ScheduleFluidTick(PendingFluidTick{Position: p, Fluid: f.Type.String()})
}
func templateProbePositions(shape int) []FeaturePosition {
	if shape == 0 {
		return []FeaturePosition{{-1, 60, 15}}
	}
	if shape == 1 {
		return []FeaturePosition{{-1, 60, 15}, {0, 60, 15}, {0, 60, 16}}
	}
	var out []FeaturePosition
	for x := int32(0); x < 3; x++ {
		for y := int32(0); y < 3; y++ {
			for z := int32(0); z < 3; z++ {
				if x == 0 || x == 2 || y == 0 || y == 2 || z == 0 || z == 2 {
					out = append(out, FeaturePosition{x - 1, y + 60, z + 15})
				}
			}
		}
	}
	return out
}

func TestTemplateGenerationTicksAgainstVanilla(t *testing.T) {
	// Scheduling does not mutate terrain. All cases can borrow these complete
	// chunk arrays while owning their queues independently.
	terrain := make([]NoiseMaterial, 98304)
	var chunks []*NoiseChunkResult
	for _, x := range []int32{-1, 0} {
		for _, z := range []int32{0, 1} {
			chunks = append(chunks, &NoiseChunkResult{X: x, Z: z, MinY: -64, Height: 384, Blocks: terrain})
		}
	}
	updaters := []struct {
		name string
		fn   structureTemplateShapeUpdater
	}{
		{"igloo", iglooTemplateUpdateShape}, {"ruined_portal", ruinedPortalTemplateUpdateShape},
		{"shipwreck", shipwreckTemplateUpdateShape}, {"ocean_ruin", oceanRuinTemplateUpdateShape},
		{"end_city", endCityTemplateUpdateShape}, {"mansion", mansionTemplateUpdateShape},
		{"runtime", func(w FeatureBlockWorld, p FeaturePosition, s SurfaceBlockState, d string, n SurfaceBlockState) (SurfaceBlockState, error) {
			return UpdateFluidNeighborShape(w, p, s, map[string]int{"down": 0, "up": 1, "north": 2, "south": 3, "west": 4, "east": 5}[d], n)
		}},
	}
	state := func(id uint32) SurfaceBlockState {
		s, err := blockstate.BlockStateByID(id)
		if err != nil {
			t.Fatal(err)
		}
		return SurfaceBlockState{Name: s.Name, Properties: s.Properties}
	}
	cases, requests, packed := 0, 0, 0
	scanPostprocessingReference(t, "../../testdata/worldgen/template_ticks_1.21.1.jsonl.gz", func(raw []byte) {
		var row struct {
			State, Fill      uint32
			Shape            int
			Finish           bool
			Blocks           [][4]int32
			Requests, Packed []treeTickRecord
			Next             string
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			t.Fatal(err)
		}
		placed := templateProbePositions(row.Shape)
		for _, updater := range updaters {
			region, err := NewOverworldFeatureRegion(chunks...)
			if err != nil {
				t.Fatal(err)
			}
			w := &templateTickProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{}, fill: state(row.Fill), region: region, seen: map[templateTickGroup]bool{}}
			w.random = NewLegacyRandom(1234)
			for _, p := range placed {
				w.blocks[p] = state(row.State)
			}
			if row.Finish {
				err = finalizeStructureTemplateShape(w, placed, updater.fn)
			} else {
				err = updateStructureTemplateEdges(w, placed, updater.fn)
			}
			if err != nil {
				t.Fatal(err)
			}
			label := fmt.Sprintf("%s state=%d fill=%d shape=%d finish=%t", updater.name, row.State, row.Fill, row.Shape, row.Finish)
			if !slices.Equal(w.requests, row.Requests) {
				first := 0
				for first < min(len(w.requests), len(row.Requests)) && w.requests[first] == row.Requests[first] {
					first++
				}
				t.Fatalf("%s: request counts %d/%d first difference %d: got %v want %v", label, len(w.requests), len(row.Requests), first, w.requests[max(0, first-1):min(len(w.requests), first+2)], row.Requests[max(0, first-1):min(len(row.Requests), first+2)])
			}
			var blocks [][4]int32
			for p, s := range w.blocks {
				id, err := blockstate.BlockStateID(s.Name, s.Properties)
				if err != nil {
					t.Fatal(err)
				}
				blocks = append(blocks, [4]int32{p.X, p.Y, p.Z, int32(id)})
			}
			sort.Slice(blocks, func(i, j int) bool {
				for axis := 0; axis < 3; axis++ {
					if blocks[i][axis] != blocks[j][axis] {
						return blocks[i][axis] < blocks[j][axis]
					}
				}
				return false
			})
			if !slices.Equal(blocks, row.Blocks) {
				t.Fatalf("%s: final block states differ", label)
			}
			var got []treeTickRecord
			for _, group := range w.groups {
				if group.kind == "block" {
					for _, r := range region.blockTicksInChunk(group.x, group.z) {
						got = append(got, treeTickRecord{"block", r.Block, r.Position.X, r.Position.Y, r.Position.Z, 0, r.Priority})
					}
				} else {
					for _, r := range region.fluidTicksInChunk(group.x, group.z) {
						got = append(got, treeTickRecord{"fluid", r.Fluid, r.Position.X, r.Position.Y, r.Position.Z, 0, r.Priority})
					}
				}
			}
			if !slices.Equal(got, row.Packed) {
				t.Fatalf("%s: packed native queues differ", label)
			}
			next, err := strconv.ParseUint(row.Next, 10, 64)
			if err != nil || w.random.NextLong() != next {
				t.Fatalf("%s: original shape RNG differs: %v", label, err)
			}
		}
		if row.Fill == 0 && row.Shape == 1 && row.Finish {
			checkTemplateTickRecording(t, chunks, placed, state(row.State), row.Packed)
		}
		cases++
		requests += len(row.Requests)
		packed += len(row.Packed)
	})
	if cases != 27792 {
		t.Fatal("incomplete native template coverage", cases)
	}
	t.Logf("matched %d native template cases through %d adapters, %d ordered requests and %d packed ticks", cases, len(updaters), requests, packed)
}

// The actual structure recorder must retain these callbacks, including its
// delegated region mode. This exercises production adapters, not manually
// populated cached requests, at both X and Z chunk boundaries.
func checkTemplateTickRecording(t *testing.T, chunks []*NoiseChunkResult, placed []FeaturePosition, state SurfaceBlockState, want []treeTickRecord) {
	t.Helper()
	for _, delegated := range []bool{false, true} {
		region, err := NewOverworldFeatureRegion(chunks...)
		if err != nil {
			t.Fatal(err)
		}
		region.borrowed = map[featureChunkPosition]bool{}
		region.shapeRandom = NewLegacyRandom(1234)
		for _, c := range chunks {
			region.borrowed[featureChunkPosition{c.X, c.Z}] = true
		}
		record := &structureBlockResult{blocks: map[FeaturePosition]SurfaceBlockState{}}
		w := &structureRecordingWorld{region: region, result: record}
		if delegated {
			region.structureRecording = record
		}
		for _, p := range placed {
			if err := w.SetBlock(p, state); err != nil {
				t.Fatal(err)
			}
		}
		if err := finalizeStructureTemplateShape(w, placed, mansionTemplateUpdateShape); err != nil {
			t.Fatal(err)
		}
		for _, c := range chunks {
			var expectedB, recordB []PendingBlockTick
			var expectedF, recordF []PendingFluidTick
			for _, r := range want {
				if r.X>>4 == c.X && r.Z>>4 == c.Z {
					p := FeaturePosition{r.X, r.Y, r.Z}
					if r.Kind == "block" {
						expectedB = append(expectedB, PendingBlockTick{p, r.Type, r.Priority})
					} else {
						expectedF = append(expectedF, PendingFluidTick{p, r.Type, r.Priority})
					}
				}
			}
			for _, r := range record.blockTicks.ticks {
				if r.Position.X>>4 == c.X && r.Position.Z>>4 == c.Z {
					recordB = append(recordB, r)
				}
			}
			for _, r := range record.fluidTicks.ticks {
				if r.Position.X>>4 == c.X && r.Position.Z>>4 == c.Z {
					recordF = append(recordF, r)
				}
			}
			if !slices.Equal(recordB, expectedB) || !slices.Equal(region.blockTicksInChunk(c.X, c.Z), expectedB) || !slices.Equal(recordF, expectedF) || !slices.Equal(region.fluidTicksInChunk(c.X, c.Z), expectedF) {
				t.Fatalf("recording %s %v delegated=%v chunk=%d,%d lost/reordered native template queues", state.Name, state.Properties, delegated, c.X, c.Z)
			}
		}
	}
}
