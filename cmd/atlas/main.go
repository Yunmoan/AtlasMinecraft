// Command atlas runs the Atlas Minecraft server (Java Edition 1.21.1,
// protocol 767, with configurable offline or online authentication).
package main

import (
	"atlas/internal/config"
	"atlas/internal/hub"
	"atlas/internal/logging"
	"atlas/internal/server"
	"atlas/internal/world"
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", ":25565", "listen address")
	motd := flag.String("motd", "Atlas — a Go-powered Minecraft server", "MOTD line")
	maxPlayers := flag.Int("max", 20, "maximum players")
	genWorkers := flag.Int("gen", 0, "chunk generation workers (0=automatic)")
	spawnRadius := flag.Int("spawn-radius", 3, "spawn preload radius in chunks (0=center only)")
	debug := flag.Bool("debug", false, "verbose protocol-level logging")
	tracePackets := flag.Bool("trace-packets", false, "log every play packet (also enables -debug)")
	worldDir := flag.String("world", "world", "Atlas world save directory")
	logFile := flag.String("log-file", "logs/latest.log", "rotating server log file")
	debugLogFile := flag.String("debug-log-file", "logs/debug.log", "rotating debug log file (with -debug)")
	opsFlag := flag.String("ops", "", "comma-separated operator names for world commands")
	propertiesPath := flag.String("properties", "server.properties", "Minecraft server.properties path")
	opsPath := flag.String("ops-file", "ops.json", "Minecraft ops.json path")
	seedFlag := flag.String("seed", "", "world seed for a new Java world (number or text)")
	levelType := flag.String("level-type", "", "new world type: minecraft:normal or minecraft:flat")
	flag.Parse()
	if *tracePackets {
		*debug = true
	}
	logOutput, err := logging.Setup(*logFile, *debugLogFile, *debug)
	if err != nil {
		_, _ = os.Stderr.WriteString("log setup: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logOutput.Close()
	logStartupBanner()
	slog.Info("Starting Minecraft server version 1.21.1")
	slog.Info("Loading properties")
	properties, err := config.Load(*propertiesPath)
	if err != nil {
		slog.Error("server.properties load failed", "error", err)
		os.Exit(1)
	}
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { setFlags[f.Name] = true })
	if !setFlags["addr"] {
		*addr = fmt.Sprintf("%s:%d", properties.ServerIP, properties.ServerPort)
	}
	if !setFlags["motd"] {
		*motd = properties.MOTD
	}
	if !setFlags["max"] {
		*maxPlayers = properties.MaxPlayers
	}
	if !setFlags["world"] {
		*worldDir = properties.LevelName
	}
	if !setFlags["seed"] {
		*seedFlag = properties.LevelSeed
	}
	if !setFlags["level-type"] {
		*levelType = properties.LevelType
	}
	if !setFlags["gen"] {
		*genWorkers = properties.ChunkGenerationWorkers
	}
	if !setFlags["spawn-radius"] {
		*spawnRadius = properties.SpawnPreloadRadius
	}
	if *genWorkers < 0 || *genWorkers > 256 || *spawnRadius < 0 || *spawnRadius > 16 {
		slog.Error("invalid generation settings", "gen", *genWorkers, "spawn-radius", *spawnRadius)
		os.Exit(1)
	}
	if *genWorkers == 0 {
		*genWorkers = max(1, min(8, runtime.GOMAXPROCS(0)/2))
	}
	slog.Info("Default game type: CREATIVE")
	if len(properties.Unsupported) > 0 {
		slog.Warn("Unsupported server.properties keys are ignored", "keys", strings.Join(properties.Unsupported, ", "))
	}
	var requestedSeed *int64
	if *seedFlag != "" {
		seed, err := config.ParseSeed(*seedFlag)
		if err != nil {
			slog.Error("invalid level-seed", "error", err)
			os.Exit(1)
		}
		requestedSeed = &seed
	}
	if err := createWorldIfMissing(*worldDir, *levelType, requestedSeed); err != nil {
		slog.Error("world initialization refused", "error", err)
		os.Exit(1)
	}
	operators, err := server.OpenOperators(*opsPath)
	if err != nil {
		slog.Error("operator list load failed", "path", *opsPath, "error", err)
		os.Exit(1)
	}
	if properties.OnlineMode && strings.TrimSpace(*opsFlag) != "" {
		slog.Error("-ops cannot assign offline UUIDs in online mode; use console /op after the player joins")
		os.Exit(1)
	}
	for _, name := range strings.Split(*opsFlag, ",") {
		if name = strings.TrimSpace(name); name != "" {
			if _, err := operators.Add(name); err != nil {
				slog.Error("could not add operator", "player", name, "error", err)
				os.Exit(1)
			}
		}
	}

	h := hub.New()
	slog.Info(fmt.Sprintf("Preparing level %q", *worldDir))
	w, err := world.Open(h, *genWorkers, *worldDir)
	if err != nil {
		slog.Error("world load failed", "error", err)
		os.Exit(1)
	}
	age, day := w.Time()
	slog.Info(fmt.Sprintf("World loaded: %s (seed %d, age %d, daytime %d, weather %s)", *worldDir, w.Seed(), age, day, w.Weather()))
	absWorld, _ := filepath.Abs(*worldDir)
	slog.Info("World generator", "path", absWorld, "mode", w.GenerationMode())
	if warning := w.GenerationWarning(); warning != "" {
		slog.Warn("vanilla generation is incomplete", "reason", warning)
	}
	spawnX, spawnZ := w.SpawnXZ()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	slog.Info("Preparing start region for dimension minecraft:overworld", "radius", *spawnRadius, "workers", *genWorkers)
	prepareStart := time.Now()
	lastProgress := time.Time{}
	if err := w.PrepareSpawn(ctx, spawnX>>4, spawnZ>>4, *spawnRadius, func(p world.SpawnPreparationProgress) {
		if p.Completed == 0 || p.Completed == p.Total || time.Since(lastProgress) >= time.Second {
			slog.Info("Preparing spawn area", "completed", p.Completed, "total", p.Total, "percent", 100*p.Completed/p.Total)
			lastProgress = time.Now()
		}
	}); err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Info("Spawn preparation canceled; stopping server")
			if err := w.Close(); err != nil {
				slog.Error("world save failed", "error", err)
				os.Exit(1)
			}
			return
		}
		slog.Error("spawn area load failed", "error", err)
		if err := w.Close(); err != nil {
			slog.Error("world save failed", "error", err)
		}
		os.Exit(1)
	}
	slog.Info("Spawn area ready", "chunks", (2*(*spawnRadius)+1)*(2*(*spawnRadius)+1), "elapsed", time.Since(prepareStart).Round(time.Millisecond))
	s := server.New(server.Config{
		Addr:               *addr,
		Motd:               *motd,
		MaxPlayers:         *maxPlayers,
		OnlineMode:         properties.OnlineMode,
		ViewDistance:       properties.ViewDistance,
		SimulationDistance: properties.SimulationDistance,
		Debug:              *debug,
		TracePackets:       *tracePackets,
		Operators:          operators,
	}, h, w)

	go runWorldClock(ctx, s)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if line := strings.TrimSpace(scanner.Text()); line != "" {
				if strings.EqualFold(strings.TrimPrefix(line, "/"), "stop") {
					slog.Info("Stopping server")
					stop()
					return
				}
				s.RunConsoleCommand(line)
			}
		}
		if err := scanner.Err(); err != nil {
			slog.Error("console input failed", "error", err)
		}
	}()
	go func() {
		<-ctx.Done()
		s.Close()
	}()
	serveErr := s.ListenAndServe()
	if serveErr != nil {
		slog.Error("server stopped", "error", serveErr)
		s.Close()
	}
	s.Close()
	stop()
	if err := w.Close(); err != nil {
		slog.Error("world save failed", "error", err)
		os.Exit(1)
	}
	if serveErr != nil {
		os.Exit(1)
	}
	slog.Info("server stopped cleanly")
}

func createWorldIfMissing(dir, levelType string, seed *int64) error {
	_, err := os.Stat(filepath.Join(dir, "level.dat"))
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect %s: %w", filepath.Join(dir, "level.dat"), err)
	}
	if levelType == "flat" || levelType == "minecraft:flat" {
		return world.CreateFlatWorld(dir, seed)
	}
	if levelType != "normal" && levelType != "minecraft:normal" {
		return fmt.Errorf("unsupported level-type %q", levelType)
	}
	return world.CreateNormalWorld(dir, seed)
}

func runWorldClock(ctx context.Context, s *server.Server) {
	s.StartTickMetrics()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	lastTick := time.Now()
	save := time.NewTicker(30 * time.Second)
	defer save.Stop()
	for {
		select {
		case <-tick.C:
			started := time.Now()
			elapsed := time.Since(lastTick)
			ticks := int64(elapsed / (50 * time.Millisecond))
			if ticks > 0 {
				s.World.AdvanceTime(ticks)
				lastTick = lastTick.Add(time.Duration(ticks) * 50 * time.Millisecond)
			}
			s.RecordTick(started)
		case <-save.C:
			started := time.Now()
			s.SavePlayers()
			if err := s.World.Save(); err != nil {
				slog.Error("periodic world save failed", "error", err)
			} else {
				slog.Debug("world metadata saved", "component", "world")
			}
			s.RecordAutosave(started)
		case <-ctx.Done():
			return
		}
	}
}
