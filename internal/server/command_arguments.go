package server

import (
	"atlas/internal/player"
	"atlas/internal/world"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func resolveCoordinates(source *player.Player, args []string, blockPosition bool) ([3]float64, error) {
	var origin, result [3]float64
	var yaw, pitch float32
	if source != nil {
		origin[0], origin[1], origin[2], yaw, pitch, _ = source.Pos()
	}
	if len(args) != 3 {
		return result, fmt.Errorf("Invalid coordinates: expected x y z")
	}
	local := strings.HasPrefix(args[0], "^") || strings.HasPrefix(args[1], "^") || strings.HasPrefix(args[2], "^")
	var values [3]float64
	for i, arg := range args {
		relative := strings.HasPrefix(arg, "~")
		if local && !strings.HasPrefix(arg, "^") {
			return result, fmt.Errorf("Invalid coordinates: ^ cannot be mixed with world coordinates")
		}
		number := arg
		if local || relative {
			number = arg[1:]
			if number == "" {
				number = "0"
			}
		}
		if blockPosition && !local && !relative && strings.ContainsAny(number, ".eE") {
			return result, fmt.Errorf("Invalid block coordinates: absolute coordinates must be integers")
		}
		value, err := strconv.ParseFloat(number, 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return result, fmt.Errorf("Invalid coordinates")
		}
		values[i] = value
		if !local {
			result[i] = value
			if relative {
				result[i] += origin[i]
			} else if !blockPosition && i != 1 && !strings.Contains(number, ".") {
				result[i] += .5
			}
		}
	}
	if local {
		y, p := float64(yaw)*math.Pi/180, float64(pitch)*math.Pi/180
		forward := [3]float64{-math.Sin(y) * math.Cos(p), -math.Sin(p), math.Cos(y) * math.Cos(p)}
		left := [3]float64{math.Cos(y), 0, math.Sin(y)}
		up := [3]float64{-math.Sin(y) * math.Sin(p), math.Cos(p), math.Cos(y) * math.Sin(p)}
		for i := range result {
			result[i] = origin[i] + left[i]*values[0] + up[i]*values[1] + forward[i]*values[2]
		}
	}
	for i, v := range result {
		if blockPosition {
			v = math.Floor(v)
			result[i] = v
		}
	}
	return result, validateCommandCoordinates(result, blockPosition)
}

func validateCommandCoordinates(position [3]float64, blockPosition bool) error {
	for _, v := range position {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("Invalid coordinates")
		}
	}
	// Vanilla TeleportCommand checks Level.isInSpawnableBounds(BlockPos).
	if position[0] < -30000000 || position[0] >= 30000000 || position[2] < -30000000 || position[2] >= 30000000 {
		return fmt.Errorf("Invalid coordinates: X/Z must be >= -30000000 and < 30000000")
	}
	if blockPosition {
		if position[1] < world.MinY || position[1] >= world.MaxY {
			return fmt.Errorf("Invalid coordinates: block Y must be >= %d and < %d", world.MinY, world.MaxY)
		}
	} else if position[1] < -20000000 || position[1] >= 20000000 {
		return fmt.Errorf("Invalid coordinates: teleport Y must be >= -20000000 and < 20000000")
	}
	return nil
}

func (s *Server) resolveCommandPlayer(source *player.Player, name string) *player.Player {
	if name == "@s" {
		return source
	}
	if name == "@p" {
		var nearest *player.Player
		best := math.Inf(1)
		var sx, sy, sz float64
		if source != nil {
			sx, sy, sz, _, _, _ = source.Pos()
		}
		for _, p := range s.Hub.Snapshot() {
			x, y, z, _, _, _ := p.Pos()
			d := (x-sx)*(x-sx) + (y-sy)*(y-sy) + (z-sz)*(z-sz)
			if d < best {
				nearest, best = p, d
			}
		}
		return nearest
	}
	return s.Hub.FindName(name)
}

// parseCommandAngle mirrors vanilla AngleArgument: an optional "~" prefix
// makes the angle relative to the source's yaw, empty relative values are 0,
// and the result wraps to vanilla's [-180, 180) degree range.
func parseCommandAngle(source *player.Player, value string) (float32, error) {
	relative := strings.HasPrefix(value, "~")
	number := strings.TrimPrefix(value, "~")
	if number == "" {
		number = "0"
	}
	angle, err := strconv.ParseFloat(number, 64)
	if err != nil || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return 0, fmt.Errorf("Invalid angle")
	}
	if relative && source != nil {
		_, _, _, yaw, _, _ := source.Pos()
		angle += float64(yaw)
	}
	return wrapCommandDegrees(angle), nil
}

// wrapCommandDegrees matches vanilla Mth.wrapDegrees.
func wrapCommandDegrees(degrees float64) float32 {
	wrapped := math.Mod(degrees, 360)
	if wrapped >= 180 {
		wrapped -= 360
	}
	if wrapped < -180 {
		wrapped += 360
	}
	return float32(wrapped)
}

func parseCommandTime(value string) (int64, error) {
	factor := float64(1)
	if len(value) > 0 {
		switch value[len(value)-1] {
		case 'd':
			factor = 24000
			value = value[:len(value)-1]
		case 's':
			factor = 20
			value = value[:len(value)-1]
		case 't':
			value = value[:len(value)-1]
		}
	}
	number, err := strconv.ParseFloat(value, 64)
	ticks := math.Round(number * factor)
	if err != nil || math.IsNaN(ticks) || math.IsInf(ticks, 0) || number < 0 || ticks > math.MaxInt32 {
		return 0, fmt.Errorf("Invalid time")
	}
	return int64(ticks), nil
}

func parseCommandBlockState(value string) (uint32, error) {
	if numeric, err := strconv.ParseUint(value, 10, 32); err == nil {
		if numeric > uint64(world.MaxBlockState) {
			return 0, fmt.Errorf("Unknown block state ID")
		}
		return uint32(numeric), nil
	}
	name := value
	properties := map[string]string{}
	if opening := strings.IndexByte(value, '['); opening >= 0 {
		if !strings.HasSuffix(value, "]") {
			return 0, fmt.Errorf("Invalid block state properties")
		}
		name = value[:opening]
		for _, property := range strings.Split(value[opening+1:len(value)-1], ",") {
			if property == "" {
				continue
			}
			key, val, ok := strings.Cut(property, "=")
			if !ok || key == "" || val == "" {
				return 0, fmt.Errorf("Invalid block state properties")
			}
			if _, duplicate := properties[key]; duplicate {
				return 0, fmt.Errorf("Invalid block state: duplicate property %s", key)
			}
			properties[key] = val
		}
	}
	return world.BlockStateID(name, properties)
}
