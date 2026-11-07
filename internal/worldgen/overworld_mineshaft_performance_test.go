package worldgen

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"testing"
)

func minePerformanceStart(tb testing.TB, g *OverworldGenerator, mesa bool) StructureStart {
	tb.Helper()
	x, z, name := int32(-16), int32(22), "minecraft:mineshaft"
	if mesa {
		x, z, name = -365, 369, "minecraft:mineshaft_mesa"
	}
	starts, err := g.structureStartsAt(x, z)
	if err != nil {
		tb.Fatal(err)
	}
	for _, start := range starts {
		if start.Structure == name {
			return start
		}
	}
	tb.Fatalf("performance fixture rejected at %d,%d", x, z)
	return StructureStart{}
}

// Keep the original full bounding-box snapshot preparation as an independent
// optimization baseline. The production replay requests only actual piece clips.
func fullBoxMineshaftReplay(g *OverworldGenerator, start StructureStart) (*structureBlockResult, error) {
	order, err := OriginalStructureDecorationOrder(start.Structure)
	if err != nil {
		return nil, err
	}
	region, err := g.structureRegionBefore(start.Chunk.X, start.Chunk.Z, order.Step)
	if err != nil {
		return nil, err
	}
	box, _ := start.BoundingBox()
	if err = g.extendJigsawReadRegion(region, box, order.Step); err != nil {
		return nil, err
	}
	biomes, err := newOverworldBiomeSampler(g.seed, g.noise.settings)
	if err != nil {
		return nil, err
	}
	region.biomeAtQuart = func(x, y, z int32) (string, error) {
		return biomes.BiomeAtQuart(x, max(int32(-16), min(int32(79), y)), z)
	}
	return g.replayMineshaftStructureBlocks(start, region, order)
}

func mineshaftRecordDigest(result *structureBlockResult) [32]byte {
	// fmt sorts maps, including nested state properties and NBT compounds.
	return sha256.Sum256([]byte(fmt.Sprintf("%v %v %v %v %v %v", result.start, result.blocks, result.postProcessing, result.blockEntities, result.pending, result.fluidTicks.ticks)))
}

func TestMineshaftRuntimeOptimizationParity(t *testing.T) {
	for _, mesa := range []bool{false, true} {
		t.Run(fmt.Sprintf("mesa=%t", mesa), func(t *testing.T) {
			g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
			if err != nil {
				t.Fatal(err)
			}
			start := minePerformanceStart(t, g, mesa)
			original, err := fullBoxMineshaftReplay(g, start)
			if err != nil {
				t.Fatal(err)
			}
			optimized, err := g.generateMineshaftStructureBlocks(start)
			if err != nil {
				t.Fatal(err)
			}
			want, got := mineshaftRecordDigest(original), mineshaftRecordDigest(optimized)
			if want != got {
				t.Fatalf("optimized mine changed blocks/flags/updates/NBT/actors: got%x want%x", got, want)
			}
			if !slices.Equal(original.postProcessingOrder, optimized.postProcessingOrder) {
				t.Fatal("optimized mine changed ordered postprocessing events")
			}
			// Version 13 records fluid requests separately from PostProcessing;
			// include their ordered type/position/priority records in this digest.
			baseline := "be43582a7c7dc50ded24fc8f8ee4e2a9aac6788459dc04baa030cb1dbd67c9fa"
			if mesa {
				// Version 13 freezes WG heights at CARVERS, as independently
				// checked by TestGenerationHeightmapsAgainstVanillaJAR. Cutting
				// exposed mesa passages no longer lowers this generation map.
				baseline = "e5c2902f4b4f92f72e1715445b2030c184b39e79d31fd38f136bcba9b7ffe4c3"
			}
			if fmt.Sprintf("%x", got) != baseline {
				t.Fatalf("mine optimization baseline changed: got%x want%s", got, baseline)
			}
			t.Logf("preserved complete mine record digest: %x", got)
		})
	}
}

// One operation includes a new generator and all cold preceding terrain.
// Use -benchtime=1x to compare eager empty-gap snapshots with actual piece clips.
func BenchmarkMineshaftCanonicalReplay(b *testing.B) {
	for _, mesa := range []bool{false, true} {
		for _, fullBox := range []bool{true, false} {
			b.Run(fmt.Sprintf("mesa=%t/full_box=%t", mesa, fullBox), func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					g, err := NewOverworldGeneratorWithOptions(1234, OverworldGeneratorOptions{GenerateStructures: true})
					if err != nil {
						b.Fatal(err)
					}
					start := minePerformanceStart(b, g, mesa)
					var result *structureBlockResult
					if fullBox {
						result, err = fullBoxMineshaftReplay(g, start)
					} else {
						result, err = g.generateMineshaftStructureBlocks(start)
					}
					if err != nil {
						b.Fatal(err)
					}
					if len(result.blocks) == 0 {
						b.Fatal("mine benchmark produced no blocks")
					}
				}
			})
		}
	}
}
