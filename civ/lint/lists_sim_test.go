package main

import (
	"strings"
	"testing"
)

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
