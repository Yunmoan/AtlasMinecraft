package worldgen

import (
	"atlas/internal/nbt"
	"math"
)

// The authored template BlockPosition clips placement; the actual double Pos
// determines biome lookup and entity Anvil ownership. Vanilla's camel_spawn
// deliberately gives these different coordinates, including after rotation.
func pendingEntityBlockPosition(entity PendingStructureEntity) FeaturePosition {
	values, ok := entity.Data["Pos"].Elements()
	if !ok || len(values) != 3 {
		return entity.Position
	}
	var coordinates [3]int32
	for i, tag := range values {
		value, ok := tag.Value.(float64)
		if !ok || tag.Type != nbt.Double || math.IsNaN(value) || math.IsInf(value, 0) {
			return entity.Position
		}
		coordinates[i] = int32(math.Floor(value))
	}
	return FeaturePosition{coordinates[0], coordinates[1], coordinates[2]}
}

// Precise entity coordinates use the original one-block mirror offset; integer
// block positions use zero. Keep pivot arithmetic in int32 before conversion.
func transformTemplateEntityPosition(p [3]float64, pivot FeaturePosition, mirror TemplateMirror, rotation TemplateRotation) [3]float64 {
	x, z := p[0], p[2]
	switch mirror {
	case TemplateMirrorLeftRight:
		z = 1 - z
	case TemplateMirrorFrontBack:
		x = 1 - x
	}
	switch rotation {
	case TemplateRotationClockwise90:
		return [3]float64{float64(pivot.X+pivot.Z+1) - z, p[1], float64(pivot.Z-pivot.X) + x}
	case TemplateRotationClockwise180:
		return [3]float64{float64(pivot.X+pivot.X+1) - x, p[1], float64(pivot.Z+pivot.Z+1) - z}
	case TemplateRotationCounterclockwise90:
		return [3]float64{float64(pivot.X-pivot.Z) + z, p[1], float64(pivot.X+pivot.Z+1) - x}
	default:
		return [3]float64{x, p[1], z}
	}
}

// StructureTemplate combines Entity.rotate and Entity.mirror in this order.
// moveTo keeps the resulting yaw without an additional [-180,180) wrap.
func transformTemplateEntityYaw(yaw float32, mirror TemplateMirror, rotation TemplateRotation) float32 {
	wrapped := float32(math.Mod(float64(yaw), 360))
	if wrapped >= 180 {
		wrapped -= 360
	}
	if wrapped < -180 {
		wrapped += 360
	}
	rotated := wrapped + float32(rotation)*90
	mirrored := wrapped
	switch mirror {
	case TemplateMirrorLeftRight:
		mirrored = 180 - wrapped
	case TemplateMirrorFrontBack:
		mirrored = -wrapped
	}
	delta := mirrored - yaw
	return rotated + delta
}
