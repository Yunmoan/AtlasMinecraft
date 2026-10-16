package server

import (
	"atlas/internal/proto"
	"atlas/internal/world"
	"math"
)

// LivingEntity.load applies persisted attributes before clamping Health. A
// village horse can retain its authored Health above the randomized maximum
// until its first save/load; immutable snapshots are encoded after that load.
func generatedSnapshotHealth(entity world.WorldEntity) (float32, bool) {
	data := entity.Data
	health, ok := data["Health"].Value.(float32)
	if !ok {
		return 0, false
	}
	if !entity.LoadedFromDisk {
		return health, true
	}
	attributes, _ := data["attributes"].Elements()
	for _, tag := range attributes {
		attribute, _ := tag.Object()
		id, _ := attribute.Text("id")
		if id != "minecraft:generic.max_health" {
			continue
		}
		base, valid := attribute["base"].Value.(float64)
		if !valid {
			continue
		}
		modifiers, _ := attribute["modifiers"].Elements()
		for _, tag := range modifiers {
			modifier, _ := tag.Object()
			operation, _ := modifier.Text("operation")
			if amount, valid := modifier["amount"].Value.(float64); valid && operation == "add_value" {
				base += amount
			}
		}
		value := base
		for _, tag := range modifiers {
			modifier, _ := tag.Object()
			operation, _ := modifier.Text("operation")
			if amount, valid := modifier["amount"].Value.(float64); valid && operation == "add_multiplied_base" {
				value += base * amount
			}
		}
		for _, tag := range modifiers {
			modifier, _ := tag.Object()
			operation, _ := modifier.Text("operation")
			if amount, valid := modifier["amount"].Value.(float64); valid && operation == "add_multiplied_total" {
				value *= 1 + amount
			}
		}
		maximum := float32(math.Max(1, math.Min(1024, value)))
		return max(float32(0), min(maximum, health)), true
	}
	return max(float32(0), health), true
}

func generatedArmorStandMetadata(id int32, entity world.WorldEntity) []byte {
	w := proto.NewWriter().VarInt(id)
	if invisible, _ := entity.Data.Number("Invisible"); invisible != 0 {
		w.U8(0).VarInt(0).U8(0x20)
	}
	if health, ok := generatedSnapshotHealth(entity); ok && health != 1 {
		w.U8(9).VarInt(3).F32(health)
	}
	var flags byte
	for _, entry := range []struct {
		key  string
		flag byte
	}{{"Small", 1}, {"ShowArms", 4}, {"NoBasePlate", 8}, {"Marker", 16}} {
		if value, _ := entity.Data.Number(entry.key); value != 0 {
			flags |= entry.flag
		}
	}
	if flags != 0 {
		w.U8(15).VarInt(0).U8(flags)
	}
	pose, _ := entity.Data.Object("Pose")
	for index, entry := range []struct {
		key          string
		defaultValue [3]float32
	}{
		{"Head", [3]float32{}}, {"Body", [3]float32{}}, {"LeftArm", [3]float32{-10, 0, -10}}, {"RightArm", [3]float32{-15, 0, 10}}, {"LeftLeg", [3]float32{-1, 0, -1}}, {"RightLeg", [3]float32{1, 0, 1}},
	} {
		values, ok := pose[entry.key].Elements()
		if !ok || len(values) != 3 {
			continue
		}
		var rotation [3]float32
		valid := true
		for i, tag := range values {
			var ok bool
			rotation[i], ok = tag.Value.(float32)
			valid = valid && ok
		}
		if valid && rotation != entry.defaultValue {
			w.U8(byte(16 + index)).VarInt(9).F32(rotation[0]).F32(rotation[1]).F32(rotation[2])
		}
	}
	return w.U8(0xff).Payload()
}

func appendGeneratedCamelMetadata(w *proto.Writer, entity world.WorldEntity) {
	// Camel.isTamed saves true even when a freshly finalized actor's tracked
	// flag is clear. AbstractHorse.load sets that bit on a disk-loaded actor.
	var flags byte
	if tame, _ := entity.Data.Number("Tame"); entity.LoadedFromDisk && tame != 0 {
		flags |= 2
	}
	if saddle, ok := entity.Data.Object("SaddleItem"); ok {
		if id, _ := saddle.Text("id"); id != "" && id != "minecraft:air" {
			flags |= 4
		}
	}
	if bred, _ := entity.Data.Number("Bred"); bred != 0 {
		flags |= 8
	}
	if eating, _ := entity.Data.Number("EatingHaystack"); eating != 0 {
		flags |= 16
	}
	if flags != 0 {
		w.U8(17).VarInt(0).U8(flags)
	}
	if tick, ok := entity.Data.Number("LastPoseTick"); ok && tick != 0 {
		w.U8(19).VarInt(2).VarLong(tick)
	}
}
