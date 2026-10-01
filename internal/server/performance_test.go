package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"math"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTPSStartupAndStall(t *testing.T) {
	start := time.Unix(1000, 0)
	m := newPerformanceMetrics(start)
	if s := m.snapshot(start); s.hasTicks || !strings.Contains(s.tpsText(), "collecting") {
		t.Fatalf("unstarted clock reports TPS: %+v", s)
	}
	m.startClock(start)
	for i := 1; i <= 10; i++ {
		m.recordTick(start.Add(time.Duration(i)*50*time.Millisecond), time.Millisecond)
	}
	s := m.snapshot(start.Add(500 * time.Millisecond))
	for _, tps := range s.tps {
		if tps != 20 {
			t.Fatalf("short sample biased by a full window: %+v", s)
		}
	}
	// World-time catch-up advances many ticks, but there is still only one
	// completed loop after a stall. Idle queries include time since that loop.
	m.recordTick(start.Add(10*time.Second), 2*time.Millisecond)
	s = m.snapshot(start.Add(10 * time.Second))
	if math.Abs(s.tps[0]-1.1) > 1e-9 || s.maxInterval != 9500*time.Millisecond {
		t.Fatalf("stall hidden by catch-up: %+v", s)
	}
	s = m.snapshot(start.Add(70 * time.Second))
	if s.tps[0] != 0 || s.count != 0 || s.lastAge != time.Minute {
		t.Fatalf("stalled server retains healthy TPS: %+v", s)
	}
}

func TestTPSRollingWindowsAndHistoryWrap(t *testing.T) {
	start := time.Unix(1000, 0)
	m := newPerformanceMetrics(start)
	m.startClock(start)
	for i := 1; i <= 20*16*60; i++ {
		m.recordTick(start.Add(time.Duration(i)*50*time.Millisecond), time.Millisecond)
	}
	s := m.snapshot(start.Add(16 * time.Minute))
	if s.tps != [3]float64{20, 20, 20} || s.count != 1200 {
		t.Fatalf("wrapped history lost recent samples: %+v", s)
	}
	s = m.snapshot(start.Add(17 * time.Minute))
	want := [3]float64{0, 16, 20 * 14.0 / 15.0}
	for i, tps := range s.tps {
		if math.Abs(tps-want[i]) > 1e-9 {
			t.Fatalf("window %d TPS=%f, want %f", i, tps, want[i])
		}
	}
	if m.count != tickHistorySize {
		t.Fatal("history is not bounded")
	}
}

func TestMSPTDistributionAndConcurrentSnapshots(t *testing.T) {
	start := time.Unix(1000, 0)
	m := newPerformanceMetrics(start)
	m.startClock(start)
	for i := 1; i <= 100; i++ {
		m.recordTick(start.Add(time.Duration(i)*200*time.Millisecond), time.Duration(i)*time.Millisecond)
	}
	s := m.snapshot(start.Add(20 * time.Second))
	if s.average != 50500*time.Microsecond || s.p95 != 95*time.Millisecond || s.max != 100*time.Millisecond || s.overBudget != 50 {
		t.Fatalf("incorrect latency distribution: %+v", s)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 101; i <= 1000; i++ {
			m.recordTick(start.Add(time.Duration(i)*200*time.Millisecond), time.Millisecond)
		}
	}()
	go func() {
		defer wg.Done()
		for range 100 {
			m.snapshot(start.Add(201 * time.Second))
		}
	}()
	wg.Wait()
}

func TestPerformanceCommandsPermissionsAndReplies(t *testing.T) {
	h := hub.New()
	w := world.New(h, 1)
	t.Cleanup(func() { w.Close() })
	s := New(Config{MaxPlayers: 20, Ops: map[string]bool{"admin": true}}, h, w)
	guest := player.New("Guest", [16]byte{1}, 1, nil)
	admin := player.New("Admin", [16]byte{2}, 2, nil)
	want := map[string]string{"tps": "TPS (1m/5m/15m)", "mspt": "MSPT work", "status": "Atlas status:", "sysinfo": "Atlas status:"}
	for command, prefix := range want {
		if got := s.executeCommand(guest, command); !strings.Contains(got, "do not have permission") {
			t.Fatalf("guest can query %s: %s", command, got)
		}
		for _, source := range []*player.Player{nil, admin} {
			if got := s.executeCommand(source, "/"+command); !strings.HasPrefix(got, prefix) {
				t.Fatalf("%s response: %s", command, got)
			}
			if got := s.executeCommand(source, command+" extra"); got != "Usage: /"+command {
				t.Fatalf("%s invalid arguments: %s", command, got)
			}
		}
		for _, level := range []int{0, 1, 2, 3, 4} {
			if present := strings.Contains(string(declareCommandsPayloadForLevel(level)), command); present != (level >= 2) {
				t.Fatalf("level %d command tree mismatch for %s", level, command)
			}
		}
	}
	s.performance.startClock(time.Now().Add(-time.Second))
	s.RecordTick(time.Now().Add(-time.Millisecond))
	s.RecordAutosave(time.Now().Add(-20 * time.Millisecond))
	s.RunCommand(admin, "status")
	packet := <-admin.Out
	if packet.ID != proto.SSystemChat {
		t.Fatalf("performance result not delivered via chat: %d", packet.ID)
	}
	for _, text := range []string{"Memory:", "GC:", "Process CPU:", "Runtime:", "World: cached chunks 0", "autosaves: 1", performanceScope} {
		if !strings.Contains(string(packet.Payload), text) {
			t.Fatalf("missing %q in status packet: %s", text, packet.Payload)
		}
	}
	if runtime.GOOS == "windows" {
		usage := readProcessUsage()
		if !usage.cpuOK || !usage.residentOK || usage.resident == 0 {
			t.Fatalf("Windows process counters unavailable: %+v", usage)
		}
	}
}
