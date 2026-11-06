package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"fmt"
	"strings"
)

func fallingBlock(name string) bool   { return blockTickHandlers[name] == "diu" }
func IsFallingBlock(name string) bool { return fallingBlock(name) }

func fallingDelay(name string) int32 {
	if name == "minecraft:dragon_egg" {
		return 5
	}
	return 2
}

func scheduleFallingBlock(world any, p FeaturePosition, name string) error {
	w, ok := world.(interface {
		ScheduleBlockTick(FeaturePosition, string, int32, int32) error
	})
	if !ok {
		return fmt.Errorf("falling block update requires block scheduling")
	}
	return w.ScheduleBlockTick(p, name, fallingDelay(name), 0)
}

func NotifyFallingBlockPlacement(world any, p FeaturePosition, state SurfaceBlockState) error {
	if !fallingBlock(state.Name) {
		return nil
	}
	return scheduleFallingBlock(world, p, state.Name)
}

// ConcretePowderBlock.touchesLiquid checks the original current position for
// its DOWN exception, then the other five neighbors in Direction order. Merely
// standing above water does not harden dry powder before it starts falling.
func concreteTouchesWater(world FeatureBlockWorld, p FeaturePosition) (bool, error) {
	for face, d := range [6]FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}} {
		if face == 0 {
			current, err := world.BlockAt(p)
			if err != nil {
				return false, err
			}
			fluid, err := FluidStateForBlock(current)
			if err != nil {
				return false, err
			}
			if !fluid.Type.same(FluidWater) {
				continue
			}
		}
		s, err := world.BlockAt(treeOffset(p, d.X, d.Y, d.Z))
		if err != nil {
			return false, err
		}
		fluid, err := FluidStateForBlock(s)
		if err != nil {
			return false, err
		}
		if !fluid.Type.same(FluidWater) {
			continue
		}
		id, err := blockstate.BlockStateID(s.Name, s.Properties)
		if err != nil {
			return false, err
		}
		rules, err := blockrules.Lookup(id)
		if err != nil {
			return false, err
		}
		if !rules.FaceSturdy[face^1] {
			return true, nil
		}
	}
	return false, nil
}

func updateFallingNeighborShape(world FeatureBlockWorld, p FeaturePosition, state SurfaceBlockState) (SurfaceBlockState, error) {
	if strings.HasSuffix(state.Name, "_concrete_powder") {
		wet, err := concreteTouchesWater(world, p)
		if err != nil {
			return state, err
		}
		if wet {
			return SurfaceBlockState{Name: strings.TrimSuffix(state.Name, "_powder")}, nil
		}
	}
	return state, scheduleFallingBlock(world, p, state.Name)
}

// FallingBlockSpawn is the result of the original scheduled FallingBlock.tick.
// The consumer must replace Source by its fluid legacy block (flags 3), then
// add the entity. Data is constructor NBT without the separately assigned UUID
// and entity type. It includes FallingBlock.falling's post-creation settings.
type FallingBlockSpawn struct {
	Source      FeaturePosition
	Replacement SurfaceBlockState
	Data        nbt.Object
}

func tickFallingBlock(world ScheduledBlockWorld, p FeaturePosition, state SurfaceBlockState) error {
	below, err := world.BlockAt(treeOffset(p, 0, -1, 0))
	if err != nil {
		return err
	}
	id, err := blockstate.BlockStateID(below.Name, below.Properties)
	if err != nil {
		return err
	}
	rules, err := blockrules.Lookup(id)
	if err != nil {
		return err
	}
	if !rules.FallThrough {
		return nil
	}
	bounds, ok := world.(interface{ MinBuildHeight() int32 })
	if !ok {
		return fmt.Errorf("falling block tick requires minimum build height")
	}
	if p.Y < bounds.MinBuildHeight() {
		return nil
	}
	spawn, err := newFallingBlockSpawn(p, state)
	if err != nil {
		return err
	}
	if brushableBlock(state.Name) {
		spawn.Data["CancelDrop"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	}
	consumer, ok := world.(interface{ SpawnFallingBlock(FallingBlockSpawn) error })
	if !ok {
		return fmt.Errorf("%w: %s requires falling entity execution", ErrUnsupportedScheduledBlock, state.Name)
	}
	return consumer.SpawnFallingBlock(spawn)
}

func newFallingBlockSpawn(p FeaturePosition, state SurfaceBlockState) (FallingBlockSpawn, error) {
	fluid, err := FluidStateForBlock(state)
	if err != nil {
		return FallingBlockSpawn{}, err
	}
	replacement := structureState("air")
	if fluid.Type != FluidEmpty {
		replacement = fluid.LegacyBlock()
	}
	state = cloneSurfaceState(state)
	if _, ok := state.Properties["waterlogged"]; ok {
		state.Properties["waterlogged"] = "false"
	}
	block := nbt.Object{"Name": {Type: nbt.String, Value: state.Name}}
	if len(state.Properties) > 0 {
		props := nbt.Object{}
		for k, v := range state.Properties {
			props[k] = nbt.Tag{Type: nbt.String, Value: v}
		}
		block["Properties"] = nbt.Tag{Type: nbt.Compound, Value: props}
	}
	list := func(values ...float64) nbt.Tag {
		items := make([]nbt.Tag, len(values))
		for i, v := range values {
			items[i] = nbt.Tag{Type: nbt.Double, Value: v}
		}
		return nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: items}
	}
	hurt, amount := int8(0), float32(0)
	if state.Name == "minecraft:anvil" || state.Name == "minecraft:chipped_anvil" || state.Name == "minecraft:damaged_anvil" {
		hurt, amount = 1, 2
	}
	data := nbt.Object{
		"Air": {Type: nbt.Short, Value: int16(300)}, "Fire": {Type: nbt.Short, Value: int16(-1)},
		"BlockState": {Type: nbt.Compound, Value: block}, "Time": {Type: nbt.Int, Value: int32(0)},
		"DropItem": {Type: nbt.Byte, Value: int8(1)}, "CancelDrop": {Type: nbt.Byte, Value: int8(0)},
		"HurtEntities": {Type: nbt.Byte, Value: hurt}, "FallHurtAmount": {Type: nbt.Float, Value: amount}, "FallHurtMax": {Type: nbt.Int, Value: int32(40)},
		"FallDistance": {Type: nbt.Float, Value: float32(0)}, "Invulnerable": {Type: nbt.Byte, Value: int8(0)}, "OnGround": {Type: nbt.Byte, Value: int8(0)},
		"PortalCooldown": {Type: nbt.Int, Value: int32(0)}, "Motion": list(0, 0, 0), "Pos": list(float64(p.X)+0.5, float64(p.Y), float64(p.Z)+0.5),
		"Rotation": {Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{{Type: nbt.Float, Value: float32(0)}, {Type: nbt.Float, Value: float32(0)}}},
	}
	return FallingBlockSpawn{p, replacement, data}, nil
}
