package worldgen

import (
	"fmt"
	"slices"
	"testing"
)

func TestPostprocessingReplayRetainsRepeatedAndOverwrittenMarks(t *testing.T) {
	for _, borrowed := range []bool{false, true} {
		for _, delegated := range []bool{false, true} {
			t.Run(fmt.Sprintf("borrowed=%v/delegated=%v", borrowed, delegated), func(t *testing.T) {
				base := newFlatForestRockChunk(-1, 0)
				base.PostProcessingOrder = []int{513, 513}
				neighbor := newFlatForestRockChunk(0, 0)
				makeRegion := func() *OverworldFeatureRegion {
					r, err := NewFeatureRegion(cloneTerrain(base), cloneTerrain(neighbor))
					if err != nil {
						t.Fatal(err)
					}
					r.borrowed = map[featureChunkPosition]bool{{-1, 0}: borrowed, {0, 0}: borrowed}
					return r
				}
				r := makeRegion()
				record := &structureBlockResult{blocks: map[FeaturePosition]SurfaceBlockState{}}
				w := &structureRecordingWorld{region: r, result: record}
				if delegated {
					r.structureRecording = record
				}
				p, q, n := FeaturePosition{-1, 64, 15}, FeaturePosition{-2, 80, 14}, FeaturePosition{0, 64, 15}
				must := func(err error) {
					t.Helper()
					if err != nil {
						t.Fatal(err)
					}
				}
				must(w.SetBlock(p, structureState("magma_block")))
				must(w.SetBlock(q, structureState("brown_mushroom")))
				must(w.SetBlock(p, structureState("magma_block"))) // unchanged writes still mark
				must(w.SetBlock(n, structureState("magma_block")))
				must(w.MarkPostProcessing(p))
				must(w.SetBlock(p, structureState("stone")))
				must(w.MarkPostProcessing(FeaturePosition{-1, 320, 15}))
				want := []FeaturePosition{p, q, p, n, p}
				if !slices.Equal(record.postProcessingOrder, want) {
					t.Fatal("recording changed append log", record.postProcessingOrder)
				}
				wantIndices := append([]int(nil), base.PostProcessingOrder...)
				for _, pos := range []FeaturePosition{p, q, p, p} {
					i, _ := base.Index(pos.X&15, pos.Y, pos.Z&15)
					wantIndices = append(wantIndices, i)
				}
				if !slices.Equal(r.cloneChunk(-1, 0).PostProcessingOrder, wantIndices) {
					t.Fatal("generation reordered marks")
				}
				for _, clipped := range []bool{false, true} {
					replayed := makeRegion()
					must(replayed.restoreStructureBlocks(record, -1, 0, clipped))
					got := replayed.cloneChunk(-1, 0)
					if !slices.Equal(got.PostProcessingOrder, wantIndices) {
						t.Fatal("cache replay changed marks", got.PostProcessingOrder, wantIndices)
					}
					if state, err := got.BlockStateAt(15, 64, 15); err != nil || state.Name != "minecraft:stone" {
						t.Fatal("cache lost last write", state, err)
					}
					count := len(replayed.cloneChunk(0, 0).PostProcessingOrder)
					if clipped && count != 1 || !clipped && count != 0 {
						t.Fatal("cache clip changed ownership", count)
					}
					got.PostProcessingOrder[0] = 99
					if replayed.cloneChunk(-1, 0).PostProcessingOrder[0] != 513 {
						t.Fatal("clone aliases append log")
					}
				}
			})
		}
	}
}

func TestPostprocessingSnapshotOwnsAppendLog(t *testing.T) {
	base := newFlatForestRockChunk(0, 0)
	snapshot := &structureChunkSnapshot{base: &overworldBase{terrain: base}, overlay: &featureChunkOverlay{
		blocks: map[int]SurfaceBlockState{}, postProcessing: map[int]bool{17: true}, postProcessingOrder: append(make([]int, 0, 32), 17, 17),
	}}
	var regions [2]*OverworldFeatureRegion
	for i := range regions {
		r, err := NewFeatureRegion(base)
		if err != nil {
			t.Fatal(err)
		}
		r.borrowed = map[featureChunkPosition]bool{}
		r.biomeBases = map[featureChunkPosition]*overworldBase{}
		r.installStructureChunk(featureChunkPosition{0, 0}, snapshot)
		regions[i] = r
	}
	regions[0].overlays[featureChunkPosition{0, 0}].postProcessingOrder[0] = 99
	if err := regions[0].MarkPostProcessing(FeaturePosition{0, 64, 0}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(regions[1].cloneChunk(0, 0).PostProcessingOrder, []int{17, 17}) || !slices.Equal(snapshot.overlay.postProcessingOrder, []int{17, 17}) {
		t.Fatal("snapshot shares append log")
	}
}
