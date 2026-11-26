package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEngineValidationAndSourceFromText(t *testing.T) {
	if !validEngine("corenlp") || !validEngine("go") {
		t.Fatal("corenlp and go must be valid engines")
	}
	for _, bad := range []string{"", "java", "Go", "fast"} {
		if validEngine(bad) {
			t.Errorf("%q should not be an engine", bad)
		}
	}
	// "One paragraph." is 14 bytes; the second paragraph is 37 bytes (é is two,
	// 𐐷 is four); a run of three newlines holds two blank-line positions.
	src := sourceFromText("t", "One paragraph.\n\nSecond, with a café and a 𐐷 rune.\n\n\nThird.")
	want := []int{0, 16, 55, 56}
	if len(src.ParaStart) != len(want) {
		t.Fatalf("paragraph starts %v, want %v", src.ParaStart, want)
	}
	for i, off := range want {
		if src.ParaStart[i] != off {
			t.Fatalf("paragraph %d starts at %d, want %d", i, src.ParaStart[i], off)
		}
	}
	// Paragraph starts are byte offsets into Text: each one is the first byte of a rune.
	for _, off := range src.ParaStart[1:] {
		if src.Text[off-2] != '\n' || src.Text[off-1] != '\n' {
			t.Errorf("paragraph start %d does not follow a blank line", off)
		}
	}
	if parseChapters("ch13, ch14,,ch02")["ch14"] != true || len(parseChapters(" ch13,ch14,,ch02 ")) != 3 {
		t.Error("chapter list parsing")
	}
}

func TestCachedText(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{"in", "out", "scratch"} {
		if err := os.MkdirAll(filepath.Join(dir, "ext-ref", "linear", sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := cachedText(dir, "ch01"); err == nil {
		t.Fatal("a chapter without a cached input should fail")
	}
	// Chapters: the text CoreNLP parsed is kept under in/ (out/ holds the parses).
	if err := os.WriteFile(filepath.Join(dir, "ext-ref", "linear", "out", "ch01.txt"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ext-ref", "linear", "in", "ch01.txt"), []byte("Chapter."), 0o644); err != nil {
		t.Fatal(err)
	}
	if text, err := cachedText(dir, "ch01"); err != nil || text != "Chapter." {
		t.Fatalf("chapter input: %q, %v", text, err)
	}
	// Drafts checked earlier live under scratch/ by hash.
	if err := os.WriteFile(filepath.Join(dir, "ext-ref", "linear", "scratch", "abc.txt"), []byte("Draft."), 0o644); err != nil {
		t.Fatal(err)
	}
	if text, err := cachedText(dir, "abc"); err != nil || text != "Draft." {
		t.Fatalf("scratch fallback: %q, %v", text, err)
	}
}
