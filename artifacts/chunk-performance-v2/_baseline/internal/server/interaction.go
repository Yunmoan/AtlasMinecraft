package server

import (
	"atlas/internal/blockrules"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"atlas/internal/worldgen"
	"math"
	"strings"
)

func playerHeight(p *player.Player) float64 {
	if p.Crouching() {
		return 1.5
	}
	return 1.8
}

func canReachBlock(p *player.Player, target world.BlockPosition) bool {
	x, y, z, _, _, _ := p.Pos()
	eye := 1.62
	if p.Crouching() {
		eye = 1.27
	}
	y += eye
	var distance float64
	for axis, value := range []float64{x, y, z} {
		minimum := []float64{float64(target.X), float64(target.Y), float64(target.Z)}[axis]
		delta := max(minimum-value, max(0, value-minimum-1))
		distance += delta * delta
	}
	// Creative block interaction range (5) plus the vanilla packet tolerance (1).
	return distance < 36
}

func (s *Server) resendBlock(p *player.Player, position world.BlockPosition) {
	states, err := s.World.ReadLoadedBlocks([]world.BlockPosition{position})
	if err != nil {
		return
	}
	p.SendPacket(proto.SBlockChange, proto.NewWriter().I64(proto.PackPosition(position.X, position.Y, position.Z)).VarInt(int32(states[0])).Payload())
}

func horizontalFacing(yaw float32) string {
	index := int(math.Floor(float64(yaw)/90+0.5)) & 3
	return []string{"south", "west", "north", "east"}[index]
}

func (s *Server) handlePlace(p *player.Player, payload []byte) {
	r := proto.NewReader(payload)
	hand, err := r.VarInt()
	if err != nil {
		return
	}
	x, y, z, err := r.Position()
	if err != nil {
		return
	}
	face, err := r.VarInt()
	if err != nil {
		return
	}
	var cursor [3]float32
	for i := range cursor {
		cursor[i], err = r.F32()
		if err != nil {
			return
		}
	}
	if _, err = r.Bool(); err != nil {
		return
	}
	sequence, err := r.VarInt()
	if err != nil || sequence < 0 || r.Remaining() != 0 {
		return
	}
	defer p.SendPacket(proto.SAckBlockChange, proto.NewWriter().VarInt(sequence).Payload())
	clicked := world.BlockPosition{X: x, Y: y, Z: z}
	if face < 0 || face > 5 {
		s.resendBlock(p, clicked)
		return
	}
	dx, dy, dz := faceOffset(face)
	target := world.BlockPosition{X: x + dx, Y: y + dy, Z: z + dz}
	defer func() {
		s.resendBlock(p, clicked)
		s.resendBlock(p, target)
		s.resendBlock(p, world.BlockPosition{X: target.X, Y: target.Y + 1, Z: target.Z})
	}()
	if hand < 0 || hand > 1 || !canReachBlock(p, clicked) || y < world.MinY || y >= world.MaxY {
		return
	}
	for _, value := range cursor {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || value < 0 || value > 1 {
			return
		}
	}
	read, err := s.World.ReadLoadedBlocks([]world.BlockPosition{clicked})
	if err != nil {
		return
	}
	clickedState, err := world.BlockStateByID(read[0])
	if err != nil {
		return
	}
	inventory, _ := p.Inventory()
	held := inventory.Slots[36+int(inventory.Selected)]
	if hand == 1 {
		held = inventory.Slots[45]
	}
	if !p.Crouching() || held.Empty() {
		if s.useBlock(p, clicked, read[0], clickedState) {
			return
		}
	}
	heldBlock := p.HeldBlock()
	if hand == 1 {
		heldBlock = world.StateAir
		if !held.Empty() {
			heldBlock = world.ItemBlock[held.ID]
		}
	}
	state, err := world.BlockStateByID(heldBlock)
	if err != nil || state.Name == "minecraft:air" {
		return
	}
	clickedRules, err := blockrules.Lookup(read[0])
	if err != nil {
		return
	}
	properties := state.Properties
	mergeSlab := strings.HasSuffix(state.Name, "_slab") && clickedState.Name == state.Name && clickedState.Properties["type"] != "double" &&
		((clickedState.Properties["type"] == "bottom" && (face == 1 || face >= 2 && cursor[1] > .5)) || (clickedState.Properties["type"] == "top" && (face == 0 || face >= 2 && cursor[1] <= .5)))
	if clickedRules.Replaceable || mergeSlab {
		target = clicked
	}
	if target.Y < world.MinY || target.Y >= world.MaxY {
		return
	}
	read, err = s.World.ReadLoadedBlocks([]world.BlockPosition{target})
	if err != nil {
		return
	}
	expected := read[0]
	targetState, err := world.BlockStateByID(expected)
	if err != nil {
		return
	}
	targetRules, err := blockrules.Lookup(expected)
	if err != nil || !targetRules.Replaceable && !mergeSlab {
		return
	}
	if properties == nil {
		properties = make(map[string]string)
	}
	if _, hasAxis := properties["axis"]; hasAxis {
		properties["axis"] = "y"
		if face == 2 || face == 3 {
			properties["axis"] = "z"
		}
		if face == 4 || face == 5 {
			properties["axis"] = "x"
		}
	}
	if _, waterlogged := properties["waterlogged"]; waterlogged {
		properties["waterlogged"] = "false"
		if targetState.Name == "minecraft:water" && targetState.Properties["level"] == "0" {
			properties["waterlogged"] = "true"
		}
	}
	_, _, _, yaw, _, _ := p.Pos()
	if _, facing := properties["facing"]; facing && !strings.HasSuffix(state.Name, "_stairs") {
		facing := horizontalFacing(yaw)
		properties["facing"] = oppositeFacing(facing)
		if strings.HasSuffix(state.Name, "_fence_gate") {
			properties["facing"] = facing
		}
		if state.Name == "minecraft:piston" || state.Name == "minecraft:sticky_piston" || state.Name == "minecraft:observer" || state.Name == "minecraft:dispenser" || state.Name == "minecraft:dropper" {
			_, _, _, _, pitch, _ := p.Pos()
			if pitch > 45 {
				properties["facing"] = "up"
			} else if pitch < -45 {
				properties["facing"] = "down"
			}
		}
	}
	if state.Name == "minecraft:torch" || state.Name == "minecraft:soul_torch" || state.Name == "minecraft:redstone_torch" {
		if face == 0 {
			return
		}
		if face >= 2 {
			wallName := strings.TrimSuffix(state.Name, "torch") + "wall_torch"
			state.Name = wallName
			properties = map[string]string{"facing": []string{"down", "up", "north", "south", "west", "east"}[face]}
		}
	}
	if state.Name == "minecraft:lantern" || state.Name == "minecraft:soul_lantern" {
		properties["hanging"] = "false"
		if face == 0 {
			properties["hanging"] = "true"
		}
	}
	if strings.HasSuffix(state.Name, "_button") || state.Name == "minecraft:lever" {
		properties["face"] = "wall"
		properties["facing"] = []string{"north", "north", "north", "south", "west", "east"}[face]
		if face == 0 {
			properties["face"] = "ceiling"
			properties["facing"] = horizontalFacing(yaw)
		}
		if face == 1 {
			properties["face"] = "floor"
			properties["facing"] = horizontalFacing(yaw)
		}
		properties["powered"] = "false"
	}
	if strings.HasSuffix(state.Name, "_trapdoor") {
		properties["open"] = "false"
		properties["half"] = "bottom"
		properties["facing"] = horizontalFacing(yaw)
		if face >= 2 {
			properties["facing"] = []string{"north", "north", "north", "south", "west", "east"}[face]
		}
		if face == 0 || face >= 2 && cursor[1] > .5 {
			properties["half"] = "top"
		}
	}
	if strings.HasSuffix(state.Name, "_stairs") {
		properties["facing"] = horizontalFacing(yaw)
		properties["shape"] = "straight"
		properties["half"] = "bottom"
		if face == 0 || face >= 2 && cursor[1] > .5 {
			properties["half"] = "top"
		}
	}
	if strings.HasSuffix(state.Name, "_slab") {
		properties["type"] = "bottom"
		if face == 0 || face >= 2 && cursor[1] > .5 {
			properties["type"] = "top"
		}
		if mergeSlab {
			properties["type"] = "double"
			properties["waterlogged"] = "false"
		}
	}
	if strings.HasSuffix(state.Name, "_leaves") {
		properties["persistent"] = "true"
	}
	var upper *uint32
	if world.IsDoor(state.Name) {
		properties["facing"] = horizontalFacing(yaw)
		properties["open"] = "false"
		properties["powered"] = "false"
		properties["half"] = "upper"
		properties["hinge"] = doorHinge(properties["facing"], cursor[0], cursor[2])
		upperID, e := world.BlockStateID(state.Name, properties)
		if e != nil {
			return
		}
		upper = &upperID
		properties["half"] = "lower"
	}
	if worldgen.SupportedPlant(state.Name) {
		below, err := s.World.ReadLoadedBlocks([]world.BlockPosition{{X: target.X, Y: target.Y - 1, Z: target.Z}})
		if err != nil {
			return
		}
		ground, err := world.BlockStateByID(below[0])
		if err != nil || !worldgen.PlantGroundSurvives(state.Name, ground.Name) {
			return
		}
		if worldgen.DoublePlant(state.Name) {
			properties["half"] = "upper"
			id, err := world.BlockStateID(state.Name, properties)
			if err != nil {
				return
			}
			upper = &id
			properties["half"] = "lower"
		}
	}
	id, err := world.BlockStateID(state.Name, properties)
	if err != nil {
		return
	}
	if supported, e := s.World.CanSupportBlock(target, id); e != nil || !supported {
		return
	}
	s.entityMu.Lock()
	defer s.entityMu.Unlock()
	blocked := false
	s.Hub.Each(func(other *player.Player) {
		ox, oy, oz, _, _, _ := other.Pos()
		hit, e := world.BlockIntersectsPlayer(id, target, ox, oy, oz, playerHeight(other))
		if e != nil || hit {
			blocked = true
		}
		if upper != nil {
			hit, e := world.BlockIntersectsPlayer(*upper, world.BlockPosition{X: target.X, Y: target.Y + 1, Z: target.Z}, ox, oy, oz, playerHeight(other))
			if e != nil || hit {
				blocked = true
			}
		}
	})
	if blocked {
		return
	}
	if err := s.World.TryPlaceBlock(target, expected, id, upper); err != nil {
		s.Dbg("placement rejected for %s: %v", p.Name, err)
	}
}
