package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"testing"
)

type blockShapeProbeWorld struct {
	*fluidStepProbeWorld
	brightness int32
}

func (w blockShapeProbeWorld) SetBlock(p FeaturePosition, s SurfaceBlockState) error {
	return w.SetFluidBlock(p, s, 2)
}
func (w blockShapeProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }
func (w blockShapeProbeWorld) FeatureRawBrightness(FeaturePosition, int32) (int32, error) {
	return w.brightness, nil // Defined native probe has constant raw brightness.
}
func (w blockShapeProbeWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	s, err := w.BlockAt(p)
	if err != nil {
		return err
	}
	f, err := FluidStateForBlock(s)
	if err != nil {
		return err
	}
	return w.ScheduleFluid(p, f.Type, FluidTickDelay(f.Type, false))
}

func (w blockShapeProbeWorld) ScheduleBlockTick(p FeaturePosition, name string, delay, priority int32) error {
	w.events = append(w.events, fmt.Sprintf("block_tick %d %d %d %s %d %d", p.X, p.Y, p.Z, name, delay, priority))
	return nil
}

// Optional whole-registry diagnostic. Unsupported runtime block families remain
// visible here; this is deliberately not an acceptance test that blesses them.
func TestBlockShapeRegistryDiagnostic(t *testing.T) {
	if os.Getenv("ATLAS_BLOCK_SHAPE_DIAGNOSTIC") == "" {
		t.Skip("optional original block-shape inventory")
	}
	runBlockShapeReference(t, "registry")
}

func TestPlantNeighborShapesAgainstVanilla(t *testing.T) {
	runBlockShapeReference(t, "plants")
}

func TestLeafNeighborShapesAgainstVanilla(t *testing.T) { runBlockShapeReference(t, "leaves") }

func TestBubbleNeighborShapesAgainstVanilla(t *testing.T)  { runBlockShapeReference(t, "bubbles") }
func TestFallingNeighborShapesAgainstVanilla(t *testing.T) { runBlockShapeReference(t, "falling") }
func TestLiquidNeighborShapesAgainstVanilla(t *testing.T)  { runBlockShapeReference(t, "liquids") }
func TestKelpNeighborShapesAgainstVanilla(t *testing.T)    { runBlockShapeReference(t, "kelp") }

func runBlockShapeReference(t *testing.T, scope string) {
	diagnostic := scope == "registry"
	f, err := os.Open("../../testdata/worldgen/block_shapes_1.21.1.jsonl.gz")
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
	counts := make(map[string]int)
	first := make(map[string]string)
	count := 0
	checked := 0
	for scan.Scan() {
		var row struct {
			State, Result uint32
			Face, Scene   int
			Events        []string
			Next          string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		s, err := blockstate.BlockStateByID(row.State)
		if err != nil {
			t.Fatal(err)
		}
		state := SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		if int(row.State)*36+row.Face*6+row.Scene != count {
			t.Fatal("missing or reordered reference", count)
		}
		count++
		if scope == "plants" && !runtimePlantShape(s.Name) || scope == "leaves" && blockTickHandlers[s.Name] != "dki" {
			continue
		}
		if scope == "bubbles" && s.Name != "minecraft:magma_block" && s.Name != "minecraft:soul_sand" && s.Name != "minecraft:bubble_column" {
			continue
		}
		if scope == "falling" && !fallingBlock(s.Name) {
			continue
		}
		if scope == "liquids" && s.Name != "minecraft:water" && s.Name != "minecraft:lava" {
			continue
		}
		if scope == "kelp" && !kelpBlock(s.Name) {
			continue
		}
		checked++
		p := FeaturePosition{8, 64, 8}
		neighbors := []SurfaceBlockState{structureState("air"), structureState("stone"), structureState("water"), structureState("grass_block"), structureState("farmland"), state}
		d := [6]FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}}[row.Face]
		world := blockShapeProbeWorld{&fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{p: state, treeOffset(p, d.X, d.Y, d.Z): neighbors[row.Scene]}, floor: 63}, 15}
		world.shapeRandom = NewLegacyRandom(int64(row.State)*31 + int64(row.Scene*6+row.Face))
		got, err := UpdateFluidNeighborShape(world, p, state, row.Face, neighbors[row.Scene])
		id, idErr := blockstate.BlockStateID(got.Name, got.Properties)
		if err != nil || idErr != nil || id != row.Result || !slices.Equal(world.events, row.Events) {
			if !diagnostic {
				t.Fatalf("%s id=%d face=%d scene=%d result=%d want=%d events=%v want=%v err=%v/%v", s.Name, row.State, row.Face, row.Scene, id, row.Result, world.events, row.Events, err, idErr)
			}
			counts[s.Name]++
			if first[s.Name] == "" {
				first[s.Name] = fmt.Sprintf("state=%d face=%d scene=%d result=%d want=%d events=%v want=%v err=%v/%v", row.State, row.Face, row.Scene, id, row.Result, world.events, row.Events, err, idErr)
			}
		}
		if !diagnostic {
			next, err := strconv.ParseUint(row.Next, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			if world.shapeRandom.NextLong() != next {
				t.Fatal("plant unexpectedly consumes randomness", row)
			}
		}
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(counts))
	total := 0
	for key, n := range counts {
		keys = append(keys, key)
		total += n
	}
	sort.Strings(keys)
	for _, key := range keys {
		t.Logf("%s: %d %s", key, counts[key], first[key])
	}
	t.Logf("%d mismatches across %d block types / %d calls", total, len(keys), count)
	t.Logf("checked %d callbacks", checked)
	if count != 26684*36 {
		t.Fatal("incomplete reference", count)
	}
	if scope == "plants" && checked != 6912 || scope == "leaves" && checked != 10080 || scope == "bubbles" && checked != 144 || scope == "falling" && checked != 1152 || scope == "liquids" && checked != 1152 {
		t.Fatal("plant callback coverage changed", checked)
	}
	if scope == "kelp" && checked != 972 {
		t.Fatal("incomplete kelp callbacks", checked)
	}
}

func TestPlantShapeContextsAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/plant_shapes_1.21.1.jsonl.gz")
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
	count, checked := 0, 0
	seen := map[[3]int]bool{}
	for scan.Scan() {
		var row struct {
			State, Result uint32
			Face, Scene   int
			Brightness    int32
			Inputs        [][4]int32
			Events        []string
			Next          string
		}
		if err := json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		key := [3]int{int(row.State), row.Face, row.Scene}
		if seen[key] {
			t.Fatal("duplicate context", key)
		}
		seen[key] = true
		count++
		s, err := blockstate.BlockStateByID(row.State)
		if err != nil {
			t.Fatal(err)
		}
		if !runtimePlantShape(s.Name) {
			continue
		}
		checked++
		state := SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		p := FeaturePosition{8, 64, 8}
		w := blockShapeProbeWorld{&fluidStepProbeWorld{blocks: map[FeaturePosition]SurfaceBlockState{}, floor: 63}, row.Brightness}
		for _, input := range row.Inputs {
			s, err := blockstate.BlockStateByID(uint32(input[3]))
			if err != nil {
				t.Fatal(err)
			}
			w.blocks[FeaturePosition{input[0], input[1], input[2]}] = SurfaceBlockState{Name: s.Name, Properties: s.Properties}
		}
		direction := [6]string{"down", "up", "north", "south", "west", "east"}[row.Face]
		neighbor, err := w.BlockAt(iglooDirectionOffset(p, direction))
		if err != nil {
			t.Fatal(err)
		}
		got, err := UpdateFluidNeighborShape(w, p, state, row.Face, neighbor)
		id, idErr := blockstate.BlockStateID(got.Name, got.Properties)
		if err != nil || idErr != nil || id != row.Result || !slices.Equal(w.events, row.Events) {
			t.Fatalf("%s state=%d face=%d context=%d result=%d want=%d events=%v want=%v err=%v/%v", s.Name, row.State, row.Face, row.Scene, id, row.Result, w.events, row.Events, err, idErr)
		}
		next, err := strconv.ParseUint(row.Next, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if uint64(NewLegacyRandom(int64(row.State)*31+int64(row.Scene*6+row.Face)).NextLong()) != next {
			t.Fatal("plant context consumed random", row)
		}
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 32100 || checked != 28800 {
		t.Fatal("incomplete contexts", count, checked)
	}
	t.Logf("matched %d plant contexts", checked)
}
