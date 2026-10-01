package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"context"
	"fmt"
	"log/slog"
	"math"
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
		if err := s.World.Save(); err != nil {
			return "Weather changed, but saving failed: " + err.Error()
		}
		return "Weather set to " + s.World.Weather()
	case "tp":
		return s.commandTeleport(source, args)
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
	target.MovementMu.Lock()
	defer target.MovementMu.Unlock()
	target.CancelPendingChunks()
	if err := s.World.PrepareDestination(context.Background(), int32(math.Floor(xyz[0]/16)), int32(math.Floor(xyz[2]/16))); err != nil {
		cx, cz := target.ChunkPos()
		s.requestChunksAround(target, cx, cz)
		return "Teleport destination unavailable: " + err.Error()
	}
	_, _, _, yaw, pitch, _ := target.Pos()
	if destinationRotation != nil {
		yaw, pitch = destinationRotation[0], destinationRotation[1]
	}
	s.entityMu.Lock()
	changed := target.SetPos(xyz[0], xyz[1], xyz[2], yaw, pitch, false)
	s.syncPosition(target)
	tp := proto.NewWriter().VarInt(target.EID).F64(xyz[0]).F64(xyz[1]).F64(xyz[2]).
		I8(angleI8(yaw)).I8(angleI8(pitch)).Bool(false).Payload()
	s.sendTracked(target, proto.SEntityTeleport, tp)
	s.updateVisibilityLocked(target)
	s.entityMu.Unlock()
	if changed {
		s.handleChunkChange(target)
	} else {
		cx, cz := target.ChunkPos()
		s.requestChunksAround(target, cx, cz)
		s.updateVisibility(target)
	}
	return fmt.Sprintf("Teleported %s to %.2f %.2f %.2f", target.Name, xyz[0], xyz[1], xyz[2])
}

func (s *Server) commandSetBlock(source *player.Player, args []string) string {
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
		err = s.World.TryPlaceBlock(world.BlockPosition{X: xyz[0], Y: xyz[1], Z: xyz[2]}, world.StateAir, state, nil)
	} else if mode == "replace" {
		err = s.World.SetBlock(xyz[0], xyz[1], xyz[2], state)
	} else {
		return "Usage: /setblock <x y z> <block[state=value]|state_id> [replace|keep]"
	}
	if err != nil {
		return "Block edit failed: " + err.Error()
	}
	return fmt.Sprintf("Set block at %d %d %d to state %d", xyz[0], xyz[1], xyz[2], state)
}
