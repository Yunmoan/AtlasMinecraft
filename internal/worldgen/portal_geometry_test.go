package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type portalPlatformProbe struct {
	base   string
	writes []string
}

func (w *portalPlatformProbe) BlockAt(FeaturePosition) (SurfaceBlockState, error) {
	return SurfaceBlockState{Name: "minecraft:" + w.base}, nil
}
func (w *portalPlatformProbe) InWorld(p FeaturePosition) bool { return p.Y >= 0 && p.Y < 256 }
func (w *portalPlatformProbe) SetBlock(p FeaturePosition, s SurfaceBlockState) error {
	w.writes = append(w.writes, fmt.Sprintf("end_platform %s %d %d %d %s", w.base, p.X, p.Y, p.Z, strings.TrimPrefix(s.Name, "minecraft:")))
	return nil
}

func TestPortalGeometryAgainstVanillaJAR(t *testing.T) {
	f, err := os.Open("../../docs/worldgen/portal_geometry_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	borders := [][3]float64{{0, 0, 59999968}, {-17.25, 32.75, 64.5}, {29999980, -29999980, 64}, {0.5, -0.5, 1}}
	points := [][3]float64{{0, 64, 0}, {-1.5, 63.9, -0.01}, {128.9, -63.01, -128.9}, {29999999, 255.5, -29999999}, {-29999999, 400, 29999999}, {-17.25, 70, 32.75}, {-17.251, 70, 64}, {16, 70, -16}}
	scales := [][2]float64{{1, 8}, {8, 1}, {1, 1}}
	writes := map[string][]string{}
	scan := bufio.NewScanner(f)
	targets, arrivals := 0, 0
	for scan.Scan() {
		line := scan.Text()
		parts := strings.Fields(line)
		switch parts[0] {
		case "portal_target":
			var b, s, p int
			var want FeaturePosition
			if _, err := fmt.Sscan(line, &parts[0], &b, &s, &p, &want.X, &want.Y, &want.Z); err != nil {
				t.Fatal(err)
			}
			border, err := NewPortalBorder(borders[b][0], borders[b][1], borders[b][2])
			if err != nil {
				t.Fatal(err)
			}
			got, err := NetherPortalSearchTarget(border, points[p][0], points[p][1], points[p][2], scales[s][0], scales[s][1])
			if err != nil || got != want {
				t.Fatalf("%s: got %+v %v", line, got, err)
			}
			targets++
		case "end_arrival":
			var want FeaturePosition
			var x, y, z, yaw float64
			if _, err := fmt.Sscan(line, &parts[0], &want.X, &want.Y, &want.Z, &x, &y, &z, &yaw); err != nil {
				t.Fatal(err)
			}
			spawn := EndSpawnPoint()
			if want != spawn || x != float64(spawn.X)+0.5 || y != float64(spawn.Y)-1 || z != float64(spawn.Z)+0.5 || yaw != 90 {
				t.Fatal("End arrival differs", line)
			}
			arrivals++
		case "end_platform":
			writes[parts[1]] = append(writes[parts[1]], line)
		default:
			t.Fatal(line)
		}
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if targets != 96 || arrivals != 1 || len(writes) != 3 {
		t.Fatal("incomplete portal geometry fixture", targets, arrivals, len(writes))
	}
	for base, want := range writes {
		world := &portalPlatformProbe{base: base}
		origin := EndSpawnPoint()
		origin.Y--
		if err := PlaceEndPlatform(world, origin); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(world.writes, want) {
			t.Fatalf("%s platform writes differ", base)
		}
	}
}
