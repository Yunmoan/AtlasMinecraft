package world

import (
	"atlas/internal/blockrules"
	"atlas/internal/light"
	"atlas/internal/proto"
	"atlas/internal/worldgen"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrChunkNotLoaded = errors.New("chunk is not loaded")
var ErrBlockChanged = errors.New("block changed before the edit was applied")

type BlockPosition struct{ X, Y, Z int32 }
type blockRead struct {
	Positions []BlockPosition
	Result    chan blockReadResult
}
type blockReadResult struct {
	States []uint32
	Err    error
}

// ReadLoadedBlocks never generates terrain in response to a movement or click.
func (w *World) ReadLoadedBlocks(positions []BlockPosition) ([]uint32, error) {
	groups := make(map[chunkKey][]int)
	for i, p := range positions {
		if p.Y >= MinY && p.Y < MaxY {
			groups[chunkKey{p.X >> 9, p.Z >> 9}] = append(groups[chunkKey{p.X >> 9, p.Z >> 9}], i)
		}
	}
	states := make([]uint32, len(positions))
	for _, indices := range groups {
		points := make([]BlockPosition, len(indices))
		for i, index := range indices {
			points[i] = positions[index]
		}
		result := make(chan blockReadResult, 1)
		if err := w.post(context.Background(), points[0].X>>4, points[0].Z>>4, blockRead{Positions: points, Result: result}); err != nil {
			return nil, err
		}
		var read blockReadResult
		select {
		case read = <-result:
		case <-w.done:
			return nil, ErrWorldClosed
		}
		if read.Err != nil {
			return nil, read.Err
		}
		for i, index := range indices {
			states[index] = read.States[i]
		}
	}
	return states, nil
}

func (w *World) BlockAt(x, y, z int32) (uint32, error) {
	positions := []BlockPosition{{x, y, z}}
	states, err := w.ReadLoadedBlocks(positions)
	if errors.Is(err, ErrChunkNotLoaded) {
		if _, err = w.LoadChunk(x>>4, z>>4); err != nil {
			return 0, err
		}
		states, err = w.ReadLoadedBlocks(positions)
	}
	if err != nil {
		return 0, err
	}
	return states[0], nil
}

// TryPlaceBlock checks the clicked snapshot in the actor. Double plants are
// committed together with their upper half, updates and the final Anvil write.
func (w *World) TryPlaceBlock(p BlockPosition, expected, state uint32, upper *uint32) error {
	if p.Y < MinY || p.Y >= MaxY || state > MaxBlockState || upper != nil && (p.Y+1 >= MaxY || *upper > MaxBlockState) {
		return errors.New("invalid block placement")
	}
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	changed, err := w.editBlock(setBlockMsg{x: p.X, y: p.Y, z: p.Z, state: state, expected: &expected, upper: upper})
	if err != nil {
		return err
	}
	return w.updateBlockNeighbors(changed)
}

// settleColumn provides the initial block-update implementation: supported
// plants lose invalid support; sand/red sand/gravel settle vertically. Falling
// entities and their animation, item drops and other tick behaviors are future
// work. The whole column mutation is persisted in one chunk write.
func settleColumn(ch *Chunk, x, z int, start int32, change func(int32, uint32)) error {
	for y := max(int32(MinY), start); y < MaxY; y++ {
		id := ch.GetBlock(x, z, y)
		if isAir(id) {
			continue
		}
		state, err := BlockStateByID(id)
		if err != nil {
			return err
		}
		if state.Name == "minecraft:sand" || state.Name == "minecraft:red_sand" || state.Name == "minecraft:gravel" {
			target := y
			for target > MinY {
				below, err := blockrules.Lookup(ch.GetBlock(x, z, target-1))
				if err != nil {
					return err
				}
				if !below.FallThrough {
					break
				}
				target--
			}
			if target != y {
				change(y, StateAir)
				if target > MinY {
					change(target, id)
				}
			}
			continue
		}
	}
	// A falling block can land below the originally edited Y and replace the
	// lower half of a plant. Validate plants after all vertical moves, including
	// positions below the source; otherwise the upper half remains suspended.
	for y := int32(MinY); y < MaxY; y++ {
		id := ch.GetBlock(x, z, y)
		if isAir(id) {
			continue
		}
		state, err := BlockStateByID(id)
		if err != nil {
			return err
		}
		if !worldgen.SupportedPlant(state.Name) {
			continue
		}
		below, err := BlockStateByID(ch.GetBlock(x, z, y-1))
		if err != nil {
			return err
		}
		survives := worldgen.PlantGroundSurvives(state.Name, below.Name)
		if worldgen.DoublePlant(state.Name) {
			if state.Properties["half"] == "upper" {
				survives = below.Name == state.Name && below.Properties["half"] == "lower"
			} else {
				above, err := BlockStateByID(ch.GetBlock(x, z, y+1))
				if err != nil {
					return err
				}
				survives = survives && above.Name == state.Name && above.Properties["half"] == "upper"
			}
		}
		if !survives {
			change(y, StateAir)
		}
	}
	return nil
}

func (r *Region) applyColumnEdit(ch *Chunk, msg setBlockMsg) error {
	var changed []BlockPosition
	if msg.changed != nil {
		defer func() { msg.changed <- changed }()
	}
	if r.world.store == nil {
		return ErrVanillaGenerationUnavailable
	}
	x, z := int(msg.x&15), int(msg.z&15)
	if msg.expected != nil && ch.GetBlock(x, z, msg.y) != *msg.expected {
		return ErrBlockChanged
	}
	if msg.upper != nil {
		if msg.upperExpected != nil {
			if ch.GetBlock(x, z, msg.y+1) != *msg.upperExpected {
				return ErrBlockChanged
			}
		} else if !isAir(ch.GetBlock(x, z, msg.y+1)) {
			return ErrBlockChanged
		}
	}
	previous := make(map[int32]uint32)
	change := func(y int32, state uint32) {
		if _, ok := previous[y]; !ok {
			previous[y] = ch.GetBlock(x, z, y)
		}
		ch.SetBlock(x, z, y, state)
	}
	rollback := func() {
		for y, state := range previous {
			ch.SetBlock(x, z, y, state)
		}
	}
	change(msg.y, msg.state)
	if msg.upper != nil {
		change(msg.y+1, *msg.upper)
	}
	if err := settleColumn(ch, x, z, msg.y-1, change); err != nil {
		rollback()
		return err
	}
	var ys []int
	for y, state := range previous {
		if ch.GetBlock(x, z, y) != state {
			ys = append(ys, int(y))
		}
	}
	if len(ys) == 0 {
		return nil
	}
	ch.lightStale = true
	if err := r.world.store.vanilla.persistChunk(ch); err != nil {
		rollback()
		return fmt.Errorf("block update save failed: %w", err)
	}
	if r.world.lighting != nil {
		changes := make([]light.Change, 0, len(ys))
		for _, y := range ys {
			changes = append(changes, light.Change{X: x, Y: y, Z: z, State: uint16(ch.GetBlock(x, z, int32(y)))})
		}
		position := light.Position{X: ch.cx, Z: ch.cz}
		updates, err := r.world.lighting.Edit(position, changes)
		if err != nil {
			return fmt.Errorf("block saved but light update failed: %w", err)
		}
		// Block-change packets precede the new lighting on the connection.
		r.applyLight(ch, updates[position], false)
		ch.lights = r.world.lighting.Snapshot(position)
		ch.lightStale = false
		if err := r.world.store.vanilla.persistChunk(ch); err != nil {
			return fmt.Errorf("block saved but lighting save failed: %w", err)
		}
		r.world.notifyLight(updates, position)
	}
	sort.Ints(ys)
	for _, value := range ys {
		y := int32(value)
		changed = append(changed, BlockPosition{X: msg.x, Y: y, Z: msg.z})
		payload := proto.NewWriter().I64(proto.PackPosition(msg.x, y, msg.z)).VarInt(int32(ch.GetBlock(x, z, y))).Payload()
		for _, v := range r.world.hub.ViewersOf(msg.x, msg.z) {
			if !chunkDelivered(v, ch.cx, ch.cz) {
				continue
			}
			v.SendPacket(proto.SBlockChange, payload)
		}
	}
	if ch.lights != nil {
		writer := proto.NewWriter().VarInt(ch.cx).VarInt(ch.cz)
		writeLightData(writer, ch.lights)
		for _, sink := range r.world.hub.ViewersOf(msg.x, msg.z) {
			if !chunkDelivered(sink, ch.cx, ch.cz) {
				continue
			}
			sink.SendPacket(proto.SUpdateLight, writer.Payload())
		}
	}
	return nil
}

func overlaps(a, b blockrules.Box) bool {
	const epsilon = 1e-7
	return a[0] < b[3]-epsilon && a[3] > b[0]+epsilon && a[1] < b[4]-epsilon && a[4] > b[1]+epsilon && a[2] < b[5]-epsilon && a[5] > b[2]+epsilon
}

func playerBox(x, y, z, height float64) blockrules.Box {
	return blockrules.Box{x - .3, y, z - .3, x + .3, y + height, z + .3}
}

func (w *World) collisionBoxes(bounds blockrules.Box) ([]blockrules.Box, error) {
	var points []BlockPosition
	// Shapes of fences and walls can reach above their block cell.
	const contactEpsilon = 1e-7
	minX, minY, minZ := int32(math.Floor(bounds[0]+contactEpsilon)), int32(math.Floor(bounds[1]+contactEpsilon))-1, int32(math.Floor(bounds[2]+contactEpsilon))
	maxX, maxY, maxZ := int32(math.Floor(bounds[3]-contactEpsilon)), int32(math.Floor(bounds[4]-contactEpsilon)), int32(math.Floor(bounds[5]-contactEpsilon))
	minY = max(minY, int32(MinY))
	maxY = min(maxY, int32(MaxY-1))
	if int64(maxX-minX+1)*int64(maxY-minY+1)*int64(maxZ-minZ+1) > 32768 {
		return nil, errors.New("movement crosses too many blocks")
	}
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			for y := minY; y <= maxY; y++ {
				points = append(points, BlockPosition{x, y, z})
			}
		}
	}
	states, err := w.ReadLoadedBlocks(points)
	if err != nil {
		return nil, err
	}
	var boxes []blockrules.Box
	for i, id := range states {
		rules, err := blockrules.Lookup(id)
		if err != nil {
			return nil, err
		}
		p := points[i]
		for _, shape := range rules.Boxes {
			for axis, offset := range []float64{float64(p.X), float64(p.Y), float64(p.Z)} {
				shape[axis] += offset
				shape[axis+3] += offset
			}
			if overlaps(shape, bounds) {
				boxes = append(boxes, shape)
			}
		}
	}
	return boxes, nil
}

func segmentHitsPlayerBox(from, to [3]float64, obstacle blockrules.Box, height float64) bool {
	const epsilon = 1e-7
	low, high := 0.0, 1.0
	minima := [3]float64{obstacle[0] - .3 + epsilon, obstacle[1] - height + epsilon, obstacle[2] - .3 + epsilon}
	maxima := [3]float64{obstacle[3] + .3 - epsilon, obstacle[4] - epsilon, obstacle[5] + .3 - epsilon}
	for axis := range from {
		delta := to[axis] - from[axis]
		if math.Abs(delta) < epsilon {
			if from[axis] <= minima[axis] || from[axis] >= maxima[axis] {
				return false
			}
			continue
		}
		a, b := (minima[axis]-from[axis])/delta, (maxima[axis]-from[axis])/delta
		if a > b {
			a, b = b, a
		}
		low = max(low, a)
		high = min(high, b)
		if low >= high {
			return false
		}
	}
	return high > 0 && low < 1
}

func (w *World) CanMovePlayer(from, to [3]float64, height float64, canStep bool) (bool, error) {
	old, next := playerBox(from[0], from[1], from[2], height), playerBox(to[0], to[1], to[2], height)
	bounds := old
	for axis := 0; axis < 3; axis++ {
		bounds[axis] = min(old[axis], next[axis])
		bounds[axis+3] = max(old[axis+3], next[axis+3])
	}
	boxes, err := w.collisionBoxes(bounds)
	if err != nil {
		return false, err
	}
	pathClear := func(a, b [3]float64) bool {
		for _, box := range boxes {
			if overlaps(playerBox(a[0], a[1], a[2], height), box) {
				// Existing shallow overlap may retreat along its nearest exit.
				// Skipping the box entirely would permit crossing the whole wall.
				minima := [3]float64{box[0] - .3, box[1] - height, box[2] - .3}
				maxima := [3]float64{box[3] + .3, box[4], box[5] + .3}
				nearest := math.Inf(1)
				axis := 0
				sign := 1.0
				for i := 0; i < 3; i++ {
					if distance := a[i] - minima[i]; distance < nearest {
						nearest = distance
						axis = i
						sign = -1
					}
					if distance := maxima[i] - a[i]; distance < nearest {
						nearest = distance
						axis = i
						sign = 1
					}
				}
				if (b[axis]-a[axis])*sign >= -1e-7 {
					continue
				}
				return false
			}
			if segmentHitsPlayerBox(a, b, box, height) {
				return false
			}
		}
		return true
	}
	if pathClear(from, to) {
		return true, nil
	}
	// Vanilla resolves vertical and horizontal components separately. A
	// legitimate diagonal slide can have a blocked straight chord even though
	// the client's axis-resolved path around a wall corner is clear.
	axisPathClear := func(a, b [3]float64) bool {
		vertical := [3]float64{a[0], b[1], a[2]}
		if !pathClear(a, vertical) {
			return false
		}
		for _, firstAxis := range []int{0, 2} {
			corner := vertical
			corner[firstAxis] = b[firstAxis]
			if pathClear(vertical, corner) && pathClear(corner, b) {
				return true
			}
		}
		return false
	}
	if axisPathClear(from, to) {
		return true, nil
	}
	// A legitimate client step moves vertically before crossing the ledge.
	if canStep && to[1] > from[1] && to[1]-from[1] <= .600001 {
		step := [3]float64{from[0], to[1], from[2]}
		return pathClear(from, step) && pathClear(step, to), nil
	}
	return false, nil
}

func BlockIntersectsPlayer(state uint32, p BlockPosition, x, y, z, height float64) (bool, error) {
	rules, err := blockrules.Lookup(state)
	if err != nil {
		return false, err
	}
	player := playerBox(x, y, z, height)
	for _, shape := range rules.Boxes {
		shape[0] += float64(p.X)
		shape[3] += float64(p.X)
		shape[1] += float64(p.Y)
		shape[4] += float64(p.Y)
		shape[2] += float64(p.Z)
		shape[5] += float64(p.Z)
		if overlaps(player, shape) {
			return true, nil
		}
	}
	return false, nil
}
