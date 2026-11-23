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
	if !strings.Contains(tt.Dump(), "No item selected") {
		t.Fatalf("the only open item should clear after it is decided\n%s", tt.Dump())
	}
	tt.Tap("Decided")
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
	s := &reviewView{
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

func TestEmptyQueueClearsSelection(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte("Alpha sentence stays.\nBeta sentence stays.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261122-120000-queue.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{
			{
				"id": "q-alpha", "path": "doc.md", "quote": "Alpha sentence stays.",
				"problem":    "Alpha is still open.",
				"alternates": []map[string]string{{"id": "a", "label": "alpha-alt", "text": "Alpha rewritten."}},
			},
			{
				"id": "q-beta", "path": "doc.md", "quote": "Beta sentence stays.",
				"problem":    "Beta is still open.",
				"alternates": []map[string]string{{"id": "a", "label": "beta-alt", "text": "Beta rewritten."}},
			},
		},
	})
	dec := filepath.Join(root, "decisions.json")
	sess, err := LoadSession(root, packets, dec)
	if err != nil {
		t.Fatal(err)
	}
	tt := gowakutest.New(t, reviewApp{Session: sess}, gowakutest.Options{Width: 1440, Height: 900})
	if !strings.Contains(tt.Dump(), "Alpha is still open.") {
		t.Fatalf("expected the first open item\n%s", tt.Dump())
	}
	tt.Tap("Reject")
	dump := tt.Dump()
	if !strings.Contains(dump, "Beta is still open.") || strings.Contains(dump, "Alpha is still open.") {
		t.Fatalf("expected the remaining open item\n%s", dump)
	}
	tt.Tap("Reject")
	dump = tt.Dump()
	for _, want := range []string{"No items", "No item selected"} {
		if !strings.Contains(dump, want) {
			t.Fatalf("missing %q after the queue emptied\n%s", want, dump)
		}
	}
	for _, gone := range []string{"Alpha is still open.", "Beta is still open.", "Use my wording", "alpha-alt", "beta-alt"} {
		if strings.Contains(dump, gone) {
			t.Fatalf("finished queue still shows %q\n%s", gone, dump)
		}
	}

	// All still lists a decided item, so the last one stays selected.
	tt.Tap("All")
	tt.Tap("Next")
	tt.Tap("Reject")
	dump = tt.Dump()
	if strings.Contains(dump, "No item selected") || !strings.Contains(dump, "Beta is still open.") {
		t.Fatalf("the last item in All should stay selected\n%s", dump)
	}
}

func TestOpenQueueAlreadyEmptySelectsNothing(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte("Gamma sentence stays.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "20261122-120100-done.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "q-gamma", "path": "doc.md", "quote": "Gamma sentence stays.",
			"problem": "Gamma was already decided.",
		}},
	})
	dec := filepath.Join(root, "decisions.json")
	writeJSON(t, dec, map[string]any{
		"version": 1,
		"items": map[string]any{
			"q-gamma": map[string]string{"disposition": "reject"},
		},
	})
	sess, err := LoadSession(root, packets, dec)
	if err != nil {
		t.Fatal(err)
	}
	tt := gowakutest.New(t, reviewApp{Session: sess}, gowakutest.Options{Width: 1440, Height: 900})
	dump := tt.Dump()
	if !strings.Contains(dump, "No items") || !strings.Contains(dump, "No item selected") {
		t.Fatalf("empty open queue should select nothing\n%s", dump)
	}
	if strings.Contains(dump, "Gamma was already decided.") || strings.Contains(dump, "Use my wording") {
		t.Fatalf("empty open queue still shows the decided item\n%s", dump)
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
