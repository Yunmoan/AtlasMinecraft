package worldgen

import (
	"encoding/json"
	"fmt"
	"math"
)

type SurfaceBlockState struct {
	Name       string            `json:"Name"`
	Properties map[string]string `json:"Properties,omitempty"`
}

type SurfaceContext struct {
	X, Y, Z                          int32
	Biome                            string
	SurfaceDepth                     int32
	SecondaryDepth                   float64
	StoneDepthAbove, StoneDepthBelow int32
	WaterHeight                      int32
	MinSurfaceLevel                  int32
	Cold, Steep                      bool
}

type surfaceAnchor struct {
	Absolute    *int32 `json:"absolute"`
	AboveBottom *int32 `json:"above_bottom"`
	BelowTop    *int32 `json:"below_top"`
}

func (anchor surfaceAnchor) resolve(settings *DensitySettings) (int32, error) {
	count := 0
	var result int32
	if anchor.Absolute != nil {
		count++
		result = *anchor.Absolute
	}
	if anchor.AboveBottom != nil {
		count++
		result = settings.MinY + *anchor.AboveBottom
	}
	if anchor.BelowTop != nil {
		count++
		result = settings.MinY + settings.Height - 1 - *anchor.BelowTop
	}
	if count != 1 {
		return 0, fmt.Errorf("surface anchor must contain exactly one position")
	}
	return result, nil
}

type surfaceCondition struct {
	Type                   string            `json:"type"`
	BiomeIs                []string          `json:"biome_is"`
	Noise                  string            `json:"noise"`
	MinThreshold           float64           `json:"min_threshold"`
	MaxThreshold           float64           `json:"max_threshold"`
	RandomName             string            `json:"random_name"`
	TrueAtAndBelow         surfaceAnchor     `json:"true_at_and_below"`
	FalseAtAndAbove        surfaceAnchor     `json:"false_at_and_above"`
	Anchor                 surfaceAnchor     `json:"anchor"`
	SurfaceDepthMultiplier int32             `json:"surface_depth_multiplier"`
	AddStoneDepth          bool              `json:"add_stone_depth"`
	Offset                 int32             `json:"offset"`
	AddSurfaceDepth        bool              `json:"add_surface_depth"`
	SecondaryDepthRange    int32             `json:"secondary_depth_range"`
	SurfaceType            string            `json:"surface_type"`
	Invert                 *surfaceCondition `json:"invert"`
}

type surfaceRule struct {
	Type        string             `json:"type"`
	ResultState *SurfaceBlockState `json:"result_state"`
	IfTrue      *surfaceCondition  `json:"if_true"`
	ThenRun     *surfaceRule       `json:"then_run"`
	Sequence    []*surfaceRule     `json:"sequence"`
}

type SurfaceRuleSampler struct {
	settings *DensitySettings
	rule     *surfaceRule
	random   PositionalRandom
	noises   map[string]*NormalNoise
	randoms  map[string]PositionalRandom
	bands    [192]SurfaceBlockState
}

func NewSurfaceRuleSampler(seed int64, settings *DensitySettings) (*SurfaceRuleSampler, error) {
	if settings == nil || settings.Height <= 0 || settings.Legacy {
		return nil, fmt.Errorf("surface sampler requires non-legacy noise settings")
	}
	var raw struct {
		Rule *surfaceRule `json:"surface_rule"`
	}
	if err := json.Unmarshal(settings.Raw, &raw); err != nil {
		return nil, err
	}
	sampler := &SurfaceRuleSampler{
		settings: settings, rule: raw.Rule,
		random: NewXoroshiro(seed).ForkPositional(),
		noises: make(map[string]*NormalNoise), randoms: make(map[string]PositionalRandom),
	}
	if err := sampler.validateRule(raw.Rule); err != nil {
		return nil, err
	}
	sampler.initializeBands()
	return sampler, nil
}

func (sampler *SurfaceRuleSampler) validateRule(rule *surfaceRule) error {
	if rule == nil {
		return fmt.Errorf("surface rule is missing")
	}
	switch rule.Type {
	case "minecraft:block":
		if rule.ResultState == nil || rule.ResultState.Name == "" {
			return fmt.Errorf("surface block state is missing")
		}
	case "minecraft:sequence":
		if len(rule.Sequence) == 0 {
			return fmt.Errorf("surface sequence is empty")
		}
		for _, child := range rule.Sequence {
			if err := sampler.validateRule(child); err != nil {
				return err
			}
		}
	case "minecraft:condition":
		if err := sampler.validateCondition(rule.IfTrue); err != nil {
			return err
		}
		return sampler.validateRule(rule.ThenRun)
	case "minecraft:bandlands":
	default:
		return fmt.Errorf("unsupported surface rule %q", rule.Type)
	}
	return nil
}

func (sampler *SurfaceRuleSampler) validateCondition(condition *surfaceCondition) error {
	if condition == nil {
		return fmt.Errorf("surface condition is missing")
	}
	switch condition.Type {
	case "minecraft:biome":
		if len(condition.BiomeIs) == 0 {
			return fmt.Errorf("surface biome condition is empty")
		}
	case "minecraft:noise_threshold":
		if condition.MinThreshold > condition.MaxThreshold {
			return fmt.Errorf("surface noise thresholds are inverted")
		}
		_, err := Parameters(condition.Noise)
		return err
	case "minecraft:vertical_gradient":
		low, err := condition.TrueAtAndBelow.resolve(sampler.settings)
		if err != nil {
			return err
		}
		high, err := condition.FalseAtAndAbove.resolve(sampler.settings)
		if err != nil {
			return err
		}
		if high <= low || condition.RandomName == "" {
			return fmt.Errorf("invalid surface vertical gradient")
		}
	case "minecraft:y_above":
		_, err := condition.Anchor.resolve(sampler.settings)
		return err
	case "minecraft:water", "minecraft:temperature", "minecraft:steep", "minecraft:hole", "minecraft:above_preliminary_surface":
	case "minecraft:not":
		return sampler.validateCondition(condition.Invert)
	case "minecraft:stone_depth":
		if condition.SurfaceType != "floor" && condition.SurfaceType != "ceiling" {
			return fmt.Errorf("invalid stone depth surface %q", condition.SurfaceType)
		}
	default:
		return fmt.Errorf("unsupported surface condition %q", condition.Type)
	}
	return nil
}

func (sampler *SurfaceRuleSampler) noise(name string) (*NormalNoise, error) {
	if noise := sampler.noises[name]; noise != nil {
		return noise, nil
	}
	parameters, err := Parameters(name)
	if err != nil {
		return nil, err
	}
	noise, err := NewNormalNoise(sampler.random.FromHashOf(densityName(name)), parameters.FirstOctave, parameters.Amplitudes)
	if err != nil {
		return nil, err
	}
	sampler.noises[name] = noise
	return noise, nil
}

func (sampler *SurfaceRuleSampler) SurfaceDepth(x, z int32) (int32, error) {
	noise, err := sampler.noise("minecraft:surface")
	if err != nil {
		return 0, err
	}
	value := noise.Noise(float64(x), 0, float64(z))
	return int32(value*2.75 + 3 + sampler.random.At(x, 0, z).NextDouble()*0.25), nil
}

func (sampler *SurfaceRuleSampler) SecondaryDepth(x, z int32) (float64, error) {
	noise, err := sampler.noise("minecraft:surface_secondary")
	if err != nil {
		return 0, err
	}
	return noise.Noise(float64(x), 0, float64(z)), nil
}

func (sampler *SurfaceRuleSampler) Apply(context SurfaceContext) (*SurfaceBlockState, error) {
	return sampler.applyRule(sampler.rule, context)
}

func (sampler *SurfaceRuleSampler) applyRule(rule *surfaceRule, context SurfaceContext) (*SurfaceBlockState, error) {
	switch rule.Type {
	case "minecraft:block":
		state := SurfaceBlockState{Name: rule.ResultState.Name}
		if rule.ResultState.Properties != nil {
			state.Properties = make(map[string]string, len(rule.ResultState.Properties))
			for key, value := range rule.ResultState.Properties {
				state.Properties[key] = value
			}
		}
		return &state, nil
	case "minecraft:sequence":
		for _, child := range rule.Sequence {
			state, err := sampler.applyRule(child, context)
			if err != nil || state != nil {
				return state, err
			}
		}
		return nil, nil
	case "minecraft:condition":
		match, err := sampler.matches(rule.IfTrue, context)
		if err != nil || !match {
			return nil, err
		}
		return sampler.applyRule(rule.ThenRun, context)
	case "minecraft:bandlands":
		return sampler.BandAt(context.X, context.Y, context.Z)
	default:
		return nil, fmt.Errorf("unsupported surface rule %q", rule.Type)
	}
}

func (sampler *SurfaceRuleSampler) matches(condition *surfaceCondition, context SurfaceContext) (bool, error) {
	switch condition.Type {
	case "minecraft:biome":
		for _, name := range condition.BiomeIs {
			if name == context.Biome {
				return true, nil
			}
		}
		return false, nil
	case "minecraft:noise_threshold":
		noise, err := sampler.noise(condition.Noise)
		if err != nil {
			return false, err
		}
		value := noise.Noise(float64(context.X), 0, float64(context.Z))
		return value >= condition.MinThreshold && value <= condition.MaxThreshold, nil
	case "minecraft:vertical_gradient":
		low, err := condition.TrueAtAndBelow.resolve(sampler.settings)
		if err != nil {
			return false, err
		}
		high, err := condition.FalseAtAndAbove.resolve(sampler.settings)
		if err != nil {
			return false, err
		}
		if context.Y <= low {
			return true, nil
		}
		if context.Y >= high {
			return false, nil
		}
		random, exists := sampler.randoms[condition.RandomName]
		if !exists {
			random = sampler.random.FromHashOf(condition.RandomName).ForkPositional()
			sampler.randoms[condition.RandomName] = random
		}
		threshold := 1 + float64(context.Y-low)/float64(high-low)*(-1)
		return float64(random.At(context.X, context.Y, context.Z).NextFloat()) < threshold, nil
	case "minecraft:y_above":
		anchor, err := condition.Anchor.resolve(sampler.settings)
		depth := int32(0)
		if condition.AddStoneDepth {
			depth = context.StoneDepthAbove
		}
		return context.Y+depth >= anchor+context.SurfaceDepth*condition.SurfaceDepthMultiplier, err
	case "minecraft:water":
		depth := int32(0)
		if condition.AddStoneDepth {
			depth = context.StoneDepthAbove
		}
		return context.WaterHeight == math.MinInt32 || context.Y+depth >= context.WaterHeight+condition.Offset+context.SurfaceDepth*condition.SurfaceDepthMultiplier, nil
	case "minecraft:stone_depth":
		depth := context.StoneDepthAbove
		if condition.SurfaceType == "ceiling" {
			depth = context.StoneDepthBelow
		}
		added := int32(0)
		if condition.AddSurfaceDepth {
			added = context.SurfaceDepth
		}
		secondary := int32(0)
		if condition.SecondaryDepthRange != 0 {
			secondary = int32((context.SecondaryDepth + 1) / 2 * float64(condition.SecondaryDepthRange))
		}
		return depth <= 1+condition.Offset+added+secondary, nil
	case "minecraft:temperature":
		return context.Cold, nil
	case "minecraft:steep":
		return context.Steep, nil
	case "minecraft:hole":
		return context.SurfaceDepth <= 0, nil
	case "minecraft:above_preliminary_surface":
		return context.Y >= context.MinSurfaceLevel, nil
	case "minecraft:not":
		match, err := sampler.matches(condition.Invert, context)
		return !match, err
	default:
		return false, fmt.Errorf("unsupported surface condition %q", condition.Type)
	}
}

func (sampler *SurfaceRuleSampler) initializeBands() {
	random := sampler.random.FromHashOf("minecraft:clay_bands")
	state := func(name string) SurfaceBlockState {
		return SurfaceBlockState{Name: "minecraft:" + name + "_terracotta"}
	}
	for index := range sampler.bands {
		sampler.bands[index] = SurfaceBlockState{Name: "minecraft:terracotta"}
	}
	for index := 0; index < len(sampler.bands); index++ {
		index += int(random.NextIntBound(5)) + 1
		if index < len(sampler.bands) {
			sampler.bands[index] = state("orange")
		}
	}
	for _, band := range []struct {
		base  int32
		color string
	}{{1, "yellow"}, {2, "brown"}, {1, "red"}} {
		count := random.NextIntBound(10) + 6
		for index := int32(0); index < count; index++ {
			width := band.base + random.NextIntBound(3)
			start := random.NextIntBound(192)
			for offset := int32(0); offset < width && start+offset < 192; offset++ {
				sampler.bands[start+offset] = state(band.color)
			}
		}
	}
	count := random.NextIntBound(7) + 9
	position := int32(0)
	for index := int32(0); index < count && position < 192; index++ {
		sampler.bands[position] = state("white")
		if position-1 > 0 && random.NextBoolean() {
			sampler.bands[position-1] = state("light_gray")
		}
		if position+1 < 192 && random.NextBoolean() {
			sampler.bands[position+1] = state("light_gray")
		}
		position += random.NextIntBound(16) + 4
	}
}

func (sampler *SurfaceRuleSampler) BandAt(x, y, z int32) (*SurfaceBlockState, error) {
	noise, err := sampler.noise("minecraft:clay_bands_offset")
	if err != nil {
		return nil, err
	}
	offset := int32(math.Floor(noise.Noise(float64(x), 0, float64(z))*4 + 0.5))
	index := (y + offset + 192) % 192
	if index < 0 {
		return nil, fmt.Errorf("badlands band position outside supported height")
	}
	state := sampler.bands[index]
	return &state, nil
}
