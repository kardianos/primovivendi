package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewPartAPatterns(t *testing.T) {
	hits := proseHitsForText("ch.md", strings.Join([]string{
		"#### 1. Form vs. Substance",
		"Between Time A and Time B, Clan A raided Clan B.",
		"#### Layer A: Scope",
		"In Hangar A and Hangar B, mechanics checked bolts.",
	}, "\n"))
	joined := joinTexts(hits)
	if !strings.Contains(joined, "Form vs. Substance") {
		t.Errorf("expected vs. in heading to be flagged, got: %s", joined)
	}
	if !strings.Contains(joined, "Time A") {
		t.Errorf("expected Time A / Clan A placeholder to be flagged, got: %s", joined)
	}
	if strings.Contains(joined, "Layer A") {
		t.Errorf("Layer A heading should not be flagged as algebraic placeholder: %s", joined)
	}
	if strings.Contains(joined, "Hangar A") {
		t.Errorf("Hangar A should not be flagged as algebraic placeholder: %s", joined)
	}
}

func TestListBlockAudit(t *testing.T) {
	doc := strings.Join([]string{
		"## Short List",
		"* Apples and pears.",
		"* Bread and butter.",
		"* Salt and water.",
		"",
		"## Prose In Bullets",
		"1. **First Essay:** " + strings.Repeat("word ", 65),
		"2. **Second Essay:** " + strings.Repeat("word ", 65),
		"",
		"## Nested Multi-Para",
		"1. **Heading:**",
		"",
		"    Indented paragraph continuing the item.",
	}, "\n")
	blocks := scanListBlocks("sample.md", doc, listOptions{maxAvg: 55, pairMaxAvg: 35})
	if len(blocks) != 3 {
		t.Fatalf("expected 3 list blocks, got %d", len(blocks))
	}
	if blocks[0].Reason != "" {
		t.Errorf("short list should be KEEP, got reason %q", blocks[0].Reason)
	}
	if blocks[1].Reason == "" {
		t.Errorf("long 2-item block should be flagged as prose-like")
	}
	if !blocks[2].HasMultiPara || blocks[2].Reason == "" {
		t.Errorf("indented multi-para list should be flagged, got %+v", blocks[2])
	}
}

func TestPacketSimulation(t *testing.T) {
	td := t.TempDir()
	chDir := filepath.Join(td, "civ", "chapters")
	if err := os.MkdirAll(chDir, 0o755); err != nil {
		t.Fatal(err)
	}
	chPath := filepath.Join(chDir, "01_ch01.md")
	orig := "## Intro\nFirst, institutional frameworks fundamentally require operational alignment, structural coherence, and procedural verification across your administrative organization.\n"
	if err := os.WriteFile(chPath, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	// Clean replacement should pass with 0 regressions.
	pkt := simPacket{
		Version: 1,
		Items: []simItem{{
			ID:    "call:civ/chapters/01_ch01.md:12345678:voice/plain-style",
			Path:  "civ/chapters/01_ch01.md",
			Quote: "First, institutional frameworks fundamentally require operational alignment, structural coherence, and procedural verification across your administrative organization.",
			Alternates: []simAlt{{
				ID:   "a",
				Text: "A working office keeps clear rules and checks its own ledger.",
			}},
		}},
	}
	res, err := simulatePacket(td, pkt, "a", 5, 1.2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.NewProseHits) != 0 {
		t.Errorf("expected 0 regressions, got %v", res.NewProseHits)
	}
	if res.AfterMean >= res.BeforeMean {
		t.Errorf("expected voice mean to improve: %+.2f -> %+.2f", res.BeforeMean, res.AfterMean)
	}

	// Regression introducing an em-dash must be caught.
	pktBad := simPacket{
		Version: 1,
		Items: []simItem{{
			ID:         "call:civ/chapters/01_ch01.md:12345678:voice/bad",
			Path:       "civ/chapters/01_ch01.md",
			Quote:      pkt.Items[0].Quote,
			Alternates: []simAlt{{ID: "a", Text: "A working office — checks its own ledger."}},
		}},
	}
	resBad, err := simulatePacket(td, pktBad, "a", 5, 1.2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resBad.NewProseHits) == 0 {
		t.Errorf("expected em-dash regression to be detected")
	}

	// Overlapping items must fail.
	pktOverlap := simPacket{
		Version: 1,
		Items: []simItem{
			{ID: "i1", Path: "civ/chapters/01_ch01.md", Quote: "First, institutional frameworks", Alternates: []simAlt{{ID: "a", Text: "Rules"}}},
			{ID: "i2", Path: "civ/chapters/01_ch01.md", Quote: "institutional frameworks fundamentally", Alternates: []simAlt{{ID: "a", Text: "rules"}}},
		},
	}
	if _, err := simulatePacket(td, pktOverlap, "a", 5, 1.2); err == nil {
		t.Errorf("expected overlap error, got nil")
	}
}
