package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"context"
	"fmt"
	"log/slog"
	"math"
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
	p.SendPacket(proto.SSystemChat, systemChat(response))
	slog.Debug("command result", "source", p.Name, "result", response)
}

// RunConsoleCommand has operator privileges and reports its result in logs.
func (s *Server) RunConsoleCommand(line string) {
	slog.Info("Console issued server command: /" + strings.TrimPrefix(strings.TrimSpace(line), "/"))
	response := s.executeCommand(nil, line)
	slog.Info(response)
}

func (s *Server) executeCommand(source *player.Player, line string) string {
	fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/"))
	if len(fields) == 0 {
		return "Usage: /help"
	}
	cmd, args := strings.ToLower(fields[0]), fields[1:]
	required := 2
	switch cmd {
	case "help", "list", "where":
		required = 0
	case "op", "deop":
		required = 3
	}
	if s.operatorLevel(source) < required {
		return "You do not have permission to use this command."
	}
	switch cmd {
	case "help":
		if s.operatorLevel(source) >= 3 {
			return "Commands: /help, /list, /where, /tps, /mspt, /status (alias /sysinfo), /seed, /op <player>, /deop <player>, /time <query|set|add>, /weather <clear|rain|thunder>, /tp [player] <x> <y> <z>, /setblock <x> <y> <z> <state>, /save"
		}
		if s.operatorLevel(source) >= 2 {
			return "Commands: /help, /list, /where, /tps, /mspt, /status (alias /sysinfo), /seed, /time <query|set|add>, /weather <clear|rain|thunder>, /tp [player] <x> <y> <z>, /setblock <x> <y> <z> <state>, /save"
		}
		return "Commands: /help, /list, /where"
	case "tps", "mspt", "status", "sysinfo":
		if len(args) != 0 {
			return "Usage: /" + cmd
		}
		return s.commandPerformance(cmd)
	case "list":
		names := s.Hub.Names()
		return fmt.Sprintf("Players online (%d/%d): %s", len(names), s.Cfg.MaxPlayers, strings.Join(names, ", "))
	case "where":
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
		return s.commandSetBlock(args)
	case "save":
		if err := s.World.Save(); err != nil {
			return "World save failed: " + err.Error()
		}
		return "World saved."
	default:
		return "Unknown command. Use /help."
	}
}

func (s *Server) commandTime(args []string) string {
	usage := "Usage: /time query | /time set <day|noon|night|midnight|ticks> | /time add <ticks>"
	if len(args) == 1 && args[0] == "query" {
		age, day := s.World.Time()
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
		ticks, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil || ticks < 0 || ticks > (1<<63-1)-24000 {
			return usage
		}
	}
	switch args[0] {
	case "set":
		s.World.SetDayTime(ticks)
	case "add":
		if ticks > 24000000 {
			return "Time increment must be at most 24000000 ticks."
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
	return fmt.Sprintf("Time set to %d ticks", day%24000)
}

func (s *Server) commandTeleport(source *player.Player, args []string) string {
	usage := "Usage: /tp [player] <x> <y> <z>"
	var target *player.Player
	if len(args) == 3 {
		target = source
	} else if len(args) == 4 {
		target = s.Hub.FindName(args[0])
		args = args[1:]
	} else {
		return usage
	}
	if target == nil {
		return "Player not found; use /list."
	}
	var xyz [3]float64
	for i, v := range args {
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || math.Abs(parsed) > 30000000 {
			return usage
		}
		xyz[i] = parsed
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

func (s *Server) commandSetBlock(args []string) string {
	if len(args) != 4 {
		return "Usage: /setblock <x> <y> <z> <air|stone|grass_block|dirt|bedrock|state_id>"
	}
	var xyz [3]int32
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseInt(args[i], 10, 32)
		if err != nil || v < -30000000 || v > 30000000 {
			return "Invalid block coordinates."
		}
		xyz[i] = int32(v)
	}
	states := map[string]uint32{"air": world.StateAir, "stone": world.StateStone,
		"grass_block": world.StateGrass, "dirt": world.StateDirt, "bedrock": world.StateBedrock}
	name := strings.TrimPrefix(strings.ToLower(args[3]), "minecraft:")
	state, ok := states[name]
	if !ok {
		v, err := strconv.ParseUint(name, 10, 15)
		if err != nil {
			return "Unknown block state. Use a listed name or a numeric 1.21.1 state ID."
		}
		state = uint32(v)
	}
	if err := s.World.SetBlock(xyz[0], xyz[1], xyz[2], state); err != nil {
		return "Block edit failed: " + err.Error()
	}
	return fmt.Sprintf("Set block at %d %d %d to state %d", xyz[0], xyz[1], xyz[2], state)
}
