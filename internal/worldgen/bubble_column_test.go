package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type bubbleFixtureWorld struct {
	blocks  map[FeaturePosition]SurfaceBlockState
	changed map[FeaturePosition]SurfaceBlockState
}

func (world *bubbleFixtureWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := world.blocks[p]; ok && world.InWorld(p) {
		return state, nil
	}
	return SurfaceBlockState{Name: "minecraft:air"}, nil
}

func (world *bubbleFixtureWorld) SetBlock(p FeaturePosition, state SurfaceBlockState) error {
	if !world.InWorld(p) {
		return fmt.Errorf("bubble fixture write outside bounds: %v", p)
	}
	world.blocks[p], world.changed[p] = state, state
	return nil
}

func (world *bubbleFixtureWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }

func TestBubbleColumnsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../testdata/worldgen/bubble_columns_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	parse := func(text string) int32 {
		n, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		return int32(n)
	}
	states := map[string]SurfaceBlockState{}
	for _, name := range []string{"air", "water", "magma_block", "soul_sand", "stone", "kelp", "kelp_plant", "seagrass", "tall_seagrass", "lava"} {
		states[name] = SurfaceBlockState{Name: "minecraft:" + name}
	}
	states["bubble_down"] = SurfaceBlockState{Name: "minecraft:bubble_column", Properties: map[string]string{"drag": "true"}}
	states["bubble_up"] = SurfaceBlockState{Name: "minecraft:bubble_column", Properties: map[string]string{"drag": "false"}}
	states["wet_slab"] = SurfaceBlockState{Name: "minecraft:oak_slab", Properties: map[string]string{"waterlogged": "true"}}
	for level := 0; level < 16; level++ {
		states[fmt.Sprintf("water_%d", level)] = SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": strconv.Itoa(level)}}
	}
	hashEdits := func(edits map[FeaturePosition]SurfaceBlockState) string {
		rows := make([]string, 0, len(edits))
		for p, state := range edits {
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
		}
		sort.Strings(rows)
		hash := sha256.Sum256([]byte(strings.Join(rows, "\n")))
		return hex.EncodeToString(hash[:])
	}
	checked, neighborRows := 0, 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) == 6 && f[0] == "bubble_neighbor" {
			p := FeaturePosition{8, 39, 8}
			w := blockShapeProbeWorld{fluidStepProbeWorld: &fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: states[f[1]]}, floor: -64}}
			face := map[string]int{"down": 0, "up": 1, "north": 2, "south": 3, "west": 4, "east": 5}[f[2]]
			if _, err := UpdateFluidNeighborShape(w, p, states[f[1]], face, states[f[3]]); err != nil {
				t.Fatal(err)
			}
			if len(w.events) != int(parse(f[4])) {
				t.Fatalf("%s: events %v", scanner.Text(), w.events)
			}
			if len(w.events) > 0 && w.events[0] != fmt.Sprintf("block_tick 8 39 8 minecraft:%s %s 0", f[1], f[5]) {
				t.Fatalf("%s: %v", scanner.Text(), w.events)
			}
			neighborRows++
			continue
		}
		if len(f) != 8 || f[0] != "bubble_column" {
			t.Fatalf("invalid bubble oracle row: %s", scanner.Text())
		}
		name, mode := f[1], int(parse(f[2]))
		p := FeaturePosition{parse(f[3]), parse(f[4]), parse(f[5])}
		wantCount, wantHash := int(parse(f[6])), f[7]
		world := &bubbleFixtureWorld{blocks: make(map[FeaturePosition]SurfaceBlockState), changed: make(map[FeaturePosition]SurfaceBlockState)}
		world.blocks[treeOffset(p, 0, -1, 0)] = states[name]
		top := min(int32(320), p.Y+8)
		if mode == 31 {
			top = 320
		}
		for y := p.Y; y < top; y++ {
			world.blocks[FeaturePosition{p.X, y, p.Z}] = states["water"]
			world.blocks[FeaturePosition{p.X + 1, y, p.Z}] = states["water"]
		}
		switch {
		case mode < 16:
			world.blocks[p] = states[fmt.Sprintf("water_%d", mode)]
		case mode == 17:
			world.blocks[treeOffset(p, 0, 3, 0)] = states["water_1"]
		case mode == 18:
			world.blocks[p] = states["kelp"]
		case mode >= 19 && mode <= 26:
			world.blocks[treeOffset(p, 0, 3, 0)] = states[[]string{"kelp_plant", "seagrass", "tall_seagrass", "wet_slab", "bubble_down", "bubble_up", "stone", "lava"}[mode-19]]
		case mode == 27:
			world.blocks[p] = states["air"]
		case mode == 28:
			world.blocks[treeOffset(p, 0, 3, 0)] = states["air"]
		case mode == 29:
			world.blocks[p] = states["bubble_up"]
		case mode == 30:
			for y := p.Y; y < top; y++ {
				world.blocks[FeaturePosition{p.X, y, p.Z}] = states["bubble_down"]
			}
		}
		n, err := UpdateBubbleColumn(world, p, states[name])
		if err != nil {
			t.Fatal(err)
		}
		if n != wantCount || len(world.changed) != wantCount || hashEdits(world.changed) != wantHash {
			t.Fatalf("%s mode=%d at %v: changed=%d hash=%s, JAR changed=%d hash=%s", name, mode, p, n, hashEdits(world.changed), wantCount, wantHash)
		}
		checked++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if checked != 960 || neighborRows != 48 {
		t.Fatalf("incomplete bubble coverage: direct=%d neighbor=%d", checked, neighborRows)
	}
	t.Logf("matched %d actual JAR column updates", checked)
}
