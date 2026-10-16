package worldgen

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestScatteredOreAgainstVanillaJAR(t *testing.T) {
	compiler := func(name string) (configuredFeaturePlacer, error) {
		if !strings.Contains(name, "scatter_custom_") {
			def, err := LoadConfiguredFeatureDefinition(name)
			if err != nil {
				return nil, err
			}
			return compileScatteredOrePlan(name, def)
		}
		variant, err := strconv.Atoi(strings.TrimPrefix(name, "minecraft:scatter_custom_"))
		if err != nil {
			return nil, err
		}
		size := []int{0, 1, 8, 16, 64, 64, 64, 8}[variant]
		chance := []float32{0, 0, .25, .5, 1, .5, .5, 0}[variant]
		base, output := "netherrack", "basalt"
		if variant == 6 {
			base = "stone"
		}
		if variant == 7 {
			output = "netherrack"
		}
		target := func(base, output string) string {
			return fmt.Sprintf(`{"target":{"predicate_type":"minecraft:block_match","block":"minecraft:%s"},"state":{"Name":"minecraft:%s"}}`, base, output)
		}
		targets := target(base, output)
		if variant == 5 || variant == 6 {
			targets += "," + target("netherrack", "blackstone")
		}
		def := ConfiguredFeatureDefinition{Type: "minecraft:scattered_ore", Config: json.RawMessage(fmt.Sprintf(`{"size":%d,"discard_chance_on_air_exposure":%g,"targets":[%s]}`, size, chance, targets))}
		return compileScatteredOrePlan(name, def)
	}
	testNetherDecorationVectors(t, "scattered_ore_1.21.1.txt", "scattered_ore", compiler, 2240, 2240)
}
