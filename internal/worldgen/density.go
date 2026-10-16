package worldgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// DensityNode is a parsed 1.21.1 density function. It retains the complete
// expression tree; evaluation support is added separately as each vanilla
// generation stage is reproduced. References share the same parsed node.
type DensityNode struct {
	Kind       string
	Name       string
	Noise      string
	Children   map[string]*DensityNode
	Numbers    map[string]float64
	Spline     *DensitySpline
	Text       map[string]string
	evaluation *densityEvaluation
}

type DensitySpline struct {
	Coordinate *DensityNode
	Points     []DensitySplinePoint
}

type DensitySplinePoint struct {
	Location   float64
	Derivative float64
	Constant   *float64
	Spline     *DensitySpline
}

type DensitySettings struct {
	Name   string
	Router map[string]*DensityNode
	Raw    json.RawMessage
	// Vanilla noise cells are four blocks per horizontal size unit and four
	// blocks per vertical size unit. Interpolated density nodes use these sizes.
	CellWidth  int32
	CellHeight int32
	MinY       int32
	Height     int32
	SeaLevel   int32
	Aquifers   bool
	OreVeins   bool
	Legacy     bool
	DefaultBlock SurfaceBlockState
	DefaultFluid SurfaceBlockState
}

type densityCompiler struct {
	registry *densityRegistry
	parsed   map[string]*DensityNode
	visiting map[string]bool
}

func newDensityCompiler() (*densityCompiler, error) {
	r, err := getDensityRegistry()
	if err != nil {
		return nil, err
	}
	return &densityCompiler{registry: r, parsed: make(map[string]*DensityNode), visiting: make(map[string]bool)}, nil
}

// LoadDensityFunction resolves and validates one built-in registry entry.
func LoadDensityFunction(name string) (*DensityNode, error) {
	c, err := newDensityCompiler()
	if err != nil {
		return nil, err
	}
	return c.reference(densityName(name))
}

// LoadDensitySettings resolves all fifteen router functions and preserves the
// original settings JSON for the later surface-rule and aquifer stages.
func LoadDensitySettings(name string) (*DensitySettings, error) {
	c, err := newDensityCompiler()
	if err != nil {
		return nil, err
	}
	name = densityName(name)
	raw, ok := c.registry.Settings[name]
	if !ok {
		return nil, fmt.Errorf("unknown 1.21.1 noise settings %q", name)
	}
	var settings struct {
		Router   map[string]json.RawMessage `json:"noise_router"`
		SeaLevel int32                      `json:"sea_level"`
		Aquifers bool                       `json:"aquifers_enabled"`
		OreVeins bool                       `json:"ore_veins_enabled"`
		Legacy   bool                       `json:"legacy_random_source"`
		DefaultBlock SurfaceBlockState       `json:"default_block"`
		DefaultFluid SurfaceBlockState       `json:"default_fluid"`
		Noise    struct {
			MinY       int32 `json:"min_y"`
			Height     int32 `json:"height"`
			Horizontal int32 `json:"size_horizontal"`
			Vertical   int32 `json:"size_vertical"`
		} `json:"noise"`
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, fmt.Errorf("noise settings %s: %w", name, err)
	}
	if len(settings.Router) != 15 {
		return nil, fmt.Errorf("noise settings %s: expected 15 router fields, got %d", name, len(settings.Router))
	}
	if settings.Noise.Horizontal < 1 || settings.Noise.Vertical < 1 {
		return nil, fmt.Errorf("noise settings %s: invalid cell size", name)
	}
	result := &DensitySettings{Name: name, Router: make(map[string]*DensityNode), Raw: append([]byte(nil), raw...), CellWidth: settings.Noise.Horizontal * 4, CellHeight: settings.Noise.Vertical * 4}
	result.MinY, result.Height, result.SeaLevel = settings.Noise.MinY, settings.Noise.Height, settings.SeaLevel
	result.Aquifers, result.OreVeins, result.Legacy = settings.Aquifers, settings.OreVeins, settings.Legacy
	result.DefaultBlock, result.DefaultFluid = settings.DefaultBlock, settings.DefaultFluid
	for field, expression := range settings.Router {
		node, err := c.parse(expression, name+"/noise_router/"+field)
		if err != nil {
			return nil, err
		}
		result.Router[field] = node
	}
	return result, nil
}

func (c *densityCompiler) reference(name string) (*DensityNode, error) {
	if n := c.parsed[name]; n != nil {
		return n, nil
	}
	if c.visiting[name] {
		return nil, fmt.Errorf("density function reference cycle at %s", name)
	}
	raw, ok := c.registry.Functions[name]
	if !ok {
		return nil, fmt.Errorf("unknown 1.21.1 density function %q", name)
	}
	c.visiting[name] = true
	n, err := c.parse(raw, name)
	delete(c.visiting, name)
	if err != nil {
		return nil, err
	}
	n.Name = name
	c.parsed[name] = n
	return n, nil
}

func (c *densityCompiler) parse(raw json.RawMessage, path string) (*DensityNode, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%s: missing density expression", path)
	}
	if raw[0] == '"' {
		var name string
		if err := json.Unmarshal(raw, &name); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		return c.reference(densityName(name))
	}
	if raw[0] != '{' {
		var value float64
		if err := json.Unmarshal(raw, &value); err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return nil, fmt.Errorf("%s: expected finite constant or density object", path)
		}
		n := &DensityNode{Kind: "constant", Numbers: map[string]float64{"value": value}}
		n.evaluation = prepareDensityEvaluation(n)
		return n, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var fullKind string
	if err := json.Unmarshal(fields["type"], &fullKind); err != nil || !strings.HasPrefix(fullKind, "minecraft:") {
		return nil, fmt.Errorf("%s: missing or invalid density type", path)
	}
	n := &DensityNode{Kind: strings.TrimPrefix(fullKind, "minecraft:"), Children: make(map[string]*DensityNode), Numbers: make(map[string]float64), Text: make(map[string]string)}
	allowed := map[string]bool{"type": true}
	child := func(key string) error {
		allowed[key] = true
		a, err := c.parse(fields[key], path+"/"+key)
		if err == nil {
			n.Children[key] = a
		}
		return err
	}
	number := func(key string) error {
		allowed[key] = true
		v, err := parseDensityNumber(fields[key])
		if err != nil {
			return fmt.Errorf("%s/%s: %w", path, key, err)
		}
		n.Numbers[key] = v
		return nil
	}
	text := func(key string) error {
		allowed[key] = true
		var value string
		if err := json.Unmarshal(fields[key], &value); err != nil || value == "" {
			return fmt.Errorf("%s/%s: expected nonempty string", path, key)
		}
		n.Text[key] = value
		return nil
	}
	var err error
	switch n.Kind {
	case "blend_alpha", "blend_offset", "end_islands":
	case "flat_cache", "cache_2d", "cache_once", "interpolated", "blend_density", "abs", "cube", "half_negative", "quarter_negative", "square", "squeeze":
		err = child("argument")
	case "add", "mul", "min", "max":
		if err = child("argument1"); err == nil {
			err = child("argument2")
		}
	case "clamp":
		if err = child("input"); err == nil {
			err = number("min")
		}
		if err == nil {
			err = number("max")
		}
	case "range_choice":
		for _, key := range []string{"input", "when_in_range", "when_out_of_range"} {
			if err = child(key); err != nil {
				break
			}
		}
		if err == nil {
			err = number("min_inclusive")
		}
		if err == nil {
			err = number("max_exclusive")
		}
	case "noise", "shift_a", "shift_b", "shift":
		key := "noise"
		if n.Kind != "noise" {
			key = "argument"
		}
		if err = text(key); err == nil {
			n.Noise = densityName(n.Text[key])
			_, err = Parameters(n.Noise)
		}
		if err == nil && n.Kind == "noise" {
			if err = number("xz_scale"); err == nil {
				err = number("y_scale")
			}
		}
	case "shifted_noise":
		if err = text("noise"); err == nil {
			n.Noise = densityName(n.Text["noise"])
			_, err = Parameters(n.Noise)
		}
		for _, key := range []string{"shift_x", "shift_y", "shift_z"} {
			if err == nil {
				err = child(key)
			}
		}
		for _, key := range []string{"xz_scale", "y_scale"} {
			if err == nil {
				err = number(key)
			}
		}
	case "y_clamped_gradient":
		for _, key := range []string{"from_y", "to_y", "from_value", "to_value"} {
			if err = number(key); err != nil {
				break
			}
		}
	case "spline":
		allowed["spline"] = true
		n.Spline, err = c.parseSpline(fields["spline"], path+"/spline")
	case "old_blended_noise":
		for _, key := range []string{"xz_scale", "y_scale", "xz_factor", "y_factor", "smear_scale_multiplier"} {
			if err = number(key); err != nil {
				break
			}
		}
	case "weird_scaled_sampler":
		if err = child("input"); err == nil {
			err = text("noise")
		}
		if err == nil {
			n.Noise = densityName(n.Text["noise"])
			_, err = Parameters(n.Noise)
		}
		if err == nil {
			err = text("rarity_value_mapper")
		}
		if err == nil && n.Text["rarity_value_mapper"] != "type_1" && n.Text["rarity_value_mapper"] != "type_2" {
			err = fmt.Errorf("%s/rarity_value_mapper: unknown mapper %q", path, n.Text["rarity_value_mapper"])
		}
	default:
		return nil, fmt.Errorf("%s: unsupported density type %s", path, fullKind)
	}
	if err != nil {
		return nil, err
	}
	for key := range fields {
		if !allowed[key] {
			return nil, fmt.Errorf("%s: unexpected field %q for %s", path, key, fullKind)
		}
	}
	n.evaluation = prepareDensityEvaluation(n)
	return n, nil
}

func parseDensityNumber(raw json.RawMessage) (float64, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("missing number")
	}
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, fmt.Errorf("expected finite number")
	}
	return value, nil
}

func (c *densityCompiler) parseSpline(raw json.RawMessage, path string) (*DensitySpline, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("%s: expected spline object", path)
	}
	if len(fields) != 2 {
		return nil, fmt.Errorf("%s: expected coordinate and points", path)
	}
	coordinate, err := c.parse(fields["coordinate"], path+"/coordinate")
	if err != nil {
		return nil, err
	}
	var points []json.RawMessage
	if err := json.Unmarshal(fields["points"], &points); err != nil || len(points) == 0 {
		return nil, fmt.Errorf("%s: expected nonempty spline points", path)
	}
	s := &DensitySpline{Coordinate: coordinate}
	for i, rawPoint := range points {
		pointPath := fmt.Sprintf("%s/points/%d", path, i)
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(rawPoint, &fields); err != nil || len(fields) != 3 {
			return nil, fmt.Errorf("%s: expected location, derivative, and value", pointPath)
		}
		location, err := parseDensityNumber(fields["location"])
		if err != nil {
			return nil, fmt.Errorf("%s/location: %w", pointPath, err)
		}
		derivative, err := parseDensityNumber(fields["derivative"])
		if err != nil {
			return nil, fmt.Errorf("%s/derivative: %w", pointPath, err)
		}
		if len(s.Points) > 0 && location <= s.Points[len(s.Points)-1].Location {
			return nil, fmt.Errorf("%s: spline locations must increase", pointPath)
		}
		point := DensitySplinePoint{Location: location, Derivative: derivative}
		value := bytes.TrimSpace(fields["value"])
		if len(value) == 0 {
			return nil, fmt.Errorf("%s/value: missing value", pointPath)
		}
		if value[0] == '{' {
			point.Spline, err = c.parseSpline(value, pointPath+"/value")
		} else {
			var constant float64
			constant, err = parseDensityNumber(value)
			point.Constant = &constant
		}
		if err != nil {
			return nil, err
		}
		s.Points = append(s.Points, point)
	}
	return s, nil
}
