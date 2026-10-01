package logging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// RotatingFile keeps the active log plus three bounded backups. slog writes
// whole records, and Write holds the lock across rotation and append.
type RotatingFile struct {
	mu       sync.Mutex
	path     string
	file     *os.File
	size     int64
	maxBytes int64
}

func Open(path string, maxBytes int64) (*RotatingFile, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	return &RotatingFile{path: path, file: f, size: st.Size(), maxBytes: maxBytes}, nil
}

// openSession starts a fresh latest.log while retaining the previous session
// as .1, matching the way vanilla keeps old server runs separately.
func openSession(path string, maxBytes int64) (*RotatingFile, error) {
	r, err := Open(path, maxBytes)
	if err != nil {
		return nil, err
	}
	if r.size > 0 {
		if err := r.rotate(); err != nil {
			r.Close()
			return nil, err
		}
	}
	return r, nil
}

func (r *RotatingFile) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.maxBytes > 0 && r.size > 0 && r.size+int64(len(p)) > r.maxBytes {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	if r.file == nil {
		return 0, errors.New("log file unavailable after failed rotation")
	}
	n, err := r.file.Write(p)
	r.size += int64(n)
	return n, err
}

func (r *RotatingFile) rotate() (err error) {
	if err := r.file.Close(); err != nil {
		return err
	}
	r.file = nil
	defer func() {
		if err != nil {
			if f, reopenErr := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); reopenErr == nil {
				r.file = f
				if st, statErr := f.Stat(); statErr == nil {
					r.size = st.Size()
				}
			}
		}
	}()
	os.Remove(r.path + ".3")
	for i := 2; i >= 1; i-- {
		from := r.path + fmt.Sprintf(".%d", i)
		to := r.path + fmt.Sprintf(".%d", i+1)
		if err := os.Rename(from, to); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.Rename(r.path, r.path+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(r.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	r.file, r.size = f, 0
	return nil
}

func (r *RotatingFile) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

type fanoutHandler struct {
	normal slog.Handler
	debug  slog.Handler
}

func (h fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.normal.Enabled(ctx, level) || (h.debug != nil && h.debug.Enabled(ctx, level))
}

func (h fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var first error
	if h.normal.Enabled(ctx, r.Level) {
		first = h.normal.Handle(ctx, r)
	}
	if h.debug != nil && h.debug.Enabled(ctx, r.Level) {
		if err := h.debug.Handle(ctx, r); first == nil {
			first = err
		}
	}
	return first
}

func (h fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := fanoutHandler{normal: h.normal.WithAttrs(attrs)}
	if h.debug != nil {
		next.debug = h.debug.WithAttrs(attrs)
	}
	return next
}

func (h fanoutHandler) WithGroup(name string) slog.Handler {
	next := fanoutHandler{normal: h.normal.WithGroup(name)}
	if h.debug != nil {
		next.debug = h.debug.WithGroup(name)
	}
	return next
}

type closers []io.Closer

func (c closers) Close() error {
	var result error
	for _, item := range c {
		result = errors.Join(result, item.Close())
	}
	return result
}

// Setup writes INFO and above to console/latest.log. With debug enabled it
// also writes all levels, with source locations, to a separate debug log.
func Setup(path, debugPath string, debug bool) (io.Closer, error) {
	if debug && filepath.Clean(path) == filepath.Clean(debugPath) {
		return nil, errors.New("debug log path must differ from the main log path")
	}
	f, err := openSession(path, 10<<20)
	if err != nil {
		return nil, err
	}
	h := fanoutHandler{normal: NewVanillaHandler(io.MultiWriter(os.Stdout, f), slog.LevelInfo, false)}
	opened := closers{f}
	if debug {
		debugFile, err := openSession(debugPath, 10<<20)
		if err != nil {
			f.Close()
			return nil, err
		}
		opened = append(opened, debugFile)
		h.debug = NewVanillaHandler(debugFile, slog.LevelDebug, true)
	}
	slog.SetDefault(slog.New(h))
	return opened, nil
}
