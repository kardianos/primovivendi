package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaterPacketKeepsRevise(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "p.md"), []byte("Hello there.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "a.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "same", "path": "p.md", "quote": "Hello there.",
			"problem": "old", "byte_offset": 0,
			"alternates": []map[string]string{{"id": "a", "text": "Hi."}},
		}},
	})
	writeJSON(t, filepath.Join(packets, "b.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{{
			"id": "same", "path": "p.md", "quote": "Hello there.",
			"problem": "new", "byte_offset": 0,
			"alternates": []map[string]string{{"id": "a", "text": "Hello."}},
		}},
	})
	decPath := filepath.Join(root, "decisions.json")
	writeJSON(t, decPath, map[string]any{
		"version": 1,
		"items": map[string]any{
			"same": map[string]any{"disposition": "revise", "text": "Author.", "alternate": "a"},
		},
	})
	sess, err := LoadSession(root, packets, decPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.Items) != 1 {
		t.Fatalf("items: %d", len(sess.Items))
	}
	if sess.Items[0].Problem != "new" {
		t.Fatalf("problem %q", sess.Items[0].Problem)
	}
	if sess.Items[0].Alternates[0].Text != "Hello." {
		t.Fatalf("alternate %q", sess.Items[0].Alternates[0].Text)
	}
	d := sess.Decisions.Items["same"]
	if d.Disposition != "revise" || d.Text != "Author." {
		t.Fatalf("decision %+v", d)
	}
	if sess.Hits["same"].Status != StatusOffset {
		t.Fatalf("hit %+v", sess.Hits["same"])
	}
}

func TestAppliedOnlyWhenReplacementIsInTheFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "chap.md"), []byte("alpha NEW omega OLD"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(packets, "items.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{
			{"id": "yes", "path": "chap.md", "quote": "alpha", "alternates": []map[string]string{{"id": "a", "text": "NEW"}}},
			{"id": "no", "path": "chap.md", "quote": "OLD", "alternates": []map[string]string{{"id": "a", "text": "GONE"}}},
			{"id": "skip", "path": "chap.md", "quote": "alpha", "alternates": []map[string]string{{"id": "a", "text": "NEW"}}},
		},
	})
	writeJSON(t, filepath.Join(packets, "response.json"), map[string]any{
		"version": 1,
		"applied": []map[string]any{
			{"id": "yes", "path": "chap.md", "ok": true},
			{"id": "no", "path": "chap.md", "ok": true},
			{"id": "skip", "path": "chap.md", "ok": false},
		},
	})
	decPath := filepath.Join(root, "decisions.json")
	writeJSON(t, decPath, map[string]any{
		"version": 1,
		"items": map[string]any{
			"yes":  map[string]string{"disposition": "accept", "alternate": "a"},
			"no":   map[string]string{"disposition": "accept", "alternate": "a"},
			"skip": map[string]string{"disposition": "accept", "alternate": "a"},
		},
	})
	sess, err := LoadSession(root, packets, decPath)
	if err != nil {
		t.Fatal(err)
	}
	if !sess.Applied["yes"] {
		t.Fatal("yes should be applied: the file contains NEW")
	}
	if sess.Applied["no"] {
		t.Fatal("no should stay unapplied: the file still has the old text and not GONE")
	}
	if sess.Applied["skip"] {
		t.Fatal("ok false is not applied")
	}
}

func TestBatchSections(t *testing.T) {
	s := &Session{
		Items: []Item{
			{ID: "acc", Path: "a.md", Quote: "old", Before: "\n\n", After: " ", Problem: "p1", Rule: "r1", Alternates: []Alternate{{ID: "a", Text: "new"}}},
			{ID: "done", Path: "a.md", Quote: "qdone", Alternates: []Alternate{{ID: "a", Text: "replaced"}}},
			{ID: "rev", Path: "a.md", Quote: "q2"},
			{ID: "hold", Path: "a.md", Quote: "q3", Problem: "why this claim?", Rule: "call/claim", Alternates: []Alternate{{ID: "a", Label: "shorter", Text: "alt3"}}},
			{ID: "rej", Path: "a.md", Quote: "q4", Rule: "r4"},
			{ID: "open", Path: "a.md", Quote: "q5"},
		},
		Hits: map[string]Hit{
			"acc":  {Status: StatusOffset, Offset: 1, End: 4, Line: 2},
			"done": {Status: StatusMissing},
			"rev":  {Status: StatusMoved, Offset: 4, End: 6, Line: 3},
			"hold": {Status: StatusOffset, Offset: 8, End: 10, Line: 4},
			"rej":  {Status: StatusMissing},
		},
		Applied: map[string]bool{"done": true},
		Answers: map[string][]Answer{"hold": {{ID: "hold", Text: "prior note"}}},
		Decisions: DecisionFile{Items: map[string]Decision{
			"acc":  {Disposition: "accept", Alternate: "a", AcceptedText: "new"},
			"done": {Disposition: "accept", Alternate: "a", AcceptedText: "replaced"},
			"rev":  {Disposition: "revise", Text: "mine"},
			"hold": {Disposition: "hold", HoldFor: "research", Note: "look up"},
			"rej":  {Disposition: "reject", Note: "no"},
		}},
	}
	md := s.BatchMarkdown()
	apply := section(md, "Apply")
	answer := section(md, "Answer")
	closed := section(md, "Closed")
	for _, want := range []string{"### acc", "### rev", "new", "mine", "match: offset byte 1..4 line 2", `before: "\n\n"`, "auto_apply: yes"} {
		if !strings.Contains(apply, want) {
			t.Errorf("Apply missing %q\n%s", want, apply)
		}
	}
	if strings.Contains(apply, "### done") {
		t.Fatalf("already-applied item appeared in Apply\n%s", apply)
	}
	if strings.Contains(apply, "### rej") || strings.Contains(apply, "### hold") {
		t.Fatalf("Apply has a closed or held item\n%s", apply)
	}
	for _, want := range []string{"### hold", "research", "look up", "why this claim?", "call/claim", "alt3", "prior note"} {
		if !strings.Contains(answer, want) {
			t.Fatalf("Answer missing %q:\n%s", want, answer)
		}
	}
	if !strings.Contains(closed, "### rej") || !strings.Contains(closed, "no") || !strings.Contains(closed, "r4") {
		t.Fatalf("Closed:\n%s", closed)
	}
	if strings.Contains(md, "### open") {
		t.Fatal("an open item is not in the batch")
	}
	if !strings.Contains(md, "Do not run civ/build.fish") {
		t.Fatal("guardrail missing")
	}
	if s.IsStale("done") || s.IsStale("rej") {
		t.Fatal("applied or rejected item must not be counted as stale")
	}
	rawJSON, err := s.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}
	var payload ExportPayload
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Apply) != 2 || len(payload.Answer) != 1 || len(payload.Closed) != 1 || len(payload.Applied) != 1 {
		t.Fatalf("unexpected export json counts: %+v", payload)
	}
}

func TestSamplePacket(t *testing.T) {
	root, err := findRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	sess, err := LoadSession(root, filepath.Join(root, "civ", "review", "testdata", "packets"), filepath.Join(t.TempDir(), "decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.ParseErrors) > 0 {
		t.Fatal(sess.ParseErrors)
	}
	want := map[string]struct {
		status HitStatus
		offset int
	}{
		"fixture:offset-household":    {StatusOffset, 0},
		"fixture:moved-household":     {StatusMoved, 200},
		"fixture:ambiguous-household": {StatusAmbiguous, 0},
		"fixture:missing":             {StatusMissing, 0},
		"fixture:fire-rewrite":        {StatusOffset, 99},
		"fixture:emdash":              {StatusOffset, 318},
	}
	if len(sess.Items) != len(want) {
		t.Fatalf("items %d", len(sess.Items))
	}
	for id, w := range want {
		h := sess.Hits[id]
		if h.Status != w.status {
			t.Errorf("%s status %s, want %s (%+v)", id, h.Status, w.status, h)
		}
		if w.status == StatusOffset || w.status == StatusMoved {
			if h.Offset != w.offset {
				t.Errorf("%s offset %d, want %d", id, h.Offset, w.offset)
			}
		}
	}
	amb := sess.Hits["fixture:ambiguous-household"]
	if len(amb.Candidates) != 2 || amb.Candidates[0] != 0 || amb.Candidates[1] != 200 {
		t.Fatalf("ambiguous candidates %+v", amb)
	}
	off := sess.Hits["fixture:offset-household"]
	if !CanAccept(off, false) || CanAccept(amb, false) || CanAccept(sess.Hits["fixture:missing"], false) {
		t.Fatal("accept rules")
	}
}

func section(md, title string) string {
	start := strings.Index(md, "## "+title)
	if start < 0 {
		return ""
	}
	rest := md[start+len("## "+title):]
	if i := strings.Index(rest, "\n## "); i >= 0 {
		return rest[:i]
	}
	return rest
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
