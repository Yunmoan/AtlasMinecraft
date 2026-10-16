package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type jigsawFeatureProbeWorld struct {
	*structureRecordingWorld
	placed bool
}

func (world *jigsawFeatureProbeWorld) PlaceJigsawFeature(raw json.RawMessage, random StructurePieceRandom, origin FeaturePosition) (bool, error) {
	placed, err := world.structureRecordingWorld.PlaceJigsawFeature(raw, random, origin)
	world.placed = placed
	return placed, err
}

func TestJigsawFeaturePoolAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/jigsaw_feature_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	regions := make(map[int]*OverworldFeatureRegion)
	cases, crossChunk := 0, 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 13 || fields[0] != "jigsaw_feature" {
			t.Fatalf("invalid feature vector%s", scanner.Text())
		}
		parse := func(index int) int64 {
			value, err := strconv.ParseInt(fields[index], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		origin := FeaturePosition{int32(parse(4)), int32(parse(5)), int32(parse(6))}
		mode := int(parse(3))
		region := regions[mode]
		if region == nil {
			region = sculkPatchProbeRegion(t, origin, mode)
			regions[mode] = region
		}
		region.overlays = nil
		region.blockEntities = nil
		result := &structureBlockResult{blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
		world := &jigsawFeatureProbeWorld{structureRecordingWorld: &structureRecordingWorld{region: region, result: result}}
		random := NewWorldgenRandom(parse(2))
		raw, _ := json.Marshal(fields[1])
		element := JigsawPoolElement{Type: "minecraft:feature_pool_element", Feature: raw, Projection: "rigid"}
		_, err := placeJigsawElement(world, element, origin, FeaturePosition{999, 999, 999}, TemplateRotation(parse(12)), "ignored", random, StructureBoundingBox{256, -64, 256, 271, 319, 271})
		if err != nil {
			t.Fatal(err)
		}
		var rows []string
		outside := false
		for position, state := range result.blocks {
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, fmt.Sprintf("%d %d %d %d", position.X, position.Y, position.Z, id))
			outside = outside || position.X>>4 != origin.X>>4 || position.Z>>4 != origin.Z>>4
		}
		sort.Strings(rows)
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
		if world.placed != (parse(7) == 1) || len(rows) != int(parse(8)) || int64(random.NextLong()) != parse(9) || hash != fields[10] || len(result.postProcessing) != int(parse(11)) {
			t.Fatalf("%s: placed%v writes%d hash%s marked%d", scanner.Text(), world.placed, len(rows), hash, len(result.postProcessing))
		}
		if region.structureRecording != nil {
			t.Fatal("structure recorder leaked into later features")
		}
		if outside {
			crossChunk++
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 480 || crossChunk == 0 {
		t.Fatalf("matched%d cases, cross-chunk%d", cases, crossChunk)
	}
	t.Logf("matched%d actual FeaturePoolElement calls, including%d cross-chunk outputs with a disjoint clip", cases, crossChunk)
}
