package md

import (
	"strings"
	"testing"
)

func TestParseHeadingsParagraphsLists(t *testing.T) {
	src := `# Part One

## Chapter 1: Reality

The first sentence.

Second paragraph with **bold** and *italic*.

### Section A

1. First item
2. Second item with a [link](https://example.com)
`
	blocks := Parse(src)
	if len(blocks) < 6 {
		t.Fatalf("expected >=6 blocks, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Kind != KindHeading || blocks[0].Level != 1 || blocks[0].Text != "Part One" {
		t.Fatalf("block0: %+v", blocks[0])
	}
	if blocks[1].Kind != KindHeading || blocks[1].Level != 2 {
		t.Fatalf("block1: %+v", blocks[1])
	}
	if blocks[2].Kind != KindParagraph || !strings.Contains(blocks[2].Text, "first sentence") && blocks[2].Text != "The first sentence." {
		// case preserved
		if blocks[2].Text != "The first sentence." {
			t.Fatalf("para: %q", blocks[2].Text)
		}
	}
	// bold/italic stripped
	found := false
	for _, b := range blocks {
		if b.Kind == KindParagraph && strings.Contains(b.Text, "bold") {
			found = true
			if strings.Contains(b.Text, "*") {
				t.Fatalf("emphasis markers left: %q", b.Text)
			}
		}
	}
	if !found {
		t.Fatal("missing bold paragraph")
	}
	// list items
	var items int
	for _, b := range blocks {
		if b.Kind == KindListItem {
			items++
			if strings.Contains(b.Text, "https://") {
				t.Fatalf("url leaked: %q", b.Text)
			}
		}
	}
	if items != 2 {
		t.Fatalf("list items: %d", items)
	}
}

func TestStripFrontMatterAndVars(t *testing.T) {
	raw := []byte(`---
title: "The Conservative Frame"
bundle: "the bundle"
geometry:
  - margin=0.6in
---

## Hello

This is {{.bundle}} from {{.title}}.
`)
	vars, err := ParseFrontMatterVars(raw)
	if err != nil {
		t.Fatal(err)
	}
	if vars["title"] != "The Conservative Frame" {
		t.Fatalf("title: %q", vars["title"])
	}
	if vars["bundle"] != "the bundle" {
		t.Fatalf("bundle: %q", vars["bundle"])
	}
	body := string(StripFrontMatter(raw))
	if strings.Contains(body, "geometry") {
		t.Fatalf("front matter not stripped: %q", body)
	}
	blocks := ParseChapter(raw, vars)
	if len(blocks) != 2 {
		t.Fatalf("blocks: %+v", blocks)
	}
	if !strings.Contains(blocks[1].Text, "the bundle") {
		t.Fatalf("var not expanded: %q", blocks[1].Text)
	}
	if strings.Contains(blocks[1].Text, "{{") {
		t.Fatalf("template left: %q", blocks[1].Text)
	}
}

func TestStripFencedCode(t *testing.T) {
	src := "Before\n\n```go\nfmt.Println(1)\n```\n\nAfter"
	blocks := Parse(src)
	for _, b := range blocks {
		if strings.Contains(b.Text, "Println") {
			t.Fatalf("code not stripped: %+v", blocks)
		}
	}
}

func TestIsMetadataOnly(t *testing.T) {
	raw := []byte(`---
title: "X"
bundle: "y"
---
`)
	blocks := ParseChapter(raw, map[string]string{"title": "X"})
	if !IsMetadataOnly(blocks) {
		t.Fatalf("expected metadata only, got %+v", blocks)
	}
}
