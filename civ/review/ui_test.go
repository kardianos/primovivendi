package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bankai-tensei/gowaku/gowakutest"
)

func TestTapAcceptWritesDecision(t *testing.T) {
	root, err := findRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	dec := filepath.Join(t.TempDir(), "decisions.json")
	sess, err := LoadSession(root, filepath.Join(root, "civ", "review", "testdata", "packets"), dec)
	if err != nil {
		t.Fatal(err)
	}
	tt := gowakutest.New(t, reviewApp{Session: sess}, gowakutest.Options{Width: 1440, Height: 900})
	if !strings.Contains(tt.Dump(), "Fire burns cotton") {
		t.Fatalf("two lines below line 1 should reach line 3\n%s", tt.Dump())
	}
	if strings.Contains(tt.Dump(), "keeping what was built") {
		t.Fatal("two lines below line 1 should stop short of line 7")
	}
	tt.Tap("Below")
	if !strings.Contains(tt.Dump(), "keeping what was built") {
		t.Fatalf("Below did not open to the end of the fixture\n%s", tt.Dump())
	}
	tt.Tap("Tight")
	if strings.Contains(tt.Dump(), "keeping what was built") {
		t.Fatal("Tight did not return to two lines")
	}
	tt.Tap("Chapter")
	if !strings.Contains(tt.Dump(), "keeping what was built") {
		t.Fatal("Chapter did not show the end of the file")
	}
	tt.Tap("Tight")
	tt.Tap("Compare")
	tt.Tap("Accept")
	b, err := os.ReadFile(dec)
	if err != nil {
		t.Fatal(err)
	}
	var got DecisionFile
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	d := got.Items["fixture:offset-household"]
	if d.Disposition != "accept" || d.Alternate != "a" {
		t.Fatalf("decision %+v\n%s", d, tt.Dump())
	}
	if d.AcceptedText != "The household rears its own children." {
		t.Fatalf("accepted_text not snapshotted: %+v", d)
	}
	if d.At == "" {
		t.Fatal("decision has no timestamp")
	}
	dump := tt.Dump()
	if strings.Contains(dump, "The sentence is sound") {
		t.Fatalf("stayed on the decided item\n%s", dump)
	}
	if !strings.Contains(dump, "The recorded offset is wrong") {
		t.Fatalf("did not show the next item\n%s", dump)
	}
}

func TestUIShowsFromAndIgnore(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte("The household keeps its own children.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261121-120000-from.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "from1", "from": "grok-4.7", "path": "doc.md",
			"quote":   "The household keeps its own children.",
			"problem": "Two agents offered drafts.",
			"alternates": []map[string]any{
				{"id": "a", "label": "shorter", "text": "The household keeps its children."},
				{"id": "b", "label": "other", "from": "opus-5.5", "text": "The household rears its children."},
				{"id": "no", "from": "cabon", "ignore": true, "reason": "this was actually good"},
			},
		}},
	})
	dec := filepath.Join(root, "decisions.json")
	sess, err := LoadSession(root, packets, dec)
	if err != nil {
		t.Fatal(err)
	}
	tt := gowakutest.New(t, reviewApp{Session: sess}, gowakutest.Options{Width: 1440, Height: 900})
	dump := tt.Dump()
	for _, want := range []string{"from: grok-4.7", "from grok-4.7", "Word diff (Quote -> shorter · grok-4.7)", "opus-5.5", "ignore"} {
		if !strings.Contains(dump, want) {
			t.Fatalf("missing %q\n%s", want, dump)
		}
	}
	if strings.Contains(dump, "from: opus-5.5") || strings.Contains(dump, "this was actually good") {
		t.Fatalf("focused alternate should show the item from, not the other votes\n%s", dump)
	}
	tt.Tap("other")
	dump = tt.Dump()
	if !strings.Contains(dump, "from: opus-5.5") || !strings.Contains(dump, "Word diff (Quote -> other · opus-5.5)") {
		t.Fatalf("other alternate from not shown\n%s", dump)
	}
	tt.Tap("Ignore")
	dump = tt.Dump()
	for _, want := range []string{"Ignore", "from: cabon", "this was actually good"} {
		if !strings.Contains(dump, want) {
			t.Fatalf("missing %q\n%s", want, dump)
		}
	}
	if strings.Contains(dump, "Word diff (Quote -> Ignore") {
		t.Fatalf("an ignore vote must not be shown as a replacement diff\n%s", dump)
	}
	tt.Tap("Ignore this")
	b, err := os.ReadFile(dec)
	if err != nil {
		t.Fatal(err)
	}
	var got DecisionFile
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	d := got.Items["from1"]
	if d.Disposition != "reject" || d.Alternate != "no" || d.Note != "this was actually good" || d.AcceptedText != "" {
		t.Fatalf("ignore vote was not recorded as reject: %+v", d)
	}
	tt.Tap("Compare")
	dump = tt.Dump()
	if !strings.Contains(dump, "from: grok-4.7") || !strings.Contains(dump, "from: opus-5.5") || !strings.Contains(dump, "from: cabon") {
		t.Fatalf("compare columns missing from\n%s", dump)
	}
}

func TestFollowingIDSkipsASavedItemAndWrapsLast(t *testing.T) {
	s := &reviewState{
		filter:   filterOpen,
		selected: "a",
		session: &Session{
			Items:     []Item{{ID: "a"}, {ID: "b"}, {ID: "c"}},
			Hits:      map[string]Hit{},
			Decisions: DecisionFile{Items: map[string]Decision{"b": {Disposition: "accept"}}},
		},
	}
	if got := s.followingID(); got != "c" {
		t.Fatalf("followingID = %q, want c", got)
	}
	// On last open item while earlier open item "a" remains, wrap back to "a".
	s.selected = "c"
	if got := s.followingID(); got != "a" {
		t.Fatalf("last followingID = %q, want a", got)
	}
	// When only one open item remains, return empty.
	s.session.Decisions.Items["a"] = Decision{Disposition: "reject"}
	if got := s.followingID(); got != "" {
		t.Fatalf("sole followingID = %q, want empty", got)
	}
}

func TestUICopyToMineLockedBannerAndReopen(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte("Justice is scoped procedural symmetry.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261120-120000-test.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "lock1", "path": "doc.md", "quote": "Justice is scoped procedural symmetry.",
			"problem": "Check locked banner.", "locked": true,
			"alternates": []map[string]string{{"id": "a", "label": "alt", "text": "Alt wording."}},
		}},
	})
	dec := filepath.Join(root, "decisions.json")
	sess, err := LoadSession(root, packets, dec)
	if err != nil {
		t.Fatal(err)
	}
	tt := gowakutest.New(t, reviewApp{Session: sess}, gowakutest.Options{Width: 1440, Height: 900})
	if !strings.Contains(tt.Dump(), "LOCKED LINE") {
		t.Fatalf("locked banner not shown\n%s", tt.Dump())
	}
	tt.Tap("Quote -> Mine")
	if !strings.Contains(tt.Dump(), "Word diff (Quote -> Mine)") {
		t.Fatalf("expected live diff against Mine\n%s", tt.Dump())
	}
	tt.Tap("Reject")
	tt.Tap("Decided")
	tt.Tap("Reopen")
	b, err := os.ReadFile(dec)
	if err != nil {
		t.Fatal(err)
	}
	var got DecisionFile
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if _, exists := got.Items["lock1"]; exists {
		t.Fatalf("Reopen did not clear decision: %+v", got.Items)
	}
}
