package worldgen

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type monumentProbeWorld struct {
	hutProbeWorld
	ticks map[FeaturePosition]bool
}

func (w *monumentProbeWorld) GenerationSeaLevel() int32 {
	if w.scenario == 1 {
		return 70
	}
	return 63
}
func (w *monumentProbeWorld) MinBuildHeight() int32 { return -64 }
func (w *monumentProbeWorld) MaxBuildHeight() int32 { return 320 }
func (w *monumentProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	w.ticks[p] = true
	return nil
}
func (w *monumentProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, found := w.edits[p]; found {
		return state, nil
	}
	floor := int32(30)
	mod := func(n, d int32) int32 {
		v := n % d
		if v < 0 {
			v += d
		}
		return v
	}
	if w.scenario == 1 {
		floor = 50 + mod(p.X*3+p.Z, 7)
	} else if w.scenario == 3 {
		floor = 80
	}
	if p.Y <= floor {
		return structureState("stone"), nil
	}
	if p.Y >= w.GenerationSeaLevel() {
		return structureState("air"), nil
	}
	if w.scenario == 2 {
		kind := mod(p.X+p.Y+p.Z, 11)
		if kind < 4 {
			return structureState([]string{"ice", "packed_ice", "blue_ice", "seagrass"}[kind]), nil
		}
	}
	return structureState("water"), nil
}

func TestMonumentBlocksAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/structure_monument_blocks_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	clips := [][4]int32{{-40, -4, 40, 80}, {-32, 0, -17, 15}, {-16, 16, -1, 31}, {16, 48, 31, 63}}
	digest := func(lines []string) string { sort.Strings(lines); return stringHexHash(strings.Join(lines, "\n")) }
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		if len(f) != 12 || f[0] != "monument_blocks" {
			t.Fatalf("invalid block vector %q", line)
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		seed := parse(1)
		layout, err := assembleMonument(NewLegacyRandom(seed), -29, 7, f[2])
		if err != nil {
			t.Fatal(err)
		}
		world := &monumentProbeWorld{hutProbeWorld: hutProbeWorld{scenario: int(parse(3)), edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}, ticks: make(map[FeaturePosition]bool)}
		box := clips[parse(4)]
		clip := StructureBoundingBox{box[0], -64, box[1], box[2], 319, box[3]}
		random := NewLegacyRandom(seed ^ 0x123456789abcdef)
		result, err := placeMonumentLayout(world, layout, random, clip, world.GenerationSeaLevel())
		if err != nil {
			t.Fatal(err)
		}
		blocks, _, err := structureBlockDigest(&world.hutProbeWorld)
		if err != nil {
			t.Fatal(err)
		}
		var ticks, actors []string
		for p := range world.ticks {
			ticks = append(ticks, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
		}
		for _, actor := range result.PendingEntities {
			if actor.ID != "minecraft:elder_guardian" {
				t.Fatal("unexpected monument actor", actor.ID)
			}
			p := actor.Position
			actors = append(actors, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
		}
		next := int64(random.NextLong())
		if len(world.edits) != int(parse(5)) || blocks != f[6] || len(ticks) != int(parse(7)) || digest(ticks) != f[8] || len(actors) != int(parse(9)) || digest(actors) != f[10] || next != parse(11) {
			t.Fatalf("%s: got %d blocks %s, %d ticks %s, %d actors %s, next %d", line, len(world.edits), blocks, len(ticks), digest(ticks), len(actors), digest(actors), next)
		}
		count++
	}
	if count != 192 {
		t.Fatalf("got %d block vectors, want 192", count)
	}
	t.Logf("matched %d original full/clipped monument placements: block states, fluid ticks, guardian positions and subsequent RNG", count)
}
