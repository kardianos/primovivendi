// Package prof provides a lightweight, opt-in span profiler that periodically
// writes timing breakdowns, slow-call traces, and Go runtime memory/goroutine
// snapshots to a file so interactive latency bottlenecks can be diagnosed live.
package prof

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var enabled atomic.Bool

type stat struct {
	count   int64
	totalNs int64
	maxNs   int64
	lastNs  int64
}

type slowEvent struct {
	at     time.Time
	name   string
	dur    time.Duration
	detail string
}

type profiler struct {
	mu       sync.Mutex
	path     string
	interval time.Duration
	started  time.Time
	stats    map[string]*stat
	slow     []slowEvent
	stopCh   chan struct{}
	doneCh   chan struct{}
}

var (
	globalMu sync.Mutex
	global   *profiler
)

// Enable activates periodic profiling to path every interval (default 5s).
// Each tick appends a timestamped summary table + recent slow spans (>=50ms)
// so the file can be inspected live with `tail -f`.
func Enable(path string, interval time.Duration) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Truncate/initialize the file for a clean session header.
	hdr := fmt.Sprintf("# civ/compose profiler started %s (interval=%s)\n\n",
		time.Now().Format(time.RFC3339), interval)
	if err := os.WriteFile(path, []byte(hdr), 0o644); err != nil {
		return err
	}

	globalMu.Lock()
	if global != nil {
		old := global
		global = nil
		enabled.Store(false)
		globalMu.Unlock()
		old.shutdown()
		globalMu.Lock()
	}
	p := &profiler{
		path:     path,
		interval: interval,
		started:  time.Now(),
		stats:    map[string]*stat{},
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
	global = p
	enabled.Store(true)
	globalMu.Unlock()

	go p.loop()
	return nil
}

// Stop flushes a final profiling snapshot and stops the background writer.
func Stop() {
	globalMu.Lock()
	p := global
	global = nil
	enabled.Store(false)
	globalMu.Unlock()
	if p != nil {
		p.shutdown()
	}
}

// Enabled reports whether profiling is currently active.
func Enabled() bool { return enabled.Load() }

// Span begins timing name and returns a finish function suitable for defer.
// When profiling is disabled this is a single atomic load and a no-op closure.
func Span(name string) func() {
	if !enabled.Load() {
		return noop
	}
	start := time.Now()
	return func() {
		record(name, time.Since(start), "")
	}
}

// SpanDetail is like Span but attaches a short detail string to slow events.
func SpanDetail(name, detail string) func() {
	if !enabled.Load() {
		return noop
	}
	start := time.Now()
	return func() {
		record(name, time.Since(start), detail)
	}
}

// Count increments a lightweight counter (recorded as a 0ns span).
func Count(name string) {
	if !enabled.Load() {
		return
	}
	record(name, 0, "")
}

func noop() {}

func record(name string, dur time.Duration, detail string) {
	globalMu.Lock()
	p := global
	globalMu.Unlock()
	if p == nil {
		return
	}
	ns := dur.Nanoseconds()
	p.mu.Lock()
	st, ok := p.stats[name]
	if !ok {
		st = &stat{}
		p.stats[name] = st
	}
	st.count++
	st.totalNs += ns
	st.lastNs = ns
	if ns > st.maxNs {
		st.maxNs = ns
	}
	if dur >= 50*time.Millisecond {
		p.slow = append(p.slow, slowEvent{
			at:     time.Now(),
			name:   name,
			dur:    dur,
			detail: detail,
		})
		if len(p.slow) > 40 {
			p.slow = p.slow[len(p.slow)-40:]
		}
	}
	p.mu.Unlock()
}

func (p *profiler) loop() {
	defer close(p.doneCh)
	tk := time.NewTicker(p.interval)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			p.flush(false)
		case <-p.stopCh:
			p.flush(true)
			return
		}
	}
}

func (p *profiler) shutdown() {
	close(p.stopCh)
	<-p.doneCh
}

func (p *profiler) flush(final bool) {
	type row struct {
		name string
		st   stat
	}
	p.mu.Lock()
	rows := make([]row, 0, len(p.stats))
	for k, v := range p.stats {
		rows = append(rows, row{name: k, st: *v})
	}
	slow := append([]slowEvent(nil), p.slow...)
	p.slow = p.slow[:0]
	p.mu.Unlock()

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].st.totalNs != rows[j].st.totalNs {
			return rows[i].st.totalNs > rows[j].st.totalNs
		}
		return rows[i].name < rows[j].name
	})

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	var sb strings.Builder
	tag := "TICK"
	if final {
		tag = "FINAL"
	}
	fmt.Fprintf(&sb, "=== [%s] %s (uptime %s · goroutines=%d · heap=%.1fMB · gc=%d) ===\n",
		tag,
		time.Now().Format("15:04:05.000"),
		time.Since(p.started).Round(100*time.Millisecond),
		runtime.NumGoroutine(),
		float64(ms.HeapAlloc)/(1024*1024),
		ms.NumGC,
	)
	if len(rows) == 0 {
		sb.WriteString("  (no spans recorded yet)\n")
	} else {
		fmt.Fprintf(&sb, "  %-34s %7s %10s %10s %10s %10s\n", "SPAN", "COUNT", "TOTAL", "AVG", "MAX", "LAST")
		for _, r := range rows {
			tot := time.Duration(r.st.totalNs)
			avg := time.Duration(0)
			if r.st.count > 0 {
				avg = time.Duration(r.st.totalNs / r.st.count)
			}
			mx := time.Duration(r.st.maxNs)
			last := time.Duration(r.st.lastNs)
			fmt.Fprintf(&sb, "  %-34s %7d %10s %10s %10s %10s\n",
				r.name, r.st.count, fmtDur(tot), fmtDur(avg), fmtDur(mx), fmtDur(last))
		}
	}
	if len(slow) > 0 {
		sb.WriteString("  Recent slow calls (>=50ms):\n")
		for _, e := range slow {
			if e.detail != "" {
				fmt.Fprintf(&sb, "    %s  %-28s %8s  %s\n", e.at.Format("15:04:05.000"), e.name, fmtDur(e.dur), e.detail)
			} else {
				fmt.Fprintf(&sb, "    %s  %-28s %8s\n", e.at.Format("15:04:05.000"), e.name, fmtDur(e.dur))
			}
		}
	}
	sb.WriteByte('\n')

	f, err := os.OpenFile(p.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(sb.String())
	_ = f.Close()
}

func fmtDur(d time.Duration) string {
	switch {
	case d == 0:
		return "0ms"
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fms", float64(d)/float64(time.Millisecond))
	case d < 10*time.Second:
		return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond))
	default:
		return d.Round(10 * time.Millisecond).String()
	}
}
