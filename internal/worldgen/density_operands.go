package worldgen

const (
	densityArgument = iota
	densityInput
	densityLeft
	densityRight
	densityWhenIn
	densityWhenOut
	densityShiftX
	densityShiftY
	densityShiftZ
)

const (
	densityValue = iota
	densityMin
	densityMax
	densityMinInclusive
	densityMaxExclusive
	densityFromY
	densityToY
	densityFromValue
	densityToValue
	densityXZScale
	densityYScale
	densityXZFactor
	densityYFactor
	densitySmearScale
)

// Resolve string-keyed operands once while parsing an immutable expression
// tree. Sampling millions of blocks then uses array loads, retaining the
// original tree for diagnostics and the original arithmetic order.
type densityEvaluation struct {
	op       densityOpcode
	children [9]*DensityNode
	numbers  [14]float64
	program  *densityProgram
	slot     int
}

// Each parsed settings graph owns its slot namespace. A sampler can bind this
// immutable graph to private runtime state without hashing node pointers at
// every block. Nodes from a different graph keep using the general maps.
type densityProgram struct{ nodes int }

type densityRuntime struct {
	noise   *NormalNoise
	cell    *interpolatedDensityCell
	corners *densityCornerValues
}

func (c *densityCompiler) prepare(n *DensityNode) {
	n.evaluation = prepareDensityEvaluation(n)
	n.evaluation.program, n.evaluation.slot = c.program, c.program.nodes
	c.program.nodes++
}

func (s *DensitySampler) bind(settings *DensitySettings) {
	s.program = settings.program
	if s.program != nil {
		s.runtime = make([]densityRuntime, s.program.nodes)
	}
}

func (s *DensitySampler) nodeRuntime(n *DensityNode) *densityRuntime {
	if n == nil {
		return nil
	}
	if e := n.evaluation; e != nil && e.program == s.program && e.slot >= 0 && e.slot < len(s.runtime) {
		return &s.runtime[e.slot]
	}
	return nil
}

func prepareDensityEvaluation(n *DensityNode) *densityEvaluation {
	e := &densityEvaluation{op: densityOpcodes[n.Kind]}
	for i, key := range [...]string{"argument", "input", "argument1", "argument2", "when_in_range", "when_out_of_range", "shift_x", "shift_y", "shift_z"} {
		e.children[i] = n.Children[key]
	}
	for i, key := range [...]string{"value", "min", "max", "min_inclusive", "max_exclusive", "from_y", "to_y", "from_value", "to_value", "xz_scale", "y_scale", "xz_factor", "y_factor", "smear_scale_multiplier"} {
		e.numbers[i] = n.Numbers[key]
	}
	return e
}

func (n *DensityNode) argument() *DensityNode {
	if n.evaluation != nil {
		return n.evaluation.children[densityArgument]
	}
	return n.Children["argument"]
}

// Dispatch parsed density operators with integer codes. Only the
// dispatch representation changes; operand evaluation order is unchanged.
type densityOpcode uint8

const (
	densityOpUnknown densityOpcode = iota
	densityOpCache2D
	densityOpConstant
	densityOpBlendAlpha
	densityOpBlendOffset
	densityOpBeardifier
	densityOpEndIslands
	densityOpFlatCache
	densityOpCacheAllInCell
	densityOpCacheOnce
	densityOpBlendDensity
	densityOpInterpolated
	densityOpAdd
	densityOpMul
	densityOpMin
	densityOpMax
	densityOpAbs
	densityOpSquare
	densityOpCube
	densityOpHalfNegative
	densityOpQuarterNegative
	densityOpSqueeze
	densityOpClamp
	densityOpRangeChoice
	densityOpYClampedGradient
	densityOpNoise
	densityOpShift
	densityOpShiftA
	densityOpShiftB
	densityOpShiftedNoise
	densityOpSpline
	densityOpOldBlendedNoise
	densityOpWeirdScaledSampler
)

var densityOpcodes = map[string]densityOpcode{
	"cache_2d":             densityOpCache2D,
	"constant":             densityOpConstant,
	"blend_alpha":          densityOpBlendAlpha,
	"blend_offset":         densityOpBlendOffset,
	"beardifier":           densityOpBeardifier,
	"end_islands":          densityOpEndIslands,
	"flat_cache":           densityOpFlatCache,
	"cache_all_in_cell":    densityOpCacheAllInCell,
	"cache_once":           densityOpCacheOnce,
	"blend_density":        densityOpBlendDensity,
	"interpolated":         densityOpInterpolated,
	"add":                  densityOpAdd,
	"mul":                  densityOpMul,
	"min":                  densityOpMin,
	"max":                  densityOpMax,
	"abs":                  densityOpAbs,
	"square":               densityOpSquare,
	"cube":                 densityOpCube,
	"half_negative":        densityOpHalfNegative,
	"quarter_negative":     densityOpQuarterNegative,
	"squeeze":              densityOpSqueeze,
	"clamp":                densityOpClamp,
	"range_choice":         densityOpRangeChoice,
	"y_clamped_gradient":   densityOpYClampedGradient,
	"noise":                densityOpNoise,
	"shift":                densityOpShift,
	"shift_a":              densityOpShiftA,
	"shift_b":              densityOpShiftB,
	"shifted_noise":        densityOpShiftedNoise,
	"spline":               densityOpSpline,
	"old_blended_noise":    densityOpOldBlendedNoise,
	"weird_scaled_sampler": densityOpWeirdScaledSampler,
}
