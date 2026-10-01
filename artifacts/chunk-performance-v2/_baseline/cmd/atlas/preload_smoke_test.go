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

func TestBinarySpawnRadiusFlagOverride(t *testing.T) {
	serverExe := os.Getenv("ATLAS_SERVER_EXE")
	if serverExe == "" {
		t.Skip("set ATLAS_SERVER_EXE to the built binary")
	}
	serverExe, _ = filepath.Abs(serverExe)
	dir := t.TempDir()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	properties := fmt.Sprintf("server-ip=127.0.0.1\nserver-port=%d\nonline-mode=false\nlevel-type=minecraft:flat\nlevel-seed=1234\nspawn-preload-radius=0\nchunk-generation-workers=1\n", port)
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(properties), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, serverExe, "-gen", "3", "-spawn-radius", "2")
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
	ready, drained := make(chan bool, 1), make(chan struct{})
	var logs bytes.Buffer
	go func() {
		defer close(drained)
		started := false
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			logs.WriteString(line + "\n")
			if !started && strings.Contains(line, "listening on ") {
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
		t.Fatalf("startup failed:\n%s", logs.String())
	}
	for x := int32(-2); x <= 2; x++ {
		for z := int32(-2); z <= 2; z++ {
			if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "world", "region"), x, z); err != nil || !exists {
				t.Fatalf("flag override chunk %d,%d missing: %v", x, z, err)
			}
		}
	}
	if _, _, exists, err := anvil.ReadChunk(filepath.Join(dir, "world", "region"), 3, 0); err != nil || exists {
		t.Fatalf("radius exceeded: %v", err)
	}
	_, _ = stdin.Write([]byte("stop\n"))
	stdin.Close()
	err = command.Wait()
	<-drained
	if err != nil || !strings.Contains(logs.String(), "(radius: 2, workers: 3)") || !strings.Contains(logs.String(), "Spawn area ready (chunks: 25,") || !strings.Contains(logs.String(), "server stopped cleanly") {
		t.Fatalf("flags or clean shutdown failed: %v\n%s", err, logs.String())
	}
}
