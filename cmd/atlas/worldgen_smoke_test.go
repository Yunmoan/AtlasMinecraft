package main

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/world"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math/bits"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Opt-in binary smoke test: start in an empty directory with the default
// normal preset, exercise the real protocol, stop cleanly, then reconnect to
// the saved position and check weather/time after restart.
func TestNormalWorldBinaryCreateAndRestart(t *testing.T) {
	serverExe, clientExe := os.Getenv("ATLAS_SERVER_EXE"), os.Getenv("ATLAS_TEST_EXE")
	if serverExe == "" || clientExe == "" {
		t.Skip("set ATLAS_SERVER_EXE and ATLAS_TEST_EXE to built binaries")
	}
	serverExe, _ = filepath.Abs(serverExe)
	clientExe, _ = filepath.Abs(clientExe)
	for _, test := range []struct {
		name      string
		seed      int64
		workers   int
		wantSpawn [3]int32
	}{
		{"seed_1234_workers_4", 1234, 4, [3]int32{-400, 65, -62}},
		{"seed_9212937586943976949_workers_8", 9212937586943976949, 8, [3]int32{96, 67, 64}},
	} {
		t.Run(test.name, func(t *testing.T) {
			runNormalWorldBinaryCreateAndRestart(t, serverExe, clientExe, test.seed, test.workers, test.wantSpawn)
		})
	}
}

func runNormalWorldBinaryCreateAndRestart(t *testing.T, serverExe, clientExe string, seed int64, workers int, wantSpawn [3]int32) {
	t.Helper()
	dir := t.TempDir()
	if base := os.Getenv("ATLAS_SMOKE_DIR"); base != "" {
		if err := os.MkdirAll(base, 0755); err != nil {
			t.Fatal(err)
		}
		var err error
		dir, err = os.MkdirTemp(base, "normal-smoke-")
		if err != nil {
			t.Fatal(err)
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	properties := fmt.Sprintf("server-ip=127.0.0.1\nserver-port=%d\nonline-mode=false\nview-distance=2\nsimulation-distance=2\nlevel-seed=%d\nchunk-generation-workers=1\n", port, seed)
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(properties), 0644); err != nil {
		t.Fatal(err)
	}
	var initialSpawn [3]int32
	for run := 0; run < 2; run++ {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, serverExe, "-gen", strconv.Itoa(workers), "-ops", "TestBot")
			command.Dir = dir
			stdout, err := command.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			command.Stderr = command.Stdout
			stdin, err := command.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			defer command.Process.Kill()
			ready := make(chan bool, 1)
			drained := make(chan struct{})
			var logs bytes.Buffer
			go func() {
				defer close(drained)
				scanner := bufio.NewScanner(stdout)
				started := false
				for scanner.Scan() {
					line := scanner.Text()
					logs.WriteString(line + "\n")
					if strings.Contains(line, "listening on ") && !started {
						started = true
						ready <- true
					}
				}
				if !started {
					ready <- false
				}
			}()
			if !<-ready {
				command.Wait()
				<-drained
				t.Fatalf("server did not start:\n%s", logs.String())
			}
			spawn := readSmokeWorldSpawn(t, filepath.Join(dir, "world", "level.dat"))
			if spawn != wantSpawn {
				t.Fatalf("seed %d spawn=%v, want %v", seed, spawn, wantSpawn)
			}
			if run == 0 {
				initialSpawn = spawn
			} else if spawn != initialSpawn {
				t.Fatalf("restart moved world spawn: %v, initially %v", spawn, initialSpawn)
			}
			// Read only the stored data while the real server is running. The
			// preload square follows the final spawn selected by the terrain scan.
			// It must be durable before any player logs in, including a restart.
			regionDir := filepath.Join(dir, "world", "region")
			for x := spawn[0]>>4 - 3; x <= spawn[0]>>4+3; x++ {
				for z := spawn[2]>>4 - 3; z <= spawn[2]>>4+3; z++ {
					if _, _, exists, err := anvil.ReadChunk(regionDir, x, z); err != nil || !exists {
						t.Fatalf("server accepted connections before spawn chunk %d,%d was saved: %v", x, z, err)
					}
				}
			}
			assertSmokeSpawnSafe(t, regionDir, spawn)
			args := []string{"-addr", address, "-client-view", "2", "-expect-view", "2", "-expect-simulation", "2", "-operator"}
			if run == 0 {
				args = append(args, "-expect-position", fmt.Sprintf("%g,%d,%g", float64(spawn[0])+.5, spawn[1], float64(spawn[2])+.5))
			} else {
				args = append(args, "-expect-position", "34.5,85,-16.5", "-expect-weather", "rain", "-expect-day", "13000", "-expect-saved-items")
			}
			client := exec.CommandContext(ctx, clientExe, args...)
			output, clientErr := client.CombinedOutput()
			_, _ = stdin.Write([]byte("save\nstop\n"))
			stdin.Close()
			serverErr := command.Wait()
			<-drained
			_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("client-%d.txt", run+1)), output, 0644)
			_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("server-%d.txt", run+1)), logs.Bytes(), 0644)
			if clientErr != nil || serverErr != nil {
				t.Fatalf("run %d: client=%v server=%v\n%s\n%s", run+1, clientErr, serverErr, output, logs.String())
			}
			if !strings.Contains(logs.String(), "server stopped cleanly") {
				t.Fatal("server did not save and stop cleanly")
			}
			if !strings.Contains(logs.String(), "Spawn area ready (chunks: 49,") || !strings.Contains(logs.String(), fmt.Sprintf("(radius: 3, workers: %d)", workers)) {
				t.Fatalf("missing preload readiness or worker override:\n%s", logs.String())
			}
			t.Logf("run %d:\n%s", run+1, output)
		}()
	}
	t.Logf("validated normal world and logs: %s", dir)
}

func readSmokeWorldSpawn(t *testing.T, path string) [3]int32 {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	data, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	object, _ := root.Object()
	level, ok := object.Object("Data")
	if !ok {
		t.Fatal("level.dat Data compound missing")
	}
	marker, ok := level.Object("AtlasWorldgen")
	ready, _ := marker.Number("SpawnReady")
	if !ok || ready != 1 {
		t.Fatalf("server accepted connections before spawn initialization was saved: SpawnReady=%d", ready)
	}
	var spawn [3]int32
	for i, key := range []string{"SpawnX", "SpawnY", "SpawnZ"} {
		value, ok := level.Number(key)
		if !ok {
			t.Fatalf("level.dat %s missing", key)
		}
		spawn[i] = int32(value)
	}
	return spawn
}

func assertSmokeSpawnSafe(t *testing.T, regionDir string, spawn [3]int32) {
	t.Helper()
	x, y, z := spawn[0], spawn[1], spawn[2]
	if y <= world.MinY || y+1 >= world.MaxY {
		t.Fatalf("spawn outside the safe build height: %v", spawn)
	}
	_, root, exists, err := anvil.ReadChunk(regionDir, x>>4, z>>4)
	if err != nil || !exists {
		t.Fatalf("read spawn chunk: exists=%v error=%v", exists, err)
	}
	for by := y - 1; by <= y+1; by++ {
		state := smokeStoredBlockState(t, root, x, by, z)
		block, err := world.BlockStateByID(state)
		if err != nil {
			t.Fatal(err)
		}
		if block.Name == "minecraft:water" || block.Name == "minecraft:lava" || block.Properties["waterlogged"] == "true" {
			t.Fatalf("spawn %v contains fluid at Y=%d: %s", spawn, by, block.Name)
		}
		position := world.BlockPosition{X: x, Y: by, Z: z}
		intersects, err := world.BlockIntersectsPlayer(state, position, float64(x)+.5, float64(y), float64(z)+.5, 1.8)
		if err != nil || intersects {
			t.Fatalf("spawn %v intersects %s at Y=%d: %v", spawn, block.Name, by, err)
		}
		if by == y-1 {
			supported, err := world.BlockIntersectsPlayer(state, position, float64(x)+.5, float64(y)-.01, float64(z)+.5, 1.8)
			if err != nil || !supported {
				t.Fatalf("spawn %v has no solid support at its feet: %s, error=%v", spawn, block.Name, err)
			}
		}
	}
	t.Logf("stored world spawn is safe: %v", spawn)
}

// Decode only the requested persisted block, without opening a second World
// actor or generating chunks while the binary under test owns the save.
func smokeStoredBlockState(t *testing.T, root nbt.Tag, x, y, z int32) uint32 {
	t.Helper()
	object, _ := root.Object()
	sections, _ := object["sections"].Elements()
	for _, tag := range sections {
		section, _ := tag.Object()
		sy, ok := section.Number("Y")
		if !ok || sy != int64(y>>4) {
			continue
		}
		states, _ := section.Object("block_states")
		palette, ok := states["palette"].Elements()
		if !ok || len(palette) == 0 {
			t.Fatalf("stored section %d has no block palette", sy)
		}
		index := 0
		if len(palette) > 1 {
			width := max(4, bits.Len(uint(len(palette)-1)))
			perLong := 64 / width
			data, ok := states["data"].Longs()
			blockIndex := int((y&15)<<8 | (z&15)<<4 | x&15)
			if !ok || blockIndex/perLong >= len(data) {
				t.Fatalf("stored section %d has invalid packed block data", sy)
			}
			index = int(uint64(data[blockIndex/perLong]) >> (uint(blockIndex%perLong) * uint(width)) & uint64((1<<width)-1))
			if index >= len(palette) {
				t.Fatalf("stored section %d block palette index %d out of range", sy, index)
			}
		}
		block, _ := palette[index].Object()
		name, _ := block.Text("Name")
		properties := make(map[string]string)
		if values, ok := block.Object("Properties"); ok {
			for key, value := range values {
				properties[key], _ = value.Text()
			}
		}
		state, err := world.BlockStateID(name, properties)
		if err != nil {
			t.Fatal(err)
		}
		return state
	}
	t.Fatalf("stored spawn column section %d missing", y>>4)
	return 0
}
