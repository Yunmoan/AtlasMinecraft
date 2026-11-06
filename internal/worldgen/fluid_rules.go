package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"sync"
)

// FluidType distinguishes the five registered 1.21.1 fluid objects. Source and
// flowing types share flow rules but remain distinct scheduled-tick keys.
type FluidType uint8

const (
	FluidEmpty FluidType = iota
	FluidFlowingWater
	FluidWater
	FluidFlowingLava
	FluidLava
)

func (kind FluidType) String() string {
	names := [...]string{"minecraft:empty", "minecraft:flowing_water", "minecraft:water", "minecraft:flowing_lava", "minecraft:lava"}
	if int(kind) >= len(names) {
		return ""
	}
	return names[kind]
}

func (kind FluidType) same(other FluidType) bool {
	return kind == other || (kind == FluidWater || kind == FluidFlowingWater) && (other == FluidWater || other == FluidFlowingWater) ||
		(kind == FluidLava || kind == FluidFlowingLava) && (other == FluidLava || other == FluidFlowingLava)
}

type FluidState struct {
	Type    FluidType
	Amount  int32
	Falling bool
}

func (s FluidState) Source() bool { return s.Type == FluidWater || s.Type == FluidLava }

type fluidBlockRules struct {
	state FluidState
	flags byte
}

func (r fluidBlockRules) canHold(kind FluidType) bool {
	if kind == FluidEmpty {
		return r.flags&128 != 0
	}
	return kind >= FluidFlowingWater && kind <= FluidLava && r.flags&(1<<(kind-1)) != 0
}
func (r fluidBlockRules) solid() bool        { return r.flags&16 != 0 }
func (r fluidBlockRules) blocksMotion() bool { return r.flags&32 != 0 }
func (r fluidBlockRules) container() bool    { return r.flags&64 != 0 }

// These two-byte records come directly from original getFluidState,
// canHoldFluid, isSolid, blocksMotion and LiquidBlockContainer predicates.
//
//go:embed fluid_rules_1.21.1.bin.gz
var fluidRuleCompressed []byte

// The collision AABBs alone lose empty outer cells in vanilla's voxel grid.
// Shapes.mergedFaceOccludes slices grid index 0/size-1, including those cells.
//
//go:embed fluid_collision_slices_1.21.1.json.gz
var fluidCollisionSliceCompressed []byte

var fluidCollisionSlices struct {
	Points  [][6]float64 `json:"slice_points"`
	Indices []uint16     `json:"slice_indices"`
}
var fluidRuleOnce sync.Once
var fluidRuleData []byte
var fluidRuleError error

func fluidRulesForID(id uint32) (fluidBlockRules, error) {
	fluidRuleOnce.Do(func() {
		gz, err := gzip.NewReader(bytes.NewReader(fluidRuleCompressed))
		if err != nil {
			fluidRuleError = err
			return
		}
		defer gz.Close()
		fluidRuleData, fluidRuleError = io.ReadAll(io.LimitReader(gz, 26684*2+1))
		if fluidRuleError == nil && len(fluidRuleData) != 26684*2 {
			fluidRuleError = fmt.Errorf("incomplete 1.21.1 fluid rules")
		}
		if fluidRuleError != nil {
			return
		}
		slices, err := gzip.NewReader(bytes.NewReader(fluidCollisionSliceCompressed))
		if err != nil {
			fluidRuleError = err
			return
		}
		defer slices.Close()
		fluidRuleError = json.NewDecoder(slices).Decode(&fluidCollisionSlices)
		if fluidRuleError != nil {
			return
		}
		if len(fluidCollisionSlices.Indices) != 26684 {
			fluidRuleError = fmt.Errorf("incomplete fluid collision grids")
			return
		}
		for _, i := range fluidCollisionSlices.Indices {
			if int(i) >= len(fluidCollisionSlices.Points) {
				fluidRuleError = fmt.Errorf("invalid fluid collision grid index")
				return
			}
		}
	})
	if fluidRuleError != nil {
		return fluidBlockRules{}, fluidRuleError
	}
	if id >= 26684 {
		return fluidBlockRules{}, fmt.Errorf("unknown fluid block-state ID %d", id)
	}
	bits := fluidRuleData[id*2]
	return fluidBlockRules{FluidState{Type: FluidType(bits & 7), Amount: int32(bits >> 3 & 15), Falling: bits&128 != 0}, fluidRuleData[id*2+1]}, nil
}

func fluidRulesForBlock(state SurfaceBlockState) (fluidBlockRules, error) {
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return fluidBlockRules{}, err
	}
	return fluidRulesForID(id)
}

func FluidStateForBlock(state SurfaceBlockState) (FluidState, error) {
	rules, err := fluidRulesForBlock(state)
	return rules.state, err
}

type fluidFaceRectangle [4]float64

// Fluid uses collision faces, not sturdy-face or full-block classifications.
// A pair of individually incomplete faces can jointly close the passage.
func fluidCanPassThroughWall(first SurfaceBlockState, p FeaturePosition, second SurfaceBlockState, q FeaturePosition, face int) (bool, error) {
	if face < 0 || face >= 6 {
		return false, fmt.Errorf("invalid fluid face %d", face)
	}
	a, err := fluidCollisionFace(first, p, face)
	if err != nil {
		return false, err
	}
	b, err := fluidCollisionFace(second, q, face^1)
	if err != nil {
		return false, err
	}
	return !fluidFacesCoverSquare(append(a, b...)), nil
}

func fluidCollisionFace(state SurfaceBlockState, p FeaturePosition, face int) ([]fluidFaceRectangle, error) {
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return nil, err
	}
	rules, err := blockrules.Lookup(id)
	if err != nil {
		return nil, err
	}
	axis := []int{1, 1, 2, 2, 0, 0}[face]
	other := []int{}
	for i := 0; i < 3; i++ {
		if i != axis {
			other = append(other, i)
		}
	}
	const epsilon = 1e-7
	if _, err := fluidRulesForID(id); err != nil {
		return nil, err
	}
	offset, zero := BlockOffset(state.Name, p), BlockOffset(state.Name, FeaturePosition{})
	plane := fluidCollisionSlices.Points[fluidCollisionSlices.Indices[id]][axis+(face&1)*3] + offset[axis] - zero[axis]
	// Shapes.mergedFaceOccludes only takes a boundary slice when the
	// shape's outermost extent equals the shared block boundary. Merely
	// intersecting that plane (offset dripstone, fences taller than 1) is
	// insufficient and would incorrectly close some partial passages.
	extent := math.Inf(1)
	if face&1 != 0 {
		extent = math.Inf(-1)
	}
	for _, box := range rules.Boxes {
		if face&1 == 0 {
			extent = min(extent, box[axis]+offset[axis]-zero[axis])
		} else {
			extent = max(extent, box[axis+3]+offset[axis]-zero[axis])
		}
	}
	boundary := float64(face & 1)
	if math.Abs(extent-boundary) > epsilon {
		return nil, nil
	}
	var rectangles []fluidFaceRectangle
	for _, box := range rules.Boxes {
		for i := 0; i < 3; i++ {
			box[i] += offset[i] - zero[i]
			box[i+3] += offset[i] - zero[i]
		}
		if box[axis] > plane || box[axis+3] <= plane {
			continue
		}
		r := fluidFaceRectangle{max(0.0, box[other[0]]), max(0.0, box[other[1]]), min(1.0, box[other[0]+3]), min(1.0, box[other[1]+3])}
		if r[2]-r[0] > epsilon && r[3]-r[1] > epsilon {
			rectangles = append(rectangles, r)
		}
	}
	return rectangles, nil
}

func fluidFacesCoverSquare(rectangles []fluidFaceRectangle) bool {
	const epsilon = 1e-7
	if len(rectangles) == 0 {
		return false
	}
	xs, ys := []float64{0, 1}, []float64{0, 1}
	for _, r := range rectangles {
		if r[0] <= epsilon && r[1] <= epsilon && r[2] >= 1-epsilon && r[3] >= 1-epsilon {
			return true
		}
		xs = append(xs, r[0], r[2])
		ys = append(ys, r[1], r[3])
	}
	sort.Float64s(xs)
	sort.Float64s(ys)
	for i := 1; i < len(xs); i++ {
		if xs[i]-xs[i-1] > epsilon {
			for j := 1; j < len(ys); j++ {
				if ys[j]-ys[j-1] > epsilon {
					x, y := (xs[i]+xs[i-1])*.5, (ys[j]+ys[j-1])*.5
					covered := false
					for _, r := range rectangles {
						if x >= r[0] && x <= r[2] && y >= r[1] && y <= r[3] {
							covered = true
							break
						}
					}
					if !covered {
						return false
					}
				}
			}
		}
	}
	return true
}
