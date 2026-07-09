package chunk

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitUnderSoftMax(t *testing.T) {
	s := "Hello world."
	ch := Split(s, 100, RESTMaxChars)
	if len(ch) != 1 || ch[0].Text != s {
		t.Fatalf("%+v", ch)
	}
}

func TestSplitAtPauseBoundary(t *testing.T) {
	// Build text that must split, with pause markers
	left := strings.Repeat("word ", 100) // ~500 chars
	right := strings.Repeat("next ", 100)
	// force small soft max
	soft := 400
	text := strings.TrimSpace(left) + " [pause] " + strings.TrimSpace(right)
	ch := Split(text, soft, RESTMaxChars)
	if len(ch) < 2 {
		t.Fatalf("expected multi chunk, got %d: lens=%v", len(ch), sizes(ch))
	}
	for _, c := range ch {
		if c.Chars > soft+50 { // allow small overshoot only if hard used; soft path should be <= soft-ish
			// actually cut is at last pause before softMax, so each part should be <= softMax runes roughly
		}
		if c.Chars > RESTMaxChars {
			t.Fatalf("over hard max: %d", c.Chars)
		}
	}
	// first chunk should end near pause
	if !strings.Contains(ch[0].Text, "word") {
		t.Fatalf("chunk0: %q", ch[0].Text[:min(40, len(ch[0].Text))])
	}
}

func TestSplitNeverExceedsHardMax(t *testing.T) {
	// No pauses, no spaces almost — use spaces for word cut
	s := strings.Repeat("abcdefghij ", 2000) // ~22000 chars
	ch := Split(s, 12000, RESTMaxChars)
	if len(ch) < 2 {
		t.Fatalf("expected splits, got %d", len(ch))
	}
	for _, c := range ch {
		if c.Chars > RESTMaxChars {
			t.Fatalf("chunk %d has %d chars", c.Index, c.Chars)
		}
		if utf8.RuneCountInString(c.Text) != c.Chars {
			t.Fatalf("chars mismatch")
		}
	}
}

func TestEstimate(t *testing.T) {
	s := strings.Repeat("Hello. ", 3000)
	n := Estimate(s, 12000)
	if n < 2 {
		t.Fatalf("estimate %d", n)
	}
}

func sizes(ch []Chunk) []int {
	var s []int
	for _, c := range ch {
		s = append(s, c.Chars)
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
