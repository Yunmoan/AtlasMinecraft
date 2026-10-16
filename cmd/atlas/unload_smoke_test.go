package main

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Check the production autosave/unload timer with a real protocol client.
// This intentionally uses the normal one-minute idle timeout rather than a
// test-only maintenance setting. All files belong to a separate temporary save.
func TestIdleChunksBinaryUnloadAndReload(t *testing.T) {
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
		dir, err = os.MkdirTemp(base, "unload-smoke-")
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
	properties := fmt.Sprintf("server-ip=127.0.0.1\nserver-port=%d\nonline-mode=false\nview-distance=2\nsimulation-distance=2\nlevel-type=minecraft:flat\nlevel-seed=1234\nspawn-preload-radius=3\nchunk-generation-workers=1\n", port)
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(properties), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, serverExe, "-ops", "TestBot", "-debug")
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
	waited := false
	defer func() {
		stdin.Close()
		if !waited {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
		<-drained
		_ = os.WriteFile(filepath.Join(dir, "server.txt"), logs.Bytes(), 0644)
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
	statusCounts := regexp.MustCompile(`World: cached chunks (\d+); light columns (\d+)`)
	status := func() (int, int) {
		t.Helper()
		if _, err := fmt.Fprintln(stdin, "status"); err != nil {
			t.Fatal(err)
		}
		response := readUntil("Scope: world-clock loop;")
		values := statusCounts.FindStringSubmatch(response)
		if len(values) != 3 {
			t.Fatalf("missing cache counters:\n%s", response)
		}
		chunks, _ := strconv.Atoi(values[1])
		columns, _ := strconv.Atoi(values[2])
		return chunks, columns
	}
	if chunks, columns := status(); chunks != 49 || columns != 49 {
		t.Fatalf("spawn cache: %d chunks, %d light columns", chunks, columns)
	}
	runClient := func(run int) {
		t.Helper()
		args := []string{"-addr", address, "-client-view", "2", "-expect-view", "2", "-expect-simulation", "2", "-operator"}
		if run == 2 {
			args = append(args, "-expect-position", "34.5,85,-16.5", "-expect-weather", "rain", "-expect-saved-items")
		}
		output, err := exec.CommandContext(ctx, clientExe, args...).CombinedOutput()
		_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("client-%d.txt", run)), output, 0644)
		if err != nil {
			t.Fatalf("client run %d: %v\n%s", run, err, output)
		}
		t.Logf("client run %d passed", run)
	}
	// Direct Anvil reads do not access the server's cache or renew idle age.
	blockSnapshot := func() nbt.Tag {
		t.Helper()
		var root nbt.Tag
		var exists bool
		var err error
		deadline := time.Now().Add(2 * time.Second)
		for {
			_, root, exists, err = anvil.ReadChunk(filepath.Join(dir, "world", "region"), 0, 0)
			// This reader is in another process and does not share the server's
			// region mutex. Windows briefly denies opening a replacement file.
			// Retry only ERROR_SHARING_VIOLATION, never parse/corruption errors.
			if err == nil || runtime.GOOS != "windows" || !errors.Is(err, syscall.Errno(32)) || time.Now().After(deadline) {
				break
			}
			select {
			case <-time.After(20 * time.Millisecond):
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		}
		if err != nil || !exists {
			t.Fatalf("edited chunk missing: %v", err)
		}
		object, _ := root.Object()
		sections, ok := object["sections"].Elements()
		if !ok {
			t.Fatal("edited chunk has no sections")
		}
		states := make([]nbt.Tag, 0, len(sections))
		for _, section := range sections {
			data, _ := section.Object()
			if state, ok := data["block_states"]; ok {
				states = append(states, nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"Y": data["Y"], "block_states": state}})
			}
		}
		return nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: states}
	}
	initial := blockSnapshot()
	runClient(1)
	if _, err := fmt.Fprintln(stdin, "save"); err != nil {
		t.Fatal(err)
	}
	readUntil("World saved.")
	edited := blockSnapshot()
	if reflect.DeepEqual(initial, edited) {
		t.Fatal("protocol client did not persist edits")
	}
	chunks, columns := status()
	if chunks == 0 || columns == 0 {
		t.Fatal("recently used cache was already released")
	}
	t.Logf("after client departure: %d chunks, %d light columns", chunks, columns)
	started := time.Now()
	for chunks != 0 || columns != 0 {
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		chunks, columns = status()
		if time.Since(started) > 110*time.Second && (chunks != 0 || columns != 0) {
			t.Fatalf("autosave did not release idle caches: %d chunks, %d light columns", chunks, columns)
		}
	}
	t.Logf("automatic unload after %s: cached chunks 0, light columns 0", time.Since(started).Round(time.Second))
	if !reflect.DeepEqual(edited, blockSnapshot()) {
		t.Fatal("automatic unload changed saved block states")
	}
	runClient(2)
	if chunks, columns := status(); chunks == 0 || columns == 0 {
		t.Fatal("returning player did not reload chunks and lighting")
	}
	_, _ = fmt.Fprintln(stdin, "stop")
	stdin.Close()
	serverErr := command.Wait()
	waited = true
	<-drained
	if serverErr != nil || !strings.Contains(logs.String(), "server stopped cleanly") {
		t.Fatalf("server shutdown: %v\n%s", serverErr, logs.String())
	}
	t.Logf("validated autosave unload/reload: %s", dir)
}
