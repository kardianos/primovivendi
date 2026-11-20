package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestWordDiffEmDash(t *testing.T) {
	got := WordDiff(
		"The work — the real work — is keeping what was built.",
		"The work, the real work, is keeping what was built.",
	)
	var deleted, inserted string
	for _, p := range got {
		switch p.Op {
		case DiffDelete:
			deleted += p.Text
		case DiffInsert:
			inserted += p.Text
		}
	}
	if !strings.Contains(deleted, "—") {
		t.Fatalf("deleted %q, pieces %#v", deleted, got)
	}
	if !strings.Contains(inserted, ",") {
		t.Fatalf("inserted %q, pieces %#v", inserted, got)
	}
	if strings.Contains(deleted, "keeping") || strings.Contains(inserted, "keeping") {
		t.Fatalf("unchanged words were marked: %#v", got)
	}
}

func TestWordDiffIdentical(t *testing.T) {
	got := WordDiff("Same line.", "Same line.")
	if len(got) != 1 || got[0].Op != DiffEqual || got[0].Text != "Same line." {
		t.Fatalf("%#v", got)
	}
}

func TestDiffFallsBackWhenLarge(t *testing.T) {
	a := make([]string, 1001)
	b := make([]string, 1001)
	for i := range a {
		a[i] = fmt.Sprintf("a%d", i)
		b[i] = fmt.Sprintf("b%d", i)
	}
	got := diffSeq(a, b)
	if len(got) != 2 || got[0].Op != DiffDelete || got[1].Op != DiffInsert {
		t.Fatalf("len %d %#v", len(got), got[:min(len(got), 3)])
	}
	if got[0].Text != strings.Join(a, "") || got[1].Text != strings.Join(b, "") {
		t.Fatal("fallback did not keep the whole run")
	}
}
