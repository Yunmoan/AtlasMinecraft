package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

type portalTestWorld map[FeaturePosition]SurfaceBlockState

func (w portalTestWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, ok := w[p]; ok {
		return state, nil
	}
	return SurfaceBlockState{Name: "minecraft:air"}, nil
}
func (w portalTestWorld) SetBlock(p FeaturePosition, state SurfaceBlockState) error {
	w[p] = state
	return nil
}
func (w portalTestWorld) InWorld(p FeaturePosition) bool { return p.Y >= 0 && p.Y < 256 }

func TestNetherPortalShapeAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/portal_shape_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	count := 0
	for scan.Scan() {
		var label, axis string
		var width, height, scenario int32
		var valid, complete bool
		var origin FeaturePosition
		var want NetherPortalShape
		if _, err := fmt.Sscan(scan.Text(), &label, &axis, &width, &height, &scenario, &valid, &complete, &origin.X, &origin.Y, &origin.Z, &want.Width, &want.Height, &want.PortalBlocks); err != nil || label != "portal_shape" {
			t.Fatal(scan.Text(), err)
		}
		want.Origin, want.Axis = origin, axis
		world := make(portalTestWorld)
		base := FeaturePosition{-31, 20, 17}
		p := func(x, y int32) FeaturePosition {
			if axis == "x" {
				return treeOffset(base, x, y, 0)
			}
			return treeOffset(base, 0, y, x)
		}
		for x := int32(-1); x <= width; x++ {
			for y := int32(-1); y <= height; y++ {
				if x >= 0 && x < width && y >= 0 && y < height || (x == -1 || x == width) && (y == -1 || y == height) {
					continue
				}
				world[p(x, y)] = SurfaceBlockState{Name: "minecraft:obsidian"}
			}
		}
		for x := int32(0); x < width; x++ {
			for y := int32(0); y < height; y++ {
				name := "minecraft:air"
				if scenario == 6 {
					name = "minecraft:nether_portal"
				}
				if x == 0 && y == 0 {
					switch scenario {
					case 1:
						name = "minecraft:fire"
					case 2:
						name = "minecraft:soul_fire"
					case 5:
						name = "minecraft:nether_portal"
					case 7:
						name = "minecraft:stone"
					}
				}
				world[p(x, y)] = SurfaceBlockState{Name: name}
			}
		}
		if scenario == 3 {
			world[p(-1, 2)] = SurfaceBlockState{Name: "minecraft:air"}
		}
		if scenario == 4 {
			world[p(width/2, height)] = SurfaceBlockState{Name: "minecraft:air"}
		}
		got, err := InspectNetherPortalShape(world, 0, base, axis)
		if err != nil {
			t.Fatal(err)
		}
		if got.Valid() != valid || got.Complete() != complete || valid && got != want {
			t.Fatalf("%s: got %+v valid=%v complete=%v, want %+v valid=%v complete=%v", scan.Text(), got, got.Valid(), got.Complete(), want, valid, complete)
		}
		if valid && scenario != 6 {
			if err := FillNetherPortalShape(world, got); err != nil {
				t.Fatal(err)
			}
			filled, err := InspectNetherPortalShape(world, 0, base, axis)
			if err != nil || !filled.Complete() {
				t.Fatalf("filled portal incomplete: %+v %v", filled, err)
			}
			for x := int32(0); x < width; x++ {
				for y := int32(0); y < height; y++ {
					state, _ := world.BlockAt(p(x, y))
					if state.Name != "minecraft:nether_portal" || state.Properties["axis"] != axis {
						t.Fatal("portal fill lost axis")
					}
				}
			}
		}
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 400 {
		t.Fatalf("got %d shape cases", count)
	}
}
