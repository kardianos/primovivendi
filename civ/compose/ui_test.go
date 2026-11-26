package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bankai-tensei/gowaku/gowakutest"
)

func TestGUILayoutAcrossAllThreeViewsAndPasses(t *testing.T) {
	root := t.TempDir()
	chapDir := filepath.Join(root, "civ", "chapters")
	if err := os.MkdirAll(chapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# Chapter One\n\nWe come into this world naked. The bridge stands firm; the valley endures. Good builders keep clear order.\n"
	if err := os.WriteFile(filepath.Join(chapDir, "01_ch01.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngine(root, filepath.Join(root, "civ", "notes", "compose"), false)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	sess, err := NewSession(eng, "ch01")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	tt := gowakutest.New(t, composeApp{Session: sess}, gowakutest.Options{Width: 1440, Height: 900})

	// 1. Default Lattice view renders without panic and shows paragraph + lattice header.
	if dump := tt.Dump(); !strings.Contains(dump, "Word Lattice") {
		t.Fatalf("expected Word Lattice in initial view:\n%s", dump)
	}

	// 2. Switch to View 2 (Dependency Syntax Diagram).
	tt.Tap("2. Diagram")
	if dump := tt.Dump(); !strings.Contains(dump, "Dependency Syntax Diagram") {
		t.Fatalf("expected Dependency Syntax Diagram after tapping 2. Diagram:\n%s", dump)
	}

	// 3. Switch to View 3 (Hybrid: Diagram + Lattice).
	tt.Tap("3. Hybrid")
	dump := tt.Dump()
	if !strings.Contains(dump, "Syntax Dependency Diagram") || !strings.Contains(dump, "Word Lattice") {
		t.Fatalf("expected both diagram and lattice in Hybrid view:\n%s", dump)
	}

	// 4. Switch passes (Forward, Loop) and back to Lattice.
	tt.Tap("Forward (F)")
	tt.Tap("1. Lattice")
	if dump := tt.Dump(); !strings.Contains(dump, "Goal (forward)") {
		t.Fatalf("expected Forward pass goal in dump:\n%s", dump)
	}

	// 5. Select a multi-word phrase, open the separate Boost / De-weight Chooser,
	// toggle Expand 2-Hop, scroll horizontally, then change sentences.
	sess.SelectSlotHeader(0)
	sess.SelectSlotHeader(2)
	tt.Tap("▶")
	dump = tt.Dump()
	if !strings.Contains(dump, "Phrase (3 words):") || !strings.Contains(dump, "Lattice horizontal scroll") || !strings.Contains(dump, "Expand 2-Hop") {
		t.Fatalf("expected Phrase Bar, Expand 2-Hop, and horizontal scroll bar in dump:\n%s", dump)
	}
	tt.Tap("Weights +/−")
	if dump := tt.Dump(); !strings.Contains(dump, "Boost / De-weight Chooser:") || !strings.Contains(dump, "OE+/Grk+/Lat−: ON") {
		t.Fatalf("expected Boost / De-weight Chooser with etymology toggle:\n%s", dump)
	}
	tt.Tap("Accept Paragraph & Next (Shift+A)")
	if dump := tt.Dump(); !strings.Contains(dump, "✓ P1") {
		t.Fatalf("expected completed paragraph badge '✓ P1' after Accept Paragraph:\n%s", dump)
	}
	tt.Tap("Next Sent (J)")
	tt.Tap("Prev Sent (K)")
}
