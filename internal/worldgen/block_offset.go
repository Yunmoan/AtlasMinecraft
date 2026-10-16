package worldgen

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed block_offset_1.21.1.txt
var blockOffsetText string

type blockOffsetRule struct {
	vertical                   bool
	maxHorizontal, maxVertical float32
}

var blockOffsetRules = func() map[string]blockOffsetRule {
	rules := make(map[string]blockOffsetRule)
	for _, line := range strings.Split(blockOffsetText, "\n") {
		if !strings.HasPrefix(line, "block_offset_rule ") {
			continue
		}
		var label, name string
		var rule blockOffsetRule
		if _, err := fmt.Sscan(line, &label, &name, &rule.vertical, &rule.maxHorizontal, &rule.maxVertical); err != nil {
			panic(err)
		}
		rules[name] = rule
	}
	if len(rules) != 33 {
		panic("incomplete Minecraft block offset rules")
	}
	return rules
}()

// BlockOffset implements the stock block state's position-dependent offset.
// Embedded collision boxes were probed at (0,0,0), so runtime callers apply
// offset(position)-offset(0) before translating those boxes.
func BlockOffset(name string, p FeaturePosition) [3]float64 {
	rule, ok := blockOffsetRules[name]
	if !ok {
		return [3]float64{}
	}
	seed := positionSeed(p.X, 0, p.Z)
	x := javaClamp((float64(float32(seed&15)/15)-0.5)*0.5, -float64(rule.maxHorizontal), float64(rule.maxHorizontal))
	z := javaClamp((float64(float32((seed>>8)&15)/15)-0.5)*0.5, -float64(rule.maxHorizontal), float64(rule.maxHorizontal))
	y := 0.0
	if rule.vertical {
		y = (float64(float32((seed>>4)&15)/15) - 1) * float64(rule.maxVertical)
	}
	return [3]float64{x, y, z}
}
