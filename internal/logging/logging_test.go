package logging

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotatingFileKeepsBoundedBackups(t *testing.T) {
	path := filepath.Join(t.TempDir(), "latest.log")
	f, err := Open(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"first\n", "second\n", "third\n", "fourth\n", "fifth\n"} {
		if _, err := f.Write([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}
	f.Close()
	for _, suffix := range []string{"", ".1", ".2", ".3"} {
		if _, err := os.Stat(path + suffix); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(path + ".4"); !os.IsNotExist(err) {
		t.Fatalf("unexpected fourth backup: %v", err)
	}
}

func TestVanillaHandlerReadableAndSingleLine(t *testing.T) {
	var output bytes.Buffer
	h := NewVanillaHandler(&output, slog.LevelInfo, false)
	r := slog.NewRecord(time.Date(2026, 9, 30, 12, 34, 56, 0, time.Local), slog.LevelInfo, "<Alice> hello\nforged", 0)
	r.AddAttrs(slog.String("player", "Alice"), slog.String("component", "server"))
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	want := "[12:34:56] [Server thread/INFO]: <Alice> hello\\nforged (player: Alice)\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

func TestDebugLogIsSeparate(t *testing.T) {
	dir := t.TempDir()
	latest := filepath.Join(dir, "latest.log")
	debug := filepath.Join(dir, "debug.log")
	old := slog.Default()
	defer slog.SetDefault(old)
	closer, err := Setup(latest, debug, true)
	if err != nil {
		t.Fatal(err)
	}
	slog.Debug("packet detail")
	slog.Info("player joined")
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
	mainBytes, _ := os.ReadFile(latest)
	debugBytes, _ := os.ReadFile(debug)
	if strings.Contains(string(mainBytes), "packet detail") || !strings.Contains(string(mainBytes), "player joined") {
		t.Fatalf("unexpected main log: %s", mainBytes)
	}
	if !strings.Contains(string(debugBytes), "packet detail") || !strings.Contains(string(debugBytes), "player joined") {
		t.Fatalf("unexpected debug log: %s", debugBytes)
	}
}

func TestSetupStartsFreshReadableSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "latest.log")
	if err := os.WriteFile(path, []byte("old raw log\n"), 0644); err != nil {
		t.Fatal(err)
	}
	old := slog.Default()
	defer slog.SetDefault(old)
	closer, err := Setup(path, filepath.Join(t.TempDir(), "debug.log"), false)
	if err != nil {
		t.Fatal(err)
	}
	slog.Info("new session")
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
	latest, _ := os.ReadFile(path)
	previous, _ := os.ReadFile(path + ".1")
	if strings.Contains(string(latest), "old raw log") || !strings.Contains(string(latest), "[Server thread/INFO]: new session") || string(previous) != "old raw log\n" {
		t.Fatalf("unexpected session rotation: latest=%q previous=%q", latest, previous)
	}
}
