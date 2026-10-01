package worldgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltinDensityGraph(t *testing.T) {
	functions, err := BuiltinDensityNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(functions) != 35 {
		t.Fatalf("density functions: got %d, want 35", len(functions))
	}
	for _, name := range functions {
		if _, err := LoadDensityFunction(name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	settings, err := BuiltinNoiseSettingsNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(settings) != 7 {
		t.Fatalf("noise settings: got %d, want 7", len(settings))
	}
	for _, name := range settings {
		s, err := LoadDensitySettings(name)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(s.Router) != 15 || len(s.Raw) == 0 {
			t.Errorf("%s: incomplete router or raw settings", name)
		}
	}
}

func TestNoiseCellInterpolation(t *testing.T) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	if settings.CellWidth != 4 || settings.CellHeight != 8 {
		t.Fatalf("overworld cell size %d x %d", settings.CellWidth, settings.CellHeight)
	}
	gradient := &DensityNode{Kind: "y_clamped_gradient", Numbers: map[string]float64{
		"from_y": -16, "to_y": 16, "from_value": -1, "to_value": 1,
	}}
	node := &DensityNode{Kind: "interpolated", Children: map[string]*DensityNode{
		"argument": {Kind: "square", Children: map[string]*DensityNode{"argument": gradient}},
	}}
	if _, err := NewDensitySampler(0).Sample(node, -3, -5, 2); err == nil {
		t.Fatal("direct sampler accepted a contextual interpolated node")
	}
	sampler := NewCellDensitySampler(0, settings)
	// -5 belongs to the [-8,0) cell. The squared values at its corners
	// are 1/4 and 0; linear interpolation must give 5/32.
	got, err := sampler.Sample(node, -3, -5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got != 5.0/32.0 {
		t.Fatalf("interpolated value at negative coordinate = %v, want 5/32", got)
	}
	for _, p := range [][3]int32{{0, 64, 0}, {-37, -48, 113}} {
		v, err := sampler.Sample(settings.Router["final_density"], p[0], p[1], p[2])
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("overworld final density at %v: %v (%v)", p, v, err)
		}
	}
}

// DensityClimateProbe builds vanilla density nodes through the supplied JAR.
// Compare raw double bits, including the Shift and ShiftedNoise coordinate math.
func TestDensityClimateReference(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "density_climate_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	shifts := make(map[string]*DensityNode)
	for _, name := range []string{"shift_x", "shift_z"} {
		shifts[name], err = LoadDensityFunction(name)
		if err != nil {
			t.Fatal(err)
		}
	}
	points := [][3]int32{{0, 64, 0}, {173, -32, -421}, {-10000, 120, 24567}}
	sc := bufio.NewScanner(f)
	var sampler *DensitySampler
	var currentSeed int64
	count := 0
	for sc.Scan() {
		var label, field string
		var seed int64
		var index int
		var bits uint64
		if _, err := fmt.Sscan(sc.Text(), &label, &seed, &field, &index, &bits); err != nil || label != "density" || index < 0 || index >= len(points) {
			t.Fatalf("invalid reference line: %q (%v)", sc.Text(), err)
		}
		if sampler == nil || seed != currentSeed {
			sampler = NewDensitySampler(seed)
			currentSeed = seed
		}
		node := settings.Router[field]
		if node == nil {
			node = shifts[field]
		}
		if node == nil {
			t.Fatalf("unknown reference field %s", field)
		}
		p := points[index]
		got, err := sampler.Sample(node, p[0], p[1], p[2])
		if err != nil || math.Float64bits(got) != bits {
			t.Errorf("seed %d %s point %d: got %d, want %d (%v)", seed, field, index, math.Float64bits(got), bits, err)
		}
		count++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 84 {
		t.Fatalf("got %d reference points, want 84", count)
	}
}

func TestBlendedNoiseReference(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "blended_noise_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	points := [][3]int32{{0, 64, 0}, {173, -32, -421}, {-10000, 120, 24567}}
	nodes := make(map[string]*DensityNode)
	sc := bufio.NewScanner(f)
	count := 0
	for sc.Scan() {
		var label, name string
		var seed int64
		var index int
		var bits uint64
		if _, err := fmt.Sscan(sc.Text(), &label, &name, &seed, &index, &bits); err != nil || label != "blended" || index < 0 || index >= len(points) {
			t.Fatalf("invalid reference line %q: %v", sc.Text(), err)
		}
		if nodes[name] == nil {
			nodes[name], err = LoadDensityFunction(name + "/base_3d_noise")
			if err != nil {
				t.Fatal(err)
			}
		}
		p := points[index]
		got, err := NewDensitySampler(seed).Sample(nodes[name], p[0], p[1], p[2])
		if err != nil || math.Float64bits(got) != bits {
			t.Errorf("%s seed %d point %d: got %d, want %d (%v)", name, seed, index, math.Float64bits(got), bits, err)
		}
		count++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 36 {
		t.Fatalf("got %d reference points, want 36", count)
	}
}

func TestWeirdScaledReference(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "weird_scaled_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	compiler, err := newDensityCompiler()
	if err != nil {
		t.Fatal(err)
	}
	points := [][3]int32{{0, 64, 0}, {173, -32, -421}, {-10000, 120, 24567}}
	inputs := []float64{-0.9, -0.6, -0.2, 0.2, 0.6, 0.9}
	nodes := make(map[string]*DensityNode)
	samplers := make(map[int64]*DensitySampler)
	sc := bufio.NewScanner(f)
	count := 0
	for sc.Scan() {
		var label, mapper string
		var seed int64
		var inputIndex, pointIndex int
		var bits uint64
		if _, err := fmt.Sscan(sc.Text(), &label, &mapper, &seed, &inputIndex, &pointIndex, &bits); err != nil || label != "weird" || inputIndex < 0 || inputIndex >= len(inputs) || pointIndex < 0 || pointIndex >= len(points) {
			t.Fatalf("invalid reference line %q: %v", sc.Text(), err)
		}
		key := fmt.Sprintf("%s/%d", mapper, inputIndex)
		if nodes[key] == nil {
			name := "spaghetti_3d_1"
			if mapper == "type_2" {
				name = "spaghetti_2d"
			}
			expression := fmt.Sprintf(`{"type":"minecraft:weird_scaled_sampler","input":%.1f,"noise":"minecraft:%s","rarity_value_mapper":"%s"}`, inputs[inputIndex], name, mapper)
			nodes[key], err = compiler.parse(json.RawMessage(expression), key)
			if err != nil {
				t.Fatal(err)
			}
		}
		if samplers[seed] == nil {
			samplers[seed] = NewDensitySampler(seed)
		}
		p := points[pointIndex]
		got, err := samplers[seed].Sample(nodes[key], p[0], p[1], p[2])
		if err != nil || math.Float64bits(got) != bits {
			t.Errorf("%s seed %d input %d point %d: got %d, want %d (%v)", mapper, seed, inputIndex, pointIndex, math.Float64bits(got), bits, err)
		}
		count++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 144 {
		t.Fatalf("got %d reference points, want 144", count)
	}
}

func TestSplineReference(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "spline_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	inputs := []float32{-2, -1.1, -0.8, -0.5, 0, 0.5, 0.75, 1, 2}
	point := func(location, value, derivative float32) DensitySplinePoint {
		v := float64(value)
		return DensitySplinePoint{Location: float64(location), Constant: &v, Derivative: float64(derivative)}
	}
	spline := &DensitySpline{Points: []DensitySplinePoint{
		point(-1.1, .044, 0), point(-.5, -.22, .4),
		point(.5, .3, -.3), point(1, .69, .4),
	}}
	sampler := NewDensitySampler(0)
	sc := bufio.NewScanner(f)
	count := 0
	for sc.Scan() {
		var label string
		var index int
		var bits uint32
		if _, err := fmt.Sscan(sc.Text(), &label, &index, &bits); err != nil || label != "spline" || index < 0 || index >= len(inputs) {
			t.Fatalf("invalid spline reference line %q: %v", sc.Text(), err)
		}
		spline.Coordinate = &DensityNode{Kind: "constant", Numbers: map[string]float64{"value": float64(inputs[index])}}
		got, err := sampler.sampleSpline(spline, 0, 0, 0)
		if err != nil || math.Float32bits(got) != bits {
			t.Errorf("input %d: got %d, want %d (%v)", index, math.Float32bits(got), bits, err)
		}
		count++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if count != len(inputs) {
		t.Fatalf("got %d spline reference points, want %d", count, len(inputs))
	}
}

func TestOverworldClimateDirectSampling(t *testing.T) {
	settings, err := LoadDensitySettings("overworld")
	if err != nil {
		t.Fatal(err)
	}
	sampler := NewDensitySampler(7122209284156038298)
	for _, field := range []string{"temperature", "vegetation", "continents", "erosion", "ridges", "depth"} {
		for _, p := range [][3]int32{{0, 64, 0}, {173, -32, -421}, {-10000, 120, 24567}} {
			v, err := sampler.Sample(settings.Router[field], p[0], p[1], p[2])
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				t.Errorf("%s at %v: %v (%v)", field, p, v, err)
			}
		}
	}
	// Continents are sampled in two dimensions; Y must not change the value.
	a, err := sampler.Sample(settings.Router["continents"], 173, -32, -421)
	if err != nil {
		t.Fatal(err)
	}
	b, err := sampler.Sample(settings.Router["continents"], 173, 240, -421)
	if err != nil || math.Float64bits(a) != math.Float64bits(b) {
		t.Fatalf("continents changed with Y: %v vs %v (%v)", a, b, err)
	}
	if _, err := sampler.Sample(settings.Router["final_density"], 0, 64, 0); err == nil {
		t.Fatal("interpolated final density should refuse direct sampling")
	}
}

func TestDensityRejectsUnknownReferenceAndCycle(t *testing.T) {
	c, err := newDensityCompiler()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.parse([]byte(`"minecraft:missing"`), "test"); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("missing reference: %v", err)
	}
	c.registry = &densityRegistry{Functions: map[string]json.RawMessage{
		"minecraft:a": json.RawMessage(`"minecraft:b"`),
		"minecraft:b": json.RawMessage(`"minecraft:a"`),
	}}
	if _, err := c.reference("minecraft:a"); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle: %v", err)
	}
}
