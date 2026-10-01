// Package config reads the supported subset of Minecraft server.properties.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"
)

type Properties struct {
	ServerIP               string
	ServerPort             int
	MOTD                   string
	MaxPlayers             int
	OnlineMode             bool
	ViewDistance           int
	SimulationDistance     int
	LevelName              string
	LevelSeed              string
	LevelType              string
	SpawnPreloadRadius     int
	ChunkGenerationWorkers int
	Unsupported            []string
}

func Defaults() Properties {
	return Properties{ServerPort: 25565, MOTD: "Atlas Minecraft server", MaxPlayers: 20, ViewDistance: 10, SimulationDistance: 10, LevelName: "world", LevelType: "minecraft:normal", SpawnPreloadRadius: 3}
}

const defaultFile = "# Atlas server.properties (Minecraft Java 1.21.1 compatible keys)\n" +
	"# Other vanilla keys are currently ignored by Atlas.\n" +
	"# Normal worlds use pure-Go terrain and supported decoration; structures are not yet generated.\n" +
	"server-ip=\nserver-port=25565\nmotd=Atlas Minecraft server\n" +
	"max-players=20\nonline-mode=false\nview-distance=10\nsimulation-distance=10\nlevel-name=world\nlevel-seed=\nlevel-type=minecraft:normal\n" +
	"# Atlas settings: radius in chunks; workers=0 selects automatically.\nspawn-preload-radius=3\nchunk-generation-workers=0\n"

// Load creates a small editable properties file when none exists. It leaves
// existing files intact, including unsupported vanilla properties.
func Load(path string) (Properties, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return Properties{}, err
		}
		if err := os.WriteFile(path, []byte(defaultFile), 0644); err != nil {
			return Properties{}, err
		}
	} else if err != nil {
		return Properties{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return Properties{}, err
	}
	defer f.Close()
	p := Defaults()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4096), 2<<20)
	lineNumber := 0
	for sc.Scan() {
		lineNumber++
		line := sc.Text()
		if lineNumber == 1 {
			line = strings.TrimPrefix(line, "\ufeff") // PowerShell 5.1 writes UTF-8 BOMs.
		}
		line = strings.TrimLeft(line, " \t\f")
		if line == "" || line[0] == '#' || line[0] == '!' {
			continue
		}
		for hasContinuation(line) {
			line = line[:len(line)-1]
			if !sc.Scan() {
				return Properties{}, fmt.Errorf("%s:%d: unterminated continuation", path, lineNumber)
			}
			lineNumber++
			line += strings.TrimLeft(sc.Text(), " \t\f")
		}
		key, value := splitProperty(line)
		key, err = unescape(key)
		if err != nil {
			return Properties{}, fmt.Errorf("%s:%d: key: %w", path, lineNumber, err)
		}
		value, err = unescape(value)
		if err != nil {
			return Properties{}, fmt.Errorf("%s:%d: %s: %w", path, lineNumber, key, err)
		}
		switch key {
		case "server-ip":
			p.ServerIP = value
		case "server-port":
			p.ServerPort, err = strconv.Atoi(value)
			if err != nil || p.ServerPort < 1 || p.ServerPort > 65535 {
				return Properties{}, fmt.Errorf("%s:%d: invalid server-port %q", path, lineNumber, value)
			}
		case "motd":
			p.MOTD = value
		case "max-players":
			p.MaxPlayers, err = strconv.Atoi(value)
			if err != nil || p.MaxPlayers < 1 || p.MaxPlayers > 100000 {
				return Properties{}, fmt.Errorf("%s:%d: invalid max-players %q", path, lineNumber, value)
			}
		case "online-mode":
			if value != "true" && value != "false" {
				return Properties{}, fmt.Errorf("%s:%d: invalid online-mode %q (expected true or false)", path, lineNumber, value)
			}
			p.OnlineMode = value == "true"
		case "view-distance":
			p.ViewDistance, err = strconv.Atoi(value)
			if err != nil || p.ViewDistance < 2 || p.ViewDistance > 16 {
				return Properties{}, fmt.Errorf("%s:%d: invalid view-distance %q (supported range 2..16)", path, lineNumber, value)
			}
		case "simulation-distance":
			p.SimulationDistance, err = strconv.Atoi(value)
			if err != nil || p.SimulationDistance < 2 || p.SimulationDistance > 16 {
				return Properties{}, fmt.Errorf("%s:%d: invalid simulation-distance %q (supported range 2..16)", path, lineNumber, value)
			}
		case "level-name":
			if value == "" {
				return Properties{}, fmt.Errorf("%s:%d: level-name is empty", path, lineNumber)
			}
			p.LevelName = value
		case "level-seed":
			p.LevelSeed = value
		case "level-type":
			if value != "minecraft:normal" && value != "normal" && value != "minecraft:flat" && value != "flat" {
				return Properties{}, fmt.Errorf("%s:%d: unsupported level-type %q", path, lineNumber, value)
			}
			p.LevelType = value
		case "spawn-preload-radius":
			p.SpawnPreloadRadius, err = strconv.Atoi(value)
			if err != nil || p.SpawnPreloadRadius < 0 || p.SpawnPreloadRadius > 16 {
				return Properties{}, fmt.Errorf("%s:%d: invalid spawn-preload-radius %q (supported range 0..16)", path, lineNumber, value)
			}
		case "chunk-generation-workers":
			p.ChunkGenerationWorkers, err = strconv.Atoi(value)
			if err != nil || p.ChunkGenerationWorkers < 0 || p.ChunkGenerationWorkers > 256 {
				return Properties{}, fmt.Errorf("%s:%d: invalid chunk-generation-workers %q (supported range 0..256)", path, lineNumber, value)
			}
		default:
			p.Unsupported = append(p.Unsupported, key)
		}
	}
	if err := sc.Err(); err != nil {
		return Properties{}, err
	}
	return p, nil
}

func hasContinuation(s string) bool {
	n := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

func splitProperty(s string) (string, string) {
	for i, escaped := 0, false; i < len(s); i++ {
		if escaped {
			escaped = false
			continue
		}
		if s[i] == '\\' {
			escaped = true
			continue
		}
		if s[i] == '=' || s[i] == ':' || s[i] == ' ' || s[i] == '\t' || s[i] == '\f' {
			key := s[:i]
			j := i
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\f') {
				j++
			}
			if j < len(s) && (s[j] == '=' || s[j] == ':') {
				j++
			}
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\f') {
				j++
			}
			return key, s[j:]
		}
	}
	return s, ""
}

func unescape(s string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 == len(s) {
			out.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 't':
			out.WriteByte('\t')
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 'f':
			out.WriteByte('\f')
		case 'u':
			if i+4 >= len(s) {
				return "", errors.New("short Unicode escape")
			}
			n, err := strconv.ParseUint(s[i+1:i+5], 16, 16)
			if err != nil {
				return "", errors.New("invalid Unicode escape")
			}
			out.WriteRune(rune(n))
			i += 4
		default:
			out.WriteByte(s[i])
		}
	}
	return out.String(), nil
}

// ParseSeed follows vanilla's numeric-or-Java-string-hash rule. An empty
// value is left to the caller to choose randomly for a new world.
func ParseSeed(value string) (int64, error) {
	if value == "" {
		return 0, errors.New("empty seed")
	}
	if number, err := strconv.ParseInt(value, 10, 64); err == nil {
		return number, nil
	}
	var hash int32
	for _, codeUnit := range utf16.Encode([]rune(value)) {
		hash = hash*31 + int32(codeUnit)
	}
	return int64(hash), nil
}
