package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPropertiesCreateAndRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.properties")
	p, err := Load(path)
	if err != nil || p.ServerPort != 25565 || p.MaxPlayers != 20 || p.ViewDistance != 10 || p.SimulationDistance != 10 || p.LevelType != "minecraft:normal" {
		t.Fatalf("defaults: %+v, %v", p, err)
	}
	contents := "# original comment\nserver-port = 25570\nmax-players: 7\n" +
		"motd=Welcome\\ to\\ Atlas\\u0021\nlevel-name=saves/test\n" +
		"level-seed=hello\nlevel-type=minecraft:flat\nonline-mode=true\nview-distance=5\nsimulation-distance=3\nunsupported-key=kept\n"
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	p, err = Load(path)
	if err != nil || p.ServerPort != 25570 || p.MaxPlayers != 7 || !p.OnlineMode || p.ViewDistance != 5 || p.SimulationDistance != 3 || p.MOTD != "Welcome to Atlas!" || p.LevelName != "saves/test" || p.LevelSeed != "hello" || p.LevelType != "minecraft:flat" {
		t.Fatalf("parsed: %+v, %v", p, err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != contents {
		t.Fatal("loading rewrote an existing properties file")
	}
}

func TestDistancePropertiesRejectUnsupportedValues(t *testing.T) {
	for _, input := range []string{"view-distance=1\n", "view-distance=17\n", "simulation-distance=0\n", "simulation-distance=32\n"} {
		path := filepath.Join(t.TempDir(), "server.properties")
		if err := os.WriteFile(path, []byte(input), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatalf("accepted unsupported distance %q", input)
		}
	}
}

func TestGenerationProperties(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.properties")
	p, err := Load(path)
	if err != nil || p.SpawnPreloadRadius != 3 || p.ChunkGenerationWorkers != 0 {
		t.Fatalf("generation defaults: %+v %v", p, err)
	}
	for _, contents := range []string{"spawn-preload-radius=0\nchunk-generation-workers=1\n", "spawn-preload-radius=16\nchunk-generation-workers=256\n", "spawn-preload-radius=2\nchunk-generation-workers=4\n"} {
		if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
		p, err = Load(path)
		if err != nil || len(p.Unsupported) != 0 {
			t.Fatalf("generation settings: %+v %v", p, err)
		}
	}
	for _, contents := range []string{"spawn-preload-radius=-1\n", "spawn-preload-radius=17\n", "spawn-preload-radius=fast\n", "chunk-generation-workers=-1\n", "chunk-generation-workers=257\n", "chunk-generation-workers=fast\n"} {
		if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatalf("accepted %q", contents)
		}
	}
}

func TestPropertiesAcceptUTF8BOM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.properties")
	if err := os.WriteFile(path, []byte("\ufeffserver-port=25570\nview-distance=6\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := Load(path)
	if err != nil || p.ServerPort != 25570 || p.ViewDistance != 6 || len(p.Unsupported) != 0 {
		t.Fatalf("BOM-prefixed properties: %+v, %v", p, err)
	}
}

func TestPropertiesRejectInvalidOnlineMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.properties")
	if err := os.WriteFile(path, []byte("online-mode=maybe\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("invalid online-mode value was accepted")
	}
}

func TestJavaStringSeed(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int64
	}{
		{"123456", 123456}, {"hello", 99162322}, {"-42", -42},
	} {
		got, err := ParseSeed(tc.text)
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %d, want %d (%v)", tc.text, got, tc.want, err)
		}
	}
}
