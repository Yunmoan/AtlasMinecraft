package worldgen

import (
	"fmt"
	"slices"
	"testing"
)

func TestGenerationTickRoutingUsesChunkOwnership(t *testing.T) {
	left := &NoiseChunkResult{X: -2, Z: 2, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
	edge := &NoiseChunkResult{X: 1875000, Z: 0, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
	r, err := NewFeatureRegion(left, edge)
	if err != nil {
		t.Fatal(err)
	}
	// The original WorldGenTickAccess delegates without testing write bounds.
	// Both ProtoChunkTicks containers retain these otherwise unwritable positions.
	for _, p := range []FeaturePosition{{-17, -65, 47}, {-17, 320, 47}, {30000000, 64, 0}} {
		if err := r.ScheduleBlockTick(p, "minecraft:cave_air", 0, 0); err != nil {
			t.Fatal(err)
		}
		if err := r.ScheduleFluidTick(PendingFluidTick{Position: p, Fluid: "minecraft:water"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []*NoiseChunkResult{left, edge} {
		got := r.cloneChunk(c.X, c.Z)
		want := 2
		if c == edge {
			want = 1
		}
		if len(got.BlockTicks) != want || len(got.FluidTicks) != want {
			t.Fatal("request lost at write bounds", got.BlockTicks, got.FluidTicks)
		}
	}
	missing := FeaturePosition{30000016, 64, 0}
	if r.ScheduleBlockTick(missing, "minecraft:cave_air", 0, 0) == nil || r.ScheduleFluidTick(PendingFluidTick{Position: missing, Fluid: "minecraft:water"}) == nil {
		t.Fatal("missing owning chunk silently discarded a request")
	}
}

func TestGenerationBlockTicksReplayAndSnapshotIsolation(t *testing.T) {
	for _, borrowed := range []bool{false, true} {
		t.Run(fmt.Sprint(borrowed), func(t *testing.T) {
			chunk := &NoiseChunkResult{X: -2, Z: 2, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
			neighbor := &NoiseChunkResult{X: -1, Z: 2, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 98304)}
			region, err := NewFeatureRegion(chunk, neighbor)
			if err != nil {
				t.Fatal(err)
			}
			region.borrowed = map[featureChunkPosition]bool{{-2, 2}: borrowed, {-1, 2}: borrowed}
			p := FeaturePosition{-17, -63, 47}
			record := &structureBlockResult{}
			w := &structureRecordingWorld{region: region, result: record}
			for _, outside := range []FeaturePosition{{30000000, 64, 47}, {-17, 64, -30000000}} {
				if err := w.ScheduleBlockTick(outside, "minecraft:oak_leaves", 1, 0); err == nil {
					t.Fatal("request outside cached chunks was silently dropped")
				}
			}
			for _, priority := range []int32{-2, 3} {
				if err := w.ScheduleBlockTick(p, "minecraft:oak_leaves", 10, priority); err != nil {
					t.Fatal(err)
				}
			}
			first := PendingBlockTick{Position: p, Block: "minecraft:oak_leaves", Priority: -2}
			if !slices.Equal(record.blockTicks.ticks, []PendingBlockTick{first}) {
				t.Fatal("record deduplication/bounds", record.blockTicks.ticks)
			}
			// A later write and another type at the same position do not cancel
			// the original request, including when replayed from a structure cache.
			if err := region.SetBlock(p, structureState("stone")); err != nil {
				t.Fatal(err)
			}
			second := PendingBlockTick{Position: p, Block: "minecraft:birch_leaves", Priority: 1}
			if err := region.ScheduleGenerationBlockTick(second); err != nil {
				t.Fatal(err)
			}
			if err := region.ScheduleGenerationBlockTick(record.blockTicks.ticks[0]); err != nil {
				t.Fatal(err)
			}
			got := region.cloneChunk(-2, 2)
			if !slices.Equal(got.BlockTicks, []PendingBlockTick{first, second}) || len(region.cloneChunk(-1, 2).BlockTicks) != 0 {
				t.Fatal("replay/chunk ownership", got.BlockTicks)
			}
			got.BlockTicks[0].Block = "changed"
			if region.cloneChunk(-2, 2).BlockTicks[0] != first || len(chunk.BlockTicks) != 0 {
				t.Fatal("queue clone aliases input")
			}
			snapshot := &structureChunkSnapshot{base: &overworldBase{terrain: chunk}, blockTicks: []PendingBlockTick{first}}
			regions := make([]*OverworldFeatureRegion, 2)
			for i := range regions {
				r, err := NewFeatureRegion(chunk)
				if err != nil {
					t.Fatal(err)
				}
				r.borrowed = make(map[featureChunkPosition]bool)
				r.biomeBases = make(map[featureChunkPosition]*overworldBase)
				r.installStructureChunk(featureChunkPosition{-2, 2}, snapshot)
				regions[i] = r
			}
			if err := regions[0].ScheduleGenerationBlockTick(second); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(regions[0].cloneChunk(-2, 2).BlockTicks, []PendingBlockTick{first, second}) || !slices.Equal(regions[1].cloneChunk(-2, 2).BlockTicks, []PendingBlockTick{first}) || !slices.Equal(snapshot.blockTicks, []PendingBlockTick{first}) {
				t.Fatal("preceding snapshot aliases queues")
			}
		})
	}
}
