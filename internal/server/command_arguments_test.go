package server

import (
	"atlas/internal/player"
	"atlas/internal/world"
	"math"
	"strings"
	"testing"
)

func TestCommandCoordinatesMatchVanillaRelativeLocalAndBlockCentering(t *testing.T) {
	p := player.New("Reference", [16]byte{}, 1, nil)
	p.SetPos(-.25, 64.75, .5, 90, 0, true)
	for _, test := range []struct {
		args  []string
		block bool
		want  [3]float64
	}{
		{[]string{"~", "~1", "~-2"}, false, [3]float64{-.25, 65.75, -1.5}},
		{[]string{"10", "64", "-10"}, false, [3]float64{10.5, 64, -9.5}},
		{[]string{"10.0", "64", "-10.0"}, false, [3]float64{10, 64, -10}},
		{[]string{"^2", "^1", "^3"}, false, [3]float64{-3.25, 65.75, 2.5}},
		{[]string{"~", "~-1", "~"}, true, [3]float64{-1, 63, 0}},
		{[]string{"10000000", "1024", "-10000000"}, false, [3]float64{10000000.5, 1024, -9999999.5}},
		{[]string{"-30000000.0", "-20000000", "-30000000.0"}, false, [3]float64{-30000000, -20000000, -30000000}},
		{[]string{"29999999.75", "19999999.75", "29999999.75"}, false, [3]float64{29999999.75, 19999999.75, 29999999.75}},
	} {
		got, err := resolveCoordinates(p, test.args, test.block)
		if err != nil {
			t.Fatal(err)
		}
		for i := range got {
			if math.Abs(got[i]-test.want[i]) > 1e-8 {
				t.Fatalf("%v got %v want %v", test.args, got, test.want)
			}
		}
	}
	for _, args := range [][]string{{"^", "~", "^"}, {"NaN", "0", "0"}, {"Inf", "0", "0"}, {"30000001", "0", "0"}, {"30000000.0", "0", "0"}, {"-30000000.01", "0", "0"}, {"0", "20000000", "0"}, {"0", "-20000000.01", "0"}} {
		if _, err := resolveCoordinates(p, args, false); err == nil {
			t.Fatalf("invalid coordinates accepted: %v", args)
		}
	}
}

func TestCommandBlockStatesTimeAliasesAndKeepBehavior(t *testing.T) {
	s, p := interactionServer(t)
	s.Cfg.Ops = map[string]bool{strings.ToLower(p.Name): true}
	if got := s.executeCommand(p, "minecraft:setblock ~ ~ ~ minecraft:oak_log[axis=x]"); !strings.HasPrefix(got, "Set block at 0 -60 0") {
		t.Fatal(got)
	}
	state, err := s.World.BlockAt(0, -60, 0)
	if err != nil {
		t.Fatal(err)
	}
	description, err := world.BlockStateByID(state)
	if err != nil || description.Name != "minecraft:oak_log" || description.Properties["axis"] != "x" {
		t.Fatalf("state properties lost: %+v %v", description, err)
	}
	if got := s.executeCommand(p, "setblock ~ ~ ~ stone keep"); !strings.HasPrefix(got, "Block edit failed:") {
		t.Fatal(got)
	}
	if current, _ := s.World.BlockAt(0, -60, 0); current != state {
		t.Fatal("keep overwrote existing block")
	}
	if got := s.executeCommand(p, "setblock ~1 ~ ~ stone keep"); !strings.HasPrefix(got, "Set block at") {
		t.Fatal(got)
	}
	if got := s.executeCommand(p, "minecraft:time set 1d"); !strings.Contains(got, "24000") {
		t.Fatal(got)
	}
	if got := s.executeCommand(p, "time add 0.5s"); !strings.Contains(got, "Added 10 ticks") {
		t.Fatal(got)
	}
	if _, day := s.World.Time(); day != 24010 {
		t.Fatalf("time units did not update saved daytime: %d", day)
	}
	if got := s.executeCommand(p, "time query daytime"); got != "Daytime: 10 ticks" {
		t.Fatal(got)
	}
	if got := s.executeCommand(p, "time query day"); got != "Day: 1" {
		t.Fatal(got)
	}
	for _, value := range []string{"NaN", "Inf", "-1", "2147483648", "1x"} {
		if _, err := parseCommandTime(value); err == nil {
			t.Fatalf("invalid time accepted: %s", value)
		}
	}
	for _, value := range []string{"minecraft:stone[axis=x]", "oak_log[axis=q]", "oak_log[axis=x,axis=y]", "26684"} {
		if _, err := parseCommandBlockState(value); err == nil {
			t.Fatalf("invalid state accepted: %s", value)
		}
	}
	for _, command := range []string{"list extra", "where extra", "seed extra", "time query invalid", "save-all bad"} {
		if got := s.executeCommand(p, command); !strings.HasPrefix(got, "Usage:") {
			t.Fatalf("invalid arguments accepted: %s: %s", command, got)
		}
	}
}
