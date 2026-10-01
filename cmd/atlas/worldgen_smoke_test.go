package main

import (
	"atlas/internal/anvil"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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
	properties := fmt.Sprintf("server-ip=127.0.0.1\nserver-port=%d\nonline-mode=false\nview-distance=2\nsimulation-distance=2\nlevel-seed=1234\nchunk-generation-workers=1\n", port)
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(properties), 0644); err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 2; run++ {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, serverExe, "-gen", "4", "-ops", "TestBot")
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
			// The default preload area must be durable before any player logs in,
			// including a restart. -gen 4 overrides the property above.
			for x := int32(-3); x <= 3; x++ {
				for z := int32(-3); z <= 3; z++ {
					if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "world", "region"), x, z); err != nil || !exists {
						t.Fatalf("server accepted connections before spawn chunk %d,%d was saved: %v", x, z, err)
					}
				}
			}
			args := []string{"-addr", address, "-client-view", "2", "-expect-view", "2", "-expect-simulation", "2", "-operator"}
			if run != 0 {
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
			if !strings.Contains(logs.String(), "Spawn area ready (chunks: 49,") || !strings.Contains(logs.String(), "(radius: 3, workers: 4)") {
				t.Fatalf("missing preload readiness or worker override:\n%s", logs.String())
			}
			t.Logf("run %d:\n%s", run+1, output)
		}()
	}
	t.Logf("validated normal world and logs: %s", dir)
}
