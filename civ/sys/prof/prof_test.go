package prof

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPeriodicProfilerWritesReport(t *testing.T) {
	out := filepath.Join(t.TempDir(), "profile.log")
	if err := Enable(out, 40*time.Millisecond); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	done := SpanDetail("test.work", "sample detail")
	time.Sleep(55 * time.Millisecond)
	done()
	Count("test.counter")
	time.Sleep(60 * time.Millisecond)
	Stop()

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	txt := string(b)
	for _, want := range []string{"[TICK]", "[FINAL]", "test.work", "test.counter", "sample detail"} {
		if !strings.Contains(txt, want) {
			t.Fatalf("profile output missing %q:\n%s", want, txt)
		}
	}
}
