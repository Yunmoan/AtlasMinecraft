package worldgen

import (
	"atlas/internal/blockrules"
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestPortalPositionsAgainstVanillaJAR(t *testing.T) {
	sizes := [][2]float32{{.6, 1.8}, {1.4, .7}, {5, 6}}
	points := [][3]float64{{.5, 1, .5}, {.1, .2, .9}, {0, 0, 0}, {1.5, 2, .3}, {-1, 3, -1}}
	relatives := [][3]float64{{.5, 0, 0}, {.1, .7, .25}, {1, 1, -.4}}
	boxes := [][]blockrules.Box{{}, {{0, 0, 0, 1, 1, 1}}, {{0, 0, -2, 1, 3, 2}}, {{0, 0, 0, 1, .5, 1}}, {{-10, -10, -10, 10, 10, 10}}, {{-.8, 0, -2, -.2, 3, 2}, {.8, 0, -2, 1.2, 3, 2}}, {{0, 0, 0, 1, .5, 1}, {0, .5, .5, 1, 1, 1}}, {{-.2, 0, -2, .1999999, 3, 2}, {.8, 0, -2, 1.2, 3, 2}}}
	f, err := os.Open("../../docs/worldgen/portal_position_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	counts := map[string]int{}
	for scan.Scan() {
		line := scan.Text()
		parts := strings.Fields(line)
		switch parts[0] {
		case "portal_rectangle":
			var axis string
			var width, scenario int32
			var want PortalRectangle
			if _, err := fmt.Sscan(line, &parts[0], &axis, &width, &scenario, &want.Origin.X, &want.Origin.Y, &want.Origin.Z, &want.Width, &want.Height); err != nil {
				t.Fatal(err)
			}
			want.Axis = axis
			world := make(portalTestWorld)
			height := int32(5)
			if width == 21 {
				height = 21
			}
			base := FeaturePosition{-7, 30, 12}
			for x := int32(0); x < width; x++ {
				for y := int32(0); y < height; y++ {
					keep := true
					switch scenario {
					case 1:
						keep = x != width-1 || y < 3
					case 2:
						keep = x != 0 || y > 1
					case 3:
						keep = x != width/2 || y == 2
					case 4:
						keep = x != width/2 || y < 3
					}
					if keep {
						p := base
						p.Y += y
						if axis == "x" {
							p.X += x
						} else {
							p.Z += x
						}
						world[p] = SurfaceBlockState{Name: "minecraft:nether_portal", Properties: map[string]string{"axis": axis}}
					}
				}
			}
			origin := base
			origin.Y += 2
			if axis == "x" {
				origin.X += width / 2
			} else {
				origin.Z += width / 2
			}
			got, err := NetherPortalRectangle(world, origin)
			if err != nil || got != want {
				t.Fatalf("%s: got %+v %v", line, got, err)
			}
		case "portal_relative":
			var axis string
			var width int32
			var s, p int
			var want [3]float64
			if _, err := fmt.Sscan(line, &parts[0], &axis, &width, &s, &p, &want[0], &want[1], &want[2]); err != nil {
				t.Fatal(err)
			}
			height := int32(3)
			if width == 21 {
				height = 21
			}
			got := PortalRelativePosition(PortalRectangle{Axis: axis, Width: width, Height: height}, points[p], sizes[s][0], sizes[s][1], false)
			for i := range got {
				if math.Abs(got[i]-want[i]) > 1e-12 {
					t.Fatalf("%s: got %+v", line, got)
				}
			}
		case "portal_collision":
			var s, p, b int
			var want [3]float64
			if _, err := fmt.Sscan(line, &parts[0], &s, &p, &b, &want[0], &want[1], &want[2]); err != nil {
				t.Fatal(err)
			}
			got := FindCollisionFreePortalPosition(points[p], sizes[s][0], sizes[s][1], boxes[b])
			for i := range got {
				if math.Abs(got[i]-want[i]) > 1e-9 {
					t.Errorf("%s: got %+v", line, got)
					break
				}
			}
		case "portal_exit":
			var targetAxis, sourceAxis string
			var s, r int
			var want, velocity [3]float64
			var yaw float32
			if _, err := fmt.Sscan(line, &parts[0], &targetAxis, &sourceAxis, &s, &r, &want[0], &want[1], &want[2], &yaw, &velocity[0], &velocity[1], &velocity[2]); err != nil {
				t.Fatal(err)
			}
			got, gotYaw, gotVelocity := PortalExitPosition(PortalRectangle{Origin: FeaturePosition{-12, 70, 31}, Axis: targetAxis, Width: 2, Height: 3}, sourceAxis, relatives[r], sizes[s][0], sizes[s][1], 23, [3]float64{1, 2, 3})
			for i := range got {
				if math.Abs(got[i]-want[i]) > 1e-9 {
					t.Fatalf("%s: got %+v", line, got)
				}
			}
			if gotYaw != yaw || gotVelocity != velocity {
				t.Fatal("portal exit rotation/velocity differs", line, gotYaw, gotVelocity)
			}
		case "portal_collision_random":
			if len(parts) != 33 {
				t.Fatal("invalid collision random vector", line)
			}
			values := make([]float64, len(parts)-1)
			for i, text := range parts[1:] {
				v, err := strconv.ParseFloat(text, 64)
				if err != nil {
					t.Fatal(err)
				}
				values[i] = v
			}
			position := [3]float64{values[2], values[3], values[4]}
			var obstacles []blockrules.Box
			for b := 0; b < 4; b++ {
				var box blockrules.Box
				copy(box[:], values[5+b*6:11+b*6])
				obstacles = append(obstacles, box)
			}
			got := FindCollisionFreePortalPosition(position, float32(values[0]), float32(values[1]), obstacles)
			for i := range got {
				if math.Abs(got[i]-values[29+i]) > 1e-9 {
					t.Errorf("%s: got %+v", line, got)
					break
				}
			}
		default:
			t.Fatal(line)
		}
		counts[parts[0]]++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if counts["portal_rectangle"] != 30 || counts["portal_relative"] != 90 || counts["portal_exit"] != 36 || counts["portal_collision"] != 120 || counts["portal_collision_random"] != 256 {
		t.Fatal("incomplete portal position fixture", counts)
	}
}
