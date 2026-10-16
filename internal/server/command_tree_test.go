package server

import (
	"atlas/internal/proto"
	"strings"
	"testing"
)

func readCommandTree(t *testing.T, payload []byte) []commandNode {
	t.Helper()
	r := proto.NewReader(payload)
	count, err := r.VarInt()
	if err != nil || count < 2 || count > 2048 {
		t.Fatalf("invalid node count: %d %v", count, err)
	}
	nodes := make([]commandNode, count)
	for i := range nodes {
		flags, err := r.U8()
		if err != nil {
			t.Fatal(err)
		}
		nodes[i].parser = -2
		nodes[i].executable = flags&4 != 0
		nodes[i].suggestions = flags&16 != 0
		children, err := r.VarInt()
		if err != nil || children < 0 || children >= count {
			t.Fatal("bad child count")
		}
		for range children {
			index, err := r.VarInt()
			if err != nil || index <= int32(i) || index >= count {
				t.Fatal("invalid tree edge")
			}
			nodes[i].children = append(nodes[i].children, index)
		}
		if flags&8 != 0 {
			t.Fatal("unexpected redirect")
		}
		if flags&3 != 0 {
			nodes[i].name, err = r.String(256)
			if err != nil {
				t.Fatal(err)
			}
		}
		switch flags & 3 {
		case 0:
			if i != 0 {
				t.Fatal("non-root root node")
			}
		case 1:
			nodes[i].parser = -1
		case 2:
			nodes[i].parser, err = r.VarInt()
			if err != nil {
				t.Fatal(err)
			}
			switch nodes[i].parser {
			case 3:
				bounds, err := r.U8()
				if err != nil {
					t.Fatal(err)
				}
				if bounds&1 != 0 {
					r.I32()
				}
				if bounds&2 != 0 {
					r.I32()
				}
			case 5:
				kind, err := r.VarInt()
				if err != nil || kind != 0 && kind != 2 {
					t.Fatal("invalid word or greedy command argument")
				}
				nodes[i].properties = proto.NewWriter().VarInt(kind).Payload()
			case 6:
				options, err := r.U8()
				if err != nil || options != 3 {
					t.Fatal("entity argument must be a single player")
				}
			case 7, 8, 10, 12, 27:
			case 42:
				minimum, err := r.I32()
				if err != nil || minimum != 0 {
					t.Fatal("invalid time properties")
				}
			default:
				t.Fatalf("unexpected parser ID %d", nodes[i].parser)
			}
		default:
			t.Fatal("invalid command node type")
		}
		if nodes[i].suggestions {
			provider, err := r.String(256)
			if err != nil || provider != "minecraft:ask_server" {
				t.Fatal("invalid suggestion provider")
			}
		}
	}
	root, err := r.VarInt()
	if err != nil || root != 0 || r.Remaining() != 0 {
		t.Fatal("wrong root or trailing command data")
	}
	return nodes
}

func treePath(nodes []commandNode, path ...string) (commandNode, bool) {
	node := nodes[0]
	for _, name := range path {
		found := false
		for _, child := range node.children {
			if nodes[child].name == name {
				node = nodes[child]
				found = true
				break
			}
		}
		if !found {
			return commandNode{}, false
		}
	}
	return node, true
}

func TestCommandTreesMatchPermissionsAndTypedArguments(t *testing.T) {
	for level := 0; level <= 4; level++ {
		nodes := readCommandTree(t, declareCommandsPayloadForLevel(level))
		for _, spec := range commandSpecs {
			for _, name := range append([]string{spec.name}, spec.aliases...) {
				if _, present := treePath(nodes, name); present != (level >= spec.level) {
					t.Fatalf("level %d command %s mismatch", level, name)
				}
			}
		}
		if level < 2 {
			continue
		}
		for _, test := range []struct {
			path       []string
			parser     int32
			executable bool
		}{
			{[]string{"execute", "in", "minecraft:the_nether", "run", "command"}, 5, true},
			{[]string{"tp", "location"}, 10, true}, {[]string{"teleport", "target", "destination"}, 6, true},
			{[]string{"setblock", "pos"}, 8, false}, {[]string{"setblock", "pos", "block"}, 12, true},
			{[]string{"setblock", "pos", "block", "keep"}, -1, true},
			{[]string{"time", "set", "time"}, 42, true}, {[]string{"time", "query", "day"}, -1, true},
			{[]string{"weather", "rain"}, -1, true}, {[]string{"save-all", "flush"}, -1, true},
			{[]string{"setworldspawn"}, -1, true}, {[]string{"setworldspawn", "pos"}, 8, false},
			{[]string{"setworldspawn", "pos", "angle"}, 27, true},
			{[]string{"spawnpoint", "targets"}, 6, false}, {[]string{"spawnpoint", "targets", "pos"}, 8, false},
			{[]string{"spawnpoint", "targets", "pos", "angle"}, 27, true},
		} {
			node, ok := treePath(nodes, test.path...)
			if !ok || node.parser != test.parser || node.executable != test.executable {
				t.Fatalf("level %d invalid path %v: %+v", level, test.path, node)
			}
		}
		for _, name := range []string{"tp", "setblock", "time", "weather"} {
			if n, _ := treePath(nodes, name); n.executable {
				t.Fatalf("incomplete /%s marked executable", name)
			}
		}
	}
}

func TestPlayerNameSuggestionsAndPermissionFiltering(t *testing.T) {
	s, p := interactionServer(t)
	p.Name = "Alice"
	s.Cfg.Ops = map[string]bool{"alice": true}
	for _, test := range []struct {
		input string
		start int32
		want  string
	}{{"/tp Al", 4, "Alice"}, {"/minecraft:tp Al", 14, "Alice"}, {"/help sta", 6, "status"}} {
		s.handleCommandSuggestions(p, proto.NewWriter().VarInt(17).String(test.input).Payload())
		packet := <-p.Out
		r := proto.NewReader(packet.Payload)
		id, _ := r.VarInt()
		start, _ := r.VarInt()
		length, _ := r.VarInt()
		count, _ := r.VarInt()
		var names []string
		for range count {
			name, _ := r.String(256)
			names = append(names, name)
			tooltip, _ := r.Bool()
			if tooltip {
				t.Fatal("unexpected tooltip")
			}
		}
		if packet.ID != proto.SCommandSuggestions || id != 17 || start != test.start || length <= 0 || r.Remaining() != 0 || !strings.Contains(strings.Join(names, ","), test.want) {
			t.Fatalf("bad suggestions for %s: %v", test.input, names)
		}
	}
	s.Cfg.Ops = nil
	s.handleCommandSuggestions(p, proto.NewWriter().VarInt(18).String("/tp Al").Payload())
	packet := <-p.Out
	r := proto.NewReader(packet.Payload)
	r.VarInt()
	r.VarInt()
	r.VarInt()
	count, _ := r.VarInt()
	if count != 0 {
		t.Fatal("guest received privileged player suggestions")
	}
}
