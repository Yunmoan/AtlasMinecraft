package worldgen

import (
	"fmt"
	"math"
)

// A NoiseChunk fills one cell before consuming its block values. Compile the
// part of the graph outside interpolators into array operations; interpolator
// inputs still use the ordinary corner sampler. Each shared node has one array
// per fill, never a coordinate cache shared with aquifer/single-point probes.
type densityCellFiller struct {
	sampler       *DensitySampler
	width, height int32
	instructions  []densityCellInstruction
	root          int
	values        []float64
}

type densityCellInstruction struct {
	node        *DensityNode
	evaluation  *densityEvaluation
	left, right int
}

func newDensityCellFiller(s *DensitySampler, root *DensityNode) *densityCellFiller {
	f := &densityCellFiller{sampler: s, width: s.cellWidth, height: s.cellHeight}
	indices := make(map[*DensityNode]int)
	var compile func(*DensityNode) int
	compile = func(n *DensityNode) int {
		if i, ok := indices[n]; ok {
			return i
		}
		instruction := densityCellInstruction{node: n, left: -1, right: -1}
		if n != nil {
			e := n.evaluation
			if e == nil {
				e = prepareDensityEvaluation(n)
			}
			instruction.evaluation = e
			switch e.op {
			case densityOpCacheAllInCell, densityOpCacheOnce, densityOpBlendDensity:
				i := compile(e.children[densityArgument])
				indices[n] = i
				return i
			case densityOpAdd, densityOpMin, densityOpMax:
				instruction.left, instruction.right = compile(e.children[densityLeft]), compile(e.children[densityRight])
			case densityOpMul:
				// General Ap2 multiplication samples only nonzero positions of
				// its right child. Keep that selective path instead of eagerly
				// filling a second array; constant MulOrAdd remains vectorized.
				if !s.zeroMultiply(e) {
					instruction.left, instruction.right = compile(e.children[densityLeft]), compile(e.children[densityRight])
				}
			case densityOpAbs, densityOpSquare, densityOpCube, densityOpHalfNegative, densityOpQuarterNegative, densityOpSqueeze:
				instruction.left = compile(e.children[densityArgument])
			case densityOpClamp, densityOpRangeChoice:
				instruction.left = compile(e.children[densityInput])
			}
		}
		i := len(f.instructions)
		f.instructions = append(f.instructions, instruction)
		indices[n] = i
		return i
	}
	f.root = compile(root)
	if count := int64(f.width) * int64(f.width) * int64(f.height); f.width > 0 && f.height > 0 && count <= 65536 {
		f.values = make([]float64, int(count)*len(f.instructions))
	}
	return f
}

func (f *densityCellFiller) fill(origin densityPosition) ([]float64, error) {
	if len(f.values) == 0 || f.sampler.chunk == nil {
		return nil, fmt.Errorf("density cell requires bounded NoiseChunk dimensions")
	}
	context := f.sampler.chunk
	previous := context.fillingCell
	context.fillingCell = true
	defer func() { context.fillingCell = previous }()
	count := int(f.width * f.width * f.height)
	for index, instruction := range f.instructions {
		out := f.values[index*count : (index+1)*count]
		if instruction.evaluation == nil {
			return nil, fmt.Errorf("nil density node")
		}
		e := instruction.evaluation
		var left, right []float64
		if instruction.left >= 0 {
			left = f.values[instruction.left*count : (instruction.left+1)*count]
		}
		if instruction.right >= 0 {
			right = f.values[instruction.right*count : (instruction.right+1)*count]
		}
		switch e.op {
		case densityOpConstant:
			for i := range out {
				out[i] = e.numbers[densityValue]
			}
		case densityOpAdd:
			for i := range out {
				out[i] = left[i] + right[i]
			}
		case densityOpMul:
			if left == nil {
				if err := f.sample(instruction, left, origin, out); err != nil {
					return nil, err
				}
			} else {
				for i := range out {
					out[i] = left[i] * right[i]
				}
			}
		case densityOpMin:
			for i := range out {
				out[i] = math.Min(left[i], right[i])
			}
		case densityOpMax:
			for i := range out {
				out[i] = math.Max(left[i], right[i])
			}
		case densityOpAbs:
			for i := range out {
				out[i] = math.Abs(left[i])
			}
		case densityOpSquare:
			for i, v := range left {
				out[i] = v * v
			}
		case densityOpCube:
			for i, v := range left {
				out[i] = v * v * v
			}
		case densityOpHalfNegative, densityOpQuarterNegative:
			factor := 0.5
			if e.op == densityOpQuarterNegative {
				factor = 0.25
			}
			for i, v := range left {
				if v < 0 {
					v *= factor
				}
				out[i] = v
			}
		case densityOpSqueeze:
			for i, v := range left {
				v = math.Max(-1, math.Min(1, v))
				out[i] = v/2 - v*v*v/24
			}
		case densityOpClamp:
			for i, v := range left {
				out[i] = math.Max(e.numbers[densityMin], math.Min(e.numbers[densityMax], v))
			}
		case densityOpInterpolated:
			if !context.singlePoint && origin.x%f.width == 0 && origin.y%f.height == 0 && origin.z%f.width == 0 {
				if err := f.interpolate(instruction.node.argument(), origin, out); err != nil {
					return nil, err
				}
				continue
			}
			fallthrough
		default:
			// RangeChoice follows vanilla's fillArray: fill the input, then
			// compute only the selected child at each position. In particular,
			// an unused invalid branch must not be evaluated eagerly.
			if err := f.sample(instruction, left, origin, out); err != nil {
				return nil, err
			}
		}
	}
	return f.values[f.root*count : (f.root+1)*count], nil
}

func (f *densityCellFiller) sample(instruction densityCellInstruction, input []float64, origin densityPosition, out []float64) error {
	e := instruction.evaluation
	i := 0
	for y := f.height - 1; y >= 0; y-- {
		for x := int32(0); x < f.width; x++ {
			for z := int32(0); z < f.width; z++ {
				n := instruction.node
				if e.op == densityOpRangeChoice {
					n = e.children[densityWhenOut]
					if input[i] >= e.numbers[densityMinInclusive] && input[i] < e.numbers[densityMaxExclusive] {
						n = e.children[densityWhenIn]
					}
				}
				v, err := f.sampler.Sample(n, origin.x+x, origin.y+y, origin.z+z)
				if err != nil {
					return fmt.Errorf("density at %d,%d,%d: %w", origin.x+x, origin.y+y, origin.z+z, err)
				}
				out[i] = v
				i++
			}
		}
	}
	return nil
}

func (f *densityCellFiller) interpolate(node *DensityNode, origin densityPosition, out []float64) error {
	var corners [2][2][2]float64
	for z := 0; z < 2; z++ {
		for x := 0; x < 2; x++ {
			for y := 0; y < 2; y++ {
				v, err := f.sampler.sampleCorner(node, origin.x+int32(x)*f.width, origin.y+int32(y)*f.height, origin.z+int32(z)*f.width)
				if err != nil {
					return err
				}
				corners[z][x][y] = v
			}
		}
	}
	// Filling a cell uses X -> Y -> Z (ordinary block interpolation uses
	// Y -> X -> Z). Reuse the four X results without changing that order.
	for x := int32(0); x < f.width; x++ {
		fx := float64(x) / float64(f.width)
		a := corners[0][0][0] + fx*(corners[0][1][0]-corners[0][0][0])
		b := corners[0][0][1] + fx*(corners[0][1][1]-corners[0][0][1])
		c := corners[1][0][0] + fx*(corners[1][1][0]-corners[1][0][0])
		d := corners[1][0][1] + fx*(corners[1][1][1]-corners[1][0][1])
		for y := f.height - 1; y >= 0; y-- {
			fy := float64(y) / float64(f.height)
			lo, hi := a+fy*(b-a), c+fy*(d-c)
			index := ((f.height-1-y)*f.width + x) * f.width
			for z := int32(0); z < f.width; z++ {
				fz := float64(z) / float64(f.width)
				out[index+z] = lo + fz*(hi-lo)
			}
		}
	}
	return nil
}
