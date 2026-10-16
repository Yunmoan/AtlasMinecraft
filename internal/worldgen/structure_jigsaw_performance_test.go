package worldgen

import (
	"atlas/internal/nbt"
	"crypto/sha256"
	"fmt"
	"testing"
)

var jigsawPerformanceTemplates = []struct {
	name, location string
}{
	{"bastion", "minecraft:bastion/treasure/bases/lava_basin"},
	{"ancient_city", "minecraft:ancient_city/city_center/city_center_1"},
	{"trial_chambers", "minecraft:trial_chambers/chamber/chamber_1"},
	{"trail_ruins", "minecraft:trail_ruins/tower/tower_1"},
}

func jigsawPerformancePiece(tb testing.TB, location string) (StructurePiece, FeaturePosition, int32, int32) {
	tb.Helper()
	registry, err := loadJigsawRegistry()
	if err != nil {
		tb.Fatal(err)
	}
	var selected JigsawPoolElement
	var collect func(JigsawPoolElement)
	collect = func(element JigsawPoolElement) {
		if element.Location == location {
			selected = element
		}
		for _, child := range element.Elements {
			collect(child)
		}
	}
	for _, pool := range registry.Pools {
		for _, entry := range pool.Elements {
			collect(entry.Element)
		}
	}
	if selected.Location == "" {
		tb.Fatalf("missing benchmark element %s", location)
	}
	origin := FeaturePosition{15, 64, 15}
	box, err := jigsawElementBox(selected, origin, TemplateRotationNone)
	if err != nil {
		tb.Fatal(err)
	}
	piece, err := (&jigsawPlacedPiece{element: selected, origin: origin, box: box}).saved()
	if err != nil {
		tb.Fatal(err)
	}
	return piece, FeaturePosition{origin.X + 5, 48, origin.Z - 7}, (box.MinX + (box.MaxX-box.MinX+1)/2) >> 4, (box.MinZ + (box.MaxZ-box.MinZ+1)/2) >> 4
}

func jigsawPerformanceWorld() *jigsawProcessorProbeWorld {
	return &jigsawProcessorProbeWorld{ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: 2, edits: make(map[FeaturePosition]SurfaceBlockState)}}, seed: 1234}
}

func jigsawPerformanceDigest(tb testing.TB, piece *StructurePiece, pivot FeaturePosition, cx, cz int32) string {
	tb.Helper()
	world := jigsawPerformanceWorld()
	random := NewWorldgenRandom(1251)
	result, err := PostProcessJigsawBlocks(world, piece, random, cx, cz, pivot)
	if err != nil {
		tb.Fatal(err)
	}
	blocks, _, err := structureBlockDigest(world.hutProbeWorld)
	if err != nil {
		tb.Fatal(err)
	}
	canonical, err := jigsawCanonicalNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{}})
	if err != nil {
		tb.Fatal(err)
	}
	text := fmt.Sprintf("%s %s %v %v %d", blocks, canonical, world.entities, result.PendingEntities, int64(random.NextLong()))
	return fmt.Sprintf("%x", sha256.Sum256([]byte(text)))
}

func TestJigsawPerformanceClipOutput(t *testing.T) {
	want := map[string]string{
		"bastion":        "6d10e577c04db7dd24acc2d8dcb1080d2950a1838cbeb35a9572660783f85f40",
		"ancient_city":   "abce61344b9798919e8ebb779544fc29c04d0ad4b3b407a5d365b6db12b348ea",
		"trial_chambers": "46222e0c8a17072cf15db37ecf29bcc4d5a89c5d16bc87c828b87613e8c333ec",
		"trail_ruins":    "3807fd9a2ad2721ade6587ba6198d64f65508bdfc6d111e4ae2a913c82bb79cf",
	}
	for _, entry := range jigsawPerformanceTemplates {
		piece, pivot, cx, cz := jigsawPerformancePiece(t, entry.location)
		got := jigsawPerformanceDigest(t, &piece, pivot, cx, cz)
		if got != want[entry.name] {
			t.Fatalf("%s clip %d,%d output changed: %s", entry.name, cx, cz, got)
		}
		t.Logf("%s clip=%d,%d digest=%s", entry.name, cx, cz, got)
	}
}

// Defined flat terrain isolates template decoding, processors, NBT handling,
// clipping and placement from runtime neighborhood generation. Each operation
// starts with a fresh writable world. cold_template also removes just that
// original template from the process cache before decoding it again.
func BenchmarkJigsawTemplateClip(b *testing.B) {
	for _, entry := range jigsawPerformanceTemplates {
		b.Run(entry.name, func(b *testing.B) {
			piece, pivot, cx, cz := jigsawPerformancePiece(b, entry.location)
			if digest := jigsawPerformanceDigest(b, &piece, pivot, cx, cz); len(digest) != 64 {
				b.Fatal("invalid clip digest")
			}
			for _, cold := range []bool{false, true} {
				name := "cached_template"
				if cold {
					name = "cold_template"
				}
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						if cold {
							jigsawTemplateCache.Delete(entry.location)
						}
						world := jigsawPerformanceWorld()
						_, err := PostProcessJigsawBlocks(world, &piece, NewWorldgenRandom(1251), cx, cz, pivot)
						if err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

func BenchmarkJigsawCachedLookups(b *testing.B) {
	location := jigsawPerformanceTemplates[0].location
	processor := []byte(`"minecraft:treasure_rooms"`)
	if _, err := LoadJigsawTemplate(location); err != nil {
		b.Fatal(err)
	}
	if _, err := loadJigsawProcessors(processor); err != nil {
		b.Fatal(err)
	}
	b.Run("template", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := LoadJigsawTemplate(location); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("processors", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := loadJigsawProcessors(processor); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("feature_origins", func(b *testing.B) {
		piece, _, _, _ := jigsawPerformancePiece(b, location)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			origins, err := jigsawFeatureOrigins(piece)
			if err != nil || len(origins) != 0 {
				b.Fatal("unexpected feature origins", err)
			}
		}
	})
	b.Run("start_connectors", func(b *testing.B) {
		piece, _, _, _ := jigsawPerformancePiece(b, jigsawPerformanceTemplates[1].location)
		element, origin, rotation, _, err := jigsawSavedBlockSettings(&piece)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := jigsawConnectors(element, origin, rotation, NewLegacyRandom(1234)); err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := jigsawConnectors(element, origin, rotation, NewLegacyRandom(1234)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
