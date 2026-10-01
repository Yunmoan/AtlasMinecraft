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
	children [9]*DensityNode
	numbers  [14]float64
}

func prepareDensityEvaluation(n *DensityNode) *densityEvaluation {
	e := &densityEvaluation{}
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
