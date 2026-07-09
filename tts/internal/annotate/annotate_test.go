package annotate

import (
	"strings"
	"testing"

	"po/tts/internal/md"
)

func TestStructurePausesOnly(t *testing.T) {
	blocks := []md.Block{
		{Kind: md.KindHeading, Level: 1, Text: "Part I"},
		{Kind: md.KindHeading, Level: 2, Text: "Chapter 1: Reality is Real"},
		{Kind: md.KindParagraph, Text: "First paragraph about reality."},
		{Kind: md.KindParagraph, Text: "Second paragraph continues."},
		{Kind: md.KindHeading, Level: 3, Text: "Reason Is Not Enough"},
		{Kind: md.KindParagraph, Text: "Section body."},
		{Kind: md.KindListItem, Text: "One"},
		{Kind: md.KindListItem, Text: "Two"},
	}
	got := Text(blocks, DefaultOptions())

	// No wrapping delivery tags
	for _, bad := range []string{"<emphasis>", "<slow>", "<soft>", "<whisper>", "</"} {
		if strings.Contains(got, bad) {
			t.Fatalf("unexpected tag %q in:\n%s", bad, got)
		}
	}
	if !strings.Contains(got, "[long-pause] Part I. [long-pause]") {
		t.Fatalf("h1 pause:\n%s", got)
	}
	if !strings.Contains(got, "[long-pause] Chapter 1: Reality is Real. [long-pause]") {
		t.Fatalf("h2 pause:\n%s", got)
	}
	if !strings.Contains(got, "[long-pause] Reason Is Not Enough. [pause]") {
		t.Fatalf("h3 pause:\n%s", got)
	}
	if !strings.Contains(got, "First paragraph about reality. [pause] Second paragraph continues.") {
		t.Fatalf("para pause:\n%s", got)
	}
	if !strings.Contains(got, "One [pause] Two") {
		t.Fatalf("list pause:\n%s", got)
	}
}

func TestEnsureTerminalPunct(t *testing.T) {
	blocks := []md.Block{{Kind: md.KindHeading, Level: 2, Text: "Already ends."}}
	got := Text(blocks, DefaultOptions())
	if strings.Contains(got, "Already ends..") {
		t.Fatalf("double period: %s", got)
	}
}
