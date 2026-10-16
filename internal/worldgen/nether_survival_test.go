package worldgen

import (
	"atlas/internal/blockstate"
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

type netherSurvivalProbeWorld struct{ ground, neighbor SurfaceBlockState }

func (w netherSurvivalProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if p == (FeaturePosition{-1, 40, 15}) {
		return w.ground, nil
	}
	if p == (FeaturePosition{0, 41, 15}) {
		return w.neighbor, nil
	}
	return SurfaceBlockState{Name: "minecraft:air"}, nil
}
func (netherSurvivalProbeWorld) SetBlock(FeaturePosition, SurfaceBlockState) error {
	panic("unexpected survival write")
}
func (netherSurvivalProbeWorld) InWorld(FeaturePosition) bool { return true }

func TestNetherPlantSurvivalAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/nether_survival_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	state := func(s string) SurfaceBlockState {
		id, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		v, err := blockstate.BlockStateByID(uint32(id))
		if err != nil {
			t.Fatal(err)
		}
		return SurfaceBlockState{Name: v.Name, Properties: v.Properties}
	}
	cases, positive := 0, 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if f[0] == "fire_burnable_states" {
			continue
		}
		if len(f) != 5 || f[0] != "nether_survival" {
			t.Fatal(scanner.Text())
		}
		world := netherSurvivalProbeWorld{state(f[2]), state(f[3])}
		s := SurfaceBlockState{Name: f[1]}
		if err := validateVegetationState(s); err != nil {
			t.Fatal(err)
		}
		got, err := featurePlantSurvives(world, FeaturePosition{-1, 41, 15}, s)
		if err != nil {
			t.Fatal(err)
		}
		if got != (f[4] == "1") {
			t.Errorf("%s got survive=%v", scanner.Text(), got)
		}
		cases++
		if got {
			positive++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 7098 {
		t.Fatalf("got %d cases, want7098", cases)
	}
	t.Logf("matched %d actual JAR canSurvive cases, %d positive", cases, positive)
}
