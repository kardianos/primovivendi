package main

import (
	"strings"
	"testing"
)

func TestApplyLineMarkersNew(t *testing.T) {
	in := strings.Join([]string{
		"Keep me.",
		"%% edit note: soft voice",
		"%- Old paragraph that was too stiff.",
		"%+ New paragraph that sounds like a human.",
		"Also keep 50% mid-line alone.",
		"%+NoSpaceStillWorks",
		"",
	}, "\n")

	out, st := applyLineMarkers([]byte(in), false)
	got := string(out)
	want := strings.Join([]string{
		"Keep me.",
		"New paragraph that sounds like a human.",
		"Also keep 50% mid-line alone.",
		"NoSpaceStillWorks",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("new mode\ngot:\n%q\nwant:\n%q", got, want)
	}
	if st.Comments != 1 || st.PlusKept != 2 || st.MinusDropped != 1 || st.PlusDropped != 0 || st.MinusKept != 0 {
		t.Fatalf("stats: %+v", st)
	}
}

func TestApplyLineMarkersOld(t *testing.T) {
	in := strings.Join([]string{
		"Keep me.",
		"%% note",
		"%- Old paragraph.",
		"%+ New paragraph.",
		"",
	}, "\n")

	out, st := applyLineMarkers([]byte(in), true)
	got := string(out)
	want := strings.Join([]string{
		"Keep me.",
		"Old paragraph.",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("old mode\ngot:\n%q\nwant:\n%q", got, want)
	}
	if st.Comments != 1 || st.MinusKept != 1 || st.PlusDropped != 1 || st.PlusKept != 0 || st.MinusDropped != 0 {
		t.Fatalf("stats: %+v", st)
	}
}

func TestApplyLineMarkersColumnZeroOnly(t *testing.T) {
	in := "Not a marker: %+ still text\n  %+ indented not a marker\n"
	out, st := applyLineMarkers([]byte(in), false)
	if string(out) != in {
		t.Fatalf("mid-line/indent must pass through unchanged:\n%q", out)
	}
	if st != (lineStats{}) {
		t.Fatalf("expected empty stats, got %+v", st)
	}
}
