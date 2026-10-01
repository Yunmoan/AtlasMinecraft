package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"sort"
	"strings"
	"unicode/utf16"
)

type commandSpec struct {
	name    string
	level   int
	usage   string
	aliases []string
}

var commandSpecs = []commandSpec{
	{"help", 0, "/help [command]", []string{"minecraft:help"}},
	{"list", 0, "/list", []string{"minecraft:list"}},
	{"where", 0, "/where", nil},
	{"tps", 2, "/tps", nil},
	{"mspt", 2, "/mspt", nil},
	{"status", 2, "/status (alias: /sysinfo)", []string{"sysinfo"}},
	{"seed", 2, "/seed", []string{"minecraft:seed"}},
	{"time", 2, "/time <query [daytime|gametime|day]|set <day|noon|night|midnight|time>|add <time>>", []string{"minecraft:time"}},
	{"weather", 2, "/weather <clear|rain|thunder>", []string{"minecraft:weather"}},
	{"tp", 2, "/tp [player] <destination player|x y z> (coordinates accept ~ and ^)", []string{"teleport", "minecraft:tp", "minecraft:teleport"}},
	{"setblock", 2, "/setblock <x y z> <block[state=value]|state_id> [replace|keep]", []string{"minecraft:setblock"}},
	{"save", 2, "/save (alias: /save-all [flush])", []string{"save-all", "minecraft:save-all"}},
	{"op", 3, "/op <player>", []string{"minecraft:op"}},
	{"deop", 3, "/deop <player>", []string{"minecraft:deop"}},
}

func findCommand(name string) (commandSpec, bool) {
	for _, spec := range commandSpecs {
		if name == spec.name {
			return spec, true
		}
		for _, alias := range spec.aliases {
			if name == alias {
				return spec, true
			}
		}
	}
	return commandSpec{}, false
}

func commandHelp(level int, args []string) string {
	if len(args) > 1 {
		return "Usage: /help [command]"
	}
	if len(args) == 1 {
		spec, ok := findCommand(strings.ToLower(strings.TrimPrefix(args[0], "/")))
		if !ok || spec.level > level {
			return "Unknown command. Use /help."
		}
		return "Commands:\n" + spec.usage
	}
	lines := []string{"Commands:"}
	for _, spec := range commandSpecs {
		if spec.level <= level {
			lines = append(lines, spec.usage)
		}
	}
	return strings.Join(lines, "\n")
}

type commandNode struct {
	name        string
	children    []int32
	parser      int32 // -1 = literal, -2 = root
	properties  []byte
	executable  bool
	suggestions bool
}

func declareCommandsPayloadForLevel(level int) []byte {
	nodes := []commandNode{{parser: -2}}
	add := func(parent int32, node commandNode) int32 {
		index := int32(len(nodes))
		nodes = append(nodes, node)
		nodes[parent].children = append(nodes[parent].children, index)
		return index
	}
	literal := func(parent int32, name string, executable bool) int32 {
		return add(parent, commandNode{name: name, parser: -1, executable: executable})
	}
	argument := func(parent int32, name string, parser int32, properties []byte, executable, suggestions bool) int32 {
		return add(parent, commandNode{name: name, parser: parser, properties: properties, executable: executable, suggestions: suggestions})
	}
	for _, spec := range commandSpecs {
		if spec.level > level {
			continue
		}
		for _, name := range append([]string{spec.name}, spec.aliases...) {
			executable := spec.name != "op" && spec.name != "deop" && spec.name != "time" && spec.name != "weather" && spec.name != "tp" && spec.name != "setblock"
			root := literal(0, name, executable)
			switch spec.name {
			case "help":
				argument(root, "command", 5, proto.NewWriter().VarInt(0).Payload(), true, true)
			case "op", "deop":
				argument(root, "player", 7, nil, true, true)
			case "time":
				set := literal(root, "set", false)
				for _, value := range []string{"day", "noon", "night", "midnight"} {
					literal(set, value, true)
				}
				argument(set, "time", 42, proto.NewWriter().I32(0).Payload(), true, false)
				addTime := literal(root, "add", false)
				argument(addTime, "time", 42, proto.NewWriter().I32(0).Payload(), true, false)
				query := literal(root, "query", true)
				for _, value := range []string{"daytime", "gametime", "day"} {
					literal(query, value, true)
				}
			case "weather":
				for _, value := range []string{"clear", "rain", "thunder"} {
					literal(root, value, true)
				}
			case "tp":
				argument(root, "location", 10, nil, true, false)
				argument(root, "destination", 6, []byte{3}, true, true) // single player
				target := argument(root, "target", 6, []byte{3}, false, true)
				argument(target, "location", 10, nil, true, false)
				argument(target, "destination", 6, []byte{3}, true, true)
			case "setblock":
				position := argument(root, "pos", 8, nil, false, false)
				block := argument(position, "block", 12, nil, true, false)
				for _, mode := range []string{"replace", "keep"} {
					literal(block, mode, true)
				}
				// Atlas also accepts numeric protocol state IDs for diagnostics.
				numeric := argument(position, "state_id", 3, proto.NewWriter().U8(3).I32(0).I32(26683).Payload(), true, false)
				for _, mode := range []string{"replace", "keep"} {
					literal(numeric, mode, true)
				}
			case "save":
				if name != "save" {
					literal(root, "flush", true)
				}
			}
		}
	}
	w := proto.NewWriter().VarInt(int32(len(nodes)))
	for _, node := range nodes {
		flags := byte(0)
		if node.parser == -1 {
			flags = 1
		} else if node.parser >= 0 {
			flags = 2
		}
		if node.executable {
			flags |= 4
		}
		if node.suggestions {
			flags |= 16
		}
		w.U8(flags).VarInt(int32(len(node.children)))
		for _, child := range node.children {
			w.VarInt(child)
		}
		if node.parser != -2 {
			w.String(node.name)
		}
		if node.parser >= 0 {
			w.VarInt(node.parser).Raw(node.properties)
		}
		if node.suggestions {
			w.String("minecraft:ask_server")
		}
	}
	return w.VarInt(0).Payload()
}

func (s *Server) handleCommandSuggestions(p *player.Player, payload []byte) {
	r := proto.NewReader(payload)
	id, err := r.VarInt()
	if err != nil {
		return
	}
	text, err := r.String(32767)
	if err != nil || r.Remaining() != 0 {
		return
	}
	start := strings.LastIndexAny(text, " \t") + 1
	prefix := text[start:]
	leadingSlash := 0
	if start == 0 && strings.HasPrefix(prefix, "/") {
		leadingSlash = 1
		prefix = prefix[1:]
	}
	var candidates []string
	fields := strings.Fields(strings.TrimPrefix(text[:start], "/"))
	helpArgument := false
	if len(fields) == 1 {
		spec, known := findCommand(strings.ToLower(fields[0]))
		helpArgument = known && spec.name == "help"
	}
	if len(fields) == 0 || helpArgument {
		for _, spec := range commandSpecs {
			if spec.level <= s.operatorLevel(p) {
				candidates = append(candidates, spec.name)
				candidates = append(candidates, spec.aliases...)
			}
		}
	} else if spec, ok := findCommand(strings.ToLower(fields[0])); ok && spec.level <= s.operatorLevel(p) && (spec.name == "op" || spec.name == "deop" || spec.name == "tp") && len(fields) <= 2 {
		candidates = append(candidates, s.Hub.Names()...)
		if spec.name == "tp" {
			candidates = append(candidates, "@s", "@p")
		}
		if spec.name == "deop" && s.Cfg.Operators != nil {
			s.Cfg.Operators.mu.RLock()
			for _, op := range s.Cfg.Operators.entries {
				candidates = append(candidates, op.Name)
			}
			s.Cfg.Operators.mu.RUnlock()
		}
	}
	sort.Strings(candidates)
	matches := make([]string, 0, len(candidates))
	for _, value := range candidates {
		if strings.HasPrefix(strings.ToLower(value), strings.ToLower(prefix)) && (len(matches) == 0 || matches[len(matches)-1] != value) {
			matches = append(matches, value)
		}
	}
	// Brigadier string ranges use UTF-16 code units, not UTF-8 byte positions.
	from := len(utf16.Encode([]rune(text[:start+leadingSlash])))
	length := len(utf16.Encode([]rune(prefix)))
	w := proto.NewWriter().VarInt(id).VarInt(int32(from)).VarInt(int32(length)).VarInt(int32(len(matches)))
	for _, value := range matches {
		w.String(value).Bool(false)
	}
	p.SendPacket(proto.SCommandSuggestions, w.Payload())
}
