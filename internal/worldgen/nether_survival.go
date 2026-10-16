package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
)

//go:embed fire_burnable_1.21.1.txt
var fireBurnableData string

// FireBlock.canBurn differs per state: waterlogged stairs/leaves/slabs do not
// burn, and Nether wood is not combustible. Extracted from the actual JAR.
var fireBurnableStates = func() []byte {
	f := strings.Fields(fireBurnableData)
	if len(f) != 3 || f[0] != "fire_burnable_states" || f[1] != "26684" {
		panic("invalid fire burnable states")
	}
	bits, err := base64.StdEncoding.DecodeString(f[2])
	if err != nil || len(bits) != (26684+7)/8 {
		panic("invalid fire burnable state bits")
	}
	return bits
}()

func featureFireSurvives(world FeatureBlockWorld, p FeaturePosition, below SurfaceBlockState) (bool, error) {
	id, err := blockstate.BlockStateID(below.Name, below.Properties)
	if err != nil {
		return false, err
	}
	rule, err := blockrules.Lookup(id)
	if err != nil {
		return false, err
	}
	if rule.FaceSturdy[1] {
		return true, nil
	}
	for d := range featureDirections {
		s, err := world.BlockAt(directionPosition(p, d))
		if err != nil {
			return false, err
		}
		id, err := blockstate.BlockStateID(s.Name, s.Properties)
		if err != nil {
			return false, err
		}
		if int(id) >= 26684 {
			return false, fmt.Errorf("invalid fire neighbor state %d", id)
		}
		if fireBurnableStates[id/8]&(1<<uint(id%8)) != 0 {
			return true, nil
		}
	}
	return false, nil
}

func featureNetherPlantSurvives(name string, below SurfaceBlockState) bool {
	ok := netherNylium(below) || below.Name == "minecraft:soul_soil" || PlantGroundSurvives(name, below.Name)
	if name == "minecraft:crimson_fungus" || name == "minecraft:warped_fungus" {
		ok = ok || below.Name == "minecraft:mycelium"
	}
	return ok
}
