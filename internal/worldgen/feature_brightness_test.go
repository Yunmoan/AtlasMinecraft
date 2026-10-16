package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The fixture calls the actual LevelLightEngine with absent light sections,
// and with queued sky DataLayers at the mushroom survival threshold.
func TestMushroomFeatureAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/mushroom_feature_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		t.Fatal(err)
	}
	features := map[string]PlacedFeature{}
	biomes := map[string]string{}
	for _, feature := range order.features {
		features[feature.Name] = feature
	}
	for biome, steps := range order.biomes {
		for _, index := range steps[9] {
			for _, feature := range order.features {
				if feature.Step == 9 && feature.Index == index {
					biomes[feature.Name] = biome
				}
			}
		}
	}
	regions := map[[3]int32]*OverworldFeatureRegion{}
	brightnessRegion, err := NewOverworldFeatureRegion(&NoiseChunkResult{MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)})
	if err != nil {
		t.Fatal(err)
	}
	parse := func(s string) int64 {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	scanner := bufio.NewScanner(file)
	brightnessCases, survivalCases, patchCases, positivePatches := 0, 0, 0, 0
	positiveFeatures := map[string]bool{}
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		switch f[0] {
		case "feature_brightness":
			got, err := brightnessRegion.FeatureRawBrightness(FeaturePosition{8, int32(parse(f[1])), 8}, int32(parse(f[2])))
			if err != nil || got != int32(parse(f[3])) {
				t.Fatalf("raw brightness %d, %v; want %s", got, err, scanner.Text())
			}
			brightnessCases++
		case "mushroom_survival":
			region := vegetationProbeRegion(t, 0, 0, 0)
			p := FeaturePosition{8, 41, 8}
			if err := region.SetBlock(treeOffset(p, 0, -1, 0), SurfaceBlockState{Name: f[2]}); err != nil {
				t.Fatal(err)
			}
			if parse(f[3]) >= 0 {
				brightness := int32(parse(f[4]))
				region.BindFeatureBrightness(func(FeaturePosition, int32) (int32, error) { return brightness, nil })
			}
			got, err := featurePlantSurvives(region, p, SurfaceBlockState{Name: f[1]})
			if err != nil || got != (parse(f[5]) == 1) {
				t.Fatalf("mushroom survival %v, %v; want %s", got, err, scanner.Text())
			}
			survivalCases++
		case "vegetation":
			seed, mode := parse(f[3]), int32(parse(f[4]))
			origin := FeaturePosition{int32(parse(f[5])), int32(parse(f[6])), int32(parse(f[7]))}
			key := [3]int32{origin.X >> 4, origin.Z >> 4, mode}
			if regions[key] == nil {
				region := vegetationProbeRegion(t, key[0], key[1], int(mode))
				region.borrowed = map[featureChunkPosition]bool{}
				for pos := range region.chunks {
					region.borrowed[pos] = true
				}
				regions[key] = region
			}
			region := regions[key]
			region.overlays = nil
			region.heights = map[featureColumnPosition]int32{}
			region.motionHeights = map[featureColumnPosition]int32{}
			region.surfaceHeights = map[featureColumnPosition]int32{}
			r := NewWorldgenRandom(seed)
			name := f[2]
			var operations []placementOperation
			if f[1] == "placed" {
				feature := features[name]
				definition, err := LoadPlacedFeatureDefinition(name)
				if err != nil {
					t.Fatal(err)
				}
				name = definition.Feature
				operations, err = bindPlacement(region, definition, &feature, func(int32, int32, int32) (string, error) { return biomes[feature.Name], nil })
				if err != nil {
					t.Fatal(err)
				}
				r = OverworldPlacedFeatureRandom(OverworldDecorationSeed(seed, key[0], key[1]), feature)
				origin.Y = 0
			}
			place, err := compileFeaturePlan(name, 0)
			if err != nil {
				t.Fatal(err)
			}
			candidates, placed := 0, 0
			err = walkPlacement(operations, r, origin, func(p FeaturePosition, r *WorldgenRandom) error {
				candidates++
				ok, err := place(region, r, p)
				if ok {
					placed++
				}
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			var output []string
			for key, overlay := range region.overlays {
				for i, s := range overlay.blocks {
					p := FeaturePosition{key.chunkX*16 + int32(i&15), -64 + int32(i/256), key.chunkZ*16 + int32((i>>4)&15)}
					id, err := blockstate.BlockStateID(s.Name, s.Properties)
					if err != nil {
						t.Fatal(err)
					}
					output = append(output, fmt.Sprintf("%d %d %d %d", p.X, p.Y, p.Z, id))
				}
			}
			sort.Strings(output)
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(output, "\n"))))
			next := int64(r.NextLong())
			if candidates != int(parse(f[8])) || placed != int(parse(f[9])) || len(output) != int(parse(f[10])) || next != parse(f[11]) || hash != f[12] {
				t.Errorf("%s seed=%d mode=%d: candidates=%d placed=%d writes=%d next=%d hash=%s; want %s", f[2], seed, mode, candidates, placed, len(output), next, hash, scanner.Text())
			}
			patchCases++
			if placed > 0 {
				positivePatches++
				positiveFeatures[f[2]] = true
			}
		default:
			t.Fatalf("unrecognized fixture %s", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if brightnessCases != 12 || survivalCases != 90 || patchCases != 1144 || positivePatches != 52 || len(positiveFeatures) != 10 {
		t.Fatalf("brightness=%d survival=%d patches=%d positive patches=%d", brightnessCases, survivalCases, patchCases, positivePatches)
	}
	g, err := NewOverworldGenerator(0)
	if err != nil {
		t.Fatal(err)
	}
	for name := range positiveFeatures {
		if strings.Contains(name, ":patch_") {
			continue
		}
		if _, ok := g.plans[name]; !ok {
			t.Fatalf("verified mushroom placed feature %s is not enabled", name)
		}
	}
	t.Logf("matched %d real-light-engine brightness vectors, %d survival vectors and %d configured/placed feature vectors (%d positive patches)", brightnessCases, survivalCases, patchCases, positivePatches)
}
