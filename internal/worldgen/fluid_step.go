package worldgen

import (
	"fmt"
	"strconv"
)

// FluidTickWorld provides the level callbacks used by FlowingFluid.tick.
// SetFluidBlock must honor the native update flags. Neighbor/block callbacks,
// container behavior and queued tick execution belong to the level adapter;
// calling TickFluid alone does not settle a world or execute a tick queue.
type FluidTickWorld interface {
	BlockAt(FeaturePosition) (SurfaceBlockState, error)
	SetFluidBlock(FeaturePosition, SurfaceBlockState, int32) error
	ScheduleFluid(FeaturePosition, FluidType, int32) error
	NotifyFluidNeighbors(FeaturePosition, string) error
	FluidLevelEvent(FeaturePosition, int32, int32) error
	BeforeFluidDestroysBlock(FeaturePosition, SurfaceBlockState, FluidType) error
	PlaceFluidInContainer(FeaturePosition, SurfaceBlockState, FluidState) error
}

type FluidSettings struct {
	UltraWarm             bool
	WaterSourceConversion bool
	LavaSourceConversion  bool
}

type FluidRandom interface{ NextIntBound(int32) int32 }

var fluidDirections = [6]FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}}
var fluidHorizontal = [4]int{2, 5, 3, 4} // Direction.Plane.HORIZONTAL: N,E,S,W.

type fluidStep struct {
	world    FluidTickWorld
	settings FluidSettings
	random   FluidRandom
	kind     FluidType
	err      error
}

func fluidNeighbor(p FeaturePosition, face int) FeaturePosition {
	d := fluidDirections[face]
	return FeaturePosition{p.X + d.X, p.Y + d.Y, p.Z + d.Z}
}

// TickFluid executes one original fluid tick, including its ordered level
// callbacks. The caller must check a scheduled request's type before calling.
func TickFluid(world FluidTickWorld, p FeaturePosition, settings FluidSettings, random FluidRandom) error {
	if world == nil {
		return fmt.Errorf("fluid tick requires a level")
	}
	e := &fluidStep{world: world, settings: settings, random: random}
	state := e.block(p)
	fluid := e.rules(state).state
	if e.err != nil || fluid.Type == FluidEmpty {
		return e.err
	}
	e.kind = fluid.Type
	if !fluid.Source() {
		next := e.newLiquid(p, state)
		delay := e.spreadDelay(p, fluid, next)
		if next.Type == FluidEmpty {
			fluid = next
			e.set(p, structureState("air"), 3)
		} else if next != fluid {
			fluid = next
			block := fluid.LegacyBlock()
			e.set(p, block, 2)
			e.schedule(p, fluid.Type, delay)
			if e.err == nil {
				e.err = world.NotifyFluidNeighbors(p, block.Name)
			}
		}
	}
	e.spread(p, fluid)
	return e.err
}

func (s FluidState) LegacyBlock() SurfaceBlockState {
	if s.Type == FluidEmpty {
		return structureState("air")
	}
	name := "water"
	if s.Type.same(FluidLava) {
		name = "lava"
	}
	level := int32(0)
	if !s.Source() {
		level = 8 - min(s.Amount, 8)
		if s.Falling {
			level += 8
		}
	}
	return SurfaceBlockState{Name: "minecraft:" + name, Properties: map[string]string{"level": strconv.Itoa(int(level))}}
}

func (e *fluidStep) block(p FeaturePosition) SurfaceBlockState {
	if e.err != nil {
		return structureState("air")
	}
	s, err := e.world.BlockAt(p)
	e.err = err
	return s
}
func (e *fluidStep) rules(s SurfaceBlockState) fluidBlockRules {
	if e.err != nil {
		return fluidBlockRules{}
	}
	r, err := fluidRulesForBlock(s)
	e.err = err
	return r
}
func (e *fluidStep) set(p FeaturePosition, s SurfaceBlockState, flags int32) {
	if e.err == nil {
		e.err = e.world.SetFluidBlock(p, s, flags)
	}
}
func (e *fluidStep) schedule(p FeaturePosition, kind FluidType, delay int32) {
	if e.err == nil {
		e.err = e.world.ScheduleFluid(p, kind, delay)
	}
}
func (e *fluidStep) wall(p FeaturePosition, s SurfaceBlockState, q FeaturePosition, t SurfaceBlockState, face int) bool {
	if e.err != nil {
		return false
	}
	open, err := fluidCanPassThroughWall(s, p, t, q, face)
	e.err = err
	return open
}
func (e *fluidStep) flowing() FluidType {
	if e.kind.same(FluidWater) {
		return FluidFlowingWater
	}
	return FluidFlowingLava
}
func (e *fluidStep) source() FluidType {
	if e.kind.same(FluidWater) {
		return FluidWater
	}
	return FluidLava
}
func (e *fluidStep) ownSource(s FluidState) bool { return s.Source() && s.Type.same(e.kind) }
func (e *fluidStep) dropOff() int32 {
	if e.kind.same(FluidLava) && !e.settings.UltraWarm {
		return 2
	}
	return 1
}
func (e *fluidStep) slopeLimit() int {
	if e.kind.same(FluidLava) && !e.settings.UltraWarm {
		return 2
	}
	return 4
}
func (e *fluidStep) delay() int32 {
	if e.kind.same(FluidWater) {
		return 5
	}
	if e.settings.UltraWarm {
		return 10
	}
	return 30
}

func (e *fluidStep) newLiquid(p FeaturePosition, state SurfaceBlockState) FluidState {
	maximum, sources := int32(0), 0
	for _, face := range fluidHorizontal {
		q := fluidNeighbor(p, face)
		other := e.block(q)
		fluid := e.rules(other).state
		if !fluid.Type.same(e.kind) || !e.wall(p, state, q, other, face) {
			continue
		}
		if fluid.Source() {
			sources++
		}
		maximum = max(maximum, fluid.Amount)
	}
	convert := e.settings.WaterSourceConversion
	if e.kind.same(FluidLava) {
		convert = e.settings.LavaSourceConversion
	}
	if convert && sources >= 2 {
		below := e.rules(e.block(fluidNeighbor(p, 0)))
		if below.solid() || e.ownSource(below.state) {
			return FluidState{Type: e.source(), Amount: 8}
		}
	}
	abovePos := fluidNeighbor(p, 1)
	above := e.block(abovePos)
	aboveFluid := e.rules(above).state
	if aboveFluid.Type != FluidEmpty && aboveFluid.Type.same(e.kind) && e.wall(p, state, abovePos, above, 1) {
		return FluidState{Type: e.flowing(), Amount: 8, Falling: true}
	}
	amount := maximum - e.dropOff()
	if amount <= 0 {
		return FluidState{}
	}
	return FluidState{Type: e.flowing(), Amount: amount}
}

func (e *fluidStep) height(p FeaturePosition, state FluidState) float32 {
	above := e.rules(e.block(fluidNeighbor(p, 1))).state
	if state.Type.same(above.Type) {
		return 1
	}
	return float32(state.Amount) / 9
}

func (e *fluidStep) spreadDelay(p FeaturePosition, old, next FluidState) int32 {
	delay := e.delay()
	if e.err != nil {
		return delay
	}
	if e.kind.same(FluidLava) && old.Type != FluidEmpty && next.Type != FluidEmpty && !old.Falling && !next.Falling && e.height(p, next) > e.height(p, old) {
		if e.err != nil {
			return delay
		}
		if e.random == nil {
			e.err = fmt.Errorf("lava spread delay requires the level random source")
			return delay
		}
		if e.random.NextIntBound(4) != 0 {
			delay *= 4
		}
	}
	return delay
}

func (e *fluidStep) replaceable(p FeaturePosition, fluid FluidState, incoming FluidType, face int) bool {
	switch {
	case fluid.Type == FluidEmpty:
		return true
	case fluid.Type.same(FluidWater):
		return face == 0 && !incoming.same(FluidWater)
	case fluid.Type.same(FluidLava):
		return e.height(p, fluid) >= float32(0.44444445) && incoming.same(FluidWater)
	}
	return false
}

func (e *fluidStep) canSpread(p FeaturePosition, state SurfaceBlockState, q FeaturePosition, target SurfaceBlockState, face int, incoming FluidType) bool {
	rules := e.rules(target)
	return e.replaceable(q, rules.state, incoming, face) && e.wall(p, state, q, target, face) && rules.canHold(incoming)
}

func (e *fluidStep) waterHole(p FeaturePosition, state SurfaceBlockState, q FeaturePosition, below SurfaceBlockState, incoming FluidType) bool {
	if !e.wall(p, state, q, below, 0) {
		return false
	}
	rules := e.rules(below)
	return rules.state.Type.same(e.kind) || rules.canHold(incoming)
}

func (e *fluidStep) canPass(p FeaturePosition, state SurfaceBlockState, q FeaturePosition, target SurfaceBlockState, face int, incoming FluidType) bool {
	rules := e.rules(target)
	return !e.ownSource(rules.state) && e.wall(p, state, q, target, face) && rules.canHold(incoming)
}

func (e *fluidStep) spread(p FeaturePosition, fluid FluidState) {
	if e.err != nil || fluid.Type == FluidEmpty {
		return
	}
	state := e.block(p)
	q := fluidNeighbor(p, 0)
	below := e.block(q)
	next := e.newLiquid(q, below)
	if e.canSpread(p, state, q, below, 0, next.Type) {
		e.spreadTo(q, below, 0, next)
		sources := 0
		for _, face := range fluidHorizontal {
			if e.ownSource(e.rules(e.block(fluidNeighbor(p, face))).state) {
				sources++
			}
		}
		if sources >= 3 {
			e.spreadSides(p, fluid, state)
		}
	} else if fluid.Source() || !e.waterHole(p, state, q, below, next.Type) {
		e.spreadSides(p, fluid, state)
	}
}

func (e *fluidStep) spreadSides(p FeaturePosition, fluid FluidState, state SurfaceBlockState) {
	amount := fluid.Amount - e.dropOff()
	if fluid.Falling {
		amount = 7
	}
	if amount <= 0 || e.err != nil {
		return
	}
	best := 1000
	var choices [6]*FluidState
	// The original cache is scoped to a single getSpread call. Its reads are
	// immutable until this search finishes; writes follow EnumMap order.
	search := fluidSlopeSearch{e: e, states: make(map[FeaturePosition]SurfaceBlockState), holes: make(map[FeaturePosition]bool)}
	for _, face := range fluidHorizontal {
		q := fluidNeighbor(p, face)
		target := search.block(q)
		next := e.newLiquid(q, target)
		if !e.canPass(p, state, q, target, face, next.Type) {
			continue
		}
		distance := 0
		if !search.hole(q, target) {
			distance = search.distance(q, target, 1, face^1)
		}
		if distance < best {
			choices = [6]*FluidState{}
		}
		if distance <= best {
			copy := next
			choices[face] = &copy
			best = distance
		}
	}
	for face := 2; face < 6; face++ {
		if next := choices[face]; next != nil {
			q := fluidNeighbor(p, face)
			target := e.block(q)
			if e.canSpread(p, state, q, target, face, next.Type) {
				e.spreadTo(q, target, face, *next)
			}
		}
	}
}

type fluidSlopeSearch struct {
	e      *fluidStep
	states map[FeaturePosition]SurfaceBlockState
	holes  map[FeaturePosition]bool
}

func (s *fluidSlopeSearch) block(p FeaturePosition) SurfaceBlockState {
	if state, ok := s.states[p]; ok {
		return state
	}
	state := s.e.block(p)
	s.states[p] = state
	return state
}
func (s *fluidSlopeSearch) hole(p FeaturePosition, state SurfaceBlockState) bool {
	if hole, ok := s.holes[p]; ok {
		return hole
	}
	q := fluidNeighbor(p, 0)
	below := s.e.block(q)
	hole := s.e.waterHole(p, state, q, below, s.e.flowing())
	s.holes[p] = hole
	return hole
}
func (s *fluidSlopeSearch) distance(p FeaturePosition, state SurfaceBlockState, depth, from int) int {
	best := 1000
	for _, face := range fluidHorizontal {
		if face == from {
			continue
		}
		q := fluidNeighbor(p, face)
		target := s.block(q)
		if !s.e.canPass(p, state, q, target, face, s.e.flowing()) {
			continue
		}
		if s.hole(q, target) {
			return depth
		}
		if depth < s.e.slopeLimit() {
			best = min(best, s.distance(q, target, depth+1, face^1))
		}
	}
	return best
}

func (e *fluidStep) spreadTo(p FeaturePosition, target SurfaceBlockState, face int, fluid FluidState) {
	if e.err != nil {
		return
	}
	rules := e.rules(target)
	if e.err != nil {
		return
	}
	if e.kind.same(FluidLava) && face == 0 && rules.state.Type.same(FluidWater) {
		if target.Name == "minecraft:water" || target.Name == "minecraft:lava" {
			e.set(p, structureState("stone"), 3)
		}
		if e.err == nil {
			e.err = e.world.FluidLevelEvent(p, 1501, 0)
		}
		return
	}
	if rules.container() {
		e.err = e.world.PlaceFluidInContainer(p, target, fluid)
		return
	}
	if !surfaceAir(target) {
		e.err = e.world.BeforeFluidDestroysBlock(p, target, e.kind)
	}
	e.set(p, fluid.LegacyBlock(), 3)
}
