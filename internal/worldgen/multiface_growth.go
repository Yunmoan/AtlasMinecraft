package worldgen

import (
	"atlas/internal/blockstate"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//go:embed multiface_support_1.21.1.txt
var multifaceSupportData string

var multifaceSupport = func() []uint8 {
	values := strings.Fields(multifaceSupportData)
	if len(values) != 26684 {
		panic("incomplete multiface support masks")
	}
	result := make([]uint8, len(values))
	for i, s := range values {
		v, err := strconv.ParseUint(s, 10, 6)
		if err != nil {
			panic(err)
		}
		result[i] = uint8(v)
	}
	return result
}()

var directionNames = [...]string{"down", "up", "north", "south", "west", "east"}

func shuffledDirections(r *WorldgenRandom, dirs []int) []int {
	dirs = append([]int(nil), dirs...)
	for i := len(dirs); i > 1; i-- {
		j := int(r.NextIntBound(int32(i)))
		dirs[i-1], dirs[j] = dirs[j], dirs[i-1]
	}
	return dirs
}

func directionPosition(p FeaturePosition, d int) FeaturePosition {
	v := featureDirections[d]
	return treeOffset(p, v.X, v.Y, v.Z)
}

func multifaceAttach(e *featureEditor, p FeaturePosition, d int) bool {
	s := e.block(directionPosition(p, d))
	id, err := blockstate.BlockStateID(s.Name, s.Properties)
	if err != nil {
		e.err = err
		return false
	}
	return multifaceSupport[id]&(1<<uint(d^1)) != 0
}

func newLichenFace(e *featureEditor, p FeaturePosition, d int) (SurfaceBlockState, bool) {
	s := e.block(p)
	if s.Name == "minecraft:glow_lichen" && s.Properties[directionNames[d]] == "true" || !multifaceAttach(e, p, d) {
		return SurfaceBlockState{}, false
	}
	if s.Name != "minecraft:glow_lichen" {
		waterlogged := "false"
		if hasSourceWater(s) {
			waterlogged = "true"
		}
		s = SurfaceBlockState{Name: "minecraft:glow_lichen", Properties: map[string]string{"down": "false", "up": "false", "north": "false", "south": "false", "west": "false", "east": "false", "waterlogged": waterlogged}}
	}
	return stateWithProperty(s, directionNames[d], "true"), true
}

func spreadLichenFace(e *featureEditor, region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition, face int) {
	for _, dir := range shuffledDirections(r, []int{0, 1, 2, 3, 4, 5}) {
		if dir/2 == face/2 {
			continue
		}
		candidates := []struct {
			p FeaturePosition
			d int
		}{{origin, dir}, {directionPosition(origin, dir), face}, {directionPosition(directionPosition(origin, dir), face), dir ^ 1}}
		for _, candidate := range candidates {
			s := e.block(candidate.p)
			if !(surfaceAir(s) || s.Name == "minecraft:glow_lichen" || s.Name == "minecraft:water" && hasSourceWater(s)) {
				continue
			}
			state, ok := newLichenFace(e, candidate.p, candidate.d)
			if !ok {
				continue
			}
			if err := region.MarkPostProcessing(candidate.p); err != nil {
				e.err = err
				return
			}
			e.set(candidate.p, state)
			return
		}
	}
}

func compileMultifaceGrowthPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Block   string   `json:"block"`
		Search  int32    `json:"search_range"`
		Floor   bool     `json:"can_place_on_floor"`
		Ceiling bool     `json:"can_place_on_ceiling"`
		Wall    bool     `json:"can_place_on_wall"`
		Spread  float32  `json:"chance_of_spreading"`
		Support []string `json:"can_be_placed_on"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Block != "minecraft:glow_lichen" || c.Search < 1 || c.Search > 64 || c.Spread < 0 || c.Spread > 1 {
		return nil, fmt.Errorf("unsupported multiface configuration")
	}
	allowed := map[string]bool{}
	for _, s := range c.Support {
		allowed[s] = true
	}
	var dirs []int
	if c.Ceiling {
		dirs = append(dirs, 1)
	}
	if c.Floor {
		dirs = append(dirs, 0)
	}
	if c.Wall {
		dirs = append(dirs, 2, 5, 3, 4)
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !dripstoneSpace(e.block(p)) {
			return false, e.err
		}
		try := func(q FeaturePosition, directions []int) bool {
			for _, dir := range directions {
				if !allowed[e.block(directionPosition(q, dir)).Name] {
					continue
				}
				state, ok := newLichenFace(&e, q, dir)
				if !ok {
					return false
				}
				e.set(q, state)
				if err := region.MarkPostProcessing(q); err != nil {
					e.err = err
					return false
				}
				if r.NextFloat() < c.Spread {
					spreadLichenFace(&e, region, r, q, dir)
				}
				return true
			}
			return false
		}
		directions := shuffledDirections(r, dirs)
		if try(p, directions) {
			return e.err == nil, e.err
		}
		for _, dir := range directions {
			var others []int
			for _, v := range dirs {
				if v != dir^1 {
					others = append(others, v)
				}
			}
			others = shuffledDirections(r, others)
			// Vanilla resets the mutable position from the origin on every attempt.
			q := directionPosition(p, dir)
			for i := int32(0); i < c.Search; i++ {
				s := e.block(q)
				if !dripstoneSpace(s) && s.Name != "minecraft:glow_lichen" {
					break
				}
				if try(q, others) {
					return e.err == nil, e.err
				}
			}
		}
		return false, e.err
	}, nil
}
