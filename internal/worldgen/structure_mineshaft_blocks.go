package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"fmt"
	"strings"
)

type MineshaftBlockWorld interface {
	StructureBlockEntityWorld
	MineshaftBiomeAt(FeaturePosition) (string, error)
	HeightmapAt(string, int32, int32) (int32, error)
	MinBuildHeight() int32
	MaxBuildHeight() int32
}

type mineshaftPieceEditor struct {
	world               MineshaftBlockWorld
	piece               *StructurePiece
	clip                StructureBoundingBox
	planks, logs, fence SurfaceBlockState
	transformed         map[uint32]SurfaceBlockState
	result              StructurePiecePostprocessResult
	err                 error
}

// PostProcessMineshaftBlocks places the original four piece types, saving
// cave-spider spawner flags and returning precise loot-minecart requests.
func PostProcessMineshaftBlocks(world MineshaftBlockWorld, piece *StructurePiece, random StructurePieceRandom, cx, cz int32) (StructurePiecePostprocessResult, error) {
	var result StructurePiecePostprocessResult
	if world == nil || piece == nil || random == nil {
		return result, fmt.Errorf("mineshaft requires world, saved piece and random")
	}
	typeID, ok := piece.Data.Number("MST")
	if !ok || typeID < 0 || typeID > 1 {
		return result, fmt.Errorf("mineshaft piece has invalid type")
	}
	wood := "oak"
	if typeID == 1 {
		wood = "dark_oak"
	}
	e := &mineshaftPieceEditor{world: world, piece: piece, clip: StructureBoundingBox{cx * 16, world.MinBuildHeight(), cz * 16, cx*16 + 15, world.MaxBuildHeight() - 1, cz*16 + 15}, planks: structureState(wood + "_planks"), logs: structureState(wood + "_log"), fence: structureState(wood + "_fence"), transformed: make(map[uint32]SurfaceBlockState)}
	if e.blocked() {
		return result, e.err
	}
	switch piece.ID {
	case "minecraft:msroom":
		e.room()
	case "minecraft:mscrossing":
		e.crossing()
	case "minecraft:msstairs":
		e.stairs()
	case "minecraft:mscorridor":
		e.corridor(random)
	default:
		return result, fmt.Errorf("unsupported mineshaft piece %q", piece.ID)
	}
	return e.result, e.err
}

func (e *mineshaftPieceEditor) read(p FeaturePosition) SurfaceBlockState {
	if e.err != nil {
		return structureState("air")
	}
	s, eErr := e.world.BlockAt(p)
	e.err = eErr
	return s
}
func (e *mineshaftPieceEditor) local(x, y, z int32) SurfaceBlockState {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) {
		return structureState("air")
	}
	return e.read(p)
}
func (e *mineshaftPieceEditor) rules(s SurfaceBlockState) blockrules.Rules {
	if e.err != nil {
		return blockrules.Rules{}
	}
	id, err := blockstate.BlockStateID(s.Name, s.Properties)
	if err != nil {
		e.err = err
		return blockrules.Rules{}
	}
	rules, err := blockrules.Lookup(id)
	e.err = err
	return rules
}
func (e *mineshaftPieceEditor) set(s SurfaceBlockState, x, y, z int32) {
	if e.err != nil {
		return
	}
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) || !e.world.InWorld(p) {
		return
	}
	old := e.read(p)
	if old.Name == e.planks.Name || old.Name == e.logs.Name || old.Name == e.fence.Name || old.Name == "minecraft:chain" {
		return
	}
	id, err := blockstate.BlockStateID(s.Name, s.Properties)
	if err != nil {
		e.err = err
		return
	}
	state, found := e.transformed[id]
	if !found {
		mirror, rotation := TemplateMirrorNone, TemplateRotationNone
		if e.piece.Orientation == "south" || e.piece.Orientation == "west" {
			mirror = TemplateMirrorLeftRight
		}
		if e.piece.Orientation == "west" || e.piece.Orientation == "east" {
			rotation = TemplateRotationClockwise90
		}
		state, e.err = TransformTemplateBlockState(s, mirror, rotation)
		if e.err != nil {
			return
		}
		e.transformed[id] = state
	}
	e.err = e.world.SetBlock(p, state)
	if e.err != nil {
		return
	}
	if featureFluid(state) {
		if schedule, ok := e.world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
			e.err = schedule.ScheduleFluidUpdate(p)
		}
	}
	if state.Name == "minecraft:oak_fence" || state.Name == "minecraft:dark_oak_fence" || state.Name == "minecraft:torch" || state.Name == "minecraft:wall_torch" {
		if marker, ok := e.world.(interface{ MarkPostProcessing(FeaturePosition) error }); ok {
			e.err = marker.MarkPostProcessing(p)
		}
	}
}
func (e *mineshaftPieceEditor) fill(x0, y0, z0, x1, y1, z1 int32, s SurfaceBlockState) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			for z := z0; z <= z1; z++ {
				e.set(s, x, y, z)
			}
		}
	}
}
func (e *mineshaftPieceEditor) underground(x, y, z int32) bool {
	p := e.piece.BlockPosition(x, y+1, z)
	if !structureContains(e.clip, p) || e.err != nil {
		return false
	}
	height, err := e.world.HeightmapAt("OCEAN_FLOOR_WG", p.X, p.Z)
	e.err = err
	return p.Y < height && err == nil
}
func (e *mineshaftPieceEditor) blocked() bool {
	b, c := e.piece.Box, e.clip
	x0, y0, z0 := max(b.MinX-1, c.MinX), max(b.MinY-1, c.MinY), max(b.MinZ-1, c.MinZ)
	x1, y1, z1 := min(b.MaxX+1, c.MaxX), min(b.MaxY+1, c.MaxY), min(b.MaxZ+1, c.MaxZ)
	biome, err := e.world.MineshaftBiomeAt(FeaturePosition{(x0 + x1) / 2, (y0 + y1) / 2, (z0 + z1) / 2})
	e.err = err
	if err != nil || biome == "minecraft:deep_dark" {
		return true
	}
	liquid := func(x, y, z int32) bool {
		s := e.read(FeaturePosition{x, y, z})
		return s.Name == "minecraft:water" || s.Name == "minecraft:lava"
	}
	for x := x0; x <= x1; x++ {
		for z := z0; z <= z1; z++ {
			if liquid(x, y0, z) || liquid(x, y1, z) {
				return true
			}
		}
	}
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			if liquid(x, y, z0) || liquid(x, y, z1) {
				return true
			}
		}
	}
	for z := z0; z <= z1; z++ {
		for y := y0; y <= y1; y++ {
			if liquid(x0, y, z) || liquid(x1, y, z) {
				return true
			}
		}
	}
	return e.err != nil
}
func (e *mineshaftPieceEditor) floor(x, y, z int32) {
	if !e.underground(x, y, z) {
		return
	}
	p := e.piece.BlockPosition(x, y, z)
	if !e.rules(e.read(p)).FaceSturdy[1] && e.err == nil {
		e.err = e.world.SetBlock(p, e.planks)
	}
}
func mineshaftReplaceable(s SurfaceBlockState) bool {
	return surfaceAir(s) || s.Name == "minecraft:water" || s.Name == "minecraft:lava" || s.Name == "minecraft:glow_lichen" || s.Name == "minecraft:seagrass" || s.Name == "minecraft:tall_seagrass"
}
func mineshaftFallingBlock(s SurfaceBlockState) bool {
	return s.Name == "minecraft:sand" || s.Name == "minecraft:red_sand" || s.Name == "minecraft:gravel" || s.Name == "minecraft:dragon_egg" || s.Name == "minecraft:anvil" || s.Name == "minecraft:chipped_anvil" || s.Name == "minecraft:damaged_anvil" || strings.HasSuffix(s.Name, "_concrete_powder")
}
func (e *mineshaftPieceEditor) ceilingAnchor(s SurfaceBlockState) bool {
	if mineshaftFallingBlock(s) || strings.HasSuffix(s.Name, "_fence_gate") || e.err != nil {
		return false
	}
	return e.rules(s).BottomCenter
}
func (e *mineshaftPieceEditor) supportLog(x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) {
		return
	}
	startY := p.Y
	down, up := true, true
	for distance := int32(1); (down || up) && e.err == nil; distance++ {
		if down {
			p.Y = startY - distance
			s := e.read(p)
			replace := mineshaftReplaceable(s) && s.Name != "minecraft:lava"
			if !replace && e.rules(s).FaceSturdy[1] {
				for y := p.Y + 1; y < startY; y++ {
					p.Y = y
					e.err = e.world.SetBlock(p, e.logs)
					if e.err != nil {
						return
					}
				}
				return
			}
			down = distance <= 20 && replace && p.Y > e.world.MinBuildHeight()+1
		}
		if up {
			p.Y = startY + distance
			s := e.read(p)
			replace := mineshaftReplaceable(s)
			if !replace && e.ceilingAnchor(s) {
				p.Y = startY + 1
				e.err = e.world.SetBlock(p, e.fence)
				for y := startY + 2; y < startY+distance && e.err == nil; y++ {
					p.Y = y
					e.err = e.world.SetBlock(p, structureState("chain"))
				}
				return
			}
			up = distance <= 50 && replace && p.Y < e.world.MaxBuildHeight()-1
		}
	}
}
func (e *mineshaftPieceEditor) supportColumns(z int32) {
	if e.local(0, -1, z).Name == e.planks.Name {
		e.supportLog(0, -1, z)
	}
	if e.local(2, -1, z).Name == e.planks.Name {
		e.supportLog(2, -1, z)
	}
}
func (e *mineshaftPieceEditor) crossing() {
	b := e.piece.Box
	air := structureState("cave_air")
	tall, _ := e.piece.Data.Number("tf")
	if tall != 0 {
		e.fill(b.MinX+1, b.MinY, b.MinZ, b.MaxX-1, b.MinY+2, b.MaxZ, air)
		e.fill(b.MinX, b.MinY, b.MinZ+1, b.MaxX, b.MinY+2, b.MaxZ-1, air)
		e.fill(b.MinX+1, b.MaxY-2, b.MinZ, b.MaxX-1, b.MaxY, b.MaxZ, air)
		e.fill(b.MinX, b.MaxY-2, b.MinZ+1, b.MaxX, b.MaxY, b.MaxZ-1, air)
		e.fill(b.MinX+1, b.MinY+3, b.MinZ+1, b.MaxX-1, b.MinY+3, b.MaxZ-1, air)
	} else {
		e.fill(b.MinX+1, b.MinY, b.MinZ, b.MaxX-1, b.MaxY, b.MaxZ, air)
		e.fill(b.MinX, b.MinY, b.MinZ+1, b.MaxX, b.MaxY, b.MaxZ-1, air)
	}
	for _, p := range [][2]int32{{b.MinX + 1, b.MinZ + 1}, {b.MinX + 1, b.MaxZ - 1}, {b.MaxX - 1, b.MinZ + 1}, {b.MaxX - 1, b.MaxZ - 1}} {
		if !surfaceAir(e.local(p[0], b.MaxY+1, p[1])) {
			e.fill(p[0], b.MinY, p[1], p[0], b.MaxY, p[1], e.planks)
		}
	}
	for x := b.MinX; x <= b.MaxX; x++ {
		for z := b.MinZ; z <= b.MaxZ; z++ {
			e.floor(x, b.MinY-1, z)
		}
	}
}
func (e *mineshaftPieceEditor) stairs() {
	air := structureState("cave_air")
	e.fill(0, 5, 0, 2, 7, 1, air)
	e.fill(0, 0, 7, 2, 2, 8, air)
	for i := int32(0); i < 5; i++ {
		extra := int32(0)
		if i < 4 {
			extra = 1
		}
		e.fill(0, 5-i-extra, 2+i, 2, 7-i, 2+i, air)
	}
}
func (e *mineshaftPieceEditor) room() {
	b := e.piece.Box
	air := structureState("cave_air")
	e.fill(b.MinX, b.MinY+1, b.MinZ, b.MaxX, min(b.MinY+3, b.MaxY), b.MaxZ, air)
	entrances, _ := e.piece.Data["Entrances"].Elements()
	for _, entry := range entrances {
		v, ok := entry.Value.([]int32)
		if !ok || len(v) != 6 {
			e.err = fmt.Errorf("invalid mineshaft entrance")
			return
		}
		e.fill(v[0], v[4]-2, v[2], v[3], v[4], v[5], air)
	}
	y0 := b.MinY + 4
	dx, dy, dz := float32(b.MaxX-b.MinX+1), float32(b.MaxY-y0+1), float32(b.MaxZ-b.MinZ+1)
	cx, cz := float32(b.MinX)+dx/2, float32(b.MinZ)+dz/2
	for y := y0; y <= b.MaxY; y++ {
		fy := float32(y-y0) / dy
		for x := b.MinX; x <= b.MaxX; x++ {
			fx := (float32(x) - cx) / (dx * 0.5)
			for z := b.MinZ; z <= b.MaxZ; z++ {
				fz := (float32(z) - cz) / (dz * 0.5)
				sum := float32(float32(fx*fx) + float32(fy*fy))
				sum = float32(sum + float32(fz*fz))
				if sum <= 1.05 {
					e.set(air, x, y, z)
				}
			}
		}
	}
}
func (e *mineshaftPieceEditor) randomFill(random StructurePieceRandom, chance float32, x0, y0, z0, x1, y1, z1 int32, state SurfaceBlockState, underground bool) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			for z := z0; z <= z1; z++ {
				if random.NextFloat() > chance || underground && !e.underground(x, y, z) {
					continue
				}
				e.set(state, x, y, z)
			}
		}
	}
}
func (e *mineshaftPieceEditor) chanceSet(random StructurePieceRandom, chance float32, state SurfaceBlockState, x, y, z int32) {
	if random.NextFloat() < chance {
		e.set(state, x, y, z)
	}
}
func (e *mineshaftPieceEditor) frame(random StructurePieceRandom, z int32) {
	for x := int32(0); x <= 2; x++ {
		if surfaceAir(e.local(x, 3, z)) {
			return
		}
	}
	e.fill(0, 0, z, 0, 1, z, stateWithProperty(e.fence, "west", "true"))
	e.fill(2, 0, z, 2, 1, z, stateWithProperty(e.fence, "east", "true"))
	if random.NextIntBound(4) == 0 {
		e.fill(0, 2, z, 0, 2, z, e.planks)
		e.fill(2, 2, z, 2, 2, z, e.planks)
	} else {
		e.fill(0, 2, z, 2, 2, z, e.planks)
		e.chanceSet(random, 0.05, stateWithProperty(structureState("wall_torch"), "facing", "south"), 1, 2, z-1)
		e.chanceSet(random, 0.05, stateWithProperty(structureState("wall_torch"), "facing", "north"), 1, 2, z+1)
	}
}
func (e *mineshaftPieceEditor) cobweb(random StructurePieceRandom, chance float32, x, y, z int32) {
	if !e.underground(x, y, z) || random.NextFloat() >= chance {
		return
	}
	p := e.piece.BlockPosition(x, y, z)
	supported := 0
	for d := 0; d < 6; d++ {
		q := directionPosition(p, d)
		if structureContains(e.clip, q) && e.rules(e.read(q)).FaceSturdy[d^1] {
			supported++
			if supported >= 2 {
				e.set(structureState("cobweb"), x, y, z)
				return
			}
		}
	}
}
func (e *mineshaftPieceEditor) minecart(random StructurePieceRandom, x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	if !structureContains(e.clip, p) || !surfaceAir(e.read(p)) || surfaceAir(e.read(treeOffset(p, 0, -1, 0))) || e.err != nil {
		return
	}
	shape := "east_west"
	if random.NextBoolean() {
		shape = "north_south"
	}
	e.set(stateWithProperty(structureState("rail"), "shape", shape), x, y, z)
	pos := []nbt.Tag{{Type: nbt.Double, Value: float64(p.X) + 0.5}, {Type: nbt.Double, Value: float64(p.Y) + 0.5}, {Type: nbt.Double, Value: float64(p.Z) + 0.5}}
	data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:chest_minecart"}, "Pos": {Type: nbt.List, ListType: nbt.Double, Value: pos}, "LootTable": {Type: nbt.String, Value: "minecraft:chests/abandoned_mineshaft"}, "LootTableSeed": {Type: nbt.Long, Value: int64(random.NextLong())}}
	e.result.PendingEntities = append(e.result.PendingEntities, PendingStructureEntity{ID: "minecraft:chest_minecart", Position: p, Data: data})
}
func (e *mineshaftPieceEditor) spawner(x, y, z int32) {
	p := e.piece.BlockPosition(x, y, z)
	e.piece.Data["hps"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	e.err = e.world.SetBlock(p, structureState("spawner"))
	if e.err != nil {
		return
	}
	data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:mob_spawner"}, "SpawnData": {Type: nbt.Compound, Value: nbt.Object{"entity": {Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:cave_spider"}}}}}, "SpawnPotentials": {Type: nbt.List, ListType: nbt.End, Value: []nbt.Tag{}}}
	for name, value := range map[string]int16{"Delay": 20, "MinSpawnDelay": 200, "MaxSpawnDelay": 800, "SpawnCount": 4, "MaxNearbyEntities": 6, "RequiredPlayerRange": 16, "SpawnRange": 4} {
		data[name] = nbt.Tag{Type: nbt.Short, Value: value}
	}
	e.err = e.world.SetBlockEntity(p, "minecraft:spawner", data)
}
func (e *mineshaftPieceEditor) corridor(random StructurePieceRandom) {
	n, ok := e.piece.Data.Number("Num")
	if !ok || n < 1 || n > 4 {
		e.err = fmt.Errorf("invalid mineshaft corridor length")
		return
	}
	length := int32(n)*5 - 1
	air := structureState("cave_air")
	rails, _ := e.piece.Data.Number("hr")
	spiders, _ := e.piece.Data.Number("sc")
	e.fill(0, 0, 0, 2, 1, length, air)
	e.randomFill(random, 0.8, 0, 2, 0, 2, 2, length, air, false)
	if spiders != 0 {
		e.randomFill(random, 0.6, 0, 0, 0, 2, 1, length, structureState("cobweb"), true)
	}
	for i := int32(0); i < int32(n); i++ {
		z := 2 + i*5
		e.frame(random, z)
		for _, v := range []struct {
			x, dz  int32
			chance float32
		}{{0, -1, 0.1}, {2, -1, 0.1}, {0, 1, 0.1}, {2, 1, 0.1}, {0, -2, 0.05}, {2, -2, 0.05}, {0, 2, 0.05}, {2, 2, 0.05}} {
			e.cobweb(random, v.chance, v.x, 2, z+v.dz)
		}
		if random.NextIntBound(100) == 0 {
			e.minecart(random, 2, 0, z-1)
		}
		if random.NextIntBound(100) == 0 {
			e.minecart(random, 0, 0, z+1)
		}
		spawned, _ := e.piece.Data.Number("hps")
		if spiders != 0 && spawned == 0 {
			candidate := z - 1 + random.NextIntBound(3)
			p := e.piece.BlockPosition(1, 0, candidate)
			if structureContains(e.clip, p) && e.underground(1, 0, candidate) {
				e.spawner(1, 0, candidate)
			}
		}
	}
	for x := int32(0); x <= 2; x++ {
		for z := int32(0); z <= length; z++ {
			e.floor(x, -1, z)
		}
	}
	e.supportColumns(2)
	if n > 1 {
		e.supportColumns(length - 2)
	}
	if rails != 0 {
		state := stateWithProperty(structureState("rail"), "shape", "north_south")
		for z := int32(0); z <= length; z++ {
			below := e.local(1, -1, z)
			if surfaceAir(below) || !e.rules(below).SolidRender {
				continue
			}
			chance := float32(0.9)
			if e.underground(1, 0, z) {
				chance = 0.7
			}
			e.chanceSet(random, chance, state, 1, 0, z)
		}
	}
}
