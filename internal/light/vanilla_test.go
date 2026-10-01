package light

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLightEngineAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/light_engine_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cases := 0
	cached := map[int]*Data{}
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if len(fields) != 4 || fields[0] != "light_case" {
			t.Fatal(scanner.Text())
		}
		scenario, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		if cached[scenario] == nil {
			engine, err := New(0, 32, true)
			if err != nil {
				t.Fatal(err)
			}
			stone := lightState(t, "stone", nil)
			torch := lightState(t, "torch", nil)
			glow := lightState(t, "glowstone", nil)
			slab := lightState(t, "stone_slab", nil)
			water := lightState(t, "water", nil)
			for cx := int32(-1); cx <= 1; cx++ {
				for cz := int32(-1); cz <= 1; cz++ {
					states := make([]uint16, 32*256)
					for y := 0; y < 32; y++ {
						for z := 0; z < 16; z++ {
							for x := 0; x < 16; x++ {
								wx, wz := int(cx)*16+x, int(cz)*16+z
								id := uint16(0)
								switch {
								case y == 0:
									id = stone
								case y == 12 && (scenario != 1 || wx != 8 || wz != 8):
									id = stone
								case scenario != 3 && wx == 15 && y == 4 && wz == 8:
									id = torch
								case wx == 8 && y == 4 && wz == 8:
									id = glow
								case wx == 10 && y == 4 && wz == 8:
									id = slab
								case scenario == 2 && wx >= 5 && wx <= 11 && wz >= 5 && wz <= 11 && y >= 13 && y <= 15:
									id = water
								}
								states[y*256+z*16+x] = id
							}
						}
					}
					if _, err = engine.Install(Position{cx, cz}, states, nil); err != nil {
						t.Fatal(err)
					}
				}
			}
			cached[scenario] = engine.Snapshot(Position{0, 0})
		}
		data := cached[scenario]
		values := make([]byte, 32*256)
		for y := 0; y < 32; y++ {
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					values[y*256+z*16+x] = level(data, fields[2] == "SKY", x, y, z, 0)
				}
			}
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(values))
		if digest != fields[3] {
			t.Errorf("scenario %d %s = %s, vanilla %s", scenario, fields[2], digest, fields[3])
		}
		cases++
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 8 {
		t.Fatal(cases)
	}
}
