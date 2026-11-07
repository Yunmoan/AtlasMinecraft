package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"fmt"
	"math"
)

// FallingEntityWorld owns block writes, game events and item creation. A failed
// callback is an error, not permission to silently discard a falling entity.
type FallingEntityWorld interface {
	ScheduledBlockWorld
	MinBuildHeight() int32
	MaxBuildHeight() int32
	FallingGameEvent(FeaturePosition, string) error
	DropFallingBlockItem([3]float64, SurfaceBlockState) error
}

// TickFallingEntity implements the shared movement/landing path for terrain
// sand, red sand, gravel and brushable blocks. Other specialized falling families
// need their landing and damage callbacks before entering this entity phase.
func TickFallingEntity(world FallingEntityWorld, input nbt.Object) (nbt.Object, bool, error) {
	data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: input}).Value.(nbt.Object)
	block, ok := data.Object("BlockState")
	if !ok {
		return nil, false, fmt.Errorf("falling entity has no BlockState")
	}
	name, _ := block.Text("Name")
	if name != "minecraft:sand" && name != "minecraft:red_sand" && name != "minecraft:gravel" && !brushableBlock(name) {
		return nil, false, fmt.Errorf("falling entity handler is not implemented: %s", name)
	}
	state := SurfaceBlockState{Name: name}
	if props, ok := block.Object("Properties"); ok {
		state.Properties = make(map[string]string, len(props))
		for key := range props {
			state.Properties[key], _ = props.Text(key)
		}
	}
	vector := func(key string) ([3]float64, error) {
		var v [3]float64
		list, ok := data[key].Elements()
		if !ok || len(list) != 3 {
			return v, fmt.Errorf("invalid falling %s", key)
		}
		for i, t := range list {
			n, ok := t.Value.(float64)
			if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
				return v, fmt.Errorf("invalid falling %s", key)
			}
			v[i] = n
		}
		return v, nil
	}
	pos, err := vector("Pos")
	if err != nil {
		return nil, false, err
	}
	velocity, err := vector("Motion")
	if err != nil {
		return nil, false, err
	}
	age, _ := data.Number("Time")
	data["Time"] = nbt.Tag{Type: nbt.Int, Value: int32(age + 1)}
	noGravity, _ := data.Number("NoGravity")
	if noGravity == 0 {
		velocity[1] -= 0.04
	}
	w := float64(float32(0.98))
	half := w / 2
	box := blockrules.Box{pos[0] - half, pos[1], pos[2] - half, pos[0] + half, pos[1] + w, pos[2] + half}
	movement, err := fallingCollision(world, box, velocity)
	if err != nil {
		return nil, false, err
	}
	length := movement[0]*movement[0] + movement[1]*movement[1] + movement[2]*movement[2]
	if length > 1e-7 {
		for axis := range pos {
			pos[axis] += movement[axis]
		}
	}
	onGround := velocity[1] != movement[1] && velocity[1] < 0
	data["OnGround"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	if onGround {
		data["OnGround"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	}
	p := FeaturePosition{int32(math.Floor(pos[0])), int32(math.Floor(pos[1])), int32(math.Floor(pos[2]))}
	distance, _ := data["FallDistance"].Value.(float32)
	if onGround {
		if distance > 0 {
			if err := world.FallingGameEvent(p, "minecraft:hit_ground"); err != nil {
				return nil, false, err
			}
		}
		distance = 0
	} else if movement[1] < 0 {
		distance -= float32(movement[1])
	}
	data["FallDistance"] = nbt.Tag{Type: nbt.Float, Value: distance}
	for _, axis := range []int{0, 2} {
		if math.Abs(velocity[axis]-movement[axis]) >= 1e-5 {
			velocity[axis] = 0
		}
	}
	if velocity[1] != movement[1] {
		velocity[1] = 0
	}
	removed := false
	if onGround {
		velocity[0] *= 0.7
		velocity[1] *= -0.5
		velocity[2] *= 0.7
		target, err := world.BlockAt(p)
		if err != nil {
			return nil, false, err
		}
		if target.Name != "minecraft:moving_piston" {
			targetID, err := blockstate.BlockStateID(target.Name, target.Properties)
			if err != nil {
				return nil, false, err
			}
			rules, err := blockrules.Lookup(targetID)
			if err != nil {
				return nil, false, err
			}
			below, err := world.BlockAt(treeOffset(p, 0, -1, 0))
			if err != nil {
				return nil, false, err
			}
			belowID, err := blockstate.BlockStateID(below.Name, below.Properties)
			if err != nil {
				return nil, false, err
			}
			support, err := blockrules.Lookup(belowID)
			if err != nil {
				return nil, false, err
			}
			cancel, _ := data.Number("CancelDrop")
			if cancel == 0 && rules.Replaceable && !support.FallThrough {
				if err := world.SetFluidBlock(p, state, 3); err != nil {
					return nil, false, err
				}
			} else if cancel == 0 {
				if drop, _ := data.Number("DropItem"); drop != 0 {
					if err := breakDroppedFallingBrushable(world, pos, state); err != nil {
						return nil, false, err
					}
					if err := world.DropFallingBlockItem(pos, state); err != nil {
						return nil, false, err
					}
				}
			} else if err := breakFallingBrushable(world, pos, state); err != nil {
				return nil, false, err
			}
			removed = true
		}
	} else if age+1 > 600 || age+1 > 100 && (p.Y <= world.MinBuildHeight() || p.Y > world.MaxBuildHeight()) {
		if drop, _ := data.Number("DropItem"); drop != 0 {
			if err := world.DropFallingBlockItem(pos, state); err != nil {
				return nil, false, err
			}
		}
		removed = true
	}
	for i := range velocity {
		velocity[i] *= 0.98
	}
	for key, vec := range map[string][3]float64{"Pos": pos, "Motion": velocity} {
		items := make([]nbt.Tag, 3)
		for i, value := range vec {
			items[i] = nbt.Tag{Type: nbt.Double, Value: value}
		}
		data[key] = nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: items}
	}
	return data, removed, nil
}

func breakFallingBrushable(world FallingEntityWorld, pos [3]float64, state SurfaceBlockState) error {
	if !brushableBlock(state.Name) {
		return nil
	}
	callback, ok := world.(interface {
		BreakFallingBrushableBlock([3]float64, SurfaceBlockState) error
	})
	if !ok {
		return fmt.Errorf("brushable falling entity requires break callback")
	}
	pos[1] += float64(float32(0.98)) / 2 // Bounding-box center, not entity feet.
	return callback.BreakFallingBrushableBlock(pos, state)
}

func breakDroppedFallingBrushable(world FallingEntityWorld, pos [3]float64, state SurfaceBlockState) error {
	if !brushableBlock(state.Name) {
		return nil
	}
	rules, ok := world.(interface{ FallingEntityDropsEnabled() (bool, error) })
	if !ok {
		return fmt.Errorf("brushable falling entity requires doEntityDrops")
	}
	enabled, err := rules.FallingEntityDropsEnabled()
	if err != nil || !enabled {
		return err
	}
	return breakFallingBrushable(world, pos, state)
}

func fallingCollision(world ScheduledBlockWorld, box blockrules.Box, delta [3]float64) ([3]float64, error) {
	var bounds blockrules.Box
	for axis := 0; axis < 3; axis++ {
		bounds[axis] = math.Min(box[axis], box[axis]+delta[axis])
		bounds[axis+3] = math.Max(box[axis+3], box[axis+3]+delta[axis])
	}
	var obstacles []blockrules.Box
	for y := int32(math.Floor(bounds[1])) - 1; y <= int32(math.Floor(bounds[4]))+1; y++ {
		for z := int32(math.Floor(bounds[2])) - 1; z <= int32(math.Floor(bounds[5]))+1; z++ {
			for x := int32(math.Floor(bounds[0])) - 1; x <= int32(math.Floor(bounds[3]))+1; x++ {
				s, err := world.BlockAt(FeaturePosition{x, y, z})
				if err != nil {
					return delta, err
				}
				id, err := blockstate.BlockStateID(s.Name, s.Properties)
				if err != nil {
					return delta, err
				}
				rules, err := blockrules.Lookup(id)
				if err != nil {
					return delta, err
				}
				if rules.Offset && len(rules.Boxes) > 0 {
					return delta, fmt.Errorf("falling collision requires block offset: %s", s.Name)
				}
				for _, shape := range rules.Boxes {
					for axis, offset := range []float64{float64(x), float64(y), float64(z)} {
						shape[axis] += offset
						shape[axis+3] += offset
					}
					obstacles = append(obstacles, shape)
				}
			}
		}
	}
	order := [3]int{1, 0, 2}
	if math.Abs(delta[0]) < math.Abs(delta[2]) {
		order = [3]int{1, 2, 0}
	}
	for _, axis := range order {
		other1, other2 := (axis+1)%3, (axis+2)%3
		for _, obstacle := range obstacles {
			if box[other1+3]-obstacle[other1] <= 1e-7 || obstacle[other1+3]-box[other1] <= 1e-7 || box[other2+3]-obstacle[other2] <= 1e-7 || obstacle[other2+3]-box[other2] <= 1e-7 {
				continue
			}
			if delta[axis] > 0 && box[axis+3] <= obstacle[axis]+1e-7 {
				delta[axis] = math.Min(delta[axis], math.Max(0, obstacle[axis]-box[axis+3]))
			}
			if delta[axis] < 0 && box[axis] >= obstacle[axis+3]-1e-7 {
				delta[axis] = math.Max(delta[axis], math.Min(0, obstacle[axis+3]-box[axis]))
			}
		}
		box[axis] += delta[axis]
		box[axis+3] += delta[axis]
	}
	return delta, nil
}
