package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// VanillaHandler writes one readable line per event. The shared lock keeps
// concurrent goroutines from interleaving lines on the console and in files.
type VanillaHandler struct {
	state  *vanillaOutput
	level  slog.Level
	source bool
	attrs  []boundAttr
	group  string
}

type boundAttr struct {
	attr  slog.Attr
	group string
}

type vanillaOutput struct {
	mu sync.Mutex
	w  io.Writer
}

func NewVanillaHandler(w io.Writer, level slog.Level, source bool) *VanillaHandler {
	return &VanillaHandler{state: &vanillaOutput{w: w}, level: level, source: source}
}

func (h *VanillaHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *VanillaHandler) Handle(_ context.Context, r slog.Record) error {
	if !h.Enabled(context.Background(), r.Level) {
		return nil
	}
	var attrs []string
	add := func(a slog.Attr, group string) {
		if a.Key == "component" && group == "" {
			return
		}
		appendReadableAttr(&attrs, a, group)
	}
	for _, a := range h.attrs {
		add(a.attr, a.group)
	}
	r.Attrs(func(a slog.Attr) bool { add(a, h.group); return true })
	stamp := r.Time
	if stamp.IsZero() {
		stamp = time.Now()
	}
	level := strings.ToUpper(r.Level.String())
	line := fmt.Sprintf("[%s] [Server thread/%s]: %s", stamp.Local().Format("15:04:05"), level, oneLine(r.Message))
	if len(attrs) > 0 {
		line += " (" + strings.Join(attrs, ", ") + ")"
	}
	if h.source && r.PC != 0 {
		if frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next(); frame.File != "" {
			line += fmt.Sprintf(" [%s:%d]", filepath.Base(frame.File), frame.Line)
		}
	}
	line += "\n"
	h.state.mu.Lock()
	_, err := io.WriteString(h.state.w, line)
	h.state.mu.Unlock()
	return err
}

func appendReadableAttr(dst *[]string, a slog.Attr, group string) {
	value := a.Value.Resolve()
	if value.Kind() == slog.KindGroup {
		next := a.Key
		if group != "" {
			next = group + "." + next
		}
		for _, nested := range value.Group() {
			appendReadableAttr(dst, nested, next)
		}
		return
	}
	key := a.Key
	if group != "" {
		key = group + "." + key
	}
	if key == "" {
		return
	}
	*dst = append(*dst, oneLine(key)+": "+oneLine(value.String()))
}

func oneLine(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func (h *VanillaHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := *h
	next.attrs = append([]boundAttr(nil), h.attrs...)
	for _, a := range attrs {
		next.attrs = append(next.attrs, boundAttr{attr: a, group: h.group})
	}
	return &next
}

func (h *VanillaHandler) WithGroup(name string) slog.Handler {
	next := *h
	if next.group != "" {
		next.group += "."
	}
	next.group += name
	return &next
}
