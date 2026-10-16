package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
)

//go:embed sculk_tags_1.21.1.json
var sculkTagData []byte

var sculkTags = func() map[string]map[string]bool {
	var values map[string][]string
	if err := json.Unmarshal(sculkTagData, &values); err != nil {
		panic(err)
	}
	result := make(map[string]map[string]bool)
	for tag, blocks := range values {
		result[tag] = make(map[string]bool)
		for _, block := range blocks {
			result[tag][block] = true
		}
	}
	return result
}()

type sculkPatchConfiguration struct {
	Charges      int32           `json:"charge_count"`
	Amount       int32           `json:"amount_per_charge"`
	Attempts     int32           `json:"spread_attempts"`
	GrowthRounds int32           `json:"growth_rounds"`
	SpreadRounds int32           `json:"spread_rounds"`
	RareGrowths  json.RawMessage `json:"extra_rare_growths"`
	Catalyst     float32         `json:"catalyst_chance"`
	rareGrowths  featureIntProvider
}

func compileSculkPatchPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c sculkPatchConfiguration
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Charges < 1 || c.Charges > 32 || c.Amount < 1 || c.Amount > 500 || c.Attempts < 1 || c.Attempts > 64 || c.GrowthRounds < 0 || c.GrowthRounds > 8 || c.SpreadRounds < 0 || c.SpreadRounds > 8 || c.Catalyst < 0 || c.Catalyst > 1 {
		return nil, fmt.Errorf("invalid sculk patch configuration")
	}
	var err error
	c.rareGrowths, err = decodeFeatureIntProvider(c.RareGrowths, 0, 2147483647)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) (bool, error) {
		return placeSculkPatch(region, r, p, c)
	}, nil
}

// SculkPatchFeature and the worldgen SculkSpreader use charge cursors rather
// than a radius-shaped replacement. Cursor order, delayed updates and each
// random shuffle are observable parts of the configured feature's RNG stream.
func placeSculkPatch(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition, c sculkPatchConfiguration) (bool, error) {
	e := featureEditor{world: region}
	s := e.block(origin)
	valid := sculkBehaviour(s) != 0
	if !valid && (surfaceAir(s) || hasSourceWater(s) && s.Name == "minecraft:water") {
		for d := range featureDirections {
			if sculkFullCollision(&e, e.block(directionPosition(origin, d))) {
				valid = true
				break
			}
		}
	}
	if !valid || e.err != nil {
		return false, e.err
	}
	for round := int32(0); round < c.SpreadRounds+c.GrowthRounds; round++ {
		cursors := make([]sculkChargeCursor, c.Charges)
		for i := range cursors {
			cursors[i] = sculkChargeCursor{p: origin, charge: c.Amount, decayDelay: 1, facings: -1}
		}
		for attempt := int32(0); attempt < c.Attempts; attempt++ {
			live := cursors[:0]
			for _, cursor := range cursors {
				cursor.update(&e, region, r, origin, round < c.SpreadRounds)
				if cursor.charge > 0 {
					// Worldgen does not merge collocated cursors.
					live = append(live, cursor)
				}
			}
			cursors = live
		}
	}
	if r.NextFloat() <= c.Catalyst && sculkFullCollision(&e, e.block(directionPosition(origin, 0))) {
		sculkSetGrowth(&e, region, origin, SurfaceBlockState{Name: "minecraft:sculk_catalyst"})
	}
	count := c.rareGrowths(r)
	for i := int32(0); i < count; i++ {
		p := treeOffset(origin, r.NextIntBound(5)-2, 0, r.NextIntBound(5)-2)
		if surfaceAir(e.block(p)) && multifaceRules(&e, e.block(directionPosition(p, 0))).FaceSturdy[1] {
			sculkSetGrowth(&e, region, p, SurfaceBlockState{Name: "minecraft:sculk_shrieker", Properties: map[string]string{"can_summon": "true"}})
		}
	}
	return e.err == nil, e.err
}

func sculkFullCollision(e *featureEditor, s SurfaceBlockState) bool {
	boxes := multifaceRules(e, s).Boxes
	// BlockRulesProbe exports VoxelShape.toAabbs. Vanilla's full cube shape
	// contains one unit box, including translucent blocks such as leaves/glass.
	return len(boxes) == 1 && boxes[0] == [6]float64{0, 0, 0, 1, 1, 1}
}

// 0 = default behaviour, 1 = vein, 2 = sculk block.
func sculkBehaviour(s SurfaceBlockState) int {
	switch s.Name {
	case "minecraft:sculk_vein":
		return 1
	case "minecraft:sculk":
		return 2
	default:
		return 0
	}
}

func sculkFaceMask(s SurfaceBlockState) int {
	mask := 0
	for d, name := range directionNames {
		if s.Properties[name] == "true" {
			mask |= 1 << uint(d)
		}
	}
	return mask
}

// MultifaceSpreader.spreadAll visits both face streams in Direction.values
// order. It uses the captured source state even after an earlier face is set.
func sculkSpreadAll(e *featureEditor, region *OverworldFeatureRegion, p FeaturePosition, source SurfaceBlockState, samePositionOnly bool) bool {
	changed := false
	mask := sculkFaceMask(source)
	fromAnyFace := source.Name != "minecraft:sculk_vein"
	for face := 0; face < 6; face++ {
		if !fromAnyFace && mask&(1<<uint(face)) == 0 {
			continue
		}
		for dir := 0; dir < 6; dir++ {
			if dir/2 == face/2 || !fromAnyFace && mask&(1<<uint(dir)) != 0 {
				continue
			}
			candidates := []struct {
				p FeaturePosition
				d int
			}{{p, dir}, {directionPosition(p, dir), face}, {directionPosition(directionPosition(p, dir), face), dir ^ 1}}
			if samePositionOnly {
				candidates = candidates[:1]
			}
			for _, candidate := range candidates {
				if !canSpreadMultiface(e, p, candidate.p, candidate.d, "minecraft:sculk_vein") {
					continue
				}
				state, ok := newMultifaceFace(e, candidate.p, candidate.d, "minecraft:sculk_vein")
				if !ok {
					continue
				}
				if e.err == nil {
					e.err = region.MarkPostProcessing(candidate.p)
				}
				e.set(candidate.p, state)
				changed = true
				break
			}
		}
	}
	return changed
}

func sculkDischarge(e *featureEditor, p FeaturePosition, source SurfaceBlockState) {
	if source.Name != "minecraft:sculk_vein" {
		return
	}
	for d, name := range directionNames {
		if source.Properties[name] == "true" && e.block(directionPosition(p, d)).Name == "minecraft:sculk" {
			source = stateWithProperty(source, name, "false")
		}
	}
	if sculkFaceMask(source) == 0 {
		name := "minecraft:air"
		if source.Properties["waterlogged"] == "true" {
			name = "minecraft:water"
		}
		source = SurfaceBlockState{Name: name}
	}
	e.set(p, source)
}

type sculkChargeCursor struct {
	p           FeaturePosition
	charge      int32
	updateDelay int32
	decayDelay  int32
	facings     int // -1 means no captured faces; 0 is an empty face set.
}

func (c *sculkChargeCursor) spread(e *featureEditor, region *OverworldFeatureRegion, source SurfaceBlockState) bool {
	if sculkBehaviour(source) != 0 {
		return sculkSpreadAll(e, region, c.p, source, false)
	}
	if c.facings < 0 {
		return sculkSpreadAll(e, region, c.p, source, true)
	}
	if c.facings == 0 {
		return sculkSpreadAll(e, region, c.p, source, false)
	}
	if !surfaceAir(source) && !hasSourceWater(source) {
		return false
	}
	state := SurfaceBlockState{Name: "minecraft:sculk_vein", Properties: map[string]string{"down": "false", "up": "false", "north": "false", "south": "false", "west": "false", "east": "false", "waterlogged": "false"}}
	found := false
	for d, name := range directionNames {
		if c.facings&(1<<uint(d)) != 0 && multifaceAttach(e, c.p, d) {
			state = stateWithProperty(state, name, "true")
			found = true
		}
	}
	if found {
		if hasSourceWater(source) {
			state = stateWithProperty(state, "waterlogged", "true")
		}
		e.set(c.p, state)
	}
	return found
}

func (c *sculkChargeCursor) update(e *featureEditor, region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition, spread bool) {
	if c.updateDelay > 0 {
		c.updateDelay--
		return
	}
	source := e.block(c.p)
	behaviour := sculkBehaviour(source)
	if spread && c.spread(e, region, source) && behaviour != 2 {
		source = e.block(c.p)
		behaviour = sculkBehaviour(source)
	}
	c.useCharge(e, region, r, origin, behaviour, spread)
	if c.charge <= 0 {
		sculkDischarge(e, c.p, source)
		return
	}
	if next, ok := sculkMove(e, r, c.p); ok {
		sculkDischarge(e, c.p, source)
		c.p = next
		dx, dz := float64(next.X-origin.X), float64(next.Z-origin.Z)
		if dx*dx+dz*dz >= 225 {
			c.charge = 0
			return
		}
		source = e.block(next)
	}
	if sculkBehaviour(source) != 0 {
		c.facings = sculkFaceMask(source)
	}
	if behaviour == 0 {
		c.decayDelay = max(c.decayDelay-1, 0)
	} else {
		c.decayDelay = 1
	}
	c.updateDelay = 1
}

func (c *sculkChargeCursor) useCharge(e *featureEditor, region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition, behaviour int, spread bool) {
	switch behaviour {
	case 0:
		if c.decayDelay <= 0 {
			c.charge = 0
		}
	case 1:
		if spread && sculkReplaceAttached(e, region, r, c.p) {
			c.charge--
		} else if r.NextIntBound(5) == 0 {
			c.charge = int32(math.Floor(float64(float32(c.charge) * 0.5)))
		}
	case 2:
		if c.charge == 0 || r.NextIntBound(5) != 0 {
			return
		}
		dx, dy, dz := float64(c.p.X-origin.X), float64(c.p.Y-origin.Y), float64(c.p.Z-origin.Z)
		near := dx*dx+dy*dy+dz*dz < 1
		if near || !sculkCanPlaceGrowth(e, c.p) {
			if r.NextIntBound(10) == 0 {
				penalty := int32(1)
				if !near {
					distance := float32(math.Sqrt(dx*dx + dy*dy + dz*dz))
					delta := distance - 1
					fraction := min(float32(1), delta*delta/float32(23*23))
					penalty = max(1, int32(float32(c.charge)*fraction*0.5))
				}
				c.charge -= penalty
			}
			return
		}
		if r.NextIntBound(50) < c.charge {
			p := directionPosition(c.p, 1)
			s := SurfaceBlockState{Name: "minecraft:sculk_sensor"}
			if r.NextIntBound(11) == 0 {
				s = SurfaceBlockState{Name: "minecraft:sculk_shrieker", Properties: map[string]string{"can_summon": "true"}}
			}
			if featureFluid(e.block(p)) {
				s = stateWithProperty(s, "waterlogged", "true")
			}
			sculkSetGrowth(e, region, p, s)
		}
		c.charge = max(0, c.charge-50)
	}
}

func sculkReplaceAttached(e *featureEditor, region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) bool {
	source := e.block(p)
	for _, d := range shuffledDirections(r, []int{0, 1, 2, 3, 4, 5}) {
		target := directionPosition(p, d)
		if source.Properties[directionNames[d]] != "true" || !sculkTags["sculk_replaceable_world_gen"][e.block(target).Name] {
			continue
		}
		sculk := SurfaceBlockState{Name: "minecraft:sculk"}
		e.set(target, sculk)
		sculkSpreadAll(e, region, target, sculk, false)
		for face := 0; face < 6; face++ {
			if face != d^1 {
				q := directionPosition(target, face)
				sculkDischarge(e, q, e.block(q))
			}
		}
		return true
	}
	return false
}

func sculkCanPlaceGrowth(e *featureEditor, p FeaturePosition) bool {
	above := e.block(directionPosition(p, 1))
	if !surfaceAir(above) && !(above.Name == "minecraft:water" && hasSourceWater(above)) {
		return false
	}
	count := 0
	for z := int32(-4); z <= 4; z++ {
		for y := int32(0); y <= 2; y++ {
			for x := int32(-4); x <= 4; x++ {
				s := e.block(treeOffset(p, x, y, z))
				if s.Name == "minecraft:sculk_sensor" || s.Name == "minecraft:sculk_shrieker" {
					count++
					if count > 2 {
						return false
					}
				}
			}
		}
	}
	return true
}

// BlockPos.betweenClosed emits x first, then y, then z. Sculk's 18 offsets
// exclude corners and the center, but keep the six direct and twelve diagonal
// neighbors. The shuffled walk retains the last non-replaceable candidate.
var sculkMoveOffsets = func() []FeaturePosition {
	var result []FeaturePosition
	for z := int32(-1); z <= 1; z++ {
		for y := int32(-1); y <= 1; y++ {
			for x := int32(-1); x <= 1; x++ {
				if (x == 0 || y == 0 || z == 0) && (x != 0 || y != 0 || z != 0) {
					result = append(result, FeaturePosition{x, y, z})
				}
			}
		}
	}
	return result
}()

func sculkMove(e *featureEditor, r *WorldgenRandom, p FeaturePosition) (FeaturePosition, bool) {
	offsets := append([]FeaturePosition(nil), sculkMoveOffsets...)
	for i := len(offsets); i > 1; i-- {
		j := int(r.NextIntBound(int32(i)))
		offsets[i-1], offsets[j] = offsets[j], offsets[i-1]
	}
	next := p
	for _, offset := range offsets {
		candidate := treeOffset(p, offset.X, offset.Y, offset.Z)
		s := e.block(candidate)
		if sculkBehaviour(s) == 0 {
			continue
		}
		if abs32(offset.X)+abs32(offset.Y)+abs32(offset.Z) != 1 {
			open := false
			for d, v := range featureDirections {
				if v.X != 0 && offset.X == v.X || v.Y != 0 && offset.Y == v.Y || v.Z != 0 && offset.Z == v.Z {
					if !multifaceRules(e, e.block(directionPosition(p, d))).FaceSturdy[d^1] {
						open = true
						break
					}
				}
			}
			if !open {
				continue
			}
		}
		next = candidate
		if s.Name == "minecraft:sculk_vein" {
			for d, name := range directionNames {
				if s.Properties[name] == "true" && sculkTags["sculk_replaceable"][e.block(directionPosition(candidate, d)).Name] {
					return next, true
				}
			}
		}
	}
	return next, next != p
}

func sculkSetGrowth(e *featureEditor, region *OverworldFeatureRegion, p FeaturePosition, s SurfaceBlockState) {
	e.set(p, s)
	if e.err != nil || !region.InWorld(p) {
		return
	}
	data := DefaultSculkBlockEntity(s)
	e.err = region.SetBlockEntity(p, s.Name, data)
}

// DefaultSculkBlockEntity is the saved NBT of the feature's newly created
// block entities. Coordinates are added by OverworldFeatureRegion.
func DefaultSculkBlockEntity(s SurfaceBlockState) nbt.Object {
	data := nbt.Object{"id": {Type: nbt.String, Value: s.Name}}
	switch s.Name {
	case "minecraft:sculk_catalyst":
		data["cursors"] = nbt.Tag{Type: nbt.List, Value: []nbt.Tag{}}
	case "minecraft:sculk_sensor":
		data["last_vibration_frequency"] = nbt.Tag{Type: nbt.Int, Value: int32(0)}
		data["listener"] = sculkVibrationListener()
	case "minecraft:sculk_shrieker":
		data["warning_level"] = nbt.Tag{Type: nbt.Int, Value: int32(0)}
		data["listener"] = sculkVibrationListener()
	}
	return data
}

func sculkVibrationListener() nbt.Tag {
	return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"event_delay": {Type: nbt.Int, Value: int32(0)},
		"selector":    {Type: nbt.Compound, Value: nbt.Object{"tick": {Type: nbt.Long, Value: int64(-1)}}},
	}}
}
