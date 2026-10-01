package worldgen

import (
	"fmt"
	"math"
)

// DensitySampler evaluates direct density expressions for a single block
// position. NoiseChunk's cell interpolation, old blended terrain noise, End
// islands, and other contextual generation stages are deliberately rejected.
// One sampler must not be used concurrently by multiple goroutines.
type DensitySampler struct {
	Seed                  int64
	noises                map[string]*NormalNoise
	blended               map[*DensityNode]*BlendedNoise
	cellWidth, cellHeight int32
	chunk                 *chunkDensityContext
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
	return s
}

func (s *DensitySampler) noise(name string) (*NormalNoise, error) {
	if n := s.noises[name]; n != nil {
		return n, nil
	}
	n, err := NewNamedNormalNoise(s.Seed, name)
	if err != nil {
		return nil, err
	}
	s.noises[name] = n
	return n, nil
}

func (s *DensitySampler) Sample(n *DensityNode, x, y, z int32) (float64, error) {
	if n == nil {
		return 0, fmt.Errorf("nil density node")
	}
	child := func(key string) (float64, error) {
		return s.Sample(n.Children[key], x, y, z)
	}
	switch n.Kind {
	case "constant":
		return n.Numbers["value"], nil
	case "blend_alpha":
		return 1, nil // No old/new chunk blending in a direct sample.
	case "blend_offset":
		return 0, nil
	case "flat_cache":
		if s.chunk != nil {
			return s.sampleFlatCache(n, x, y, z)
		}
		return child("argument")
	case "cache_2d":
		if s.chunk != nil {
			return s.sampleCache2D(n, x, y, z)
		}
		return child("argument")
	case "cache_all_in_cell":
		if s.chunk == nil {
			return 0, fmt.Errorf("cell cache requires a NoiseChunk")
		}
		previous := s.chunk.fillingCell
		s.chunk.fillingCell = true
		value, err := child("argument")
		s.chunk.fillingCell = previous
		return value, err
	case "cache_once", "blend_density":
		return child("argument")
	case "interpolated":
		if s.chunk != nil && s.chunk.singlePoint {
			return child("argument")
		}
		if s.cellWidth <= 0 || s.cellHeight <= 0 {
			return 0, fmt.Errorf("interpolated density requires a noise cell")
		}
		return s.interpolate(n.Children["argument"], x, y, z)
	case "add", "mul", "min", "max":
		a, err := child("argument1")
		if err != nil {
			return 0, err
		}
		b, err := child("argument2")
		if err != nil {
			return 0, err
		}
		switch n.Kind {
		case "add":
			return a + b, nil
		case "mul":
			return a * b, nil
		case "min":
			return math.Min(a, b), nil
		default:
			return math.Max(a, b), nil
		}
	case "abs", "square", "cube", "half_negative", "quarter_negative", "squeeze":
		v, err := child("argument")
		if err != nil {
			return 0, err
		}
		switch n.Kind {
		case "abs":
			return math.Abs(v), nil
		case "square":
			return v * v, nil
		case "cube":
			return v * v * v, nil
		case "half_negative":
			if v < 0 {
				return v * 0.5, nil
			}
		case "quarter_negative":
			if v < 0 {
				return v * 0.25, nil
			}
		case "squeeze":
			v = math.Max(-1, math.Min(1, v))
			return v/2 - v*v*v/24, nil
		}
		return v, nil
	case "clamp":
		v, err := child("input")
		if err != nil {
			return 0, err
		}
		return math.Max(n.Numbers["min"], math.Min(n.Numbers["max"], v)), nil
	case "range_choice":
		v, err := child("input")
		if err != nil {
			return 0, err
		}
		if v >= n.Numbers["min_inclusive"] && v < n.Numbers["max_exclusive"] {
			return child("when_in_range")
		}
		return child("when_out_of_range")
	case "y_clamped_gradient":
		from, to := n.Numbers["from_y"], n.Numbers["to_y"]
		if from == to {
			return 0, fmt.Errorf("zero-width y gradient")
		}
		t := math.Max(0, math.Min(1, (float64(y)-from)/(to-from)))
		return n.Numbers["from_value"] + t*(n.Numbers["to_value"]-n.Numbers["from_value"]), nil
	case "noise":
		noise, err := s.noise(n.Noise)
		if err != nil {
			return 0, err
		}
		xz, ys := n.Numbers["xz_scale"], n.Numbers["y_scale"]
		return noise.Noise(float64(x)*xz, float64(y)*ys, float64(z)*xz), nil
	case "shift", "shift_a", "shift_b":
		noise, err := s.noise(n.Noise)
		if err != nil {
			return 0, err
		}
		fx, fy, fz := float64(x), float64(y), float64(z)
		switch n.Kind {
		case "shift_a":
			fy = 0
		case "shift_b":
			fx, fy, fz = fz, fx, 0
		}
		return 4 * noise.Noise(fx*0.25, fy*0.25, fz*0.25), nil
	case "shifted_noise":
		shiftX, err := child("shift_x")
		if err != nil {
			return 0, err
		}
		shiftY, err := child("shift_y")
		if err != nil {
			return 0, err
		}
		shiftZ, err := child("shift_z")
		if err != nil {
			return 0, err
		}
		noise, err := s.noise(n.Noise)
		if err != nil {
			return 0, err
		}
		xz, ys := n.Numbers["xz_scale"], n.Numbers["y_scale"]
		return noise.Noise(float64(x)*xz+shiftX, float64(y)*ys+shiftY, float64(z)*xz+shiftZ), nil
	case "spline":
		v, err := s.sampleSpline(n.Spline, x, y, z)
		return float64(v), err
	case "old_blended_noise":
		b := s.blended[n]
		if b == nil {
			b = NewBlendedNoise(s.Seed, n.Numbers["xz_scale"], n.Numbers["y_scale"],
				n.Numbers["xz_factor"], n.Numbers["y_factor"], n.Numbers["smear_scale_multiplier"])
			s.blended[n] = b
		}
		return b.Noise(x, y, z), nil
	case "weird_scaled_sampler":
		input, err := child("input")
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
		noise, err := s.noise(n.Noise)
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
	position := densityPosition{x0, y0, z0}
	if s.chunk != nil {
		cell = s.chunk.interpolated[n]
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
				s.chunk.interpolated[n] = cell
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
