package player

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestPortalProcessorAndPlayerCooldownAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/portal_processor_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var p *Player
	count, cooldownCount := 0, 0
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		f := strings.Fields(scan.Text())
		number := func(i int) int32 {
			v, e := strconv.ParseInt(f[i], 10, 32)
			if e != nil {
				t.Fatal(e)
			}
			return int32(v)
		}
		boolean := func(i int) bool {
			v, e := strconv.ParseBool(f[i])
			if e != nil {
				t.Fatal(e)
			}
			return v
		}
		switch f[0] {
		case "portal_processor":
			if number(3) == 0 {
				p = New("Oracle", [16]byte{}, 1, nil)
			}
			p.teleportPending = !boolean(5)
			actual := p.AdvancePortalTick(boolean(4), 1, number(1))
			if actual != boolean(6) || p.portalDuration != number(7) {
				t.Fatalf("%s: trigger=%v duration=%d", scan.Text(), actual, p.portalDuration)
			}
			count++
		case "portal_player_cooldown":
			if number(2) == 0 {
				p = New("Oracle", [16]byte{}, 1, nil)
				p.SetPortalCooldown(10)
			}
			p.AdvancePortalTick(boolean(3), 1, 1)
			if p.PortalCooldown() != number(4) {
				t.Fatalf("%s: cooldown=%d", scan.Text(), p.PortalCooldown())
			}
			cooldownCount++
		}
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 1080 || cooldownCount != 30 {
		t.Fatal("incomplete portal vectors", count, cooldownCount)
	}
}
