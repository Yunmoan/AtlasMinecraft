package server

import (
	"fmt"
	"math"
	"runtime"
	"runtime/metrics"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	targetTPS       = 20
	tickHistorySize = targetTPS*15*60 + 1
)

type tickSample struct {
	elapsed  time.Duration // monotonic time since the clock started
	work     time.Duration
	interval time.Duration // between completed iterations, including waits/saves
}

type performanceMetrics struct {
	mu           sync.Mutex
	boot         time.Time
	clockStarted time.Time
	samples      []tickSample
	next         int
	count        int
	total        uint64
	last         time.Duration
	saves        uint64
	lastSave     time.Duration
	maxSave      time.Duration
	cpuAt        time.Time
	cpuUsage     processUsage
	cpuPercent   float64
	cpuWindow    time.Duration
}

type tickStats struct {
	tps         [3]float64
	observed    time.Duration
	hasTicks    bool
	count       int
	average     time.Duration
	p95         time.Duration
	max         time.Duration
	overBudget  int
	maxInterval time.Duration
	lastAge     time.Duration
	saves       uint64
	lastSave    time.Duration
	maxSave     time.Duration
}

func newPerformanceMetrics(now time.Time) *performanceMetrics {
	return &performanceMetrics{boot: now, cpuAt: now, cpuUsage: readProcessUsage()}
}

func (m *performanceMetrics) startClock(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clockStarted = now
	m.samples = make([]tickSample, tickHistorySize)
	m.next, m.count, m.total, m.last = 0, 0, 0, 0
}

func (m *performanceMetrics) recordTick(end time.Time, work time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.clockStarted.IsZero() {
		return
	}
	elapsed := end.Sub(m.clockStarted)
	if elapsed <= m.last {
		return
	}
	m.samples[m.next] = tickSample{elapsed: elapsed, work: max(0, work), interval: elapsed - m.last}
	m.next = (m.next + 1) % len(m.samples)
	m.count = min(m.count+1, len(m.samples))
	m.total++
	m.last = elapsed
}

func (m *performanceMetrics) snapshot(now time.Time) tickStats {
	m.mu.Lock()
	s := tickStats{hasTicks: m.total > 0, saves: m.saves, lastSave: m.lastSave, maxSave: m.maxSave}
	if m.clockStarted.IsZero() {
		m.mu.Unlock()
		return s
	}
	elapsed := max(0, now.Sub(m.clockStarted))
	s.observed = min(elapsed, 15*time.Minute)
	s.lastAge = max(0, elapsed-m.last)
	windows := [3]time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute}
	var counts [3]int
	work := make([]time.Duration, 0, min(m.count, targetTPS*60))
	for i := 0; i < m.count; i++ {
		sample := m.samples[i]
		if sample.elapsed > elapsed {
			continue
		}
		for j, window := range windows {
			if sample.elapsed > elapsed-window {
				counts[j]++
			}
		}
		if sample.elapsed > elapsed-time.Minute {
			work = append(work, sample.work)
			s.average += sample.work
			s.max = max(s.max, sample.work)
			s.maxInterval = max(s.maxInterval, sample.interval)
			if sample.work > time.Second/targetTPS {
				s.overBudget++
			}
		}
	}
	m.mu.Unlock()
	for i, window := range windows {
		if observed := min(elapsed, window); observed > 0 {
			s.tps[i] = min(targetTPS, float64(counts[i])/observed.Seconds())
		}
	}
	s.count = len(work)
	if len(work) > 0 {
		s.average /= time.Duration(len(work))
		sort.Slice(work, func(i, j int) bool { return work[i] < work[j] })
		s.p95 = work[int(math.Ceil(float64(len(work))*.95))-1]
	}
	return s
}

// StartTickMetrics marks the beginning of the actual world-clock loop. Startup
// preparation is excluded, and short runs use their observed elapsed time.
func (s *Server) StartTickMetrics() {
	s.performance.startClock(time.Now())
}

// RecordTick counts one completed loop, irrespective of how many world-time
// ticks were advanced to catch up. Catch-up must not disguise a stalled loop.
func (s *Server) RecordTick(start time.Time) {
	end := time.Now()
	s.performance.recordTick(end, end.Sub(start))
}

func (s *Server) RecordAutosave(start time.Time) {
	duration := time.Since(start)
	m := s.performance
	m.mu.Lock()
	m.saves++
	m.lastSave = duration
	m.maxSave = max(m.maxSave, duration)
	m.mu.Unlock()
}

func (s tickStats) tpsText() string {
	if !s.hasTicks {
		return "TPS (1m/5m/15m): collecting samples (target: 20)"
	}
	return fmt.Sprintf("TPS (1m/5m/15m): %.2f / %.2f / %.2f (target: 20; observed: %s)", s.tps[0], s.tps[1], s.tps[2], s.observed.Round(time.Millisecond))
}

func (s tickStats) msptText() string {
	if !s.hasTicks {
		return "MSPT work (last 1m): collecting samples"
	}
	if s.count == 0 {
		return fmt.Sprintf("MSPT work (last 1m): no completed ticks; last tick age: %s; autosaves: %d, last/max: %.3f / %.3f ms", s.lastAge.Round(time.Millisecond), s.saves, milliseconds(s.lastSave), milliseconds(s.maxSave))
	}
	return fmt.Sprintf("MSPT work (last 1m, avg/P95/max): %.3f / %.3f / %.3f ms; >50ms: %d/%d\nLoop gap max: %.3f ms; last tick age: %s; autosaves: %d, last/max: %.3f / %.3f ms",
		milliseconds(s.average), milliseconds(s.p95), milliseconds(s.max), s.overBudget, s.count,
		milliseconds(s.maxInterval), s.lastAge.Round(time.Millisecond), s.saves, milliseconds(s.lastSave), milliseconds(s.maxSave))
}

func milliseconds(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

const performanceScope = "Scope: world-clock loop; generation and player handlers run separately."

func (s *Server) commandPerformance(command string) string {
	now := time.Now()
	stats := s.performance.snapshot(now)
	switch command {
	case "tps":
		return stats.tpsText() + "\n" + performanceScope
	case "mspt":
		return stats.msptText() + "\n" + performanceScope
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	heap := []metrics.Sample{{Name: "/gc/heap/live:bytes"}, {Name: "/gc/heap/goal:bytes"}}
	metrics.Read(heap)
	usage, percent, cpuWindow := s.performance.sampleProcess(now)
	cpu := "unavailable"
	if usage.cpuOK {
		cpu = "collecting samples"
		if cpuWindow > 0 {
			cpu = fmt.Sprintf("%.1f%% of one core over %s", percent, cpuWindow.Round(time.Millisecond))
		}
	}
	resident := "unavailable"
	if usage.residentOK {
		resident = fmt.Sprintf("%.1f MiB", float64(usage.resident)/float64(1<<20))
	}
	lastPause := uint64(0)
	if mem.NumGC > 0 {
		lastPause = mem.PauseNs[(mem.NumGC-1)%256]
	}
	lines := []string{
		fmt.Sprintf("Atlas status: uptime %s; players %d/%d; goroutines %d", now.Sub(s.performance.boot).Round(time.Second), s.Hub.Count(), s.Cfg.MaxPlayers, runtime.NumGoroutine()),
		stats.tpsText(), stats.msptText(),
		fmt.Sprintf("Process CPU: %s; logical CPUs: %d; Go parallelism: %d", cpu, runtime.NumCPU(), runtime.GOMAXPROCS(0)),
		fmt.Sprintf("Memory: heap allocated %.1f MiB; heap in use %.1f MiB; Go used/reserved %.1f / %.1f MiB; resident working set %s",
			float64(mem.HeapAlloc)/float64(1<<20), float64(mem.HeapInuse)/float64(1<<20), float64(mem.Sys-mem.HeapReleased)/float64(1<<20), float64(mem.Sys)/float64(1<<20), resident),
		fmt.Sprintf("GC: %d collections; pause last/total %.3f / %.3f ms", mem.NumGC, float64(lastPause)/1e6, float64(mem.PauseTotalNs)/1e6),
		fmt.Sprintf("GC heap: live at last collection %.1f MiB; next collection goal %.1f MiB", float64(heap[0].Value.Uint64())/float64(1<<20), float64(heap[1].Value.Uint64())/float64(1<<20)),
		fmt.Sprintf("Runtime: %s %s/%s; view/simulation distance: %d/%d", runtime.Version(), runtime.GOOS, runtime.GOARCH, s.Cfg.ViewDistance, s.Cfg.SimulationDistance),
	}
	if s.World != nil {
		w := s.World.Stats()
		lines = append(lines, fmt.Sprintf("World: cached chunks %d; light columns %d; region actors %d; generation active/workers %d/%d; queued jobs %d; mailbox messages %d",
			w.LoadedChunks, w.LightColumns, w.Regions, w.ActiveGeneration, w.GenerationWorkers, w.QueuedGeneration, w.MailboxMessages))
	}
	return strings.Join(append(lines, performanceScope), "\n")
}

type processUsage struct {
	cpu        time.Duration
	resident   uint64
	cpuOK      bool
	residentOK bool
}

// CPU usage is the change in process kernel+user time divided by wall time.
// All goroutines are included; 100% means one core, and multicore work can
// exceed 100%. Repeated queries within 250ms reuse the last useful sample.
func (m *performanceMetrics) sampleProcess(now time.Time) (processUsage, float64, time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	usage := readProcessUsage()
	if elapsed := now.Sub(m.cpuAt); elapsed >= 250*time.Millisecond {
		if usage.cpuOK && m.cpuUsage.cpuOK && usage.cpu >= m.cpuUsage.cpu {
			m.cpuPercent = 100 * float64(usage.cpu-m.cpuUsage.cpu) / float64(elapsed)
			m.cpuWindow = elapsed
		} else {
			m.cpuWindow = 0
		}
		m.cpuAt, m.cpuUsage = now, usage
	}
	return usage, m.cpuPercent, m.cpuWindow
}
