package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bankai-tensei/gowaku/core"
	"github.com/bankai-tensei/gowaku/geom"
	"github.com/bankai-tensei/gowaku/gowakutest"
	"github.com/bankai-tensei/gowaku/input"
	"github.com/bankai-tensei/gowaku/semantics"
)

// openReview opens the window on a clock that moves clickQuiet on every
// button release, so a test can click through a queue. A frozen clock is
// what shows that a second release is ignored.
func openReview(t *testing.T, sess *Session) *gowakutest.Tester {
	t.Helper()
	now := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	return gowakutest.New(t, reviewApp{
		Session: sess,
		now: func() time.Time {
			now = now.Add(clickQuiet)
			return now
		},
	}, gowakutest.Options{Width: 1440, Height: 900})
}

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
	tt := openReview(t, sess)
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
	tt := openReview(t, sess)
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
	tt := openReview(t, sess)
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
	tt := openReview(t, sess)
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
	tt := openReview(t, sess)
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

func TestButtonQuietPeriod(t *testing.T) {
	now := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	s := &reviewState{now: func() time.Time { return now }}
	n := 0
	press := s.guard(func() { n++ })
	if s.guard(nil) != nil {
		t.Fatal("a disabled button must stay disabled")
	}
	press()
	press()
	press()
	if n != 1 {
		t.Fatalf("flutter ran %d times", n)
	}
	now = now.Add(clickQuiet - time.Millisecond)
	press()
	if n != 1 {
		t.Fatalf("press inside the quiet period ran, count %d", n)
	}
	now = now.Add(time.Millisecond)
	press()
	if n != 2 {
		t.Fatalf("press after the quiet period count %d", n)
	}
}

func TestFlutterClickDoesNotDecideTheNextItem(t *testing.T) {
	root, err := findRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	dec := filepath.Join(t.TempDir(), "decisions.json")
	sess, err := LoadSession(root, filepath.Join(root, "civ", "review", "testdata", "packets"), dec)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	tt := gowakutest.New(t, reviewApp{Session: sess, now: func() time.Time { return now }}, gowakutest.Options{Width: 1440, Height: 900})
	accept, ok := tt.Find("Accept")
	if !ok {
		t.Fatalf("no Accept button\n%s", tt.Dump())
	}
	// Three releases, close together on the pointer clock, so the toolkit
	// counts a triple click and would run the button on the first and the third.
	clickN(tt, accept.Rect.Center(), 3)
	if !strings.Contains(tt.Dump(), "The recorded offset is wrong") {
		t.Fatalf("flutter did not leave the next item showing\n%s", tt.Dump())
	}
	got := readDecisions(t, dec)
	if len(got.Items) != 1 || got.Items["fixture:offset-household"].Disposition != "accept" {
		t.Fatalf("flutter wrote %+v", got.Items)
	}
	tt.Tap("Reject")
	got = readDecisions(t, dec)
	if _, decided := got.Items["fixture:moved-household"]; decided {
		t.Fatal("a second button inside the quiet period decided the next item")
	}
	now = now.Add(clickQuiet)
	tt.Tap("Reject")
	got = readDecisions(t, dec)
	if got.Items["fixture:moved-household"].Disposition != "reject" {
		t.Fatalf("a click after the quiet period did not reject: %+v", got.Items["fixture:moved-household"])
	}
}

func TestUISelectAndCopyPassage(t *testing.T) {
	root, err := findRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	dec := filepath.Join(t.TempDir(), "decisions.json")
	sess, err := LoadSession(root, filepath.Join(root, "civ", "review", "testdata", "packets"), dec)
	if err != nil {
		t.Fatal(err)
	}
	tt := openReview(t, sess)
	const mark = "Fire burns cotton"
	n, ok := tt.FindFunc(func(n *gowakutest.Node) bool {
		return n.Role == semantics.RoleText && strings.Contains(n.Label, "\n") && strings.Contains(n.Label, mark)
	})
	if !ok {
		t.Fatalf("passage text is not on screen\n%s", tt.Dump())
	}
	// The first line of the passage, inside the glyphs rather than the list preview.
	lineY := n.Rect.Y + 8
	tt.DoubleTapAt(geom.Offset{X: n.Rect.X + 12, Y: lineY})
	tt.Key(input.KeyC, input.ModControl)
	word := tt.Clipboard()
	if word == "" || !strings.Contains(n.Label, word) {
		t.Fatalf("double-click copied %q from passage %q", word, n.Label)
	}
	from := geom.Offset{X: n.Rect.X + 4, Y: lineY}
	to := geom.Offset{X: from.X + 80, Y: from.Y}
	if maxX := n.Rect.X + n.Rect.W - 2; to.X > maxX {
		to.X = maxX
	}
	tt.DragAt(from, to)
	tt.Key(input.KeyC, input.ModControl)
	got := tt.Clipboard()
	if got == "" || !strings.Contains(n.Label, got) {
		t.Fatalf("drag copied %q from %q", got, n.Label)
	}
	tt.SecondaryTapAt(from)
	if !strings.Contains(tt.Dump(), "Copy") || !strings.Contains(tt.Dump(), "Select all") {
		t.Fatalf("right-click did not open the copy menu\n%s", tt.Dump())
	}
}

func TestProblemTextStaysAboveTheControls(t *testing.T) {
	short := openReview(t, problemFixture(t, "A short note."))
	note, above := problemAndAbove(t, short, "A short note.")
	gap := above.Rect.Y - (note.Rect.Y + note.Rect.H)
	if gap < 0 || gap > 40 {
		t.Fatalf("short problem left a gap of %v before Above\n%s", gap, short.Dump())
	}

	longText := strings.Repeat("The instruction keeps going. ", 80) + "PROBLEMEND."
	tt := openReview(t, problemFixture(t, longText))
	note, above = problemAndAbove(t, tt, "PROBLEMEND.")
	if note.Rect.H <= problemMaxHeight {
		t.Fatalf("long problem height %v, want more than the cap %v", note.Rect.H, problemMaxHeight)
	}
	if above.Rect.Y > note.Rect.Y+problemMaxHeight+40 || above.Rect.Y < note.Rect.Y+problemMaxHeight-8 {
		t.Fatalf("Above at %v, problem at %v, cap %v\n%s", above.Rect.Y, note.Rect.Y, problemMaxHeight, tt.Dump())
	}
	for _, w := range tt.Warnings() {
		if w.Kind == gowakutest.WarnOverflow && strings.Contains(w.Text, "PROBLEMEND") {
			t.Fatalf("problem overflowed its box: %+v", w)
		}
	}
	before := note.Rect.Y
	tt.ScrollAt(geom.Offset{X: note.Rect.X + 12, Y: note.Rect.Y + 12}, 0, -3, 0)
	note, _ = problemAndAbove(t, tt, "PROBLEMEND.")
	if note.Rect.Y >= before-1 {
		t.Fatalf("wheel did not move the problem: before %v after %v\n%s", before, note.Rect.Y, tt.Dump())
	}
}

func problemFixture(t *testing.T, problem string) *Session {
	t.Helper()
	root := t.TempDir()
	quote := "The household keeps its own children."
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte(quote+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "note.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "note1", "path": "doc.md", "quote": quote, "problem": problem,
		}},
	})
	sess, err := LoadSession(root, packets, filepath.Join(root, "decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

func problemAndAbove(t *testing.T, tt *gowakutest.Tester, problemMark string) (problem, above gowakutest.Node) {
	t.Helper()
	problem, ok := findNode(tt.Semantics(), problemMark, false)
	if !ok {
		t.Fatalf("no problem text %q\n%s", problemMark, tt.Dump())
	}
	above, ok = findNode(tt.Semantics(), "Above", true)
	if !ok {
		t.Fatalf("no Above button\n%s", tt.Dump())
	}
	return problem, above
}

func findNode(root *gowakutest.Node, label string, exact bool) (gowakutest.Node, bool) {
	var found gowakutest.Node
	var ok bool
	if root == nil {
		return found, false
	}
	root.Walk(func(n *gowakutest.Node) bool {
		match := n.Label == label
		if !exact {
			match = strings.Contains(n.Label, label)
		}
		if !match {
			return true
		}
		found = *n
		ok = true
		return false
	})
	return found, ok
}

func readDecisions(t *testing.T, path string) DecisionFile {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got DecisionFile
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

// clickN releases the primary button n times at pos, fast enough to count
// as one multi-click.
func clickN(tt *gowakutest.Tester, pos geom.Offset, n int) {
	var clock uint32 = 5000
	for range n {
		tt.Tree().DispatchPointer(core.PointerEvent{Kind: core.PointerMove, Pos: pos, Time: clock})
		tt.Tree().DispatchPointer(core.PointerEvent{Kind: core.PointerDown, Pos: pos, Button: input.ButtonPrimary, Time: clock})
		clock += 40
		tt.Tree().DispatchPointer(core.PointerEvent{Kind: core.PointerUp, Pos: pos, Button: input.ButtonPrimary, Time: clock})
		clock += 40
	}
	tt.Pump()
	tt.Settle()
}
