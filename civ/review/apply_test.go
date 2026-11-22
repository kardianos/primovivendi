package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyBottomUpAndGuards(t *testing.T) {
	root := t.TempDir()
	orig := "First line keeps truth.\n\nSecond line keeps truth.\n\nThird line is locked.\n\nFourth line has note.\n\nFifth overlap target here.\n"
	chap := filepath.Join(root, "chap.md")
	if err := os.WriteFile(chap, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")
	if err := os.MkdirAll(packets, 0o755); err != nil {
		t.Fatal(err)
	}
	decPath := filepath.Join(root, "decisions.json")

	// 1. Offset hit (accept)
	// 2. Duplicate quote disambiguated by before (moved, revise)
	// 3. Locked item (must be skipped without AllowLocked)
	// 4. Item with editorial note (must be skipped without IncludeNoted)
	// 5+6. Two overlapping items (both must be skipped)
	writeJSON(t, filepath.Join(packets, "20261120-100000-init.json"), map[string]any{
		"version": 1,
		"items": []map[string]any{
			{
				"id": "i1", "path": "chap.md", "byte_offset": 0, "byte_end": len("First line keeps truth."),
				"quote":      "First line keeps truth.",
				"alternates": []map[string]string{{"id": "a", "text": "First line holds truth."}},
			},
			{
				"id": "i2", "path": "chap.md", "quote": "keeps truth.", "before": "Second line ",
				"alternates": []map[string]string{{"id": "a", "text": "guards truth."}},
			},
			{
				"id": "i3-locked", "path": "chap.md", "quote": "Third line is locked.", "locked": true,
				"alternates": []map[string]string{{"id": "a", "text": "Changed."}},
			},
			{
				"id": "i4-noted", "path": "chap.md", "quote": "Fourth line has note.",
				"alternates": []map[string]string{{"id": "a", "text": "Fourth line edited."}},
			},
			{
				"id": "i5-ov1", "path": "chap.md", "quote": "Fifth overlap target",
				"alternates": []map[string]string{{"id": "a", "text": "Fifth A"}},
			},
			{
				"id": "i6-ov2", "path": "chap.md", "quote": "overlap target here.",
				"alternates": []map[string]string{{"id": "a", "text": "B here."}},
			},
		},
	})
	writeJSON(t, decPath, map[string]any{
		"version": 1,
		"items": map[string]any{
			"i1":        map[string]any{"disposition": "accept", "alternate": "a", "accepted_text": "First line holds truth.", "at": "2026-11-20T10:01:00Z"},
			"i2":        map[string]any{"disposition": "revise", "text": "guards truth.", "at": "2026-11-20T10:02:00Z"},
			"i3-locked": map[string]any{"disposition": "accept", "alternate": "a", "accepted_text": "Changed.", "at": "2026-11-20T10:03:00Z"},
			"i4-noted":  map[string]any{"disposition": "accept", "alternate": "a", "accepted_text": "Fourth line edited.", "note": "check next sentence", "at": "2026-11-20T10:04:00Z"},
			"i5-ov1":    map[string]any{"disposition": "accept", "alternate": "a", "accepted_text": "Fifth A", "at": "2026-11-20T10:05:00Z"},
			"i6-ov2":    map[string]any{"disposition": "accept", "alternate": "a", "accepted_text": "B here.", "at": "2026-11-20T10:06:00Z"},
		},
	})

	sess, err := LoadSession(root, packets, decPath)
	if err != nil {
		t.Fatal(err)
	}
	// Dry run leaves file untouched.
	dry, err := sess.RunApply(ApplyOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(dry.Applied) != 2 || len(dry.Skipped) != 4 {
		t.Fatalf("dry run: applied=%d skipped=%d (%+v)", len(dry.Applied), len(dry.Skipped), dry)
	}
	afterDry, _ := os.ReadFile(chap)
	if string(afterDry) != orig {
		t.Fatal("dry run modified file")
	}

	// Real apply splices i1 and i2 bottom-up and writes a response packet.
	res, err := sess.RunApply(ApplyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Applied) != 2 || len(res.Skipped) != 4 || res.ResponsePacket == "" {
		t.Fatalf("apply result: %+v", res)
	}
	gotBytes, _ := os.ReadFile(chap)
	got := string(gotBytes)
	want := "First line holds truth.\n\nSecond line guards truth.\n\nThird line is locked.\n\nFourth line has note.\n\nFifth overlap target here.\n"
	if got != want {
		t.Fatalf("file after apply:\n%s\nwant:\n%s", got, want)
	}

	// Reloading session marks i1 and i2 as applied, not stale, and omits them from Apply export.
	sess2, err := LoadSession(root, packets, decPath)
	if err != nil {
		t.Fatal(err)
	}
	if !sess2.Applied["i1"] || !sess2.Applied["i2"] {
		t.Fatalf("reloaded applied map: %+v", sess2.Applied)
	}
	if sess2.IsStale("i1") || sess2.IsStale("i2") {
		t.Fatal("applied items should not be stale")
	}
	exp := sess2.BuildExport()
	if len(exp.Applied) != 2 || len(exp.Apply) != 4 {
		t.Fatalf("export after apply: applied=%v remaining=%d", exp.Applied, len(exp.Apply))
	}

	// IncludeNoted applies i4-noted cleanly.
	res2, err := sess2.RunApply(ApplyOptions{IncludeNoted: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res2.Applied) != 1 || res2.Applied[0].ID != "i4-noted" {
		t.Fatalf("include-noted result: %+v", res2)
	}
	got2, _ := os.ReadFile(chap)
	if !strings.Contains(string(got2), "Fourth line edited.") {
		t.Fatalf("i4-noted not written:\n%s", string(got2))
	}
}
