package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Exercises the actual clock, console and protocol client in a separate flat
// save. It never touches the world owned by an already-running Atlas process.
func TestPerformanceBinaryConsoleAndProtocol(t *testing.T) {
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
		dir, err = os.MkdirTemp(base, "performance-smoke-")
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
	properties := fmt.Sprintf("server-ip=127.0.0.1\nserver-port=%d\nonline-mode=false\nview-distance=2\nsimulation-distance=2\nlevel-type=minecraft:flat\nlevel-seed=1234\nspawn-preload-radius=0\nchunk-generation-workers=1\n", port)
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(properties), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, serverExe, "-ops", "TestBot")
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
	var logs bytes.Buffer
	lines, drained := make(chan string, 1024), make(chan struct{})
	go func() {
		defer close(drained)
		defer close(lines)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			logs.WriteString(line + "\n")
			select {
			case lines <- line:
			default:
			}
		}
	}()
	readUntil := func(marker string) string {
		t.Helper()
		var response strings.Builder
		for {
			select {
			case line, ok := <-lines:
				if !ok {
					t.Fatalf("server output closed before %q:\n%s", marker, response.String())
				}
				response.WriteString(line + "\n")
				if strings.Contains(line, marker) {
					return response.String()
				}
			case <-ctx.Done():
				t.Fatalf("timeout waiting for %q:\n%s", marker, response.String())
			}
		}
	}
	readUntil("listening on ")
	select {
	case <-time.After(time.Second):
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	for _, cmd := range []string{"tps", "mspt", "status", "sysinfo"} {
		if _, err := fmt.Fprintln(stdin, cmd); err != nil {
			t.Fatal(err)
		}
		response := readUntil("Scope: world-clock loop;")
		if strings.Contains(response, "collecting samples") {
			t.Fatalf("running clock did not record samples:\n%s", response)
		}
		switch cmd {
		case "tps":
			values := regexp.MustCompile(`TPS \(1m/5m/15m\): ([0-9.]+) / ([0-9.]+) / ([0-9.]+)`).FindStringSubmatch(response)
			if len(values) != 4 {
				t.Fatal(response)
			}
			for _, value := range values[1:] {
				tps, err := strconv.ParseFloat(value, 64)
				if err != nil || tps < 10 || tps > 20 {
					t.Fatalf("idle clock TPS outside expected bounds: %s", response)
				}
			}
		case "mspt":
			if !strings.Contains(response, "avg/P95/max") || !strings.Contains(response, "Loop gap max:") {
				t.Fatal(response)
			}
		default:
			for _, field := range []string{"Atlas status:", "Memory:", "GC:", "Process CPU:", "World: cached chunks 1", "queued jobs 0"} {
				if !strings.Contains(response, field) {
					t.Fatalf("%s missing %s:\n%s", cmd, field, response)
				}
			}
			if runtime.GOOS == "windows" && strings.Contains(response, "unavailable") {
				t.Fatalf("native process counters unavailable:\n%s", response)
			}
		}
		t.Logf("console %s:\n%s", cmd, response)
	}
	client := exec.CommandContext(ctx, clientExe, "-addr", address, "-client-view", "2", "-expect-view", "2", "-expect-simulation", "2", "-operator")
	output, clientErr := client.CombinedOutput()
	_, _ = fmt.Fprintln(stdin, "stop")
	stdin.Close()
	serverErr := command.Wait()
	<-drained
	_ = os.WriteFile(filepath.Join(dir, "client.txt"), output, 0644)
	_ = os.WriteFile(filepath.Join(dir, "server.txt"), logs.Bytes(), 0644)
	if clientErr != nil || serverErr != nil || !strings.Contains(logs.String(), "server stopped cleanly") {
		t.Fatalf("client=%v server=%v:\n%s\n%s", clientErr, serverErr, output, logs.String())
	}
	t.Logf("protocol client:\n%s\nartifacts: %s", output, dir)
}
