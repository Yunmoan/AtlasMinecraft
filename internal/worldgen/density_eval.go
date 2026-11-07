package worldgen

import (
	"fmt"
	"math"
)

// DensitySampler evaluates density expressions for a single block position.
// NoiseChunk sampling adds interpolation and caches; actual structure starts
// must be supplied separately for nonzero terrain adjustments.
// One sampler must not be used concurrently by multiple goroutines.
type DensitySampler struct {
	Seed                  int64
	noises                map[string]*NormalNoise
	blended               map[*DensityNode]*BlendedNoise
	endIslands            *EndIslandDensity
	legacy                bool
	beardifier            *StructureBeardifier
	cellWidth, cellHeight int32
	chunk                 *chunkDensityContext
	// Only the built-in biome router opts in. Its flat_cache/cache_2d
	// arguments are independent of Y; the general direct sampler deliberately
	// retains its original semantics for arbitrary probe expressions.
	horizontal map[densityCacheKey]float64
	program    *densityProgram
	runtime    []densityRuntime
}

func NewDensitySampler(seed int64) *DensitySampler {
	return &DensitySampler{Seed: seed, noises: make(map[string]*NormalNoise), blended: make(map[*DensityNode]*BlendedNoise)}
}

// NewCellDensitySampler evaluates interpolated density nodes in vanilla's
// noise-cell grid. It is one stage of NoiseChunk sampling, not a chunk
// generator: aquifers, surface rules, structures and features still apply.
func NewCellDensitySampler(seed int64, settings *DensitySettings) *DensitySampler {
	s := NewDensitySampler(seed)
	s.cellWidth, s.cellHeight = settings.CellWidth, settings.CellHeight
	s.legacy = settings.Legacy
	s.bind(settings)
	return s
}

func (s *DensitySampler) noise(name string) (*NormalNoise, error) {
	if n := s.noises[name]; n != nil {
		return n, nil
	}
	var n *NormalNoise
	var err error
	if s.legacy {
		n, err = newLegacyNamedNormalNoise(s.Seed, name)
	} else {
		n, err = NewNamedNormalNoise(s.Seed, name)
	}
	if err != nil {
		return nil, err
	}
	s.noises[name] = n
	return n, nil
}

func (s *DensitySampler) nodeNoise(node *DensityNode) (*NormalNoise, error) {
	runtime := s.nodeRuntime(node)
	if runtime != nil && runtime.noise != nil {
		return runtime.noise, nil
	}
	noise, err := s.noise(node.Noise)
	if err == nil && runtime != nil {
		runtime.noise = noise
	}
	return noise, err
}

func (s *DensitySampler) Sample(n *DensityNode, x, y, z int32) (float64, error) {
	if n == nil {
		return 0, fmt.Errorf("nil density node")
	}
	evaluation := n.evaluation
	if evaluation == nil {
		// Manually constructed probe expressions retain their map semantics.
		// Never mutate the node here: parsed trees are shared between workers.
		evaluation = prepareDensityEvaluation(n)
	}
	child := func(index int) (float64, error) {
		return s.Sample(evaluation.children[index], x, y, z)
	}
	switch evaluation.op {
	case densityOpConstant:
		return evaluation.numbers[densityValue], nil
	case densityOpBlendAlpha:
		return 1, nil // No old/new chunk blending in a direct sample.
	case densityOpBlendOffset:
		return 0, nil
	case densityOpBeardifier:
		if s.beardifier == nil {
			return 0, nil
		}
		return s.beardifier.Sample(x, y, z), nil
	case densityOpEndIslands:
		if s.endIslands == nil {
			s.endIslands = NewEndIslandDensity(s.Seed)
		}
		return s.endIslands.Sample(x, z), nil
	case densityOpFlatCache:
		if s.chunk != nil {
			return s.sampleFlatCache(n, x, y, z)
		}
		if s.horizontal != nil {
			return s.sampleHorizontal(n, x, y, z)
		}
		return child(densityArgument)
	case densityOpCache2D:
		if s.chunk != nil {
			return s.sampleCache2D(n, x, y, z)
		}
		if s.horizontal != nil {
			return s.sampleHorizontal(n, x, y, z)
		}
		return child(densityArgument)
	case densityOpCacheAllInCell:
		if s.chunk == nil {
			return 0, fmt.Errorf("cell cache requires a NoiseChunk")
		}
		previous := s.chunk.fillingCell
		s.chunk.fillingCell = true
		value, err := child(densityArgument)
		s.chunk.fillingCell = previous
		return value, err
	case densityOpCacheOnce, densityOpBlendDensity:
		return child(densityArgument)
	case densityOpInterpolated:
		if s.chunk != nil && s.chunk.singlePoint {
			return child(densityArgument)
		}
		if s.cellWidth <= 0 || s.cellHeight <= 0 {
			return 0, fmt.Errorf("interpolated density requires a noise cell")
		}
		return s.interpolate(n.argument(), x, y, z)
	case densityOpAdd, densityOpMul, densityOpMin, densityOpMax:
		a, err := child(densityLeft)
		if err != nil {
			return 0, err
		}
		if a == 0 && s.zeroMultiply(evaluation) {
			return 0, nil
		}
		b, err := child(densityRight)
		if err != nil {
			return 0, err
		}
		switch evaluation.op {
		case densityOpAdd:
			return a + b, nil
		case densityOpMul:
			return a * b, nil
		case densityOpMin:
			return math.Min(a, b), nil
		default:
			return math.Max(a, b), nil
		}
	case densityOpAbs, densityOpSquare, densityOpCube, densityOpHalfNegative, densityOpQuarterNegative, densityOpSqueeze:
		v, err := child(densityArgument)
		if err != nil {
			return 0, err
		}
		switch evaluation.op {
		case densityOpAbs:
			return math.Abs(v), nil
		case densityOpSquare:
			return v * v, nil
		case densityOpCube:
			return v * v * v, nil
		case densityOpHalfNegative:
			if v < 0 {
				return v * 0.5, nil
			}
		case densityOpQuarterNegative:
			if v < 0 {
				return v * 0.25, nil
			}
		case densityOpSqueeze:
			v = math.Max(-1, math.Min(1, v))
			return v/2 - v*v*v/24, nil
		}
		return v, nil
	case densityOpClamp:
		v, err := child(densityInput)
		if err != nil {
			return 0, err
		}
		return math.Max(evaluation.numbers[densityMin], math.Min(evaluation.numbers[densityMax], v)), nil
	case densityOpRangeChoice:
		v, err := child(densityInput)
		if err != nil {
			return 0, err
		}
		if v >= evaluation.numbers[densityMinInclusive] && v < evaluation.numbers[densityMaxExclusive] {
			return child(densityWhenIn)
		}
		return child(densityWhenOut)
	case densityOpYClampedGradient:
		from, to := evaluation.numbers[densityFromY], evaluation.numbers[densityToY]
		if from == to {
			return 0, fmt.Errorf("zero-width y gradient")
		}
		t := math.Max(0, math.Min(1, (float64(y)-from)/(to-from)))
		return evaluation.numbers[densityFromValue] + t*(evaluation.numbers[densityToValue]-evaluation.numbers[densityFromValue]), nil
	case densityOpNoise:
		noise, err := s.nodeNoise(n)
		if err != nil {
			return 0, err
		}
		xz, ys := evaluation.numbers[densityXZScale], evaluation.numbers[densityYScale]
		return noise.Noise(float64(x)*xz, float64(y)*ys, float64(z)*xz), nil
	case densityOpShift, densityOpShiftA, densityOpShiftB:
		noise, err := s.nodeNoise(n)
		if err != nil {
			return 0, err
		}
		fx, fy, fz := float64(x), float64(y), float64(z)
		switch evaluation.op {
		case densityOpShiftA:
			fy = 0
		case densityOpShiftB:
			fx, fy, fz = fz, fx, 0
		}
		return 4 * noise.Noise(fx*0.25, fy*0.25, fz*0.25), nil
	case densityOpShiftedNoise:
		shiftX, err := child(densityShiftX)
		if err != nil {
			return 0, err
		}
		shiftY, err := child(densityShiftY)
		if err != nil {
			return 0, err
		}
		shiftZ, err := child(densityShiftZ)
		if err != nil {
			return 0, err
		}
		noise, err := s.nodeNoise(n)
		if err != nil {
			return 0, err
		}
		xz, ys := evaluation.numbers[densityXZScale], evaluation.numbers[densityYScale]
		return noise.Noise(float64(x)*xz+shiftX, float64(y)*ys+shiftY, float64(z)*xz+shiftZ), nil
	case densityOpSpline:
		v, err := s.sampleSpline(n.Spline, x, y, z)
		return float64(v), err
	case densityOpOldBlendedNoise:
		b := s.blended[n]
		if b == nil {
			if s.legacy {
				b = newLegacyBlendedNoise(s.Seed, evaluation.numbers[densityXZScale], evaluation.numbers[densityYScale],
					evaluation.numbers[densityXZFactor], evaluation.numbers[densityYFactor], evaluation.numbers[densitySmearScale])
			} else {
				b = NewBlendedNoise(s.Seed, evaluation.numbers[densityXZScale], evaluation.numbers[densityYScale],
					evaluation.numbers[densityXZFactor], evaluation.numbers[densityYFactor], evaluation.numbers[densitySmearScale])
			}
			s.blended[n] = b
		}
		return b.Noise(x, y, z), nil
	case densityOpWeirdScaledSampler:
		input, err := child(densityInput)
		if err != nil {
			return 0, err
		}
		var rarity float64
		if n.Text["rarity_value_mapper"] == "type_1" {
			switch {
			case input < -0.5:
				rarity = 0.75
			case input < 0:
				rarity = 1
			case input < 0.5:
				rarity = 1.5
			default:
				rarity = 2
			}
		} else {
			switch {
			case input < -0.75:
				rarity = 0.5
			case input < -0.5:
				rarity = 0.75
			case input < 0.5:
				rarity = 1
			case input < 0.75:
				rarity = 2
			default:
				rarity = 3
			}
		}
		noise, err := s.nodeNoise(n)
		if err != nil {
			return 0, err
		}
		return rarity * math.Abs(noise.Noise(float64(x)/rarity, float64(y)/rarity, float64(z)/rarity)), nil
	default:
		return 0, fmt.Errorf("direct sampling of %s is not implemented (node %s)", n.Kind, n.Name)
	}
}

func (s *DensitySampler) interpolate(n *DensityNode, x, y, z int32) (float64, error) {
	x0 := floorCell(x, s.cellWidth)
	y0 := floorCell(y, s.cellHeight)
	z0 := floorCell(z, s.cellWidth)
	x1, y1, z1 := x0+s.cellWidth, y0+s.cellHeight, z0+s.cellWidth
	fx := float64(x-x0) / float64(s.cellWidth)
	fy := float64(y-y0) / float64(s.cellHeight)
	fz := float64(z-z0) / float64(s.cellWidth)
	var corners [2][2][2]float64
	var cell *interpolatedDensityCell
	runtime := s.nodeRuntime(n)
	position := densityPosition{x0, y0, z0}
	if s.chunk != nil {
		if runtime != nil {
			cell = runtime.cell
		} else {
			cell = s.chunk.interpolated[n]
		}
	}
	if cell != nil && cell.position == position {
		corners = cell.corners
	} else {
		// Blocks in one noise cell reuse exactly the same eight corners. The
		// slower coordinate cache is consulted only when entering another cell.
		for zi, zz := range [...]int32{z0, z1} {
			for xi, xx := range [...]int32{x0, x1} {
				for yi, yy := range [...]int32{y0, y1} {
					v, err := s.sampleCorner(n, xx, yy, zz)
					if err != nil {
						return 0, err
					}
					corners[zi][xi][yi] = v
				}
			}
		}
		if s.chunk != nil {
			if cell == nil {
				cell = &interpolatedDensityCell{}
				if runtime != nil {
					runtime.cell = cell
				} else {
					s.chunk.interpolated[n] = cell
				}
			}
			cell.position, cell.corners = position, corners
		}
	}
	lerp := func(t, a, b float64) float64 { return a + t*(b-a) }
	if s.chunk != nil && s.chunk.fillingCell {
		return lerp(fz,
			lerp(fy, lerp(fx, corners[0][0][0], corners[0][1][0]), lerp(fx, corners[0][0][1], corners[0][1][1])),
			lerp(fy, lerp(fx, corners[1][0][0], corners[1][1][0]), lerp(fx, corners[1][0][1], corners[1][1][1]))), nil
	}
	return lerp(fz,
		lerp(fx, lerp(fy, corners[0][0][0], corners[0][0][1]), lerp(fy, corners[0][1][0], corners[0][1][1])),
		lerp(fx, lerp(fy, corners[1][0][0], corners[1][0][1]), lerp(fy, corners[1][1][0], corners[1][1][1]))), nil
}

func floorCell(v, size int32) int32 {
	if size > 0 && size&(size-1) == 0 {
		return v &^ (size - 1)
	}
	q := v / size
	if v%size < 0 {
		q--
	}
	return q * size
}

// Vanilla CubicSpline uses float arithmetic even when its coordinates are
// density functions (which themselves return doubles).
func (s *DensitySampler) sampleSpline(spline *DensitySpline, x, y, z int32) (float32, error) {
	coordinate, err := s.Sample(spline.Coordinate, x, y, z)
	if err != nil {
		return 0, err
	}
	v := float32(coordinate)
	points := spline.Points
	last := len(points) - 1
	value := func(i int) (float32, error) {
		if points[i].Constant != nil {
			return float32(*points[i].Constant), nil
		}
		return s.sampleSpline(points[i].Spline, x, y, z)
	}
	extend := func(i int) (float32, error) {
		base, err := value(i)
		if err != nil {
			return 0, err
		}
		derivative := float32(points[i].Derivative)
		if derivative == 0 {
			return base, nil
		}
		return base + derivative*(v-float32(points[i].Location)), nil
	}
	if v < float32(points[0].Location) {
		return extend(0)
	}
	if v >= float32(points[last].Location) {
		return extend(last)
	}
	for i := 0; i < last; i++ {
		left, right := float32(points[i].Location), float32(points[i+1].Location)
		if v >= right {
			continue
		}
		t := (v - left) / (right - left)
		a, err := value(i)
		if err != nil {
			return 0, err
		}
		b, err := value(i + 1)
		if err != nil {
			return 0, err
		}
		length := right - left
		m := float32(points[i].Derivative)*length - (b - a)
		n := -float32(points[i+1].Derivative)*length + (b - a)
		return a + t*(b-a) + t*(1-t)*(m+t*(n-m)), nil
	}
	return 0, fmt.Errorf("spline interval not found")
}
