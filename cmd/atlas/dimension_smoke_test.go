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
	"strconv"
	"strings"
	"testing"
	"time"
)

// Run actual server/client executables twice against the same normal save.
// This checks the wire transition and disk persistence, beyond World tests.
func TestDimensionBinaryTeleportAndRestart(t *testing.T) {
	serverExe, clientExe := os.Getenv("ATLAS_SERVER_EXE"), os.Getenv("ATLAS_TEST_EXE")
	if serverExe == "" || clientExe == "" {
		t.Skip("set ATLAS_SERVER_EXE and ATLAS_TEST_EXE to built binaries")
	}
	serverExe, _ = filepath.Abs(serverExe)
	clientExe, _ = filepath.Abs(clientExe)
	for _, seed := range []int64{1234, 9212937586943976949} {
		t.Run(strconv.FormatInt(seed, 10), func(t *testing.T) {
			dir := t.TempDir()
			if base := os.Getenv("ATLAS_SMOKE_DIR"); base != "" {
				if err := os.MkdirAll(base, 0755); err != nil {
					t.Fatal(err)
				}
				var err error
				dir, err = os.MkdirTemp(base, "dimension-smoke-")
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
			properties := fmt.Sprintf("server-ip=127.0.0.1\nserver-port=%d\nonline-mode=false\nview-distance=2\nsimulation-distance=2\nlevel-seed=%d\n", port, seed)
			if err = os.WriteFile(filepath.Join(dir, "server.properties"), []byte(properties), 0644); err != nil {
				t.Fatal(err)
			}
			var crystalUUID string
			for run := 0; run < 2; run++ {
				uuid := runDimensionBinarySmoke(t, serverExe, clientExe, dir, address, run)
				if os.Getenv("ATLAS_SMOKE_ENTITIES") != "" {
					if run == 0 {
						crystalUUID = uuid
					} else if uuid != crystalUUID {
						t.Fatalf("restart changed crystal UUID: %s want%s", uuid, crystalUUID)
					}
				}
			}
			if crystalUUID != "" {
				_, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "world", "DIM1", "entities"), 2, 0)
				if err != nil || !exists {
					t.Fatal("End entity Anvil missing", err)
				}
				object, _ := root.Object()
				entities, _ := object["Entities"].Elements()
				found := false
				for _, tag := range entities {
					data, _ := tag.Object()
					id, _ := data.Text("id")
					if id != "minecraft:end_crystal" {
						continue
					}
					parts, _ := data["UUID"].Value.([]int32)
					if len(parts) == 4 && fmt.Sprintf("%08x%08x%08x%08x", uint32(parts[0]), uint32(parts[1]), uint32(parts[2]), uint32(parts[3])) == crystalUUID {
						found = true
					}
				}
				if !found {
					t.Fatal("wire crystal UUID did not match saved vanilla entity")
				}
			}
			for _, relative := range []string{"region", "DIM-1/region", "DIM1/region"} {
				if _, err := os.Stat(filepath.Join(dir, "world", relative, "r.0.0.mca")); err != nil {
					t.Fatalf("missing persisted %s region: %v", relative, err)
				}
			}
			t.Logf("validated normal dimension save and logs: %s", dir)
		})
	}
}

func runDimensionBinarySmoke(t *testing.T, serverExe, clientExe, dir, address string, run int) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, serverExe, "-gen", "4", "-spawn-radius", "0", "-ops", "TestBot")
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
	if err = command.Start(); err != nil {
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
	args := []string{"-addr", address, "-dimension-smoke"}
	if os.Getenv("ATLAS_SMOKE_ENTITIES") != "" {
		args = append(args, "-dimension-entities")
	}
	if run > 0 {
		args = append(args, "-dimension-resume")
	}
	client := exec.CommandContext(ctx, clientExe, args...)
	output, clientErr := client.CombinedOutput()
	_, _ = stdin.Write([]byte("save\nstop\n"))
	stdin.Close()
	serverErr := command.Wait()
	<-drained
	_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("dimension-client-%d.txt", run+1)), output, 0644)
	_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("dimension-server-%d.txt", run+1)), logs.Bytes(), 0644)
	if clientErr != nil || serverErr != nil {
		t.Fatalf("run %d client=%v server=%v\n%s\n%s", run+1, clientErr, serverErr, output, logs.String())
	}
	if !strings.Contains(logs.String(), "server stopped cleanly") {
		t.Fatal("server did not stop cleanly")
	}
	t.Logf("run %d:\n%s", run+1, output)
	var uuid string
	for _, line := range strings.Split(string(output), "\n") {
		if i := strings.Index(line, "ENTITY crystal uuid="); i >= 0 {
			value := strings.Fields(line[i+len("ENTITY crystal uuid="):])[0]
			if uuid != "" && uuid != value {
				t.Fatalf("reconnect changed crystal UUID %s want%s", value, uuid)
			}
			uuid = value
		}
	}
	if os.Getenv("ATLAS_SMOKE_ENTITIES") != "" && uuid == "" {
		t.Fatal("no crystal identity in client report")
	}
	return uuid
}
