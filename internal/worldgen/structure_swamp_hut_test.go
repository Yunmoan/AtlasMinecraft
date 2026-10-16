package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type hutProbeWorld struct {
	scenario int
	edits    map[FeaturePosition]SurfaceBlockState
	post     map[FeaturePosition]bool
	entities map[FeaturePosition]featureBlockEntity
}

func (world *hutProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, found := world.edits[p]; found {
		return state, nil
	}
	floor := int32(63)
	switch world.scenario {
	case 1:
		modulo := (p.X + 2*p.Z) % 5
		if modulo < 0 {
			modulo += 5
		}
		floor = 60 + modulo
	case 2, 3:
		floor = 60
	case 4:
		floor = -65
	}
	name := "minecraft:air"
	if p.Y <= floor {
		name = "minecraft:stone"
		if p.Y == floor {
			name = "minecraft:grass_block"
		}
	} else if world.scenario == 2 && p.Y <= 66 {
		name = "minecraft:water"
	} else if world.scenario == 3 && p.Y == 72 {
		name = "minecraft:oak_leaves"
	}
	return SurfaceBlockState{Name: name}, nil
}

func (world *hutProbeWorld) SetBlock(p FeaturePosition, state SurfaceBlockState) error {
	world.edits[p] = state
	if entity, found := world.entities[p]; found && entity.block != state.Name {
		delete(world.entities, p)
	}
	return nil
}
func (world *hutProbeWorld) SetBlockEntity(p FeaturePosition, block string, data nbt.Object) error {
	if world.entities == nil {
		world.entities = make(map[FeaturePosition]featureBlockEntity)
	}
	world.entities[p] = featureBlockEntity{block, data}
	return nil
}
func (world *hutProbeWorld) InWorld(p FeaturePosition) bool { return p.Y >= -64 && p.Y < 320 }
func (world *hutProbeWorld) MarkPostProcessing(p FeaturePosition) error {
	world.post[p] = true
	return nil
}

func structureBlockDigest(world *hutProbeWorld) (string, string, error) {
	edits := make([]string, 0, len(world.edits))
	for p, state := range world.edits {
		id, err := blockstate.BlockStateID(state.Name, state.Properties)
		if err != nil {
			return "", "", err
		}
		edits = append(edits, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
	}
	post := make([]string, 0, len(world.post))
	for p := range world.post {
		post = append(post, fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z))
	}
	sort.Strings(edits)
	sort.Strings(post)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(edits, "\n")))), fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(post, "\n")))), nil
}

func TestSwampHutBlocksAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "structure_hut_blocks_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 12 || f[0] != "hut" {
			t.Fatalf("invalid hut oracle vector %q", scanner.Text())
		}
		parse := func(index int) int32 {
			value, err := strconv.ParseInt(f[index], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			return int32(value)
		}
		orientation := parse(1)
		xSize, zSize := int32(7), int32(9)
		if orientation%2 != 0 {
			xSize, zSize = zSize, xSize
		}
		piece := StructurePiece{ID: "minecraft:tesh", Orientation: [...]string{"south", "west", "north", "east"}[orientation],
			Box: StructureBoundingBox{parse(3), 64, parse(4), parse(3) + xSize - 1, 70, parse(4) + zSize - 1},
			Data: nbt.Object{
				"HPos": {Type: nbt.Int, Value: int32(-1)}, "Witch": {Type: nbt.Byte, Value: int8(1)}, "Cat": {Type: nbt.Byte, Value: int8(1)},
			}}
		world := &hutProbeWorld{scenario: int(parse(2)), edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}
		result, err := PostProcessSwampHutBlocks(world, &piece, parse(5), parse(6))
		if err != nil || len(result.PendingEntities) != 0 {
			t.Fatalf("%s: hut block postprocess result %v (%v)", scanner.Text(), result, err)
		}
		height, _ := piece.Data["HPos"].Number()
		blockHash, postHash, err := structureBlockDigest(world)
		if err != nil {
			t.Fatal(err)
		}
		if height != int64(parse(7)) || len(world.edits) != int(parse(8)) || blockHash != f[9] || len(world.post) != int(parse(10)) || postHash != f[11] {
			t.Fatalf("%s:\n got HPos=%d blocks=%d/%s post=%d/%s", scanner.Text(), height, len(world.edits), blockHash, len(world.post), postHash)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 240 {
		t.Fatalf("got %d hut block vectors, want 240", count)
	}
	t.Logf("matched %d original hut geometry / clipping / height / stair / postprocessing cases", count)
}

func TestSwampHutEntitiesRemainExplicitlyPending(t *testing.T) {
	world := &hutProbeWorld{edits: make(map[FeaturePosition]SurfaceBlockState), post: make(map[FeaturePosition]bool)}
	start, err := GenerateScatteredStructureStart(0, "swamp_hut", 0, 0, 63,
		func(int32, int32, int32) (string, error) { return "minecraft:swamp", nil },
		func(int32, int32) (int32, error) { return 63, nil })
	if err != nil {
		t.Fatal(err)
	}
	result, err := PostProcessSwampHutBlocks(world, &start.Pieces[0], 0, 0)
	if err != nil || len(result.PendingEntities) != 2 {
		t.Fatalf("expected explicit witch and cat requests, got %v/%v", result, err)
	}
	for _, entity := range result.PendingEntities {
		flag, _ := start.Pieces[0].Data[entity.SavedFlag].Number()
		if flag != 0 || !entity.Persistent {
			t.Fatalf("pending entity marked as fulfilled: %v", entity)
		}
	}
}
