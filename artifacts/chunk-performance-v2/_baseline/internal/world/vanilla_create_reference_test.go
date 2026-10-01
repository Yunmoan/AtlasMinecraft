package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
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

// ATLAS_SERVER_JAR opts into an independent interoperability check. The JAR
// runs only in this test, never in the Atlas server.
func TestNewWorldMetadataOpensInVanillaServer(t *testing.T) {
	jar := os.Getenv("ATLAS_SERVER_JAR")
	if jar == "" {
		t.Skip("set ATLAS_SERVER_JAR to the Minecraft 1.21.1 server.jar")
	}
	absJar, err := filepath.Abs(jar)
	if err != nil {
		t.Fatal(err)
	}
	for _, flat := range []bool{true, false} {
		name := "normal"
		if flat {
			name = "flat"
		}
		t.Run(name, func(t *testing.T) {
			serverDir := t.TempDir()
			worldDir := filepath.Join(serverDir, "world")
			seed := int64(1234)
			cx, cz := int32(0), int32(0)
			var trees map[[3]int32]uint32
			var generated *VanillaSave
			var chunk *Chunk
			var markerY int32
			if flat {
				if err := CreateFlatWorld(worldDir, &seed); err != nil {
					t.Fatal(err)
				}
				generated, err = openVanilla(worldDir)
				if err != nil {
					t.Fatal(err)
				}
				chunk, err = generated.generateFlat(0, 0)
				if err != nil {
					t.Fatal(err)
				}
				markerY = -60
			} else {
				seed, cx, cz = 0, -96, -128
				if err := CreateNormalWorld(worldDir, &seed); err != nil {
					t.Fatal(err)
				}
				generated, err = openVanilla(worldDir)
				if err != nil {
					t.Fatal(err)
				}
				chunk, err = generated.loadChunk(cx, cz)
				if err != nil {
					t.Fatal(err)
				}
				markerY = chunk.Height(0, 0)
				trees = treeStatesInChunk(t, chunk)
				plants := vegetationStatesInChunk(t, chunk)
				for position, id := range plants {
					trees[position] = id
				}
			}
			chunk.SetBlock(0, 0, markerY, StateStone)
			if err := generated.persistChunk(chunk); err != nil {
				t.Fatalf("write %s spawn chunk: %v", name, err)
			}
			root, _ := generated.root.Object()
			data, _ := root.Object("Data")
			rules, _ := data.Object("GameRules")
			rules["spawnChunkRadius"] = nbt.Tag{Type: nbt.String, Value: "0"}
			if err := writeGzipNBTAtomic(filepath.Join(worldDir, "level.dat"), "", generated.root); err != nil {
				t.Fatal(err)
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			port := listener.Addr().(*net.TCPAddr).Port
			listener.Close()
			if err := os.WriteFile(filepath.Join(serverDir, "eula.txt"), []byte("eula=true\n"), 0644); err != nil {
				t.Fatal(err)
			}
			properties := fmt.Sprintf("level-name=world\nonline-mode=false\nserver-ip=127.0.0.1\nserver-port=%d\nview-distance=2\nsimulation-distance=2\n", port)
			if err := os.WriteFile(filepath.Join(serverDir, "server.properties"), []byte(properties), 0644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, "java", "-jar", absJar, "nogui")
			command.Dir = serverDir
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
			ready := make(chan bool, 1)
			done := make(chan struct{})
			var output bytes.Buffer
			go func() {
				defer close(done)
				scanner := bufio.NewScanner(stdout)
				scanner.Buffer(make([]byte, 4096), 1<<20)
				started := false
				for scanner.Scan() {
					line := scanner.Text()
					output.WriteString(line + "\n")
					if strings.Contains(line, "Done (") && !started {
						started = true
						ready <- true
					}
				}
				if !started {
					ready <- false
				}
			}()
			started := <-ready
			if started {
				_, _ = stdin.Write([]byte(fmt.Sprintf("forceload add %d %d\nsave-all flush\nstop\n", cx*16, cz*16)))
			}
			stdin.Close()
			waitErr := command.Wait()
			<-done
			if !started || waitErr != nil {
				t.Fatalf("vanilla did not load %s save: %v\n%s", name, waitErr, output.String())
			}
			{
				chunkName, root, exists, err := anvil.ReadChunk(filepath.Join(worldDir, "region"), cx, cz)
				if err != nil || !exists {
					t.Fatalf("vanilla lost generated %s chunk: %v", name, err)
				}
				loaded, err := chunkFromAnvil(chunkName, root, cx, cz)
				if err != nil || loaded.GetBlock(0, 0, markerY) != StateStone {
					t.Fatalf("vanilla did not preserve generated block: %v", err)
				}
				if loaded.biomes != chunk.biomes {
					t.Fatal("vanilla changed Atlas-generated quart biomes")
				}
				for position, id := range trees {
					if loaded.GetBlock(int(position[0]), int(position[2]), position[1]) != id {
						t.Fatalf("vanilla changed generated tree/vegetation state at %+v", position)
					}
				}
				if !flat {
					t.Logf("official server loaded, saved and preserved %d generated tree/vegetation blocks", len(trees))
				}
			}
		})
	}
}
