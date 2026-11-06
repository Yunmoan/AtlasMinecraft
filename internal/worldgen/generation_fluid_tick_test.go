package worldgen

import (
	"atlas/internal/blockstate"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGenerationFluidTicksAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/generation_fluid_ticks_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var queue generationFluidTickQueue
	savedIndex := 0
	for _, row := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(row)
		number := func(i int) int32 {
			t.Helper()
			v, e := strconv.ParseInt(f[i], 10, 32)
			if e != nil {
				t.Fatal(e)
			}
			return int32(v)
		}
		switch f[0] {
		case "generation_fluid_request":
			// Original trigger/sub-tick order are deliberately absent from a
			// generation queue: ProtoChunkTicks discards them.
			queue.add(PendingFluidTick{Position: FeaturePosition{number(2), number(3), number(4)}, Fluid: f[1], Priority: number(6)})
			if len(queue.ticks) != int(number(8)) {
				t.Fatalf("%s: count %d", row, len(queue.ticks))
			}
		case "generation_fluid_saved":
			got := queue.ticks[savedIndex%len(queue.ticks)]
			want := PendingFluidTick{Position: FeaturePosition{number(3), number(4), number(5)}, Fluid: f[2], Priority: number(7)}
			if got != want || number(6) != 0 {
				t.Fatalf("%s: got %+v", row, got)
			}
			savedIndex++
		case "generation_fluid_states":
			compressed, e := base64.StdEncoding.DecodeString(f[1])
			if e != nil {
				t.Fatal(e)
			}
			gz, e := gzip.NewReader(bytes.NewReader(compressed))
			if e != nil {
				t.Fatal(e)
			}
			types, e := io.ReadAll(gz)
			gz.Close()
			if e != nil || len(types) != 26684 {
				t.Fatal("invalid fluid type reference", e)
			}
			names := []string{"minecraft:empty", "minecraft:flowing_water", "minecraft:water", "minecraft:flowing_lava", "minecraft:lava"}
			for id, kind := range types {
				state, e := blockstate.BlockStateByID(uint32(id))
				if e != nil {
					t.Fatal(e)
				}
				if got := generationFluidType(SurfaceBlockState{Name: state.Name, Properties: state.Properties}); got != names[kind] {
					t.Fatalf("state %d %+v: got %s want %s", id, state, got, names[kind])
				}
			}
		default:
			t.Fatal("invalid reference row", row)
		}
	}
	if savedIndex != 20 {
		t.Fatal("incomplete saved tick cases", savedIndex)
	}
}

func TestGenerationFluidTicksRegionReplayAndClone(t *testing.T) {
	for _, borrowed := range []bool{false, true} {
		t.Run(fmt.Sprint(borrowed), func(t *testing.T) {
			chunks := []*NoiseChunkResult{}
			for x := int32(-2); x <= -1; x++ {
				chunks = append(chunks, &NoiseChunkResult{X: x, Z: 2, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)})
			}
			region, err := NewFeatureRegion(chunks...)
			if err != nil {
				t.Fatal(err)
			}
			region.borrowed = map[featureChunkPosition]bool{{-2, 2}: borrowed, {-1, 2}: borrowed}
			p := FeaturePosition{-17, -63, 47}
			if err := region.SetBlock(p, structureState("water")); err != nil {
				t.Fatal(err)
			}
			record := &structureBlockResult{}
			world := &structureRecordingWorld{region: region, result: record}
			for _, outside := range []FeaturePosition{{-17, -65, 47}, {30000000, 64, 47}} {
				if err := world.ScheduleFluidUpdate(outside); err != nil {
					t.Fatal(err)
				}
			}
			for i := 0; i < 2; i++ {
				if err := world.ScheduleFluidUpdate(p); err != nil {
					t.Fatal(err)
				}
			}
			if len(record.fluidTicks.ticks) != 1 {
				t.Fatal("structure did not deduplicate ticks")
			}
			// The scheduled type belongs to the original request, even if later
			// structure writes remove/replace its block before cached replay.
			if err := region.SetBlock(p, structureState("lava")); err != nil {
				t.Fatal(err)
			}
			if err := region.ScheduleFluidUpdate(p); err != nil {
				t.Fatal(err)
			}
			if err := region.ScheduleFluidTick(record.fluidTicks.ticks[0]); err != nil {
				t.Fatal(err)
			}
			got := region.cloneChunk(-2, 2)
			if len(got.FluidTicks) != 2 || got.FluidTicks[0].Fluid != "minecraft:water" || got.FluidTicks[1].Fluid != "minecraft:lava" {
				t.Fatalf("ticks %+v", got.FluidTicks)
			}
			if len(region.cloneChunk(-1, 2).FluidTicks) != 0 {
				t.Fatal("tick leaked into neighbor chunk")
			}
			index, _ := got.Index(p.X&15, p.Y, p.Z&15)
			if index < len(got.PostProcessing) && got.PostProcessing[index] {
				t.Fatal("scheduled tick incorrectly became PostProcessing")
			}
			got.FluidTicks[0].Fluid = "changed"
			if region.cloneChunk(-2, 2).FluidTicks[0].Fluid != "minecraft:water" {
				t.Fatal("clone aliased queue")
			}
			if len(chunks[0].FluidTicks) != 0 {
				t.Fatal("mutated input chunk queue")
			}
		})
	}
}

func TestGenerationFluidTicksPrecedingSnapshotIsolation(t *testing.T) {
	chunk := &NoiseChunkResult{X: -2, Z: 2, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)}
	first := PendingFluidTick{Position: FeaturePosition{-17, -63, 47}, Fluid: "minecraft:water", Priority: -2}
	snapshot := &structureChunkSnapshot{base: &overworldBase{terrain: chunk}, fluidTicks: []PendingFluidTick{first}}
	regions := make([]*OverworldFeatureRegion, 2)
	for i := range regions {
		region, err := NewFeatureRegion(chunk)
		if err != nil {
			t.Fatal(err)
		}
		region.borrowed = make(map[featureChunkPosition]bool)
		region.biomeBases = make(map[featureChunkPosition]*overworldBase)
		region.installStructureChunk(featureChunkPosition{-2, 2}, snapshot)
		regions[i] = region
	}
	duplicate := first
	duplicate.Priority = 3
	if err := regions[0].ScheduleFluidTick(duplicate); err != nil {
		t.Fatal(err)
	}
	flowing := first
	flowing.Fluid = "minecraft:flowing_water"
	if err := regions[0].ScheduleFluidTick(flowing); err != nil {
		t.Fatal(err)
	}
	left, right := regions[0].cloneChunk(-2, 2), regions[1].cloneChunk(-2, 2)
	if len(left.FluidTicks) != 2 || left.FluidTicks[0] != first || len(right.FluidTicks) != 1 || right.FluidTicks[0] != first || len(snapshot.fluidTicks) != 1 {
		t.Fatalf("cache copy lost ticks, first priority or isolation: left=%v right=%v snapshot=%v", left.FluidTicks, right.FluidTicks, snapshot.fluidTicks)
	}
}
