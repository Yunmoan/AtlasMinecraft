package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

func (s *Server) isOperator(p *player.Player) bool {
	return s.operatorLevel(p) >= 2
}

func (s *Server) operatorLevel(p *player.Player) int {
	if p == nil || s.Cfg.Ops[strings.ToLower(p.Name)] {
		return 4
	}
	if s.Cfg.Operators != nil {
		return s.Cfg.Operators.LevelForPlayer(p.Name, p.UUID)
	}
	return 0
}

func (s *Server) sendTime(p *player.Player) {
	age, day := s.World.Time()
	p.SendPacket(proto.SUpdateTime, proto.NewWriter().I64(age).I64(day).Payload())
}

func weatherPacket(reason byte, value float32) []byte {
	return proto.NewWriter().U8(reason).F32(value).Payload()
}

func (s *Server) sendWeather(p *player.Player) {
	weather := s.World.Weather()
	if p.Dimension() != "minecraft:overworld" {
		weather = "clear"
	}
	if weather == "clear" {
		p.SendPacket(proto.SGameStateChange, weatherPacket(2, 0))
	} else {
		p.SendPacket(proto.SGameStateChange, weatherPacket(1, 0))
	}
	rain, thunder := float32(0), float32(0)
	if weather != "clear" {
		rain = 1
	}
	if weather == "thunder" {
		thunder = 1
	}
	p.SendPacket(proto.SGameStateChange, weatherPacket(7, rain))
	p.SendPacket(proto.SGameStateChange, weatherPacket(8, thunder))
}

func (s *Server) sendWorldState(p *player.Player) {
	s.sendTime(p)
	s.sendWeather(p)
}

func (s *Server) broadcastWorldState() {
	s.Hub.Each(func(p *player.Player) { s.sendWorldState(p) })
}

// RunCommand executes a client command and sends the result back to its source.
func (s *Server) RunCommand(p *player.Player, line string) {
	slog.Info(p.Name + " issued server command: /" + strings.TrimPrefix(strings.TrimSpace(line), "/"))
	response := s.executeCommand(p, line)
	p.SendPacket(proto.SSystemChat, commandChat(response))
	slog.Debug("command result", "source", p.Name, "result", response)
}

// RunConsoleCommand has operator privileges and reports its result in logs.
func (s *Server) RunConsoleCommand(line string) {
	slog.Info("Console issued server command: /" + strings.TrimPrefix(strings.TrimSpace(line), "/"))
	response := s.executeCommand(nil, line)
	for _, resultLine := range strings.Split(response, "\n") {
		slog.Info(resultLine)
	}
}

func (s *Server) executeCommand(source *player.Player, line string) string {
	fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/"))
	if len(fields) == 0 {
		return "Usage: /help"
	}
	cmd, args := strings.ToLower(fields[0]), fields[1:]
	spec, known := findCommand(cmd)
	if !known {
		return "Unknown command. Use /help."
	}
	if s.operatorLevel(source) < spec.level {
		return "You do not have permission to use this command."
	}
	enteredCommand := cmd
	cmd = spec.name
	switch cmd {
	case "help":
		return commandHelp(s.operatorLevel(source), args)
	case "execute":
		return s.commandExecuteIn(source, args)
	case "tps", "mspt", "status", "sysinfo":
		if len(args) != 0 {
			return "Usage: /" + enteredCommand
		}
		return s.commandPerformance(cmd)
	case "list":
		if len(args) != 0 {
			return "Usage: /list"
		}
		names := s.Hub.Names()
		return fmt.Sprintf("Players online (%d/%d): %s", len(names), s.Cfg.MaxPlayers, strings.Join(names, ", "))
	case "where":
		if len(args) != 0 {
			return "Usage: /where"
		}
		if source == nil {
			return "Usage: /tp <player> <x> <y> <z>"
		}
		x, y, z, _, _, _ := source.Pos()
		return fmt.Sprintf("Position: %.2f %.2f %.2f (overworld)", x, y, z)
	case "seed":
		if len(args) != 0 {
			return "Usage: /seed"
		}
		return fmt.Sprintf("Seed: %d", s.World.Seed())
	case "op", "deop":
		if len(args) != 1 || s.Cfg.Operators == nil {
			return "Usage: /" + cmd + " <player>"
		}
		var changed bool
		var err error
		if cmd == "op" {
			if s.Cfg.OnlineMode {
				target := s.Hub.FindName(args[0])
				if target == nil {
					return "In online mode, the player must be connected before /op can verify their UUID."
				}
				changed, err = s.Cfg.Operators.AddProfile(target.Name, target.UUID)
			} else {
				changed, err = s.Cfg.Operators.Add(args[0])
			}
		} else {
			changed, err = s.Cfg.Operators.Remove(args[0])
		}
		if err != nil {
			return "Operator list update failed: " + err.Error()
		}
		if !changed {
			if cmd == "op" {
				return args[0] + " is already an operator"
			}
			return args[0] + " is not an operator"
		}
		if target := s.Hub.FindName(args[0]); target != nil {
			target.SendPacket(proto.SDeclareCommands, declareCommandsPayloadForLevel(s.operatorLevel(target)))
		}
		if cmd == "op" {
			return "Made " + args[0] + " a server operator"
		}
		return "Removed " + args[0] + " from server operators"
	case "time":
		return s.commandTime(args)
	case "weather":
		if len(args) != 1 {
			return "Usage: /weather <clear|rain|thunder>"
		}
		if err := s.World.SetWeather(strings.ToLower(args[0])); err != nil {
			return err.Error()
		}
		s.broadcastWorldState()
		if err := s.SaveWorlds(); err != nil {
			return "Weather changed, but saving failed: " + err.Error()
		}
		return "Weather set to " + s.World.Weather()
	case "tp":
		return s.commandTeleport(source, args)
	case "setworldspawn":
		return s.commandSetWorldSpawn(source, args)
	case "spawnpoint":
		return s.commandSpawnpoint(source, args)
	case "setblock":
		return s.commandSetBlock(source, args)
	case "save":
		if len(args) > 0 && !(enteredCommand != "save" && len(args) == 1 && args[0] == "flush") {
			return "Usage: /save-all [flush]"
		}
		s.SavePlayers()
		if err := s.World.Save(); err != nil {
			return "World save failed: " + err.Error()
		}
		return "World saved."
	default:
		return "Unknown command. Use /help."
	}
}

func (s *Server) commandTime(args []string) string {
	usage := "Usage: /time query [daytime|gametime|day] | /time set <day|noon|night|midnight|time> | /time add <time>"
	if (len(args) == 1 || len(args) == 2) && args[0] == "query" {
		age, day := s.World.Time()
		if len(args) == 2 {
			switch args[1] {
			case "daytime":
				return fmt.Sprintf("Daytime: %d ticks", day%24000)
			case "gametime":
				return fmt.Sprintf("Game time: %d ticks", age)
			case "day":
				return fmt.Sprintf("Day: %d", day/24000)
			default:
				return usage
			}
		}
		return fmt.Sprintf("World age: %d ticks; daytime: %d ticks", age, day%24000)
	}
	if len(args) != 2 {
		return usage
	}
	var ticks int64
	var err error
	switch args[1] {
	case "day":
		ticks = 1000
	case "noon":
		ticks = 6000
	case "night":
		ticks = 13000
	case "midnight":
		ticks = 18000
	default:
		ticks, err = parseCommandTime(args[1])
		if err != nil {
			return usage
		}
	}
	switch args[0] {
	case "set":
		s.World.SetDayTime(ticks)
	case "add":
		if args[1] == "day" || args[1] == "noon" || args[1] == "night" || args[1] == "midnight" {
			return usage
		}
		s.World.AdvanceTime(ticks)
	default:
		return usage
	}
	s.Hub.Each(func(p *player.Player) { s.sendTime(p) })
	if err := s.World.Save(); err != nil {
		return "Time changed, but saving failed: " + err.Error()
	}
	_, day := s.World.Time()
	if args[0] == "add" {
		return fmt.Sprintf("Added %d ticks; daytime: %d ticks", ticks, day%24000)
	}
	return fmt.Sprintf("Time set to %d ticks", day)
}

func (s *Server) commandTeleport(source *player.Player, args []string) string {
	return s.commandTeleportIn(source, args, nil)
}

func (s *Server) commandTeleportIn(source *player.Player, args []string, destinationWorld *world.World) string {
	usage := "Usage: /tp [player] <destination player|x y z>"
	var target *player.Player
	var xyz [3]float64
	var err error
	var destinationRotation *[2]float32
	if len(args) == 1 || len(args) == 2 {
		target = source
		if len(args) == 2 {
			target = s.resolveCommandPlayer(source, args[0])
		}
		destination := s.resolveCommandPlayer(source, args[len(args)-1])
		if target == nil || destination == nil {
			return "Player not found; use /list."
		}
		var yaw, pitch float32
		xyz[0], xyz[1], xyz[2], yaw, pitch, _ = destination.Pos()
		destinationRotation = &[2]float32{yaw, pitch}
		destinationWorld = s.PlayerWorld(destination)
		err = validateCommandCoordinates(xyz, false)
	} else if len(args) == 3 {
		target = source
		xyz, err = resolveCoordinates(source, args, false)
	} else if len(args) == 4 {
		target = s.resolveCommandPlayer(source, args[0])
		args = args[1:]
		xyz, err = resolveCoordinates(source, args, false)
	} else {
		return usage
	}
	if target == nil {
		return "Player not found; use /list."
	}
	if err != nil {
		return err.Error()
	}
	if destinationWorld == nil {
		destinationWorld = s.PlayerWorld(target)
	}
	crossDimension := destinationWorld.DimensionName() != target.Dimension()
	if err := s.teleportToDimension(target, destinationWorld.DimensionName(), xyz, destinationRotation); err != nil {
		return "Teleport destination unavailable: " + err.Error()
	}
	if crossDimension {
		return fmt.Sprintf("Teleported %s to %.2f %.2f %.2f in %s", target.Name, xyz[0], xyz[1], xyz[2], destinationWorld.DimensionName())
	}
	return fmt.Sprintf("Teleported %s to %.2f %.2f %.2f", target.Name, xyz[0], xyz[1], xyz[2])
}

// formatCommandAngle renders a spawn angle like vanilla's float formatting,
// keeping at least one decimal place.
func formatCommandAngle(angle float32) string {
	formatted := strconv.FormatFloat(float64(angle), 'f', -1, 32)
	if !strings.ContainsAny(formatted, ".eE") {
		formatted += ".0"
	}
	return formatted
}

// commandSetWorldSpawn implements /setworldspawn [<pos> [<angle>]]. Like
// vanilla, the position defaults to the source (origin on console) and the
// angle to 0. The stored spawn is broadcast to every player in play.
func (s *Server) commandSetWorldSpawn(source *player.Player, args []string) string {
	usage := "Usage: /setworldspawn [x y z [angle]]"
	var xyz [3]float64
	var angle float32
	var err error
	switch len(args) {
	case 0:
		xyz, err = resolveCoordinates(source, []string{"~", "~", "~"}, true)
	case 3:
		xyz, err = resolveCoordinates(source, args, true)
	case 4:
		xyz, err = resolveCoordinates(source, args[:3], true)
		if err == nil {
			angle, err = parseCommandAngle(source, args[3])
		}
	default:
		return usage
	}
	if err != nil {
		return err.Error()
	}
	x, y, z := int32(xyz[0]), int32(xyz[1]), int32(xyz[2])
	changed, err := s.World.SetSpawn(x, y, z, angle)
	if err != nil {
		return "Spawn update failed: " + err.Error()
	}
	if changed {
		payload := proto.NewWriter().I64(proto.PackPosition(x, y, z)).F32(angle).Payload()
		s.Hub.Each(func(p *player.Player) {
			if p.PlayReady() && p.Dimension() == "minecraft:overworld" {
				p.SendPacket(proto.SSpawnPosition, payload)
			}
		})
	}
	// Vanilla keeps the spawn neighborhood loaded via a region ticket;
	// preparing it here covers joining players until the next restart.
	if err := s.World.PrepareDestination(context.Background(), x>>4, z>>4); err != nil {
		return "Spawn point set, but preparing the spawn area failed: " + err.Error()
	}
	return fmt.Sprintf("Set the world spawn point to %d, %d, %d [%s]", x, y, z, formatCommandAngle(angle))
}

// commandSpawnpoint implements /spawnpoint [<player>] [<x y z> [<angle>]].
// As in vanilla, an omitted position uses the source's current block
// position (not the target's), and the stored point is forced.
func (s *Server) commandSpawnpoint(source *player.Player, args []string) string {
	usage := "Usage: /spawnpoint [player] [x y z [angle]]"
	var target *player.Player
	var xyz [3]float64
	var angle float32
	var err error
	here := func() ([3]float64, error) {
		return resolveCoordinates(source, []string{"~", "~", "~"}, true)
	}
	switch len(args) {
	case 0:
		target = source
		if target == nil {
			return "Usage: /spawnpoint <player> [x y z [angle]]"
		}
		xyz, err = here()
	case 1:
		target = s.resolveCommandPlayer(source, args[0])
		if target == nil {
			return "Player not found; use /list."
		}
		xyz, err = here()
	case 3:
		target = source
		if target == nil {
			return "Usage: /spawnpoint <player> [x y z [angle]]"
		}
		xyz, err = resolveCoordinates(source, args, true)
	case 4:
		target = s.resolveCommandPlayer(source, args[0])
		if target == nil {
			return "Player not found; use /list."
		}
		xyz, err = resolveCoordinates(source, args[1:], true)
	case 5:
		target = s.resolveCommandPlayer(source, args[0])
		if target == nil {
			return "Player not found; use /list."
		}
		xyz, err = resolveCoordinates(source, args[1:4], true)
		if err == nil {
			angle, err = parseCommandAngle(source, args[4])
		}
	default:
		return usage
	}
	if err != nil {
		return err.Error()
	}
	x, y, z := int32(xyz[0]), int32(xyz[1]), int32(xyz[2])
	spawnWorld := s.PlayerWorld(source)
	if err := spawnWorld.SetPlayerSpawn(target.UUID, x, y, z, angle); err != nil {
		return "Spawn point update failed: " + err.Error()
	}
	return fmt.Sprintf("Set spawn point to %d, %d, %d [%s] in %s for %s",
		x, y, z, formatCommandAngle(angle), spawnWorld.DimensionName(), target.Name)
}

func (s *Server) commandSetBlock(source *player.Player, args []string) string {
	return s.commandSetBlockIn(source, args, s.PlayerWorld(source))
}

func (s *Server) commandSetBlockIn(source *player.Player, args []string, targetWorld *world.World) string {
	if len(args) != 4 && len(args) != 5 {
		return "Usage: /setblock <x y z> <block[state=value]|state_id> [replace|keep]"
	}
	position, err := resolveCoordinates(source, args[:3], true)
	if err != nil {
		return err.Error()
	}
	xyz := [3]int32{int32(position[0]), int32(position[1]), int32(position[2])}
	state, err := parseCommandBlockState(args[3])
	if err != nil {
		return "Unknown block state: " + err.Error()
	}
	mode := "replace"
	if len(args) == 5 {
		mode = args[4]
	}
	if mode == "keep" {
		err = targetWorld.TryPlaceBlock(world.BlockPosition{X: xyz[0], Y: xyz[1], Z: xyz[2]}, world.StateAir, state, nil)
	} else if mode == "replace" {
		err = targetWorld.SetBlock(xyz[0], xyz[1], xyz[2], state)
	} else {
		return "Usage: /setblock <x y z> <block[state=value]|state_id> [replace|keep]"
	}
	if err != nil {
		return "Block edit failed: " + err.Error()
	}
	return fmt.Sprintf("Set block at %d %d %d to state %d", xyz[0], xyz[1], xyz[2], state)
}
