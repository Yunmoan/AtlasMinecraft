package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"math"
	"testing"
)

func TestCamelTemplatePreciseChunkOwnership(t *testing.T) {
	element := JigsawPoolElement{Type: "minecraft:legacy_single_pool_element", Location: "minecraft:village/desert/camel_spawn", Projection: "rigid"}
	template, err := LoadJigsawTemplate(element.Location)
	if err != nil || len(template.Entities) != 1 {
		t.Fatalf("original camel asset: %v", err)
	}
	if template.Entities[0].BlockPosition != (FeaturePosition{0, 0, 1}) || template.Entities[0].Position != ([3]float64{}) {
		t.Fatal("camel ownership regression lost the independently audited original coordinates")
	}
	different := 0
	for rotation := TemplateRotationNone; rotation <= TemplateRotationCounterclockwise90; rotation++ {
		for _, x := range []int32{-17, -16, -1, 0, 15, 16} {
			for _, z := range []int32{-16, -1, 0, 15, 16} {
				origin := FeaturePosition{x, 64, z}
				anchor, _ := TransformTemplatePosition(template.Entities[0].BlockPosition, FeaturePosition{}, TemplateMirrorNone, rotation)
				anchor = treeOffset(origin, anchor.X, anchor.Y, anchor.Z)
				cx, cz := anchor.X>>4, anchor.Z>>4
				world := &jigsawProcessorProbeWorld{ruinedPortalProbeWorld: &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: 0, edits: make(map[FeaturePosition]SurfaceBlockState)}}}
				clip := StructureBoundingBox{cx * 16, -64, cz * 16, cx*16 + 15, 319, cz*16 + 15}
				placed, err := placeJigsawElement(world, element, origin, origin, rotation, "ignore_waterlogging", NewWorldgenRandom(0), clip)
				if err != nil || len(placed.PendingEntities) != 1 {
					t.Fatalf("original camel clipped call %v/%d: %v", origin, rotation, err)
				}
				entity := placed.PendingEntities[0]
				values, _ := entity.Data["Pos"].Elements()
				wantX := int32(math.Floor(values[0].Value.(float64))) >> 4
				wantZ := int32(math.Floor(values[2].Value.(float64))) >> 4
				region := &OverworldFeatureRegion{minY: -64, height: 384}
				if err := region.AddPendingEntity(entity); err != nil {
					t.Fatal(err)
				}
				requests := region.PendingEntitiesInChunk(wantX, wantZ)
				if len(requests) != 1 || requests[0].Position != anchor {
					t.Fatalf("camel did not reach its actual Anvil chunk %d,%d, while retaining template clip anchor %v", wantX, wantZ, anchor)
				}
				if cx != wantX || cz != wantZ {
					if len(region.PendingEntitiesInChunk(cx, cz)) != 0 {
						t.Fatal("camel duplicated into its authored anchor's chunk")
					}
					different++
				}
				requests[0].Data["Pos"] = nbt.Tag{}
				if len(region.PendingEntitiesInChunk(wantX, wantZ)) != 1 {
					t.Fatal("caller mutation changed pending entity ownership")
				}
			}
		}
	}
	if different == 0 {
		t.Fatal("camel fixture did not cross an actual chunk ownership boundary")
	}
	t.Logf("120 actual original camel template placements, %d with distinct clip/Anvil chunks; all four rotations and negative boundaries", different)
}

func TestPendingEntityPreciseNegativeCoordinates(t *testing.T) {
	entity := PendingStructureEntity{ID: "minecraft:cow", Position: FeaturePosition{0, 64, 0}, Data: nbt.Object{"Pos": animalPositionNBT([3]float64{-.25, 64, -16.25})}}
	region := &OverworldFeatureRegion{minY: -64, height: 384}
	if err := region.AddPendingEntity(entity); err != nil {
		t.Fatal(err)
	}
	if got := region.PendingEntitiesInChunk(-1, -2); len(got) != 1 {
		t.Fatal(fmt.Sprintf("precise negative coordinates were truncated rather than floored: %v", got))
	}
}
