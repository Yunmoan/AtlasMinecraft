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
	file, err := os.Open("../../docs/worldgen/bubble_columns_1.21.1.txt")
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
	checked, postprocessed, neighborRows := 0, 0, 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) == 6 && f[0] == "bubble_neighbor" {
			// These rows record the JAR's actual neighbor scheduling rules.
			// This generation stage settles the tick result, not tick timing.
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
		// The production entry point must also match the independent vectors
		// when starting from either generated source block.
		if name == "magma_block" || name == "soul_sand" {
			chunk := &NoiseChunkResult{X: p.X >> 4, Z: p.Z >> 4, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)}
			initial := make(map[int]SurfaceBlockState)
			for q, state := range world.blocks {
				if q.X>>4 != chunk.X || q.Z>>4 != chunk.Z || q.Y < -64 || q.Y >= 320 {
					continue
				}
				index, _ := chunk.Index(q.X&15, q.Y, q.Z&15)
				initial[index] = state
				chunk.setSurfaceState(q.X&15, q.Y, q.Z&15, state)
			}
			n, err := PostProcessGeneratedBubbleColumns(chunk)
			if err != nil {
				t.Fatal(err)
			}
			edits := make(map[FeaturePosition]SurfaceBlockState)
			for index, state := range chunk.SurfaceBlocks {
				old := initial[index]
				oldID, err := blockstate.BlockStateID(old.Name, old.Properties)
				if err != nil {
					t.Fatal(err)
				}
				newID, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				if oldID != newID {
					edits[FeaturePosition{chunk.X*16 + int32(index&15), chunk.MinY + int32(index/256), chunk.Z*16 + int32((index>>4)&15)}] = state
				}
			}
			if n != wantCount || len(edits) != wantCount || hashEdits(edits) != wantHash {
				t.Fatalf("postprocess %s mode=%d at %v: changed=%d hash=%s, JAR changed=%d hash=%s", name, mode, p, n, hashEdits(edits), wantCount, wantHash)
			}
			if n, err := PostProcessGeneratedBubbleColumns(chunk); err != nil || n != 0 {
				t.Fatalf("bubble postprocess is not idempotent: changed=%d error=%v", n, err)
			}
			postprocessed++
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
	if checked != 960 || postprocessed != 384 || neighborRows != 48 {
		t.Fatalf("incomplete bubble coverage: direct=%d postprocessed=%d neighbor=%d", checked, postprocessed, neighborRows)
	}
	t.Logf("matched %d actual JAR column updates, including %d generated-chunk postprocessing cases", checked, postprocessed)
}

func BenchmarkGeneratedBubbleColumns(b *testing.B) {
	chunk := &NoiseChunkResult{X: 0, Z: 10, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)}
	chunk.setSurfaceState(15, 20, 15, SurfaceBlockState{Name: "minecraft:magma_block"})
	for y := int32(21); y <= 62; y++ {
		chunk.Blocks[(y+64)*256+15*16+15] = NoiseWater
	}
	b.ResetTimer()
	for range b.N {
		// Reset just this column, excluding terrain allocation and setup.
		for y := int32(21); y <= 62; y++ {
			delete(chunk.SurfaceBlocks, int((y+64)*256+15*16+15))
		}
		if n, err := PostProcessGeneratedBubbleColumns(chunk); err != nil || n != 42 {
			b.Fatalf("bubble postprocess changed=%d, error=%v", n, err)
		}
	}
}

func TestGeneratedBubbleColumnsCrossChunkOverlayIsolation(t *testing.T) {
	// A decoration source at (0,0) can write into the corner of (1,1).
	// Each destination is finalized from its private replay; processing it
	// must preserve the cached terrain and the other destination's result.
	source := &NoiseChunkResult{X: 0, Z: 0, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: make(map[int]SurfaceBlockState)}
	target := &NoiseChunkResult{X: 1, Z: 1, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: make(map[int]SurfaceBlockState)}
	for y := int32(40); y < 48; y++ {
		target.Blocks[(y+64)*256] = NoiseWater
	}
	region, err := NewOverworldFeatureRegion(source, target)
	if err != nil {
		t.Fatal(err)
	}
	region.borrowed = map[featureChunkPosition]bool{{0, 0}: true, {1, 1}: true}
	if err := region.SetBlock(FeaturePosition{16, 39, 16}, SurfaceBlockState{Name: "minecraft:soul_sand"}); err != nil {
		t.Fatal(err)
	}
	sourceResult, targetResult := region.cloneChunk(0, 0), region.cloneChunk(1, 1)
	if n, err := PostProcessGeneratedBubbleColumns(sourceResult); err != nil || n != 0 {
		t.Fatalf("source chunk should have no bubble changes: %d, %v", n, err)
	}
	if n, err := PostProcessGeneratedBubbleColumns(targetResult); err != nil || n != 8 {
		t.Fatalf("cross-chunk source was not finalized: changed=%d, error=%v", n, err)
	}
	for y := int32(40); y < 48; y++ {
		generated, err := targetResult.BlockStateAt(0, y, 0)
		if err != nil || generated.Name != "minecraft:bubble_column" || generated.Properties["drag"] != "false" {
			t.Fatalf("target column at Y=%d: %v, %v", y, generated, err)
		}
		cached, _ := target.BlockStateAt(0, y, 0)
		replayed, _ := region.BlockAt(FeaturePosition{16, y, 16})
		if cached.Name != "minecraft:water" || replayed.Name != "minecraft:water" {
			t.Fatalf("bubble processing mutated cached/replay water at Y=%d: cached=%v replay=%v", y, cached, replayed)
		}
	}
	// Properties must be private too, so a consumer cannot change every
	// bubble in the column by changing one returned state's map.
	first, _ := targetResult.BlockStateAt(0, 40, 0)
	first.Properties["drag"] = "true"
	second, _ := targetResult.BlockStateAt(0, 41, 0)
	if second.Properties["drag"] != "false" {
		t.Fatal("generated bubble states share mutable property maps")
	}
}
