package worldgen

import (
	"atlas/internal/nbt"
	"math"
	"reflect"
	"testing"
)

func TestVillageRuntimeUsesPreciseBiomeAndWorldAge(t *testing.T) {
	g := &OverworldGenerator{seed: 1234}
	// DayTime may be changed independently: it must not become camel pose age.
	g.SetStructureDayTimeSource(func() int64 { return 240000 })
	var age int64
	g.SetStructureGameTimeSource(func() int64 { return age })
	element := JigsawPoolElement{Type: "minecraft:legacy_single_pool_element", Location: "minecraft:village/desert/camel_spawn", Projection: "rigid"}
	for rotation := TemplateRotationNone; rotation <= TemplateRotationCounterclockwise90; rotation++ {
		origin := FeaturePosition{-16, 64, 15}
		world := &jigsawProcessorProbeWorld{ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: 0, edits: make(map[FeaturePosition]SurfaceBlockState)}}}
		placed, err := placeJigsawElement(world, element, origin, origin, rotation, "ignore_waterlogging", NewWorldgenRandom(0), StructureBoundingBox{-32, -64, 0, 15, 319, 31})
		if err != nil || len(placed.PendingEntities) != 1 {
			t.Fatal("actual camel template actor unavailable", err)
		}
		input := placed.PendingEntities[0]
		angles, ok := input.Data["Rotation"].Elements()
		if !ok || len(angles) != 2 || angles[0].Value != float32(rotation)*90 || angles[1].Value != float32(0) {
			t.Fatal("camel without authored Rotation lost original template rotation", rotation, input.Data["Rotation"])
		}
		values, _ := input.Data["Pos"].Elements()
		p := FeaturePosition{int32(math.Floor(values[0].Value.(float64))), int32(math.Floor(values[1].Value.(float64))), int32(math.Floor(values[2].Value.(float64)))}
		wantX, wantY, wantZ := BiomeQuartAtBlock(g.seed, p.X, p.Y, p.Z)
		calls := 0
		region := &OverworldFeatureRegion{minY: -64, height: 384, biomeAtQuart: func(x, y, z int32) (string, error) {
			calls++
			if x != wantX || y != wantY || z != wantZ {
				t.Fatalf("runtime queried authored anchor instead of actual entity biome: %d,%d,%d want %d,%d,%d", x, y, z, wantX, wantY, wantZ)
			}
			return "minecraft:desert", nil
		}}
		for _, sample := range []struct{ age, pose int64 }{{0, 0}, {1, 0}, {53, 0}, {54, 1}, {6000, 5947}, {24000, 23947}, {-1, 0}} {
			age = sample.age
			request := input
			request.Data = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: input.Data}).Value.(nbt.Object)
			if err := g.finalizeVillageRequest(region, &request, NewStructureRegionRandom(g.seed, p.X>>4, p.Z>>4)); err != nil {
				t.Fatal(err)
			}
			if got, _ := request.Data.Number("LastPoseTick"); got != sample.pose {
				t.Fatalf("age %d pose %d want %d", age, got, sample.pose)
			}
			if request.Position != input.Position || !reflect.DeepEqual(request.Data["Pos"], input.Data["Pos"]) || !reflect.DeepEqual(request.Data["Rotation"], input.Data["Rotation"]) {
				t.Fatal("finalization changed original template coordinates/clip anchor")
			}
		}
		if calls != 7 {
			t.Fatal("runtime did not query the actual seeded biome for every actor")
		}
	}
}
